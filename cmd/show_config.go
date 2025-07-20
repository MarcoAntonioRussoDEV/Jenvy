package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ShowCurrentConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Unable to access home directory:", err)
		return
	}

	path := filepath.Join(home, ".jvm", "config.json")
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("⚠️ Configuration not found:", err)
		return
	}
	defer file.Close()

	var cfg map[string]string
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		fmt.Println("❌ File parsing error:", err)
		return
	}

	fmt.Println("📦 Current configuration:")
	for k, v := range cfg {
		if v == "" {
			v = "(empty)"
		}
		fmt.Printf(" - %s: %s\n", k, v)
	}
}
