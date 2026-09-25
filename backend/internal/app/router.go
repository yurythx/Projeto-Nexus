package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	blogTransport "github.com/yurythx/projeto-aurora/internal/modules/blog/transport"
	calendarTransport "github.com/yurythx/projeto-aurora/internal/modules/calendar/transport"
	catalogTransport "github.com/yurythx/projeto-aurora/internal/modules/catalog/transport"
	contactTransport "github.com/yurythx/projeto-aurora/internal/modules/contact/transport"
	directoryTransport "github.com/yurythx/projeto-aurora/internal/modules/directory/transport"
	egressTransport "github.com/yurythx/projeto-aurora/internal/modules/egress/transport"
	exampleTransport "github.com/yurythx/projeto-aurora/internal/modules/example/transport"
	filesTransport "github.com/yurythx/projeto-aurora/internal/modules/files/transport"
	integrationsTransport "github.com/yurythx/projeto-aurora/internal/modules/integrations/transport"
	mercurioTransport "github.com/yurythx/projeto-aurora/internal/modules/mercurio/transport"
	searchTransport "github.com/yurythx/projeto-aurora/internal/modules/search/transport"
	signumTransport "github.com/yurythx/projeto-aurora/internal/modules/signum/transport"
	tramiteTransport "github.com/yurythx/projeto-aurora/internal/modules/tramite/transport"
	usersTransport "github.com/yurythx/projeto-aurora/internal/modules/users/transport"
	wikiTransport "github.com/yurythx/projeto-aurora/internal/modules/wiki/transport"
	atendimentoTransport "github.com/yurythx/projeto-aurora/internal/modules/atendimento/transport"
	organizacaoTransport "github.com/yurythx/projeto-aurora/internal/modules/organizacao/transport"

	"github.com/yurythx/projeto-aurora/internal/platform/audit"
	"github.com/yurythx/projeto-aurora/internal/platform/auth"
	"github.com/yurythx/projeto-aurora/internal/platform/branding"
	"github.com/yurythx/projeto-aurora/internal/platform/database"
	"github.com/yurythx/projeto-aurora/internal/platform/httpserver"
	"github.com/yurythx/projeto-aurora/internal/platform/idempotency"
	"github.com/yurythx/projeto-aurora/internal/platform/keycloakconfig"
	"github.com/yurythx/projeto-aurora/internal/platform/localauth"
	"github.com/yurythx/projeto-aurora/internal/platform/modules"
	"github.com/yurythx/projeto-aurora/internal/platform/outbox"
	"github.com/yurythx/projeto-aurora/internal/platform/ws"
)

