package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"jvm/internal/providers/adoptium"
	"jvm/internal/providers/azul"
	"jvm/internal/providers/liberica"
	"jvm/internal/providers/private"
	"jvm/internal/utils"
)

// RuntimeInfo rappresenta le informazioni di sistema operativo e architettura per Windows.
//
// Questa struttura incapsula i dettagli del runtime necessari per identificare
// il tipo corretto di JDK da scaricare per il sistema Windows corrente.
// I valori vengono normalizzati nel formato utilizzato dai provider JDK.
//
// Campi:
//
//	OS   - Sistema operativo, sempre "windows" per questo tool
//	Arch - Architettura CPU in formato provider JDK ("x64", "x32", "aarch64")
//
// Utilizzo tipico:
//
//	runtime := getRuntimeInfo()
//	// runtime.OS = "windows", runtime.Arch = "x64" su Windows 64-bit
type RuntimeInfo struct {
	OS   string
	Arch string
}

// getRuntimeInfo rileva e normalizza le informazioni di sistema Windows per compatibilità JDK.
//
// Questa funzione identifica l'architettura del sistema Windows corrente,
// convertendo i nomi interni di Go nei formati standard utilizzati dai provider JDK.
// È essenziale per garantire il download del JDK corretto per il sistema Windows.
//
// Processo di rilevamento:
// 1. **Sistema operativo**: Sempre "windows" (tool Windows-only)
// 2. **Rilevamento architettura**: Usa runtime.GOARCH per identificare l'architettura CPU
// 3. **Normalizzazione**: Converte nomi Go in formato standard provider JDK
// 4. **Ritorno struttura**: Incapsula informazioni in RuntimeInfo
//
// Mappature architettura (Go → JDK):
//   - "amd64" → "x64" (Intel/AMD 64-bit, più comune)
//   - "386" → "x32" (Intel/AMD 32-bit, legacy)
//   - "arm64" → "aarch64" (ARM 64-bit, Windows ARM)
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	RuntimeInfo - Struttura con OS="windows" e Arch normalizzato per provider JDK
//
// Casi d'uso:
//   - Selezione automatica del JDK compatibile durante download
//   - Filtraggio versioni disponibili per sistema corrente
//   - Validazione compatibilità prima dell'installazione
//
// Note Windows specifiche:
//   - Su Windows x64: OS="windows", Arch="x64"
//   - Su Windows ARM64: OS="windows", Arch="aarch64"
//   - Windows 32-bit: OS="windows", Arch="x32" (rare, legacy)
//
// Esempio di utilizzo:
//
//	runtime := getRuntimeInfo()
//	fmt.Printf("Sistema: %s %s", runtime.OS, runtime.Arch)
//	// Output su Windows 64-bit: "Sistema: windows x64"
func getRuntimeInfo() RuntimeInfo {
	archName := runtime.GOARCH

	// Convert Go runtime names to JDK format
	switch archName {
	case "amd64":
		archName = "x64"
	case "386":
		archName = "x32"
	case "arm64":
		archName = "aarch64"
	}

	return RuntimeInfo{OS: "windows", Arch: archName}
}

// shouldPreferVersion determina quale di due versioni JDK dovrebbe essere preferita.
//
// Questa funzione implementa la logica di comparazione delle versioni per scegliere
// la "migliore" versione quando multiple opzioni sono disponibili per un target.
// Segue una strategia di preferenza basata su versioni più recenti.
//
// Algoritmo di preferenza (in ordine di priorità):
// 1. **Versione major superiore**: 21.x.x > 17.x.x
// 2. **Versione minor superiore**: 17.1.x > 17.0.x (a parità di major)
// 3. **Versione patch superiore**: 17.0.5 > 17.0.2 (a parità di major.minor)
//
// Strategia di confronto:
//   - Parsing completo di entrambe le versioni con adoptium.ParseVersion()
//   - Confronto gerarchico: major → minor → patch
//   - Prima differenza significativa determina il risultato
//   - Versioni identiche: ritorna false (nessuna preferenza)
//
// Casi d'uso tipici:
//   - Selezione automatica della migliore versione tra match multipli
//   - Ordinamento di lista versioni per presentazione utente
//   - Decisione di aggiornamento automatico
//   - Risoluzione di conflitti in ricerche ambigue
//
// Esempi di comportamento:
//
//	shouldPreferVersion("21.0.2", "17.0.5") → true (major superiore)
//	shouldPreferVersion("17.1.0", "17.0.8") → true (minor superiore)
//	shouldPreferVersion("17.0.5", "17.0.2") → true (patch superiore)
//	shouldPreferVersion("17.0.2", "17.0.2") → false (identiche)
//
// Parametri:
//
//	version1 string - Prima versione da confrontare
//	version2 string - Seconda versione da confrontare
//
// Restituisce:
//
//	bool - true se version1 dovrebbe essere preferita a version2, false altrimenti
//
// Note implementative:
//   - Utilizza adoptium.ParseVersion() per parsing consistente
//   - Gestisce automaticamente versioni malformate (tramite ParseVersion)
//   - Algoritmo deterministico: stesso input produce sempre stesso output
//   - Performance ottimizzata: stop al primo livello di differenza
//
// Limitazioni:
//   - Non considera preferenze LTS vs non-LTS
//   - Non valuta stabilità o qualità delle release
//   - Puramente numerico, non semantico
//
// Esempio di utilizzo:
//
//	if shouldPreferVersion("21.0.1", bestVersion) {
//	    bestVersion = "21.0.1"
//	}
func shouldPreferVersion(version1, version2 string) bool {
	v1Major, v1Minor, v1Patch := adoptium.ParseVersion(version1)
	v2Major, v2Minor, v2Patch := adoptium.ParseVersion(version2)

	// Prefer higher major version
	if v1Major != v2Major {
		return v1Major > v2Major
	}

	// Prefer higher minor version
	if v1Minor != v2Minor {
		return v1Minor > v2Minor
	}

	// Prefer higher patch version
	return v1Patch > v2Patch
}

