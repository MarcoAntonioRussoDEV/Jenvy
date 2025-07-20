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
		fmt.Println("❌ Impossibile accedere alla home:", err)
		return
	}

	path := filepath.Join(home, ".jvm", "config.json")
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("⚠️ Configurazione non trovata:", err)
		return
	}
	defer file.Close()

	var cfg map[string]string
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		fmt.Println("❌ Errore nel parsing del file:", err)
		return
	}

	fmt.Println("📦 Configurazione attuale:")
	for k, v := range cfg {
		if v == "" {
			v = "(vuoto)"
		}
		fmt.Printf(" - %s: %s\n", k, v)
	}
}
