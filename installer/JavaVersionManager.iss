[Setup]
AppName=Java Version Manager
AppVersion=1.0.0
AppPublisher=Marco Antonio Russo
AppPublisherURL=https://github.com/tuoutente/java-version-manager
AppSupportURL=https://github.com/tuoutente/java-version-manager/issues
AppUpdatesURL=https://github.com/tuoutente/java-version-manager/releases
DefaultDirName={autopf}\Java Version Manager
DefaultGroupName=Java Version Manager
AllowNoIcons=yes
LicenseFile=license.txt
InfoAfterFile=readme_after_install.txt
OutputDir=..\dist
OutputBaseFilename=JavaVersionManager-Setup
SetupIconFile=..\manager\assets\java.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesInstallIn64BitMode=x64

[Languages]
Name: "italian"; MessagesFile: "compiler:Languages\Italian.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1

[Files]
Source: "..\launcher.vbs"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\launcher-cmd.vbs"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\manager\*"; DestDir: "{app}\manager"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\Java Version Manager"; Filename: "{app}\launcher.vbs"; IconFilename: "{app}\manager\assets\java.ico"
Name: "{group}\{cm:UninstallProgram,Java Version Manager}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\Java Version Manager"; Filename: "{app}\launcher.vbs"; IconFilename: "{app}\manager\assets\java.ico"; Tasks: desktopicon
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\Java Version Manager"; Filename: "{app}\launcher.vbs"; IconFilename: "{app}\manager\assets\java.ico"; Tasks: quicklaunchicon

[Run]
Filename: "{app}\launcher.vbs"; Description: "{cm:LaunchProgram,Java Version Manager}"; Flags: shellexec postinstall skipifsilent

[UninstallDelete]
Type: files; Name: "{autodesktop}\Gestione Java.lnk"
