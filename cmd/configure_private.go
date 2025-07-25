package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ConfigurePrivateRepo(endpoint string, token string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("[ERROR] Unable to determine user directory:", err)
		return
	}

	dir := filepath.Join(home, ".jvm")
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, "config.json")
	cfg := map[string]string{
		"private_endpoint": endpoint,
		"private_token":    token,
	}

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("[ERROR] Write error:", err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(cfg)
	if err != nil {
		fmt.Println("[ERROR] JSON encoding error:", err)
		return
	}

	fmt.Println("[SUCCESS] Private repository configured successfully!")
	fmt.Println("📁 File:", path)
}
