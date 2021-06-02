package gobean

import (
	"fmt"
	"reflect"
)

type ElementField struct {
	Parent      *PopulateItem
	StructField reflect.StructField
	FieldValue  reflect.Value
	Wired       bool
	TagValue    string
	Index       int
}

func (elemField *ElementField) FullName(injectionTag string) string {
	return fmt.Sprintf("struct: %s [ %s %s `%s:\"%s\"` ]",
		elemField.Parent.BeanType.String(),
		elemField.StructField.Name,
		elemField.StructField.Type.String(),
		injectionTag,
		elemField.TagValue)
}
