package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"jenvy/internal/utils"
)

// ShowCurrentConfig visualizza la configurazione corrente dei repository privati nel sistema Windows.
//
// Questa funzione implementa la lettura sicura e la presentazione strutturata delle
// impostazioni di configurazione per repository JDK privati, fornendo visibilità
// completa sulle configurazioni attive nell'ambiente Windows:
//
// **Scopo della funzione:**
// - Lettura e parsing del file di configurazione JSON repository privati
// - Visualizzazione formattata delle impostazioni correnti
// - Validazione integrità configurazione e gestione errori
// - Supporto per troubleshooting e verifica configurazione enterprise
//
// **Gestione sicurezza Windows:**
// - Accesso controllato al profilo utente Windows (%USERPROFILE%)
// - Lettura sicura del file config.json dalla directory .jenvy
// - Gestione permessi Windows per file di configurazione
// - Protezione da accesso non autorizzato a credenziali sensibili
//
// **Integrazione ambiente Windows:**
// - Compatibilità con profili utente roaming Windows Domain
// - Supporto per percorsi Windows con caratteri speciali e spazi
// - Gestione encoding UTF-8 per configurazioni internazionali
// - Integrazione con Windows Terminal e CMD per output formattato
//
// **Struttura configurazione supportata:**
// - URL repository privati (HTTP/HTTPS con autenticazione)
// - Credenziali di accesso (username/password, token, certificati)
// - Impostazioni proxy per ambienti enterprise Windows
// - Configurazioni SSL/TLS per repository sicuri
//
// **Presentazione informazioni:**
// - Output strutturato e leggibile delle configurazioni
// - Gestione valori vuoti con indicatori appropriati
// - Formattazione consistent con altri comandi jenvy
// - Supporto per redirection output per scripting Windows
//
// **Gestione errori robusta:**
// - Verifica esistenza file configurazione
// - Validazione formato JSON e struttura dati
// - Messaggi di errore informativi per troubleshooting
// - Fallback graceful per configurazioni corrotte o incomplete
//
// **Casi d'uso tipici:**
// - Verifica configurazione repository enterprise prima del download
// - Troubleshooting problemi di autenticazione
// - Audit configurazioni in ambienti multi-utente Windows
// - Validazione setup prima di operazioni automatizzate
//
// La funzione garantisce accesso sicuro alle informazioni di configurazione
// senza esporre dati sensibili in plain text quando non necessario.
func ShowCurrentConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Unable to access user directory: %v", err))
		utils.PrintInfo("Cannot locate Windows user profile directory")
		utils.PrintInfo("Ensure proper user permissions and try again")
		return
	}

	configPath := filepath.Join(homeDir, ".jenvy", "config.json")
	jenvyDir := filepath.Join(homeDir, ".jenvy")

	// Verifica esistenza directory .jenvy
	if _, err := os.Stat(jenvyDir); os.IsNotExist(err) {
		utils.PrintInfo("Jenvy configuration directory not found")
		utils.PrintInfo("No private repository has been configured yet")
		utils.PrintInfo("Use 'jenvy configure private <URL>' to set up a repository")
		return
	}

	// Verifica esistenza file configurazione
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		utils.PrintInfo("Private repository configuration not found")
		utils.PrintInfo("No configuration file exists")
		utils.PrintInfo("Use 'jenvy configure private <URL>' to set up a repository")
		return
	}

	// Apri e leggi il file di configurazione
	file, err := os.Open(configPath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Unable to read configuration file: %v", err))
		utils.PrintInfo("This may be due to:")
		utils.PrintInfo("  - File is locked by another process")
		utils.PrintInfo("  - Insufficient Windows permissions")
		utils.PrintInfo("  - File corruption or access restrictions")
		utils.PrintInfo("Try closing applications or running as administrator")
		return
	}
	defer file.Close()

	// Parse del contenuto JSON
	var cfg map[string]string
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Configuration file parsing error: %v", err))
		utils.PrintInfo("The configuration file appears to be corrupted")
		utils.PrintInfo("Consider using 'jenvy reset-config' to reset configuration")
		utils.PrintInfo("Then reconfigure with 'jenvy configure private <URL>'")
		return
	}

	// Verifica che la configurazione non sia vuota
	if len(cfg) == 0 {
		utils.PrintInfo("Configuration file is empty")
		utils.PrintInfo("No private repository settings found")
		utils.PrintInfo("Use 'jenvy configure private <URL>' to set up a repository")
		return
	}

	// Visualizza la configurazione corrente
	utils.PrintInfo("Current Private Repository Configuration:")
	fmt.Println()

	// Ordina le chiavi per una presentazione consistente
	keys := make([]string, 0, len(cfg))
	for key := range cfg {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := cfg[key]
		displayValue := value
		if value == "" {
			displayValue = utils.ColorText("(empty)", utils.Yellow)
		} else if key == "password" || key == "token" || key == "api_key" {
			// Maschera valori sensibili per sicurezza
			displayValue = utils.ColorText("(configured - hidden for security)", utils.Green)
		} else {
			displayValue = utils.ColorText(value, utils.Cyan)
		}

		keyFormatted := utils.ColorText(fmt.Sprintf("%-15s", key+":"), utils.Blue)
		fmt.Printf("  %s %s\n", keyFormatted, displayValue)
	}

	fmt.Println()
	utils.PrintInfo("Use 'jenvy reset-config' to clear configuration")
	utils.PrintInfo("Use 'jenvy configure private <URL>' to update repository")
}
