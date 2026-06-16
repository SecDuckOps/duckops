package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type AuthManager interface {
	Load() error
	AgentID() string
	Token() string
	SetAgentID(id string)
	SetToken(token string)
	Save() error
	Clear() error
}

type Identity struct {
	AgentID string `json:"agent_id"`
	Token   string `json:"token"`
}

type Manager struct {
	identity Identity
	key      []byte
	path     string
}

func NewManager(dataDir string) *Manager {
	key := deriveKey(dataDir)
	return &Manager{
		path: filepath.Join(dataDir, "identity.json"),
		key:  key[:],
	}
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read identity: %w", err)
	}

	decrypted, err := decrypt(data, m.key)
	if err != nil {
		return fmt.Errorf("decrypt identity: %w", err)
	}

	if err := json.Unmarshal(decrypted, &m.identity); err != nil {
		return fmt.Errorf("parse identity: %w", err)
	}

	slog.Debug("loaded agent identity", "agent_id", m.identity.AgentID)
	return nil
}

func (m *Manager) AgentID() string {
	return m.identity.AgentID
}

func (m *Manager) Token() string {
	return m.identity.Token
}

func (m *Manager) SetAgentID(id string) {
	m.identity.AgentID = id
}

func (m *Manager) SetToken(token string) {
	m.identity.Token = token
}

func (m *Manager) Save() error {
	plaintext, err := json.Marshal(m.identity)
	if err != nil {
		return fmt.Errorf("marshal identity: %w", err)
	}

	encrypted, err := encrypt(plaintext, m.key)
	if err != nil {
		return fmt.Errorf("encrypt identity: %w", err)
	}

	if err := os.WriteFile(m.path, encrypted, 0o600); err != nil {
		return fmt.Errorf("write identity: %w", err)
	}

	slog.Debug("saved agent identity", "agent_id", m.identity.AgentID)
	return nil
}

func (m *Manager) Clear() error {
	m.identity = Identity{}
	if err := os.Remove(m.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func deriveKey(seed string) [32]byte {
	return sha256.Sum256([]byte("duckops-agent-v1:" + seed))
}

func encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func decrypt(data, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
