package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rsp9u/go-oidc"
	"minivmm"
)

// AuthMiddleware is a middleware resolving authentication.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, r := auth(r)
		if isAuth || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
		} else {
			redirectToAuthURL(w, r)
		}
	})
}

func auth(r *http.Request) (bool, *http.Request) {
	if minivmm.C.NoAuth {
		newCtx := minivmm.SetUserName(r, "dummy.user")
		newReq := r.WithContext(newCtx)
		return true, newReq
	}

	cookie, err := r.Cookie(minivmm.CookieName)
	if err != nil {
		return false, r
	}
	token := cookie.Value

	payload, err := minivmm.VerifyToken(token)
	if err != nil {
		return false, r
	}

	// Set user name into context
	newCtx := minivmm.SetUserName(r, payload.Subject)
	newReq := r.WithContext(newCtx)

	return true, newReq
}

func redirectToAuthURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oauth2Config, _, _, err := minivmm.SetupOIDCProvider(ctx)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}
	state := "dummy-state"
	http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusFound)
}

// HandleAuth redirects to the main page if authentication is successful.
func HandleAuth(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, minivmm.C.Origin, 302)
}

// HandleOIDCCallback obtains and verifies OIDC tokens, and set the access token to client.
func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	// NOTE: skip state verification because this service does not store the private resources.

	_, accessToken, err := generateToken(r.URL.Query().Get("code"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "Unauthorized.\n"+err.Error()+"\n")
		return
	}

	cookie := http.Cookie{
		Name:   minivmm.CookieName,
		Value:  accessToken,
		Path:   "/",
		Secure: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, minivmm.C.Origin, 302)
}

func generateToken(authCode string) (*oidc.IDToken, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oauth2Config, _, verifier, err := minivmm.SetupOIDCProvider(ctx)
	if err != nil {
		return nil, "", err
	}

	// Exchange code and token
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oauth2Token, err := oauth2Config.Exchange(ctx, authCode)
	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to fetch token from OIDC provider")
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
