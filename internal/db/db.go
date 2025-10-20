package db

import (
    "database/sql"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "time"

    _ "modernc.org/sqlite"
)

// Open opens (and creates if missing) the SQLite database and applies schema.
func Open() (*sql.DB, string, error) {
    dbPath := defaultDBPath()
    if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
        // fallback to local folder when APPDATA missing or permissions fail
        cwd, _ := os.Getwd()
        dbPath = filepath.Join(cwd, "keykeeper.db")
    }

    d, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, "", err
    }

    // ensure busy timeout and foreign keys
    if _, err := d.Exec("PRAGMA foreign_keys=ON;"); err != nil {
        d.Close()
        return nil, "", err
    }
    if _, err := d.Exec("PRAGMA busy_timeout=5000;"); err != nil {
        d.Close()
        return nil, "", err
    }

    if err := applySchema(d); err != nil {
        d.Close()
        return nil, "", err
    }
    return d, dbPath, nil
}

func defaultDBPath() string {
    appdata := os.Getenv("APPDATA")
    if appdata == "" {
        cwd, _ := os.Getwd()
        return filepath.Join(cwd, "offline-phrasevault.db")
    }
    return filepath.Join(appdata, "OfflinePhraseVault", "offline-phrasevault.db")
}

func applySchema(d *sql.DB) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS wallets(
            id INTEGER PRIMARY KEY,
            name TEXT UNIQUE NOT NULL,
            pin_hash BLOB NOT NULL,
            created_at INTEGER NOT NULL
        );`,
        `CREATE TABLE IF NOT EXISTS words(
            id INTEGER PRIMARY KEY,
            wallet_id INTEGER NOT NULL,
            slot INTEGER NOT NULL,
            ciphertext BLOB NOT NULL,
            UNIQUE(wallet_id, slot),
            FOREIGN KEY(wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
        );`,
        `CREATE INDEX IF NOT EXISTS idx_words_wallet ON words(wallet_id);`,
    }
    tx, err := d.Begin()
    if err != nil {
        return err
    }
    defer func() {
        _ = tx.Rollback()
    }()
    for _, s := range stmts {
        if _, err := tx.Exec(s); err != nil {
            return fmt.Errorf("schema error: %w", err)
        }
    }
    if err := tx.Commit(); err != nil {
        return err
    }
    // Ensure phrase_len column exists on wallets (default 24)
    if err := ensureWalletPhraseLen(d); err != nil {
        return err
    }
    // simple sanity query
    var now int64
    if err := d.QueryRow("SELECT ?", time.Now().Unix()).Scan(&now); err != nil {
        return errors.New("database not usable after schema")
    }
    return nil
}

func ensureWalletPhraseLen(d *sql.DB) error {
    // Check if column exists
    rows, err := d.Query(`PRAGMA table_info('wallets')`)
    if err != nil { return err }
    defer rows.Close()
    has := false
    for rows.Next() {
        var cid int
        var name, ctype string
        var notnull, pk int
        var dflt sql.NullString
        if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil { return err }
        if name == "phrase_len" { has = true }
    }
    if has {
        return nil
    }
    if _, err := d.Exec(`ALTER TABLE wallets ADD COLUMN phrase_len INTEGER NOT NULL DEFAULT 24`); err != nil {
        return err
    }
    return nil
}


