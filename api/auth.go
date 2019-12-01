package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rsp9u/go-oidc"
	"golang.org/x/oauth2"
	"minivmm"
)

const (
	clientID     = "minivmm"
	clientSecret = "minivmmminivmm"
)

type jwtPayload struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	ClientID string `json:"client_id"`
}

// AuthMiddleware is a middleware resolving authentication.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, r := auth(r)
		if isAuth || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			ret := map[string]string{"error": "Unauthorized", "oidc_url": os.Getenv(minivmm.EnvOIDC)}
			b, _ := json.Marshal(ret)
			w.Write(b)
		}
	})
}

func auth(r *http.Request) (bool, *http.Request) {
	envNoAuth := os.Getenv(minivmm.EnvNoAuth)
	if envNoAuth == "1" || envNoAuth == "true" {
		newCtx := minivmm.SetUserName(r, "dummy.user")
		newReq := r.WithContext(newCtx)
		return true, newReq
	}

	a := r.Header.Get("Authorization")
	if a == "" {
		return false, r
	}

	s := strings.Split(a, " ")
	token := s[1]

	// Check client ID and signature of access token
	payload, err := extractJWTPayload(token)
	if err != nil {
		log.Println("failed to parse access token: ", err)
		return false, r
	}

	if payload.ClientID != clientID {
		log.Println("failed to verify celient ID in access token")
		return false, r
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, provider, _, err := setupOIDCProvider(ctx)
	if err != nil {
		log.Println("failed to setup oidc provider: ", err)
		return false, r
	}

	if _, err = provider.RemoteKeySet.VerifySignature(context.Background(), token); err != nil {
		log.Println("failed to verify signature in access token: ", err)
		return false, r
	}

	// Set user name into context
	newCtx := minivmm.SetUserName(r, payload.Subject)
	newReq := r.WithContext(newCtx)

	return true, newReq
}

// HandleOIDCCallback obtains and verifies OIDC tokens, and set the access token to client.
func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	_, accessToken, err := generateToken(r.URL.Query().Get("code"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized.\n"+err.Error()+"\n")
		return
	}

	cookie := http.Cookie{
		Name:   "minivmm_token",
		Value:  accessToken,
		Path:   "/",
		Secure: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, os.Getenv(minivmm.EnvOrigin), 302)
}

func generateToken(authCode string) (*oidc.IDToken, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oauth2Config, _, verifier, err := setupOIDCProvider(ctx)
	if err != nil {
		return nil, "", err
	}

	// Exchange code and token
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oauth2Token, err := oauth2Config.Exchange(ctx, authCode)
	if err != nil {
		return nil, "", errors.New("Failed to fetch token from OIDC provider")
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, "", errors.New("Missing ID token")
	}

	// Parse and verify ID Token payload.
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to verify ID token")
	}

	// Extract the access Token from OAuth2 token.
	rawAccessToken, ok := oauth2Token.Extra("access_token").(string)
	if !ok {
		return nil, "", errors.New("Missing access token")
	}

	return idToken, rawAccessToken, nil
}

func setupOIDCProvider(ctx context.Context) (*oauth2.Config, *oidc.Provider, *oidc.IDTokenVerifier, error) {
	redirectURL := os.Getenv(minivmm.EnvOrigin) + "/api/v1/login"

	// set up
	iss := os.Getenv(minivmm.EnvOIDC) + "/"
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

func extractJWTPayload(jwt string) (*jwtPayload, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed jwt, expected 3 parts got %d", len(parts))
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.Wrap(err, "malformed jwt payload")
	}

	var payload jwtPayload
	if err = json.Unmarshal(rawPayload, &payload); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal payload json: "+string(rawPayload))
	}

	return &payload, nil
}
