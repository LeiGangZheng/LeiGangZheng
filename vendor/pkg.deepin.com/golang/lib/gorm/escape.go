package gorm

import "strings"

//MySQLEscape mysql字符转义
var MySQLEscape = strings.NewReplacer(
	"'", "\\'",
	`"`, `\"`,
	`\`, `\\`,
	"_", `\_`,
	"%", `\%`,
)
