' TailRouter Silent Background Launcher for Windows
Set WshShell = CreateObject("WScript.Shell")
Set FSO = CreateObject("Scripting.FileSystemObject")
ScriptDir = FSO.GetParentFolderName(WScript.ScriptFullName)
RootDir = FSO.GetParentFolderName(FSO.GetParentFolderName(ScriptDir))

' Chạy PowerShell System Tray hoặc server.py ngầm (0 = ẩn hoàn toàn cửa sổ)
WshShell.Run "powershell -ExecutionPolicy Bypass -WindowStyle Hidden -File """ & ScriptDir & "\TailRouter-Tray.ps1""", 0, False
