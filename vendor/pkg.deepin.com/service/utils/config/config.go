package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	//"reflect"
)

// Config is map of app config
type Config struct {
	cfg map[string]interface{}
}

var _cfg *Config
var _path string

// GetConfig return global config
func GetConfig() *Config {
	if nil == _cfg {
		_path = "configs/server.json"
		_cfg, _ = NewConfig(_path)
	}
	return _cfg
}

var (
	// KeyServerHost is bind host
	KeyServerHost = "ServerHost"
	// KeyServerPort is bind prot
	KeyServerPort = "ServerPort"

	// KeyMysqlHost is mysql host
	KeyMysqlHost = "MysqlHost"
	// KeyMysqlPort is mysql prot
	KeyMysqlPort = "MysqlPort"
	// KeyMysqlUsername is mysql username
	KeyMysqlUsername = "MysqlUsername"
	// KeyMysqlPassword is mysql password
	KeyMysqlPassword = "MysqlPassword"
	// KeyMysqlDatabase is mysql database name
	KeyMysqlDatabase = "MysqlDatabase"

	// KeyAdminToken is admin toke
	KeyAdminToken = "AdminToken"
	// KeyAdminScope is admin scope
	KeyAdminScope = "AdminScope"
)

// Get return raw data of key
func (c *Config) Get(key string) interface{} {
	if nil == c.cfg[key] {
		panic("Get Invalid Key " + key)
	}
	return c.cfg[key]
}

// Read return string data of key
func Read(key string) string {
	return fmt.Sprint(GetConfig().Get(key))
}

// ReadInt return int data of key
func ReadInt(key string) int {
	i := GetConfig().Get(key).(float64)
	return int(i)
}

// ReadList return array data of key
func ReadList(key string) []interface{} {
	if list, ok := GetConfig().Get(key).([]interface{}); ok {
		return list
	}
	return [](interface{}){}
}

// NewConfig read config from conf file
func NewConfig(path string) (*Config, error) {
	c := &Config{
		cfg: map[string](interface{}){},
	}

	root := os.Getenv("SERVER_ROOT")
	if "" == root {
		root = "."
	}
	confpath := root + "/" + path
	content, err := ioutil.ReadFile(confpath)
	if nil != err {
		fmt.Println(confpath, err)
		panic("Can not found conf file")
	}

	if strings.HasSuffix(confpath, ".json") {
		err = json.Unmarshal(content, &c.cfg)
	}
	if strings.HasSuffix(confpath, ".yml") {
		err = yaml.Unmarshal(content, &c.cfg)
	}
	if strings.HasSuffix(confpath, ".toml") {
		err = toml.Unmarshal(content, &c.cfg)
	}
	c.cfg["Root"] = root
	return c, err
}

func Load(path string, v interface{}) error {
	root := os.Getenv("SERVER_ROOT")
	if "" == root {
		root = "."
	}
	confpath := root + "/" + path
	content, err := ioutil.ReadFile(confpath)
	if nil != err {
		panic("Can not found conf file")
	}

	if strings.HasSuffix(confpath, ".json") {
		err = json.Unmarshal(content, v)
	}
	if strings.HasSuffix(confpath, ".yml") {
		err = yaml.Unmarshal(content, v)
	}
	if strings.HasSuffix(confpath, ".toml") {
		err = toml.Unmarshal(content, v)
	}
	return err
}

// Dump config
func Dump(v interface{}) {
	fmt.Println("config begin:--------------------------------")
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
	fmt.Println("config end:--------------------------------")
}
