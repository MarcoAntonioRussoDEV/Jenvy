[Setup]
AppName=Java Version Manager
AppVersion=1.0
DefaultDirName={autopf}\JVM
DefaultGroupName=Java Version Manager
DisableProgramGroupPage=yes
OutputBaseFilename=jvm-installer
OutputDir=distribution
ChangesEnvironment=yes
PrivilegesRequired=lowest

SetupIconFile=distribution\jvm.ico
WizardImageFile=distribution\jvm_splash.bmp
WizardSmallImageFile=distribution\jvm_splash_small.bmp

; Parametri riga di comando
; /CONFIGURE_PRIVATE=1  - Abilita configurazione repository privato
; /CONFIGURE_PRIVATE=0  - Salta configurazione repository privato

[Files]
Source: "distribution\jvm.exe";               DestDir: "{app}"; Flags: ignoreversion
Source: "distribution\README.txt";            DestDir: "{app}"; Flags: ignoreversion

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "PATH"; ValueData: "{olddata};{app}"; Check: NeedsAddPath('{app}')

[Tasks]
Name: "addtopath"; Description: "Aggiungi JVM al PATH di sistema"; GroupDescription: "Configurazione aggiuntiva:"

[Run]
Filename: "notepad.exe"; Parameters: """{app}\README.txt"""; Description: "📘 Leggi il README"; Flags: postinstall shellexec


[Code]
var
  InputPage: TInputQueryWizardPage;
  WelcomePage: TOutputMsgWizardPage;
  ConfigurePrivate: Boolean;

procedure InitializeWizard;
begin
  // Controlla se configurare repository privato (default: True)
  ConfigurePrivate := StrToIntDef(ExpandConstant('{param:CONFIGURE_PRIVATE|1}'), 1) = 1;
  
  // Pagina di presentazione iniziale
  WelcomePage := CreateOutputMsgPage(
    wpWelcome,
    '☕ Benvenuto in Java Version Manager',
    'Un tool elegante per gestire versioni OpenJDK',
    '🚀 Funzionalità principali:' + #13#10 +
    '• Elenco JDK da Adoptium, Azul, Liberica' + #13#10 +
    '• Supporto repository privati aziendali' + #13#10 +
    '• Selezione intelligente versioni (LTS prioritario)' + #13#10 +
    '• Interfaccia CLI con tabelle formattate' + #13#10 + #13#10 +
    '📦 Dopo l''installazione potrai usare il comando "jvm" da qualsiasi terminale.' + #13#10 + #13#10 +
    '🔧 Esempi d''uso:' + #13#10 +
    '  jvm remote-list' + #13#10 +
    '  jvm remote-list --provider=azul' + #13#10 +
    '  jvm remote-list --all'
  );

  // Pagina configurazione repository privato (condizionale)
  if ConfigurePrivate then
  begin
    InputPage := CreateInputQueryPage(
      wpSelectDir,
      '🔒 Configurazione Repository Privato',
      'Configura l''accesso al tuo repository aziendale',
      'Questi parametri verranno salvati in %USERPROFILE%\.jvm\config.json' + #13#10 + #13#10 +
      '⚠️ Puoi lasciare vuoto per configurare successivamente con:' + #13#10 +
      '   jvm configure-private <endpoint> [token]'
    );
    InputPage.Add('Endpoint del repository (es. https://nexus.company.com/api/jdk):', False);
    InputPage.Add('Token di accesso (opzionale):', False);
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Endpoint, Token: string;
  ConfigPath, JSON: string;
begin
  if (CurStep = ssPostInstall) and ConfigurePrivate then
  begin
    // Legge i valori inseriti solo se la pagina è stata mostrata
    Endpoint := InputPage.Values[0];
    Token    := InputPage.Values[1];

    // Se l'utente ha inserito almeno l'endpoint, salva il config
    if Endpoint <> '' then
    begin
      // Prepara il JSON
      JSON :=
        '{' + #13#10 +
        '  "private_endpoint": "' + Endpoint + '",' + #13#10 +
        '  "private_token": "'   + Token    + '"'  + #13#10 +
        '}';

      // Salva in %USERPROFILE%\.jvm\config.json
      ConfigPath := ExpandConstant('{%USERPROFILE}\.jvm\config.json');
      ForceDirectories(ExtractFileDir(ConfigPath));
      SaveStringToFile(ConfigPath, JSON, False);
    end;
  end;
end;

function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'PATH', OrigPath) then
    begin
      Result := True;
      exit;
    end;
  // Controlla se il percorso è già presente
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;