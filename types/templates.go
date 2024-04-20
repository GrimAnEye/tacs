package types

import (
	"io/fs"
	l "log/slog"
	"path/filepath"
	"strings"
	textTemplate "text/template"
)

const templateExt = ".tmpl"

type Template struct {
	*textTemplate.Template
}

var Templates Template = Template{Template: &textTemplate.Template{}}

// Load - loads .tmpl templates into memory.
// If the path is not specified, the starting point is the program directory
func (t *Template) Load(dir string) error {
	dir = filepath.Clean(dir)

	var paths []string = make([]string, 0)

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			l.Error("template paths collection error",
				l.String("dirPath", dir),
				l.Any("err", err))
			return err
		}
		if !d.IsDir() && strings.Contains(d.Name(), templateExt) {
			l.Debug("template found", l.String("path", path))
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		l.Error("template search error", l.Any("err", err))
		return err
	}

	var err error
	t.Template, err = t.Template.ParseFiles(paths...)
	if err != nil {
		l.Error("parsing templates error",
			l.Any("paths", paths),
			l.Any("err", err))
		return err
	}

	return nil
}