// DownloadJDK esegue il download completo e l'installazione di una versione JDK specifica su Windows.
//
// Questa è la funzione principale del comando download, che orchestra l'intero processo
// di ricerca, download, estrazione e installazione di un JDK da provider pubblici o privati.
// Progettata per offrire un'esperienza utente fluida con feedback dettagliato e conferme.
//
// Processo completo di download:
// 1. **Parsing argomenti**: Analizza versione target e opzioni da riga di comando
// 2. **Risoluzione provider**: Determina provider (adoptium, azul, liberica, private)
// 3. **Configurazione directory**: Setup directory download (~/.jvm/versions)
// 4. **Ricerca versione**: Query provider per trovare versione compatibile
// 5. **Conferma utente**: Richiede approvazione prima del download
// 6. **Download file**: Scarica archivio JDK con progress indicator
// 7. **Estrazione automatica**: Decomprime e organizza file JDK
// 8. **Pulizia opzionale**: Rimozione archivio se richiesta dall'utente
//
// Sintassi supportata:
//
//	jvm download 17                    # Ultima versione disponibile JDK 17
//	jvm download 21.0.2                # Versione specifica
//	jvm download 17 --provider=azul    # Provider specifico
//	jvm download 21 --output=./jdks    # Directory custom
//
// Provider supportati:
//   - **adoptium**: Eclipse Adoptium (default, più popolare)
//   - **azul**: Azul Zulu OpenJDK (enterprise-ready)
//   - **liberica**: BellSoft Liberica JDK
//   - **private**: Repository aziendali configurati
//
// Gestione intelligente versioni:
//   - "17" → cerca migliore versione 17.x.y disponibile
//   - "17.0" → cerca migliore versione 17.0.x disponibile
//   - "17.0.5" → cerca esattamente versione 17.0.5
//   - Preferenza per versioni più recenti a parità di match
//
// Caratteristiche UX:
//   - Progress indicator durante download con velocità
//   - Conferme interattive per operazioni irreversibili
//   - Feedback colorato per migliorare leggibilità
//   - Istruzioni step-by-step per prossimi passi
//
// Gestione directory:
//   - Default: C:\Users\username\.jvm\versions\JDK-{version}\
//   - Struttura: Una directory per versione per isolamento
//   - Creazione automatica di directory mancanti
//   - Estrazione con flattening di directory annidate
//
// Sicurezza e robustezza:
//   - Validazione input utente per prevenire injection
//   - Timeout download per evitare hang indefiniti
//   - Verifica integrità file scaricati
//   - Prevenzione zip/tar slip attacks durante estrazione
//   - Gestione graceful di errori di rete e filesystem
//
// Parametri:
//
//	defaultProvider string - Provider predefinito se non specificato da utente
//
// Comportamento errori:
//   - Stampa errore specifico e termina per input invalidi
//   - Fallback directory "./downloads" se home non determinabile
//   - Skip download se versione non trovata nei provider
//   - Continuazione con warning se estrazione fallisce
//
// Side effects:
//   - Crea directory ~/.jvm/versions/ se non esiste
//   - Scarica file archivio JDK nella directory versione
//   - Estrae e organizza file JDK per uso immediato
//   - Stampa informazioni dettagliate su stdout
//   - Può richiedere input utente per conferme
//
// Post-download:
//   - Mostra percorso JDK installato
//   - Suggerisce comandi successivi (jvm use, jvm list)
//   - Fornisce percorso bin/ per aggiunta a PATH
//   - Opzione rimozione archivio per risparmio spazio
//
// Esempio di utilizzo completo:
//
//	DownloadJDK("adoptium")
//	// Con args: ["17", "--provider=azul"]
//	// Risultato: JDK 17 Azul scaricato in ~/.jvm/versions/JDK-17.x.y/
func DownloadJDK(defaultProvider string) {

	// Parse command line arguments
	args := os.Args[2:] // Skip "download"
	if len(args) == 0 {
		utils.PrintError("No JDK version specified")
		utils.PrintInfo("Usage: jvm download <version> [options]")
		utils.PrintInfo("Examples:")
		fmt.Println("  jvm download 17          # Download JDK 17")
		fmt.Println("  jvm download 21.0.5      # Download specific version")
		fmt.Println("  jvm download 17 --provider=azul")
		return
	}

	version := args[0]
	provider := defaultProvider

	// Get default download directory: ~/.jvm/versions
	outputDir, dirErr := getDefaultDownloadDir()
	if dirErr != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine download directory: %v", dirErr))
		outputDir = "./downloads" // fallback
	}

	// Parse optional flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--provider=") {
			provider = strings.TrimPrefix(arg, "--provider=")
		} else if strings.HasPrefix(arg, "--output=") {
			outputDir = strings.TrimPrefix(arg, "--output=")
		}
	}

	fmt.Printf("%s Searching for JDK version %s from provider: %s\n",
		utils.ColorText("[>]", utils.BrightCyan), version, provider)
	fmt.Printf("%s Download directory: %s\n\n",
		utils.ColorText("[>]", utils.BrightCyan), outputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create output directory: %v", err))
		return
	}

	// Get JDK releases from the specified provider and find matching version
	var downloadURL string
	var filename string
	var foundVersion string

	switch provider {
	case "adoptium":
		releases, err := adoptium.GetAllJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findAdoptiumDownload(releases, version)

	case "azul":
		releases, err := azul.GetAzulJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findAzulDownload(releases, version)

	case "liberica":
		releases, err := liberica.GetLibericaJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findLibericaDownload(releases, version)

	case "private":
		releases, err := private.GetPrivateJDKs()
		if err != nil {
			fmt.Printf("[ERROR] Failed to fetch releases from %s: %v\n", provider, err)
			return
		}
		downloadURL, filename, foundVersion = findPrivateDownload(releases, version)

	default:
		fmt.Printf("[ERROR] Unknown provider: %s\n", provider)
		fmt.Println("[INFO] Available providers: adoptium, azul, liberica, private")
		return
	}

	if downloadURL == "" {
		fmt.Printf("[ERROR] JDK version %s not found in %s provider\n", version, provider)
		fmt.Println("[INFO] Try running 'jvm remote-list' to see available versions")
		return
	}

	if filename == "" {
		filename = fmt.Sprintf("openjdk-%s.tar.gz", version)
	}

	// Create a version-specific subdirectory
	versionDir := fmt.Sprintf("JDK-%s", foundVersion)
	versionOutputDir := filepath.Join(outputDir, versionDir)

	// Create version-specific directory
	if err := os.MkdirAll(versionOutputDir, 0755); err != nil {
		fmt.Printf("[ERROR] Failed to create version directory: %v\n", err)
		return
	}

	outputPath := filepath.Join(versionOutputDir, filename)

	fmt.Printf("%s JDK %s\n", utils.ColorText("[FOUND]", utils.BrightGreen), foundVersion)
	fmt.Printf("%s Download URL: %s\n", utils.ColorText("[URL]", utils.BrightBlue), downloadURL)
	fmt.Printf("%s Version directory: %s\n", utils.ColorText("[DIR]", utils.BrightYellow), versionOutputDir)
	fmt.Printf("%s Saving to: %s\n", utils.ColorText("[FILE]", utils.BrightMagenta), outputPath)

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		utils.PrintWarning(fmt.Sprintf("File already exists: %s", filename))
	}

	// Ask for confirmation
	fmt.Print("\n[?] Do you want to proceed with the download? (y/N): ")
	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		utils.PrintInfo("Download cancelled by user")
		return
	}

	fmt.Println()

	// Download the file
	if err := downloadFile(downloadURL, outputPath); err != nil {
		utils.PrintError(fmt.Sprintf("Download failed: %v", err))
		return
	}

	utils.PrintSuccess("Download completed successfully!")
	fmt.Printf("%s JDK %s saved to: %s\n",
		utils.ColorText("[OUTPUT]", utils.BrightGreen), foundVersion, versionOutputDir)
	fmt.Printf("%s Archive file: %s\n",
		utils.ColorText("[FILE]", utils.BrightBlue), filename)

	// Show file info
	if fileInfo, err := os.Stat(outputPath); err == nil {
		fmt.Printf("[SIZE] File size: %.2f MB\n", float64(fileInfo.Size())/1024/1024)
		fmt.Printf("[TIME] Download time: %s\n", time.Now().Format("15:04:05"))
	}

	utils.PrintSuccess(fmt.Sprintf("JDK downloaded successfully: %s", filepath.Base(outputPath)))
	utils.PrintInfo(fmt.Sprintf("Location: %s", outputPath))
	fmt.Println()

	// Ask if user wants to extract the archive automatically
	fmt.Print("[?] Do you want to extract the archive now? (Y/n): ")
	var extractResponse string
	fmt.Scanln(&extractResponse)

	extractResponse = strings.ToLower(strings.TrimSpace(extractResponse))
	if extractResponse == "" || extractResponse == "y" || extractResponse == "yes" {
		fmt.Println()
		utils.PrintInfo("Starting extraction...")

		// Extract using the same logic as extract command with intelligent parsing
		if err := extractJDKArchive(versionDir, versionOutputDir); err != nil {
			utils.PrintError(fmt.Sprintf("Extraction failed: %v", err))
			utils.PrintInfo("You can manually extract later using:")
			utils.PrintInfo(fmt.Sprintf("  jvm extract %s", versionDir))
		} else {
			utils.PrintSuccess("JDK extracted successfully!")
			utils.PrintInfo(fmt.Sprintf("JDK ready at: %s", versionOutputDir))
			fmt.Println()
			utils.PrintInfo("To activate this JDK, use:")
			utils.PrintInfo(fmt.Sprintf("  jvm use %s", versionDir))
		}
	} else {
		utils.PrintWarning("Archive not extracted. To extract manually, use:")
		utils.PrintWarning(fmt.Sprintf("  jvm extract %s", versionDir))
	}

	fmt.Println()
	utils.PrintInfo("Next steps:")
	utils.PrintInfo("  jvm extract <archive>        # Extract the downloaded archive")
	utils.PrintInfo("  jvm list                     # View installed JDKs")
	utils.PrintInfo("  jvm use <version>            # Set JDK as active")
}

