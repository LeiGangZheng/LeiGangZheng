package model

import (
	"database/sql"
	"time"

	"github.com/jinzhu/gorm"
)

/*
database: deepin_user
dsn = "root:123456@tcp(127.0.0.1:3306)/deepin_user?parseTime=true&loc=Local&charset=utf8mb4,utf8"      # 地址
*/

type LoginInfo struct {
	IP   string    `json:"login_ip" form:"login_ip" gorm:"type:varchar(255);not null;default:''"`
	Time time.Time `json:"login_time" form:"login_time"`
}

type UserIndex struct {
	SQLEmail       sql.NullString `json:"-" gorm:"column:email;type:varchar(191);unique;"`
	SQLPhoneNumber sql.NullString `json:"-" gorm:"column:phone_number;unique;type:varchar(32)"`
}
type UserLoginInfo struct {
	gorm.Model `json:"-"`
	UserID     int `json:"uid"`

	LoginInfo
}

type LoginInfoList []*UserLoginInfo

// User for core deepinid data
type User struct {
	UserIndex
	ID            int           `form:"uid" json:"uid" gorm:"primary_key;auto_increment"`
	Username      string        `form:"username" json:"username" gorm:"type:varchar(191);not null;default:'';index"`
	Nickname      string        `form:"nickname" json:"nickname" gorm:"type:varchar(255);not null;default:''"`
	Region        string        `form:"region" json:"region" gorm:"type:varchar(128);not null;default:''"`
	Scope         string        `json:"scope" form:"scope" gorm:"-"`
	ProfileImage  string        `json:"profile_image" form:"profile_image" gorm:"type:varchar(255);not null;default:''"`
	Active        bool          `json:"active" form:"active" gorm:"type:tinyint;not null;default:'0'"`
	Disabled      bool          `json:"disabled" form:"disabled" gorm:"type:tinyint;not null;default:'0'"`
	Switched      bool          `json:"switched" form:"switched" gorm:"type:tinyint;not null;default:'0'"`
	RegisterIP    string        `json:"reg_ip" form:"reg_ip" gorm:"type:varchar(255);not null;default:''"`
	LoginIP       string        `json:"-" form:"login_ip" gorm:"type:varchar(255);not null;default:''"`
	LoginTime     time.Time     `json:"-" form:"login_time"`
	Scopes        ScopeList     `gorm:"many2many:user_scope;" json:"-"`
	LoginInfoList LoginInfoList `json:"login_info_list"`
	Password      string        `form:"password" json:"password" gorm:"type:varchar(255);not null;default:''"`
	PasswordEmpty bool          `form:"password_empty" json:"password_empty" gorm:"type:tinyint;not null;default:'0'"`
	CreatedAt     time.Time     `json:"-"`
	UpdatedAt     time.Time     `json:"-"`
	DeletedAt     *time.Time    `json:"-"`
}

// xxx
type UserProfile struct {
	UserID         int       `json:"upid" gorm:"primary_key;not null"`
	RealName       string    `json:"real_name" gorm:"type:varchar(255);not null;default:''"`
	GenderID       int       `json:"gender" gorm:"not null;default:'0'"`
	Birthday       time.Time `json:"birthday" gorm:"type:DATETIME"`
	MobilePhone    string    `json:"mobile_phone" gorm:"type:varchar(255);not null;default:''"`
	Website        string    `json:"website" gorm:"type:varchar(255);not null;default:''"`
	Signature      string    `json:"signature" gorm:"type:varchar(255);not null;default:''"`
	MailingAddress string    `json:"mailing_address" gorm:"type:varchar(255);not null;default:''"`

	Volunteer                 int    `json:"volunteer" gorm:"not null;default:'0'"`
	VolunteerProvince         string `json:"volunteer_province" gorm:"type:varchar(255);not null;default:''"`
	VolunteerCity             string `json:"volunteer_city" gorm:"type:varchar(255);not null;default:''"`
	ValueVolunteerSkills      []int  `json:"volunteer_skills" sqlname:"VolunteerSkills" gorm:"-"`
	VolunteerSkills           string `json:"-" gorm:"type:varchar(255);not null;default:''"`
	ValueVolunteerServiceTime []int  `json:"volunteer_service_time" sqlname:"VolunteerServiceTime" gorm:"-"`
	VolunteerServiceTime      string `json:"-" gorm:"type:varchar(255);not null;default:''"`

	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
}

// 日志记录表
type UserErrlog struct {
	SID   string `gorm:"column:src_id;primary_key"` //源数据的主键ID
	SType int    `gorm:"column:src_type"`           //源数据类型
	DID   string ``
	DType int    `gorm:"column:dst_type"` //数据表类型 1:user, 2:platform, 3: client,           --写哪个数据表
	DData string `gorm:"column:dst_data"` //供帐号系统合并使用 chinauos帐号系统，目标数据库的  --目标数据的关键字段
}
