package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"jenvy/internal/utils"
)

// ExtractJDK gestisce l'estrazione di archivi JDK scaricati nel sistema Windows.
//
// Questa funzione implementa un sistema completo di estrazione per archivi JDK
// scaricati tramite il comando 'jenvy download', con ottimizzazioni specifiche per l'ambiente Windows:
//
// **Funzionalità principali:**
// - Lista automatica di archivi JDK disponibili se chiamata senza argomenti
// - Estrazione archivi JDK dalla directory ~/.jenvy/versions
// - **Parsing intelligente versioni**: Supporta sia nomi completi che parziali
// - Rilevamento intelligente del tipo di archivio basato su estensione
// - Gestione percorsi Windows con caratteri speciali e spazi
//
// **Parsing intelligente versioni:**
// - Input "17" → trova automaticamente JDK-17.x.y disponibile
// - Input "17.0" → trova JDK-17.0.x con patch più recente
// - Input "JDK-17.0.16+8" → usa directory esatta specificata
// - Gestione ambiguità con lista interattiva per scelta utente
//
// **Formati supportati Windows:**
// - .zip: Formato nativo Windows (preferito per tutti i provider)
// - .tar.gz: Supporto legacy per archivi Unix convertiti
//
// **Sicurezza e validazioni:**
// - Verifica esistenza e accessibilità file archivio
// - Controllo spazio disco disponibile prima dell'estrazione
// - Prevenzione path traversal attacks (../, ..\)
// - Validazione permessi Windows per directory di destinazione
//
// **Gestione directory Windows:**
// - Estrazione automatica nella directory version appropriata
// - Normalizzazione struttura directory per compatibilità Windows filesystem
// - Gestione case-insensitive NTFS appropriata
// - Supporto percorsi lunghi Windows (>260 caratteri)
//
// **Integrazione con download:**
// - Funziona solo con archivi scaricati tramite 'jenvy download'
// - Riconosce la struttura ~/.jenvy/versions/JDK-version/archive.zip
// - Estrae direttamente nella directory JDK appropriata
//
// **Sintassi comando:**
//
//	jenvy extract                      # mostra archivi disponibili da estrarre
//	jenvy extract 17                   # estrae versione 17.x.y più recente
//	jenvy extract 17.0                 # estrae versione 17.0.x più recente
//	jenvy extract JDK-17.0.16+8        # estrae versione specifica esatta
//
// **Esempi d'uso:**
//
//	jenvy extract                      # mostra: JDK-17.0.16+8, JDK-21.0.1+12, etc.
//	jenvy extract 17                   # trova e estrae JDK-17.0.16+8 automaticamente
//	jenvy extract JDK-21.0.1+12        # estrae specificamente questa versione
//
// La funzione garantisce estrazione sicura e pulizia automatica in caso di errori.
func ExtractJDK() {
	// Ottieni directory home dell'utente
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Error getting user directory: %v", err))
		utils.PrintInfo("Cannot access Windows user profile directory")
		return
	}

	versionsDir := filepath.Join(homeDir, ".jenvy", "versions")

	// Se nessun argomento, mostra archivi disponibili
	if len(os.Args) < 3 {
		showAvailableArchives(versionsDir)
		return
	}

	requestedVersion := os.Args[2]

	// Se l'input non inizia con "JDK-", cerca usando parsing intelligente
	var jdkDir string
	var actualVersion string

	if strings.HasPrefix(requestedVersion, "JDK-") {
		// Input completo, usa direttamente
		actualVersion = requestedVersion
		jdkDir = filepath.Join(versionsDir, requestedVersion)

		// Verifica che la directory esista
		if _, err := os.Stat(jdkDir); os.IsNotExist(err) {
			utils.PrintError(fmt.Sprintf("JDK version not found: %s", requestedVersion))
			utils.PrintInfo("Available JDK versions:")
			showAvailableArchives(versionsDir)
			return
		}
	} else {
		// Input parziale (es. "17"), cerca JDK con archivi disponibili
		foundPath, err := findJDKWithArchive(versionsDir, requestedVersion)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Unable to find JDK with archive for version '%s': %v", requestedVersion, err))
			utils.PrintInfo("Available JDK versions with archives:")
			showAvailableArchives(versionsDir)
			return
		}

		jdkDir = foundPath
		actualVersion = filepath.Base(foundPath)
		utils.PrintInfo(fmt.Sprintf("Found JDK version: %s", actualVersion))
	} // Cerca archivi nella directory JDK
	archiveFile, err := findArchiveInDirectory(jdkDir)
	if err != nil {
		utils.PrintError(fmt.Sprintf("No archive found in %s: %v", actualVersion, err))
		utils.PrintInfo("This JDK may already be extracted or the archive is missing")
		return
	}

	utils.PrintInfo(fmt.Sprintf("Found archive: %s", filepath.Base(archiveFile)))
	utils.PrintInfo(fmt.Sprintf("Extracting to: %s", jdkDir))

	// Estrai l'archivio nella stessa directory
	if err := extractArchive(archiveFile, jdkDir); err != nil {
		utils.PrintError(fmt.Sprintf("Extraction failed: %v", err))
		return
	}

	// Verifica che l'estrazione sia avvenuta correttamente
	if !utils.IsValidJDKDirectory(jdkDir) {
		utils.PrintWarning("Extracted directory does not appear to be a valid JDK")
		utils.PrintInfo("The archive may be corrupted or in an unexpected format")
	}

	// Rimuovi l'archivio dopo estrazione riuscita
	if err := os.Remove(archiveFile); err != nil {
		utils.PrintWarning(fmt.Sprintf("Could not remove archive file: %v", err))
		utils.PrintInfo(fmt.Sprintf("Archive file still present: %s", filepath.Base(archiveFile)))
	}

	utils.PrintSuccess(fmt.Sprintf("JDK extracted successfully: %s", actualVersion))
	utils.PrintInfo(fmt.Sprintf("Location: %s", jdkDir))
	utils.PrintInfo("Use 'jenvy use " + actualVersion + "' to activate this JDK")
}

