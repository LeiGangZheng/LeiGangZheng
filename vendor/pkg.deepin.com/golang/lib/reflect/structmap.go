package reflect

import (
	"reflect"
)

//StructRelationMap 将struct集合通过一定规则映射到对应集合的field上（通过field判断是否相等）
//example:
// type A struct {
// 	ID string
// 	B  []B
// }
// type B struct {
// 	AID  string
// 	Name string
// }
// arrayA := []A{{ID: "1"}, {ID: "2"}}
// arrayB := []B{{AID: "1",Name:"zhangsan"}, {AID: "2",Name:"lisi"}, {AID: "1",Name:"wangwu"}}
// StructRelationMap(arrayA,arrayB,"B","ID","AID")
// ==>[]A{{ID:"1",B:[]{{AID:"1",Name:"zhangsan"},{AID:"1",Name:"wangwu"}}},{ID:"2",B:[]{{AID:"2",Name:"lisi"}}}}
//result =>arrayA
//split =>arrayB
//resultfield=>arrayA中用于接收数据的字段
//fielda,fieldb 用于进行比较的字段
func StructRelationMap(result interface{}, split interface{}, resultfield, fieldresult, fieldsplit string) {
	resultvalue, splitvalue, resultElemTp, splitElemTp, resultElemTpPtr, splitElemTpPtr := structRelationMapValidType(result, split, resultfield)
	structRelationMapValidCompareField(resultElemTp, splitElemTp, fieldresult, fieldsplit)
	structRelationMap(resultvalue, splitvalue, resultElemTpPtr, splitElemTpPtr, resultfield, func(resultitem, splititem interface{}) bool {
		return fieldEqual(reflect.ValueOf(resultitem), reflect.ValueOf(splititem), fieldresult, fieldsplit)
	})
}

//StructRelationMapFunc 将struct集合通过一定规则映射到对应集合的field上（通过func比较是否相等）
func StructRelationMapFunc(result interface{}, split interface{}, resultfield string, f func(resultitem, splititem interface{}) bool) {
	resultvalue, splitvalue, _, _, resultElemTpPtr, splitElemTpPtr := structRelationMapValidType(result, split, resultfield)
	structRelationMap(resultvalue, splitvalue, resultElemTpPtr, splitElemTpPtr, resultfield, f)
}

//StructRelationMapIndex 将struct集合通过一定规则映射到对应集合的field上（通过func比较是否相等）主要是针对result和split比对前做过排序处理
func StructRelationMapIndex(result interface{}, split interface{}, resultfield, fieldresult, fieldsplit string) {
	resultvalue, splitvalue, _, _, resultElemTpPtr, splitElemTpPtr := structRelationMapValidType(result, split, resultfield)
	structRelationMapIndex(resultvalue, splitvalue, resultElemTpPtr, splitElemTpPtr, resultfield, func(resultitem, splititem interface{}) bool {
		return fieldEqual(reflect.ValueOf(resultitem), reflect.ValueOf(splititem), fieldresult, fieldsplit)
	})
}

//StructRelationMapFuncIndex 将struct集合通过一定规则映射到对应集合的field上（通过func比较是否相等）主要是针对result和split比对前做过排序处理
func StructRelationMapFuncIndex(result interface{}, split interface{}, resultfield string, f func(resultitem, splititem interface{}) bool) {
	resultvalue, splitvalue, _, _, resultElemTpPtr, splitElemTpPtr := structRelationMapValidType(result, split, resultfield)
	structRelationMap(resultvalue, splitvalue, resultElemTpPtr, splitElemTpPtr, resultfield, f)
}

func structRelationMap(resultvalue, splitvalue reflect.Value, resultElemTpPtr, splitElemTpPtr bool, resultfield string, f func(resultitem, splititem interface{}) bool) {
	for i := 0; i < resultvalue.Len(); i++ {
		item := resultvalue.Index(i)
		if resultElemTpPtr {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}

		coll := item.FieldByName(resultfield)
		if coll.IsNil() {
			coll.Set(reflect.MakeSlice(coll.Type(), 0, 0))
		}
		for j := 0; j < splitvalue.Len(); j++ {
			subitem := splitvalue.Index(j)
			if splitElemTpPtr {
				if f(item.Interface(), subitem.Elem().Interface()) {
					coll = reflect.Append(coll, subitem)
				}
			} else {
				if f(item.Interface(), subitem.Interface()) {
					coll = reflect.Append(coll, subitem)
				}
			}
		}
		item.FieldByName(resultfield).Set(coll)
	}
}

func structRelationMapIndex(resultvalue, splitvalue reflect.Value, resultElemTpPtr, splitElemTpPtr bool, resultfield string, f func(resultitem, splititem interface{}) bool) {
	index := 0
	for i := 0; i < resultvalue.Len(); i++ {
		item := resultvalue.Index(i)
		if resultElemTpPtr {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}

		coll := item.FieldByName(resultfield)
		if coll.IsNil() {
			coll.Set(reflect.MakeSlice(coll.Type(), 0, 0))
		}
		for j := index; j < splitvalue.Len(); j++ {
			subitem := splitvalue.Index(j)
			subitemparm := subitem
			if splitElemTpPtr {
				subitemparm = subitemparm.Elem()
			}
			if f(item.Interface(), subitemparm.Interface()) {
				coll = reflect.Append(coll, subitem)
			} else {
				index = j
				break
			}
		}
		item.FieldByName(resultfield).Set(coll)
	}
}

