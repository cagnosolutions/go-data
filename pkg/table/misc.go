package table

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/cagnosolutions/go-data/pkg/format"
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
		if !sf.IsExported() {
			// if it's not, skip it
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
	// get a value to inspect the map and make sure it is not
	// nil and has at least one entry.
	if reflect.ValueOf(m).IsNil() || reflect.ValueOf(m).Len() < 1 {
		log.Panicln("supplied map cannot be nil, or empty")
	}
	// next, let's ensure we have the correct receiver type
	// which should be a pointer to a struct
	val := reflect.ValueOf(ptr)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf("%v type can't have attributes inspected\n", val.Kind())
	}
	// get a type for the struct
	typ := val.Type()
	// loop over each field of the struct
	for i := 0; i < val.NumField(); i++ {
		// return the value for the field in this position
		fld := val.Field(i)
		// check to make sure we can set this field
		if !fld.CanSet() {
			// if it's not, skip it
			continue
		}
		// get the key (name of current field in struct)
		key := typ.Field(i).Name
		// attempt to look this field up in the map
		v, found := m[key]
		if !found {
			// if this key was not in the map, skip
			continue
		}
		// otherwise, attempt to set the struct field using
		// the value that was mapped via the key
		fld.Set(reflect.ValueOf(v))
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

func SprintAny(s string, v interface{}) string {
	// get the value of the interface
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		// dereference the pointer
		val = val.Elem()
	}
	// check the type to ensure it is actually a struct
	if val.Kind() != reflect.Struct {
		log.Panicf(
			"%v type can't have attributes inspected, expected a struct.\n",
			val.Kind(),
		)
	}
	// get the underlying type from the reflected value
	typ := val.Type()
	return os.Expand(
		s, func(attr string) string {
			// loop over each field in the struct
			for i := 0; i < val.NumField(); i++ {
				// return an instance of struct field at this position
				sf := typ.Field(i)
				// check to make sure the field is exported
				if !sf.IsExported() {
					// if it's not, skip it
					continue
				}
				// go to next one if no match is found
				if !LooseMatch(attr, sf.Name, format.ToSnakeCase) {
					continue
				}
				// get the struct field value
				sfv := val.Field(i)
				// return field value
				switch sfv.Kind() {
				case reflect.String:
					return fmt.Sprintf("'%s'", sfv.String())
				default:
					return fmt.Sprintf("%v", sfv.Interface())
				}
			}
			return ""
		},
	)
}

func LooseMatch(s1, s2 string, fn func(s string) string) bool {
	// first do direct match
	if s1 == s2 {
		return true
	}
	// next try case folding match
	if strings.EqualFold(s1, s2) {
		return true
	}
	// check for fn
	if fn == nil {
		return false
	}
	// next try formatting function match
	if strings.EqualFold(s1, fn(s2)) {
		return true
	}
	// otherwise, no match
	return false
}