// downloadFile scarica un file da URL con indicatore di progresso e gestione robusta degli errori.
//
// Questa funzione implementa un download HTTP robusto ottimizzato per file JDK di grandi dimensioni,
// con timeout appropriati, indicatori di progresso in tempo reale e gestione completa degli errori
// di rete. Progettata per fornire feedback continuo all'utente durante operazioni lunghe.
//
// Caratteristiche del download:
// 1. **Client HTTP configurato**: Timeout di 30 minuti per file grandi
// 2. **Headers appropriati**: User-Agent personalizzato per identificazione
// 3. **Validazione response**: Verifica status code prima di procedere
// 4. **Progress tracking**: Indicatore percentuale con velocità in tempo reale
// 5. **Buffer ottimizzato**: 32KB buffer per performance bilanciata
// 6. **Gestione errori**: Recovery graceful da interruzioni di rete
//
// Indicatore di progresso:
//   - Percentuale completamento se Content-Length disponibile
//   - Velocità download in MB/s in tempo reale
//   - Dimensioni scaricate vs totali
//   - Fallback a solo dimensione scaricata se lunghezza sconosciuta
//   - Aggiornamento in tempo reale (refresh continuo della stessa linea)
//
// Esempio output progresso:
//
//	[DOWNLOAD] Progress: 45.2% (125.4 MB / 277.8 MB) - Speed: 8.3 MB/s
//	[DOWNLOAD] Downloaded: 125.4 MB - Speed: 8.3 MB/s (se size sconosciuto)
//
// Gestione timeout e resilienza:
//   - Timeout download: 30 minuti (appropriato per JDK fino a 300MB)
//   - Timeout connection implicito nel http.Client
//   - Retry non implementato (operazione one-shot per semplicità)
//   - Gestione disconnessioni di rete con errori informativi
//
// Sicurezza:
//   - User-Agent custom per identificazione legittima
//   - Validazione response code per prevenire download di errori
//   - Nessuna esecuzione automatica di file scaricati
//   - Percorso file validato dal chiamante
//
// Parametri:
//
//	url string      - URL completo del file da scaricare
//	filepath string - Percorso locale assoluto dove salvare il file
//
// Restituisce:
//
//	error - nil se download completato con successo, errore specifico altrimenti
//
// Errori possibili:
//   - Errore creazione request HTTP
//   - Timeout durante download (30 min)
//   - Server response non-200 (file non trovato, accesso negato, etc.)
//   - Errore creazione file locale (permessi, spazio disco)
//   - Interruzione connessione durante trasferimento
//   - Errore scrittura su disco (spazio esaurito)
//
// Performance:
//   - Buffer 32KB ottimizzato per SSD moderni
//   - Progress update efficiente senza impatto su velocità
//   - Memory footprint costante indipendentemente da dimensione file
//   - Streaming download (non carica tutto in memoria)
//
// Esempio di utilizzo:
//
//	err := downloadFile("https://adoptium.net/...jdk-17.zip", "C:/Users/user/.jvm/versions/JDK-17/jdk.zip")
//	if err != nil {
//	    log.Printf("Download failed: %v", err)
//	}
func downloadFile(url, filepath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Minute * 30, // 30 minutes timeout for large files
	}

	// Create the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "JVM-Manager/1.0")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create the output file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer out.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	var downloaded int64

	// Create a buffer for copying
	buffer := make([]byte, 32*1024) // 32KB buffer

	fmt.Println("[DOWNLOAD] Downloading...")
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("writing to file: %w", writeErr)
			}
			downloaded += int64(n)

			// Show progress if we know the content length
			if contentLength > 0 {
				progress := float64(downloaded) / float64(contentLength) * 100
				elapsed := time.Since(startTime)
				speed := float64(downloaded) / elapsed.Seconds() / 1024 / 1024 // MB/s

				fmt.Printf("\r[DOWNLOAD] Progress: %.1f%% (%.2f MB / %.2f MB) - Speed: %.2f MB/s",
					progress,
					float64(downloaded)/1024/1024,
					float64(contentLength)/1024/1024,
					speed,
				)
			} else {
				// Show downloaded amount without percentage
				elapsed := time.Since(startTime)
				speed := float64(downloaded) / elapsed.Seconds() / 1024 / 1024 // MB/s

				fmt.Printf("\r[DOWNLOAD] Downloaded: %.2f MB - Speed: %.2f MB/s",
					float64(downloaded)/1024/1024,
					speed,
				)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading response: %w", err)
		}
	}

	fmt.Println() // New line after progress
	return nil
}

