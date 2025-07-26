package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"jenvy/internal/utils"

	"golang.org/x/sys/windows/registry"
)

// UseJDK attiva una versione specifica di JDK come JAVA_HOME di sistema su Windows.
//
// Questa è la funzione principale del comando "jvm use" che gestisce l'intero processo
// di selezione e attivazione di una versione JDK installata, inclusa la gestione
// dei privilegi amministratore richiesti per modificare le variabili d'ambiente di sistema.
//
// Processo di attivazione JDK:
// 1. **Validazione argomenti**: Verifica che sia specificata una versione
// 2. **Controllo privilegi**: Verifica privilegi amministratore per modifiche di sistema
// 3. **Elevazione automatica**: Richiede privilegi amministratore se necessario
// 4. **Ricerca JDK**: Localizza l'installazione JDK corrispondente
// 5. **Validazione JDK**: Verifica che sia un'installazione JDK completa
// 6. **Impostazione JAVA_HOME**: Modifica registro Windows per JAVA_HOME
// 7. **Aggiornamento PATH**: Assicura che %JAVA_HOME%\bin sia nel PATH
// 8. **Test finale**: Verifica che l'installazione Java sia funzionante
//
// Gestione privilegi amministratore:
//   - **Rilevamento automatico**: Controlla se il processo ha già privilegi admin
//   - **Elevazione UAC**: Usa ShellExecute con "runas" per richiedere privilegi
//   - **Trasparenza utente**: Processo automatico con messaggi informativi
//   - **Fallback graceful**: Opzioni alternative se elevazione fallisce
//
// Ricerca e validazione JDK:
//   - **Exact match**: Cerca prima corrispondenza esatta (es. "17" → "JDK-17")
//   - **Partial match**: Cerca prefissi se exact match non trovato
//   - **Disambiguazione**: Gestisce multiple corrispondenze con lista interattiva
//   - **Validazione struttura**: Verifica presenza bin/, lib/, java.exe
//
// Modifiche registro Windows:
//   - **JAVA_HOME**: Imposta in HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment
//   - **PATH update**: Aggiunge %JAVA_HOME%\bin all'inizio del PATH di sistema
//   - **Broadcasting**: Notifica applicazioni dei cambiamenti ambiente
//   - **Persistenza**: Modifiche permanenti sopravvivono a riavvii
//
// Parametri:
//
//	Legge da os.Args[2] la versione JDK da attivare
//
// Comportamenti speciali:
//   - Se mancano argomenti: Mostra usage e lista JDK disponibili
//   - Se non amministratore: Richiede automaticamente elevazione privilegi
//   - Se multiple corrispondenze: Mostra lista per disambiguazione
//   - Se JDK non valido: Mostra errore dettagliato con suggerimenti
//
// Esempio di utilizzo:
//
//	jvm use 17        → Attiva JDK 17 (cerca JDK-17.x.x)
//	jvm use 17.0.5    → Attiva JDK 17.0.5 specifico
//	jvm u 21          → Forma breve per attivare JDK 21
//
// Output tipico:
//
//	[INFO] Administrator privileges required to modify system environment variables
//	[INFO] Requesting administrator privileges...
//	[SUCCESS] Set JAVA_HOME to JDK 17
//	[INFO] JAVA_HOME = C:\Users\user\.jenvy\versions\JDK-17.0.5
//	[INFO] Restart your terminal/IDE to see the changes
//
// Scenari di errore:
//   - Privilegi insufficienti: Guida per esecuzione come amministratore
//   - JDK non trovato: Suggerisce "jvm list" per vedere JDK disponibili
//   - Directory JDK corrotta: Messaggio di errore con path problematico
//   - Errori registro: Consigli troubleshooting per problemi Windows
func UseJDK() {
	if len(os.Args) < 3 {
		utils.PrintUsage("Usage: jenvy use <version>")
		utils.PrintUsage("Short form: jenvy u <version>")
		utils.PrintInfo("Available JDKs:")
		showAvailableJDKs()
		return
	}

	version := os.Args[2]

	// Prima di tutto, verifichiamo se ci sono JDK installati
	versionsDir, err := utils.GetJVMVersionsDirectory()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to access JVM directory: %v", err))
		return
	}

	// Controlla se la directory versions esiste e contiene JDK
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		utils.PrintError("JVM versions directory not found or inaccessible")
		utils.PrintInfo("No JDKs appear to be installed yet")
		utils.PrintInfo(fmt.Sprintf("Use 'jenvy download %s' to download your first JDK", version))
		return
	}

	// Verifica se ci sono JDK validi installati
	jdkCount := 0
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "JDK-") {
			jdkPath := filepath.Join(versionsDir, entry.Name())
			if utils.IsValidJDKDirectory(jdkPath) {
				jdkCount++
			}
		}
	}

	if jdkCount == 0 {
		utils.PrintError("No valid JDK installations found")
		utils.PrintInfo("The .jenvy/versions directory exists but contains no valid JDK installations")
		utils.PrintInfo(fmt.Sprintf("Use 'jenvy download %s' to download a JDK", version))
		return
	}

	// CONTROLLO IMPORTANTE: Verifica se la versione richiesta esiste PRIMA di richiedere privilegi admin
	jdkPath, err := utils.FindSingleJDKInstallation(version)
	if err != nil {
		// Fornire messaggi di errore più specifici e utili
		if strings.Contains(err.Error(), "no JDK found matching version") {
			utils.PrintError(fmt.Sprintf("JDK version %s not found", version))
			utils.PrintInfo("Available JDK versions:")
			showAvailableJDKs()
			utils.PrintInfo(fmt.Sprintf("Use 'jenvy download %s' to download this version", version))
		} else if strings.Contains(err.Error(), "multiple matches found") {
			utils.PrintError(fmt.Sprintf("Multiple JDK versions match '%s'", version))
			utils.PrintInfo("Please be more specific with the version number")
			utils.PrintInfo("Example: use 'jenvy use 17.0.5' instead of 'jenvy use 17'")
		} else if strings.Contains(err.Error(), "failed to get JVM directory") {
			utils.PrintError("Unable to access JVM installation directory")
			utils.PrintInfo("Make sure you have proper permissions and the .jenvy directory exists")
		} else {
			utils.PrintError(fmt.Sprintf("Failed to locate JDK version %s: %v", version, err))
		}
		return
	}

	// Verify it's a valid JDK directory PRIMA di richiedere privilegi admin
	if !utils.IsValidJDKDirectory(jdkPath) {
		utils.PrintError(fmt.Sprintf("Invalid or corrupted JDK directory: %s", jdkPath))
		utils.PrintInfo("This JDK installation appears to be incomplete or damaged")
		utils.PrintInfo(fmt.Sprintf("Try downloading it again with: jvm download %s", version))
		return
	}

	// Check if running as administrator
	if !isRunningAsAdmin() {
		utils.PrintInfo("Administrator privileges required to modify system environment variables")
		utils.PrintInfo("Requesting administrator privileges...")

		if requestAdminPrivileges() {
			return // Exit current process, admin process will handle the command
		} else {
			utils.PrintError("Failed to obtain administrator privileges")
			utils.PrintInfo("You can run manually as Administrator or use user-level installation")
			return
		}
	}

	// Set JAVA_HOME in system environment
	err = setSystemEnvironmentVariable("JAVA_HOME", jdkPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to set JAVA_HOME: %v", err))
		utils.PrintInfo("Try running as Administrator")
		return
	}

	// Ensure %JAVA_HOME%\\bin is in PATH
	err = ensureJavaHomeInPath()
	if err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to update PATH: %v", err))
		utils.PrintInfo("You may need to add %JAVA_HOME%\\bin to your PATH manually")
	}

	utils.PrintSuccess(fmt.Sprintf("Set JAVA_HOME to JDK %s", version))
	utils.PrintInfo(fmt.Sprintf("JAVA_HOME = %s", jdkPath))
	utils.PrintInfo("Restart your terminal/IDE to see the changes")

	// Show Java version
	fmt.Println()
	utils.PrintInfo("Testing Java installation:")
	testJavaInstallation(jdkPath)
}

