package cmd

import (
	"fmt"

	"jenvy/internal/ui"
	"jenvy/internal/utils"
)

// Version information
const (
	Version   = "1.0.0"
	BuildDate = "2025-01-27"
	GitCommit = "main"
)

// ShowVersionWithInfo displays version information with custom build data
func ShowVersionWithInfo(version, buildDate, gitCommit string) {
	ui.ShowBanner()
	fmt.Printf("%s %s\n", utils.ColorText("Jenvy", utils.BrightCyan), utils.ColorText("v"+version, utils.BrightGreen))
	fmt.Printf("%s %s\n", utils.ColorText("Build Date:", utils.BrightYellow), buildDate)
	fmt.Printf("%s %s\n", utils.ColorText("Git Commit:", utils.BrightYellow), gitCommit)
	fmt.Printf("%s %s\n", utils.ColorText("License:", utils.BrightYellow), "MIT")
	fmt.Printf("%s %s\n", utils.ColorText("Repository:", utils.BrightYellow), "https://github.com/MarcoAntonioRussoDEV/Jenvy")
	fmt.Println("")
	fmt.Printf("%s Developer Kit Manager for Windows\n", utils.ColorText("🚀", utils.BrightGreen))
	fmt.Printf("%s Multi-provider OpenJDK management solution\n", utils.ColorText("☕", utils.BrightYellow))
	fmt.Printf("%s Support: GitHub Sponsors, Ko-fi, PayPal\n", utils.ColorText("💖", utils.BrightMagenta))
}

func ShowVersion() {
	ShowVersionWithInfo(Version, BuildDate, GitCommit)
}

func ShowHelp() {
	ui.ShowBanner()
	fmt.Println("Jenvy - Developer Kit Manager helps you explore available OpenJDK releases across providers.")
	fmt.Println("It selects one recommended version per major tag (e.g., 8, 11, 17...) using the following priority:")
	fmt.Println(" " + utils.ColorText("[LTS]", utils.BrightGreen) + " LTS availability (Long-Term Support)")
	fmt.Println(" " + utils.ColorText("[STATS]", utils.BrightYellow) + " Most-used or popular release")
	fmt.Println(" " + utils.ColorText("[LATEST]", utils.BrightCyan) + " Latest patch version")
	fmt.Println("")

	fmt.Println(utils.SectionText("[COMMANDS] AVAILABLE COMMANDS:"))
	fmt.Println("─────────────────────")
	fmt.Println("  jenvy remote-list (rl)                   # Show recommended versions (default: Adoptium)")
	fmt.Println("  jenvy remote-list --provider=azul        # Specify provider (adoptium|azul|liberica|private)")
	fmt.Println("  jenvy remote-list --all                  # Show versions from all providers")
	fmt.Println("  jenvy remote-list --latest               # Show only the latest version")
	fmt.Println("  jenvy remote-list --major-only           # Show only major releases (e.g. 17.0.0)")
	fmt.Println("  jenvy remote-list --jdk=17               # Filter only a specific version")
	fmt.Println("  jenvy remote-list --lts-only             # Show only LTS versions")
	fmt.Println("")
	fmt.Println(utils.SectionText("[DOWNLOAD] JDK DOWNLOAD:"))
	fmt.Println("────────────────")
	fmt.Println("  jenvy download (dl) <version>            # Download JDK version to ~/.jenvy/versions")
	fmt.Println("  jenvy download 17 --provider=adoptium    # Download from specific provider")
	fmt.Println("  jenvy download 21 --output=./my-jdks     # Download to custom directory")
	fmt.Println("")
	fmt.Println(utils.SectionText("[EXTRACT] JDK EXTRACTION:"))
	fmt.Println("──────────────────")
	fmt.Println("  jenvy extract (ex)                       # List available archives to extract")
	fmt.Println("  jenvy extract 17                          # Extract any JDK 17.x.y version")
	fmt.Println("  jenvy extract 21                          # Extract any JDK 21.x.y version")
	fmt.Println("  jenvy extract JDK-17.0.16+8              # Extract specific JDK version")
	fmt.Println("")
	fmt.Println(utils.SectionText("[MANAGE] JDK MANAGEMENT:"))
	fmt.Println("─────────────────")
	fmt.Println("  jenvy list (l)                           # Show installed JDK versions")
	fmt.Println("  jenvy use (u) <version>                  # Set JDK version as active (JAVA_HOME)")
	fmt.Println("  jenvy remove (rm) <version>              # Remove installed JDK version")
	fmt.Println("  jenvy remove (rm) --all                  # Remove ALL JDK installations")
	fmt.Println("")
	fmt.Println(utils.SectionText("[SHELL] SHELL COMPLETION:"))
	fmt.Println("──────────────────")
	fmt.Println("  jenvy completion                         # Generate bash completion script")
	fmt.Println("  jenvy completion install                 # Install completion to ~/.bashrc")
	fmt.Println("")
	fmt.Println(utils.SectionText("[TOOLS] SYSTEM TOOLS:"))
	fmt.Println("───────────────")
	fmt.Println("  jenvy fix-path (fp)                      # Remove duplicate PATH entries")
	fmt.Println("  jenvy init                               # Initialize Jenvy environment variables")
	fmt.Println("")
	fmt.Println(utils.SectionText("[PRIVATE] PRIVATE REPOSITORY CONFIGURATION:"))
	fmt.Println("───────────────────────────────────")
	fmt.Println("  jenvy configure-private (cp) <endpoint> [token]  # Configure enterprise repository")
	fmt.Println("  jenvy config-show (cs)                           # Show current configuration")
	fmt.Println("  jenvy config-reset (cr)                          # Remove private configuration")
	fmt.Println("")
	fmt.Println(utils.SectionText("[HELP] HELP & VERSION:"))
	fmt.Println("────────────────")
	fmt.Println("  jenvy --help, -h, help                   # Show this help message")
	fmt.Println("  jenvy --version, -v, version             # Show version information")
	fmt.Println("")
	fmt.Println(utils.ExamplesText("PRACTICAL EXAMPLES:"))
	fmt.Println("──────────────────────")
	fmt.Println("  jenvy rl --provider=azul --jdk=21")
	fmt.Println("  jenvy remote-list --all --lts-only")
	fmt.Println("  jenvy dl 17 && jenvy ex downloaded-archive.zip")
	fmt.Println("  jenvy extract adoptium-jdk-21.zip")
	fmt.Println("  jenvy cp https://nexus.company.com/api/jdk token123")
	fmt.Println("  jenvy cs")
	fmt.Println("  jenvy completion install                 # Enable tab completion")
	fmt.Println("  jenvy dl 17                              # Then use: jenvy <Tab> to see commands")
}
