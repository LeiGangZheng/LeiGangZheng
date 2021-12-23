package dst

import (
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"pkg.deepin.com/golang/lib/uuid"
	"pkg.deepin.com/service/lib/log"
	"pkg.deepin.com/service/lib/storage/db"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/dst/model"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src"
	srcm "pkg.deepin.com/web/deepinid_v2/dataMigrate/src/model"
)

var (
	_userDB  *db.Conn
	_oauthDB *db.Conn

	ErrSuccPart error = errors.New("Partial Opt succe")
)

// InitDbConn
func InitDbConn(userCf, oauthCf *db.ConnConf) error {
	var err error
	if userCf == nil && oauthCf == nil {
		panic("Not config user or oauth dst_db info")
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
	return nil
}

func Ping(dc *db.Conn) error {
	if dc == nil {
		panic("Not InitDbConn dst")
	}
	return dc.Debug().DB().Ping()
}

// WriteUsers
func WriteUsers(ins []srcm.User) error {
	if _userDB == nil {
		panic("Not InitDbConn dst_user")
	}
	if ins == nil {
		return errors.New("input param nill")
	}
	if len(ins) == 0 {
		return nil
	}

	succ := 0
	duplicate := 0
	user := &model.User{}
	userInfo := &model.UserInfo{}
	for _, u := range ins {
		tup, _ := src.ReadUserProfile(u.ID)

		// model.User{}
		err := transUser(user, &u)
		if nil != err {
			log.Error("transUser error", err.Error(), "src_data", u)
			//note
			continue
		}
		// model.UserInfo{}
		err = transUserInfo(userInfo, tup, user.ID)
		if nil != err {
			log.Error("transUserInfo error", err.Error(), "src_data", u)
			//note
			continue
		}

		session := _userDB.Begin()
		if session.Error != nil {
			log.Error("db session begin error", session.Error.Error(), "src_data", u)
			return session.Error
		}

		if result, err := session.CommonDB().Exec(fmt.Sprintf("INSERT INTO `%s` VALUES()", "uosidincrement")); err != nil {
			log.Error("write uosidincrement failed", "error", err.Error(), "src_data", u)
			err = session.Rollback().Error
			if err != nil {
				return err
			}
			//note
			continue
		} else {
			uosid, err := result.LastInsertId()
			if err != nil {
				log.Error("write uosidincrement LastInsertId failed", "error", err.Error(), "src_data", u)
				err = session.Rollback().Error
				if err != nil {
					return err
				}
				//note
				continue
			}
			user.UnionID = int(uosid)
			userInfo.UnionID = user.UnionID
		}

		if db := session.Table("user").Create(user); db.Error != nil {
			err := session.Rollback().Error
			if err != nil {
				return err
			}

			merr, ok := db.Error.(*mysql.MySQLError)
			if ok && merr.Number == 1062 {
				duplicate = duplicate + 1
			}
			log.Error("write user failed", "error", db.Error.Error(), "src_data", u, "insert_data", user)
			//note
			continue
		}
		if db := session.Table("userinfo").Create(userInfo); db.Error != nil {
			err := session.Rollback().Error
			if err != nil {
				return err
			}
			merr, ok := db.Error.(*mysql.MySQLError)
			if ok && merr.Number == 1062 {
				duplicate = duplicate + 1
			}
			log.Error("write userinfo failed", "error", db.Error.Error(), "src_data", u, "insert_data", userInfo)
			//note
			continue
		}
		if err := session.Commit().Error; err != nil {
			log.Error("write user userinfo commit failed", "error", err.Error(), "src_data", u)
			err := session.Rollback().Error
			if err != nil {
				return err
			}
			//note
			continue
		}
		//
		succ = succ + 1
		if err = PlatformBind(u.ID, user.ID); err != nil {
			log.Error("tpl data bind failed", "error", err.Error(), "src_user_data", u, "dst_user_data", user)
			//note
		}
	}

	if len(ins) == (succ + duplicate) {
		return nil
	}
	return ErrSuccPart
}

//
func PlatformBind(srcUID int, dstID string) error {
	if _userDB == nil {
		panic("Not InitDbConn dst_user")
	}
	tpls, err := src.ReadThirdPartyLogin(srcUID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error("ReadThirdPartyLogin error", err.Error())
		//note
		return err
	}
	if len(tpls) == 0 {
		return nil
	}

	succ := 0
	dstP := &model.Platform{}
	for _, tpl := range tpls {
		dstP.ID = uuid.UUID32()
		dstP.UserID = dstID                           //&model.User{ID: platform.UserID}
		dstP.PlatformType = getPlatformType(tpl.Type) //平台类型，1微信
		dstP.UnionID = src.GetTplUnionID(&tpl)
		dstP.OpenID = src.GetTplOpenID(&tpl)
		dstP.Sex = model.Secret //性别
		err = _userDB.Table("platform").Create(dstP).Error
		if err != nil {
			log.Error("table platform create error", "error", err.Error(), "insert_data", dstP)
			//note
			continue
		}
		succ = succ + 1
	}

	if len(tpls) != succ {
		return ErrSuccPart
	}
	return nil
}

// *****
func WriteClient() error {
	if _oauthDB == nil {
		panic("Not InitDbConn dst_oauth")
	}

	srcCs, err := src.ReadClient()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debug("no client data should move")
			return nil
		}
		log.Error("ReadClient error", err.Error())
		return errors.New("read client from srcdb error")
	}

	succ := 0
	client := &model.Client{}
	for _, c := range srcCs { //可重入
		if IsExistItemClient(c.ID, c.Secret) {
			succ = succ + 1
			continue
		}

		scops, err1 := src.GetClientScope(c.ID)
		if err1 != nil {
			log.Error("GetClientScope error", err1.Error(), "src_client_data", c)
			//notes
			continue
		}
		err1 = transClient(client, &c, scops)
		if err1 != nil {
			log.Error("transClient error", err1.Error(), "src_client_data", c)
			//note
			continue
		}
		//write
		err1 = _oauthDB.Table("client").Create(client).Error
		if err1 != nil {
			log.Error("table client create error", "error", err1.Error(), "src_client_data", c, "insert_client_data", client)
			//note
			continue
		}
		succ = succ + 1
	}

	if len(srcCs) != succ {
		return ErrSuccPart
	}
	return nil
}