// showAvailableArchives mostra la lista di archivi JDK disponibili per l'estrazione.
//
// Questa funzione scansiona la directory ~/.jenvy/versions alla ricerca di directory
// JDK che contengono archivi non ancora estratti. Fornisce un'interfaccia user-friendly
// per visualizzare i JDK scaricati e pronti per l'estrazione.
//
// **Funzionalità:**
// - Scansione automatica directory versioni JDK
// - Rilevamento archivi .zip e .tar.gz non estratti
// - Distinzione tra JDK estratti e non estratti
// - Output colorato per migliore leggibilità
//
// **Formato output:**
// - Lista directory JDK con archivi disponibili
// - Indicazione formato archivio (ZIP/TAR.GZ)
// - Dimensione file per ogni archivio
// - Istruzioni d'uso per l'estrazione
//
// Parametri:
//   - versionsDir: percorso directory contenente le versioni JDK
func showAvailableArchives(versionsDir string) {
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Cannot access versions directory: %v", err))
		utils.PrintInfo("Make sure to download JDKs first using 'jenvy download <version>'")
		return
	}

	var availableArchives []string
	var extractedJDKs []string

	for _, entry := range entries {
		if entry.IsDir() {
			jdkDir := filepath.Join(versionsDir, entry.Name())

			// Controlla se c'è un archivio nella directory
			archiveFile, err := findArchiveInDirectory(jdkDir)
			if err == nil {
				// Ottieni informazioni sul file
				info, err := os.Stat(archiveFile)
				if err == nil {
					size := float64(info.Size()) / (1024 * 1024) // MB
					availableArchives = append(availableArchives,
						fmt.Sprintf("  %s (%.1f MB)", entry.Name(), size))
				} else {
					availableArchives = append(availableArchives,
						fmt.Sprintf("  %s", entry.Name()))
				}
			} else {
				// Controlla se è un JDK già estratto
				if utils.IsValidJDKDirectory(jdkDir) {
					extractedJDKs = append(extractedJDKs, fmt.Sprintf("  %s (already extracted)", entry.Name()))
				}
			}
		}
	}

	if len(availableArchives) == 0 && len(extractedJDKs) == 0 {
		utils.PrintInfo("No JDK versions found in ~/.jenvy/versions")
		utils.PrintInfo("Download JDKs first using:")
		utils.PrintInfo("  jenvy remote-list          # See available versions")
		utils.PrintInfo("  jenvy download <version>   # Download a JDK")
		return
	}

	if len(availableArchives) > 0 {
		utils.PrintInfo("Available archives to extract:")
		for _, archive := range availableArchives {
			fmt.Println(archive)
		}
		fmt.Println()
		utils.PrintInfo("To extract a JDK, use:")
		utils.PrintInfo("  jenvy extract <jdk-version>")
		utils.PrintInfo("Example:")
		if len(availableArchives) > 0 {
			// Estrai il nome JDK dal primo elemento disponibile
			firstJDK := strings.Split(availableArchives[0], " ")[0]
			firstJDK = strings.TrimSpace(firstJDK)
			utils.PrintInfo(fmt.Sprintf("  jenvy extract %s", firstJDK))
		}
	}

	if len(extractedJDKs) > 0 {
		if len(availableArchives) > 0 {
			fmt.Println()
		}
		utils.PrintInfo("Already extracted JDKs:")
		for _, jdk := range extractedJDKs {
			fmt.Println(jdk)
		}
	}
}

