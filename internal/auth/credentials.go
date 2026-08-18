package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"
)

const (
	algorithm      = "pbkdf2-sha256"
	iterations     = 200000
	saltLen        = 16
	keyLen         = 32
	MinPasswordLen = 8
)

const passwordAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// Credentials is the stored password verifier. The plaintext is never written.
type Credentials struct {
	Algorithm  string    `json:"algorithm"`
	Iterations int       `json:"iterations"`
	Salt       string    `json:"salt"`
	Hash       string    `json:"hash"`
	UpdatedAt  time.Time `json:"updated_at"`

	path string
}

// GeneratePassword returns a random password in xxxx-xxxx-xxxx form, drawn from
// an alphabet without visually ambiguous characters so it can be read off a
// screen and typed on a phone.
func GeneratePassword() string {
	groups := make([]byte, 0, 14)
	for i := 0; i < 12; i++ {
		if i > 0 && i%4 == 0 {
			groups = append(groups, '-')
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(passwordAlphabet))))
		if err != nil {
			return ""
		}
		groups = append(groups, passwordAlphabet[n.Int64()])
	}
	return string(groups)
}

func derive(password string, salt []byte, iters int) ([]byte, error) {
	return pbkdf2.Key(sha256.New, password, salt, iters, keyLen)
}

func newCredentials(path, password string) (*Credentials, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	hash, err := derive(password, salt, iterations)
	if err != nil {
		return nil, err
	}

	return &Credentials{
		Algorithm:  algorithm,
		Iterations: iterations,
		Salt:       base64.StdEncoding.EncodeToString(salt),
		Hash:       base64.StdEncoding.EncodeToString(hash),
		UpdatedAt:  time.Now().UTC(),
		path:       path,
	}, nil
}

// LoadCredentials reads an existing verifier from disk.
func LoadCredentials(path string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if c.Algorithm != algorithm {
		return nil, fmt.Errorf("unsupported password algorithm %q in %s", c.Algorithm, path)
	}
	c.path = path
	return &c, nil
}

// LoadOrCreateCredentials loads the verifier at path, creating one on first run.
// envPassword seeds it when set; otherwise a password is generated and returned
// as plaintext so the caller can show it once.
func LoadOrCreateCredentials(path, envPassword string) (creds *Credentials, plaintext string, err error) {
	if c, err := LoadCredentials(path); err == nil {
		return c, "", nil
	} else if !os.IsNotExist(err) {
		return nil, "", err
	}

	password := envPassword
	generated := ""
	if password == "" {
		password = GeneratePassword()
		generated = password
	}
	if len(password) < MinPasswordLen {
		return nil, "", fmt.Errorf("password must be at least %d characters", MinPasswordLen)
	}

	c, err := newCredentials(path, password)
	if err != nil {
		return nil, "", err
	}
	if err := c.save(); err != nil {
		return nil, "", err
	}
	return c, generated, nil
}

func (c *Credentials) save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, append(data, '\n'), 0o600)
}

// Verify reports whether password matches, comparing in constant time.
func (c *Credentials) Verify(password string) bool {
	salt, err := base64.StdEncoding.DecodeString(c.Salt)
	if err != nil {
		return false
	}
	want, err := base64.StdEncoding.DecodeString(c.Hash)
	if err != nil {
		return false
	}

	iters := c.Iterations
	if iters <= 0 {
		iters = iterations
	}

	got, err := derive(password, salt, iters)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}

// Set replaces the password. Existing sessions stay valid; revoke them
// separately if that is not what you want.
func (c *Credentials) Set(password string) error {
	if len(password) < MinPasswordLen {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLen)
	}

	next, err := newCredentials(c.path, password)
	if err != nil {
		return err
	}

	c.Algorithm = next.Algorithm
	c.Iterations = next.Iterations
	c.Salt = next.Salt
	c.Hash = next.Hash
	c.UpdatedAt = next.UpdatedAt
	return c.save()
}
