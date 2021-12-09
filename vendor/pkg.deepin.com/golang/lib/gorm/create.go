package gorm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	pkgreflect "pkg.deepin.com/golang/lib/reflect"
)

var (
	beforecreate           = gorm.DefaultCallback.Create().Get("gorm:before_create")
	updatetimestamp        = gorm.DefaultCallback.Create().Get("gorm:update_time_stamp")
	create                 = gorm.DefaultCallback.Create().Get("gorm:create")
	forcereloadaftercreate = gorm.DefaultCallback.Create().Get("gorm:force_reload_after_create")
)

func init() {
	gorm.DefaultCallback.Create().Replace("gorm:before_create", func(scope *gorm.Scope) {
		value := scope.IndirectValue()
		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			beforecreateR(scope)
		}
		beforecreate(scope)
	})

	gorm.DefaultCallback.Create().Replace("gorm:update_time_stamp", func(scope *gorm.Scope) {
		value := scope.IndirectValue()
		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			updateTimeStampForCreateCallback(scope)
		} else {
			updatetimestamp(scope)
		}
	})

	gorm.DefaultCallback.Create().Replace("gorm:create", func(scope *gorm.Scope) {
		value := scope.IndirectValue()
		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			createCallback(scope)
		} else {
			create(scope)
		}
	})

	gorm.DefaultCallback.Create().Replace("gorm:force_reload_after_create", func(scope *gorm.Scope) {
		value := scope.IndirectValue()
		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			forceReloadAfterCreateCallback(scope)
		} else {
			forcereloadaftercreate(scope)
		}
	})
}

//beforecreateR 进行默认值的赋值操作
func beforecreateR(scope *gorm.Scope) {
	value := scope.IndirectValue()
	length := value.Len()
	for _, item := range scope.Fields() { //set default
		if v, has := item.TagSettingsGet("DEFAULT"); has { //hear maby muti
			v = strings.Trim(v, "'")
			for i := 0; i < length; i++ {
				field := value.Index(i) //hear is a struct
				fieldvalue := pkgreflect.GetFieldValue(field, item.Names...)
				if fieldvalue.Kind() == reflect.Invalid || !fieldvalue.IsZero() || pkgreflect.IsNil(fieldvalue) || !fieldvalue.CanSet() {
					continue
				}
				setDefaultValue(fieldvalue, v)
			}

		}
	}
}

func setDefaultValue(value reflect.Value, str string) {
	switch value.Kind() {
	case reflect.Bool:
		if strings.ToUpper(str) == "TRUE" {
			value.SetBool(true)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			panic(fmt.Errorf("default value %s not format field type", str))
		}
		value.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			panic(fmt.Errorf("default value %s not format field type", str))
		}
		value.SetUint(i)
	case reflect.String:
		value.SetString(str)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			panic(fmt.Errorf("default value %s not format field type", str))
		}
		value.SetFloat(f)
	}
}

