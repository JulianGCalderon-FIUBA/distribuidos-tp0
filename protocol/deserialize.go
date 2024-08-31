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
	var m M
	var value reflect.Value
	ty := reflect.TypeFor[M]()

	switch ty.Kind() {
	case reflect.Struct:
		var err error
		value, _, err = deserializeStruct(ty, record, 0)
		if err != nil {
			return m, err
		}

	case reflect.Slice:
		var err error
		value, _, err = deserializeSlice(ty, record, 0)
		if err != nil {
			return m, err
		}

	default:
		log.Panicf("can't deserialize type %v", ty)
	}

	m = value.Interface().(M)
	return m, nil
}

func deserializeStruct(ty reflect.Type, record []string, cursor int) (reflect.Value, int, error) {
	pValue := reflect.New(ty)
	value := pValue.Elem()

	fields := reflect.VisibleFields(ty)
	for _, fieldTy := range fields {
		valueField := value.FieldByIndex(fieldTy.Index)

		var fieldValue reflect.Value
		var err error
		fieldValue, cursor, err = deserializePrimitive(fieldTy.Type, record, cursor)
		if err != nil {
			return value, cursor, err
		}

		valueField.Set(fieldValue)
	}

	return value, cursor, nil
}

func deserializeSlice(ty reflect.Type, record []string, cursor int) (reflect.Value, int, error) {
	var value reflect.Value

	lenString, err := advance(record, cursor)
	if err != nil {
		return value, cursor, err
	}
	len, err := strconv.Atoi(lenString)
	if err != nil {
		return value, cursor, fmt.Errorf("field %v should be slice length", cursor)
	}
	cursor++

	value = reflect.MakeSlice(ty, len, len)
	elemTy := ty.Elem()

	for elemIdx := 0; elemIdx < len; elemIdx++ {
		sliceElem := value.Index(elemIdx)

		var sliceElemToSet reflect.Value
		var err error
		sliceElemToSet, cursor, err = deserializePrimitive(elemTy, record, cursor)
		if err != nil {
			return value, cursor, err
		}

		sliceElem.Set(sliceElemToSet)
	}

	return value, cursor, nil
}

func deserializePrimitive(ty reflect.Type, record []string, cursor int) (reflect.Value, int, error) {
	pValue := reflect.New(ty)
	value := pValue.Elem()

	switch value.Interface().(type) {
	case int:
		valueToParse, err := advance(record, cursor)
		if err != nil {
			return value, cursor, err
		}
		valueToSet, err := strconv.Atoi(valueToParse)
		if err != nil {
			return value, cursor, fmt.Errorf("field %v should be an int", cursor)
		}
		value.SetInt(int64(valueToSet))
		cursor++
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
		log.Panicf("can't serialize type %v", ty)
	}

	return value, cursor, nil
}

func advance(record []string, cursor int) (string, error) {
	if len(record) <= cursor {
		return "", fmt.Errorf("field %v is missing", cursor)
	}
	return record[cursor], nil
}
