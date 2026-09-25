package transport

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registra todas as rotas do módulo de Atendimento Socioassistencial na API protegida.
func RegisterRoutes(r chi.Router, h *Handlers, logger *slog.Logger) {
	r.Route("/atendimentos", func(api chi.Router) {
		api.Get("/", h.ListAtendimentos)
		api.Post("/", h.CreateAtendimento)
		api.Get("/stats", h.GetStats)
		api.Get("/busca-cidadao/{cpf}", h.SearchByCPF)
		api.Get("/{id}", h.GetAtendimento)
		api.Put("/{id}", h.UpdateAtendimento)
	})

	r.Get("/servicos/meus-servicos", h.ResolveMyServices)

	r.Route("/centro-pop/prontuarios", func(api chi.Router) {
		api.Get("/", h.ListProntuarios)
		api.Post("/", h.CreateProntuario)
		api.Get("/{id}", h.GetProntuario)
	})
}
