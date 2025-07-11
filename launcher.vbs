Set shell = CreateObject("WScript.Shell")
Set fso = CreateObject("Scripting.FileSystemObject")

folder = fso.GetParentFolderName(WScript.ScriptFullName)
script = folder & "\manager\java-manager.ps1"

' Esegue PowerShell come amministratore senza -NoExit per permettere la chiusura
shell.Run "powershell.exe -ExecutionPolicy Bypass -Command ""Start-Process powershell -ArgumentList '-ExecutionPolicy Bypass -File \""" & script & "\""' -Verb RunAs""", 0, false
