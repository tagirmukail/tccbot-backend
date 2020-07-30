package config

import "github.com/spf13/viper"

// initDBPath inialized db path
func initDBPath() string {
	return viper.GetString("db_path")
}