func structRelationMapValidType(result interface{}, split interface{}, resultfield string) (reflect.Value, reflect.Value, reflect.Type, reflect.Type, bool, bool) {
	resultvalue, splitvalue := structRelationMapValidArray(result, split)
	resultElemTp := resultvalue.Type().Elem()
	splitElemTp := splitvalue.Type().Elem()
	resultElemTpPtr := structRelationMapValidResultFieldType(resultElemTp, splitElemTp, resultfield)
	if resultElemTpPtr {
		resultElemTp = resultElemTp.Elem()
	}
	splitElemTp, splitElemTpPtr := structRelationMapValidSplitItemType(splitElemTp)
	return resultvalue, splitvalue, resultElemTp, splitElemTp, resultElemTpPtr, splitElemTpPtr
}

//验证result和split是否是array或slice类型
//如果是指针，则先获取elem后再进行判断，并返回value
func structRelationMapValidArray(result interface{}, split interface{}) (reflect.Value, reflect.Value) {
	resultvalue := reflect.ValueOf(result)
	if resultvalue.Kind() == reflect.Ptr || resultvalue.Kind() == reflect.Interface {
		resultvalue = resultvalue.Elem()
	}
	splitvalue := reflect.ValueOf(split)
	if splitvalue.Kind() == reflect.Ptr || resultvalue.Kind() == reflect.Interface {
		splitvalue = splitvalue.Elem()
	}

	if !IsArray(resultvalue.Kind()) || !IsArray(splitvalue.Kind()) {
		panic("result 或 split 必须为array或slice")
	}
	return resultvalue, splitvalue
}

//验证result需要填充的字段的类型是否是slice并进行数据类型的比对
//并返回元素是否是指针类型
func structRelationMapValidResultFieldType(resultElemTp, splitElemTp reflect.Type, resultfield string) bool {
	resultElemTpPtr := false
	if resultElemTp.Kind() == reflect.Ptr || resultElemTp.Kind() == reflect.Interface {
		resultElemTp = resultElemTp.Elem()
		if resultElemTp.Kind() != reflect.Struct {
			panic("result elem must struct")
		}
		resultElemTpPtr = true
	}
	resultElemResultFieldTp, has := resultElemTp.FieldByName(resultfield)
	if !has {
		panic("result elem has no field")
	}

	if resultElemResultFieldTp.Type.Kind() != reflect.Slice {
		panic("result elem result field is not a slice")
	}

	if resultElemResultFieldTp.Type.Elem() != splitElemTp {
		panic("result elem result field elem type not equl split elem type")
	}
	return resultElemTpPtr
}

//structRelationMapValidSplitItemType 验证split的元素类型是否是struct
func structRelationMapValidSplitItemType(splitElemTp reflect.Type) (reflect.Type, bool) {
	splitElemTpPtr := false
	if splitElemTp.Kind() == reflect.Ptr || splitElemTp.Kind() == reflect.Interface {
		splitElemTpPtr = true
		splitElemTp = splitElemTp.Elem()
	}
	if splitElemTp.Kind() != reflect.Struct {
		panic("split elem must struct")
	}
	return splitElemTp, splitElemTpPtr
}

//structRelationMapValidCompareField 验证用于比较的字段类型是否相同
func structRelationMapValidCompareField(resultElemTp, splitElemTp reflect.Type, fieldresult, fieldsplit string) {
	resultElemFieldTp, has := resultElemTp.FieldByName(fieldresult)
	if !has {
		panic("result elem has no field")
	}
	splitElemFieldTp, has := splitElemTp.FieldByName(fieldsplit)
	if !has {
		panic("split elem has no field")
	}

	if resultElemFieldTp.Type != splitElemFieldTp.Type {
		panic("compare field type not same")
	}
}

func fieldEqual(v1, v2 reflect.Value, field1, field2 string) bool {
	value1 := v1.FieldByName(field1)
	value2 := v2.FieldByName(field2)
	if value1.Kind() != value2.Kind() {
		return false
	}
	if !value1.IsValid() || !value2.IsValid() {
		return false
	}
	switch value1.Kind() {
	case reflect.Bool:
		return value1.Bool() == value2.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value1.Int() == value2.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value1.Uint() == value2.Uint()
	case reflect.Float32, reflect.Float64:
		return value1.Float() == value2.Float()
	case reflect.Complex64, reflect.Complex128:
		return value1.Complex() == value2.Complex()
	case reflect.Array, reflect.Slice, reflect.Interface, reflect.Ptr, reflect.Struct, reflect.Map, reflect.Func:
		return reflect.DeepEqual(value1.Interface(), value2.Interface())
	case reflect.Chan:
		return value1.Pointer() == value2.Pointer()
	case reflect.String:
		return value1.String() == value2.String()
	case reflect.UnsafePointer:
		return value1.Pointer() == value2.Pointer()
	default:
		panic("not support type")
	}
}
