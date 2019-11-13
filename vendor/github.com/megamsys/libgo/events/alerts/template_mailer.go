package alerts

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"path/filepath"
)

func body(name string, mp map[string]string, dir string) (string, error) {
	if dir == "" {
		return "", errors.New(`[meta] Is it there in vertice.conf ?`)
	}

	f := filepath.Join(dir, "mailer", name+".html")
	if _, err := os.Stat(f); err != nil {
		return "", err
	}
	var w bytes.Buffer
	t, err := template.ParseFiles(f)
	if err != nil {
		return "", err
	}

	if err = t.Execute(&w, mp); err != nil {
		return "", err
	}
	return w.String(), nil
}
