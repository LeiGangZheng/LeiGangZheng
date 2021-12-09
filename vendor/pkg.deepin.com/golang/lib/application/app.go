package application

import (
	"fmt"
	"os"
	"path/filepath"
)

var vcsInfo string

var buildTime string

var appName string

var appID string

var appMode string

// todo
func init() {

}

// AppName gets application name.
func AppName() string {
	if appName == "" {
		appName = filepath.Base(os.Args[0])
	}
	return appName
}

// BuildTime gets building time.
func BuildTime() string {
	return buildTime
}

// VcsInfo gets vcs revision.
func VcsInfo() string {
	return vcsInfo
}

// AppMode gets app mode
func AppMode() string {
	if appMode == "" {
		appMode = EnvAppMode()
	}
	if appMode == "" {
		appMode = "local"
	}
	return appMode
}

// AppID gets app id
func AppID() string {
	if appID == "" || appID == "0" {
		appID = EnvAppID()
	}
	if appID == "" || appID == "0" {
		appID = "0"
	}
	return appID
}

// LogDir ...
func LogDir() string {
	// LogDir gets application log directory.
	logDir := EnvAppLogDir()
	if logDir == "" {
		logDir = "/home/www/logs/"
	}
	return fmt.Sprintf("%s%s/", logDir, AppName())
}