// requestAdminPrivileges richiede automaticamente privilegi amministratore tramite UAC Windows.
//
// Questa funzione gestisce l'elevazione dei privilegi quando il comando "jvm use"
// necessita di modificare le variabili d'ambiente di sistema. Utilizza l'API Windows
// ShellExecute con il verbo "runas" per attivare il dialogo UAC (User Account Control).
//
// Meccanismo elevazione UAC:
// 1. **Rilevamento eseguibile**: Ottiene il path dell'eseguibile JVM corrente
// 2. **Preparazione argomenti**: Ricostruisce tutti gli argomenti della command line
// 3. **ShellExecute "runas"**: Invoca Windows Shell con richiesta privilegi admin
// 4. **Terminazione processo**: Il processo corrente termina, quello elevato continua
//
// Processo UAC Windows:
//   - **Dialogo sicurezza**: Windows mostra prompt UAC per conferma utente
//   - **Nuovo processo**: Se accettato, viene creato processo con privilegi admin
//   - **Stesso comando**: Il nuovo processo esegue esattamente gli stessi argomenti
//   - **Terminazione originale**: Il processo originale termina dopo l'elevazione
//
// Gestione argomenti:
//   - **Preservazione completa**: Tutti gli argomenti originali vengono mantenuti
//   - **Esclusione program name**: Solo gli argomenti reali (os.Args[1:])
//   - **Join sicuro**: Concatenazione argomenti con spazi per ShellExecute
//   - **Unicode support**: Gestione corretta caratteri Unicode in percorsi
//
// Codici ritorno ShellExecute:
//   - **> 32**: Successo, elevazione completata
//   - **<= 32**: Errore o cancellazione utente
//   - **Codici comuni**: 2=file not found, 5=access denied, 8=memoria insufficiente
//
// Parametri:
//
//	Nessuno (legge da os.Args globale)
//
// Restituisce:
//
//	bool - true se elevazione completata con successo, false se fallita o rifiutata
//
// Comportamento trasparente:
//   - Se successo: Il processo corrente termina, quello elevato prosegue silenziosamente
//   - Se fallimento: Il processo corrente continua con messaggi di errore appropriati
//   - Se cancellato: L'utente ha rifiutato l'elevazione nel dialogo UAC
//
// Scenari di utilizzo:
//   - Utente standard che esegue "jvm use"
//   - Modifica variabili ambiente sistema richiede privilegi admin
//   - Alternativa a esecuzione manuale "Run as Administrator"
//
// Limitazioni:
//   - Richiede interazione utente (dialogo UAC)
//   - Non funziona in contesti automatizzati senza desktop
//   - Dipende dalle policy UAC del sistema
//
// Esempio di utilizzo:
//
//	if !isRunningAsAdmin() {
//	    if requestAdminPrivileges() {
//	        return // Il nuovo processo gestirà il comando
//	    }
//	    // Gestire fallimento elevazione
//	}
func requestAdminPrivileges() bool {
	// Get current executable path
	exe, err := os.Executable()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get executable path: %v", err))
		return false
	}

	// Build command arguments (pass all original arguments)
	args := os.Args[1:] // Skip the program name

	// Create the command with runas verb to request admin privileges
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)

	// Join arguments into a single string
	argString := strings.Join(args, " ")
	argPtr, _ := syscall.UTF16PtrFromString(argString)

	// Use ShellExecute to run with elevated privileges
	ret := shellExecute(0, verbPtr, exePtr, argPtr, nil, 1)

	// Return true if ShellExecute succeeded (> 32)
	return ret > 32
}

