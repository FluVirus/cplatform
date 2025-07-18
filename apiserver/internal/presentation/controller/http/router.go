package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct {
	controller *Controller
}

func NewRouter(controller *Controller) *Router {
	return &Router{
		controller: controller,
	}
}

func (r *Router) GetHandler() http.Handler {
	m := mux.NewRouter()

	v1 := m.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/users", r.controller.RegisterUserHandler).Methods(http.MethodPost)
	v1.HandleFunc("/users", r.controller.DeleteSelfUserHandler).Methods(http.MethodDelete)

	return m
}
