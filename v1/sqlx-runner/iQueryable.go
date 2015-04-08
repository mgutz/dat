package runner

import "github.com/mgutz/dat/v1"

// IQueryable is an object that can be queried.
type IQueryable interface {
	DeleteFrom(table string) *dat.DeleteBuilder
	Exec(cmd string, args ...interface{}) (*dat.Result, error)
	ExecBuilder(b dat.Builder) error
	ExecMulti(commands ...*dat.Expression) (int, error)
	InsertInto(table string) *dat.InsertBuilder
	Insect(table string) *dat.InsectBuilder
	Select(columns ...string) *dat.SelectBuilder
	SelectDoc(columns ...string) *dat.SelectDocBuilder
	SQL(sql string, args ...interface{}) *dat.RawBuilder
	Update(table string) *dat.UpdateBuilder
	Upsert(table string) *dat.UpsertBuilder
}