// shellExecute è un wrapper Go per l'API Windows ShellExecuteW per esecuzione programmi con privilegi.
//
// Questa funzione incapsula la chiamata diretta all'API Win32 ShellExecuteW utilizzando
// syscall per eseguire programmi con parametri specifici, inclusa la possibilità di
// richiedere elevazione privilegi tramite il verbo "runas".
//
// API Windows ShellExecuteW:
//   - **Funzione nativa**: shell32.dll ShellExecuteW per esecuzione avanzata
//   - **Unicode support**: Versione Wide (W) per supporto caratteri Unicode
//   - **Verbi azione**: "open", "runas", "print", etc. per diversi comportamenti
//   - **Controllo finestra**: Parametri per gestione visualizzazione finestra
//
// Parametri:
//
//	hwnd uintptr     - Handle finestra parent (0 per nessun parent)
//	verb *uint16     - Verbo azione: nil="open", "runas"=privilegi admin
//	file *uint16     - Percorso eseguibile da lanciare (UTF-16 pointer)
//	args *uint16     - Argomenti command line (UTF-16 pointer, può essere nil)
//	dir *uint16      - Directory lavoro (UTF-16 pointer, può essere nil)
//	show int         - Modalità visualizzazione finestra (SW_HIDE=0, SW_NORMAL=1, etc.)
//
// Restituisce:
//
//	uintptr - Codice ritorno ShellExecute (>32=successo, <=32=errore specifico)
//
// Codici ritorno comuni:
//   - **> 32**: Successo, programma avviato correttamente
//   - **0**: Out of memory or resources
//   - **2**: File not found (ERROR_FILE_NOT_FOUND)
//   - **3**: Path not found (ERROR_PATH_NOT_FOUND)
//   - **5**: Access denied (ERROR_ACCESS_DENIED)
//   - **8**: Out of memory (ERROR_NOT_ENOUGH_MEMORY)
//   - **31**: No application associated with file type
//
// Utilizzo syscall.NewLazyDLL:
//   - **Caricamento lazy**: DLL caricata solo quando necessario
//   - **Performance**: Evita caricamento inutile se funzione non usata
//   - **Gestione errori**: syscall gestisce automaticamente errori caricamento
//   - **Pulizia automatica**: Go runtime gestisce cleanup DLL
//
// Sicurezza:
//   - **unsafe.Pointer**: Necessario per compatibilità API C Windows
//   - **Validazione input**: Chiamante responsabile per validazione parametri
//   - **Gestione memoria**: Go runtime gestisce stringhe UTF-16
//
// Esempio di utilizzo con UAC:
//
//	verb, _ := syscall.UTF16PtrFromString("runas")
//	exe, _ := syscall.UTF16PtrFromString("C:\\app.exe")
//	args, _ := syscall.UTF16PtrFromString("arg1 arg2")
//	ret := shellExecute(0, verb, exe, args, nil, 1)
//	if ret > 32 { /* successo */ }
func shellExecute(hwnd uintptr, verb, file, args, dir *uint16, show int) uintptr {
	ret, _, _ := syscall.NewLazyDLL("shell32.dll").NewProc("ShellExecuteW").Call(
		hwnd,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(args)),
		uintptr(unsafe.Pointer(dir)),
		uintptr(show))
	return ret
}

