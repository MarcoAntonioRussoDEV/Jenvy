package main

import (
	"fmt"
	"jvm/cmd"
	"jvm/ui"
	"os"
)

func main() {
    if len(os.Args) < 2 {
       if len(os.Args) < 2 {
    ui.ShowBanner()
    fmt.Println("Java Version Manager helps you explore available OpenJDK releases across providers.")
    fmt.Println("It selects one recommended version per major tag (e.g., 8, 11, 17...) using the following priority:")
    fmt.Println(" ✅ LTS availability (Long-Term Support)")
    fmt.Println(" 📈 Most-used or popular release")
    fmt.Println(" 🆕 Latest patch version")
    fmt.Println()
    fmt.Println("❗ Missing subcommand. Use: jvm remote-list [--provider] [--all]")
    return
}

        os.Exit(1)
    }

    switch os.Args[1] {
    case "remote-list":
        cmd.RemoteList()
    default:
        fmt.Printf("❌ Comando sconosciuto: %s\n", os.Args[1])
    }
}
