package main

import (
	"fmt"
	"os"

	"jvm/cmd"
	"jvm/ui"
	"jvm/utils"
)

func main() {
	if len(os.Args) < 2 {
		ui.ShowBanner()
		fmt.Println("Java Version Manager helps you explore available OpenJDK releases across providers.")
		fmt.Println("It selects one recommended version per major tag (e.g., 8, 11, 17...) using the following priority:")
		fmt.Println(" ✅ LTS availability (Long-Term Support)")
		fmt.Println(" 📈 Most-used or popular release")
		fmt.Println(" 🆕 Latest patch version\n")
		fmt.Println("❗ Missing subcommand. Use: jvm remote-list [--provider] [--all]")
		return
	}

	// 👇 Provider predefinito centralizzato
	provider := utils.DefaultProvider()

	switch os.Args[1] {
	case "remote-list":
		cmd.RemoteList(provider)

	case "configure-private":
		if len(os.Args) < 3 {
			fmt.Println("❗ Usa: jvm configure-private <endpoint> [token]")
			return
		}
		endpoint := os.Args[2]
		token := ""
		if len(os.Args) > 3 {
			token = os.Args[3]
		}
		cmd.ConfigurePrivateRepo(endpoint, token)

	case "show-config":
		cmd.ShowCurrentConfig()

	case "reset-config":
		cmd.ResetConfigFile()

	default:
		fmt.Printf("❌ Comando sconosciuto: %s\n", os.Args[1])
	}
}
