package http

import (
	"cplatform/internal/application/authentication/basic"
	"net/http"
)

func (c *Controller) DeleteSelfUserHandler(w http.ResponseWriter, r *http.Request) {
	user := basic.GetUser(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: complete implementation
}