// updateTimeStampForCreateCallback will set `CreatedAt`, `UpdatedAt` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	now := reflect.ValueOf(gorm.NowFunc())
	value := scope.IndirectValue()

	createdAtField, cok := scope.FieldByName("CreatedAt")
	if cok {
		if now.Type().ConvertibleTo(createdAtField.StructField.Struct.Type) {
			cok = true
		} else {
			cok = false
		}
	}

	updatedAtField, uok := scope.FieldByName("UpdatedAt")
	if uok {
		if now.Type().ConvertibleTo(updatedAtField.StructField.Struct.Type) {
			uok = true
		} else {
			uok = false
		}
	}

	if cok || uok {
		length := value.Len()
		for i := 0; i < length; i++ {
			item := value.Index(i)
			if cok {
				//hear need by path
				field := pkgreflect.GetFieldValue(item, createdAtField.Names...)
				if field.CanSet() && field.IsZero() {
					field.Set(now)
				}
			}
			if uok {
				field := pkgreflect.GetFieldValue(item, updatedAtField.Names...)
				if field.CanSet() && field.IsZero() {
					field.Set(now)
				}
			}
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}

// createCallback the callback used to insert data into database
func createCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var fields []*gorm.Field
		var returnfields []*gorm.Field
		for _, field := range scope.Fields() {
			if field.IsNormal && !field.IsIgnored {
				if field.HasDefaultValue { //default value 主要分两种：1.自增长类型的 2.默认值类型的
					if _, has := field.TagSettingsGet("DEFAULT"); has { //自增不计算
						fields = append(fields, field)
					} else {
						returnfields = append(returnfields, field) //自增序列，输出
					}
				} else {
					fields = append(fields, field)
				}
			} else if field.Relationship != nil && field.Relationship.Kind == "belongs_to" { //value
				for _, foreignKey := range field.Relationship.ForeignDBNames { //外键信息
					if foreignField, ok := scope.FieldByName(foreignKey); ok {
						fields = append(fields, foreignField)
					}
				}
			}
		}

		var (
			returningColumn = "*"
			quotedTableName = scope.QuotedTableName()
			primaryField    = scope.PrimaryField()
			extraOption     string
			insertModifier  string
		)

		if str, ok := scope.Get("gorm:insert_option"); ok {
			extraOption = fmt.Sprint(str)
		}
		if str, ok := scope.Get("gorm:insert_modifier"); ok {
			insertModifier = strings.ToUpper(fmt.Sprint(str))
			if insertModifier == "INTO" {
				insertModifier = ""
			}
		}

		if primaryField != nil {
			returnfields = append(returnfields, primaryField)
		}

		columns := make([]string, len(fields))
		for i, item := range fields {
			columns[i] = scope.Quote(item.DBName)
		}

		value := scope.IndirectValue()
		length := value.Len()
		fieldlength := len(fields)
		placeholders := make([]string, length)
		for i := 0; i < length; i++ {
			item := value.Index(i)
			arr := make([]string, fieldlength)
			for j := 0; j < fieldlength; j++ {
				v := pkgreflect.GetFieldValue(item, fields[j].Names...)
				if v.Kind() == reflect.Invalid {
					arr[j] = scope.AddToVars(nil)
				} else {
					arr[j] = scope.AddToVars(v.Interface())
				}

			}
			placeholders[i] = "(" + strings.Join(arr, ",") + ")"
		}

		lastInsertIDOutputInterstitial := ""
		lastInsertIDReturningSuffix := ""

		hasoutput := false
		if v, has := scope.DB().Get("returning"); has {
			if pass, ok := v.(bool); ok && pass {
				isall := false
				if v, has := scope.DB().Get("returning_all"); has {
					if pass, ok := v.(bool); pass && ok {
						isall = true
					}
				}

				outfields := []string{}
				for _, item := range returnfields {
					outfields = append(outfields, scope.Quote(item.DBName))
				}
				if isall {
					for _, item := range columns {
						outfields = append(outfields, item)
					}
				}

				//返回输出
				lastInsertIDOutputInterstitial = scope.Dialect().LastInsertIDOutputInterstitial("", "", columns)
				if lastInsertIDOutputInterstitial != "" {
					//output 类型
					hasoutput = true
					keyword := lastInsertIDOutputInterstitial[:strings.Index(lastInsertIDOutputInterstitial, " ")]
					for i, item := range outfields {
						outfields[i] = fmt.Sprintf(`%v.%v`, quotedTableName, item)
					}
					lastInsertIDOutputInterstitial = fmt.Sprintf(`%v %v`, keyword, strings.Join(outfields, ","))

				} else {
					lastInsertIDReturningSuffix = scope.Dialect().LastInsertIDReturningSuffix(quotedTableName, returningColumn)
					if lastInsertIDReturningSuffix != "" {
						//returning 类型
						hasoutput = true
						keyword := lastInsertIDReturningSuffix[:strings.Index(lastInsertIDReturningSuffix, " ")]
						for i, item := range outfields {
							outfields[i] = fmt.Sprintf(`%v.%v`, quotedTableName, item)
						}
						lastInsertIDReturningSuffix = fmt.Sprintf(`%v %v`, keyword, strings.Join(outfields, ","))
					}
				}
			}
		}

		if len(columns) == 0 {
			panic("insert operate has no columns")
		} else {
			scope.Raw(fmt.Sprintf(
				"INSERT%v INTO %v (%v)%v VALUES %v%v%v",
				addExtraSpaceIfExist(insertModifier),
				scope.QuotedTableName(),
				strings.Join(columns, ","),
				addExtraSpaceIfExist(lastInsertIDOutputInterstitial),
				strings.Join(placeholders, ","),
				addExtraSpaceIfExist(extraOption),
				addExtraSpaceIfExist(lastInsertIDReturningSuffix),
			))
		}

		if hasoutput {
			if rows, err := scope.SQLDB().Query(scope.SQL, scope.SQLVars...); scope.Err(err) == nil {
				defer rows.Close()
				result := reflect.New(value.Type()).Elem()
				resultType := value.Type().Elem()
				isPtr := false
				if resultType.Kind() == reflect.Ptr {
					isPtr = true
					resultType = resultType.Elem()
				}

				for rows.Next() {
					scope.DB().RowsAffected++
					elem := reflect.New(resultType).Elem()
					scope.DB().ScanRows(rows, elem.Addr().Interface())
					if isPtr {
						result.Set(reflect.Append(result, elem.Addr()))
					} else {
						result.Set(reflect.Append(result, elem))
					}
				}

				if scope.Err(err) == nil {
					scope.DB().InstantSet("returning_value", result.Interface())
				}
			}
		} else {
			if result, err := scope.SQLDB().Exec(scope.SQL, scope.SQLVars...); scope.Err(err) == nil {
				scope.DB().RowsAffected, _ = result.RowsAffected()
				if lastid, err := result.LastInsertId(); err == nil {
					scope.DB().InstantSet("returning_lastinsertid", lastid)
				}
			}
		}
		return
	}
}

// forceReloadAfterCreateCallback will reload columns that having default value, and set it back to current object
func forceReloadAfterCreateCallback(scope *gorm.Scope) {
}
