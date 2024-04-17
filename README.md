## Cross-compiling from windows powershell
```powershell
$Env:GOOS = "linux"; $Env:GOARCH = "amd64"
$Env:TELEGRAM_BOT_TOKEN = "0000000000:ABCDEFGHIJKLMNOPQRSTUVWXYZABEXAMPLE"
$Env:CONFIG_FILE_PATH = "/path/to/config.json"
go build main.go
```

## Running the bot on debian and probably many other linux distributions
```bash
env TELEGRAM_BOT_TOKEN=0000000000:ABCDEFGHIJKLMNOPQRSTUVWXYZABEXAMPLE CONFIG_FILE_PATH=/path/to/config.json /path/to/your/executable
```