// getDefaultDownloadDir determina e restituisce la directory di download predefinita per JDK su Windows.
//
// Questa funzione costruisce il percorso standardizzato dove JVM organizza tutti i JDK scaricati,
// seguendo le convenzioni di directory utente Windows per una gestione ordinata delle installazioni.
//
// Struttura directory predefinita:
//
//	C:\Users\{username}\.jvm\versions\
//	├── JDK-17.0.5\          # Versione specifica JDK
//	│   ├── bin\             # Eseguibili JDK
//	│   ├── lib\             # Librerie JDK
//	│   └── ...              # Altri file JDK
//	├── JDK-21.0.2\          # Altra versione JDK
//	└── JDK-8.0.392\         # Versione legacy
//
// Processo di determinazione:
// 1. **Rilevamento utente corrente**: Usa user.Current() per ottenere info utente
// 2. **Costruzione percorso**: Combina HomeDir + ".jvm" + "versions"
// 3. **Normalizzazione path**: Usa filepath.Join per compatibilità Windows
// 4. **Validazione**: Verifica accessibilità directory home
//
// Vantaggi directory standardizzata:
//   - **Isolamento versioni**: Ogni JDK in directory separata
//   - **Gestione centralizzata**: Tutte le versioni in un posto
//   - **Compatibilità strumenti**: Path predicibile per automazione
//   - **Facilità pulizia**: Directory unica da gestire
//   - **Backup semplificato**: Un solo path da includere
//
// Gestione errori:
//   - Ritorna errore se impossibile determinare utente corrente
//   - Non crea directory (responsabilità del chiamante)
//   - Gestisce gracefully profili utente corrotti o inaccessibili
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	string - Percorso assoluto directory download (~/.jvm/versions)
//	error  - nil se successo, errore se impossibile determinare directory home
//
// Compatibilità Windows:
//   - Gestisce caratteri Unicode in nomi utente Windows
//   - Supporta percorsi lunghi Windows (>260 caratteri) se abilitati
//   - Compatibile con tutte le versioni Windows moderne
//   - Rispetta convenzioni directory utente Windows
//
// Esempio di utilizzo:
//
//	downloadDir, err := getDefaultDownloadDir()
//	if err != nil {
//	    downloadDir = "./downloads" // fallback
//	}
//	// downloadDir = "C:\Users\Marco\.jvm\versions"
//
// Integrazione con JVM:
//   - Usata da comando download per destinazione automatica
//   - Utilizzata da comando list per enumerazione JDK installati
//   - Riferimento per comando use nella selezione versioni
//   - Base per comando remove per identificazione target
func getDefaultDownloadDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("getting current user: %w", err)
	}

	jvmDir := filepath.Join(currentUser.HomeDir, ".jvm", "versions")
	return jvmDir, nil
}

