package reflect

import (
	"reflect"
)

//IsArray 判断数据是否是数组类型
func IsArray(kind reflect.Kind) bool {
	if kind == reflect.Array || kind == reflect.Slice {
		return true
	}
	return false
}
