package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort               int
	DefaultScpiSocketPort    int
	DefaultScpiSocketAddress string
	PreferencesFilePath      string
  ConnectionMode           string
}

func loadConfig() (*Config, error) {
	pflag.Int("server-port", 8080, "HTTP server port")
	pflag.Int("scpi-port", 5025, "Default SCPI socket port")
	pflag.String("scpi-address", "localhost", "Default SCPI socket address")
	pflag.String("preferences-file", "$HOME/.scpir/preferences.json", "Preferences file path")
	pflag.String("connection-mode", "per-client", "Connection mode (per-client or server-default)")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.scpir")
	viper.AddConfigPath("/etc/scpir")

	viper.SetDefault("serverPort", 8080)
	viper.SetDefault("defaultScpiSocketPort", 5025)
	viper.SetDefault("defaultScpiSocketAddress", "localhost")
	viper.SetDefault("preferencesFilePath", "$HOME/.scpir/preferences.json")
	viper.SetDefault("connectionMode", "per-client")

	viper.SetEnvPrefix("SCPIR")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	connectionMode := viper.GetString("connection-mode")
	defaultAddress := viper.GetString("scpi-address")
	if connectionMode == "per-client" && !pflag.CommandLine.Changed("scpi-address") {
		defaultAddress = ""
	}

	config := &Config{
		ServerPort:               viper.GetInt("server-port"),
		DefaultScpiSocketPort:    viper.GetInt("scpi-port"),
		DefaultScpiSocketAddress: defaultAddress,
		PreferencesFilePath:      os.ExpandEnv(viper.GetString("preferences-file")),
		ConnectionMode:           connectionMode,
	}

	log.Printf("Config: %+v", config)

	return config, nil
}
