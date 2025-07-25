package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResetPrivateConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("[ERROR] Error accessing user directory:", err)
		return
	}

	path := filepath.Join(homeDir, ".jvm", "config.json")
	err = os.Remove(path)
	if err != nil {
		fmt.Println("[WARN] Unable to delete file:", err)
		return
	}

	fmt.Println("[SUCCESS] Private configuration removed successfully.")
}
