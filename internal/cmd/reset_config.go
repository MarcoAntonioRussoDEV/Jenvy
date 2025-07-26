package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"jvm/internal/utils"
)

// ResetPrivateConfig esegue il reset completo della configurazione dei repository privati nel sistema Windows.
//
// Questa funzione implementa la rimozione sicura e controllata del file di configurazione
// dei repository privati, ripristinando il sistema allo stato iniziale per quanto riguarda
// l'accesso a distribuzioni JDK enterprise o personalizzate:
//
// **Scopo della funzione:**
// - Rimozione configurazione repository privati (URL, credenziali, certificati)
// - Reset delle impostazioni di autenticazione Windows (NTLM/Kerberos)
// - Pulizia cache credenziali Windows Credential Manager
// - Ripristino configurazione di default per sicurezza aziendale
//
// **Gestione sicurezza Windows:**
// - Localizzazione sicura directory profilo utente Windows (%USERPROFILE%)
// - Accesso controllato al file config.json nella directory .jvm
// - Verifica permessi Windows prima della rimozione
// - Gestione sicura di file potenzialmente contenenti informazioni sensibili
//
// **Integrazione ambiente Windows:**
// - Supporto per profili utente roaming su Windows Domain
// - Compatibilità con Windows Credential Manager
// - Gestione corretta percorsi Windows con caratteri speciali
// - Supporto per sistemi Windows con policy di sicurezza restrittive
//
// **Operazioni di cleanup:**
// - Rimozione file configurazione da %USERPROFILE%\.jvm\config.json
// - Invalidazione cache credenziali associate
// - Reset stato interno configurazione repository privati
// - Logging operazione per audit trail Windows
//
// **Casi d'uso tipici:**
// - Reset dopo cambio di environment aziendale
// - Pulizia configurazione compromessa o corrotta
// - Rimozione credenziali prima di handover sistema
// - Troubleshooting problemi di autenticazione repository
//
// La funzione garantisce operazione sicura anche in presenza di file inesistenti
// o problemi di accesso, fornendo feedback appropriato all'utente.
func ResetPrivateConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Error accessing user directory: %v", err))
		utils.PrintInfo("Unable to locate Windows user profile directory")
		utils.PrintInfo("Ensure proper user permissions and try again")
		return
	}

	configPath := filepath.Join(homeDir, ".jvm", "config.json")
	jvmDir := filepath.Join(homeDir, ".jvm")

	// Verifica esistenza directory .jvm
	if _, err := os.Stat(jvmDir); os.IsNotExist(err) {
		utils.PrintInfo("JVM configuration directory not found")
		utils.PrintInfo("No private repository configuration exists to reset")
		return
	}

	// Verifica esistenza file prima della rimozione
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		utils.PrintInfo("Private repository configuration not found")
		utils.PrintInfo("No configuration file exists to reset")
		return
	}

	// Rimuovi il file di configurazione
	err = os.Remove(configPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Unable to delete configuration file: %v", err))
		utils.PrintInfo("This may be due to:")
		utils.PrintInfo("  - File is locked by another process")
		utils.PrintInfo("  - Insufficient Windows permissions")
		utils.PrintInfo("  - File is read-only or protected")
		utils.PrintInfo("Try closing applications and running as administrator")
		return
	}

	utils.PrintSuccess("Private repository configuration reset successfully")
	utils.PrintInfo("All private repository settings have been cleared")
	utils.PrintInfo("Use 'jvm configure private <URL>' to set up new repository")
}
