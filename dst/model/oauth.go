package model

import "time"

/*
database:  deepinid
dsn = "root:123456@tcp(127.0.0.1:3306)/deepinid?parseTime=true&loc=Local&charset=utf8mb4,utf8"      # 地址
*/

//Application 公司内部对外提供服务的应用程序相关信息
//table: application
type Application struct {
	ID          string `gorm:"column:ID;type:varchar(32);primary_key"`
	Name        string `gorm:"column:Name;type:varchar(200);not null"`           //应用名称
	WebSite     string `gorm:"column:WebSite;type:varchar(500);not null;unique"` //应用对外提供服务站点
	Description string `gorm:"column:Description;type:varchar(500);not null"`    //应用描述信息
	Author      string `gorm:"column:Author;type:varchar(50);not null"`          //责任人
	WorkNo      string `gorm:"column:WorkNo;type:varchar(10);not null"`          //工号
}

//授权类型
const (
	GruntTypeAuthorizationCode        int = 1 << iota //code码
	GruntTypeImplicit                                 //隐式
	GruntTypeOwnerPasswordCredentials                 //用户身份
	GruntTypeClientCredentials                        //客户端
)

//table: client
//Client oauth client
type Client struct {
	ID                string    `gorm:"column:ID;type:varchar(32);primary_key"`
	CompanyID         string    `gorm:"column:CompanyID;type:varchar(32)"`
	Name              string    `gorm:"column:Name;type:varchar(100)"` //应用名称
	WebSite           string    `gorm:"column:WebSite;type:varchar(2000)"`
	AccessID          string    `gorm:"column:AccessID;type:varchar(100)"`
	AccessKey         string    `gorm:"column:AccessKey;type:varchar(100)"`
	Scopes            string    `gorm:"column:Scopes;type:varchar(2000)"`
	UnGrantURI        string    `gorm:"column:UnGrantURI;type:varchar(200)"`       //撤销授权操作的应用回调地址
	CallBackAccessID  string    `gorm:"column:CallBackAccessID;type:varchar(32)"`  //账号注销相关回调的身份验证id
	CallBackAccessKey string    `gorm:"column:CallBackAccessKey;type:varchar(32)"` //账号注销相关回调的身份验证key
	LogoffAccountURI  string    `gorm:"column:LogoffAccountURI;type:varchar(200)"` //注销账户的回调地址
	Description       string    `gorm:"column:Description"`
	CreatedAt         time.Time `gorm:"column:CreatedAt"`   //创建时间
	UpdatedAt         time.Time `gorm:"column:UpdatedAt"`   //更新时间
	GruntType         int       `gorm:"column:GruntType"`   //授权类型
	IsGrunt           bool      `gorm:"column:IsGrunt"`     //是否需要在界面显示授权
	IsUOS             bool      `gorm:"column:IsUOS"`       //是否是uos项目
	IsDeepin          bool      `gorm:"column:IsDeepin"`    //是否是deepin项目
	IsUngranted       bool      `gorm:"column:IsUngranted"` //代表能否取消的授权应用标签，true代表能取消授权，false代表不能
}

//CompanyType 公司类型
type CompanyType int

//公司类型
const (
	CompanyTypeUnionTech CompanyType = iota + 1
	CompanyTypeOther
)

//table:  company
//Company client's company
type Company struct {
	ID          string      `gorm:"column:ID;primary_key;type:varchar(32)"` //
	Name        string      `gorm:"column:Name;type:varchar(200)"`          //
	CompanyType CompanyType `gorm:"column:CompanyType;type:tinyint"`        //公司类型
	ApplyUserID string      `gorm:"column:ApplyUserID;type:varchar(32)"`    //申请帐号id
}

const (
	AtDEEPIN int32 = iota + 1 //account deepin
	AtUOS                     //account uos
)

//table: openid
//此处确定用户的openid时，需要注意一下几点：
//1.如果用户是已经注册的uos帐号，unionid为已经注册的uosid，
//2.如果用户是已经注册的deepin帐号，unionid为已经注册的deepinid
type OpenID struct {
	ID          string `gorm:"column:ID;type:varchar(32);primary_key"`                        //唯一key
	UserID      string `gorm:"column:UserID;type:varchar(100);unique_index:UserID_ClientID"`  //用户uuid
	CompanyID   string `gorm:"column:CompanyID;type:varchar(100)"`                            //注册应用公司ｉｄ
	ClientID    string `gorm:"column:ClientID;type:varchar(32);unique_index:UserID_ClientID"` //注册应用ｉｄ
	OpenID      string `gorm:"column:OpenID;type:varchar(32)"`                                //openid
	UnionID     string `gorm:"column:UnionID;type:varchar(32)"`                               //unionid
	UID         int32  `gorm:"column:UID"`                                                    //uosid or deepinid
	UIDType     int32  `gorm:"column:UIDType;unique_index:UserID_ClientID"`                   //uditype, 用于解决clientid重复的问题
	IsForbidden bool   `gorm:"column:IsForbidden"`                                            //是否撤销授权
	Scopes      string `gorm:"column:Scopes;type:varchar(4000)"`                              //授权范围
}

//table: scope oauth2协议中的scope
//scope的相关信息可以参考：ref:https://oauth.net/2/scope/
//scope的value格式: modul(.submodul):action
//需要注意的是，scope一旦写入是不允许进行修改的，只能进行复制后重命名操作
type Scope struct {
	ID            string `gorm:"column:ID;type:varchar(32);primary_key"`         //id pk
	ApplicationID string `gorm:"column:ApplicationID;type:varchar(32)"`          //应用id
	Scope         string `gorm:"column:Scope;not null;unique;type:varchar(100)"` //scope
	Description   string `gorm:"column:Description;type:varchar(100);size:255"`  //description
}

//table: unionid
type UnionID struct {
	ID        string `gorm:"column:ID;type:varchar(32);primary_key"`
	UserID    string `gorm:"column:UserID;type:varchar(100);unique_index:idx1"`
	CompanyID string `gorm:"column:CompanyID;type:varchar(100);unique_index:idx1"`
	UnionID   string `gorm:"column:UnionID;type:varchar(100)"`
}
