package main

import (
	"flag"
	"fmt"
	"time"

	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/service/utils/config"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/dst"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src"
)

type MyfConfig struct {
	DB_SRC_Oauth *db.ConnConf // DB配置
	DB_DST_Oauth *db.ConnConf // DB配置
	Log          log.Options  // log配置
}

var (
	cfgPath = flag.String("f", "./oauth_cfg.toml", "the config file path")
)

func main() {
	// 加载配置文件
	var cfg MyfConfig
	err := config.Load(*cfgPath, &cfg)
	if nil != err {
		panic(fmt.Errorf("load cfg error:[%s]", err.Error()))
	}

	// init
	log.InitLogger(cfg.Log)
	err = src.InitDbConn(nil, cfg.DB_SRC_Oauth)
	if nil != err {
		panic(fmt.Errorf("init src db error:[%s]", err.Error()))
	}
	err = dst.InitDbConn(nil, cfg.DB_DST_Oauth)
	if nil != err {
		panic(fmt.Errorf("init dst db error:[%s]", err.Error()))
	}

	// 开始数据迁移
	fmt.Println("deal client start...", "time:", time.Now())
	err = dst.WriteClient()
	if nil != err {
		panic(fmt.Errorf("write client db error:[%s]", err.Error()))
	}
	fmt.Println("deal client end...", "time:", time.Now())
}
