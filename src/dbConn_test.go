package src

import (
	"fmt"
	"testing"

	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/service/utils/config"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src/model"
)

func init() {
	// 加载配置文件
	type MyfConfig struct {
		DB_SRC_User  *db.ConnConf // DB配置
		DB_SRC_Oauth *db.ConnConf // DB配置
		Log          log.Options  // log配置
	}
	var cfg MyfConfig
	var cfgPath string = "../cfg/cfg.toml"
	err := config.Load(cfgPath, &cfg)
	if nil != err {
		panic(err)
	}

	log.InitLogger(cfg.Log)
	_userDB, err = db.NewConn1(cfg.DB_SRC_User)
	if err != nil {
		log.Error("New db conn error", "error", err.Error())
		panic("xxxxxxxx")
	}

	_oauthDB, err = db.NewConn1(cfg.DB_SRC_Oauth)
	if err != nil {
		panic("xxxxxxxx")
	}
}

//
func TestGetUnionID(t *testing.T) {
	aim := "x111111111"
	tpl := &model.ThirdPartyLogin{
		ID:   "sina_" + fmt.Sprint(aim),
		Type: model.ServiceNameSina,
	}
	unionID := GetTplUnionID(tpl)
	if unionID != aim {
		t.FailNow()
	}

	tpl.ID = fmt.Sprintf("qq_%s", "yzzzzzzzz")
	tpl.Type = model.ServiceNameQQ
	unionID = GetTplUnionID(tpl)
	if unionID != aim {
		t.FailNow()
	}

}

func TestReadClient(t *testing.T) {
	clients, err := ReadClient()
	if err != nil {
		t.FailNow()
	}
	t.Log(clients)

	for _, c := range clients {
		scops, err1 := GetClientScope(c.ID)
		if err1 != nil {
			continue
		}
		t.Log(scops)
	}
}

func TestGetClientScope(t *testing.T) {
	scops, err := GetClientScope("0634ab169bf76a5df39812c4350778c83b3450e4")
	if err != nil {
		t.FailNow()
	}
	t.Log(scops)
}

func TestReadUsers(t *testing.T) {
	us, err := ReadUsers(0, 500)
	if err != nil {
		t.FailNow()
	}
	t.Log(us)
}

func TestReadThirdPartyLogin(t *testing.T) {
	uid := 272206
	tpls, err := ReadThirdPartyLogin(uid)
	if err != nil {
		t.FailNow()
	}

	for _, v := range tpls {
		tplID := GetTplUnionID(&v)
		t.Log(tplID, v.Type)
	}
	t.Log(tpls)
}

func TestGetWebSite(t *testing.T) {
	url := "https://cdn-client-store6.deepin.com/api/user/login;(:);https://store.deepin.com/api/user/login;(:);http://store.deepin.com/api/user/login;(:);https://community-store.deepin.com/api/user/login"
	//url := "1111" //";(:);22222222;(:);333333333gin;(:);5555555555//login"
	s := GetWebSite(url)

	if len(s) != len("https://cdn-client-store6.deepin.com/api/user/login,https://store.deepin.com/api/user/login,http://store.deepin.com/api/user/login,https://community-store.deepin.com/api/user/login") {
		t.FailNow()
	}
	t.Log(s)
}

func TestReadUserProfile(t *testing.T) {
	uid := 307335
	rst, err := ReadUserProfile(uid)
	if err != nil {
		t.FailNow()
	}

	t.Log(rst)
}
