package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Preferences struct {
	ScpiPort int `json:"scpiPort"`
	ScpiAddress string `json:"scpiAddress"`
}

func getPreferencesPath() (string, error) {
	path := config.PreferencesFilePath

	prefsDir := filepath.Dir(path)
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create preferences directory: %w", err)
	}

	return path, nil
}

func loadPreferences() (*Preferences, error) {
	path, err := getPreferencesPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read preferences file: %w", err)
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("failed to parse preferences file: %w", err)
	}

  if prefs.ScpiPort == 0 {
      prefs.ScpiPort = config.DefaultScpiSocketPort
  }
  if prefs.ScpiAddress == "" {
      prefs.ScpiAddress = config.DefaultScpiSocketAddress
  }

	return &prefs, nil
}

func (p *Preferences) save() error {
	path, err := getPreferencesPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences file: %w", err)
	}

	log.Printf("Saved preferences: %+v", p)
	return nil
}

func (p *Preferences) delete() error {
	path, err := getPreferencesPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete preferences file: %w", err)
	}

	p.ScpiPort = config.DefaultScpiSocketPort
	p.ScpiAddress = config.DefaultScpiSocketAddress

	log.Println("Deleted preferences file and preset preferences")
	return nil
}
