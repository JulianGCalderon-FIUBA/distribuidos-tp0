package protocol

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

// Deserializes any value from a CSV record (list of strings)
// It uses reflect package to access the desired value type in runtime
func Deserialize[M any](record []string) (M, error) {
	ty := reflect.TypeFor[M]()
	var m M
	var value reflect.Value
	var err error

	switch ty.Kind() {
	case reflect.Struct:
		value, _, err = deserializeStruct(ty, record, 0)
	case reflect.Slice:
		value, _, err = deserializeSlice(ty, record, 0)
	default:
		log.Panicf("unimplemented: deserialization of type %v", ty.Kind())
	}

	m = value.Interface().(M)
	return m, err
}

// Deserializes record into a struct
// Panics if `ty` is not a struct type
func deserializeStruct(ty reflect.Type, record []string, cursor int) (reflect.Value, int, error) {
	pValue := reflect.New(ty)
	value := pValue.Elem()

	fields := reflect.VisibleFields(ty)
	for _, fieldTy := range fields {
		field := value.FieldByIndex(fieldTy.Index)

		var fieldValue reflect.Value
		var err error
		fieldValue, cursor, err = deserializePrimitive(fieldTy.Type, record, cursor)
		if err != nil {
			return value, cursor, err
		}

		field.Set(fieldValue)
	}

	return value, cursor, nil
}

// Deserializes record into a slice
// Panics if `ty` is not a slice type
func deserializeSlice(ty reflect.Type, record []string, cursor int) (reflect.Value, int, error) {
	var value reflect.Value

	len, cursor, err := deserializeInt(record, cursor)
	if err != nil {
		return value, cursor, err
	}

	value = reflect.MakeSlice(ty, len, len)
	elemTy := ty.Elem()

	for elemIdx := 0; elemIdx < len; elemIdx++ {
		elem := value.Index(elemIdx)

		var elemValue reflect.Value
		var err error
		elemValue, cursor, err = deserializePrimitive(elemTy, record, cursor)
		if err != nil {
			return value, cursor, err
		}

		elem.Set(elemValue)
	}

	return value, cursor, nil
}

func deserializePrimitive(ty reflect.Type, record []string, cursor int) (reflect.Value, int, error) {
	pValue := reflect.New(ty)
	value := pValue.Elem()

	switch value.Interface().(type) {
	case int:
		var valueToSet int
		var err error
		valueToSet, cursor, err = deserializeInt(record, cursor)
		if err != nil {
			return value, cursor, err
		}
		value.SetInt(int64(valueToSet))
	case string:
		valueToSet, err := advance(record, cursor)
		if err != nil {
			return value, cursor, err
		}
		value.SetString(valueToSet)
		cursor++
	case time.Time:
		valueToParse, err := advance(record, cursor)
		if err != nil {
			return value, cursor, err
		}
		valueToSet, err := time.Parse(time.DateOnly, valueToParse)
		if err != nil {
			return value, cursor, fmt.Errorf("field %v should be a date", cursor)
		}
		value = reflect.ValueOf(valueToSet)
		cursor++

	default:
		log.Panicf("unimplemented: deserialization of type %v", ty)
	}

	return value, cursor, nil
}

func deserializeInt(record []string, cursor int) (int, int, error) {
	var value int
	valueToParse, err := advance(record, cursor)
	if err != nil {
		return value, cursor, err
	}
	value, err = strconv.Atoi(valueToParse)
	if err != nil {
		return value, cursor, fmt.Errorf("field %v should be an int", cursor)
	}
	return value, cursor + 1, nil
}

func advance(record []string, cursor int) (string, error) {
	if len(record) <= cursor {
		return "", fmt.Errorf("field %v is missing", cursor)
	}
	return record[cursor], nil
}