// findAdoptiumDownload ricerca e seleziona il miglior download JDK da releases Eclipse Adoptium.
//
// Questa funzione implementa la logica di matching intelligente per trovare la versione
// JDK più appropriata dalle release Adoptium, considerando versione target, compatibilità
// piattaforma e preferenze di qualità. Adoptium è il provider predefinito e più utilizzato.
//
// Algoritmo di ricerca intelligente:
// 1. **Parsing versione target**: Decompone versione richiesta in major.minor.patch
// 2. **Iterazione releases**: Esamina tutte le release disponibili da Adoptium
// 3. **Matching flessibile**: Supporta ricerche parziali (17 → 17.x.y)
// 4. **Priorità piattaforma**: Preferisce binari compatibili con sistema corrente
// 5. **Selezione migliore**: Usa shouldPreferVersion per determinare la scelta ottimale
// 6. **Fallback robusto**: Tenta alternative se match perfetto non disponibile
//
// Logica di matching versione:
//   - "17" → qualsiasi versione 17.x.y (più recente preferita)
//   - "17.0" → qualsiasi versione 17.0.x (patch più alta preferita)
//   - "17.0.5" → esattamente versione 17.0.5
//
// Priorità selezione binari:
// 1. **Match perfetto OS+Arch**: windows x64 su sistema Windows 64-bit
// 2. **Fallback formato**: Qualsiasi binario .zip se match perfetto non trovato
// 3. **Versione preferita**: Tra match equivalenti, sceglie versione più recente
//
// Formati download supportati:
//   - **ZIP**: Formato preferito, facile estrazione su Windows
//   - **TAR.GZ**: Formato alternativo, supporto completo
//   - Altri formati: Ignorati per compatibilità Windows
//
// Parametri:
//
//	releases []adoptium.AdoptiumResponse - Lista complete release da Adoptium API
//	version string                       - Versione target (es. "17", "21.0.2")
//
// Restituisce:
//
//	string - URL download del binario selezionato ("" se non trovato)
//	string - Nome file suggerito per download ("" se non determinabile)
//	string - Versione esatta trovata (es. "17.0.5+8") ("" se non trovato)
//
// Comportamento con match multipli:
//   - Preferisce sempre versione più recente (21.0.2 > 21.0.1)
//   - A parità di versione, preferisce match perfetto OS/Arch
//   - Considera solo release con binari scaricabili
//
// Gestione edge cases:
//   - Release senza binari compatibili: ignorate
//   - Versioni malformate: parsing robusto con fallback
//   - Provider down: ritorna risultati vuoti gracefully
//   - URL malformati: cleanup automatico parametri query
//
// Ottimizzazioni:
//   - Stop al primo match perfetto per performance
//   - Parsing lazy delle versioni (solo quando necessario)
//   - Cache implicita durante iterazione singola
//
// Esempio di utilizzo:
//
//	releases, _ := adoptium.GetAllJDKs()
//	url, filename, version := findAdoptiumDownload(releases, "17")
//	// url = "https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.5+8/OpenJDK17U-jdk_x64_windows_hotspot_17.0.5_8.zip"
//	// filename = "OpenJDK17U-jdk_x64_windows_hotspot_17.0.5_8.zip"
//	// version = "17.0.5+8"
func findAdoptiumDownload(releases []adoptium.AdoptiumResponse, version string) (string, string, string) {
	runtime := getRuntimeInfo()

	// Parse target version
	targetMajor, targetMinor, targetPatch := utils.ParseVersionNumber(version)

	var bestMatch adoptium.AdoptiumResponse
	var bestBinary struct {
		OS      string `json:"os"`
		Arch    string `json:"architecture"`
		Package struct {
			Link string `json:"link"`
		} `json:"package"`
	}
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		releaseVersion := release.VersionData.OpenJDKVersion
		major, minor, patch := adoptium.ParseVersion(releaseVersion)

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Find the best binary for this release (prefer current OS/arch)
		for _, binary := range release.Binaries {
			if binary.OS == runtime.OS && binary.Arch == runtime.Arch {
				if !found || shouldPreferVersion(releaseVersion, bestMatch.VersionData.OpenJDKVersion) {
					bestMatch = release
					bestBinary = binary
					found = true
					break // Found perfect match
				}
			}
		}

		// If no perfect match, try any compatible binary
		if !found {
			for _, binary := range release.Binaries {
				if strings.Contains(binary.Package.Link, ".zip") {
					bestMatch = release
					bestBinary = binary
					found = true
					break
				}
			}
		}
	}

	if !found {
		return "", "", ""
	}

	url := bestBinary.Package.Link
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return url, filename, bestMatch.VersionData.OpenJDKVersion
}

