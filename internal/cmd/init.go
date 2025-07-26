package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// Init inizializza completamente l'ambiente Java Version Manager (JVM) su Windows.
//
// Questa funzione esegue le seguenti operazioni di setup:
// 1. Verifica che l'eseguibile jvm sia presente nel PATH di sistema
// 2. Crea la directory di configurazione ~/.jenvy se non esiste
// 3. Installa gli script di autocompletamento per Bash, PowerShell e CMD
// 4. Controlla l'esistenza di un file di configurazione e ne crea uno di default se necessario
// 5. Mostra un riepilogo dei comandi disponibili all'utente
//
// La funzione fornisce feedback dettagliato durante ogni fase dell'inizializzazione,
// utilizzando colori per migliorare l'esperienza utente.
//
// Esempi di utilizzo:
//
//	jvm init                    # Inizializzazione completa
//
// Note:
//   - Dopo l'esecuzione è necessario riavviare il terminale per abilitare l'autocompletamento
//   - La directory di configurazione viene creata in C:\Users\username\.jenvy
//   - Il comando 'jenvy use' richiede privilegi amministratore per modificare il Registry Windows
//   - Il file di configurazione di default include impostazioni per provider preferito e LTS
//   - Supporta solo Windows (richiede Registry di sistema e UAC per gestione JAVA_HOME)
func Init() {
	fmt.Println("Initializing Java Version Manager...")

	// Verifica se jvm è nel PATH
	if !isJvmInPath() {
		fmt.Println("[WARNING] JVM executable not found in PATH")
		fmt.Println("[INFO] Consider adding the JVM directory to your PATH for global access")
	}

	// Crea directory di configurazione se non esiste
	if err := createConfigDirectory(); err != nil {
		fmt.Printf("[ERROR] Failed to create config directory: %v\n", err)
	} else {
		fmt.Println("[SUCCESS] Configuration directory ready")
	}

	// Installa completamento per tutte le shell
	fmt.Println("\nSetting up shell completions...")
	InstallCompletionForAllShells()

	// Controlla configurazione esistente
	fmt.Println("\nChecking current configuration...")
	if hasExistingConfig() {
		fmt.Println("[SUCCESS] Configuration file found")
	} else {
		fmt.Println("[INFO] No configuration file found - creating default configuration")
		if err := createDefaultConfig(); err != nil {
			fmt.Printf("  [WARNING] Failed to create default config: %v\n", err)
		} else {
			fmt.Println("[SUCCESS] Default configuration created")
		}
	}

	fmt.Println("\nJava Version Manager initialization complete!")
	fmt.Println("[INFO] Available commands:")
	fmt.Println("   jvm remote-list       - Show available JDK versions")
	fmt.Println("   jvm list               - Show installed JDK versions")
	fmt.Println("   jvm download <version> - Download and install a JDK version")
	fmt.Println("   jvm use <version>     - Set JAVA_HOME system-wide")
	fmt.Println("   jvm remove <version>  - Remove installed JDK version")
	fmt.Println("   jvm help              - Show detailed help")
	fmt.Println("\n🔧 Restart your terminal to enable tab completion!")
}

// isJvmInPath verifica se l'eseguibile jvm è accessibile tramite la variabile PATH di Windows.
//
// Questa funzione confronta la directory dell'eseguibile corrente con tutte le directory
// presenti nella variabile d'ambiente PATH per determinare se JVM è installato globalmente.
//
// Processo di verifica:
// 1. Ottiene il percorso assoluto dell'eseguibile corrente usando os.Executable()
// 2. Estrae la directory contenente l'eseguibile
// 3. Legge la variabile d'ambiente PATH e la suddivide in singole directory
// 4. Confronta ogni directory nel PATH con la directory dell'eseguibile
// 5. Utilizza filepath.Clean() per normalizzare i percorsi prima del confronto
//
// Restituisce:
//
//	true  - se jvm è nel PATH e accessibile globalmente
//	false - se jvm non è nel PATH o si è verificato un errore
//
// Casi d'uso:
//   - Validazione durante l'inizializzazione del sistema
//   - Avviso all'utente se il tool non è installato globalmente
//   - Diagnostica per problemi di accessibilità del comando
//
// Note tecniche Windows:
//   - filepath.Clean() risolve riferimenti relativi (., ..) e separatori multipli
//   - filepath.SplitList() usa il separatore ; appropriato per Windows PATH
//   - Gestisce automaticamente i percorsi Windows con backslash
func isJvmInPath() bool {
	// Ottieni la directory dell'eseguibile corrente
	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	exeDir := filepath.Dir(exePath)
	pathEnv := os.Getenv("PATH")

	paths := filepath.SplitList(pathEnv)
	for _, path := range paths {
		if filepath.Clean(path) == filepath.Clean(exeDir) {
			return true
		}
	}

	return false
}

// createConfigDirectory crea la directory di configurazione .jenvy nel profilo utente Windows.
//
// Questa funzione è responsabile della creazione della struttura di directory necessaria
// per memorizzare i file di configurazione, cache e dati temporanei di JVM su Windows.
//
// Struttura creata:
//
//	C:\Users\username\.jenvy\    # Directory principale di configurazione
//	├── config.json           # File di configurazione principale (creato separatamente)
//	└── versions\             # Directory per JDK scaricati (creata al primo download)
//
// Processo di creazione:
// 1. Ottiene la directory home dell'utente corrente usando os.UserHomeDir()
// 2. Costruisce il percorso completo: C:\Users\username\.jenvy
// 3. Utilizza os.MkdirAll() per creare ricorsivamente tutte le directory necessarie
// 4. Imposta i permessi a 0755 (equivalente Windows: full control per owner, read per altri)
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	error - nil se la creazione è riuscita, errore specifico altrimenti
//
// Gestione errori:
//   - Errore se non è possibile determinare la directory home Windows
//   - Errore se non si hanno i permessi per creare la directory
//   - Errore se la directory esiste già come file regolare
//
// Sicurezza Windows:
//   - La directory viene creata nel profilo utente per isolamento
//   - Eredita le ACL (Access Control List) della directory padre
//   - Solo l'utente corrente ha accesso completo alla directory
//
// Integrazione Windows:
//   - Compatibile con tutte le versioni moderne di Windows (7, 8, 10, 11)
//   - Supporta percorsi lunghi Windows (>260 caratteri) se abilitati
//   - Gestisce caratteri Unicode nei nomi utente Windows
func createConfigDirectory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".jenvy")
	return os.MkdirAll(configDir, 0755)
}

