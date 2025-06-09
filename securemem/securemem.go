package securemem

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// MemoryVault securely stores encrypted values in memory.
type MemoryVault struct {
	key  []byte
	db   map[string][]byte
	lock sync.RWMutex
	gcm  cipher.AEAD
}

// NewMemoryVault creates a new vault and generates a random AES-GCM key.
func NewMemoryVault() (*MemoryVault, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &MemoryVault{
		key: key,
		db:  make(map[string][]byte),
		gcm: gcm,
	}, nil
}

// Put stores an encrypted version of the struct under the given key.
func (v *MemoryVault) Put(key string, val interface{}) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	plaintext, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	nonce := make([]byte, v.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := v.gcm.Seal(nonce, nonce, plaintext, nil)
	v.db[key] = ciphertext
	return nil
}

// Get retrieves and decrypts the value associated with key into dest.
func (v *MemoryVault) Get(key string, dest interface{}) error {
	v.lock.RLock()
	defer v.lock.RUnlock()

	ciphertext, ok := v.db[key]
	if !ok {
		return errors.New("key not found")
	}

	nonceSize := v.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return errors.New("invalid ciphertext")
	}

	nonce := ciphertext[:nonceSize]
	enc := ciphertext[nonceSize:]

	plaintext, err := v.gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return json.Unmarshal(plaintext, dest)
}
