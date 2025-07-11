Set objFSO = CreateObject("Scripting.FileSystemObject")
strFolder = objFSO.GetParentFolderName(WScript.ScriptFullName)

Set shell = CreateObject("Shell.Application")
shell.ShellExecute "cmd.exe", "/c """ & strFolder & "\manager\java-version.bat""", "", "runas", 1
