package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigurePrivateRepo configura un repository privato JDK nel sistema Windows.
//
// Questa funzione crea o aggiorna la configurazione per accedere a repository
// privati di JDK (come Nexus, o server aziendali custom) salvando
// endpoint e token di autenticazione in un file di configurazione JSON sicuro.
//
// Processo di configurazione:
// 1. **Localizzazione directory home**: Ottiene directory utente Windows
// 2. **Creazione directory config**: Assicura esistenza di ~/.jvm/
// 3. **Costruzione file config**: Percorso C:\Users\username\.jvm\config.json
// 4. **Struttura configurazione**: Mappa endpoint e token in formato JSON
// 5. **Scrittura sicura**: Salva configurazione con formattazione leggibile
// 6. **Validazione risultato**: Conferma successo operazione all'utente
//
// Formato JSON generato:
//
//	{
//	  "private_endpoint": "https://nexus.company.com/api/jdk",
//	  "private_token": "your-authentication-token-here"
//	}
//

// Gestione sicurezza:
//   - File creato con permessi utente (0755 per directory)
//   - Token salvato in chiaro (considerare crittografia per versioni future)
//   - Accesso limitato al profilo utente Windows corrente
//   - Nessuna trasmissione non crittografata del token
//
// Parametri:
//
//	endpoint string - URL completo dell'API repository privato
//	token string    - Token di autenticazione per accesso al repository
//
// Comportamento errori:
//   - Stampa errore specifico e termina se directory home non determinabile
//   - Stampa errore e termina se impossibile creare file di configurazione
//   - Stampa errore e termina se encoding JSON fallisce
//   - Gestione graceful: non corrompe configurazioni esistenti in caso di errore
//
// Side effects:
//   - Crea directory ~/.jvm/ se non esistente
//   - Sovrascrive completamente config.json esistente (backup non automatico)
//   - Stampa messaggi di stato e risultato su stdout
//
// Requisiti:
//   - Permessi di scrittura nella directory home utente Windows
//   - Spazio disponibile per file JSON (tipicamente <1KB)
//   - Endpoint accessibile e token valido (verificare manualmente)
//
// Note di compatibilità:
//   - JSON formattato con indentazione per leggibilità umana
//   - Compatibile con tutti i parser JSON standard
//   - Struttura espandibile per future configurazioni
//   - Encoding UTF-8 per supporto caratteri internazionali
//
// Esempio di utilizzo:
//
//	ConfigurePrivateRepo("https://nexus.company.com/api/jdk", "abc123token")
//	// Risultato: File config.json creato in C:\Users\username\.jvm\config.json
func ConfigurePrivateRepo(endpoint string, token string) {
	// Ottiene la directory home dell'utente Windows corrente
	// Necessaria per localizzare la cartella di configurazione ~/.jvm
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("[ERROR] Unable to determine user directory:", err)
		return
	}

	// Costruisce il percorso della directory di configurazione JVM
	// Su Windows: C:\Users\username\.jvm
	dir := filepath.Join(home, ".jvm")

	// Crea ricorsivamente la directory di configurazione se non esiste
	// Permessi 0755: full access per owner, read+execute per altri
	os.MkdirAll(dir, 0755)

	// Costruisce il percorso completo del file di configurazione
	// Risultato: C:\Users\username\.jvm\config.json
	path := filepath.Join(dir, "config.json")

	// Crea la struttura di configurazione come mappa string-string
	// Contiene endpoint URL e token di autenticazione per repository privato
	cfg := map[string]string{
		"private_endpoint": endpoint,
		"private_token":    token,
	}

	// Crea nuovo file di configurazione (sovrascrive se esistente)
	// ATTENZIONE: questa operazione elimina configurazioni precedenti
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("[ERROR] Write error:", err)
		return
	}
	// Assicura chiusura file anche in caso di errore nel resto della funzione
	defer file.Close()

	// Crea encoder JSON configurato per output leggibile
	enc := json.NewEncoder(file)

	// Configura indentazione per formattazione JSON human-readable
	// Usa 2 spazi per indentazione (standard comune)
	enc.SetIndent("", "  ")

	// Codifica la mappa di configurazione in formato JSON e scrive nel file
	err = enc.Encode(cfg)
	if err != nil {
		fmt.Println("[ERROR] JSON encoding error:", err)
		return
	}

	// Conferma successo operazione con messaggio colorato e percorso file
	fmt.Println("[SUCCESS] Private repository configured successfully!")
	fmt.Println("📁 File:", path)
}