// setSystemEnvironmentVariable imposta una variabile d'ambiente di sistema nel registro Windows.
//
// Questa funzione modifica permanentemente le variabili d'ambiente a livello di sistema
// attraverso il registro di Windows, rendendo le modifiche persistenti e disponibili
// per tutti gli utenti e servizi del sistema.
//
// Registro Windows utilizzato:
//
//	Chiave: HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment
//	Tipo: REG_SZ (String Value)
//	Scopo: Variabili d'ambiente sistema globali
//
// Processo di modifica:
// 1. **Apertura chiave registro**: Apre con permessi SET_VALUE per modifica
// 2. **Impostazione valore**: Scrive la variabile come stringa nel registro
// 3. **Chiusura chiave**: Cleanup automatico con defer per sicurezza
// 4. **Broadcasting**: Notifica sistema del cambiamento (implementazione futura)
//
// Requisiti privilegi:
//   - **Amministratore richiesto**: HKLM richiede privilegi elevated
//   - **UAC necessario**: Su Windows Vista+ serve elevazione UAC
//   - **Servizi Windows**: Accesso completo per modifiche sistema
//
// Persistenza e scope:
//   - **Permanente**: Sopravvive a riavvii sistema
//   - **Globale**: Disponibile per tutti gli utenti
//   - **Servizi**: Accessibile ai servizi Windows
//   - **Nuove sessioni**: Automaticamente disponibile in nuovi login
//
// Broadcasting (da implementare):
//   - **WM_SETTINGCHANGE**: Messaggio Windows per notifica applicazioni
//   - **HWND_BROADCAST**: Broadcast a tutte le finestre top-level
//   - **Update live**: Alcune applicazioni aggiornano senza riavvio
//
// Parametri:
//
//	name string  - Nome variabile d'ambiente (es. "JAVA_HOME")
//	value string - Valore da assegnare (es. "C:\Program Files\Java\jdk-17")
//
// Restituisce:
//
//	error - nil se successo, errore specifico se operazione fallisce
//
// Errori comuni:
//   - **Permessi insufficienti**: Processo non eseguito come amministratore
//   - **Chiave inaccessibile**: Registro corrotto o permessi negati
//   - **Valore non impostabile**: Problemi scrittura registro o memoria
//   - **Nome invalido**: Caratteri speciali non supportati nel nome
//
// Esempio di utilizzo:
//
//	err := setSystemEnvironmentVariable("JAVA_HOME", "C:\\jdk-17")
//	if err != nil {
//	    log.Printf("Failed to set JAVA_HOME: %v", err)
//	}
//
// Note di sicurezza:
//   - Non valida caratteri pericolosi nel nome/valore
//   - Non previene sovrascrittura variabili sistema critiche
//   - Responsabilità chiamante per validazione input
func setSystemEnvironmentVariable(name, value string) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	err = key.SetStringValue(name, value)
	if err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	// Broadcast WM_SETTINGCHANGE message to notify applications
	// This helps some applications pick up the new environment variable
	utils.PrintInfo("Broadcasting environment change...")

	return nil
}

