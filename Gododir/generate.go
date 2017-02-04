package main

import (
	"bytes"
	"io/ioutil"

	"text/template"

	do "gopkg.in/godo.v2"
)

var builderTemplate = `
package dat

//// DO NOT EDIT, auto-generated: godo builder-boilerplate

{{ range $idx, $builder := .builders }}
	// Interpolate interpolates this builders sql.
	func (b *{{$builder}}) Interpolate() (string, []interface{}, error) {
		return interpolate(b)
	}

	// IsInterpolated determines if this builder will interpolate when
	// Interpolate() is called.
	func (b *{{$builder}}) IsInterpolated() bool {
		return b.isInterpolated
	}

	// SetIsInterpolated sets whether this builder should interpolate.
	func (b *{{$builder}}) SetIsInterpolated(enable bool) *{{$builder}} {
		b.isInterpolated = enable
		return b
	}

	// CanJSON determines if a builder can output JSON.
	func (b *{{$builder}}) CanJSON() bool {
		{{if canJson $builder}}
		return true
		{{else}}
		return false
		{{end}}
	}
{{ end }}
`

func generateTasks(p *do.Project) {
	p.Task("builder-boilerplate", nil, func(c *do.Context) {
		t, err := template.New("t1").
			Funcs(template.FuncMap{
				"canJson": func(builder string) bool {
					switch builder {
					default:
						return false
					case "JSQLBuilder", "SelectDocBuilder":
						return true
					}
				},
			}).Parse(builderTemplate)
		c.Check(err, "Could not parse template")

		context := do.M{
			"builders": []string{
				"CallBuilder",
				"DeleteBuilder", "InsectBuilder",
				"InsertBuilder", "JSQLBuilder", "RawBuilder",
				"SelectBuilder", "SelectDocBuilder",
				"UpdateBuilder", "UpsertBuilder",
			},
		}

		var tmpl bytes.Buffer
		err = t.Execute(&tmpl, context)
		c.Check(err, "Cannot execute template")
		s := tmpl.String()
		ioutil.WriteFile("dat/builders_generated.go", []byte(s), 0644)
		c.Run("go fmt dat/builders_generated.go")
	}).Desc("Generates builder boilerplate code")
}
