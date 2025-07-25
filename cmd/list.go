package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"jvm/ui"
)

// ListInstalledJDKs mostra tutte le versioni JDK installate localmente
func ListInstalledJDKs() {
	ui.ShowBanner()
	fmt.Println("📦 JDK INSTALLATI LOCALMENTE")
	fmt.Println()

	// Ottieni directory home dell'utente
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("[ERROR] Errore nell'ottenere la directory home: %v\n", err)
		return
	}

	versionsDir := filepath.Join(homeDir, ".jvm", "versions")

	// Controlla se la directory esiste
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		fmt.Println("📂 Nessuna installazione JDK trovata")
		fmt.Printf("[INFO] La directory %s non esiste ancora\n", versionsDir)
		fmt.Println("   Usa 'jvm download <version>' per scaricare una versione")
		return
	}

	// Leggi il contenuto della directory
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		fmt.Printf("[ERROR] Errore nella lettura della directory: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("📂 Nessuna installazione JDK trovata")
		fmt.Printf("[INFO] La directory %s è vuota\n", versionsDir)
		fmt.Println("   Usa 'jvm download <version>' per scaricare una versione")
		return
	}

	// Raccogli informazioni sui JDK installati
	var jdks []JDKInstallation
	for _, entry := range entries {
		if entry.IsDir() {
			jdkPath := filepath.Join(versionsDir, entry.Name())
			installation := analyzeJDKInstallation(entry.Name(), jdkPath)
			jdks = append(jdks, installation)
		}
	}

	if len(jdks) == 0 {
		fmt.Println("📂 Nessuna installazione JDK valida trovata")
		return
	}

	// Ordina per versione (più recenti prima)
	sort.Slice(jdks, func(i, j int) bool {
		return compareVersions(jdks[i].Version, jdks[j].Version) > 0
	})

	// Mostra le installazioni in formato tabella
	displayJDKTable(jdks)
}

// JDKInstallation rappresenta un'installazione JDK locale
type JDKInstallation struct {
	Version     string
	Provider    string
	Path        string
	Size        string
	InstallDate string
	IsExtracted bool
	ArchiveType string
}

// analyzeJDKInstallation analizza una directory JDK per estrarre informazioni
func analyzeJDKInstallation(dirName, jdkPath string) JDKInstallation {
	installation := JDKInstallation{
		Version: dirName,
		Path:    jdkPath,
	}

	// Prova a identificare il provider dal nome
	installation.Provider = identifyProvider(dirName)

	// Calcola dimensione directory
	size := calculateDirSize(jdkPath)
	installation.Size = formatSize(size)

	// Ottieni data di installazione (modificata della directory)
	if stat, err := os.Stat(jdkPath); err == nil {
		installation.InstallDate = stat.ModTime().Format("2006-01-02 15:04")
	}

	// Controlla se contiene file estratti o archivi
	installation.IsExtracted, installation.ArchiveType = checkExtractionStatus(jdkPath)

	return installation
}

// identifyProvider cerca di identificare il provider dal nome della directory
func identifyProvider(dirName string) string {
	lower := strings.ToLower(dirName)

	if strings.Contains(lower, "adoptium") || strings.Contains(lower, "temurin") {
		return "Adoptium"
	}
	if strings.Contains(lower, "azul") || strings.Contains(lower, "zulu") {
		return "Azul"
	}
	if strings.Contains(lower, "liberica") || strings.Contains(lower, "bellsoft") {
		return "Liberica"
	}
	if strings.Contains(lower, "oracle") {
		return "Oracle"
	}

	return "Sconosciuto"
}

// calculateDirSize calcola la dimensione totale di una directory
func calculateDirSize(dirPath string) int64 {
	var size int64

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				size += info.Size()
			}
		}
		return nil
	})

	if err != nil {
		return 0
	}

	return size
}

// formatSize formatta una dimensione in byte in una stringa leggibile
func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", size, units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}

