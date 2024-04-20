package server

import (
	l "log/slog"
	"net/http"

	t "tacs/types"
)

func Analysis(w http.ResponseWriter, r *http.Request) {
	rawUsername := r.PathValue("username")
	username := loginValidation(rawUsername)
	if username == "" {
		l.Warn("invalid username", l.String("raw", rawUsername), l.String("parse", username))
		http.Error(w, "invalid username", http.StatusUnauthorized)
		return
	}

	// Checking if a user exists in a local handler
	if t.Scheme.Local != nil {
		if analysisLocal(w, r, username) {
			return
		}
		l.Debug("local user not found", l.String("user", username))
	}

	// Checking whether a user is in the LDAP handler
	if t.Scheme.Ldap != nil {
		if analysisLdap(w, r, username) {
			return
		}
		l.Debug("ldap user not found", l.String("user", username))
	}
	http.NotFound(w, r)
}
