package templater

import (
	_ "embed"
	"io"
	"text/template"

	"github.com/myceliums/gdb/model"
)

//go:embed main.tmpl
var tmpl string

// WriteTemplate writes the model to the given writer
func WriteTemplate(wr io.Writer, pkgname string, mdl model.Model) error {
	t := template.New(`main`)

	if _, err := t.Parse(tmpl); err != nil {
		return err
	}

	var p struct {
		PkgName          string
		RawConfiguration string
	}

	p.PkgName = pkgname
	p.RawConfiguration = string(mdl.Config())

	return t.Execute(wr, p)
}
