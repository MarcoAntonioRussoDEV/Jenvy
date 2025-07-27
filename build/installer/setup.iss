#define MyAppVersion "1.0.0"

[Setup]
AppName=Jenvy - Developer Kit Manager
AppVersion={#MyAppVersion}
DefaultDirName={autopf}\Jenvy
DefaultGroupName=Jenvy - Developer Kit Manager
DisableProgramGroupPage=yes
OutputBaseFilename=jenvy-installer-{#MyAppVersion}
OutputDir=..\..\release
ChangesEnvironment=yes
PrivilegesRequired=admin

SetupIconFile=..\..\assets\icons\jenvy.ico
WizardImageFile=..\..\assets\splash\jenvy_splash.bmp
WizardSmallImageFile=..\..\assets\splash\jenvy_splash_small.bmp

; Command line parameters
; /CONFIGURE_PRIVATE=1  - Enable private repository configuration
; /CONFIGURE_PRIVATE=0  - Skip private repository configuration

[Files]
Source: "..\..\build\dist\jenvy.exe";               DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\README.md";               DestDir: "{app}"; Flags: ignoreversion

[Registry]
; PATH viene gestito nella sezione [Code] per un controllo più preciso

[Tasks]
Name: "addtopath"; Description: "Add Jenvy to system PATH"; GroupDescription: "Additional configuration:"

[Run]
Filename: "{app}\jenvy.exe"; Parameters: "init"; WorkingDir: "{app}"; Flags: runhidden waituntilterminated skipifdoesntexist; Description: "Initialize Jenvy environment"; StatusMsg: "Setting up Jenvy configuration and shell completions..."
Filename: "notepad.exe"; Parameters: """{app}\README.md"""; Description: "📘 Open README"; Flags: postinstall shellexec unchecked


[Code]
var
  InputPage: TInputQueryWizardPage;
  WelcomePage: TOutputMsgWizardPage;
  ConfigurePrivate: Boolean;

// Forward declarations
procedure CleanupPATH(); forward;
function NeedsAddPath(Param: string): Boolean; forward;

procedure InitializeWizard;
var
  ExistingInstallPath: string;
  ResultCode: Integer;
begin
  // Check if configuring private repository (default: True)
  ConfigurePrivate := StrToIntDef(ExpandConstant('{param:CONFIGURE_PRIVATE|1}'), 1) = 1;
  
  // Check for existing jenvy installation
  if RegQueryStringValue(HKLM, 'SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Jenvy - Developer Kit Manager_is1', 'InstallLocation', ExistingInstallPath) or
     RegQueryStringValue(HKCU, 'SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Jenvy - Developer Kit Manager_is1', 'InstallLocation', ExistingInstallPath) then
  begin
    if MsgBox('An existing Jenvy - Developer Kit Manager installation has been detected at:' + #13#10 + 
              ExistingInstallPath + #13#10#13#10 + 
              'Do you want to uninstall it first and proceed with a fresh installation?' + #13#10 +
              '(Recommended to avoid conflicts)', mbConfirmation, MB_YESNO) = IDYES then
    begin
      // Try to run uninstaller
      if FileExists(ExistingInstallPath + '\unins000.exe') then
      begin
        if not Exec(ExistingInstallPath + '\unins000.exe', '/SILENT', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
          MsgBox('Could not automatically uninstall the previous version. Please uninstall manually.', mbError, MB_OK);
      end;
    end;
  end;
  
  // Initial welcome page
  WelcomePage := CreateOutputMsgPage(
    wpWelcome,
    '☕ Welcome to Jenvy - Developer Kit Manager',
    'A professional OpenJDK management solution for Windows',
    '🌟 Jenvy is an open source project for managing Java Development Kits.' + #13#10 +
    '📋 GitHub: https://github.com/MarcoAntonioRussoDEV/Jenvy' + #13#10 +
    '📖 Documentation: Full README with examples and guides' + #13#10 +
    '💖 Support: GitHub Sponsors, Ko-fi, PayPal donations welcome' + #13#10 + #13#10 +
    '🚀 Key Features:' + #13#10 +
    '• Multi-provider support (Adoptium, Azul Zulu, BellSoft Liberica)' + #13#10 +
    '• Private enterprise repository integration' + #13#10 +
    '• Smart version selection with LTS priority' + #13#10 +
    '• Automatic JAVA_HOME and PATH management' + #13#10 +
    '• Shell completions (Bash, PowerShell, CMD)' + #13#10 +
    '• Professional CLI with formatted tables' + #13#10 + #13#10 +
    '📦 After installation, Jenvy will be available globally.' + #13#10 +
    '🔧 The installer will automatically run "jenvy init" to configure your environment.' + #13#10 + #13#10 +
    '� Quick Start Examples:' + #13#10 +
    '  jenvy remote-list          # Browse available JDK versions' + #13#10 +
    '  jenvy download 21          # Download JDK 21 LTS' + #13#10 +
    '  jenvy use 21              # Set JDK 21 as active' + #13#10 +
    '  jenvy list                # Show installed versions' + #13#10 +
    '  jenvy --help              # Full command reference'
  );

  // Private repository configuration page (conditional)
  if ConfigurePrivate then
  begin
    InputPage := CreateInputQueryPage(
      wpSelectDir,
      '🔒 Private Repository Configuration',
      'Configure access to your enterprise repository',
      'These parameters will be saved to %USERPROFILE%\.jenvy\config.json' + #13#10 + #13#10 +
      '⚠️ You can leave empty to configure later with:' + #13#10 +
      '   jenvy configure-private <endpoint> [token]'
    );
    InputPage.Add('Repository endpoint (e.g. https://nexus.company.com/api/jdk):', False);
    InputPage.Add('Access token (optional):', False);
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Endpoint, Token: string;
  ConfigPath, JSON: string;
  CurrentPath: string;
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    // Clean PATH from duplicates first
    CleanupPATH();
    
    // Add jenvy to PATH if not already present
    if NeedsAddPath(ExpandConstant('{app}')) then
    begin
      if RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'PATH', CurrentPath) then
      begin
        if CurrentPath <> '' then
          CurrentPath := CurrentPath + ';' + ExpandConstant('{app}')
        else
          CurrentPath := ExpandConstant('{app}');
        
        RegWriteExpandStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'PATH', CurrentPath);
        Log(Format('Added to SYSTEM PATH: %s', [ExpandConstant('{app}')]));
      end;
    end;
    
    // Configure private repository if requested
    if ConfigurePrivate then
    begin
      // Read entered values only if page was shown
      Endpoint := InputPage.Values[0];
      Token    := InputPage.Values[1];

      // If user entered at least the endpoint, save config
      if Endpoint <> '' then
      begin
        // Prepare JSON
        JSON :=
          '{' + #13#10 +
          '  "private_endpoint": "' + Endpoint + '",' + #13#10 +
          '  "private_token": "'   + Token    + '"'  + #13#10 +
          '}';

        // Save to %USERPROFILE%\.jenvy\config.json
        ConfigPath := ExpandConstant('{%USERPROFILE}\.jenvy\config.json');
        ForceDirectories(ExtractFileDir(ConfigPath));
        SaveStringToFile(ConfigPath, JSON, False);
      end;
    end;
  end;
end;

function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
  CleanPath: string;
  SearchPattern1, SearchPattern2: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'PATH', OrigPath) then
    begin
      Result := True;
      exit;
    end;
  
  // Normalizza il PATH per la ricerca
  CleanPath := ';' + Uppercase(OrigPath) + ';';
  
  // Normalizza anche la directory da cercare
  SearchPattern1 := ';' + Uppercase(Param) + ';';
  SearchPattern2 := ';' + Uppercase(Param) + '\;';  // Con backslash finale
  
  // Check if path is already present (case insensitive)
  Result := (Pos(SearchPattern1, CleanPath) = 0) and (Pos(SearchPattern2, CleanPath) = 0);
  
  if not Result then
    Log(Format('PATH already contains: %s', [Param]));
end;

procedure CleanupPATH();
var
  OrigPath, NewPath: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'PATH', OrigPath) then
    exit;
    
  // Simple cleanup: just remove double semicolons
  NewPath := OrigPath;
  while Pos(';;', NewPath) > 0 do
    StringChange(NewPath, ';;', ';');
  
  // Remove leading/trailing semicolons
  while (Length(NewPath) > 0) and (NewPath[1] = ';') do
    Delete(NewPath, 1, 1);
  while (Length(NewPath) > 0) and (NewPath[Length(NewPath)] = ';') do
    Delete(NewPath, Length(NewPath), 1);
  
  // Update registry only if changed
  if NewPath <> OrigPath then
  begin
    RegWriteExpandStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'PATH', NewPath);
    Log('SYSTEM PATH cleaned: removed empty entries');
  end;
end;