# 🗝️ Offline PhraseVault (Go)  
**Your Offline Wallet Recovery Phrase Keeper**

---

## 🚀 Features

- **Truly Offline:** No network required; runs locally as a single executable.
- **Compact & Portable:** Single .exe terminal app, Windows 7–11 support.
- **Secure Storage:** Saves 24-word wallet phrases in an encrypted SQLite database.
- **PIN Protection:** Numeric PIN lock to prevent unauthorized access.
- **Custom Obfuscation:** Extra reversible layer to deter casual snooping.

---

## 🛠️ Quick Start (Windows 64-bit)

### 1. Prerequisites
- Install [Go 1.19+](https://go.dev/dl/)

### 2. Build Instructions
Open **PowerShell** in your project root and run:

```powershell
go mod tidy
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -trimpath -ldflags "-s -w" -o offline-phrasevault.exe ./cmd/keykeeper
```

### 3. Run the App
- **Double-click** `offline-phrasevault.exe`  
  *or*
- Run from terminal:  
  ```powershell
  .\offline-phrasevault.exe
  ```

---

## 💾 Data Location

- Your data is stored at:  
  `%APPDATA%\OfflinePhraseVault\offline-phrasevault.db`
- If `%APPDATA%` does not exist, it falls back to the local folder.

---

## ⚠️ Security Notice

- Obfuscation is intended to protect against *casual* inspection only.
- **Not suitable** for advanced attackers or if adversaries have full system access.

---

## 📝 Windows 7 Users

- For best compatibility, **use Go 1.19 toolchain** when building.

---

Enjoy secure, offline protection for your wallet recovery phrases!  
Questions or issues? Create an issue on GitHub.
