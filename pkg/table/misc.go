package table

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// StructToMap takes a struct or struct pointer and transforms it
// into a map. The fields must be exported for them to be added to
// the map. In the case of an error, a panic or a nil map may be
// returned.
func StructToMap(ptr interface{}) map[string]interface{} {
	// get the value of the struct
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	// get the underlying type from the reflected value
	typ := val.Type()
	// instantiate a new map
	m := make(map[string]interface{}, val.NumField())
	// loop over each field in the struct
	for i := 0; i < val.NumField(); i++ {
		// return an instance of struct field at this position
		sf := typ.Field(i)
		// check to make sure the field is exported
		if sf.Anonymous {
			// if it's not, skip it
			i--
			continue
		}
		// attempt to add it to the map
		m[sf.Name] = val.Field(i).Interface()
	}
	// we are all done, return the map
	return m
}

// MapToStruct takes a map[string]interface{} type along with a
// pointer to the struct that you wish to have filled out. It will
// attempt to fill out the struct fields using the keys and values
// within the map. Any map keys that do not match the struct field
// names, or values that are not able to be aligned will be ignored.
func MapToStruct(m map[string]interface{}, ptr interface{}) {
	// get the value of the map
	mval := reflect.ValueOf(m)
	if mval.Kind() == reflect.Ptr {
		// if it is a pointer to a map then
		// dereference the pointer (maybe we
		// should panic in this instance)
		mval = mval.Elem()
		// since we got a pointer, better
		// make sure it was in fact a map
		// check the type to ensure it is actually a struct
		if mval.Kind() != reflect.Map {
			log.Panicf("%v type can't have attributes inspected\n", mval.Kind())
		}
	}
	// next, let's ensure we have the correct receiver type
	// which should be a pointer to a struct
	sval := reflect.ValueOf(ptr)
	if sval.Kind() == reflect.Ptr {
		// dereference the pointer
		sval = sval.Elem()
	}
	// check the type to ensure it is actually a struct
	if sval.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", sval.Kind())
	}
	// get a type for the struct
	styp := sval.Type()
	// now that we have ensured that we have a map along
	// with the correct receiver type, we can begin to
	// loop over each field in the struct and attempt to
	// look up the map key with the same struct field name
	// and assign the value of the struct field using the
	// value from the map.
	for i := 0; i < sval.NumField(); i++ {
		// get the struct field value at this index
		sfv := styp.Field(i)
		// sfv := sval.Field(i)
		// check to make sure the value can be set
		if !sfv.IsExported() {
			// if not, skip it
			continue
		}
		// now, we should look to see if we can find
		// a map value for the supplied struct field
		// value that we are on.
		mkv := mval.MapIndex(reflect.ValueOf(sfv.Name))
		fmt.Printf("mkv=%#v\n", mkv)
		// finally, we should set the struct field value
		// to the value found at this map key
		switch mkv.Kind() {
		case reflect.Int:
			sval.Field(i).SetInt(mkv.Int())
		case reflect.String:
			sval.Field(i).Set(mkv.Elem())
		}
	}
}

// GetFieldFromStruct returns a *Field from the struct, looking it up
// by its name. The field in the struct must be exported, else it will
// return nil.
func GetFieldFromStruct(ptr interface{}, name string) *Field {
	// get the value of the struct
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	// attempt to grab the field using the provided name
	sf, ok := val.Type().FieldByName(name)
	if !ok {
		// if it is not exported or could not be found
		// return a nil value instead of *Field
		return nil
	}
	// otherwise, we should be good to go. fill out a
	// new *Field, and return it
	return &Field{
		Name:  sf.Name,
		Kind:  sf.Type.Kind(),
		Value: val.FieldByName(name).Interface(),
	}
}

// GetStructField returns a reflect.StructField from the struct, looking
// it up by its name. The field in the struct must be exported, else it
// will return an empty struct field and a boolean value of false.
func GetStructField(ptr interface{}, name string) (reflect.StructField, bool) {
	// get the value of the struct
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	// attempt to grab the field using the provided name
	// and return it (may be empty if not exported) along
	// with the boolean value provided
	return val.Type().FieldByName(name)
}

// GetStructTag returns the struct tag for an exported field within a
// struct, if it has one. If it does not have one, or if the field is
// not exported, it will return "" along with a boolean value indicating
// false. Otherwise, it will return the struct tag as a reflect.StructTag
// and value of true
func GetStructTag(ptr interface{}, name string) (reflect.StructTag, bool) {
	// get the value of the struct
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	// attempt to grab the field using the provided name
	sf, ok := val.Type().FieldByName(name)
	if !ok {
		// if it is not exported or could not be found
		// return a nil value instead of *Field
		return "", false
	}
	// otherwise, we should be good to go, so just return
	// the struct tag.
	return sf.Tag, true
}

// ParseTagString takes a struct tag (string) and parses and
// returns the key and value of the tag in the form of
// two string values.
func ParseTagString(tag string) (string, string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}