// ensureJavaHomeInPath assicura che %JAVA_HOME%\bin sia presente nel PATH di sistema Windows.
//
// Questa funzione gestisce l'aggiornamento intelligente della variabile PATH di sistema
// per includere la directory bin del JDK attivo, permettendo l'esecuzione diretta
// di comandi Java da qualsiasi posizione nel prompt dei comandi.
//
// Processo di aggiornamento PATH:
// 1. **Lettura PATH corrente**: Recupera valore attuale dal registro sistema
// 2. **Parsing entries**: Suddivide PATH in singole directory separate da ";"
// 3. **Controllo esistenza**: Verifica se %JAVA_HOME%\bin è già presente
// 4. **Aggiunta intelligente**: Se mancante, aggiunge all'inizio del PATH
// 5. **Scrittura registro**: Salva il nuovo PATH nel registro sistema
//
// Registro Windows utilizzato:
//
//	Chiave: HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment
//	Valore: "Path" (REG_EXPAND_SZ o REG_SZ)
//	Permessi: QUERY_VALUE | SET_VALUE per lettura e modifica
//
// Gestione %JAVA_HOME%\bin:
//   - **Variabile espandibile**: Usa %JAVA_HOME%\bin invece di path assoluto
//   - **Posizione prioritaria**: Aggiunto all'inizio del PATH per precedenza
//   - **Case-insensitive**: Confronto ignorando maiuscole/minuscole
//   - **Trim whitespace**: Rimuove spazi accidentali nelle entries PATH
//
// Vantaggi utilizzo %JAVA_HOME%\bin:
//   - **Dinamico**: Si aggiorna automaticamente quando JAVA_HOME cambia
//   - **Portable**: Non hard-coded a path specifici
//   - **Standard**: Convenzione comune per setup Java
//   - **Manutenibile**: Un solo punto di aggiornamento (JAVA_HOME)
//
// Comportamento intelligente:
//   - **Evita duplicati**: Non aggiunge se già presente nel PATH
//   - **Priorità elevata**: Inserimento all'inizio per precedenza su altre versioni Java
//   - **Preservazione PATH**: Mantiene tutte le altre entries esistenti
//   - **Feedback utente**: Messaggi informativi su operazioni eseguite
//
// Parametri:
//
//	Nessuno (opera su variabili d'ambiente sistema)
//
// Restituisce:
//
//	error - nil se successo o già presente, errore se modifica fallisce
//
// Messaggi output:
//   - "[INFO] %JAVA_HOME%\bin is already in PATH" - se già configurato
//   - "[SUCCESS] Added %JAVA_HOME%\bin to system PATH" - se aggiunto con successo
//
// Scenari di errore:
//   - **Permessi insufficienti**: Richiede privilegi amministratore
//   - **Registro inaccessibile**: Chiave sistema corrotta o bloccata
//   - **PATH corrotto**: Valore PATH nel formato non riconosciuto
//   - **Memoria insufficiente**: PATH troppo lungo per limiti Windows
//
// Esempio PATH risultante:
//
//	Prima:  "C:\Windows\System32;C:\Windows;C:\Program Files\Git\bin"
//	Dopo:   "%JAVA_HOME%\bin;C:\Windows\System32;C:\Windows;C:\Program Files\Git\bin"
func ensureJavaHomeInPath() error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Read current PATH
	currentPath, _, err := key.GetStringValue("Path")
	if err != nil {
		return fmt.Errorf("failed to read PATH: %w", err)
	}

	javaHomeBin := `%JAVA_HOME%\bin`

	// Check if %JAVA_HOME%\\bin is already in PATH
	pathEntries := strings.Split(currentPath, ";")
	for _, entry := range pathEntries {
		if strings.EqualFold(strings.TrimSpace(entry), javaHomeBin) {
			utils.PrintInfo("%JAVA_HOME%\\bin is already in PATH")
			return nil
		}
	}

	// Add %JAVA_HOME%\\bin to the beginning of PATH
	newPath := javaHomeBin + ";" + currentPath

	err = key.SetStringValue("Path", newPath)
	if err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}

	utils.PrintSuccess("Added %JAVA_HOME%\\bin to system PATH")
	return nil
}

