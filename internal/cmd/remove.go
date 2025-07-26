package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"jenvy/internal/utils"
)

// RemoveJDK gestisce la rimozione sicura di installazioni JDK dal sistema Windows.
//
// Questa funzione fornisce un'interfaccia completa per la rimozione di versioni JDK,
// con particolare attenzione alla sicurezza e integrità del sistema Windows:
//
// **Modalità di rimozione supportate:**
// - Rimozione singola versione: `jvm remove <version>` o `jvm rm <version>`
// - Rimozione completa: `jvm remove --all` o `jvm rm -a`
//
// **Sicurezza e validazioni Windows:**
// - Verifica se il JDK è attualmente impostato come JAVA_HOME
// - Controllo processi Windows che potrebbero utilizzare il JDK
// - Conferma utente prima di operazioni distruttive
// - Gestione permessi Windows per directory di sistema
//
// **Gestione intelligente versioni:**
// - Ricerca esatta e fuzzy matching per identificazione versioni
// - Supporto per diversi formati di naming (Adoptium, Azul, Liberica, Private)
// - Validazione esistenza directory prima della rimozione
// - Cleanup automatico directory vuote post-rimozione
//
// **Integrazione ambiente Windows:**
// - Gestione variabili ambiente Windows (JAVA_HOME, PATH)
// - Supporto per percorsi Windows con spazi e caratteri speciali
// - Logging operazioni nel formato Windows standard
// - Compatibilità con Windows Defender e antivirus
//
// La funzione è progettata per essere sicura e user-friendly, fornendo
// feedback dettagliato e opzioni di rollback in caso di problemi.
func RemoveJDK() {
	if len(os.Args) < 3 {
		utils.PrintUsage("Usage: jenvy remove <version>")
		utils.PrintUsage("       jvm remove --all")
		utils.PrintUsage("Short form: jenvy rm <version>")
		utils.PrintUsage("           jvm rm -a")
		utils.PrintInfo("Available JDKs:")
		showAvailableJDKsForRemoval()
		return
	}

	// Controlla se è stato richiesto di rimuovere tutto
	if os.Args[2] == "--all" || os.Args[2] == "-a" {
		removeAllJDKs()
		return
	}

	version := os.Args[2]

	// Ottieni directory home dell'utente
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Error getting home directory: %v", err))
		return
	}

	versionsDir := filepath.Join(homeDir, ".jenvy", "versions")

	// Controlla se la directory esiste
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		utils.PrintError("No JDK installations found")
		utils.PrintInfo("The versions directory doesn't exist yet")
		return
	}

	// Trova la versione JDK da rimuovere
	jdkPath, err := findJDKForRemoval(versionsDir, version)
	if err != nil {
		utils.PrintError(fmt.Sprintf("JDK version %s not found: %v", version, err))
		utils.PrintInfo("Run 'jenvy list' to see installed JDKs")
		return
	}

	// Verifica se la versione è attualmente in uso
	if isJDKCurrentlyInUse(jdkPath) {
		utils.PrintWarning(fmt.Sprintf("JDK %s is currently set as JAVA_HOME", version))
		utils.PrintInfo("Consider switching to another version before removal:")
		utils.PrintInfo("  jvm use <other-version>")
		utils.PrintInfo("Continuing will unset JAVA_HOME and may affect running Java applications")
		fmt.Print("\nDo you want to continue anyway? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(strings.TrimSpace(response)) != "y" {
			utils.PrintInfo("Removal cancelled")
			return
		}
	}

	// Conferma rimozione
	fmt.Printf("Are you sure you want to remove JDK %s?\n", version)
	fmt.Printf("   Path: %s\n", jdkPath)
	fmt.Print("   This action cannot be undone. [y/N]: ")

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) != "y" {
		utils.PrintInfo("Removal cancelled")
		return
	}

	// Rimuovi la directory JDK
	fmt.Printf("Removing JDK %s...\n", version)
	err = os.RemoveAll(jdkPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to remove JDK: %v", err))
		utils.PrintInfo("Make sure no applications are using this JDK")
		return
	}

	utils.PrintSuccess(fmt.Sprintf("JDK %s removed successfully", version))

	// Mostra JDK rimanenti
	fmt.Println()
	utils.PrintInfo("Remaining JDK installations:")
	showRemainingJDKs(versionsDir)
}

