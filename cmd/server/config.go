package main

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port               int
	ScriptStoragePath  string
	HelpFilePath       string
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.sclipi")
	viper.AddConfigPath("/etc/sclipi")

	viper.SetDefault("port", 8080)
	viper.SetDefault("scriptStoragePath", "./scripts")
	viper.SetDefault("helpFilePath", "./SCPI.txt")

	viper.SetEnvPrefix("SCLIPI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	config := &Config{
		Port:              viper.GetInt("port"),
		ScriptStoragePath: viper.GetString("scriptStoragePath"),
		HelpFilePath:      viper.GetString("helpFilePath"),
	}

	return config, nil
}
