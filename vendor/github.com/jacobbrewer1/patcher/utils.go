package patcher

import (
	"reflect"
	"strings"
)

// ptr returns a pointer to the value passed in.
func ptr[T any](v T) *T {
	return &v
}

func isPointerToStruct[T any](t T) bool {
	rv := reflect.ValueOf(t)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return false
	}

	return rv.Elem().Kind() == reflect.Struct
}

func dereferenceIfPointer(resource any) any {
	if reflect.TypeOf(resource).Kind() == reflect.Ptr && !reflect.ValueOf(resource).IsNil() {
		return reflect.ValueOf(resource).Elem().Interface()
	}
	return resource
}

func ensureStruct(resource any) {
	if reflect.TypeOf(resource).Kind() != reflect.Struct {
		// Intentionally panic here as this should never happen. This is a programming error.
		panic("resource is not a struct")
	}
}

func getTag(fType *reflect.StructField, tagName string) string {
	tag := fType.Tag.Get(tagName)
	if tag == "" {
		tag = strings.ToLower(fType.Name)
	}

	tags := strings.Split(tag, TagOptSeparator)
	if len(tags) > 1 {
		return tags[0]
	}
	return tag
}

func getValue(fVal reflect.Value) any {
	if fVal.Kind() == reflect.Ptr && fVal.IsNil() {
		return nil
	} else if fVal.Kind() == reflect.Ptr {
		return fVal.Elem().Interface()
	}
	return fVal.Interface()
}

// IsValidType checks if the given value is of a type that can be stored as a database field.
func IsValidType(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Struct, reflect.Ptr:
		return true
	default:
		return false
	}
}

func getTableName(resource any) string {
	typeOf := reflect.TypeOf(resource)
	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}
	return toSnakeCase(typeOf.Name())
}

func toSnakeCase(s string) string {
	result := make([]rune, 0, len(s)*2)
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
