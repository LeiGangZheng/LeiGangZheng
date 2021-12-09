package reflect

import "reflect"

//ToInterface 将数据转换为interface数组
func ToInterface(v interface{})[]interface{}{
	if v==nil{
		return nil
	}
	value:=reflect.ValueOf(v)
	if value.Kind()==reflect.Ptr || value.Kind()==reflect.Interface{
		value=value.Elem()
	}
	if value.Kind()==reflect.Array || value.Kind()==reflect.Slice{
		result:=make([]interface{},value.Len())
		for i:=0;i<len(result);i++{
			if value.Index(i).CanInterface(){
				result[i]=value.Index(i).Interface()
			}else{
				panic("array or slice has item canot interface")
			}
		}
		return result
	}

	if value.CanInterface() {
		return []interface{}{value.Interface()}
	}
	panic("value canot transform to interface")
}
