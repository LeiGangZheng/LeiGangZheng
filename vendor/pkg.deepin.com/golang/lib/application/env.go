package application

import (
	"os"
)

// EnvAppMode gets UNIONTECH_APP_MODE.
func EnvAppMode() string {
	return os.Getenv("UNIONTECH_APP_MODE")
}

// EnvAppID gets UNIONTECH_APP_ID.
func EnvAppID() string {
	return os.Getenv("UNIONTECH_APP_ID")
}

// EnvAppLogDir gets UNIONTECH_APP_LOG_DIR
func EnvAppLogDir() string {
	return os.Getenv("UNIONTECH_APP_LOG_DIR")
}
