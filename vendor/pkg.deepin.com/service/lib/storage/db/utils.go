package db

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"
)

// Utils ...
type Utils struct{}

// FileWithLineNum 获取db_caller文件及代码位置
func (utils *Utils) FileWithLineNum() string {
	var goSrcRegexp = regexp.MustCompile(`jinzhu/gorm(@.*)?/.*.go`)
	var goTestRegexp = regexp.MustCompile(`jinzhu/gorm(@.*)?/.*test.go`)
	var connSrcRegexp = regexp.MustCompile(`conn.go`) // TOFIX: 'pkg.deepin.com/service/lib/gorm/conn.go'

	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && ((!goSrcRegexp.MatchString(file) || goTestRegexp.MatchString(file)) && !connSrcRegexp.MatchString(file)) {
			return fmt.Sprintf("%v:%v", file, line)
		}
	}
	return ""
}

// HandleScopeSQL 拼接sql
func (utils *Utils) HandleScopeSQL(logSql bool, tp string, oriSql string, args []interface{}) (sql string) {
	var (
		sqlRegexp                = regexp.MustCompile(`\?`)
		numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
	)

	if !logSql {
		return oriSql
	}

	formattedValues := make([]string, 0)
	for _, value := range args {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		if indirectValue.IsValid() {
			value = indirectValue.Interface()
			if t, ok := value.(time.Time); ok {
				if t.IsZero() {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", "0000-00-00 00:00:00"))
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
				}
			} else if b, ok := value.([]byte); ok {
				if str := string(b); isPrintable(str) {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
				} else {
					formattedValues = append(formattedValues, "'<binary>'")
				}
			} else if r, ok := value.(driver.Valuer); ok {
				if value, err := r.Value(); err == nil && value != nil {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			} else {
				switch value.(type) {
				case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
					formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
				default:
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			}
		} else {
			formattedValues = append(formattedValues, "NULL")
		}
	}

	// 区分$n和?符号
	if numericPlaceHolderRegexp.MatchString(oriSql) {
		for _, value := range formattedValues {
			compile := regexp.MustCompile(`\$[0-9]+`)
			submatch := compile.FindStringSubmatch(oriSql)
			oriSql = strings.Replace(oriSql, submatch[0], value, 1)
		}
		sql = oriSql
	} else {
		formattedValuesLength := len(formattedValues)
		for index, value := range sqlRegexp.Split(oriSql, -1) {
			sql += value
			if index < formattedValuesLength {
				sql += formattedValues[index]
			}
		}
	}
	return
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