// hasExistingConfig verifica l'esistenza di un file di configurazione nella directory Windows.
//
// Questa funzione controlla se l'utente ha già un file di configurazione config.json
// nella directory C:\Users\username\.jenvy, che contiene le impostazioni personalizzate per JVM.
//
// Percorso verificato:
//
//	C:\Users\username\.jenvy\config.json    # File di configurazione principale
//
// Processo di verifica:
// 1. Ottiene la directory home dell'utente Windows usando os.UserHomeDir()
// 2. Costruisce il percorso completo del file Windows
// 3. Utilizza os.Stat() per verificare l'esistenza e l'accessibilità del file
// 4. Interpreta il risultato: nil = file esiste, errore = file non trovato o inaccessibile
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	bool - true se il file di configurazione esiste ed è accessibile, false altrimenti
//
// Comportamento con errori:
//   - Se non è possibile determinare la home directory Windows: ritorna false
//   - Se il file non esiste: ritorna false
//   - Se il file esiste ma non è accessibile (permessi Windows): ritorna false
//   - Se la directory .jenvy non esiste: ritorna false
//
// Casi d'uso:
//   - Determinare se creare una configurazione di default durante l'inizializzazione
//   - Validare la presenza di configurazioni prima di operazioni avanzate
//   - Evitare sovrascrittura accidentale di configurazioni esistenti
//
// Note Windows:
//   - Non valida il contenuto del file JSON, solo l'esistenza
//   - Non distingue tra file vuoto e file con contenuto valido
//   - Funzione read-only, non modifica il filesystem
//   - Rispetta le ACL (Access Control List) di Windows per l'accesso ai file
func hasExistingConfig() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	configPath := filepath.Join(homeDir, ".jenvy", "config.json")
	_, err = os.Stat(configPath)
	return err == nil
}

// createDefaultConfig genera e scrive un file di configurazione predefinito per Windows.
//
// Questa funzione crea un file config.json con impostazioni predefinite ottimizzate
// per l'ambiente Windows e fornisce un'esperienza utente ottimale sin dal primo utilizzo.
//
// Configurazione predefinita generata:
//
//	{
//	  "defaultProvider": "adoptium",     // Provider Eclipse Adoptium come default (più popolare)
//	  "downloadPath": "",                // Usa directory default Windows (C:\Users\username\.jenvy\versions)
//	  "privateRepositories": [],         // Nessun repository privato configurato inizialmente
//	  "lastCheck": "",                   // Nessun check precedente delle versioni remote
//	  "autoUpdate": true,                // Abilita controlli automatici degli aggiornamenti
//	  "preferLTS": true                  // Preferenza per versioni Long Term Support
//	}
//
// Processo di creazione:
// 1. Definisce la struttura JSON di configurazione con valori sensati per Windows
// 2. Ottiene la directory home dell'utente Windows
// 3. Costruisce il percorso completo: C:\Users\username\.jenvy\config.json
// 4. Scrive il file con permessi appropriati Windows
//
// Parametri:
//
//	Nessuno
//
// Restituisce:
//
//	error - nil se la creazione è riuscita, errore specifico altrimenti
//
// Gestione errori:
//   - Errore se non è possibile determinare la directory home Windows
//   - Errore se non si hanno i permessi di scrittura nella directory .jenvy
//   - Errore se il filesystem è pieno o in sola lettura
//   - Errore se Windows UAC blocca la scrittura nella directory utente
//
// Comportamento:
//   - Sovrascrive file di configurazione esistenti (usare con hasExistingConfig())
//   - Crea un file JSON ben formattato e indentato per leggibilità umana
//   - Usa valori conservativi che funzionano nell'ambiente Windows
//
// Sicurezza Windows:
//   - Il file eredita le ACL della directory padre
//   - Solo l'utente corrente ha accesso di scrittura al file
//   - Non include informazioni sensibili nel file di configurazione di default
//   - Repository privati devono essere configurati esplicitamente dall'utente
//
// Integrazione Windows:
//   - Il JSON generato è compatibile con tutti i parser JSON standard
//   - Le chiavi seguono la convenzione camelCase per coerenza
//   - Valori booleani e stringhe seguono le specifiche JSON RFC 7159
//   - Supporta caratteri Unicode nei percorsi Windows
func createDefaultConfig() error {
	// Definisce la configurazione JSON predefinita con impostazioni ottimali
	// per un nuovo utente di Java Version Manager su Windows
	defaultConfig := `{
  "defaultProvider": "adoptium",
  "downloadPath": "",
  "privateRepositories": [],
  "lastCheck": "",
  "autoUpdate": true,
  "preferLTS": true
}`

	// Ottiene la directory home dell'utente Windows corrente
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Costruisce il percorso completo del file di configurazione Windows
	configPath := filepath.Join(homeDir, ".jenvy", "config.json")

	// Scrive il file di configurazione nel profilo utente Windows
	// Il file avrà permessi appropriati per l'ambiente Windows
	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