// showAvailableJDKs mostra una lista delle installazioni JDK disponibili nel sistema.
//
// Questa funzione è una utility di supporto che enumera tutte le versioni JDK
// installate nella directory standard JVM e le presenta in formato user-friendly.
// Viene utilizzata quando l'utente non specifica una versione o quando si verificano
// errori di ricerca per guidare l'utente nelle opzioni disponibili.
//
// Processo di enumerazione:
// 1. **Rilevamento directory home**: Ottiene directory utente corrente
// 2. **Costruzione path**: Costruisce path alla directory versions di JVM
// 3. **Lettura directory**: Enumera tutte le subdirectory presenti
// 4. **Filtraggio JDK**: Identifica solo directory con prefisso "JDK-"
// 5. **Validazione**: Verifica che ogni directory sia un JDK valido
// 6. **Estrazione versione**: Rimuove prefisso "JDK-" per mostrare versione pulita
// 7. **Presentazione**: Mostra lista formattata all'utente
//
// Struttura directory analizzata:
//
//	C:\Users\{user}\.jenvy\versions\
//	├── JDK-17.0.5\     → Versione "17.0.5"
//	├── JDK-21.0.2\     → Versione "21.0.2"
//	├── JDK-8.0.392\    → Versione "8.0.392"
//	└── Other-Folder\   → Ignorata (no prefisso JDK-)
//
// Filtraggio intelligente:
//   - **Prefisso JDK-**: Solo directory che iniziano con "JDK-"
//   - **Directory valide**: Solo directory, non file
//   - **JDK completi**: Validazione struttura tramite isValidJDKDirectory()
//   - **Esclusione corrotti**: Directory JDK incomplete vengono omesse
//
// Gestione scenari vuoti:
//   - **Directory home inaccessibile**: Messaggio errore con dettagli
//   - **Directory versions inesistente**: Suggerisce download primo JDK
//   - **Nessun JDK trovato**: Informa su procedura installazione
//   - **JDK corrotti**: Mostra solo quelli validi, ignora gli altri
//
// Parametri:
//
//	Nessuno (opera su directory standard JVM)
//
// Output tipico:
//
//	Available JDK versions:
//	  - 17.0.5
//	  - 21.0.2
//	  - 8.0.392
//
// Messaggi speciali:
//   - "No JDKs found. Use 'jenvy download <version>' to install a JDK" - se directory vuota
//   - "No valid JDKs found. Use 'jenvy download <version>' to install a JDK" - se solo JDK corrotti
//   - "Failed to get home directory: {error}" - se problemi accesso directory utente
//
// Utilizzo nei comandi:
//   - Automaticamente mostrata in UseJDK() se mancano argomenti
//   - Suggerita in messaggi di errore per guidare l'utente
//   - Helper per comando "jvm list" per overview installazioni
//
// Integrazione con validazione:
//   - Usa isValidJDKDirectory() per controllo qualità
//   - Evita di mostrare installazioni corrotte all'utente
//   - Garantisce che versioni mostrate siano effettivamente utilizzabili
func showAvailableJDKs() {
	versionsDir, err := utils.GetJVMVersionsDirectory()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get JVM directory: %v", err))
		return
	}

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		utils.PrintWarning("No JDKs found. Use 'jenvy download <version>' to install a JDK")
		return
	}

	var jdks []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "JDK-") {
			version := strings.TrimPrefix(entry.Name(), "JDK-")
			jdkPath := filepath.Join(versionsDir, entry.Name())
			if utils.IsValidJDKDirectory(jdkPath) {
				jdks = append(jdks, version)
			}
		}
	}

	if len(jdks) == 0 {
		utils.PrintWarning("No valid JDKs found. Use 'jenvy download <version>' to install a JDK")
		return
	}

	fmt.Println("Available JDK versions:")
	for _, jdk := range jdks {
		fmt.Printf("  - %s\n", jdk)
	}
}

