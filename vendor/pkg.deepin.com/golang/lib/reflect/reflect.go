package reflect

import (
	"reflect"
	"strconv"
)

//GetFieldValue 获取struct/slice/array/map 中的value
func GetFieldValue(value reflect.Value, path ...string) reflect.Value {
	var kind reflect.Kind
	length := len(path)
	for i := 0; i < length; i++ {
		item := path[i]
		kind = value.Kind()
		if IsNil(value) {
			return value
		}
		if kind == reflect.Ptr || kind == reflect.Interface {
			value = value.Elem()
		}

		switch kind {
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			i, err := strconv.ParseInt(item, 10, 64)
			if err != nil || i < 0 {
				return reflect.Value{}
			}
			length := value.Len()
			if int(i) < length {
				value = value.Index(int(i))
			} else {
				return reflect.Value{}
			}
		case reflect.Struct:
			value = value.FieldByName(item)
		case reflect.Map:
			if value.Type().Key().Kind() != reflect.String {
				return reflect.Value{}
			}
			value = value.MapIndex(reflect.ValueOf(item))
		case reflect.Ptr, reflect.Interface:
			return GetFieldValue(value, path[i:]...)
		default:
			return reflect.Value{}
		}
	}
	return value
}

//IsNil 是否为空
//封装reflect中value.isnil(not panic)
func IsNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
