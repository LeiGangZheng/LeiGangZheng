package reflect

import (
	"go/token"
	"reflect"
)

//Struct2Map 将struct结构转换为map结构(可以用于gorm的更新操作)
//v:为转化的结构体
//cols:为空值是，全部转化，不为空时，转化指定的列
func Struct2Map(v interface{}, fields ...string) map[string]interface{} {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}
	if value.Kind() != reflect.Struct {
		return nil
	}
	tp := value.Type()
	result := make(map[string]interface{})

	if len(fields) > 0 {
		for _, item := range fields {
			if !token.IsExported(item) {
				continue
			}
			v := value.FieldByName(item)
			if v.Kind() != reflect.Invalid {
				result[item] = v.Interface() //受保护的和Invalid调用interface会panic
			}
		}
	} else {
		for i := 0; i < tp.NumField(); i++ {
			name := tp.Field(i).Name
			if !token.IsExported(name) {
				continue
			}
			result[name] = value.Field(i).Interface()
		}
	}
	return result
}

//FieldString 获取array/slice struct中string 字段
//v:array/slice struct
//f:field
//result:field value
func FieldString(v interface{}, f string) []string {
	pass, value, tpisprt, _, fieldt := fieldValidType(v, f)
	if !pass {
		return nil
	}
	if fieldt.Type.Kind() != reflect.String {
		return nil
	}
	result := make([]string, 0)
	field(value, tpisprt, fieldt, &result)
	return result
}

//FieldInt 获取array/slice struct中int 字段
//v:array/slice struct
//f:field
//result:field value
func FieldInt(v interface{}, f string) []int {
	pass, value, tpisprt, _, fieldt := fieldValidType(v, f)
	if !pass {
		return nil
	}
	if fieldt.Type.Kind() != reflect.Int {
		return nil
	}
	result := make([]int, 0)
	field(value, tpisprt, fieldt, &result)
	return result
}

//FieldInt64 获取array/slice struct中int64 字段
//v:array/slice struct
//f:field
//result:field value
func FieldInt64(v interface{}, f string) []int64 {
	pass, value, tpisprt, _, fieldt := fieldValidType(v, f)
	if !pass {
		return nil
	}
	if fieldt.Type.Kind() != reflect.Int {
		return nil
	}
	result := make([]int64, 0)
	field(value, tpisprt, fieldt, &result)
	return result
}

//Field 获取array/slice struct中 字段
//v:array/slice struct
//f:field
//result:field value
func Field(v interface{}, f string, result interface{}) {
	pass, value, tpisprt, _, fieldt := fieldValidType(v, f)
	if !pass {
		return
	}
	field(value, tpisprt, fieldt, result)
}

func field(value reflect.Value, tpisprt bool, fieldt *reflect.StructField, result interface{}) {
	rvalue := reflect.ValueOf(result)
	if rvalue.Kind() == reflect.Interface || rvalue.Kind() == reflect.Ptr {
		rvalue = reflect.Indirect(rvalue)
	}
	if rvalue.Kind() != reflect.Slice {
		return
	}
	if rvalue.Type().Elem() != fieldt.Type {
		return
	}
	length := value.Len()
	for i := 0; i < length; i++ {
		item := value.Index(i)
		if tpisprt {
			item = item.Elem()
		}
		v := item.FieldByIndex(fieldt.Index)
		rvalue.Set(reflect.Append(rvalue, v))
	}
}

//是否验证通过
//value
//elem 是否是指针
//elem tp
//字段类型
func fieldValidType(v interface{}, field string) (bool, reflect.Value, bool, reflect.Type, *reflect.StructField) {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = reflect.Indirect(value)
	}

	if value.Kind() != reflect.Array && value.Kind() != reflect.Slice {
		return false, value, false, nil, nil
	}
	tp := value.Type().Elem()
	tpisprt := false
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
		tpisprt = true
	}
	if tp.Kind() != reflect.Struct {
		return false, value, false, nil, nil
	}
	ft, has := tp.FieldByName(field)
	if !has {
		return false, value, false, nil, nil
	}
	return true, value, tpisprt, tp, &ft
}
