[Setup]
AppName=Java Version Manager
AppVersion=1.0
DefaultDirName={commonpf}\JVM
DefaultGroupName=JVM CLI
DisableDirPage=yes
DisableProgramGroupPage=yes
OutputDir=.
OutputBaseFilename=jvm-installer
Compression=lzma
SolidCompression=yes
SetupIconFile=jvm.ico
WizardImageFile=jvm_splash.bmp
WizardSmallImageFile=jvm_splash_small.bmp
LanguageDetectionMethod=locale
; SignTool="\"C:\\Program Files (x86)\\Windows Kits\\10\\bin\\x64\\signtool.exe\" sign /f \"distribution\\jvm-dev-cert.pfx\" /p jvm-password /tr http://timestamp.digicert.com /td sha256 $f"



[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "italian"; MessagesFile: "compiler:Languages\Italian.isl"

[Files]
Source: "jvm.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "jvm.ico"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
  ValueType: string; ValueName: "Path"; ValueData: "{olddata};{app}"; Flags: preservestringtype uninsdeletevalue

[Run]
Filename: "{app}\jvm.exe"; Description: "{cm:LaunchProgram,JVM}"; Flags: postinstall skipifsilent
