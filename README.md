Offline PhraseVault (Go) - Offline Wallet Phrase Keeper

Overview
- Offline, single-EXE terminal app for Windows 7–11
- Stores 24-word wallet phrases in SQLite
- Numeric PIN lock; reversible custom obfuscation

Build (Windows 64-bit)
1. Install Go 1.19+
2. In PowerShell from project root:
   - go mod tidy
   - set GOOS=windows
   - set GOARCH=amd64
   - go build -trimpath -ldflags "-s -w" -o offline-phrasevault.exe ./cmd/keykeeper

Run
- Double-click `offline-phrasevault.exe` or run in terminal.

Notes for Windows 7
- Prefer Go 1.19 toolchain to ensure compatibility.

Data location
- `%APPDATA%/OfflinePhraseVault/offline-phrasevault.db` (fallback to local folder if APPDATA missing).

Security note
- Obfuscation protects against casual inspection only. Not suitable against skilled attackers with full system access.


