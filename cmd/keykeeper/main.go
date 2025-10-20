package main

import (
    "fmt"
    "log"

    "keykeeper/internal/db"
    "keykeeper/internal/ui"
)

func main() {
    database, dbPath, err := db.Open()
    if err != nil {
        log.Fatalf("failed to open database: %v", err)
    }
    defer database.Close()

    fmt.Printf("Offline PhraseVault - Database: %s\n", dbPath)
    if err := ui.RunMenu(database); err != nil {
        log.Fatalf("error: %v", err)
    }
}


