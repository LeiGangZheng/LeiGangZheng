package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/jinzhu/gorm"
	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/service/utils/config"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/dst"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src"
)

type MyfConfig struct {
	DB_SRC_Oauth *db.ConnConf // DB配置
	DB_SRC_User  *db.ConnConf // DB配置
	DB_DST_User  *db.ConnConf // DB配置
	Log          log.Options  // log配置
}

var (
	cfgPath = flag.String("f", "./user_cfg.toml", "the config file path")
	start   = flag.Int64("start", 0, "the start of read user database")
	step    = flag.Int64("step", 2000, "the step of read user database echo time")
	test    = flag.Bool("test", true, "test the db connect config")
)

func main() {
	flag.Parse()
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
	err = dst.InitDbConn(cfg.DB_DST_User, nil)
	if nil != err {
		panic(fmt.Errorf("init dst db error:[%s]", err.Error()))
	}

	dir, err := ioutil.ReadDir(cfg.Log.Dir)
	for _, d := range dir {
		fn := path.Join([]string{cfg.Log.Dir, d.Name()}...)
		fmt.Println("remove name: ", fn)
		os.RemoveAll(fn)
	}

	// if *test {
	// 	err := src.Ping()
	// 	if err != nil {
	// 		panic(fmt.Errorf("src db test connect error:[%s]", err.Error()))
	// 	}
	// 	err = dst.Ping()
	// 	if err != nil {
	// 		panic(fmt.Errorf("dst db test connect error:[%s]", err.Error()))
	// 	}
	// 	fmt.Println("test success...")
	// 	return
	// }

	// 开始数据迁移
	// user CN or DE (修改配置，切换不同的数据库)
	startTm := time.Now()
	fmt.Println("deal users start...", "start_time:", startTm)
	totalCnt, err := src.GetUserCount()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("no users data should move", "user_count:", totalCnt)
		} else {
			panic(fmt.Errorf("get user total count error:[%s]", err.Error()))
		}
	}

	var i int64
	for i = *start; i < totalCnt+1; i = i + *step {
		var idxEnd int64 = i + *step - 1
		fmt.Printf("deal users [ %d - %d ] ...\n", i, idxEnd)

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

	fmt.Println("deal users end...", "use_time:", time.Now().Sub(startTm).Minutes(), "min")
}
