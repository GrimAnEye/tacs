package server

import (
	"fmt"
	l "log/slog"
	"maps"
	"net/http"

	ld "tacs/ldap"
	t "tacs/types"

	"github.com/go-ldap/ldap/v3"
)

// analysisLdap - searches for a login in LDAP, and if available, runs a template.
//
// If the branch was involved - the login was found, the template was launched, an error occurred
// - returns true, otherwise - false
func analysisLdap(w http.ResponseWriter, r *http.Request, username string) bool {
	// Check that required parameters are declared
	if t.Scheme.Ldap.Uid == "" {
		l.Error("unique ldap identifier not specified", l.String("uid", ""))
		http.Error(w, "invalid ldap config", http.StatusInternalServerError)
		return true
	}

	// Connecting to LDAP
	lConn, err := ld.Connect()
	if err != nil {
		l.Error("ldap connection error", l.Any("err", err))
		http.Error(w, "ldap connection error", http.StatusInternalServerError)
		return true
	}
	defer func() {
		if err := lConn.Close(); err != nil {
			l.Error("error closing LDAP connection", l.Any("err", err))
		}
	}()

	// Making a filter to search for a user
	userFilter := fmt.Sprintf("(&(%s=%s)%s)", t.Scheme.Ldap.Uid, username, t.Scheme.Ldap.Filter)
	l.Debug("user search in ldap", l.String("filter", userFilter))

	// Searching for a user
	ldapUser, err := ld.Request(lConn, t.Scheme.Ldap.UsersSearchBase, userFilter, nil, false)
	if err != nil || len(ldapUser) == 0 {

		// If the user is not found in LDAP, it returns 404 - content not found
		if ldap.IsErrorAnyOf(err, ldap.LDAPResultNoSuchObject) || len(ldapUser) == 0 {
			l.Warn("user not found in LDAP", l.String("username", username))
			http.NotFound(w, r)

		} else {
			l.Error("ldap search error",
				l.String("filter", userFilter),
				l.Any("err", err))
			http.Error(w, "ldap search error", http.StatusInternalServerError)
		}
		return true
	}

	// Having received the user's DN, I look for his groups in LDAP

	// Is searching in subgroups required?
	subGroup := ""
	if t.Scheme.Ldap.Subgroups {
		subGroup = ld.LDAP_matching_rule_in_chain
	}
	// Creating a filter to search for user groups
	groupFilter := fmt.Sprintf("(member%s=%s)", subGroup, ldapUser[0]["dn"][0])
	l.Debug("search for user groups in LDAP", l.String("filter", groupFilter))

	userGroups, err := ld.Request(lConn, t.Scheme.Ldap.GroupsSearchBase, groupFilter, nil, false)
	if err != nil || len(userGroups) == 0 {
		if ldap.IsErrorAnyOf(err, ldap.LDAPResultNoSuchObject) || len(userGroups) == 0 {
			l.Warn("user groups not found in LDAP", l.String("filter", groupFilter))
		} else {
			l.Error("error searching for user groups in LDAP",
				l.String("filter", groupFilter),
				l.Any("err", err))
			http.Error(w, "ldap search error", http.StatusInternalServerError)
			return true
		}
	}
	// Searches for declared groups in the returned list
	params := func() t.Params {
		for x := range t.Scheme.Ldap.List {
			for y := range userGroups {
				if t.Scheme.Ldap.List[x].Group == userGroups[y]["dn"][0] {
					// When a target group is detected, the loops are completed
					return t.Params{
						Template: t.Scheme.Ldap.List[x].Params.Template,
						Fields:   maps.Clone(t.Scheme.Ldap.List[x].Params.Fields),
					}
				}
			}
		}
		return t.Params{}
	}()

	// If no parameters were found for the user
	if params.Template == "" && !t.Scheme.Ldap.AllowWithoutGroups {
		// I check if it is possible to use a scheme without groups (i.e. with default fields)
		l.Warn("LDAP group is not assigned to the user", l.String("username", username))
		http.NotFound(w, r)
		return true
	}
	// Preparing data for output
	params = params.Merge(t.Scheme.Ldap.Default)

	// Checking that a template has been assigned
	if params.Template == "" {
		// If the template was not assigned, return 404 - content not found
		l.Warn("no template assigned to user", l.String("username", username))
		http.NotFound(w, r)
		return true
	}

	// Finding out which fields should be requested from ldap to fill the template
	userFields := params.Values()
	if len(userFields) != 0 {
		ldapFields, err := ld.Request(lConn,
			ldapUser[0]["dn"][0], "(objectClass=*)", userFields, false)
		if err != nil {
			l.Error("error requesting user fields from ldap",
				l.String("baseDN", ldapUser[0]["dn"][0]),
				l.Any("properties", userFields),
				l.Any("err", err))
			http.Error(w, "ldap search error", http.StatusInternalServerError)
			return true
		}
		// Convert LDAP format to local
		for _, key := range userFields {
			if v, ok := ldapFields[0][key]; ok {
				if len(v) == 0 {
					params.Set(key, "", false)
					continue
				}
				params.Set(key, ldapFields[0][key][0], false)
			}
		}
	}

	// Same for raw bytes
	userFields = params.RawValues()
	if len(userFields) != 0 {
		ldapFields, err := ld.Request(lConn,
			ldapUser[0]["dn"][0], "(objectClass=*)", userFields, true)
		if err != nil {
			l.Error("error requesting user fields from ldap",
				l.String("baseDN", ldapUser[0]["dn"][0]),
				l.Any("properties", userFields),
				l.Any("err", err))
			http.Error(w, "ldap search error", http.StatusInternalServerError)
			return true
		}
		// Convert LDAP format to local
		for _, key := range userFields {
			if v, ok := ldapFields[0][key]; ok {
				if len(v) == 0 {
					params.Set(key, "", true)
					continue
				}
				params.Set(key, ldapFields[0][key][0], true)
			}
		}
	}

	if err := t.Templates.ExecuteTemplate(w, params.Template, params.Fields); err != nil {
		l.Error("execute template error",
			l.String("method", r.Method),
			l.String("requestURI", r.RequestURI),
			l.String("requesterIP", r.RemoteAddr),
			l.Any("err", err))
		http.Error(w, "execute template error", http.StatusInternalServerError)
	}
	return true
}
