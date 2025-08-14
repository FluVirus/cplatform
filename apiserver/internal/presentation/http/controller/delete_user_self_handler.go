package controller

import (
	"cplatform/internal/application/authentication/basic"
	"net/http"
)

func (c *Controller) DeleteSelfUserHandler(w http.ResponseWriter, r *http.Request) {
	panic("ABOBA")

	user := basic.GetUser(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(user.Email))

	// TODO: complete implementation
}
