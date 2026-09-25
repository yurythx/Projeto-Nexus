package transport

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handlers, logger *slog.Logger) {
	r.Route("/organizacao", func(api chi.Router) {
		// Acesso pessoal
		api.Get("/meu-acesso", h.ResolveMeuAcesso)

		// Localidades
		api.Get("/localidades", h.ListLocalidades)
		api.Post("/localidades", h.CreateLocalidade)
		api.Put("/localidades/{id}", h.UpdateLocalidade)
		api.Delete("/localidades/{id}", h.DeleteLocalidade)

		// Setores por Localidade
		api.Get("/localidades/{id}/setores", h.ListSetores)
		api.Post("/localidades/{id}/setores", h.CreateSetor)
		api.Put("/setores/{id}", h.UpdateSetor)
		api.Delete("/setores/{id}", h.DeleteSetor)

		// Perfis
		api.Get("/perfis", h.ListPerfis)
		api.Post("/perfis", h.CreatePerfil)
		api.Put("/perfis/{id}", h.UpdatePerfil)

		// Lotação de Usuário
		api.Get("/usuarios/{id}/lotacao", h.GetUsuarioLotacao)
		api.Put("/usuarios/{id}/lotacao", h.SetUsuarioLotacao)
	})
}