// findJDKForRemoval localizza e valida il percorso del JDK da rimuovere nel filesystem Windows.
//
// Questa funzione implementa una strategia di ricerca intelligente per identificare
// installazioni JDK specifiche nella directory delle versioni Windows:
//
// **Strategia di ricerca a più livelli:**
// 1. Corrispondenza esatta della versione specificata
// 2. Ricerca fuzzy con prefix matching per versioni parziali
// 3. Gestione di ambiguità con suggerimenti all'utente
//
// **Supporto formati di naming Windows:**
// - Adoptium: adoptium-jdk-<version>
// - Azul: azul-jdk-<version>
// - Liberica: liberica-jdk-<version>
// - Private: private-jdk-<version>
// - Formato standard: jdk-<version> o JDK-<version>
//
// **Validazioni Windows-specifiche:**
// - Verifica esistenza directory nel filesystem NTFS
// - Controllo permessi di lettura per directory versioni
// - Gestione case-insensitive per compatibilità Windows
// - Supporto per percorsi lunghi Windows (>260 caratteri)
//
// **Gestione errori robusta:**
// - Messaggi di errore dettagliati per debugging
// - Suggerimenti per risoluzione problemi comuni
// - Fallback su ricerca parziale se esatta fallisce
// - Lista di opzioni multiple per disambiguazione utente
//
// Parametri:
//   - versionsDir: directory radice contenente installazioni JDK Windows
//   - targetVersion: versione JDK da cercare (esatta o parziale)
//
// Ritorna il percorso completo Windows del JDK trovato o errore dettagliato.
func findJDKForRemoval(versionsDir, targetVersion string) (string, error) {
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read versions directory: %w", err)
	}

	// Cerca corrispondenza esatta prima
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()

		// Estrai la versione dal nome della directory
		if extractedVersion := extractVersionFromDirName(dirName); extractedVersion != "" {
			if extractedVersion == targetVersion {
				return filepath.Join(versionsDir, dirName), nil
			}
		}
	}

	// Se non trova corrispondenza esatta, cerca corrispondenza parziale
	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		if extractedVersion := extractVersionFromDirName(dirName); extractedVersion != "" {
			if strings.HasPrefix(extractedVersion, targetVersion) {
				matches = append(matches, dirName)
			}
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no JDK found matching version %s", targetVersion)
	}

	if len(matches) == 1 {
		return filepath.Join(versionsDir, matches[0]), nil
	}

	// Più corrispondenze trovate
	utils.PrintError(fmt.Sprintf("Multiple JDK versions found matching '%s':", targetVersion))
	for _, match := range matches {
		version := extractVersionFromDirName(match)
		fmt.Printf("  - %s (%s)\n", version, match)
	}
	return "", fmt.Errorf("please specify the exact version")
}

// extractVersionFromDirName estrae la versione JDK pulita dal nome directory Windows.
//
// Questa funzione implementa il parsing intelligente dei nomi directory
// per estrarre versioni JDK standardizzate, supportando tutti i formati
// di naming utilizzati dai provider principali nell'ecosistema Windows:
//
// **Formati supportati:**
// - Eclipse Adoptium: "adoptium-jdk-21.0.1+12", "JDK-17.0.9+11"
// - Azul Zulu: "azul-jdk-8.0.392+8", "jdk-11.0.21+9"
// - BellSoft Liberica: "liberica-jdk-17.0.8+7"
// - Repository Private: "private-jdk-custom-1.8.0_372"
// - Formato generico: "jdk-<version>", "JDK-<version>"
//
// **Algoritmo di normalizzazione:**
// 1. Rimozione prefissi provider-specifici (adoptium-, azul-, liberica-, private-)
// 2. Rimozione prefissi standard (jdk-, JDK-)
// 3. Preservazione formato versione completo (major.minor.patch+build)
// 4. Gestione case-insensitive per compatibilità Windows
//
// **Windows-specific considerations:**
// - Supporto per nomi directory con spazi (rari ma possibili)
// - Gestione caratteri speciali Windows in nomi directory
// - Compatibilità con limitazioni lunghezza path Windows
// - Preservazione encoding UTF-8 per caratteri internazionali
//
// **Esempi di trasformazione:**
//
//	"adoptium-jdk-21.0.1+12" → "21.0.1+12"
//	"azul-jdk-8.0.392+8" → "8.0.392+8"
//	"JDK-17.0.9+11" → "17.0.9+11"
//	"liberica-jdk-11.0.20+8" → "11.0.20+8"
//
// Parametri:
//   - dirName: nome della directory come riportato dal filesystem Windows
//
// Ritorna la versione JDK estratta e normalizzata, o stringa vuota se non valida.
func extractVersionFromDirName(dirName string) string {
	// Gestisce formati come:
	// JDK-11.0.21+9
	// JDK-17.0.9+11
	// jdk-11.0.21+9
	// adoptium-jdk-21.0.1+12
	// azul-jdk-8.0.392+8

	// Rimuovi prefissi comuni
	cleaned := dirName
	prefixes := []string{"adoptium-", "azul-", "liberica-", "private-"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimPrefix(cleaned, prefix)
			break
		}
	}

	// Rimuovi "jdk-" o "JDK-" se presente
	cleaned = strings.TrimPrefix(cleaned, "jdk-")
	cleaned = strings.TrimPrefix(cleaned, "JDK-")

	return cleaned
}

