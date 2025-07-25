package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// FixPath rimuove le voci duplicate dal PATH di sistema
func FixPath() {
	if runtime.GOOS != "windows" {
		fmt.Println("⚠️ PATH fix is currently only supported on Windows")
		return
	}

	fmt.Println("🔧 JVM PATH REPAIR UTILITY")
	fmt.Println("==========================")
	fmt.Println()

	// Leggi il PATH attuale usando cmd.exe che funziona anche con bash
	cmdPath := "C:\\Windows\\System32\\cmd.exe"
	cmd := exec.Command(cmdPath, "/c", "reg query \"HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment\" /v PATH")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[ERROR] Error reading system PATH: %v\n", err)
		return
	}

	// Parse dell'output di reg.exe
	lines := strings.Split(string(output), "\n")
	var currentPath string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "PATH") && strings.Contains(line, "REG_") {
			parts := strings.Split(line, "REG_EXPAND_SZ")
			if len(parts) >= 2 {
				currentPath = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	if currentPath == "" {
		fmt.Println("[ERROR] Current PATH is empty or not found")
		return
	}

	fmt.Printf("📋 Current SYSTEM PATH entries: %d\n", len(strings.Split(currentPath, ";")))

	// Dividi il PATH in voci separate
	pathEntries := strings.Split(currentPath, ";")
	var cleanEntries []string
	seen := make(map[string]bool)
	duplicatesRemoved := 0
	emptyEntriesRemoved := 0

	for _, entry := range pathEntries {
		trimmedEntry := strings.TrimSpace(entry)

		// Rimuovi voci vuote
		if trimmedEntry == "" {
			emptyEntriesRemoved++
			continue
		}

		// Controlla duplicati (case-insensitive)
		upperEntry := strings.ToUpper(trimmedEntry)
		if seen[upperEntry] {
			duplicatesRemoved++
			fmt.Printf("[CLEAN] Removing duplicate: %s\n", trimmedEntry)
			continue
		}

		seen[upperEntry] = true
		cleanEntries = append(cleanEntries, trimmedEntry)
	}

	if duplicatesRemoved == 0 && emptyEntriesRemoved == 0 {
		fmt.Println("[SUCCESS] PATH is already clean, no duplicates found")
		return
	}

	// Ricostruisci il PATH pulito
	cleanPath := strings.Join(cleanEntries, ";")

	fmt.Printf("\n[SUMMARY] CLEANUP SUMMARY:\n")
	fmt.Printf("   • Duplicate entries removed: %d\n", duplicatesRemoved)
	fmt.Printf("   • Empty entries removed: %d\n", emptyEntriesRemoved)
	fmt.Printf("   • Final PATH entries: %d\n", len(cleanEntries))
	fmt.Printf("\n[UPDATE] Updating SYSTEM PATH in registry...\n")

	// Aggiorna il registro di sistema (richiede privilegi amministratore)
	cmd = exec.Command(cmdPath, "/c",
		fmt.Sprintf(`reg add "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v PATH /t REG_EXPAND_SZ /d "%s" /f`, cleanPath))
	err = cmd.Run()
	if err != nil {
		fmt.Printf("[ERROR] Error updating SYSTEM PATH: %v\n", err)
		fmt.Printf("[INFO] TIP: You may need to run as Administrator\n")
		return
	}

	fmt.Println("[SUCCESS] SYSTEM PATH cleaned successfully!")
	fmt.Println()
	fmt.Println("[INFO] IMPORTANT: Restart your terminal or VS Code to see the changes")
	fmt.Println("   Or run: refreshenv (if you have Chocolatey installed)")
	fmt.Println()
	fmt.Printf("🔍 You can verify the changes by running: echo $PATH\n")
}
