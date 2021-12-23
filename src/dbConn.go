package src

import (
	"errors"
	"strings"

	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src/model"
)

var (
	_userDB  *db.Conn = nil
	_oauthDB *db.Conn = nil
	//_errnoteDB *db.Conn = nil
)

// InitDbConn
func InitDbConn(userCf, oauthCf *db.ConnConf) error {
	var err error
	if userCf == nil && oauthCf == nil {
		panic("Not config user or oauth src_db info")
	}

	// 初始化默认的DB链接
	if userCf != nil {
		_userDB, err = db.NewConn1(userCf)
		if err != nil {
			log.Error("New db conn error", "error", err.Error())
			return err
		}
	}

	// 初始化默认的DB链接
	if oauthCf != nil {
		_oauthDB, err = db.NewConn1(oauthCf)
		if err != nil {
			log.Error("New db conn error", "error", err.Error())
			return err
		}
	}

	// 写入并记录（用于排错）
	return nil
}

func Ping(dc *db.Conn) error {
	if dc == nil {
		panic("Not InitDbConn src")
	}
	return dc.Debug().DB().Ping()
}

func GetUserCount() (int64, error) {
	if _userDB == nil {
		panic("Not InitDbConn src_user")
	}

	u := model.User{}
	err := _userDB.Table("user").Last(&u).Error
	if err != nil {
		return 0, err
	}
	return int64(u.ID), nil
}

// user 读取  --- ok
func ReadUsers(idxStart, idxEnd int64) ([]model.User, error) {
	if _userDB == nil {
		panic("Not InitDbConn src_user")
	}

	if idxStart < 0 || idxEnd < idxStart {
		return nil, errors.New("input param error")
	}

	users := []model.User{}
	err := _userDB.Table("user").Where("id between ? and ?", idxStart, idxEnd).Find(&users).Error //大于等于,小于等于
	return users, err
}

// ok  -- 一个账号绑定多个媒体  (没有图像)
func ReadThirdPartyLogin(uid int) ([]model.ThirdPartyLogin, error) {
	if _oauthDB == nil {
		panic("Not InitDbConn src_oauth")
	}

	tpl := []model.ThirdPartyLogin{}
	err := _oauthDB.Table("third_party_login").Find(&tpl, "user_id = ?", uid).Error
	return tpl, err
}

// ok
func GetTplUnionID(tpl *model.ThirdPartyLogin) string {
	switch tpl.Type {
	case model.ServiceNameQQ:
		return tpl.ID[len(model.ServiceNameQQ)+1:]
	case model.ServiceNameSina:
		return tpl.ID[len(model.ServiceNameSina)+1:]
	case model.ServiceNameWechat:
		return tpl.ID[len(model.ServiceNameWechat)+1:]
	case model.ServiceNameGithub:
		return tpl.ID[len(model.ServiceNameGithub)+1:]
	}
	return ""
}

func GetTplOpenID(tpl *model.ThirdPartyLogin) string {
	switch tpl.Type {
	case model.ServiceNameWechat:
		return tpl.ID[len(model.ServiceNameWechat)+1:]
	}
	return ""
}

//  -- ok
func GetWebSite(url string) string {
	us := strings.Split(url, ";(:);")
	if len(us) == 1 {
		return us[0]
	}

	srst := ""
	for _, v := range us {
		srst = srst + v + ","
	}
	return srst[:len(srst)-1]
}

//  -- ok
func ReadClient() ([]model.Client, error) {
	if _oauthDB == nil {
		panic("Not InitDbConn src_oauth")
	}

	cs := []model.Client{}
	err := _oauthDB.Table("client").Find(&cs).Error
	if err != nil {
		return nil, err
	}
	return cs, err
}

//  -- ok
func GetClientScope(clientID string) (string, error) {
	if _oauthDB == nil {
		panic("Not InitDbConn src_oauth")
	}

	css := []model.ClientScope{}
	// err := _oauthDB.Table("client_scope").Select("scope_id").Where("client_id = ?", clientID).Find(&css).Error
	err := _oauthDB.Table("client_scope").Select("scope_id").GetList(&css, "client_id = ?", clientID)
	if err != nil {
		return "", err
	}
	if len(css) <= 0 {
		return "", errors.New("query client_scope empty")
	}

	ids := []int{}
	for _, vc := range css {
		ids = append(ids, vc.ScopeID)
	}
	cs := []model.Scope{}
	err = _oauthDB.ExecSQL(&cs, "SELECT * FROM scope WHERE id IN (?)", ids)
	if err != nil {
		return "", err
	}

	scopes := ""
	for _, v := range cs {
		scopes += v.Name + ","
	}
	if 0 == len(scopes) {
		return "", errors.New("query scope empty")
	}
	return scopes[:len(scopes)-1], nil
}

////////////////////////////////////////////////
//
func ReadUserProfile(uid int) (*model.UserProfile, error) {
	if _userDB == nil {
		panic("Not InitDbConn src_user")
	}

	rst := model.UserProfile{} //Find(&tpl, "user_id = ?", uid).Error
	err := _userDB.Table("user_profile").Last(&rst, "user_id = ?", uid).Error
	if err != nil {
		return nil, err
	}
	return &rst, err
}

// ReadListThirdPartyLogin
func ReadListThirdPartyLogin() ([]model.ThirdPartyLogin, error) {
	return nil, nil
	// if _oauthDB == nil {
	// 	return nil, errors.New("db or input param error")
	// }

	// tpl := []model.ThirdPartyLogin{}
	// err := _oauthDB.Table("third_party_login").Find(&tpl, "user_id = ?", uid).Error
	// return tpl, err
}

// write Errlog
// func Errloged(srcID, unionID int) error {
// 	if _errnoteDB == nil {
// 		return errors.New("input param error")
// 	}
// 	return nil
// 	//return _errnoteDB.Create(&model.UserErrlog{SID: srcID, UnionID: unionID}).Error
// }
