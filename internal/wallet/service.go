package wallet

import (
    "crypto/sha256"
    "database/sql"
    "encoding/binary"
    "errors"
    "fmt"
    "strings"
    "time"

    crypt "keykeeper/internal/crypto"
)

type Wallet struct {
    ID        int64
    Name      string
    PhraseLen int
    CreatedAt time.Time
}

func le64(i int64) []byte {
    var b [8]byte
    binary.LittleEndian.PutUint64(b[:], uint64(i))
    return b[:]
}

func pinHash(pin string, walletID int64) []byte {
    h := sha256.Sum256(append([]byte(pin), le64(walletID)...))
    return h[:]
}

func Create(db *sql.DB, name, pin string, words []string) (int64, error) {
    n := len(words)
    if n != 12 && n != 18 && n != 24 {
        return 0, errors.New("phrase length must be 12, 18, or 24")
    }
    for _, w := range words {
        if strings.TrimSpace(w) == "" {
            return 0, errors.New("words must be non-empty")
        }
    }
    tx, err := db.Begin()
    if err != nil { return 0, err }
    defer func(){ _ = tx.Rollback() }()

    res, err := tx.Exec(`INSERT INTO wallets(name, pin_hash, created_at, phrase_len) VALUES(?,?,?,?)`, name, []byte{0}, time.Now().Unix(), n)
    if err != nil {
        return 0, fmt.Errorf("create wallet: %w", err)
    }
    id, err := res.LastInsertId()
    if err != nil { return 0, err }

    // update pin_hash with salted id
    if _, err := tx.Exec(`UPDATE wallets SET pin_hash=? WHERE id=?`, pinHash(pin, id), id); err != nil {
        return 0, err
    }

    pairs := crypt.EncryptWords(words, pin, id)
    for _, p := range pairs {
        if _, err := tx.Exec(`INSERT INTO words(wallet_id, slot, ciphertext) VALUES(?,?,?)`, id, p.Slot, p.Data); err != nil {
            return 0, err
        }
    }
    if err := tx.Commit(); err != nil { return 0, err }
    return id, nil
}

func List(db *sql.DB) ([]Wallet, error) {
    rows, err := db.Query(`SELECT id, name, created_at, phrase_len FROM wallets ORDER BY created_at DESC`)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []Wallet
    for rows.Next() {
        var w Wallet
        var ts int64
        if err := rows.Scan(&w.ID, &w.Name, &ts, &w.PhraseLen); err != nil { return nil, err }
        w.CreatedAt = time.Unix(ts, 0)
        out = append(out, w)
    }
    return out, rows.Err()
}

func VerifyPIN(db *sql.DB, walletID int64, pin string) (bool, error) {
    var h []byte
    if err := db.QueryRow(`SELECT pin_hash FROM wallets WHERE id=?`, walletID).Scan(&h); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return false, errors.New("wallet not found") }
        return false, err
    }
    expected := pinHash(pin, walletID)
    return string(h) == string(expected), nil
}

func GetOrderedWords(db *sql.DB, walletID int64, pin string) ([]string, error) {
    ok, err := VerifyPIN(db, walletID, pin)
    if err != nil { return nil, err }
    if !ok { return nil, errors.New("invalid PIN") }

    rows, err := db.Query(`SELECT slot, ciphertext FROM words WHERE wallet_id=?`, walletID)
    if err != nil { return nil, err }
    defer rows.Close()
    var rslot int
    var rdata []byte
    var tmp []crypt.SlotCipher
    for rows.Next() {
        if err := rows.Scan(&rslot, &rdata); err != nil { return nil, err }
        tmp = append(tmp, crypt.SlotCipher{Slot: rslot, Data: append([]byte(nil), rdata...)})
    }
    ordered := crypt.DecryptWords(tmp, pin, walletID)
    return ordered, rows.Err()
}

func Delete(db *sql.DB, walletID int64) error {
    _, err := db.Exec(`DELETE FROM wallets WHERE id=?`, walletID)
    return err
}

func DeleteWordAt(db *sql.DB, walletID int64, pin string, index1based int) error {
    ok, err := VerifyPIN(db, walletID, pin)
    if err != nil { return err }
    if !ok { return errors.New("invalid PIN") }
    // get phrase length for validation
    var n int
    if err := db.QueryRow(`SELECT phrase_len FROM wallets WHERE id=?`, walletID).Scan(&n); err != nil { return err }
    if index1based < 1 || index1based > n { return errors.New("index out of range") }
    key := crypt.DeriveKey(pin, walletID)
    perm := crypt.BuildPermutationN(key, n)
    slot := perm[index1based-1]
    _, err = db.Exec(`DELETE FROM words WHERE wallet_id=? AND slot=?`, walletID, slot)
    return err
}


