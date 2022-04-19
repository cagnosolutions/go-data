package data

import (
	"log"
	"reflect"

	"github.com/cagnosolutions/go-data/pkg/format"
)

type DataModelField struct {
	Name  string
	Value interface{}
	Type  reflect.Kind
}

type DataModel struct {
	Name         string
	Fields       []*DataModelField
	FieldsByName map[string]*DataModelField
}

func NewDataModel(ptr interface{}, tag string) *DataModel {
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	typ := val.Type()
	mod := &DataModel{
		Name:         format.ToSnakeCase(typ.Name()),
		Fields:       make([]*DataModelField, 0, val.NumField()),
		FieldsByName: make(map[string]*DataModelField),
	}
	for i := 0; i < val.NumField(); i++ {
		sf := typ.Field(i)
		if !sf.IsExported() {
			continue
		}
		name := sf.Tag.Get(tag)
		if name == "" {
			name = format.ToSnakeCase(sf.Name)
		}
		f := &DataModelField{
			Name:  name,
			Value: val.Field(i).Interface(),
			Type:  sf.Type.Kind(),
		}
		mod.FieldsByName[name] = f
		mod.Fields = append(mod.Fields, f)
	}
	return mod
}
