package gorm

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

//QueryPageOrder 排序
type QueryPageOrder struct {
	Field string
	ASC   bool
}

//QuerySearchField 查询
type QuerySearchField struct {
	Field     string
	Values    []interface{}
	IsRange   bool //是否是区间 int time.time
	IsIn      bool //是否是in
	IsPrecise bool //是否like
}

//Between between format
func (qsf *QuerySearchField) Between() string {
	return " %s BETWEEN ? AND ? "
}

//In in format
func (qsf *QuerySearchField) In() string {
	return " %s IN (?) "
}

//Like like format
func (qsf *QuerySearchField) Like() string {
	return " %s LIKE ? "
}

//Where where format
func (qsf *QuerySearchField) Where() string {
	return " %s = ?"
}

//QueryPage 分页查询
//page:查询页
//count:数量
//fieldtrans：转换函数
//qpo：排序
//qf：查询
//fields:输出列
//result:输出结果
func QueryPage(page, count int, qpo []*QueryPageOrder, qf []*QuerySearchField, fields []string, result interface{}, db *gorm.DB) (int64, error) {
	value := reflect.ValueOf(result)
	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		value = value.Elem()
	}
	if value.Kind() != reflect.Slice {
		return 0, errors.New("result must be slice")
	}
	elemtype := value.Type().Elem()
	if elemtype.Kind() == reflect.Interface {
		return 0, errors.New("result slice element must be struct or point to struct ptr")
	}
	if elemtype.Kind() == reflect.Ptr {
		//elemtypeisprt = true
		elemtype = elemtype.Elem()
	}
	if elemtype.Kind() != reflect.Struct {
		return 0, errors.New("result slice element must be struct or point to struct ptr")
	}
	tableinfo := Table(elemtype)
	searchdb := db.New()
	searchdb = searchdb.Table(tableinfo.name)

	if len(fields) != 0 {
		//select field transform
		for i, item := range fields {
			fields[i] = tableinfo.ColumnName(item)
			if fields[i] == "" {
				return 0, errors.New("bad output fields")
			}
		}
		searchdb = searchdb.Select(fields)
	}
	if len(qpo) != 0 {
		orders := make([]string, len(qpo))
		for i, item := range qpo {
			name := tableinfo.ColumnName(item.Field)
			if name == "" {
				return 0, errors.New("bad order parm")
			}
			if item.ASC {
				orders[i] = name + " ASC "
			} else {
				orders[i] = name + " DESC"
			}
		}
		searchdb = searchdb.Order(strings.Join(orders, ","))
	}

	for _, item := range qf {
		name, has := tableinfo.fields[item.Field]
		if !has {
			return 0, errors.New("bad search field parm")
		}
		if item.IsRange {
			if len(item.Values) < 2 {
				var defaultvalue [2]interface{}
				switch name.field.Type.Kind() {
				case reflect.Int:
					if strconv.IntSize == 32 {
						defaultvalue = [2]interface{}{math.MinInt32, math.MaxInt32}
					} else {
						defaultvalue = [2]interface{}{math.MinInt32, math.MaxInt64}
					}
				case reflect.Int8:
					defaultvalue = [2]interface{}{math.MinInt8, math.MaxInt8}
				case reflect.Int16:
					defaultvalue = [2]interface{}{math.MinInt16, math.MaxInt16}
				case reflect.Int32:
					defaultvalue = [2]interface{}{math.MinInt32, math.MaxInt32}
				case reflect.Int64:
					defaultvalue = [2]interface{}{math.MinInt64, math.MaxInt64}
				case reflect.Uint:
					if strconv.IntSize == 32 {
						defaultvalue = [2]interface{}{0, math.MaxUint32}
					} else {
						defaultvalue = [2]interface{}{0, uint64(math.MaxUint64)}
					}
				case reflect.Uint8:
					defaultvalue = [2]interface{}{0, math.MaxUint8}
				case reflect.Uint16:
					defaultvalue = [2]interface{}{0, math.MaxUint16}
				case reflect.Uint32:
					defaultvalue = [2]interface{}{0, math.MaxUint32}
				case reflect.Uint64:
					defaultvalue = [2]interface{}{0, uint64(math.MaxUint64)}
				case reflect.Struct:
					if name.field.Type == reflect.TypeOf(time.Time{}) {
						defaultvalue = [2]interface{}{time.Unix(0, 0), time.Now().Add(20 * 365 * 24 * time.Hour)}
					} else {
						return 0, errors.New("not support type")
					}
				}
				if len(item.Values) == 0 {
					item.Values = append(item.Values, defaultvalue[:]...)
				} else if len(item.Values) == 1 {
					item.Values = append(item.Values, defaultvalue[1])
				}
			}
			searchdb = searchdb.Where(fmt.Sprintf(item.Between(), tableinfo.ColumnName(item.Field)), item.Values...)
		} else if item.IsIn && len(item.Values) > 0 {
			if len(item.Values) == 1 {
				searchdb = searchdb.Where(fmt.Sprintf(item.Where(), tableinfo.ColumnName(item.Field)), item.Values...)
			} else {
				searchdb = searchdb.Where(fmt.Sprintf(item.In(), tableinfo.ColumnName(item.Field)), item.Values)
			}
		} else if item.IsPrecise && (name.field.Type.Kind() == reflect.String || name.field.Type == reflect.TypeOf(sql.NullString{})) && len(item.Values) > 0 {
			if v, ok := item.Values[0].(string); ok {
				searchdb = searchdb.Where(fmt.Sprintf(item.Like(), tableinfo.ColumnName(item.Field)), "%"+MySQLEscape.Replace(v)+"%")
			} else {
				return 0, errors.New("bad parm")
			}
		} else if len(item.Values) > 0 {
			searchdb = searchdb.Where(fmt.Sprintf(item.Where(), tableinfo.ColumnName(item.Field)), item.Values[0])
		}
	}

	total := int64(0)
	if err := searchdb.NewScope(nil).DB().Table(tableinfo.name).Count(&total).Error; err != nil {
		return 0, err
	}
	if page < 1 {
		page = 1
	}
	if count < 1 {
		count = 20
	}
	if count > 2000 {
		return 0, errors.New("invalid count")
	}
	s := (page - 1) * count
	if err := searchdb.Limit(count).Offset(s).Find(result).Error; err != nil {
		return 0, err
	}
	return total, nil
}
