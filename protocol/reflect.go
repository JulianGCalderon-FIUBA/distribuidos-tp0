package protocol

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func Serialize[M any](m M) []string {
	v := reflect.ValueOf(m)

	fields := reflect.VisibleFields(v.Type())
	length := len(fields)
	data := make([]string, 0, length)

	for _, field := range fields {

		value := v.FieldByName(field.Name).Interface()
		switch value := value.(type) {
		case int:
			data = append(data, strconv.Itoa(value))
		case string:
			data = append(data, value)
		case time.Time:
			data = append(data, value.Format(time.DateOnly))
		}
	}

	return data
}

func Deserialize[M any](record []string) (m M, err error) {
	ty := reflect.TypeOf(m)

	fields := reflect.VisibleFields(ty)
	expectedSize := len(fields)
	if len(record) != expectedSize {
		err = fmt.Errorf("record should contains %v fields", expectedSize)
		return
	}

	pointerToValue := reflect.ValueOf(&m)
	value := pointerToValue.Elem()

	for idx, field := range fields {
		valueField := value.FieldByIndex(field.Index)

		switch valueField.Interface().(type) {
		case int:
			valueToParse := record[idx]
			var valueToSet int
			valueToSet, err = strconv.Atoi(valueToParse)
			if err != nil {
				err = fmt.Errorf("field %v should be an int", idx)
				return
			}

			valueField.SetInt(int64(valueToSet))
		case string:
			valueToSet := record[idx]
			valueField.SetString(valueToSet)
		case time.Time:
			valueToParse := record[idx]
			var valueToSet time.Time
			valueToSet, err = time.Parse(time.DateOnly, valueToParse)
			if err != nil {
				err = fmt.Errorf("field %v should be a date", idx)
				return
			}
			valueField.Set(reflect.ValueOf(valueToSet))
		}
	}

	return
}