// get
func IsExistItemClient(AccessID, AccessKey string) bool {
	if _oauthDB == nil {
		panic("Not InitDbConn dst_oauth")
	}

	cs := []model.Client{}
	err := _oauthDB.Table("client").Find(&cs, "AccessID = ? AND AccessKey = ?", AccessID, AccessKey).Error
	if err == nil && len(cs) > 0 {
		return true
	}
	return false
}

//xxxxxxxxxxxxxxxxxxxxxxxxxxx        end          xxxxxxxxxxxxxxxxxxxxxxxxxxx
/*
func xxxxx(user *model.User) error {
	union := &model.UnionID{
		ID:        uuid.UUID32(),
		UserID:    user.ID,
		CompanyID: "CompanyID",
		UnionID:   uuid.UUID32(),
	}

	openid := &model.OpenID{
		ID:        uuid.UUID32(),
		UserID:    user.ID,
		CompanyID: union.CompanyID,
		ClientID:  "client.ID",
		OpenID:    uuid.UUID32(),
		UnionID:   union.UnionID,
		UID:       int32(user.UnionID), //uosid or deepinid
		UIDType:   model.AtDEEPIN,
		Scopes:    "", //授权范围
	}

	session := _oauthDB.Begin()
	if session.Error != nil {
		log.Error("", session.Error.Error())
		return session.Error
	}
	if db := session.Create(union); db.Error != nil {
		log.Error("write union failed", "error", db.Error.Error())
		err := session.Rollback().Error
		if err != nil {
			return err
		}
		// note error data
		err = src.Errloged(sid, int(uosid))
		if err != nil {
			log.Error("note failed data error", "error", err.Error())
			return err
		}
		return db.Error
	}
	if db := session.Create(openid); db.Error != nil {
		log.Error("write openid failed", "error", db.Error.Error())
		err := session.Rollback().Error
		if err != nil {
			return err
		}
		// note error data
		err = src.Errloged(sid, int(uosid))
		if err != nil {
			log.Error("note failed data error", "error", err.Error())
			return err
		}
		return db.Error
	}
	if err1 := session.Commit().Error; err1 != nil {
		log.Error("write union openid commit failed", "error", err1.Error())
		err := session.Rollback().Error
		if err != nil {
			return err
		}
		// note error data
		err = src.Errloged(sid, int(uosid))
		if err != nil {
			log.Error("note failed data error", "error", err.Error())
			return err
		}
		return err1
	}
	return nil
}

*/
