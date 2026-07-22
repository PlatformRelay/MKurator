package mqrest

import (
	"context"
	"net/http"
)

// requestAuthenticator applies the authentication state for one mqweb request.
// Implementations may use the request context to acquire or refresh state.
type requestAuthenticator interface {
	authenticate(context.Context, *http.Request) error
}

type basicRequestAuthenticator struct {
	username string
	password string
}

func (a basicRequestAuthenticator) authenticate(_ context.Context, req *http.Request) error {
	req.SetBasicAuth(a.username, a.password)
	return nil
}
