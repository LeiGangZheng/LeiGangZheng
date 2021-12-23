package dst

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"pkg.deepin.com/golang/lib/uuid"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/dst/model"
	"pkg.deepin.com/web/deepinid_v2/dataMigrate/src"
	srcm "pkg.deepin.com/web/deepinid_v2/dataMigrate/src/model"
)

func userUOSID() (int64, error) {
	result, err := _userDB.CommonDB().Exec(fmt.Sprintf("INSERT INTO `%s` VALUES()", "uosidincrement"))
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func nullStringValid(s string) bool {
	if len(s) > 0 {
		return true
	}
	return false
}

func getPlatformType(t string) model.PlatformType {
	switch t {
	case srcm.ServiceNameQQ:
		return model.PlatfromTypeQQ
	case srcm.ServiceNameSina:
		return model.PlatfromTypeSina
	case srcm.ServiceNameWechat:
		return model.PlatfromTypeWechat
	case srcm.ServiceNameGithub:
		return model.PlatfromTypeGithub
	}
	return model.PlatfromTypeNULL
}

//////////////////////////////
//https://cdn-client-store6.deepin.com/api/user/login;(:);https://store.deepin.com/api/user/login;(:);http://store.deepin.com/api/user/login;(:);https://community-store.deepin.com/api/user/login
func getWebSite(url string) string {
	us := strings.Split(url, ";(:);")
	if len(us) == 1 {
		return us[0]
	}

	srst := ""
	for _, v := range us {
		srst = srst + v + ","
	}
	return srst
}

// authorization_code,refresh_token
const (
	GruntType_GruntTypeNone                     int = 0
	GruntType_GruntTypeAuthorizationCode        int = 1 //0001
	GruntType_GruntTypeImplicit                 int = 2 //0010
	GruntType_GruntTypeOwnerPasswordCredentials int = 4 //0100
	GruntType_GruntTypeClientCredentials        int = 8 //1000
)

func getGruntType(accTps string) int {
	us := strings.Split(accTps, ",")

	rst := GruntType_GruntTypeNone
	for _, v := range us {
		switch v {
		case "authorization_code":
			rst = rst + GruntType_GruntTypeAuthorizationCode
		}
	}
	return rst
}

func transClient(dstC *model.Client, srcC *srcm.Client, scopes string) error {
	if dstC == nil || srcC == nil || len(scopes) == 0 {
		return errors.New("input param nil")
	}
	dstC.ID = uuid.UUID32()
	dstC.CompanyID = ""
	dstC.Name = srcC.AppName
	dstC.WebSite = src.GetWebSite(srcC.RedirectURI)
	dstC.AccessID = srcC.ID
	dstC.AccessKey = srcC.Secret
	dstC.Scopes = scopes
	dstC.Description = srcC.AppDescription //
	dstC.UnGrantURI = ""                   //撤销授权操作的应用回调地址
	dstC.CallBackAccessID = ""             //账号注销相关回调的身份验证id
	dstC.CallBackAccessKey = ""            //账号注销相关回调的身份验证key
	dstC.LogoffAccountURI = ""             //注销账户的回调地址
	dstC.CreatedAt = time.Now()            //创建时间
	dstC.UpdatedAt = time.Now()            //更新时间
	dstC.GruntType = 9                     //授权类型  -- ok
	dstC.IsGrunt = false                   //是否需要在界面显示授权   -- 三方应用标识
	dstC.IsUOS = true                      //是否是uos项目
	dstC.IsDeepin = false                  //是否是deepin项目
	return nil
}

func transUser(dstU *model.User, srcU *srcm.User) error {
	if dstU == nil || srcU == nil {
		return errors.New("input param nil")
	}
	dstU.ID = uuid.UUID32()
	dstU.UserName = srcU.Username //用户帐号
	dstU.NickName = sql.NullString{
		String: srcU.Nickname,
		Valid:  nullStringValid(srcU.Nickname)} //昵称
	dstU.Email = srcU.SQLEmail //邮箱
	dstU.PhoneNumber = srcU.SQLPhoneNumber
	dstU.Region = model.RegionArea(srcU.Region) //注册区
	dstU.ProfileImage = srcU.ProfileImage       //头像
	dstU.Disabled = srcU.Disabled
	dstU.Switched = srcU.Switched
	dstU.RegisterIP = srcU.RegisterIP //注册时ip
	dstU.LoginIP = srcU.LoginIP       //登入ip
	dstU.LoginTime = srcU.LoginTime   //登入时间
	dstU.FailCount = 0
	dstU.Password = srcU.Password
	dstU.PasswordEmpty = srcU.PasswordEmpty
	//dstU.PasswordUpdatedAt =
	dstU.CreatedAt = srcU.CreatedAt
	dstU.UpdatedAt = srcU.UpdatedAt
	//dstU.UnionID = unionID //供帐号系统合并使用 chinauos帐号系统
	//dstU.DeepinID =            //供帐号系统合并使用 deepin帐号系统
	dstU.UserNameModifyCount = 0
	dstU.NickNameModifyCount = 0
	dstU.NickNameUpdatedAt = time.Now()
	dstU.Status = model.StatusNormal //用户状态(1.正常 2.冻结 3.注销中 4.限制登录)
	return nil
}

func transUserInfo(dstU *model.UserInfo, srcC *srcm.UserProfile, uid string) error {
	if dstU == nil || len(uid) == 0 {
		return errors.New("input param nil")
	}
	dstU.ID = uid
	//dstU.UnionID = unionID
	dstU.Gender = func() model.Gender {
		return model.Secret
	}()
	if srcC != nil {
		dstU.Gender = model.Gender(srcC.GenderID)
		dstU.Birthday = srcC.Birthday
		//dstU.Description = srcC.Signature
	}
	// dstU.RealName = ""
	// dstU.Birthday  =
	// dstU.Description =
	// dstU.HomePage  =
	// dstU.Industry  =
	// dstU.Country =
	// dstU.Province =
	// dstU.City    =
	// dstU.Area    =
	// dstU.Address
	return nil
}