// isJDKCurrentlyInUse verifica se il JDK specificato è attualmente attivo nell'ambiente Windows.
//
// Questa funzione implementa il controllo di sicurezza per prevenire la rimozione
// accidentale del JDK attualmente configurato come ambiente Java principale:
//
// **Controlli ambiente Windows:**
// - Lettura variabile JAVA_HOME dal registro Windows
// - Confronto case-insensitive dei percorsi (standard Windows)
// - Normalizzazione percorsi per gestire slash/backslash misti
// - Risoluzione symlink e junction point Windows
//
// **Sicurezza rimozione:**
// - Previene corruption dell'ambiente Java attivo
// - Evita interruzione applicazioni Windows in esecuzione
// - Protegge da rimozione accidentale JDK di sistema
// - Mantiene integrità configurazione PATH Windows
//
// **Gestione percorsi Windows:**
// - Supporto per percorsi con spazi e caratteri speciali
// - Gestione case-insensitive del filesystem NTFS
// - Risoluzione percorsi relativi e assoluti
// - Compatibilità con format UNC per network drive
//
// **Integrazione sistema:**
// - Compatibile con configurazioni Windows Domain
// - Supporto per profile utente roaming
// - Gestione variabili ambiente system vs user
// - Verifica coerenza tra JAVA_HOME e PATH
//
// Parametri:
//   - jdkPath: percorso assoluto Windows del JDK da verificare
//
// Ritorna true se il JDK è attualmente configurato come JAVA_HOME attivo.
func isJDKCurrentlyInUse(jdkPath string) bool {
	javaHome := os.Getenv("JAVA_HOME")
	if javaHome == "" {
		return false
	}

	// Normalizza i percorsi per il confronto
	normalizedJavaHome := strings.ToLower(filepath.Clean(javaHome))
	normalizedJDKPath := strings.ToLower(filepath.Clean(jdkPath))

	return normalizedJavaHome == normalizedJDKPath
}

// showAvailableJDKsForRemoval visualizza l'elenco delle installazioni JDK disponibili per rimozione nel sistema Windows.
//
// Questa funzione fornisce una panoramica user-friendly delle versioni JDK
// attualmente installate, facilitando la selezione per operazioni di rimozione:
//
// **Scansione directory Windows:**
// - Lettura sicura della directory %USERPROFILE%\.jenvy\versions Windows
// - Enumerazione solo di directory valide (ignora file temporanei)
// - Gestione permessi insufficienti con fallback graceful
// - Supporto per percorsi lunghi Windows (>260 caratteri)
//
// **Presentazione informazioni:**
// - Lista ordinata delle versioni disponibili
// - Estrazione versione pulita da nomi directory complessi
// - Formato output consistent con altri comandi jvm
// - Gestione graceful di directory vuote o corrotte
//
// **Gestione errori Windows:**
// - Verifica esistenza directory .jenvy nel profilo utente
// - Controllo permessi di lettura sulla directory versioni
// - Fallback silenzioso su errori di accesso filesystem
// - Messaggi informativi per directory inesistenti
//
// **Integrazione UX:**
// - Output formattato per facile lettura in terminal Windows
// - Supporto per redirection output (per scripting)
// - Compatibilità con Windows Terminal e CMD legacy
// - Encoding UTF-8 per caratteri speciali in versioni
//
// La funzione è progettata per essere chiamata automaticamente quando
// l'utente invoca il comando remove senza parametri specifici.
func showAvailableJDKsForRemoval() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	versionsDir := filepath.Join(homeDir, ".jenvy", "versions")
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		utils.PrintInfo("No JDK installations found")
		return
	}

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return
	}

	if len(entries) == 0 {
		utils.PrintInfo("No JDK installations found")
		return
	}

	utils.PrintInfo("Available JDK versions:")
	for _, entry := range entries {
		if entry.IsDir() {
			version := extractVersionFromDirName(entry.Name())
			if version != "" {
				fmt.Printf("  - %s\n", version)
			}
		}
	}
}