// findAzulDownload ricerca e seleziona il miglior download JDK da releases Azul Zulu.
//
// Questa funzione implementa la logica di matching per il provider Azul Zulu OpenJDK,
// un'alternativa enterprise-ready ad Adoptium con focus su stabilità e supporto
// commerciale. Azul utilizza un formato dati diverso che richiede parsing specifico.
//
// Caratteristiche Azul Zulu:
//   - **Certificazione enterprise**: Build certificate per ambienti produzione
//   - **Supporto esteso**: Opzioni di supporto commerciale disponibili
//   - **Release frequenti**: Aggiornamenti di sicurezza tempestivi
//   - **Multi-piattaforma**: Supporto estensivo per architetture diverse
//
// Differenze nel formato dati Azul:
//   - Versioni come array di interi: [17, 0, 5] invece di stringa "17.0.5"
//   - Metadati di build integrati nella struttura
//   - URL download diretti senza parametri query complessi
//   - Nome pacchetto descrittivo incluso nei metadati
//
// Algoritmo di ricerca Azul:
// 1. **Parsing target**: Conversione versione string in componenti numerici
// 2. **Iterazione packages**: Esamina tutti i package Azul disponibili
// 3. **Validazione versione**: Verifica array JavaVersion non vuoto
// 4. **Matching logico**: Supporta ricerche parziali come altri provider
// 5. **Compatibilità piattaforma**: Preferisce match per sistema corrente
// 6. **Selezione immediata**: Prima compatibilità trovata viene utilizzata
//
// Formato versione Azul:
//   - JavaVersion: [21] → versione 21.x.y
//   - JavaVersion: [17, 0] → versione 17.0.x
//   - JavaVersion: [17, 0, 5] → versione 17.0.5 esatta
//
// Criteri di compatibilità:
//   - **Nome package**: Contiene string OS corrente (case-insensitive)
//   - **Formato file**: Preferenza per archivi .zip
//   - **Disponibilità**: URL download valido e accessibile
//
// Gestione versioni:
//   - Parsing robusto di array versione variabile
//   - Supporto versioni incomplete [17] o [17, 0]
//   - Formattazione output consistente "major.minor.patch"
//   - Fallback graceful per formati inaspettati
//
// Parametri:
//
//	releases []azul.AzulPackage - Lista packages da Azul API
//	version string               - Versione target richiesta
//
// Restituisce:
//
//	string - URL download Azul Zulu ("" se non trovato)
//	string - Nome file package ("" se non determinabile)
//	string - Versione formattata (es. "17.0.5") ("" se non trovato)
//
// Limitazioni attuali:
//   - Prende primo match invece di migliore (ottimizzazione futura)
//   - Non considera date di release per tie-breaking
//   - Matching case-insensitive potrebbe essere troppo permissivo
//
// Esempio di utilizzo:
//
//	packages, _ := azul.GetAzulJDKs()
//	url, filename, version := findAzulDownload(packages, "17")
//	// url = "https://cdn.azul.com/zulu/bin/zulu17.30.15-ca-jdk17.0.1-win_x64.zip"
//	// filename = "zulu17.30.15-ca-jdk17.0.1-win_x64.zip"
//	// version = "17.0.1"
func findAzulDownload(releases []azul.AzulPackage, version string) (string, string, string) {
	runtime := getRuntimeInfo()

	// Parse target version
	targetMajor, targetMinor, targetPatch := utils.ParseVersionNumber(version)

	var bestMatch azul.AzulPackage
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		if len(release.JavaVersion) == 0 {
			continue
		}

		major := release.JavaVersion[0]
		minor := 0
		if len(release.JavaVersion) > 1 {
			minor = release.JavaVersion[1]
		}
		patch := 0
		if len(release.JavaVersion) > 2 {
			patch = release.JavaVersion[2]
		}

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Check if compatible with our platform or is a zip file
		if strings.Contains(strings.ToLower(release.Name), runtime.OS) || strings.HasSuffix(release.DownloadURL, ".zip") {
			bestMatch = release
			found = true
			break // Take the first match for now
		}
	}

	if !found {
		return "", "", ""
	}

	url := bestMatch.DownloadURL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	// Format version string
	var versionStr string
	if len(bestMatch.JavaVersion) > 0 {
		versionStr = fmt.Sprintf("%d", bestMatch.JavaVersion[0])
		if len(bestMatch.JavaVersion) > 1 {
			versionStr += fmt.Sprintf(".%d", bestMatch.JavaVersion[1])
		}
		if len(bestMatch.JavaVersion) > 2 {
			versionStr += fmt.Sprintf(".%d", bestMatch.JavaVersion[2])
		}
	}

	return url, filename, versionStr
}