// NewRouter monta o router HTTP completo para o Projeto Aurora: a base da plataforma
// (health, metrics, request id, recovery, CORS, security headers), o /ready
// conectado a toda dependência essencial, e as rotas de negócio versionadas em /api/v1 e /ws.
func NewRouter(deps *Dependencies) chi.Router {
	r := httpserver.New(httpserver.Options{
		Logger:         deps.Logger,
		AllowedOrigins: []string{deps.Config.FrontendURL},
		RequestTimeout: 30 * time.Second,
		MetricsToken:   deps.Config.Security.MetricsToken,
	})

	checks := []httpserver.Check{
		{Name: "postgres", Fn: database.Ping(deps.DB)},
		{Name: "rabbitmq", Fn: deps.Messaging.Ping},
		// MinIO — achado de auditoria: o card correspondente no painel de
		// Monitoramento sempre mostrava "Desconhecido", porque nada nunca
		// perguntava ao MinIO se ele estava de pé (só postgres/rabbitmq
		// eram checados aqui). deps.Storage.Ping reusa o client MinIO já
		// autenticado — sem ele, esta checagem só existiria se o MinIO
		// estivesse fora do ar E algo tentasse fazer upload/download na
		// hora, tarde demais para um painel de monitoramento.
		{Name: "minio", Fn: deps.Storage.Ping},
	}
	readiness := httpserver.ReadyHandler(checks, 3*time.Second)
	r.Get("/ready", readiness)
	r.Get("/readyz", readiness) // alias canônico k8s (gap G-14)

	// WebSocket autenticado por ticket. O handler interativo do Mercurio
	// (subscribe/read/open_dm) é sempre conectado — ele mesmo reavalia
	// deps.ModuleRegistry a cada frame recebido (ver newMercurioInboundHandler),
	// não mais uma decisão travada no boot: achado real, um módulo
	// desativado em runtime não podia deixar a interação por WS
	// continuando funcionando pra quem já tinha (ou abria) uma conexão. O
	// envio de mensagens em si é por HTTP (POST /mercurio/rooms/.../messages).
	wsInbound := newMercurioInboundHandler(deps.Modules.Mercurio.Service, deps.ModuleRegistry, deps.Logger)
	r.Get("/ws", ws.UpgradeHandler(deps.Hub, deps.Tickets, deps.Config.FrontendURL, wsInbound, deps.Logger))

	// Rota pública de documentação de API (OpenAPI 3.0 / Swagger UI)
	r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.json")
	})
	// Assets do Swagger UI vendorados (F2.1): swagger-ui-dist@5.17.14 +
	// swagger-initializer.js próprio — sem CDN, então a CSP de /docs pode
	// ficar em 'self' (só style-src precisa de 'unsafe-inline', porque o
	// Swagger UI injeta <style> em runtime). http.FileServer/http.Dir já
	// bloqueia path traversal (rejeita e limpa "..").
	r.Handle("/docs/assets/*", http.StripPrefix("/docs/assets/", http.FileServer(http.Dir("docs/swagger-ui"))))
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'")
		w.Write([]byte(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Documentação da API — Projeto Aurora (OpenAPI 3.0)</title>
  <link rel="stylesheet" href="/docs/assets/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="/docs/assets/swagger-ui-bundle.js"></script>
  <script src="/docs/assets/swagger-initializer.js"></script>
</body>
</html>`))
	})

	// Rota de login local (pública) — caminho ADICIONAL ao Keycloak
	// (break-glass), nunca um substituto do Identity Broker.
	localauth.RegisterRoutes(r, deps.Modules.LocalAuth.Handlers, deps.Logger, deps.RateLimiters.LocalLogin)

	r.Route("/api/v1", func(api chi.Router) {
		// --- grupo PÚBLICO — pivô Aurora v3 (docs/analise-para-aurora-v3.md) ---
		// Sem auth.RequireAuthentication: esta é a única fração de /api/v1
		// alcançável por um visitante anônimo da internet. Cada módulo
		// aqui dentro é responsável por autenticar as próprias rotas que
		// de fato exigem identidade (engajamento do Blog, CRUD
		// administrativo do Catalog) via um sub-Group interno — ver
		// catalog/transport.RegisterRoutes e blog/transport.RegisterRoutes.
		// Nunca adicione um módulo aqui sem antes confirmar que TODA rota
		// sensível dele aplica seu próprio auth.RequireAuthentication.
		//
		// Construído uma vez e passado por valor (é só uma func) aos três
		// — cada um decide sozinho ONDE dentro do próprio mount aplicá-lo,
		// nunca no nível deste Group (ver o comentário de
		// catalog/transport.RegisterRoutes sobre por que).
		requireAuth := auth.RequireAuthentication(deps.Verifier, deps.Logger)

		// GET /api/v1/system/public-features — estado de blog/catalog/contact
		// pra um visitante SEM sessão (nav do site público, PublicShell).
		// Sem ModuleRegistry.Guard de propósito: precisa responder mesmo
		// com um desses 3 módulos desativado (é o que ela reporta).
		modules.RegisterAnonymousRoutes(api, deps.Modules.SystemFeatures.Handlers)

		api.Group(func(gr chi.Router) {
			gr.Use(deps.ModuleRegistry.Guard(modules.KeyCatalog, deps.Logger))
			catalogTransport.RegisterRoutes(gr, deps.Modules.Catalog.Handlers, requireAuth, deps.RateLimiters.PublicAPI, deps.Logger)
		})
		api.Group(func(gr chi.Router) {
			gr.Use(deps.ModuleRegistry.Guard(modules.KeyBlog, deps.Logger))
			blogTransport.RegisterRoutes(gr, deps.Modules.Blog.Handlers, requireAuth, deps.RateLimiters.PublicAPI, deps.Logger)
		})
		api.Group(func(gr chi.Router) {
			gr.Use(deps.ModuleRegistry.Guard(modules.KeyContact, deps.Logger))
			contactTransport.RegisterRoutes(gr, deps.Modules.Contact.Handlers, requireAuth, deps.RateLimiters.Contact, deps.Logger)
		})
		api.Group(func(gr chi.Router) {
			gr.Use(deps.ModuleRegistry.Guard(modules.KeySignum, deps.Logger))
			signumTransport.RegisterRoutes(gr, deps.Modules.Signum.Handlers, requireAuth, deps.RateLimiters.PublicAPI, deps.Logger)
		})

		// --- grupo AUTENTICADO — tudo que exige sessão (Keycloak ou local) ---
		api.Group(func(authed chi.Router) {
			authed.Use(auth.RequireAuthentication(deps.Verifier, deps.Logger))
			// Rate limit por identidade autenticada (fallback: IP) em todo
			// este grupo — gap G-01: antes só login e /ws/ticket tinham
			// teto, qualquer token válido podia marretar as demais rotas.
			authed.Use(httpserver.RateLimit(deps.Logger, deps.RateLimiters.APIGlobal, apiRateLimitKey))
			authed.Use(idempotency.Middleware(deps.Idempotency, deps.Logger))

			authed.With(httpserver.RateLimit(deps.Logger, deps.RateLimiters.WSTicket, wsTicketRateLimitKey)).
				Post("/ws/ticket", ws.TicketHandler(deps.Tickets, deps.Logger))

			// Atendimento e Organização (Localidades, Setores, Perfis)
			atendimentoTransport.RegisterRoutes(authed, deps.Modules.Atendimento.Handlers, deps.Logger)
			organizacaoTransport.RegisterRoutes(authed, deps.Modules.Organizacao.Handlers, deps.Logger)

			// POST /api/v1/auth/logout — precisa de identidade, então mora
			// aqui dentro (o login fica fora, em localauth.RegisterRoutes).
			localauth.RegisterAuthedRoutes(authed, deps.Modules.LocalAuth.Handlers)

			// Registro central de módulos: manifesto público consumido pelo
			// FeatureFlagProvider do frontend + CRUD administrativo (aurora-admin).
			modules.RegisterPublicRoutes(authed, deps.Modules.SystemFeatures.Handlers)
			modules.RegisterAdminRoutes(authed, deps.Modules.SystemFeatures.Handlers, deps.Logger)

			// Módulos de negócio — cada bloco responde 404 em toda a sua
			// superfície quando o módulo está desativado em system_features
			// (modules.Registry.Guard avaliado por requisição, sem reiniciar
			// o processo).
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyUsers, deps.Logger))
				usersTransport.RegisterRoutes(gr, deps.Modules.Users.Handlers, deps.Logger)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyIntegrations, deps.Logger))
				integrationsTransport.RegisterRoutes(gr, deps.Modules.Integrations.Handlers, deps.Logger)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyExample, deps.Logger))
				exampleTransport.RegisterRoutes(gr, deps.Modules.Example.Handlers)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyMercurio, deps.Logger))
				mercurioTransport.RegisterRoutes(gr, deps.Modules.Mercurio.Handlers, deps.Logger)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyDirectory, deps.Logger))
				directoryTransport.RegisterRoutes(gr, deps.Modules.Directory.Handlers, deps.Logger)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyCalendar, deps.Logger))
				calendarTransport.RegisterRoutes(gr, deps.Modules.Calendar.Handlers, deps.Logger)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyFiles, deps.Logger))
				filesTransport.RegisterRoutes(gr, deps.Modules.Files.Handlers)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyWiki, deps.Logger))
				wikiTransport.RegisterRoutes(gr, deps.Modules.Wiki.Handlers)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeySearch, deps.Logger))
				searchTransport.RegisterRoutes(gr, deps.Modules.Search.Handlers)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyTramite, deps.Logger))
				tramiteTransport.RegisterRoutes(gr, deps.Modules.Tramite.Handlers, deps.Logger)
			})
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyEgress, deps.Logger))
				egressTransport.RegisterRoutes(gr, deps.Modules.Egress.Handlers, deps.Logger)
			})

			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyAudit, deps.Logger))
				audit.RegisterRoutes(gr, deps.Modules.Audit.Handlers, deps.Logger)
			})

			// keycloak_admin é módulo de núcleo (Locked, Default: true — a
			// tela de Módulos nem oferece desativá-lo pela API, ver
			// ErrLocked em modules/transport.go), mas era o ÚNICO módulo
			// do catálogo cujas rotas nunca passavam pelo Guard (achado da
			// auditoria "módulo desativado não pode ficar acessível") —
			// os outros dois módulos Locked (users, audit) já se protegem
			// mesmo sendo Locked, por defesa em profundidade contra uma
			// edição direta no banco. Fica alinhado aqui pelo mesmo motivo.
			authed.Group(func(gr chi.Router) {
				gr.Use(deps.ModuleRegistry.Guard(modules.KeyKeycloakAdmin, deps.Logger))
				keycloakconfig.RegisterRoutes(gr, deps.Modules.KeycloakConfig.Handlers, deps.Logger)
			})
			branding.RegisterPublicRoutes(authed, deps.Modules.Branding.Handlers)
			branding.RegisterAdminRoutes(authed, deps.Modules.Branding.Handlers, deps.Logger)
			outbox.RegisterStatsRoutes(authed, deps.Modules.OutboxStats.Handlers, deps.Logger)
		})
	})

	return r
}

func wsTicketRateLimitKey(r *http.Request) string {
	if identity, ok := auth.IdentityFromContext(r.Context()); ok && identity.Subject != "" {
		return identity.Subject
	}
	return httpserver.ClientIPKey(r)
}

// apiRateLimitKey identifica o chamador do rate limiter global do grupo
// AUTENTICADO de /api/v1: o subject do token autenticado quando disponível
// (o normal, já que o grupo está atrás de RequireAuthentication), caindo
// para o IP só em caminhos de borda. Mesma forma de wsTicketRateLimitKey —
// nome próprio só para documentar a intenção no ponto de uso. O grupo
// PÚBLICO usa deps.RateLimiters.PublicAPI/Contact com httpserver.ClientIPKey
// puro — nunca há identidade autenticada ali.
func apiRateLimitKey(r *http.Request) string {
	if identity, ok := auth.IdentityFromContext(r.Context()); ok && identity.Subject != "" {
		return identity.Subject
	}
	return httpserver.ClientIPKey(r)
}