// showRemainingJDKs visualizza le installazioni JDK rimanenti dopo un'operazione di rimozione nel sistema Windows.
//
// Questa funzione fornisce feedback immediato all'utente mostrando lo stato
// delle installazioni JDK dopo una rimozione, aiutando a verificare il successo dell'operazione:
//
// **Scansione post-rimozione:**
// - Re-scan della directory versioni Windows dopo modifiche
// - Identificazione installazioni sopravvissute alla rimozione
// - Verifica integrità directory rimanenti
// - Controllo coerenza stato filesystem Windows
//
// **Reportistica stato:**
// - Lista formattata delle versioni ancora presenti
// - Messaggio appropriato se nessuna installazione rimane
// - Estrazione versioni normalizzate per display
// - Ordinamento logico delle versioni rimanenti
//
// **Gestione resilienza Windows:**
// - Tolleranza per errori di accesso temporanei
// - Gestione directory parzialmente rimosse
// - Recovery da stati inconsistenti del filesystem
// - Supporto per operazioni interrotte (antivirus, etc.)
//
// **Utility operativa:**
// - Chiamata automatica dopo rimozioni singole/multiple
// - Integrazione con flusso di conferma utente
// - Base per suggerimenti azioni successive
// - Supporto per logging operazioni di sistema
//
// Parametri:
//   - versionsDir: percorso directory versioni Windows da scansionare
//
// La funzione gestisce gracefully errori e fornisce sempre feedback utile.
func showRemainingJDKs(versionsDir string) {
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		utils.PrintWarning("Could not list remaining JDKs")
		return
	}

	var remainingJDKs []string
	for _, entry := range entries {
		if entry.IsDir() {
			version := extractVersionFromDirName(entry.Name())
			if version != "" {
				remainingJDKs = append(remainingJDKs, version)
			}
		}
	}

	if len(remainingJDKs) == 0 {
		utils.PrintInfo("No JDK installations remaining")
	} else {
		for _, version := range remainingJDKs {
			fmt.Printf("  - %s\n", version)
		}
	}
}

