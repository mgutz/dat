package main

import (
	"io/ioutil"

	g "github.com/mgutz/godo/v2"
	"github.com/mgutz/godo/v2/util"
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

func generateTasks(p *g.Project) {
	p.Task("builder-boilerplate", func() error {
		context := g.M{
			"builders": []string{"DeleteBuilder", "InsectBuilder",
				"InsertBuilder", "RawBuilder", "SelectBuilder", "SelectJSONBuilder",
				"UpdateBuilder", "UpsertBuilder"},
		}

		s, err := util.StrTemplate(builderTemplate, context)
		if err != nil {
			return err
		}

		ioutil.WriteFile("v1/builders_generated.go", []byte(s), 0644)
		g.Run("go fmt v1/builders_generated.go")
		return nil
	}).Desc("Generates builder boilerplate code")
}
