package v0

import (
	xena "github.com/e-zhydzetski/tt-xena"
)

func NewHandler() *Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(job xena.Job) error {
	return nil
}
