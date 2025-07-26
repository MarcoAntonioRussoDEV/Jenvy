package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

// FixPath esegue la pulizia automatica della variabile d'ambiente PATH di sistema Windows.
//
// Questa funzione implementa un'utilità di manutenzione per il PATH di sistema che:
//
//  1. **Lettura PATH corrente**: Accede al registro di sistema Windows tramite reg.exe
//     per leggere la variabile PATH dalla chiave HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment
//
//  2. **Rilevamento duplicati**: Identifica e rimuove voci duplicate utilizzando
//     confronto case-insensitive per gestire variazioni di maiuscole/minuscole
//
//  3. **Pulizia voci vuote**: Elimina automaticamente entry vuote o contenenti
//     solo spazi bianchi che possono accumularsi nel tempo
//
//  4. **Aggiornamento registro**: Scrive il PATH pulito nel registro di sistema
//     Windows utilizzando il comando reg.exe con privilegi amministratore
//
// **Caratteristiche Windows-specifiche:**
// - Utilizza il separatore ";" tipico di Windows per le variabili PATH
// - Accede direttamente al registro di sistema tramite HKLM
// - Gestisce variabili di tipo REG_EXPAND_SZ per supportare variabili d'ambiente
// - Richiede elevazione UAC per modifiche al PATH di sistema
//
// **Requisiti di sicurezza:**
// - Necessita di privilegi amministratore per modificare il PATH di sistema
// - Utilizza percorsi assoluti (C:\Windows\System32\cmd.exe) per sicurezza
//
// **Output diagnostico:**
// - Mostra statistiche dettagliate delle operazioni di pulizia
// - Elenca ogni voce duplicata rimossa durante il processo
// - Fornisce istruzioni per applicare le modifiche (restart terminale)
//
// Esempi di utilizzo:
//
//	jvm fix-path  # Pulisce automaticamente il PATH di sistema
func FixPath() {

	fmt.Println("JVM PATH REPAIR UTILITY")
	fmt.Println("==========================")
	fmt.Println()

	// readSystemPath legge la variabile PATH di sistema dal registro Windows.
	// Utilizza il comando reg.exe per accedere alla chiave del registro
	// HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment
	// che contiene le variabili d'ambiente di sistema permanenti.
	cmdPath := "C:\\Windows\\System32\\cmd.exe"
	cmd := exec.Command(cmdPath, "/c", "reg query \"HKLM\\SYSTEM\\CurrentControlSet\\Control\\Session Manager\\Environment\" /v PATH")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("[ERROR] Error reading system PATH: %v\n", err)
		return
	}

	// parseRegistryOutput analizza l'output del comando reg.exe per estrarre
	// il valore della variabile PATH. Il formato dell'output è:
	// "PATH    REG_EXPAND_SZ    C:\Windows\system32;C:\Windows;..."
	// Viene cercata la riga contenente "PATH" e "REG_" per identificare
	// la definizione della variabile ed estrarre il valore dopo "REG_EXPAND_SZ".
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

	fmt.Printf("Current SYSTEM PATH entries: %d\n", len(strings.Split(currentPath, ";")))

	// cleanPathEntries esegue la logica di pulizia del PATH rimuovendo:
	// 1. Voci duplicate (confronto case-insensitive per Windows)
	// 2. Voci vuote o contenenti solo spazi bianchi
	// 3. Mantiene l'ordine originale delle voci valide
	//
	// Utilizza una mappa per tracciare le voci già viste e contatori
	// per statistiche di pulizia che verranno mostrate all'utente.
	pathEntries := strings.Split(currentPath, ";")
	var cleanEntries []string
	seen := make(map[string]bool)
	duplicatesRemoved := 0
	emptyEntriesRemoved := 0

	for _, entry := range pathEntries {
		trimmedEntry := strings.TrimSpace(entry)

		// Rimuove voci vuote o contenenti solo spazi bianchi.
		// Queste possono accumularsi nel PATH a causa di modifiche
		// manuali o installazioni/disinstallazioni di software.
		if trimmedEntry == "" {
			emptyEntriesRemoved++
			continue
		}

		// Controllo duplicati utilizzando confronto case-insensitive.
		// Su Windows i percorsi non sono case-sensitive, quindi
		// "C:\Windows" e "c:\windows" sono considerati duplicati.
		upperEntry := strings.ToUpper(trimmedEntry)
		if seen[upperEntry] {
			duplicatesRemoved++
			fmt.Printf("[CLEAN] Removing duplicate: %s\n", trimmedEntry)
			continue
		}

		seen[upperEntry] = true
		cleanEntries = append(cleanEntries, trimmedEntry)
	}

	// Verifica se è necessaria la pulizia. Se il PATH è già pulito,
	// non è necessario modificare il registro di sistema.
	if duplicatesRemoved == 0 && emptyEntriesRemoved == 0 {
		fmt.Println("[SUCCESS] PATH is already clean, no duplicates found")
		return
	}

	// Ricostruisce il PATH pulito utilizzando il separatore ";" di Windows.
	// Mantiene l'ordine originale delle voci, rimuovendo solo i duplicati.
	cleanPath := strings.Join(cleanEntries, ";")

	fmt.Printf("\n[SUMMARY] CLEANUP SUMMARY:\n")
	fmt.Printf("   • Duplicate entries removed: %d\n", duplicatesRemoved)
	fmt.Printf("   • Empty entries removed: %d\n", emptyEntriesRemoved)
	fmt.Printf("   • Final PATH entries: %d\n", len(cleanEntries))
	fmt.Printf("\n[UPDATE] Updating SYSTEM PATH in registry...\n")

	// updateSystemPath scrive il PATH pulito nel registro di sistema Windows.
	// Utilizza il comando reg.exe per aggiornare la chiave HKLM che contiene
	// le variabili d'ambiente di sistema. Il tipo REG_EXPAND_SZ permette
	// l'espansione di variabili d'ambiente come %SYSTEMROOT%.
	//
	// IMPORTANTE: Questa operazione richiede privilegi amministratore perché
	// modifica una chiave di sistema in HKEY_LOCAL_MACHINE.
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
	fmt.Printf("You can verify the changes by running: echo $PATH\n")
}
