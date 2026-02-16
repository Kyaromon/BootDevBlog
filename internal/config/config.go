package config

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type Config struct {
    DatabaseURL     string `json:"db_url"`
    CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
    var config Config
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return config, err
    }
    configPath := filepath.Join(homeDir, ".gatorconfig.json")

    file, err := os.Open(configPath)
    if err != nil {
        return config, err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    return config, err
}

func (c *Config) SetUser(username string) error {
    c.CurrentUserName = username
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }
    configPath := filepath.Join(homeDir, ".gatorconfig.json")

    file, err := os.Create(configPath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(c)
}