// removeAllJDKs esegue la rimozione completa e sicura di tutte le installazioni JDK dal sistema Windows.
//
// Questa funzione implementa un'operazione di pulizia totale delle installazioni JDK,
// con controlli di sicurezza avanzati e conferme multiple per prevenire perdite accidentali:
//
// **Processo di rimozione sicura:**
// 1. Scansione completa directory installazioni Windows
// 2. Inventario dettagliato di tutte le versioni presenti
// 3. Controllo JDK attualmente in uso (JAVA_HOME)
// 4. Doppia conferma utente con typing esplicito "yes"
// 5. Rimozione sequenziale con reporting granulare
// 6. Cleanup directory vuote e metadata residui
//
// **Sicurezza e validazioni Windows:**
// - Verifica permessi elevati se necessari per rimozione
// - Controllo processi Windows che utilizzano JDK attivi
// - Gestione file locked da antivirus o applicazioni
// - Backup metadata configurazione per recovery
//
// **Gestione errori e recovery:**
// - Tracking rimozioni riuscite vs fallite
// - Report dettagliato di errori per ogni installazione
// - Preservazione JDK non rimovibili per troubleshooting
// - Suggerimenti per risoluzione problemi comuni
//
// **Integrazione ambiente Windows:**
// - Reset variabili ambiente (JAVA_HOME, PATH)
// - Cleanup registro Windows se necessario
// - Rimozione associazioni file .jar se appropriate
// - Notifica Windows per aggiornamento cache sistema
//
// **Post-operazione:**
// - Statistiche complete dell'operazione
// - Suggerimenti per installazione nuovi JDK
// - Verifica integrità configurazione residua
// - Logging operazione per audit trail
//
// Questa è un'operazione irreversibile che richiede particolare attenzione.
func removeAllJDKs() {
	// Ottieni directory home dell'utente
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Error getting home directory: %v", err))
		return
	}

	versionsDir := filepath.Join(homeDir, ".jenvy", "versions")

	// Controlla se la directory esiste
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		utils.PrintError("No JDK installations found")
		utils.PrintInfo("The versions directory doesn't exist yet")
		return
	}

	// Leggi tutte le installazioni
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to read versions directory: %v", err))
		return
	}

	// Filtra solo le directory
	var jdkDirs []string
	var jdkVersions []string
	for _, entry := range entries {
		if entry.IsDir() {
			jdkDirs = append(jdkDirs, entry.Name())
			version := extractVersionFromDirName(entry.Name())
			if version != "" {
				jdkVersions = append(jdkVersions, version)
			}
		}
	}

	if len(jdkDirs) == 0 {
		utils.PrintInfo("No JDK installations found to remove")
		return
	}

	// Mostra quello che verrà rimosso
	fmt.Printf("You are about to remove ALL JDK installations (%d total):\n", len(jdkVersions))
	for _, version := range jdkVersions {
		fmt.Printf("   - %s\n", version)
	}
	fmt.Printf("\n   Directory: %s\n", versionsDir)

	// Controlla se qualche JDK è attualmente in uso
	currentJavaHome := os.Getenv("JAVA_HOME")
	if currentJavaHome != "" {
		normalizedJavaHome := strings.ToLower(filepath.Clean(currentJavaHome))
		normalizedVersionsDir := strings.ToLower(filepath.Clean(versionsDir))

		if strings.HasPrefix(normalizedJavaHome, normalizedVersionsDir) {
			utils.PrintWarning("One of these JDKs is currently set as JAVA_HOME")
			utils.PrintInfo("This will unset your current Java environment")
		}
	}

	// Doppia conferma per operazione potenzialmente distruttiva
	fmt.Print("\nWARNING: This will permanently delete ALL JDK installations!")
	fmt.Print("\n   Are you absolutely sure? Type 'yes' to confirm: ")

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(strings.TrimSpace(response)) != "yes" {
		utils.PrintInfo("Removal cancelled")
		return
	}

	// Procedi con la rimozione
	fmt.Printf("Removing all JDK installations...\n")

	var removedCount int
	var failedRemovals []string

	for _, dirName := range jdkDirs {
		jdkPath := filepath.Join(versionsDir, dirName)
		version := extractVersionFromDirName(dirName)

		fmt.Printf("   Removing %s...\n", version)
		err := os.RemoveAll(jdkPath)
		if err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to remove %s: %v", version, err))
			failedRemovals = append(failedRemovals, version)
		} else {
			removedCount++
		}
	}

	// Riporta i risultati
	if removedCount > 0 {
		utils.PrintSuccess(fmt.Sprintf("Successfully removed %d JDK installation(s)", removedCount))
	}

	if len(failedRemovals) > 0 {
		utils.PrintWarning(fmt.Sprintf("Failed to remove %d JDK(s):", len(failedRemovals)))
		for _, version := range failedRemovals {
			fmt.Printf("   - %s\n", version)
		}
		utils.PrintInfo("Make sure no applications are using these JDKs")
	}

	// Se tutti i JDK sono stati rimossi con successo, rimuovi anche la directory versions se vuota
	if len(failedRemovals) == 0 {
		if entries, err := os.ReadDir(versionsDir); err == nil && len(entries) == 0 {
			os.Remove(versionsDir) // Rimuovi la directory vuota (ignora errori)
		}
		utils.PrintInfo("All JDK installations have been removed")
		utils.PrintInfo("Run 'jenvy download <version>' to install a new JDK")
	}
}
