package crypto

import (
    "crypto/sha256"
    "encoding/binary"
)

// SlotCipher is a stored word ciphertext mapped to its permuted slot.
type SlotCipher struct { Slot int; Data []byte }

// DeriveKey derives a 32-byte key from PIN and walletID.
func DeriveKey(pin string, walletID int64) [32]byte {
    var le [8]byte
    binary.LittleEndian.PutUint64(le[:], uint64(walletID))
    return sha256.Sum256(append([]byte(pin), le[:]...))
}

// xorshift128+ PRNG
type xorshift128plus struct {
    s0 uint64
    s1 uint64
}

func newXorShiftFromSeed(seed []byte) *xorshift128plus {
    // seed expects at least 16 bytes; if shorter, pad via hashing
    var s0, s1 [8]byte
    if len(seed) >= 16 {
        copy(s0[:], seed[:8])
        copy(s1[:], seed[8:16])
    } else {
        h := sha256.Sum256(seed)
        copy(s0[:], h[:8])
        copy(s1[:], h[8:16])
    }
    return &xorshift128plus{ s0: le64(s0[:]), s1: le64(s1[:]) }
}

func (x *xorshift128plus) next() uint64 {
    s1 := x.s0
    s0 := x.s1
    x.s0 = s0
    s1 ^= s1 << 23
    s1 ^= s1 >> 17
    s1 ^= s0
    s1 ^= s0 >> 26
    x.s1 = s1
    return x.s0 + x.s1
}

func le64(b []byte) uint64 {
    var v [8]byte
    copy(v[:], b)
    return binary.LittleEndian.Uint64(v[:])
}

// BuildPermutation returns a deterministic permutation of 0..23 from the key.
func BuildPermutationN(key [32]byte, n int) []int {
    perm := make([]int, n)
    for i := 0; i < n; i++ {
        perm[i] = i
    }
    prng := newXorShiftFromSeed(key[:16])
    for i := n-1; i > 0; i-- {
        j := int(prng.next() % uint64(i+1))
        perm[i], perm[j] = perm[j], perm[i]
    }
    return perm
}

func BuildPermutation(key [32]byte) []int { return BuildPermutationN(key, 24) }

func invertPermutation(perm []int) []int {
    inv := make([]int, len(perm))
    for i, p := range perm {
        inv[p] = i
    }
    return inv
}

func nonceForIndex(key [32]byte, i int) []byte {
    h := sha256.New()
    h.Write(key[:])
    h.Write([]byte{byte(i)})
    sum := h.Sum(nil)
    return sum[:16]
}

func maskBytes(seed []byte, n int) []byte {
    prng := newXorShiftFromSeed(seed)
    out := make([]byte, n)
    off := 0
    for off < n {
        v := prng.next()
        var buf [8]byte
        binary.LittleEndian.PutUint64(buf[:], v)
        toCopy := 8
        if n-off < 8 { toCopy = n-off }
        copy(out[off:off+toCopy], buf[:toCopy])
        off += toCopy
    }
    return out
}

// EncryptWords returns slot->ciphertext pairs using permutation and XOR mask.
func EncryptWords(words []string, pin string, walletID int64) []SlotCipher {
    key := DeriveKey(pin, walletID)
    perm := BuildPermutationN(key, len(words))
    out := make([]SlotCipher, len(words))
    for i, w := range words {
        seed := nonceForIndex(key, i)
        mask := maskBytes(seed, len(w))
        b := []byte(w)
        for k := range b {
            b[k] ^= mask[k]
        }
        out[i] = SlotCipher{ Slot: perm[i], Data: append([]byte(nil), b...) }
    }
    return out
}

// DecryptWords reconstructs ordered words from stored slot-cipher rows.
func DecryptWords(rows []SlotCipher, pin string, walletID int64) []string {
    key := DeriveKey(pin, walletID)
    // infer n from rows length (max slot +1)
    n := 0
    for _, r := range rows { if r.Slot+1 > n { n = r.Slot+1 } }
    if n == 0 { n = 24 }
    perm := BuildPermutationN(key, n)
    inv := invertPermutation(perm)
    ordered := make([]string, n)
    for _, r := range rows {
        i := inv[r.Slot]
        seed := nonceForIndex(key, i)
        mask := maskBytes(seed, len(r.Data))
        b := append([]byte(nil), r.Data...)
        for k := range b {
            b[k] ^= mask[k]
        }
        ordered[i] = string(b)
    }
    return ordered
}


