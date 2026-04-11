package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWT struct {
	jwksURL   string
	publicKeys map[string]*ecdsa.PublicKey
}

type GoogleClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
	jwt.RegisteredClaims
}


type GoogleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Scope         string `json:"scope"`
}

func NewJWT(jwksURL string) (*JWT, error) {
	j := &JWT{
		jwksURL:   jwksURL,
		publicKeys: make(map[string]*ecdsa.PublicKey),
	}
	if err := j.refreshKeys(); err != nil {
		return nil, err
	}
	return j, nil
}

// ================= JWKS =================

type jwksResponse struct {
	Keys []struct {
		Kid string `json:"kid"`
		Kty string `json:"kty"`
		Crv string `json:"crv"`
		X   string `json:"x"`
		Y   string `json:"y"`
	} `json:"keys"`
}

func (j *JWT) refreshKeys() error {
	resp, err := http.Get(j.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	for _, key := range jwks.Keys {
		if key.Kty != "EC" || key.Crv != "P-256" {
			continue
		}

		xBytes, _ := base64.RawURLEncoding.DecodeString(key.X)
		yBytes, _ := base64.RawURLEncoding.DecodeString(key.Y)

		pubKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}

		j.publicKeys[key.Kid] = pubKey
	}

	return nil
}

// ================= Validate Google Access Token =================

func validateGoogleAccessToken(tokenString string) (*GoogleTokenInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?access_token=%s", tokenString))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid access token")
	}

	var tokenInfo GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}

	return &tokenInfo, nil
}

// ================= JWT VALIDATION =================

func (j *JWT) ValidateToken(tokenString string) (interface{}, error) {
	// Cek apakah token adalah Google Access Token (opaque ya29.)
	if strings.HasPrefix(tokenString, "ya29.") {
		return validateGoogleAccessToken(tokenString)
	}

	// Cek apakah token adalah JWT (Google ID Token atau Supabase Token)
	if strings.HasPrefix(tokenString, "ey.") {
		// Try Google ID Token first
		claims := &GoogleClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != jwt.SigningMethodES256.Alg() {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}

				kid, ok := t.Header["kid"].(string)
				if !ok {
					return nil, errors.New("missing kid in token header")
				}

				key, ok := j.publicKeys[kid]
				if !ok {
					// auto refresh kalau key belum ada (rotated key)
					if err := j.refreshKeys(); err != nil {
						return nil, err
					}
					key, ok = j.publicKeys[kid]
					if !ok {
						return nil, errors.New("public key not found for kid")
					}
				}

				return key, nil
			},
			jwt.WithLeeway(1*time.Minute),
		)
		if err == nil && token.Valid {
			return claims, nil
		}

	}

	return nil, errors.New("invalid token")
}

// ================= HELPER =================

func (j *JWT) GetUserID(tokenString string) (uuid.UUID, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	switch c := claims.(type) {
	case *GoogleClaims:
		// Buat UUID yang konsisten dari sub Google
		return uuid.NewSHA1(uuid.Nil, []byte(c.Sub)), nil
	case *GoogleTokenInfo:
		// Buat UUID yang konsisten dari sub Google
		return uuid.NewSHA1(uuid.Nil, []byte(c.Sub)), nil
	default:
		return uuid.Nil, errors.New("unsupported claims type")
	}
}
