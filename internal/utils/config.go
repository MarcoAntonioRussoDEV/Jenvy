package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
    PrivateEndpoint string `json:"private_endpoint"`
    PrivateToken    string `json:"private_token"`
}

func LoadConfig() (*Config, error) {
    home, _ := os.UserHomeDir()
    path := filepath.Join(home, ".jenvy", "config.json")
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var cfg Config
    err = json.NewDecoder(file).Decode(&cfg)
    return &cfg, err
}
