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
}

func getPreferencesPath() (string, error) {
	path := config.PreferencesFilePath

	// Ensure directory exists
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
			return nil, nil // No preferences file is not an error
		}
		return nil, fmt.Errorf("failed to read preferences file: %w", err)
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return nil, fmt.Errorf("failed to parse preferences file: %w", err)
	}

	log.Printf("Loaded preferences: %+v", prefs)
	return &prefs, nil
}

func savePreferences(prefs *Preferences) error {
	path, err := getPreferencesPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write preferences file: %w", err)
	}

	log.Printf("Saved preferences: %+v", prefs)
	return nil
}

func deletePreferences() error {
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

	log.Println("Deleted preferences file")
	return nil
}