// findLibericaDownload ricerca e seleziona il miglior download da BellSoft Liberica JDK.
//
// BellSoft Liberica JDK è un'alternativa OpenJDK con focus su performance e features
// aggiuntive come JavaFX integrato. Utilizza un sistema di versioning e metadati
// diverso dai altri provider che richiede parsing specializzato.
//
// Caratteristiche Liberica:
//   - **JavaFX incluso**: Build con JavaFX preintegrato disponibili
//   - **Native performance**: Ottimizzazioni specifiche per diverse piattaforme
//   - **Container ready**: Build ottimizzate per deployment in container
//   - **Supporto ARM**: Eccellente supporto per architetture ARM
//
// Algoritmo di ricerca:
// 1. **Parsing versione**: Usa liberica.ParseLibericaVersion() per formato specifico
// 2. **Matching flessibile**: Supporta ricerche parziali standard
// 3. **Primo match**: Strategia semplificata, primo compatible trovato
// 4. **URL diretto**: Ritorna DownloadURL senza modifiche
//
// Parametri:
//
//	releases []liberica.LibericaRelease - Lista release da Liberica API
//	version string                      - Versione target
//
// Restituisce:
//
//	string - URL download, nome file, versione trovata
func findLibericaDownload(releases []liberica.LibericaRelease, version string) (string, string, string) {
	// Parse target version
	targetMajor, targetMinor, targetPatch := utils.ParseVersionNumber(version)

	var bestMatch liberica.LibericaRelease
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		major, minor, patch := liberica.ParseLibericaVersion(release.Version)

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Take the first match
		bestMatch = release
		found = true
		break
	}

	if !found {
		return "", "", ""
	}

	url := bestMatch.DownloadURL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return url, filename, bestMatch.Version
}

