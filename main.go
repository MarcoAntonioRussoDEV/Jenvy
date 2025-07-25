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

	// Provider predefinito centralizzato
	provider := utils.DefaultProvider()

	switch os.Args[1] {
	case "remote-list", "rl":
		cmd.RemoteList(provider)

	case "download", "dl":
		cmd.DownloadJDK(provider)

	case "extract", "ex":
		cmd.ExtractJDK()

	case "list", "l":
		cmd.ListInstalledJDKs()

	case "use", "u":
		cmd.UseJDK()

	case "remove", "rm":
		cmd.RemoveJDK()

	case "completion":
		if len(os.Args) > 2 {
			switch os.Args[2] {
			case "install", "--install-all":
				cmd.InstallCompletionForAllShells()
			case "bash":
				cmd.GenerateCompletion()
			case "powershell":
				fmt.Print(cmd.GeneratePowerShellCompletion())
			case "cmd":
				fmt.Print(cmd.GenerateCmdCompletion())
			default:
				fmt.Println("Usage: jvm completion [install|bash|powershell|cmd]")
				fmt.Println("  install     - Install completion for all available shells")
				fmt.Println("  bash        - Generate Bash completion script")
				fmt.Println("  powershell  - Generate PowerShell completion script")
				fmt.Println("  cmd         - Generate CMD completion script")
			}
		} else {
			cmd.GenerateCompletion() // Default: Bash
		}

	case "fix-path", "fp":
		cmd.FixPath()

	case "init":
		cmd.InitializeJVMEnvironment()

	case "configure-private", "cp":
		if len(os.Args) < 3 {
			utils.PrintUsage("Usage: jvm configure-private <endpoint> [token]")
			utils.PrintUsage("Short form: jvm cp <endpoint> [token]")
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
		cmd.ResetPrivateConfig()

	case "--help", "-h", "help":
		cmd.ShowHelp()

	default:
		utils.PrintError(fmt.Sprintf("Unknown command: %s", os.Args[1]))
		utils.PrintInfo("Use 'jvm --help' to see all available commands")
	}
}
