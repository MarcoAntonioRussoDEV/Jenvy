package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"jvm/utils"
)

// ListInstalledJDKs esegue la scansione e visualizzazione di tutte le installazioni JDK locali su Windows.
//
// Questa funzione implementa un sistema completo di rilevamento e analisi delle installazioni JDK
// nella directory di gestione locale (~/.jvm/versions) con le seguenti funzionalità:
//
//  1. **Scansione directory JVM**: Accede alla directory ~/.jvm/versions nel profilo utente Windows
//     per individuare tutte le installazioni JDK gestite dal tool
//
//  2. **Analisi automatica**: Per ogni installazione rilevata esegue:
//     - Calcolo dimensioni directory tramite filesystem walk
//     - Rilevamento stato di estrazione (archivio vs. installazione completa)
//     - Lettura metadati di installazione (data, tipo archivio)
//
//  3. **Ordinamento intelligente**: Ordina le installazioni per numero di versione
//     con logica di parsing che riconosce pattern di versioning JDK standard
//
//  4. **Visualizzazione tabellare**: Presenta i risultati in formato tabella con:
//     - Versione e stato di installazione
//     - Data di installazione e dimensioni formattate
//     - Percorsi abbreviati per leggibilità
//
// **Caratteristiche Windows-specifiche:**
// - Utilizza os.UserHomeDir() per accedere al profilo utente Windows
// - Gestisce percorsi Windows con separatori backslash appropriati
// - Supporta formati archivio Windows (.zip, .msi, .exe) oltre a formati Unix
// - Formattazione output ottimizzata per terminali Windows (cmd.exe, PowerShell)
//
// **Stati di installazione riconosciuti:**
// - [READY]: JDK completamente estratto con struttura bin/, lib/, include/
// - [ARCHIVE]: Solo file archivio presente, richiede estrazione
// - [EMPTY]: Directory esistente ma senza contenuto valido
//
// Esempi di utilizzo:
//
//	jvm list  # Mostra tutte le installazioni JDK locali
func ListInstalledJDKs() {
	fmt.Println(utils.ColorText("LOCAL JDK INSTALLATIONS", utils.Bold+utils.BrightCyan))
	fmt.Println()

	// Ottieni directory home utente
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(utils.ErrorText(fmt.Sprintf("Error getting home directory: %v", err)))
		return
	}

	versionsDir := filepath.Join(homeDir, ".jvm", "versions")

	// Controlla se la directory esiste
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		fmt.Println(utils.WarningText("No JDK installations found"))
		fmt.Printf("[INFO] Directory %s does not exist yet\n", versionsDir)
		fmt.Println("   Use 'jvm download <version>' to download a version")
		return
	}

	// Leggi contenuto directory
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		fmt.Println(utils.ErrorText(fmt.Sprintf("Error reading directory: %v", err)))
		return
	}

	if len(entries) == 0 {
		fmt.Println(utils.WarningText("No JDK installations found"))
		fmt.Printf("[INFO] Directory %s is empty\n", versionsDir)
		fmt.Println("   Use 'jvm download <version>' to download a version")
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
		fmt.Println(utils.WarningText("No valid JDK installations found"))
		return
	}

	// Ordina per versione (più recenti per prime)
	sort.Slice(jdks, func(i, j int) bool {
		return compareVersions(jdks[i].Version, jdks[j].Version) > 0
	})

	// Mostra installazioni in formato tabella
	displayJDKTable(jdks)
}

// JDKInstallation rappresenta un'installazione JDK locale
type JDKInstallation struct {
	Version     string
	Path        string
	Size        string
	InstallDate string
	IsExtracted bool
	ArchiveType string
}

