package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"tacs/types"
)

func Test(t *testing.T) {
	type exp struct {
		data string
		code int
	}
	type args struct {
		env    map[string]string
		login  string
		exp    exp
		scheme types.Config
	}
	type test struct {
		name string
		args args
	}
	ErrNotFound := exp{data: "404 page not found\n", code: http.StatusNotFound}
	templ := `{{define "default"}}{{.cn}}
{{.mail}}{{end}}`
	types.Templates.Template, _ = template.New("default").Parse(templ)

	// Structure for declaring tests
	tests := []test{
		{name: "request favicon.ico", args: args{login: "root", exp: ErrNotFound}},
		{name: "not configured server", args: args{login: "root", exp: ErrNotFound}},
		{name: "bad login", args: args{login: "_root", exp: exp{data: "invalid username\n", code: 401}}},
		{name: "bad login2", args: args{login: "root_", exp: exp{data: "invalid username\n", code: 401}}},

		{
			name: "local default", args: args{
				login: "root",
				scheme: types.Config{Local: &types.Local{
					Default: types.Params{
						Template: "default",
						Fields:   map[string]string{"cn": "root", "mail": "root@example.com"},
					},
					List: map[string]types.Params{"root": {}},
				}}, exp: exp{
					data: "root\nroot@example.com",
					code: 200,
				},
			},
		},
	}

	// Creating a router and declaring endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", http.NotFound)
	mux.HandleFunc("/{username}", Analysis)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%s", i, tt.name), func(t *testing.T) {
			// Setting environment variables
			for k, v := range tt.args.env {
				t.Setenv(k, v)
			}
			// Setting program variables
			types.Scheme = tt.args.scheme
			t.Cleanup(func() {
				types.Scheme = types.Config{}
			})

			// Forming and running a request
			r := httptest.NewRequest(http.MethodGet, "/"+tt.args.login, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			res := w.Result()
			defer res.Body.Close()

			if w.Code != tt.args.exp.code {
				t.Errorf("expected code: %d. Got %d", tt.args.exp.code, w.Code)
			}

			data, err := io.ReadAll(res.Body)
			if err != nil {
				t.Errorf("expected error to be nil. Got %v", err)
			}

			if string(data) != tt.args.exp.data {
				t.Errorf("expected %s. Got %v", tt.args.exp.data, string(data))
			}
		})
	}
}
