package main

import (
	"fmt"
	"os"

	"jvm/cmd"
	"jvm/utils"
)

func main() {
	if len(os.Args) < 2 {
		cmd.ShowHelp()
		return
	}

	// 👇 Provider predefinito centralizzato
	provider := utils.DefaultProvider()

	switch os.Args[1] {
	case "remote-list", "rl":
		cmd.RemoteList(provider)

	case "download", "dl":
		cmd.DownloadJDK(provider)

	case "configure-private", "cp":
		if len(os.Args) < 3 {
			fmt.Println("❗ Usage: jvm configure-private <endpoint> [token]")
			fmt.Println("❗ Short form: jvm cp <endpoint> [token]")
			return
		}
		endpoint := os.Args[2]
		token := ""
		if len(os.Args) > 3 {
			token = os.Args[3]
		}
		cmd.ConfigurePrivateRepo(endpoint, token)

	case "config-show", "cs":
		cmd.ShowCurrentConfig()

	case "config-reset", "cr":
		cmd.ResetConfigFile()

	case "--help", "-h", "help":
		cmd.ShowHelp()

	default:
		fmt.Printf("❌ Unknown command: %s\n", os.Args[1])
		fmt.Println("💡 Use 'jvm --help' to see all available commands")
	}
}
