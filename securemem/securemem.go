package securemem

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// MemoryVault securely stores encrypted values in memory and supports persistence.
type MemoryVault struct {
	key  []byte
	db   map[string][]byte
	lock sync.RWMutex
	gcm  cipher.AEAD
}

// NewMemoryVault creates a new vault with a random key.
func NewMemoryVault() (*MemoryVault, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	return NewMemoryVaultWithKey(key)
}

// NewMemoryVaultWithKey creates a vault with a provided AES-256 key.
func NewMemoryVaultWithKey(key []byte) (*MemoryVault, error) {
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

// PersistToFile saves encrypted in-memory data to disk.
func (v *MemoryVault) PersistToFile(path string) error {
	v.lock.RLock()
	defer v.lock.RUnlock()

	blob, err := json.Marshal(v.db)
	if err != nil {
		return fmt.Errorf("marshal vault map failed: %w", err)
	}

	// Encrypt entire blob
	nonce := make([]byte, v.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce failed: %w", err)
	}
	ciphertext := v.gcm.Seal(nonce, nonce, blob, nil)

	return os.WriteFile(path, ciphertext, 0600)
}

// LoadFromFile loads encrypted data from disk and restores the vault.
func (v *MemoryVault) LoadFromFile(path string) error {
	v.lock.Lock()
	defer v.lock.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	nonceSize := v.gcm.NonceSize()
	if len(data) < nonceSize {
		return errors.New("invalid file data")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]
	plaintext, err := v.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return json.Unmarshal(plaintext, &v.db)
}
