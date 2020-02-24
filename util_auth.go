package minivmm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rsp9u/go-oidc"
	"golang.org/x/oauth2"
)

type userNameContextKey string

const (
	k            = userNameContextKey("userName")
	clientID     = "minivmm"
	clientSecret = "minivmmminivmm"
	CookieName   = "minivmm_token"
)

type JWTPayload struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	ClientID string `json:"client_id"`
}

// SetUserName sets a user name to http request context.
func SetUserName(r *http.Request, userName string) context.Context {
	return context.WithValue(r.Context(), k, userName)
}

// GetUserName gets a user name from http request context.
func GetUserName(r *http.Request) string {
	return r.Context().Value(k).(string)
}

// VerifyToken verifies the given OIDC access token and returns its jwt payload.
func VerifyToken(token string) (*JWTPayload, error) {
	// Check client ID and signature of access token
	payload, err := extractJWTPayload(token)
	if err != nil {
		log.Println("failed to parse access token: ", err)
		return nil, err
	}

	if payload.ClientID != clientID {
		log.Println("failed to verify celient ID in access token")
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, provider, _, err := SetupOIDCProvider(ctx)
	if err != nil {
		log.Println("failed to setup oidc provider: ", err)
		return nil, err
	}

	if _, err = provider.RemoteKeySet.VerifySignature(context.Background(), token); err != nil {
		log.Println("failed to verify signature in access token: ", err)
		return nil, err
	}

	return payload, nil
}

// SetupOIDCProvider returns OIDC configurations.
func SetupOIDCProvider(ctx context.Context) (*oauth2.Config, *oidc.Provider, *oidc.IDTokenVerifier, error) {
	redirectURL := os.Getenv(EnvOrigin) + "/api/v1/login"

	// set up
	iss := os.Getenv(EnvOIDC) + "/"
	provider, err := oidc.NewProvider(ctx, iss)
	if err != nil {
		return nil, nil, nil, errors.New("Failed to set up OIDC provider")
	}
	ep := provider.Endpoint()
	ep.AuthStyle = oauth2.AuthStyleInHeader
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     ep,
		Scopes:       []string{oidc.ScopeOpenID},
	}
	oidcConfig := oidc.Config{
		ClientID:          clientID,
		SkipClientIDCheck: false,
		SkipExpiryCheck:   false,
		SkipIssuerCheck:   false,
	}
	verifier := provider.Verifier(&oidcConfig)

	return oauth2Config, provider, verifier, nil
}

func extractJWTPayload(jwt string) (*JWTPayload, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed jwt, expected 3 parts got %d", len(parts))
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.Wrap(err, "malformed jwt payload")
	}

	var payload JWTPayload
	if err = json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload json: "+string(rawPayload))
	}

	return &payload, nil
}