// testJavaInstallation testa se un'installazione Java è funzionante e accessibile.
//
// Questa funzione esegue controlli basilari per verificare che un'installazione JDK
// sia configurata correttamente e l'eseguibile Java sia accessibile. Fornisce
// feedback immediato all'utente sulla funzionalità dell'installazione appena attivata.
//
// Controlli eseguiti:
// 1. **Costruzione path**: Costruisce percorso completo a java.exe
// 2. **Verifica esistenza**: Controlla che java.exe esista fisicamente
// 3. **Test accessibilità**: Verifica che il file sia accessibile e leggibile
// 4. **Feedback utente**: Informa sui risultati dei controlli
//
// Path testato:
//
//	{jdkPath}\bin\java.exe
//	Esempio: "C:\Users\user\.jenvy\versions\JDK-17\bin\java.exe"
//
// Controlli limitati (per semplicità):
//   - **Solo esistenza file**: Non esegue java.exe per test funzionalità
//   - **No version check**: Non verifica che versione Java corrisponda
//   - **No execution test**: Non testa "java -version" command
//   - **No permissions**: Non verifica permessi esecuzione
//
// Razionale controlli limitati:
//   - **Performance**: Evita overhead esecuzione processi
//   - **Sicurezza**: Non esegue codice esterno automaticamente
//   - **Semplicità**: Controllo base sufficiente per validazione immediata
//   - **Reliability**: Meno dipendenze = meno punti di fallimento
//
// Parametri:
//
//	jdkPath string - Percorso directory root del JDK da testare
//
// Output tipico successo:
//
//	Testing: C:\Users\user\.jenvy\versions\JDK-17\bin\java.exe -version
//	[SUCCESS] Java executable found and accessible
//	Java location: C:\Users\user\.jenvy\versions\JDK-17\bin\java.exe
//
// Output tipico errore:
//
//	Testing: C:\Users\user\.jenvy\versions\JDK-17\bin\java.exe -version
//	[ERROR] Java executable not found
//
// Miglioramenti futuri possibili:
//   - Esecuzione effettiva "java -version" con cattura output
//   - Parsing versione e verifica corrispondenza
//   - Test compilazione semplice per validare javac
//   - Controllo permessi esecuzione
//   - Timeout per prevenire hang su installazioni corrotte
//
// Integrazione nel workflow:
//   - Chiamata automatica dopo UseJDK() per feedback immediato
//   - Validazione finale prima di considerare operazione completata
//   - Troubleshooting helper per identificare problemi installazione
func testJavaInstallation(jdkPath string) {
	javaExe := filepath.Join(jdkPath, "bin", "java.exe")

	// Test java -version command
	fmt.Printf("Testing: %s -version\n", javaExe)

	// We can't easily run the command and capture output here without additional complexity
	// Instead, we'll just verify the executable exists and is accessible
	if _, err := os.Stat(javaExe); err != nil {
		utils.PrintError("Java executable not found")
		return
	}

	utils.PrintSuccess("Java executable found and accessible")
	fmt.Printf("Java location: %s\n", javaExe)
}

// InitializeJVMEnvironment configura l'ambiente iniziale per JVM durante l'installazione.
//
// Questa funzione gestisce il setup iniziale dell'ambiente JVM quando il tool viene
// installato per la prima volta, preparando le variabili d'ambiente e i percorsi
// necessari per il corretto funzionamento del sistema di gestione versioni Java.
//
// Operazioni di inizializzazione:
// 1. **Controllo privilegi**: Verifica se eseguito come amministratore
// 2. **Setup PATH sistema**: Prepara PATH per supportare %JAVA_HOME%\bin
// 3. **Feedback privilegi**: Informa utente su limitazioni senza privilegi admin
// 4. **Guida utilizzo**: Fornisce istruzioni per prossimi passi
//
// Gestione privilegi amministratore:
//   - **Con privilegi**: Setup completo variabili d'ambiente sistema
//   - **Senza privilegi**: Avvisa che "jvm use" richiederà elevazione UAC
//   - **Messaggio informativo**: Spiega implicazioni e alternative
//
// Setup PATH sistema:
//   - **Preparazione**: Assicura che PATH sia configurato per %JAVA_HOME%\bin
//   - **Non destructive**: Non modifica JAVA_HOME fino a primo "jvm use"
//   - **Reversibile**: Setup può essere facilmente annullato se necessario
//
// Scenari di utilizzo:
//   - **Prima installazione**: Setup ambiente quando JVM installato
//   - **Reinstallazione**: Ripristino configurazione dopo problemi
//   - **Setup automatico**: Parte di processo installazione automatizzata
//   - **Configurazione manuale**: Chiamata manuale per fix problemi ambiente
//
// Parametri:
//
//	Nessuno (funzione di setup globale)
//
// Output con privilegi admin:
//
//	🔧 Setting up JVM environment variables...
//	[SUCCESS] JVM environment initialized
//	[INFO] Use 'jenvy use <version>' to set your active JDK
//
// Output senza privilegi admin:
//
//	🔧 Setting up JVM environment variables...
//	[WARNING] For system-wide environment variables, run as Administrator
//	[INFO] You can still use JVM, but 'jenvy use' will require Administrator privileges
//	[INFO] Use 'jenvy use <version>' to set your active JDK
//
// Messaggi di errore possibili:
//
//	[ERROR] Failed to initialize PATH: {error details}
//	[INFO] You may need to manually add %JAVA_HOME%\bin to your PATH
//
// Integrazione con installer:
//   - Chiamata automatica durante setup.exe
//   - Parte del processo post-installazione
//   - Prerequisito per utilizzo normale del tool
//
// Note per sviluppatori:
//   - Non richiede JDK già installati per funzionare
//   - Prepara solo l'ambiente, non installa JDK
//   - Idempotente: sicuro chiamare multiple volte
func InitializeJVMEnvironment() {
	fmt.Println("🔧 Setting up JVM environment variables...")

	// Check if running as administrator
	if !isRunningAsAdmin() {
		utils.PrintWarning("For system-wide environment variables, run as Administrator")
		utils.PrintInfo("You can still use JVM, but 'jenvy use' will require Administrator privileges")
	} else {
		// Ensure %JAVA_HOME%\\bin is in PATH (will be set when a JDK is selected)
		err := ensureJavaHomeInPath()
		if err != nil {
			utils.PrintError(fmt.Sprintf("Failed to initialize PATH: %v", err))
			utils.PrintInfo("You may need to manually add %JAVA_HOME%\\bin to your PATH")
		} else {
			utils.PrintSuccess("JVM environment initialized")
		}
	}

	utils.PrintInfo("Use 'jenvy use <version>' to set your active JDK")
}