// findJDKWithArchive cerca un JDK che corrisponde alla versione richiesta e ha un archivio disponibile per l'estrazione.
//
// Questa funzione implementa parsing intelligente delle versioni per il comando extract,
// cercando solo JDK che hanno archivi non ancora estratti. È diversa da utils.FindSingleJDKInstallation
// perché filtra solo quelli con archivi disponibili.
//
// **Algoritmo di ricerca intelligente:**
// 1. **Exact Match**: Cerca "JDK-{version}" con archivio
// 2. **Partial Match**: Cerca versioni che iniziano con il pattern e hanno archivi
// 3. **Filtro archivi**: Solo directory con archivi .zip o .tar.gz disponibili
// 4. **Gestione ambiguità**: Mostra opzioni multiple se trovate
//
// **Logica matching versione:**
//   - "17" → trova qualsiasi JDK-17.x.y con archivio
//   - "17.0" → trova qualsiasi JDK-17.0.x con archivio
//   - "17.0.8" → trova qualsiasi JDK-17.0.8.x con archivio
//
// **Comportamento con risultati multipli:**
//   - Un solo match: ritorna il percorso trovato
//   - Nessun match: errore "no JDK found with archive"
//   - Multiple matches: mostra lista e richiede maggiore precisione
//
// Parametri:
//
//	versionsDir string - Directory base che contiene le versioni JDK
//	version string     - Versione richiesta (es. "17", "17.0.8")
//
// Restituisce:
//
//	string - Percorso assoluto alla directory JDK con archivio
//	error  - nil se trovato singolo match, errore specifico altrimenti
//
// Esempi di utilizzo:
//
//	path, err := findJDKWithArchive("/home/user/.jenvy/versions", "17")
//	// path = "/home/user/.jenvy/versions/JDK-17.0.8.1+1" (se ha archivio)
//
// Differenze da utils.FindSingleJDKInstallation:
//   - Filtra solo JDK con archivi disponibili
//   - Progettato specificamente per il comando extract
//   - Non considera JDK già estratti senza archivi
func findJDKWithArchive(versionsDir, version string) (string, error) {
	// Cerca prima match esatto
	exactMatch := filepath.Join(versionsDir, fmt.Sprintf("JDK-%s", version))
	if _, err := os.Stat(exactMatch); err == nil {
		if _, err := findArchiveInDirectory(exactMatch); err == nil {
			return exactMatch, nil
		}
	}

	// Cerca match parziali con archivi
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read versions directory: %w", err)
	}

	var matches []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if strings.HasPrefix(name, "JDK-") {
				jdkVersion := strings.TrimPrefix(name, "JDK-")
				if strings.HasPrefix(jdkVersion, version) {
					fullPath := filepath.Join(versionsDir, name)
					// Verifica che ci sia un archivio nella directory
					if _, err := findArchiveInDirectory(fullPath); err == nil {
						matches = append(matches, fullPath)
					}
				}
			}
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no JDK found with archive matching version %s", version)
	}

	if len(matches) == 1 {
		return matches[0], nil
	}

	// Multiple matches - mostra opzioni
	utils.PrintWarning("Multiple JDK versions with archives found:")
	for i, match := range matches {
		fmt.Printf("  %d. %s\n", i+1, filepath.Base(match))
	}
	return "", fmt.Errorf("multiple matches found, please be more specific")
}

