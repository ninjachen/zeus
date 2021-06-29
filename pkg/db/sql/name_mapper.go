package sql

import "github.com/iancoleman/strcase"

type lowerCamelMapper struct {
}

func (l lowerCamelMapper) Obj2Table(s string) string {
	return strcase.ToLowerCamel(s)
}

func (l lowerCamelMapper) Table2Obj(s string) string {
	return strcase.ToLowerCamel(s)
}
