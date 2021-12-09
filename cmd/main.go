package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/service/utils/config"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/dst"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src"
)

type MyfConfig struct {
	DB_SRC_User  *db.ConnConf // DB配置
	DB_SRC_Oauth *db.ConnConf // DB配置
	DB_DST_User  *db.ConnConf // DB配置
	DB_DST_Oauth *db.ConnConf // DB配置
	Log          log.Options  // log配置
}

var (
	cfgPath = flag.String("f", "./cfg/cfg.toml", "the config file path")
	step    = flag.Int64("s", 20000, "the step of read user database echo time")
	test    = flag.Bool("test", true, "test the db connect config")
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
	err = src.InitDbConn(cfg.DB_SRC_User, cfg.DB_SRC_Oauth)
	if nil != err {
		panic(fmt.Errorf("init src db error:[%s]", err.Error()))
	}
	err = dst.InitDbConn(cfg.DB_DST_User, cfg.DB_DST_Oauth)
	if nil != err {
		panic(fmt.Errorf("init dst db error:[%s]", err.Error()))
	}

	if *test {
		err := src.Ping()
		if err != nil {
			panic(fmt.Errorf("src db test connect error:[%s]", err.Error()))
		}
		err = dst.Ping()
		if err != nil {
			panic(fmt.Errorf("dst db test connect error:[%s]", err.Error()))
		}
		fmt.Println("test success...")
		return
	}

	// 开始数据迁移
	log.Info("deal client start...", "time:", time.Now())
	err = dst.WriteClient()
	if nil != err {
		panic(fmt.Errorf("write client db error:[%s]", err.Error()))
	}
	log.Info("deal client end...", "time:", time.Now())

	// user CN or DE (修改配置，切换不同的数据库)
	log.Info("deal users start...", "time:", time.Now())
	totalCnt, err := src.GetUserCount()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("no users data should move", "user_count:", totalCnt)
		} else {
			panic(fmt.Errorf("get user total count error:[%s]", err.Error()))
		}
	}

	var i int64
	for i = 1; i < totalCnt+1; i = i + *step {
		var idxEnd int64 = i + *step
		uls, err := src.ReadUsers(i, idxEnd)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Debug("no users data should move", "idxStart:", i, "idxEnd:", idxEnd)
				continue
			} else {
				panic(fmt.Errorf("src read users error:[%s], idxStart:[%d], idxEnd:[%d]", err.Error(), i, idxEnd))
			}
		}
		err = dst.WriteUsers(uls)
		if err != nil {
			log.Error("dst write users error", "error", err.Error())
		}
	}
	log.Info("deal users end...", "time:", time.Now())

	//check success
}