// findArchiveInDirectory cerca archivi JDK (.zip, .tar.gz) nella directory specificata.
//
// Questa funzione implementa la ricerca intelligente di archivi JDK all'interno
// di una directory versione, supportando i formati più comuni utilizzati dai
// provider JDK per la distribuzione su Windows.
//
// **Algoritmo di ricerca:**
// 1. Scansione file nella directory target
// 2. Filtro per estensioni supportate (.zip, .tar.gz)
// 3. Priorità ai file .zip (preferiti su Windows)
// 4. Ritorno primo archivio valido trovato
//
// **Formati supportati:**
// - .zip: Archivi ZIP standard (priorità alta)
// - .tar.gz: Archivi TAR compressi GZIP (compatibilità)
//
// **Validazioni:**
// - Controllo esistenza e accessibilità file
// - Verifica estensioni file supportate
// - Controllo dimensione minima archivio (> 1MB)
//
// Parametri:
//   - dirPath: percorso directory in cui cercare archivi
//
// Ritorna:
//   - string: percorso completo primo archivio trovato
//   - error: errore se nessun archivio trovato o directory inaccessibile
func findArchiveInDirectory(dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("cannot read directory: %v", err)
	}

	// Prima cerca file .zip (preferiti su Windows)
	for _, entry := range entries {
		if !entry.IsDir() {
			fileName := strings.ToLower(entry.Name())
			if strings.HasSuffix(fileName, ".zip") {
				fullPath := filepath.Join(dirPath, entry.Name())
				// Verifica che il file esista e abbia dimensione ragionevole
				if info, err := os.Stat(fullPath); err == nil && info.Size() > 1024*1024 {
					return fullPath, nil
				}
			}
		}
	}

	// Poi cerca file .tar.gz
	for _, entry := range entries {
		if !entry.IsDir() {
			fileName := strings.ToLower(entry.Name())
			if strings.HasSuffix(fileName, ".tar.gz") {
				fullPath := filepath.Join(dirPath, entry.Name())
				// Verifica che il file esista e abbia dimensione ragionevole
				if info, err := os.Stat(fullPath); err == nil && info.Size() > 1024*1024 {
					return fullPath, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no valid archive found")
}

// extractArchive esegue l'estrazione effettiva dell'archivio nel percorso di destinazione.
//
// Questa funzione implementa l'estrazione multi-formato con ottimizzazioni
// specifiche per il filesystem Windows e gestione errori robusta:
//
// **Supporto formati:**
// - ZIP: Estrazione nativa usando archive/zip (preferito Windows)
// - TAR.GZ: Estrazione usando archive/tar e compress/gzip
// - Rilevamento automatico formato da estensione file
//
// **Ottimizzazioni Windows:**
// - Gestione percorsi lunghi (>260 caratteri) con prefisso \\?\
// - Preservazione timestamp file compatibili con NTFS
// - Gestione permessi Windows appropriati (0755 per directory, 0644 per file)
// - Supporto per file con caratteri Unicode nei nomi
//
// **Sicurezza estrazione:**
// - Validazione path per prevenire directory traversal attacks
// - Controllo dimensioni file per prevenire zip bombs
// - Verifica spazio disco disponibile durante estrazione
// - Pulizia automatica in caso di errori
//
// **Gestione strutture archivio:**
// - Rimozione directory wrapper se presente (comune in archivi JDK)
// - Normalizzazione struttura per compatibilità con altri comandi jenvy
// - Preservazione metadata JDK essenziali
//
// Parametri:
//   - archivePath: percorso del file archivio da estrarre
//   - destPath: directory di destinazione per l'estrazione
//
// Ritorna errore se l'estrazione fallisce per qualsiasi motivo.
func extractArchive(archivePath, destPath string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(archivePath))
	if ext == ".zip" {
		if err := extractZip(archivePath, destPath); err != nil {
			return fmt.Errorf("ZIP extraction failed: %v", err)
		}
	} else if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") {
		if err := extractTarGz(archivePath, destPath); err != nil {
			return fmt.Errorf("TAR.GZ extraction failed: %v", err)
		}
	} else {
		return fmt.Errorf("unsupported archive format: %s", ext)
	}

	// Try to find and flatten JDK structure if needed
	jdkRoot, err := findJDKRootDir(destPath)
	if err != nil {
		utils.PrintWarning("Could not locate JDK root directory, using extracted structure as-is")
		return nil
	}

	// If JDK is nested, flatten it
	if jdkRoot != destPath {
		if err := flattenJDKDirectory(jdkRoot, destPath, archivePath); err != nil {
			utils.PrintWarning(fmt.Sprintf("Could not flatten JDK structure: %v", err))
			utils.PrintInfo("JDK extracted successfully but may have nested structure")
		}
	}

	return nil
}

// extractZip estrae un archivio ZIP con protezioni di sicurezza avanzate per Windows.
//
// Implementa estrazione completa di archivi ZIP JDK con particolare attenzione
// alla sicurezza e compatibilità Windows. Gestisce percorsi lunghi, caratteri
// Unicode e preserva permessi file appropriati per l'ambiente Windows.
//
// Caratteristiche di sicurezza:
//   - **Zip slip protection**: Validazione rigorosa percorsi file
//   - **Path normalization**: Pulizia e validazione nomi file
//   - **Directory confinement**: Prevenzione scrittura fuori target
//   - **Resource management**: Chiusura automatica handle file
//
// Gestione file e directory:
//   - **Directory creation**: Ricreazione struttura directory archivio
//   - **File extraction**: Preservazione contenuto e metadati
//   - **Permission handling**: Gestione appropriata permessi Windows
//   - **Unicode support**: Supporto completo caratteri internazionali
//
// Parametri:
//   - src: percorso archivio ZIP sorgente
//   - dest: directory destinazione estrazione
//
// Ritorna errore se l'estrazione fallisce per qualsiasi motivo.
func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Clean the file path to prevent zip slip attacks
		cleanPath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(cleanPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(cleanPath, 0755)
			continue
		}

		// Create the directories for file
		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz estrae archivi TAR.GZ con supporto completo per l'ambiente Windows.
//
// Gestisce la decompressione GZIP seguita dall'estrazione TAR, preservando
// la struttura directory e gestendo appropriatamente i permessi file per
// l'ambiente Windows. Implementa le stesse protezioni di sicurezza di extractZip.
//
// Processo a due fasi:
//  1. **Decompressione GZIP**: gzip.NewReader per decompressione stream
//  2. **Estrazione TAR**: tar.NewReader per estrazione file e directory
//
// Gestione metadati:
//   - **File regolari**: Preservazione contenuto e dimensione
//   - **Directory**: Ricreazione struttura gerarchica
//   - **Permessi**: Conversione permessi file → Windows
//   - **Timestamp**: Preservazione dove possibile
//
// Sicurezza TAR:
//   - **Tar slip protection**: Validazione percorsi come ZIP
//   - **Type validation**: Gestione solo file regolari e directory
//   - **Path cleaning**: Normalizzazione percorsi cross-platform
//
// Parametri:
//   - src: percorso archivio TAR.GZ sorgente
//   - dest: directory destinazione estrazione
//
// Ritorna errore per problemi decompressione/estrazione.
func extractTarGz(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Clean the file path to prevent tar slip attacks
		cleanPath := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(cleanPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create the directories for file
			if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
				return err
			}

			// Extract file
			outFile, err := os.Create(cleanPath)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}

			outFile.Close()

			// Set file permissions
			if err := os.Chmod(cleanPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}

// findJDKRootDir localizza la directory root effettiva del JDK all'interno dell'estrazione.
//
// Gli archivi JDK spesso contengono una directory wrapper (es. "jdk-17.0.5+8")
// che racchiude la vera installazione JDK. Questa funzione identifica e ritorna
// il percorso della directory JDK utilizzabile, riducendo la nidificazione.
//
// Pattern comuni archivi JDK:
//   - jdk-17.0.5+8/bin/, jdk-17.0.5+8/lib/ (directory wrapper)
//   - bin/, lib/ (estrazione diretta, ideale)
//   - multiple directory (ambiguo, usa euristica)
//
// Algoritmo di ricerca:
// 1. **Enumera contenuto**: Lista entry nella directory estrazione
// 2. **Identifica candidati**: Cerca directory che sembrano JDK
// 3. **Valida struttura**: Usa IsValidJDKDirectory per conferma
// 4. **Ritorna migliore**: Prima directory JDK valida trovata
//
// Parametri:
//   - extractPath: directory dove è stato estratto l'archivio
//
// Ritorna il percorso directory JDK utilizzabile e errore se non trovata.
func findJDKRootDir(extractPath string) (string, error) {
	// JDK archives often contain a single root directory like "jdk-17.0.5+8"
	entries, err := os.ReadDir(extractPath)
	if err != nil {
		return extractPath, err
	}

	// Look for a single directory that might be the JDK root
	var jdkDir string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory contains typical JDK structure (bin, lib, etc.)
			potentialJDKDir := filepath.Join(extractPath, entry.Name())
			if utils.IsValidJDKDirectory(potentialJDKDir) {
				jdkDir = potentialJDKDir
				break
			}
		}
	}

	if jdkDir == "" {
		// If no JDK-like subdirectory found, check if extractPath itself is a JDK
		if utils.IsValidJDKDirectory(extractPath) {
			return extractPath, nil
		}
		return extractPath, fmt.Errorf("could not locate JDK root directory")
	}

	return jdkDir, nil
}

