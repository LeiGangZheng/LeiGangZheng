package gorm

import (
	"go/ast"
	"reflect"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
)

type tabler interface {
	TableName() string
}

type dbTabler interface {
	TableName(*gorm.DB) string
}

type dbTable struct {
	tp     reflect.Type
	name   string
	fields map[string]*dbField
}

func (dt *dbTable) ColumnName(field string) string {
	if f, has := dt.fields[field]; has {
		return f.name
	}
	return ""
}

type dbField struct {
	field    reflect.StructField
	name     string            //db field name
	isignore bool              //is not export
	ispk     bool              //is pk
	tag      map[string]string //sql tag
}

func (field *dbField) parseSQLTag() {
	setting := map[string]string{}
	for _, str := range []string{field.field.Tag.Get("sql"), field.field.Tag.Get("gorm")} {
		if str == "" {
			continue
		}
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	field.tag = setting
}

var nameCache = sync.Map{} // map[reflect.Type]*dbTable{}
//SingularTable need set with gorm func SingularTable
var SingularTable bool

//Table 获取数据结构的表结构
func Table(v interface{}) *dbTable {
	var tp reflect.Type
	var vistp bool //标记v是否是reflect type
	if _, pass := v.(reflect.Type); pass {
		vistp = true
		tp = v.(reflect.Type)
	} else {
		tp = reflect.TypeOf(v)
	}

	if tp.Kind() == reflect.Ptr || tp.Kind() == reflect.Interface {
		tp = tp.Elem()
	}
	if tp.Kind() != reflect.Struct {
		panic("must be struct")
	}
	if v, has := nameCache.Load(tp); has {
		table := v.(*dbTable)
		return table
	}
	//calc struct
	tableinfo := &dbTable{
		tp:     tp,
		fields: map[string]*dbField{},
	}
	if vistp {
		v = reflect.New(tp).Interface()
	}

	//calc table name,需要注意的是，如果设置gorm.DefaultTableNameHandler在使用时需要在进行一次转换
	if tn, ok := v.(tabler); ok {
		tableinfo.name = tn.TableName()
	} else {
		tableinfo.name = gorm.ToTableName(tp.Name())
		if !SingularTable { //同gorm的规则
			tableinfo.name = inflection.Plural(tableinfo.name)
		}
	}

	for i := 0; i < tp.NumField(); i++ {
		if field := tp.Field(i); ast.IsExported(field.Name) {
			tbfield := &dbField{
				field: field,
			}
			tbfield.parseSQLTag()
			if _, ok := tbfield.tag["-"]; ok {
				tbfield.isignore = true
			} else {
				if _, ok := tbfield.tag["PRIMARY_KEY"]; ok {
					tbfield.ispk = true
				}
			}
			if value, ok := tbfield.tag["COLUMN"]; ok {
				tbfield.name = value
			} else {
				tbfield.name = gorm.ToColumnName(field.Name)
			}
			tableinfo.fields[field.Name] = tbfield
		}
	}
	nameCache.Store(tp, tableinfo)
	return tableinfo
}

//TableName 获取表名
func TableName(v interface{}) string {
	return Table(v).name
}

//ColumnName 获取列名
func ColumnName(v interface{}, field string) string {
	table := Table(v)
	if v, has := table.fields[field]; has {
		return v.name
	}
	return ""
}
