package main

import (
	"io/ioutil"

	do "gopkg.in/godo.v2"
	"gopkg.in/godo.v2/util"
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
{{ end }}
`

func generateTasks(p *do.Project) {
	p.Task("builder-boilerplate", nil, func(c *do.Context) {
		context := do.M{
			"builders": []string{"CallBuilder", "DeleteBuilder", "InsectBuilder",
				"InsertBuilder", "RawBuilder", "SelectBuilder", "SelectDocBuilder",
				"UpdateBuilder", "UpsertBuilder"},
		}

		s, err := util.StrTemplate(builderTemplate, context)
		c.Check(err, "Unalbe ")

		ioutil.WriteFile("builders_generated.go", []byte(s), 0644)
		c.Run("go fmt builders_generated.go")
	}).Desc("Generates builder boilerplate code")
}
