# Cross-compiling from windows powershell
```powershell
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"
$Env:TELEGRAM_BOT_TOKEN = "0000000000:ABCDEFGHIJKLMNOPQRSTUVWXYZABEXAMPLE"
$Env:CONFIG_FILE_PATH = "/path/to/config.json"
go build main.go
```
