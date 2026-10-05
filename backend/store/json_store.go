package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"web-ssh/models"
)

const (
	encryptionKey = "web-ssh-encryption-key-32bytes!!"
)

type JSONStore struct {
	mu       sync.RWMutex
	filePath string
	data     *StoreData
}

type StoreData struct {
	Servers map[string]*models.Server `json:"servers"`
	Users   map[string]*models.User   `json:"users"`
}

func NewJSONStore(dir string) (*JSONStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	filePath := filepath.Join(dir, "data.json")
	store := &JSONStore{
		filePath: filePath,
		data: &StoreData{
			Servers: make(map[string]*models.Server),
			Users:   make(map[string]*models.User),
		},
	}

	if _, err := os.Stat(filePath); err == nil {
		if err := store.load(); err != nil {
			return nil, fmt.Errorf("failed to load store: %w", err)
		}
	} else {

		store.data.Users["admin"] = &models.User{
			ID:       "admin",
			Username: "admin",
			Password: "admin123",
		}
		if err := store.save(); err != nil {
			return nil, fmt.Errorf("failed to initialize store: %w", err)
		}
	}

	return store, nil
}

func (s *JSONStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	decrypted, err := decrypt(data)
	if err != nil {

		plain := &StoreData{
			Servers: make(map[string]*models.Server),
			Users:   make(map[string]*models.User),
		}
		if err := json.Unmarshal(data, plain); err != nil {
			return fmt.Errorf("failed to parse store data: %w", err)
		}
		s.data = plain
		return nil
	}

	plain := &StoreData{
		Servers: make(map[string]*models.Server),
		Users:   make(map[string]*models.User),
	}
	if err := json.Unmarshal(decrypted, plain); err != nil {
		return fmt.Errorf("failed to parse decrypted store data: %w", err)
	}
	s.data = plain
	return nil
}

func (s *JSONStore) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	encrypted, err := encrypt(data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, encrypted, 0600)
}

func (s *JSONStore) GetServers(userID string) []*models.Server {
	s.mu.RLock()
	defer s.mu.RUnlock()

	servers := make([]*models.Server, 0, len(s.data.Servers))
	for _, server := range s.data.Servers {
		servers = append(servers, server)
	}
	return servers
}

func (s *JSONStore) GetServer(id, userID string) (*models.Server, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, ok := s.data.Servers[id]
	if !ok {
		return nil, false
	}
	return server, true
}

func (s *JSONStore) CreateServer(server *models.Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	server.CreatedAt = now
	server.UpdatedAt = now

	if server.Password != "" {
		enc, err := encryptString(server.Password)
		if err == nil {
			server.Password = enc
		}
	}

	s.data.Servers[server.ID] = server
	return s.save()
}

func (s *JSONStore) UpdateServer(server *models.Server) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data.Servers[server.ID]; !ok {
		return fmt.Errorf("server not found")
	}

	server.UpdatedAt = time.Now()

	if server.Password != "" {
		existing := s.data.Servers[server.ID]
		decrypted, err := decryptString(existing.Password)
		if err != nil || decrypted != server.Password {
			enc, err := encryptString(server.Password)
			if err == nil {
				server.Password = enc
			}
		} else {
			server.Password = existing.Password
		}
	}

	s.data.Servers[server.ID] = server
	return s.save()
}

func (s *JSONStore) DeleteServer(id, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data.Servers, id)
	return s.save()
}

func (s *JSONStore) DecryptServerPassword(serverID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, ok := s.data.Servers[serverID]
	if !ok {
		return "", fmt.Errorf("server not found")
	}

	if server.Password == "" {
		return "", nil
	}

	return decryptString(server.Password)
}

func (s *JSONStore) GetUser(username string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.data.Users[username]
	if !ok {
		return nil, false
	}
	return user, true
}

func (s *JSONStore) ValidateUser(username, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.data.Users[username]
	if !ok {
		return false
	}
	return user.Password == password
}

func encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func encryptString(s string) (string, error) {
	encrypted, err := encrypt([]byte(s))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", encrypted), nil
}

func decryptString(s string) (string, error) {
	var ciphertext []byte
	_, err := fmt.Sscanf(s, "%x", &ciphertext)
	if err != nil {
		return s, nil
	}

	decrypted, err := decrypt(ciphertext)
	if err != nil {
		return s, nil
	}
	return string(decrypted), nil
}

func (s *JSONStore) GetTerminalSettings(userID, serverID string) (*models.TerminalSettings, error) {
	return nil, fmt.Errorf("not implemented: use MySQL store")
}

func (s *JSONStore) SaveTerminalSettings(settings *models.TerminalSettings) error {
	return fmt.Errorf("not implemented: use MySQL store")
}

func (s *JSONStore) CreateUser(id, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Users[id] = &models.User{
		ID:       id,
		Username: id,
		Password: password,
	}
	return s.save()
}

func (s *JSONStore) UserExists(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data.Users[userID]
	return ok
}

func (s *JSONStore) GetServerByID(id string) (*models.Server, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	server, ok := s.data.Servers[id]
	if !ok {
		return nil, false
	}
	return server, true
}

func (s *JSONStore) GetUserGroups(userID string) map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int)
	for _, server := range s.data.Servers {
		g := "默认"
		if server.Group != "" {
			g = server.Group
		}
		result[g]++
	}
	return result
}
