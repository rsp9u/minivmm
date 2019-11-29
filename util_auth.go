package minivmm

import (
	"context"
	"net/http"
)

type userNameContextKey string

const k = userNameContextKey("userName")

// SetUserName set user name to http request context.
func SetUserName(r *http.Request, userName string) context.Context {
	return context.WithValue(r.Context(), k, userName)
}

// GetUserName get user name from http request context.
func GetUserName(r *http.Request) string {
	return r.Context().Value(k).(string)
}
