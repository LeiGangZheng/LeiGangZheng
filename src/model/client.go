package model

import "time"

/*
database:  deepinid
dsn = "root:123456@tcp(127.0.0.1:3306)/deepinid?parseTime=true&loc=Local&charset=utf8mb4,utf8"      # 地址
*/

type Scope struct {
	ID     int    `gorm:"type:int(11);not null;primary_key;auto_increment" json:"id"`
	Name   string `gorm:"type:varchar(191);not null;default:'';unique" json:"name"`
	Depend string `gorm:"type:varchar(255);not null;default:''" json:"-"`
}

type ScopeList []Scope

// Client stores all data in struct variables
type Client struct {
	ID             string    `gorm:"type:varchar(191);primary_key;not null" json:"client_id"`
	UserID         int       `gorm:"type:int(11);not null;default:'0'" json:"user_id"`
	AppName        string    `gorm:"type:varchar(255);not null;default:'';unique" json:"app_name"`
	Secret         string    `gorm:"type:varchar(255);not null;default:''" json:"client_secret"`
	RedirectURI    string    `gorm:"type:varchar(255);not null;default:''" json:"redirect_uri"`
	RevokeURI      string    `gorm:"type:varchar(255);not null;default:''" json:"revoke_uri"`
	AppHomeURI     string    `gorm:"type:varchar(255);not null;default:''" json:"app_home_uri"`
	AppDescription string    `gorm:"type:varchar(255);not null;default:''" json:"app_description"`
	AccessType     string    `gorm:"type:varchar(255);not null;default:''" json:"access_type"`
	Scope          string    `sql:"-" json:"scope"`
	Scopes         ScopeList `gorm:"many2many:client_scope" json:"-"`
}

// Authorize data
type Authorize struct {
	ID          int    `gorm:"type:int(11);primary_key;auto_increment"`
	Code        string `gorm:"type:varchar(255);not null;default:''"`
	UserID      int    `gorm:"type:int(11);not null;default:'0'"`
	ClientID    string `gorm:"type:varchar(255);not null;default:''"`
	ExpiresIn   int32  `gorm:"type:varchar(255);not null;default:'0'"`
	Scope       string `gorm:"type:varchar(255);not null;default:''"`
	RedirectURI string `gorm:"type:varchar(255);not null;default:''"`
	State       string `gorm:"type:varchar(255);not null;default:''"`
	CreatedAt   time.Time
}

const (
	ServiceNameQQ     = "qq"
	ServiceNameSina   = "sina"
	ServiceNameWechat = "wechat"
	ServiceNameGithub = "github"
)

// ThirdPartyLogin is third party login info
type ThirdPartyLogin struct {
	ID     string `gorm:"type:varchar(191);not null;default:''" json:"-"`
	Token  string `gorm:"type:varchar(255);not null;default:''" json:"-"`
	Type   string `gorm:"type:varchar(16);not null;default:''" json:"type"`
	UserID int    `gorm:"type:int(11);not null;default:'0'" json:"uid"` //user.ID
}

// client_scope
type ClientScope struct {
	ClientID string `gorm:"column:client_id;type:varchar(191);not null;default:''" json:"client_id"`
	ScopeID  int    `gorm:"column:scope_id;type:int(11);not null;default:'0'" json:"scope_id"` //user.ID
}
