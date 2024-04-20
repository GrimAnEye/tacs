package server

import (
	l "log/slog"
	"net/http"

	t "tacs/types"
)

// analysisLocal - searches for a login in the local section of the schema
// and if present, launches the template.
//
// If the branch was involved - the login was found, the template was launched, an error occurred
// - returns true, otherwise - false
func analysisLocal(w http.ResponseWriter, r *http.Request, username string) bool {
	var user t.Params
	user, isLocal := t.Scheme.Local.List[username]

	// If the login is found locally
	if isLocal {

		// Preparing data for output
		user := user.Merge(t.Scheme.Local.Default)

		// Checking that a template has been assigned
		if user.Template == "" {
			// If the template was not assigned, return 404 - content not found
			l.Warn("no template assigned to user", l.String("username", username))
			http.NotFound(w, r)
			return true
		}

		if err := t.Templates.ExecuteTemplate(w, user.Template, user.Fields); err != nil {
			l.Error("execute template error",
				l.String("method", r.Method),
				l.String("requestURI", r.RequestURI),
				l.String("requesterIP", r.RemoteAddr),
				l.Any("err", err))
			http.Error(w, "execute template error", http.StatusInternalServerError)
		}

		return true
	}
	return false
}
