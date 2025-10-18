package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort            int
	DefaultScpiSocketPort int
	ScriptStoragePath     string
	HelpFilePath          string
	PreferencesFilePath   string
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.sclipi")
	viper.AddConfigPath("/etc/sclipi")

	viper.SetDefault("serverPort", 8080)
	viper.SetDefault("defaultScpiSocketPort", 5025)
	viper.SetDefault("scriptStoragePath", "$HOME/.sclipi/scripts")
	viper.SetDefault("helpFilePath", "TODO")
	viper.SetDefault("preferencesFilePath", "$HOME/.sclipi/preferences.json")

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
		ServerPort:            viper.GetInt("serverPort"),
		DefaultScpiSocketPort: viper.GetInt("defaultScpiSocketPort"),
		ScriptStoragePath:     os.ExpandEnv(viper.GetString("scriptStoragePath")),
		HelpFilePath:          os.ExpandEnv(viper.GetString("helpFilePath")),
		PreferencesFilePath:   os.ExpandEnv(viper.GetString("preferencesFilePath")),
	}

	log.Printf("Config: %+v", config)

	return config, nil
}
