package model

import (
	"database/sql"
	"time"
)

/*
database:  deepinid_user
dsn = "root:123456@tcp(127.0.0.1:3306)/deepinid_user?parseTime=true&loc=Local&charset=utf8mb4,utf8"      # 地址
*/

//PlatformType 第三方平台类型
type PlatformType int32

const (
	PlatfromTypeNULL PlatformType = iota
	PlatfromTypeWechat
	PlatfromTypeQQ
	PlatfromTypeSina
	PlatfromTypeGithub
)

//Valid 验证注册区域
func (pt PlatformType) Valid() bool {
	switch pt {
	case PlatfromTypeWechat, PlatfromTypeQQ, PlatfromTypeSina, PlatfromTypeGithub:
		return true
	default:
		return false
	}
}

//table: platform
//Platform 第三方平台记录表
type Platform struct {
	ID           string       `gorm:"column:ID;type:char(32);primary_key"`   //使用uuid作为帐号的id，需要注意的是，为了加快数据库的查询速度，uuid的格式采用无-的格式
	UserID       string       `gorm:"column:UserID;type:char(32);index"`     //用户id
	PlatformType PlatformType `gorm:"column:PlatformType;type:int(2)"`       //平台类型，1微信
	OpenID       string       `gorm:"column:OpenID;type:varchar(50);index;"` //我们的开放平台id
	UnionID      string       `gorm:"column:UnionID;type:varchar(50);index"` //第三方平台unionid
	NickName     string       `gorm:"column:NickName;type:varchar(100)"`     //昵称
	Avatar       string       `gorm:"column:Avatar;type:varchar(4000)"`      //头像
	Version      int32        `gorm:"column:Version;type:int(2);"`           //版本
	Sex          Gender       `gorm:"column:Sex;type:int(1);"`               //性别
	Extend       string       `gorm:"column:Extend;type:tinytext;"`          //扩展信息
}

//PlatformFlag 绑定时是否覆盖原值的标识
type PlatformFlag struct {
	NickNameFlag bool
	AvatarFlag   bool
}

//user
//RegionArea 注册区域
type RegionArea string

//Gender 性别
type Gender int

//Valid 验证注册区域
func (r RegionArea) Valid() bool {
	if r == RegionAreaCN || r == RegionAreaDE {
		return true
	}
	return false
}

//Valid 验证注册区域
func (g Gender) Valid() bool {
	switch g {
	case Male:
		return true
	case Female:
		return true
	case Secret:
		return true
	default:
		return false
	}
}

//Value 返回注册区域的string值
func (r RegionArea) String() string {
	switch r {
	case RegionAreaCN:
		return "CN"
	case RegionAreaDE:
		return "DE"
	default:
		return ""
	}
}

//注册区域
const (
	RegionAreaCN RegionArea = "CN"
	RegionAreaDE RegionArea = "DE"
)

//性别类型
const (
	Secret Gender = iota
	Male
	Female
)

const (
	//China ...
	China string = "中国"
)

//Status 账号状态
type Status int

const (
	//StatusUnKnow 未知
	StatusUnKnow Status = iota
	//StatusNormal //正常
	StatusNormal
	//StatusFreeze    //冻结
	StatusFreeze
	//StatusClosing  //注销中
	StatusClosing
	//StatusLimited //限制登录
	StatusLimited
)

//User 数据库用户信息
type User struct {
	ID                  string         `gorm:"column:ID;primary_key"`      //使用uuid作为帐号的id，需要注意的是，为了加快数据库的查询速度，uuid的格式采用无-的格式
	UserName            string         `gorm:"column:UserName"`            //用户帐号
	NickName            sql.NullString `gorm:"column:NickName"`            //昵称
	Email               sql.NullString `gorm:"column:Email"`               //用户邮箱
	PhoneNumber         sql.NullString `gorm:"column:PhoneNumber"`         //用户手机号，（中国区用户）
	Region              RegionArea     `gorm:"column:Region"`              //注册区
	ProfileImage        string         `gorm:"column:ProfileImage"`        //头像
	Disabled            bool           `gorm:"column:Disabled"`            //禁用
	Switched            bool           `gorm:"column:Switched"`            //
	RegisterIP          string         `gorm:"column:RegisterIP"`          //注册时ip
	LoginIP             string         `gorm:"column:LoginIP"`             //登入ip
	LoginTime           time.Time      `gorm:"column:LoginTime"`           //登入时间
	FailCount           int            `gorm:"column:FailCount"`           //密码错误次数
	Password            string         `gorm:"column:Password"`            //密码
	PasswordEmpty       bool           `gorm:"column:PasswordEmpty"`       //密码是否为空
	PasswordUpdatedAt   time.Time      `gorm:"column:PasswordUpdatedAt"`   //密码修改时间
	CreatedAt           time.Time      `gorm:"column:CreatedAt" json:"-"`  //创建时间
	UpdatedAt           time.Time      `gorm:"column:UpdatedAt" json:"-"`  //更新时间
	UnionID             int            `gorm:"column:UnionID"`             //供帐号系统合并使用 chinauos帐号系统
	DeepinID            sql.NullInt32  `gorm:"column:DeepinID"`            //供帐号系统合并使用 deepin帐号系统
	UserNameModifyCount int            `gorm:"column:UserNameModifyCount"` //用户名修改次数
	NickNameModifyCount int            `gorm:"column:NickNameModifyCount"` //昵称修改次数
	NickNameUpdatedAt   time.Time      `gorm:"column:NickNameUpdatedAt"`   //昵称修改时间
	Status              Status         `gorm:"column:Status;default:1"`    //用户状态(1.正常 2.冻结 3.注销中 4.限制登录)
}

//## userinfo
//UserInfo 数据库用户拓展信息
type UserInfo struct {
	ID          string    `gorm:"column:ID;primary_key"` //与User表里面ID字段一致
	UnionID     int       `gorm:"column:UnionID"`        //添加该字段仅仅用于进行api接口兼容和分页查询
	RealName    string    `gorm:"column:RealName"`
	Gender      Gender    `gorm:"column:Gender"`
	Birthday    time.Time `gorm:"column:Birthday"`
	Description string    `gorm:"column:Description"`
	HomePage    string    `gorm:"column:HomePage"`
	Industry    int       `gorm:"column:Industry"`
	Country     string    `gorm:"column:Country"`
	Province    string    `gorm:"column:Province"`
	City        string    `gorm:"column:City"`
	Area        string    `gorm:"column:Area"`
	Address     string    `gorm:"column:Address"`
}

//Area 地区
type Area struct {
	Value  string `gorm:"column:Value";primary_key` //行政编码
	Label  string `gorm:"column:Label"`             //对应的名称
	Parent string `gorm:"column:Parent"`            //父级归属
}

//table: uosidincrement
//UOSIDINCREMENT 生成user uosid的id
type UOSIDINCREMENT struct {
	ID int `gorm:"AUTO_INCREMENT"`
}
