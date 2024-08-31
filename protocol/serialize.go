package protocol

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

// Serializes any value into a CSV record (list of strings)
// It uses reflect package to access the given value type in runtime
func Serialize[M any](m M) []string {
	v := reflect.ValueOf(m)
	ty := reflect.TypeOf(m)
	data := make([]string, 0)

	switch ty.Kind() {
	case reflect.Struct:
		data = append(data, serializeStruct(v)...)
	case reflect.Slice:
		data = append(data, serializeSlice(v)...)
	default:
		log.Panicf("can't serialize type %v", ty)
	}

	fmt.Printf("%v", data)

	return data
}

func serializeStruct(value reflect.Value) []string {
	ty := value.Type()
	data := make([]string, 0)

	fields := reflect.VisibleFields(ty)
	for _, fieldTy := range fields {
		field := value.FieldByName(fieldTy.Name)
		data = append(data, serializePrimitive(field)...)
	}
	return data
}

func serializeSlice(value reflect.Value) []string {
	data := make([]string, 0)

	length := value.Len()
	data = append(data, strconv.Itoa(length))

	for elemIdx := 0; elemIdx < length; elemIdx++ {
		elem := value.Index(elemIdx)
		data = append(data, serializePrimitive(elem)...)
	}

	return data
}

func serializePrimitive(value reflect.Value) []string {
	ty := value.Type()
	data := make([]string, 0)

	switch concreteValue := value.Interface().(type) {
	case int:
		data = append(data, strconv.Itoa(concreteValue))
	case string:
		data = append(data, concreteValue)
	case time.Time:
		data = append(data, concreteValue.Format(time.DateOnly))
	default:
		log.Panicf("can't serialize type %v", ty)
	}

	return data
}
