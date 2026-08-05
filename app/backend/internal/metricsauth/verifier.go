package metricsauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type Verifier struct {
	digest [sha256.Size]byte
}

func Load(tokenFile string) (*Verifier, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("read metrics bearer token file %q: %w", tokenFile, err)
	}

	token := string(data)
	switch {
	case strings.HasSuffix(token, "\r\n"):
		token = strings.TrimSuffix(token, "\r\n")
	case strings.HasSuffix(token, "\n"):
		token = strings.TrimSuffix(token, "\n")
	}

	verifier, err := New(token)
	if err != nil {
		return nil, fmt.Errorf("validate metrics bearer token file %q: %w", tokenFile, err)
	}
	return verifier, nil
}

func New(token string) (*Verifier, error) {
	if token == "" {
		return nil, fmt.Errorf("metrics bearer token must not be empty")
	}
	if strings.IndexFunc(token, unicode.IsSpace) >= 0 {
		return nil, fmt.Errorf("metrics bearer token must be a single value without whitespace")
	}

	return &Verifier{digest: sha256.Sum256([]byte(token))}, nil
}

func (v *Verifier) Authorize(header string) bool {
	if v == nil {
		return false
	}

	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" || strings.IndexFunc(token, unicode.IsSpace) >= 0 {
		return false
	}

	candidate := sha256.Sum256([]byte(token))
	return subtle.ConstantTimeCompare(candidate[:], v.digest[:]) == 1
}
