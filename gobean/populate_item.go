package gobean

import (
	"fmt"
	"github.com/linuzilla/summer/utils"
	"reflect"
	"runtime"
	"strings"
)

type PopulateItem struct {
	Bean      interface{}
	BeanType  reflect.Type
	BeanValue reflect.Value
	// structPtr  bool
	Wired      bool
	Fields     []*ElementField
	WiredCount int
	Source     string
}

func (item *PopulateItem) CheckIsWired() bool {
	item.Wired = item.WiredCount == len(item.Fields)
	return item.Wired
}

func (item *PopulateItem) String() string {
	var str strings.Builder

	elements := item.BeanValue.Elem()
	typeOfT := elements.Type()

	str.WriteString(fmt.Sprintln(">> TypeOf:", item.BeanType))

	for i := 0; i < elements.NumField(); i++ {
		f := elements.Field(i)

		if f.CanInterface() {
			str.WriteString(fmt.Sprintf(">>   %d: %s %s = %v `%s`", i,
				typeOfT.Field(i).Name,
				f.Type(),
				f.Interface(),
				typeOfT.Field(i).Tag))
		} else {
			str.WriteString(fmt.Sprintf(">>   %d: %s %s `%s`", i,
				typeOfT.Field(i).Name,
				f.Type(),
				typeOfT.Field(i).Tag))
		}

		if typeOfT.Field(i).Anonymous {
			str.WriteString(fmt.Sprintln(" (embedded)"))
		} else {
			str.WriteString("\n")
		}
	}
	return str.String()
}

func New(bean interface{}, skip int, injectionTag string) (*PopulateItem, error) {
	function, file, line, _ := runtime.Caller(skip)

	beanType := reflect.TypeOf(bean)

	// structPtr:  BeanType.Kind() == reflect.Ptr && BeanType.Elem().Kind() == reflect.Struct,

	item := &PopulateItem{
		Bean:       bean,
		Wired:      false,
		BeanType:   beanType,
		BeanValue:  reflect.ValueOf(bean),
		WiredCount: 0,
		Source:     fmt.Sprintf("Bean [%s] add via file: [%s:%d], function: [%s]", beanType.String(), utils.Basename(file), line, runtime.FuncForPC(function).Name()),
	}

	var elementFieldList []*ElementField

	if newList, err := item.retrieveFieldsRecursively(injectionTag, reflect.ValueOf(item.Bean).Elem(), elementFieldList); err != nil {
		return item, fmt.Errorf("%s\n%s\n%v", item.Source, item.String(), err)
	} else {
		item.Fields = newList
		item.CheckIsWired()
		return item, nil
	}
}

func (item *PopulateItem) retrieveFieldsRecursively(injectionTag string, elemValue reflect.Value, prevList []*ElementField) (newList []*ElementField, err error) {
	elemType := elemValue.Type()
	newList = prevList

	for i := 0; i < elemValue.NumField(); i++ {
		typeField := elemType.Field(i)

		if tag := typeField.Tag.Get(injectionTag); len(tag) > 0 {
			valueField := elemValue.Field(i)

			newElementField := &ElementField{
				Parent:      item,
				Wired:       false,
				StructField: elemValue.Type().Field(i),
				FieldValue:  elemValue.Field(i),
				Index:       i,
				TagValue:    tag,
			}

			if tag == `+` {
				if !newElementField.StructField.Anonymous {
					return newList, fmt.Errorf(`ERROR: "+" type of inject should only be used in anonymous field`)
				} else {
					newList, err = item.retrieveFieldsRecursively(injectionTag, valueField, newList)
				}
			} else {
				newList = append(newList, newElementField)
			}
		}
	}
	return newList, err
}
