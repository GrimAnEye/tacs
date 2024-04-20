// Wrapper for connecting and querying LDAP
package ldap

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	l "log/slog"
	"os"
	"path/filepath"

	ld "github.com/go-ldap/ldap/v3"
)

// LDAP_matching_rule_in_chain - LDAP operator that allows you to recursively search for an object in subgroups
const LDAP_matching_rule_in_chain string = ":1.2.840.113556.1.4.1941:"

// Connect - performs connection to LDAP by provided parameters
// and returns the connection object.
//
// Requires `TACS_LDAP_SERVER` and `TACS_LDAP_PORT` in environment variables
//
// If `TACS_LDAP_CERT` and `TACS_LDAP_KEY` are set, TLS is enabled and traffic is encrypted.
// Minimum TLS version = 1.2
//
// If `TACS_LDAP_USER` and `TACS_LDAP_PASSWORD` are set, Bind() is executed.
func Connect() (*ld.Conn, error) {
	host := os.Getenv("TACS_LDAP_SERVER")
	port := os.Getenv("TACS_LDAP_PORT")
	certificatePath := os.Getenv("TACS_LDAP_CERT")
	certKeyPath := os.Getenv("TACS_LDAP_KEY")
	username := os.Getenv("TACS_LDAP_USER")
	password := os.Getenv("TACS_LDAP_PASSWORD")

	conn, err := ld.DialURL(fmt.Sprintf("ldap://%s:%s", host, port))
	if err != nil {
		l.Error("dialing error", l.String("host", host), l.String("port", port), l.Any("err", err))
		return nil, err
	}

	// if SSL certificate is available - switching to encrypted communication channel
	if certificatePath != "" && certKeyPath != "" {

		cert, err := os.ReadFile(filepath.Clean(certificatePath))
		if err != nil {
			l.Error("ssl certificate read error",
				l.String("ldapCertPath", certificatePath),
				l.Any("err", err))

			if cErr := conn.Close(); cErr != nil {
				l.Error("error closing ldap connection", l.Any("err", cErr))
				return conn, errors.Join(err, cErr)
			}
			return conn, err
		}

		key, err := os.ReadFile(filepath.Clean(certKeyPath))
		if err != nil {
			l.Error("ssl key read error", l.String("ldapKeyPath", certKeyPath), l.Any("err", err))

			if cErr := conn.Close(); cErr != nil {
				l.Error("error closing ldap connection", l.Any("err", err))
				return conn, errors.Join(err, cErr)
			}
			return conn, err
		}

		// Analyzing the key-certificate mapping
		certAndKey, err := tls.X509KeyPair(cert, key)
		if err != nil {
			l.Error("ssl certificate-key mapping analysis error", l.Any("err", err))

			if cErr := conn.Close(); cErr != nil {
				l.Error("error closing ldap connection", l.Any("err", err))
				return conn, errors.Join(err, cErr)
			}
			return conn, err
		}

		// Trying to switch to TLS
		if err := conn.StartTLS(&tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{certAndKey},
			ServerName:   host,
		}); err != nil {
			l.Error("error of switching to secure channel", l.Any("err", err))
			return conn, err
		}

	}
	// Authenticate under the user, in case of failure - just return the connection
	if username != "" && password != "" {
		if err := conn.Bind(username, password); err != nil {
			l.Error("LDAP authentication error",
				l.String("username", username),
				l.String("host", host),
				l.String("post", port),
				l.Any("err", err),
			)
			return conn, err
		}
	}
	return conn, nil
}

// Accepts query data and returns data from LDAP.
//
// The requested attributes must be in the exact case as they are in LDAP,
// otherwise they cannot be retrieved from the returned data.
//
// Returns raw data in base64 format
func Request(
	conn *ld.Conn,
	searchBase, filter string, requestAttributes []string, rawFormat bool) (
	[]map[string][]string, error,
) {
	searchRequest := ld.NewSearchRequest(
		searchBase,

		ld.ScopeWholeSubtree,
		ld.NeverDerefAliases, 0, 0, false,

		filter, requestAttributes, nil)

	// Запрос к LDAP

	sr, err := conn.Search(searchRequest)
	if err != nil {
		if ld.IsErrorAnyOf(err, ld.LDAPResultNoSuchObject) {
			l.Warn("no objects found",
				l.String("baseDn", searchBase),
				l.String("ldapFilter", filter),
				l.Any("requestAttributes", requestAttributes))
		} else {
			l.Error("LDAP search error",
				l.String("baseDn", searchBase),
				l.String("ldapFilter", filter),
				l.Any("requestAttributes", requestAttributes),
				l.Any("err", err))
		}
		return nil, err
	}

	// Response parsing
	var objects []map[string][]string = make([]map[string][]string, 0)

	for _, entry := range sr.Entries {
		var obj map[string][]string = make(map[string][]string)

		obj["dn"] = []string{entry.DN}
		for _, attr := range requestAttributes {
			if rawFormat {
				rawValues := entry.GetRawAttributeValues(attr)

				var encObj []string
				for _, b := range rawValues {
					encObj = append(encObj, base64.StdEncoding.EncodeToString(b))
				}

				obj[attr] = encObj
				continue
			}
			obj[attr] = entry.GetAttributeValues(attr)
		}
		objects = append(objects, obj)
	}
	return objects, nil
}
