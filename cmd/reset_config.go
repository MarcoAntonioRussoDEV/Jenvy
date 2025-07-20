package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func ResetConfigFile() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("❌ Errore nell'accesso alla directory utente:", err)
		return
	}

	path := filepath.Join(home, ".jvm", "config.json")
	err = os.Remove(path)
	if err != nil {
		fmt.Println("⚠️ Impossibile cancellare il file:", err)
		return
	}

	fmt.Println("✅ Configurazione privata rimossa con successo.")
}