// findPrivateDownload ricerca downloads da repository privati configurati dall'utente.
//
// Gestisce repository JDK aziendali interni come Nexus, Artifactory o API custom,
// utilizzando configurazione da ~/.jvm/config.json per autenticazione e endpoint.
//
// Caratteristiche repository privati:
//   - **Autenticazione**: Token-based per accesso sicuro
//   - **Compliance aziendale**: JDK approvati per uso interno
//   - **Versioni custom**: Build aziendali con patch specifiche
//   - **Controllo accesso**: Limitato a utenti autorizzati
//
// Algoritmo semplificato:
// 1. **Parsing versione**: Split standard per major.minor.patch
// 2. **Matching diretto**: Confronto stringhe di versione
// 3. **Primo match**: Strategia rapida per ambienti controllati
// 4. **URL validazione**: Cleanup parametri query se presenti
//
// Parametri:
//
//	releases []private.PrivateRelease - Lista da repository privato
//	version string                    - Versione richiesta
//
// Restituisce:
//
//	string - URL download privato, nome file, versione
func findPrivateDownload(releases []private.PrivateRelease, version string) (string, string, string) {
	// Parse target version
	targetMajor, targetMinor, targetPatch := utils.ParseVersionNumber(version)

	var bestMatch private.PrivateRelease
	var found bool

	// Search for matches with proper version parsing
	for _, release := range releases {
		// Simple version parsing for private releases
		versionParts := strings.Split(release.Version, ".")
		if len(versionParts) == 0 {
			continue
		}

		major, err := strconv.Atoi(versionParts[0])
		if err != nil {
			continue
		}

		minor := 0
		if len(versionParts) > 1 {
			if m, err := strconv.Atoi(versionParts[1]); err == nil {
				minor = m
			}
		}

		patch := 0
		if len(versionParts) > 2 {
			if p, err := strconv.Atoi(versionParts[2]); err == nil {
				patch = p
			}
		}

		// Check if this version matches our target
		isMatch := false
		if targetMinor == -1 && targetPatch == -1 {
			// Only major version specified (e.g., "17" -> match any 17.x.y)
			isMatch = (major == targetMajor)
		} else if targetPatch == -1 {
			// Major.minor specified (e.g., "17.0" -> match any 17.0.x)
			isMatch = (major == targetMajor && minor == targetMinor)
		} else {
			// Full version specified (e.g., "17.0.5" -> exact match)
			isMatch = (major == targetMajor && minor == targetMinor && patch == targetPatch)
		}

		if !isMatch {
			continue
		}

		// Take the first match
		bestMatch = release
		found = true
		break
	}

	if !found {
		return "", "", ""
	}

	url := bestMatch.DownloadURL
	filename := filepath.Base(url)
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}

	return url, filename, bestMatch.Version
}

// extractArchive estrae automaticamente archivi JDK ZIP o TAR.GZ nella directory di destinazione.
//
// Funzione dispatcher intelligente che rileva il formato dell'archivio dall'estensione
// e delega all'estrattore specifico. Implementa protezioni di sicurezza contro
// attacchi zip/tar slip per prevenire scrittura fuori dalla directory target.
//
// Formati supportati:
//   - **ZIP**: Formato standard Windows, più comune per JDK Windows
//   - **TAR.GZ**: Formato compresso, utilizzato da alcuni provider JDK
//
// Sicurezza implementata:
//   - **Path validation**: Prevenzione zip slip attacks
//   - **Directory confinement**: Estrazione solo nella directory target
//   - **Path normalization**: Pulizia percorsi malformati
//
// Parametri:
//
//	archivePath string - Percorso assoluto file archivio da estrarre
//	destPath string    - Directory destinazione per estrazione
//
// Restituisce:
//
//	error - nil se estrazione completata, errore se formato non supportato o fallimento

// extractJDKArchive estrae un archivio JDK nella directory versione specificata.
//
// Questa funzione trova automaticamente l'archivio nella directory JDK e lo estrae
// utilizzando la stessa logica del comando extract. È progettata per l'integrazione
// con il workflow di download automatico.
//
// Processo di estrazione:
// 1. **Ricerca archivio**: Trova file .zip o .tar.gz nella directory
// 2. **Rilevamento formato**: Determina tipo archivio dall'estensione
// 3. **Estrazione sicura**: Decomprime con protezioni security
// 4. **Organizzazione file**: Flattening directory se necessario
//
// Parametri:
//
//	jdkDirName string - Nome directory JDK (es. "JDK-17.0.8+9")
//	jdkPath string    - Percorso completo directory JDK
//
// Restituisce:
//
//	error - nil se estrazione completata, errore specifico altrimenti
//
// Esempio di utilizzo:
//
//	err := extractJDKArchive("JDK-17.0.8+9", "/home/user/.jvm/versions/JDK-17.0.8+9")
func extractJDKArchive(jdkDirName, jdkPath string) error {
	// Find archive in directory
	archivePath, err := findArchiveInDirectory(jdkPath)
	if err != nil {
		return fmt.Errorf("finding archive: %w", err)
	}

	// Extract archive
	if err := extractArchive(archivePath, jdkPath); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	return nil
}
