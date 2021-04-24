package gormjqdt

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ReflectionOfGivenInf struct {
	Origin    string
	SnakeCase string
	FromTag   reflect.StructTag
	Kind      reflect.Kind
}

type ArrayReflector map[int]ReflectionOfGivenInf

// GetAllStructField is method to get all field in given struct
func GetAllStructField(i interface{}, toSnakeCase ...bool) ArrayReflector {
	// Check is snake case param is given
	isToLower := true
	if len(toSnakeCase) < 1 {
		isToLower = false
	} else if !toSnakeCase[0] {
		isToLower = false
	}

	// Make map string cosntructor
	arrayReflector := make(ArrayReflector)

	// Loop the struct fields using reflector
	elems := reflect.ValueOf(i).Elem()
	for i := 0; i < elems.NumField(); i++ {
		kind := elems.Field(i).Kind()
		field := elems.Type().Field(i)

		name := field.Name
		nameSnakeCase := ""
		if isToLower {
			nameSnakeCase = ToSnakeCase(name)
		}

		arrayReflector[i] = ReflectionOfGivenInf{
			Origin:    name,
			SnakeCase: nameSnakeCase,
			FromTag:   field.Tag,
			Kind:      kind,
		}
	}

	// Return
	return arrayReflector
}

// ToSnakeCase is method to convert String to snake_case format
// Thanks to https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(s string) string {
	snakeCase := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snakeCase = matchAllCap.ReplaceAllString(snakeCase, "${1}_${2}")

	return strings.ToLower(snakeCase)
}

// StringInArraySimple is method for searching given value in given array
// Thanks to https://codereview.stackexchange.com/users/1361/oneofone
func StringInArraySimple(val string, array map[int]string) (ok bool, i int) {
	for i = range array {
		if ok = array[i] == val; ok {
			return
		}
	}

	return
}

// ParamsValuesProcessing method to process given params into the type of it should be
func ParamsValuesProcessing(i interface{}) (string, bool) {
	var unboxedValue string
	var isArray bool

	switch v := i.(type) {
	case string:
		isArray = false
		unboxedValue = i.(string)

	case []string:
		if len(v) > 1 {
			isArray = true
			unboxedValue = "('"
			unboxedValue += strings.Join(v, "','")
			unboxedValue += "')"
		} else {
			isArray = false
			unboxedValue = v[0]
		}

	case []interface{}:
		ii := i.([]interface{})
		if len(ii) > 1 {
			isArray = true
			var stringSlice []string
			for _, casv := range ii {
				stringSlice = append(stringSlice, ConvertInJsonValToString(casv))
			}

			unboxedValue = "('"
			unboxedValue += strings.Join(stringSlice, "','")
			unboxedValue += "')"
		} else {
			isArray = false
			unboxedValue = ConvertInJsonValToString(ii[0])
		}
	}

	return unboxedValue, isArray
}

// Get gets the first value associated with the given key.
// If there are no values associated with the key, Get returns
// the empty string. To access multiple values, use the map
// directly.
//   // This function copy paste from url go official package
func GetValFromSlice(v map[string][]string, key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func ConvertInJsonValToString(i interface{}, key ...string) string {
	switch v := i.(type) {
	case string:
		return v

	case bool:
		return strconv.FormatBool(v)

	case int, float64:
		return fmt.Sprintf("%v", v)

	case map[string]interface{}:
		if len(key) <= 0 {
			return ""
		}
		return ConvertInJsonValToString(v[key[0]])
	}

	return ""
}

// GetPointerName is method to get slice of pointer names from given slice interface
func GetPointerName(ins map[int]interface{}) (names map[int]string) {
	names = make(map[int]string)

	for k, v := range ins {
		names[k] = reflect.TypeOf(v).String()

		switch true {
		case strings.Contains(names[k], "."):
			names[k] = strings.Split(names[k], ".")[1]

		case strings.Contains(names[k], "*"):
			names[k] = strings.Replace(names[k], "*", "", -1)
		}
	}

	return names
}

// GetDbColumns is method to retrive all column collections from go type struct (or model)
func GetDbColumns(model interface{}) map[int]string {
	modelFields := GetAllStructField(model, true)
	dbColumns := make(map[int]string)

	for i, v := range modelFields {
		var dbColumn string
		var ok bool

		dbColumn, ok = v.FromTag.Lookup("column")
		if !ok {
			dbColumn = v.SnakeCase
		}

		dbColumns[i] = dbColumn
	}

	return dbColumns
}

// GetDbColumnTypes is method to retrive all column type collections from go type struct (or model)
func GetDbColumnTypes(model interface{}) map[string]reflect.Kind {
	modelFields := GetAllStructField(model, true)
	dbColumnTypes := make(map[string]reflect.Kind)

	for _, v := range modelFields {
		var dbColumn string
		var ok bool

		dbColumn, ok = v.FromTag.Lookup("column")
		if !ok {
			dbColumn = v.SnakeCase
		}

		dbColumnTypes[dbColumn] = v.Kind
	}

	return dbColumnTypes
}