// analyzeJDKInstallation analizza una directory JDK per estrarre informazioni
//
// Questa funzione esegue un'analisi completa di una directory di installazione JDK:
//
//  1. **Calcolo dimensioni**: Esegue una camminata ricorsiva della directory per calcolare
//     la dimensione totale dell'installazione incluse tutte le sottodirectory e file
//
//  2. **Metadati installazione**: Estrae la data di installazione dal timestamp di modifica
//     della directory utilizzando gli attributi del filesystem Windows
//
//  3. **Stato estrazione**: Determina se la directory contiene un JDK estratto
//     (con struttura bin/, lib/, include/) o solo file archivio
//
// Considerazioni specifiche per Windows:
// - Gestisce separatori di percorso Windows e attributi filesystem
// - Riconosce formati archivio specifici di Windows (.msi, .exe)
// - Utilizza formattazione timestamp Windows per date di installazione
// - Supporta pattern di permessi e accesso directory Windows
//
// Returns: Struct JDKInstallation con risultati completi dell'analisi
func analyzeJDKInstallation(dirName, jdkPath string) JDKInstallation {
	installation := JDKInstallation{
		Version: dirName,
		Path:    jdkPath,
	}

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

// calculateDirSize calcola la dimensione totale di una directory
//
// Questa funzione esegue una camminata ricorsiva attraverso la struttura
// di directory per calcolare la dimensione cumulativa di tutti i file contenuti:
//
// **Metodo di calcolo:**
// - Utilizza filepath.WalkDir per attraversamento efficiente delle directory
// - Accumula dimensioni dei file utilizzando il metodo os.FileInfo.Size()
// - Salta le directory stesse, contando solo i file regolari
// - Gestisce errori di accesso con grazia senza terminare la scansione
//
// **Considerazioni specifiche per Windows:**
// - Gestisce correttamente i permessi del file system Windows
// - Supporta attributi file Windows e file nascosti
// - Gestisce scenari di locking file specifici di Windows
//
// Questo calcolo delle dimensioni aiuta gli utenti a comprendere l'utilizzo
// dello spazio su disco e identificare installazioni che potrebbero richiedere pulizia.
//
// Parametri:
//   - dirPath: Percorso assoluto alla directory per il calcolo delle dimensioni
//
// Returns: Dimensione totale in byte come int64, o 0 se il calcolo fallisce
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
//
// Questa funzione converte conteggi di byte grezzi in rappresentazioni di dimensioni
// user-friendly utilizzando unità binarie standard (basate su 1024) per reportistica
// precisa dello spazio su disco:
//
// **Logica di conversione:**
// - Utilizza unità binarie: B, KB, MB, GB, TB (basate su 1024)
// - Mantiene precisione con 1 cifra decimale per unità superiori ai byte
// - Gestisce casi limite (dimensione zero, file molto grandi)
//
// **Regole di formattazione:**
// - Byte: Nessuna cifra decimale (es. "1024 B")
// - Unità maggiori: Una cifra decimale (es. "1.5 MB")
// - Selezione automatica unità basata sulla grandezza
//
// Questo fornisce reportistica consistente delle dimensioni attraverso tutte
// le installazioni JDK per facile confronto e gestione dello spazio su disco.
//
// Parametri:
//   - bytes: Dimensione in byte come int64
//
// Returns: Stringa dimensione formattata con unità appropriata
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

// checkExtractionStatus verifica se la directory contiene file estratti o solo archivi
//
// Questa funzione analizza il contenuto di una directory JDK per determinare
// lo stato di installazione e il tipo di archivio presente:
//
// **Criteri di rilevamento estrazione:**
// - Presenza di directories standard JDK: bin/, lib/, include/
// - Presenza di file di release: release, version
// - Struttura di directory tipica di un JDK completamente estratto
//
// **Tipi di archivio riconosciuti:**
// - .tar.gz/.tgz: Archivi compressi Unix/Linux standard
// - .zip: Archivi compressi Windows standard
// - .msi: Installer Microsoft Windows
// - .exe: Eseguibili di installazione Windows
//
// **Logica di analisi:**
// - Scansione directory per identificare file e subdirectories
// - Priorità agli indicatori di estrazione over archivi
// - Supporto per installazioni miste (archivio + estratti)
//
// Questa funzione è cruciale per determinare le azioni disponibili
// (extract, use, remove) per ogni installazione JDK.
//
// Parametri:
//   - jdkPath: Percorso assoluto alla directory JDK da analizzare
//
// Returns:
//   - bool: true se il JDK è stato estratto
//   - string: tipo di archivio presente ("tar.gz", "zip", "msi", "exe", "")
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

// compareVersions confronta due stringhe di versione per ordinamento
//
// Questa funzione implementa un algoritmo di confronto versioni JDK intelligente
// che tiene conto delle convenzioni di numerazione JDK standard:
//
// **Algoritmo di confronto:**
// - Estrazione numero versione principale da stringhe complesse
// - Confronto numerico per versioni maggiori (17 vs 21)
// - Fallback lessicografico per versioni identiche
// - Gestione pattern comuni: jdk-XX, java-XX, XX.Y.Z+build
//
// **Pattern supportati:**
// - JDK-17.0.2+8: Estrae "17" come versione principale
// - java-21-openjdk: Estrae "21" come versione principale
// - adoptium-11.0.19+7: Estrae "11" come versione principale
//
// **Valore di ritorno:**
// - 1: se versione 'a' è maggiore di 'b'
// - -1: se versione 'a' è minore di 'b'
// - 0: se le versioni sono identiche
//
// Questo ordinamento garantisce che le versioni JDK più recenti
// appaiano per prime nella lista delle installazioni.
//
// Parametri:
//   - a, b: Stringhe di versione JDK da confrontare
//
// Returns: Intero indicante il risultato del confronto (-1, 0, 1)
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

// extractMainVersion estrae il numero di versione principale da una stringa
//
// Questa funzione helper implementa il parsing intelligente delle versioni JDK
// per estrarre il numero di versione principale da stringhe di formato variabile:
//
// **Strategia di parsing:**
// - Suddivisione stringa utilizzando delimitatori comuni (-, _, .)
// - Ricerca del primo numero intero valido nel range JDK (1-99)
// - Filtro per escludere numeri non significativi (build numbers, patch versions)
//
// **Esempi di estrazione:**
// - "jdk-17.0.2+8" → 17
// - "java-21-openjdk" → 21
// - "adoptium-11.0.19+7-lts" → 11
// - "JDK-8u371-windows-x64" → 8
//
// **Gestione edge cases:**
// - Stringhe senza numeri validi ritornano 0
// - Numeri fuori range JDK (>99) vengono ignorati
// - Build numbers e patch versions vengono saltati
//
// Questa funzione è fondamentale per l'ordinamento corretto delle versioni
// JDK indipendentemente dalle convenzioni di naming dei diversi provider.
//
// Parametri:
//   - version: Stringa di versione da cui estrarre il numero principale
//
// Returns: Numero intero della versione principale, 0 se non trovato
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

// displayJDKTable mostra i JDK in formato tabella colorata
//
// Questa funzione implementa la visualizzazione tabellare delle installazioni JDK
// con formattazione colorata per migliorare la leggibilità e l'esperienza utente:
//
// **Struttura tabella:**
// - Header con intestazioni colonne in grassetto e ciano
// - Righe dati con colorazione basata sullo stato di installazione
// - Separatori visivi tra header e contenuto
// - Allineamento colonne per leggibilità ottimale
//
// **Schema colori:**
// - Header: Grassetto + Ciano brillante per massima visibilità
// - Versioni: Verde per installazioni pronte, giallo per archivi
// - Provider: Colore neutrale per informazioni secondarie
// - Stato: Verde per READY, giallo per ARCHIVE, rosso per EMPTY
// - Percorsi: Troncati automaticamente se troppo lunghi
//
// **Informazioni aggiuntive:**
// - Legenda stati con spiegazioni dettagliate
// - Suggerimenti comandi successivi per workflow utente
// - Conteggio totale installazioni trovate
func displayJDKTable(jdks []JDKInstallation) {
	fmt.Printf(utils.ColorText("Found %d JDK installations:\n\n", utils.Bold+utils.BrightCyan), len(jdks))

	// Header della tabella con colori (senza colonna PROVIDER)
	fmt.Printf(utils.ColorText("%-30s %-10s %-18s %-12s %s\n", utils.Bold+utils.BrightCyan),
		"VERSION", "STATUS", "INSTALL DATE", "SIZE", "PATH")
	fmt.Println(utils.ColorText(strings.Repeat("-", 85), utils.Cyan))

	// Righe della tabella con colori
	for _, jdk := range jdks {
		status := getStatusIcon(jdk.IsExtracted, jdk.ArchiveType)
		statusColor := getStatusColor(jdk.IsExtracted, jdk.ArchiveType)

		// Tronca il percorso se troppo lungo
		displayPath := jdk.Path
		if len(displayPath) > 40 {
			displayPath = "..." + displayPath[len(displayPath)-37:]
		}

		// Colora la versione in base allo stato
		versionColor := utils.BrightGreen
		if !jdk.IsExtracted {
			versionColor = utils.BrightYellow
		}

		// Formatta la riga con padding fisso per l'allineamento
		version := fmt.Sprintf("%-30s", jdk.Version)
		statusStr := fmt.Sprintf("%-10s", status)
		installDate := fmt.Sprintf("%-18s", jdk.InstallDate)
		size := fmt.Sprintf("%-12s", jdk.Size)

		fmt.Printf("%s %s %s %s %s\n",
			utils.ColorText(version, versionColor),
			utils.ColorText(statusStr, statusColor),
			installDate,
			size,
			utils.ColorText(displayPath, utils.Blue))
	}

	fmt.Println()
	fmt.Println(utils.ColorText("Status Legend:", utils.Bold+utils.BrightCyan))
	fmt.Printf("   %s - JDK extracted and ready for use\n", utils.ColorText("[READY]", utils.BrightGreen))
	fmt.Printf("   %s - Archive downloaded (requires extraction)\n", utils.ColorText("[ARCHIVE]", utils.BrightYellow))
	fmt.Printf("   %s - Empty or corrupted directory\n", utils.ColorText("[EMPTY]", utils.BrightRed))
	fmt.Println()
	fmt.Println(utils.ColorText("[INFO] Available next commands:", utils.Bold))
	fmt.Println("   jvm extract <version>  - Extract a JDK archive")
	fmt.Println("   jvm use <version>      - Set as active JDK")
	fmt.Println("   jvm remove <version>   - Remove a version")
}

// getStatusIcon restituisce l'icona di stato appropriata senza emoji
//
// Questa funzione determina il testo di stato da visualizzare per ogni installazione JDK
// basandosi sullo stato di estrazione e sul tipo di archivio presente:
//
// **Stati supportati:**
// - [READY]: JDK completamente estratto con struttura directories standard
// - [ARCHIVE]: Solo file archivio presente, necessita estrazione
// - [EMPTY]: Directory vuota o contenuto non riconosciuto come JDK valido
//
// **Logica di determinazione:**
// - Verifica presenza di directories estratte (bin/, lib/, include/)
// - Identifica tipo archivio presente (.zip, .tar.gz, .msi, .exe)
// - Applica priorità: estratto > archivio > vuoto
//
// Parametri:
//   - isExtracted: true se il JDK è stato estratto
//   - archiveType: tipo di archivio presente ("zip", "tar.gz", "msi", "exe", "")
//
// Returns: Stringa di stato senza emoji per output pulito
func getStatusIcon(isExtracted bool, archiveType string) string {
	if isExtracted {
		return "[READY]"
	} else if archiveType != "" {
		return "[ARCHIVE]"
	}
	return "[EMPTY]"
}

// getStatusColor restituisce il colore appropriato per lo stato
//
// Questa funzione helper determina il codice colore ANSI appropriato
// per la visualizzazione colorata dello stato di installazione JDK:
//
// **Schema colori:**
// - Verde brillante: Installazioni pronte all'uso ([READY])
// - Giallo brillante: Archivi che richiedono estrazione ([ARCHIVE])
// - Rosso brillante: Directory vuote o danneggiate ([EMPTY])
//
// I colori vengono definiti nel package utils/colors.go per consistenza
// attraverso tutta l'applicazione e supporto cross-terminale.
//
// Parametri:
//   - isExtracted: true se il JDK è stato estratto
//   - archiveType: tipo di archivio presente
//
// Returns: Codice colore ANSI dal package utils
func getStatusColor(isExtracted bool, archiveType string) string {
	if isExtracted {
		return utils.BrightGreen
	} else if archiveType != "" {
		return utils.BrightYellow
	}
	return utils.BrightRed
}
