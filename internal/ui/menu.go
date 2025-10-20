package ui

import (
    "bufio"
    "database/sql"
    "fmt"
    "os"
    "strconv"
    "strings"

    "keykeeper/internal/wallet"
)

func RunMenu(db *sql.DB) error {
    in := bufio.NewReader(os.Stdin)
    for {
        fmt.Println()
        fmt.Println("Offline PhraseVault - Menu")
        fmt.Println("1) Add wallet")
        fmt.Println("2) List wallets")
        fmt.Println("3) View wallet")
        fmt.Println("4) Delete wallet")
        fmt.Println("5) Delete word in wallet")
        fmt.Println("6) Export wallet")
        fmt.Println("7) Quit")
        fmt.Print("Choose: ")
        choice, _ := in.ReadString('\n')
        choice = strings.TrimSpace(choice)
        switch choice {
        case "1":
            if err := addWalletFlow(db, in); err != nil { fmt.Println("Error:", err) }
        case "2":
            if err := listWalletsFlow(db); err != nil { fmt.Println("Error:", err) }
        case "3":
            if err := viewWalletFlow(db, in); err != nil { fmt.Println("Error:", err) }
        case "4":
            if err := deleteWalletFlow(db, in); err != nil { fmt.Println("Error:", err) }
        case "5":
            if err := deleteWordFlow(db, in); err != nil { fmt.Println("Error:", err) }
        case "6":
            if err := exportWalletFlow(db, in); err != nil { fmt.Println("Error:", err) }
        case "7":
            return nil
        default:
            fmt.Println("Invalid choice")
        }
    }
}

func prompt(in *bufio.Reader, label string) (string, error) {
    fmt.Print(label)
    s, err := in.ReadString('\n')
    if err != nil { return "", err }
    return strings.TrimSpace(s), nil
}

func promptPIN(in *bufio.Reader) (string, error) {
    for {
        pin, err := prompt(in, "Enter numeric PIN: ")
        if err != nil { return "", err }
        if pin != "" && isNumeric(pin) {
            conf, err := prompt(in, "Confirm PIN: ")
            if err != nil { return "", err }
            if conf == pin { return pin, nil }
            fmt.Println("PINs do not match")
        } else {
            fmt.Println("PIN must be digits only")
        }
    }
}

func isNumeric(s string) bool {
    for _, r := range s { if r < '0' || r > '9' { return false } }
    return len(s) > 0
}

func parseWords(input string) []string {
    if strings.Contains(input, ",") {
        parts := strings.Split(input, ",")
        out := make([]string, 0, len(parts))
        for _, p := range parts { out = append(out, strings.TrimSpace(p)) }
        return out
    }
    // space separated
    parts := strings.Fields(input)
    return parts
}

func addWalletFlow(db *sql.DB, in *bufio.Reader) error {
    name, err := prompt(in, "Wallet name: ")
    if err != nil { return err }
    pin, err := promptPIN(in)
    if err != nil { return err }
    // choose phrase length
    fmt.Println("Choose phrase length: 12 / 18 / 24")
    lstr, err := prompt(in, "Length: ")
    if err != nil { return err }
    lstr = strings.TrimSpace(lstr)
    if lstr == "" { lstr = "24" }
    n, err := strconv.Atoi(lstr)
    if err != nil || (n != 12 && n != 18 && n != 24) { return fmt.Errorf("invalid length") }
    fmt.Printf("Enter %d words in order (comma or space separated):\n", n)
    line, err := in.ReadString('\n')
    if err != nil { return err }
    words := parseWords(strings.TrimSpace(line))
    if len(words) != n {
        return fmt.Errorf("expected %d words, got %d", n, len(words))
    }
    _, err = wallet.Create(db, name, pin, words)
    if err == nil { fmt.Println("Wallet added.") }
    return err
}

func listWalletsFlow(db *sql.DB) error {
    ws, err := wallet.List(db)
    if err != nil { return err }
    if len(ws) == 0 { fmt.Println("No wallets."); return nil }
    for _, w := range ws {
        fmt.Printf("%d) %s  (%s)\n", w.ID, w.Name, w.CreatedAt.Format("2006-01-02 15:04"))
    }
    return nil
}

func selectWalletID(db *sql.DB, in *bufio.Reader) (int64, error) {
    if err := listWalletsFlow(db); err != nil { return 0, err }
    s, err := prompt(in, "Enter wallet ID: ")
    if err != nil { return 0, err }
    id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
    if err != nil { return 0, fmt.Errorf("invalid ID") }
    return id, nil
}

func viewWalletFlow(db *sql.DB, in *bufio.Reader) error {
    id, err := selectWalletID(db, in)
    if err != nil { return err }
    pin, err := prompt(in, "Enter PIN: ")
    if err != nil { return err }
    words, err := wallet.GetOrderedWords(db, id, pin)
    if err != nil { return err }
    for i, w := range words {
        fmt.Printf("%2d: %s\n", i+1, w)
    }
    return nil
}

func deleteWalletFlow(db *sql.DB, in *bufio.Reader) error {
    id, err := selectWalletID(db, in)
    if err != nil { return err }
    conf, err := prompt(in, "Type DELETE to confirm: ")
    if err != nil { return err }
    if conf != "DELETE" { fmt.Println("Cancelled."); return nil }
    if err := wallet.Delete(db, id); err != nil { return err }
    fmt.Println("Wallet deleted.")
    return nil
}

func deleteWordFlow(db *sql.DB, in *bufio.Reader) error {
    id, err := selectWalletID(db, in)
    if err != nil { return err }
    pin, err := prompt(in, "Enter PIN: ")
    if err != nil { return err }
    s, err := prompt(in, "Index to delete (1-24): ")
    if err != nil { return err }
    idx, err := strconv.Atoi(strings.TrimSpace(s))
    if err != nil { return fmt.Errorf("invalid index") }
    if err := wallet.DeleteWordAt(db, id, pin, idx); err != nil { return err }
    fmt.Println("Word deleted.")
    return nil
}

func exportWalletFlow(db *sql.DB, in *bufio.Reader) error {
    id, err := selectWalletID(db, in)
    if err != nil { return err }
    pin, err := prompt(in, "Enter PIN: ")
    if err != nil { return err }
    words, err := wallet.GetOrderedWords(db, id, pin)
    if err != nil { return err }
    fmt.Println(strings.Join(words, ", "))
    return nil
}


