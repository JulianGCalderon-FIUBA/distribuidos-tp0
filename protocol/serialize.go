package protocol

import (
	"log"
	"reflect"
	"strconv"
	"time"
)

// Serializes any value into a CSV record (list of strings)
// It uses reflect package to access the given value type in runtime
func Serialize(v any) []string {
	value := reflect.ValueOf(v)

	switch value.Kind() {
	case reflect.Struct:
		return serializeStruct(value)
	case reflect.Slice:
		return serializeSlice(value)
	default:
		log.Panicf("unimplemented: serialization of type %v", value.Kind())
	}

	return nil
}

// Serializes a struct value into a CSV record
// Panics if `value` is not a struct
func serializeStruct(value reflect.Value) []string {
	data := make([]string, 0)

	fields := reflect.VisibleFields(value.Type())
	for _, fieldTy := range fields {
		field := value.FieldByIndex(fieldTy.Index)
		data = append(data, serializePrimitive(field)...)
	}

	return data
}

// Serializes a slice value into a CSV record
// Panics if `value` is not a slice
func serializeSlice(value reflect.Value) []string {
	length := value.Len()
	data := []string{strconv.Itoa(length)}

	// range function syntax is not supported in gopls yet
	value.Seq2()(func(_, element reflect.Value) bool {
		data = append(data, serializePrimitive(element)...)
		return true
	})

	return data
}

// Serializes a primitive value into a CSV record
func serializePrimitive(value reflect.Value) []string {
	switch concreteValue := value.Interface().(type) {
	case int:
		return []string{strconv.Itoa(concreteValue)}
	case string:
		return []string{concreteValue}
	case time.Time:
		return []string{concreteValue.Format(time.DateOnly)}
	default:
		log.Panicf("unimplemented: serialization of type %v", value.Type())
	}

	return nil
}
