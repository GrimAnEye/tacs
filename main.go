/*
Copyright GrimAnEye

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	l "log/slog"
	"os"

	_ "tacs/help"

	"tacs/server"
	t "tacs/types"
)

func main() {
	if err := t.Scheme.Load(); err != nil {
		l.Error("scheme loading error", l.Any("err", err))
		os.Exit(1)
	}
	if err := t.Templates.Load(t.Scheme.TemplateDir); err != nil {
		l.Error("templates loading error", l.Any("err", err))
		os.Exit(1)
	}

	server.Start()
}