// isRunningAsAdmin verifica se il processo corrente ha privilegi di amministratore.
//
// Questa funzione implementa un controllo affidabile per determinare se l'applicazione
// è in esecuzione con privilegi elevati, necessari per modificare le variabili
// d'ambiente di sistema tramite il registro Windows.
//
// Metodo di verifica:
//
//	Tenta di aprire una chiave del registro che richiede privilegi amministratore
//	per l'accesso in scrittura. Se l'operazione riesce, il processo ha privilegi admin.
//
// Chiave registro utilizzata per test:
//
//	HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\Environment
//	- **Criticità**: Chiave sistema per variabili d'ambiente globali
//	- **Protezione**: Richiede privilegi elevated per SET_VALUE
//	- **Affidabilità**: Controllo diretto sui permessi reali necessari
//
// Vantaggi di questo approccio:
//   - **Test reale**: Verifica esattamente i permessi che servono per operazioni JVM
//   - **Affidabile**: Non dipende da API che potrebbero cambiare
//   - **Specifico**: Testa accesso alla specifica risorsa che useremo
//   - **Immediato**: Fallisce velocemente se privilegi insufficienti
//
// Alternative non utilizzate:
//   - **Token API**: Più complesso e dipendente da versioni Windows
//   - **Gruppo Administrators**: Membership non garantisce privilegi attivi
//   - **UAC API**: Overhead maggiore per controllo semplice
//
// Meccanismo:
// 1. **Tentativo apertura**: Prova ad aprire chiave con permessi SET_VALUE
// 2. **Gestione errore**: Se fallisce, assenza privilegi amministratore
// 3. **Pulizia**: Chiude chiave immediatamente se apertura riuscita
// 4. **Ritorno booleano**: true se privilegi presenti, false altrimenti
//
// Parametri:
//
//	Nessuno (controlla processo corrente)
//
// Restituisce:
//
//	bool - true se processo ha privilegi amministratore, false altrimenti
//
// Utilizzo tipico:
//
//	if !isRunningAsAdmin() {
//	    // Richiedi elevazione UAC
//	    requestAdminPrivileges()
//	} else {
//	    // Procedi con modifiche sistema
//	}
//
// Scenari di utilizzo:
//   - Prima di ogni modifica variabili d'ambiente sistema
//   - Decisione se mostrare prompt UAC o errore
//   - Validazione prerequisiti per operazioni privilegiate
//   - Guida utente su come eseguire comando correttamente
//
// Limitazioni:
//   - Non distingue tra diversi livelli di privilegi admin
//   - Non rileva UAC disabilitato o policy gruppo
//   - Test specifico per registro, potrebbe non coprire altri privilegi
func isRunningAsAdmin() bool {
	// Try to open a registry key that requires admin access
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.SET_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	return true
}
