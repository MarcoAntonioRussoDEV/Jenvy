╔═════════════════════════════════════════════╗
║       Java Version Manager - README         ║
╚═════════════════════════════════════════════╝

📁 STRUTTURA DELLA CARTELLA:

JavaVersionManager
├── launcher.vbs              → Avvio principale (PowerShell GUI con diritti admin)
├── java-version-launcher.vbs → Avvio alternativo (Batch con diritti admin)
├── README.txt                → Questo file!
├── manager
│   ├── java-manager.ps1      → Script PowerShell con selezione grafica
│   ├── java-version.bat      → Script compatibile con cmd
│   └── assets
│       └── java.ico          → Icona personalizzata

🧩 COME SI USA:

① Doppio clic su `launcher.vbs`  
   → Si aprirà una finestra grafica per selezionare la versione di Java installata  
   → La variabile JAVA_HOME verrà aggiornata automaticamente  
   → Verrà creato un collegamento "Gestione Java" sul desktop con