// checkExtractionStatus controlla se la directory contiene file estratti o solo archivi
func checkExtractionStatus(jdkPath string) (bool, string) {
	entries, err := os.ReadDir(jdkPath)
	if err != nil {
		return false, ""
	}

	hasExtracted := false
	archiveType := ""

	for _, entry := range entries {
		name := strings.ToLower(entry.Name())

		// Cerca archivi
		if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") {
			archiveType = "tar.gz"
		} else if strings.HasSuffix(name, ".zip") {
			archiveType = "zip"
		} else if strings.HasSuffix(name, ".msi") {
			archiveType = "msi"
		} else if strings.HasSuffix(name, ".exe") {
			archiveType = "exe"
		}

		// Cerca directory tipiche di JDK estratto
		if entry.IsDir() && (name == "bin" || name == "lib" || name == "include") {
			hasExtracted = true
		}

		// Cerca file tipici di JDK estratto
		if !entry.IsDir() && (name == "release" || name == "version") {
			hasExtracted = true
		}
	}

	return hasExtracted, archiveType
}

// compareVersions confronta due stringhe di versione (ritorna: 1 se a > b, -1 se a < b, 0 se uguali)
func compareVersions(a, b string) int {
	// Estrai il numero di versione principale dalle stringhe
	versionA := extractMainVersion(a)
	versionB := extractMainVersion(b)

	if versionA != versionB {
		if versionA > versionB {
			return 1
		}
		return -1
	}

	// Se le versioni principali sono uguali, confronta lexicograficamente
	if a > b {
		return 1
	} else if a < b {
		return -1
	}
	return 0
}

// extractMainVersion estrae il numero di versione principale (es. "17" da "jdk-17.0.2")
func extractMainVersion(version string) int {
	// Cerca pattern comuni: jdk-XX, java-XX, XX, etc.
	parts := strings.FieldsFunc(version, func(c rune) bool {
		return c == '-' || c == '_' || c == '.'
	})

	for _, part := range parts {
		if num, err := strconv.Atoi(part); err == nil && num > 0 && num < 100 {
			return num
		}
	}

	return 0
}

// displayJDKTable mostra i JDK in formato tabella
func displayJDKTable(jdks []JDKInstallation) {
	fmt.Printf("📋 Trovate %d installazioni JDK:\n\n", len(jdks))

	// Header della tabella
	fmt.Printf("%-25s %-12s %-8s %-16s %-10s %s\n",
		"VERSIONE", "PROVIDER", "STATO", "DATA INSTALL", "DIMENSIONE", "PERCORSO")
	fmt.Println(strings.Repeat("-", 90))

	// Righe della tabella
	for _, jdk := range jdks {
		status := getStatusIcon(jdk.IsExtracted, jdk.ArchiveType)

		// Tronca il percorso se troppo lungo
		displayPath := jdk.Path
		if len(displayPath) > 35 {
			displayPath = "..." + displayPath[len(displayPath)-32:]
		}

		fmt.Printf("%-25s %-12s %-8s %-16s %-10s %s\n",
			jdk.Version,
			jdk.Provider,
			status,
			jdk.InstallDate,
			jdk.Size,
			displayPath)
	}

	fmt.Println()
	fmt.Println("📌 Legenda stato:")
	fmt.Println("   [READY] PRONTO   - JDK estratto e pronto per l'uso")
	fmt.Println("   📦 ARCHIVIO - Solo archivio scaricato (richiede estrazione)")
	fmt.Println("   ❓ VUOTO    - Directory vuota o danneggiata")
	fmt.Println()
	fmt.Println("[INFO] Prossimi comandi disponibili:")
	fmt.Println("   jvm extract <version>  - Estrai un archivio JDK")
	fmt.Println("   jvm use <version>      - Imposta come JDK attivo")
	fmt.Println("   jvm remove <version>   - Rimuovi una versione")
}

// getStatusIcon restituisce l'icona di stato appropriata
func getStatusIcon(isExtracted bool, archiveType string) string {
	if isExtracted {
		return "[READY] PRONTO"
	} else if archiveType != "" {
		return "📦 ARCHIVIO"
	}
	return "❓ VUOTO"
}
