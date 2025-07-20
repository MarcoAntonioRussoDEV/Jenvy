package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResetConfigFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Error accessing user directory:", err)
		return
	}

	path := filepath.Join(home, ".jvm", "config.json")
	err = os.Remove(path)
	if err != nil {
		fmt.Println("⚠️ Unable to delete file:", err)
		return
	}

	fmt.Println("✅ Private configuration removed successfully.")
}