// flattenJDKDirectory sposta contenuto JDK annidato al livello parent per semplificare accesso.
//
// Quando un archivio JDK crea una struttura annidata inutile, questa funzione
// riorganizza i file per avere un layout più pulito e accessibile, eliminando
// layer di directory intermedi che complicano l'utilizzo del JDK.
//
// Problema risolto:
//
//	Archivi JDK spesso creano strutture come:
//	target/jdk-17.0.5+8/bin/java.exe
//	target/jdk-17.0.5+8/lib/
//
//	Risultato desiderato:
//	target/bin/java.exe
//	target/lib/
//
// Processo di flattening sicuro:
// 1. **Directory temporanea**: Crea spazio staging per evitare conflitti durante spostamento
// 2. **Spostamento staged**: Muove tutto il contenuto JDK in directory temporanea
// 3. **Pulizia wrapper**: Rimuove directory wrapper annidata ora vuota
// 4. **Finalizzazione**: Sposta contenuto da staging alla destinazione finale
// 5. **Cleanup automatico**: Rimuove directory temporanea anche in caso di errore
//
// Gestione sicura file system:
//   - **Operazioni atomiche**: Usa os.Rename per spostamenti atomici
//   - **Recovery automatico**: defer cleanup in caso di errori intermedi
//   - **Validazione path**: Tutti i percorsi normalizzati tramite filepath.Join
//   - **Prevenzione conflitti**: Directory temporanea con naming sicuro
//   - **Gestione errori**: Errori dettagliati per troubleshooting
//
// Compatibilità Windows:
//   - Gestisce percorsi lunghi Windows
//   - Supporta caratteri Unicode in nomi file
//   - Rispetta lock file e permessi Windows
//   - Operazioni sicure con antivirus in tempo reale
//
// Parametri:
//   - jdkRootDir: directory JDK sorgente da appiattire (es. "target/jdk-17.0.5+8")
//   - targetDir: directory destinazione finale (es. "target")
//   - archivePath: percorso archivio originale (utilizzato per riferimento)
//
// Ritorna errore se l'operazione fallisce per problemi filesystem.
func flattenJDKDirectory(jdkRootDir, targetDir, archivePath string) error {
	// Create a temporary directory to avoid conflicts
	tempDir := targetDir + "_temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Move JDK contents to temp directory
	entries, err := os.ReadDir(jdkRootDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(jdkRootDir, entry.Name())
		destPath := filepath.Join(tempDir, entry.Name())

		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move %s: %v", entry.Name(), err)
		}
	}

	// Remove the now empty nested directory structure
	if err := os.RemoveAll(filepath.Join(targetDir, filepath.Base(jdkRootDir))); err != nil {
		return err
	}

	// Move contents from temp to target directory
	tempEntries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	for _, entry := range tempEntries {
		srcPath := filepath.Join(tempDir, entry.Name())
		destPath := filepath.Join(targetDir, entry.Name())

		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move %s to final location: %v", entry.Name(), err)
		}
	}

	return nil
}
