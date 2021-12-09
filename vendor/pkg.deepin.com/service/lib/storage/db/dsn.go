package db

import (
	"net/url"
	"strings"
)

// DSN ...
type DSN struct {
	User     string            // 用户名
	Password string            // 密码
	Net      string            // 网络协议
	Addr     string            // url地址
	DBName   string            // 数据库名称
	Params   map[string]string // 连接参数
}

//
// ParseDSN 解析dsn到DSN结构体
func ParseDSN(conf *ConnConf) (cfg *DSN, err error) {
	cfg = new(DSN)
	dsn := conf.DSN

	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	// 找到最后一个 '/'
	if conf.Dialect == "mysql" {
		foundSlash := false
		for i := len(dsn) - 1; i >= 0; i-- {
			if dsn[i] == '/' {
				foundSlash = true
				var j, k int

				// 如果i<0,则左半部分为空
				if i > 0 {
					// [username[:password]@][protocol[(address)]]
					// 在dsn[:i]找到最后一个'@'
					for j = i; j >= 0; j-- {
						if dsn[j] == '@' {
							// username[:password]
							// 在dsn[:j]中找到第一个':'
							for k = 0; k < j; k++ {
								if dsn[k] == ':' {
									cfg.Password = dsn[k+1 : j]
									break
								}
							}
							cfg.User = dsn[:k]

							break
						}
					}

					// [protocol[(address)]]
					// 在 dsn[j+1:i]中找到第一个 '('
					for k = j + 1; k < i; k++ {
						if dsn[k] == '(' {
							if dsn[i-1] != ')' {
								if strings.ContainsRune(dsn[k+1:i], ')') {
									return nil, ERR_DSN_INVALID
								}
								return nil, ERR_DSN_INVALID
							}
							cfg.Addr = dsn[k+1 : i-1]
							break
						}
					}
					cfg.Net = dsn[j+1 : k]
				}

				// dbname[?param1=value1&...&paramN=valueN]
				// 在dsn[i+1:]中找到第一个'?'
				for j = i + 1; j < len(dsn); j++ {
					if dsn[j] == '?' {
						if err = parseDSNParams(cfg, dsn[j+1:]); err != nil {
							return
						}
						break
					}
				}
				cfg.DBName = dsn[i+1 : j]

				break
			}
		}
		if !foundSlash && len(dsn) > 0 {
			return nil, ERR_DSN_INVALID
		}
		return
	}

	// host=127.0.0.1 user=postgres dbname=test sslmode=disable password=123456
	if conf.Dialect == "postgres" {
		arr := strings.Split(conf.DSN, " ")
		m := map[string]interface{}{}
		for i := 0; i < len(arr); i++ {
			kv := strings.Split(arr[i], "=")
			if len(kv) < 2 {
				break
			}
			m[strings.Trim(kv[0], " ")] = kv[1]
		}
		if _, ok := m["user"]; ok {
			cfg.User = m["user"].(string)
		}
		if _, ok := m["password"]; ok {
			cfg.Password = m["password"].(string)
		}
		if _, ok := m["dbname"]; ok {
			cfg.DBName = m["dbname"].(string)
		}
		if _, ok := m["host"]; ok {
			cfg.Addr = m["host"].(string)
		}
	}

	return
}

func parseDSNParams(cfg *DSN, params string) (err error) {
	for _, v := range strings.Split(params, "&") {
		param := strings.SplitN(v, "=", 2)
		if len(param) != 2 {
			continue
		}

		if cfg.Params == nil {
			cfg.Params = make(map[string]string)
		}
		value := param[1]
		if cfg.Params[param[0]], err = url.QueryUnescape(value); err != nil {
			return
		}
	}
	return
}
