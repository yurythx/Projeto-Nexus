package app

import (
	"context"

	exampleApp "github.com/yurythx/projeto-aurora/internal/modules/example/application"
	exampleInfra "github.com/yurythx/projeto-aurora/internal/modules/example/infrastructure"
	exampleTransport "github.com/yurythx/projeto-aurora/internal/modules/example/transport"

	blogApp "github.com/yurythx/projeto-aurora/internal/modules/blog/application"
	blogInfra "github.com/yurythx/projeto-aurora/internal/modules/blog/infrastructure"
	blogTransport "github.com/yurythx/projeto-aurora/internal/modules/blog/transport"

	catalogApp "github.com/yurythx/projeto-aurora/internal/modules/catalog/application"
	catalogInfra "github.com/yurythx/projeto-aurora/internal/modules/catalog/infrastructure"
	catalogTransport "github.com/yurythx/projeto-aurora/internal/modules/catalog/transport"

	contactApp "github.com/yurythx/projeto-aurora/internal/modules/contact/application"
	contactInfra "github.com/yurythx/projeto-aurora/internal/modules/contact/infrastructure"
	contactTransport "github.com/yurythx/projeto-aurora/internal/modules/contact/transport"

	directoryApp "github.com/yurythx/projeto-aurora/internal/modules/directory/application"
	directoryInfra "github.com/yurythx/projeto-aurora/internal/modules/directory/infrastructure"
	directoryTransport "github.com/yurythx/projeto-aurora/internal/modules/directory/transport"

	calendarApp "github.com/yurythx/projeto-aurora/internal/modules/calendar/application"
	calendarInfra "github.com/yurythx/projeto-aurora/internal/modules/calendar/infrastructure"
	calendarTransport "github.com/yurythx/projeto-aurora/internal/modules/calendar/transport"

	filesApp "github.com/yurythx/projeto-aurora/internal/modules/files/application"
	filesInfra "github.com/yurythx/projeto-aurora/internal/modules/files/infrastructure"
	filesTransport "github.com/yurythx/projeto-aurora/internal/modules/files/transport"

	wikiApp "github.com/yurythx/projeto-aurora/internal/modules/wiki/application"
	wikiInfra "github.com/yurythx/projeto-aurora/internal/modules/wiki/infrastructure"
	wikiTransport "github.com/yurythx/projeto-aurora/internal/modules/wiki/transport"

	searchApp "github.com/yurythx/projeto-aurora/internal/modules/search/application"
	searchDomain "github.com/yurythx/projeto-aurora/internal/modules/search/domain"
	searchTransport "github.com/yurythx/projeto-aurora/internal/modules/search/transport"

	signumApp "github.com/yurythx/projeto-aurora/internal/modules/signum/application"
	signumInfra "github.com/yurythx/projeto-aurora/internal/modules/signum/infrastructure"
	signumTransport "github.com/yurythx/projeto-aurora/internal/modules/signum/transport"

	tramiteApp "github.com/yurythx/projeto-aurora/internal/modules/tramite/application"
	tramiteInfra "github.com/yurythx/projeto-aurora/internal/modules/tramite/infrastructure"
	tramiteTransport "github.com/yurythx/projeto-aurora/internal/modules/tramite/transport"

	egressApp "github.com/yurythx/projeto-aurora/internal/modules/egress/application"
	egressInfra "github.com/yurythx/projeto-aurora/internal/modules/egress/infrastructure"
	egressTransport "github.com/yurythx/projeto-aurora/internal/modules/egress/transport"

	mercurioApp "github.com/yurythx/projeto-aurora/internal/modules/mercurio/application"
	mercurioInfra "github.com/yurythx/projeto-aurora/internal/modules/mercurio/infrastructure"
	mercurioTransport "github.com/yurythx/projeto-aurora/internal/modules/mercurio/transport"

	integrationsApp "github.com/yurythx/projeto-aurora/internal/modules/integrations/application"
	integrationsInfra "github.com/yurythx/projeto-aurora/internal/modules/integrations/infrastructure"
	integrationsTransport "github.com/yurythx/projeto-aurora/internal/modules/integrations/transport"

	usersApp "github.com/yurythx/projeto-aurora/internal/modules/users/application"
	usersInfra "github.com/yurythx/projeto-aurora/internal/modules/users/infrastructure"
	usersTransport "github.com/yurythx/projeto-aurora/internal/modules/users/transport"

	atendimentoApp "github.com/yurythx/projeto-aurora/internal/modules/atendimento/application"
	atendimentoInfra "github.com/yurythx/projeto-aurora/internal/modules/atendimento/infrastructure"
	atendimentoTransport "github.com/yurythx/projeto-aurora/internal/modules/atendimento/transport"

	organizacaoApp "github.com/yurythx/projeto-aurora/internal/modules/organizacao/application"
	organizacaoInfra "github.com/yurythx/projeto-aurora/internal/modules/organizacao/infrastructure"
	organizacaoTransport "github.com/yurythx/projeto-aurora/internal/modules/organizacao/transport"

	"github.com/yurythx/projeto-aurora/internal/platform/audit"
	"github.com/yurythx/projeto-aurora/internal/platform/branding"
	"github.com/yurythx/projeto-aurora/internal/platform/keycloakconfig"
	"github.com/yurythx/projeto-aurora/internal/platform/localauth"
	"github.com/yurythx/projeto-aurora/internal/platform/modules"
	"github.com/yurythx/projeto-aurora/internal/platform/outbox"
)

// Modules guarda o serviço de aplicação e os handlers HTTP de todo módulo
// de negócio da plataforma. Servindo como o "ponto de montagem" central
// que conecta domain/application/infrastructure/transport de cada módulo.
//
// Os handlers são construídos incondicionalmente (é só wiring de struct,
// sem I/O); a ativação/desativação de um módulo é aplicada em
// internal/app/router.go, que embrulha as rotas de cada módulo no
// middleware modules.Registry.Guard — um módulo desativado responde 404
// em toda a sua superfície e seu handler nunca roda. Consumidores de
// eventos, por não poderem ser desmontados de um canal AMQP em runtime,
// são decididos no boot (internal/app/worker.go, notifications.go).
type Modules struct {
	Users struct {
		Handlers *usersTransport.Handlers
	}
	Integrations struct {
		Service  *integrationsApp.Service
		Handlers *integrationsTransport.Handlers
	}
	Mercurio struct {
		Service  *mercurioApp.Service
		Handlers *mercurioTransport.Handlers
	}
	Egress struct {
		Service  *egressApp.Service
		Handlers *egressTransport.Handlers
	}
	Blog struct {
		Service  *blogApp.Service
		Handlers *blogTransport.Handlers
	}
	Catalog struct {
		Service  *catalogApp.Service
		Handlers *catalogTransport.Handlers
	}
	Contact struct {
		Service  *contactApp.Service
		Handlers *contactTransport.Handlers
	}
	Directory struct {
		Service  *directoryApp.Service
		Handlers *directoryTransport.Handlers
	}
	Calendar struct {
		Service  *calendarApp.Service
		Handlers *calendarTransport.Handlers
	}
	Files struct {
		Service  *filesApp.Service
		Handlers *filesTransport.Handlers
	}
	Wiki struct {
		Service  *wikiApp.Service
		Handlers *wikiTransport.Handlers
	}
	Search struct {
		Service  *searchApp.Service
		Handlers *searchTransport.Handlers
	}
	Signum struct {
		Service  *signumApp.Service
		Handlers *signumTransport.Handlers
	}
	Tramite struct {
		Service  *tramiteApp.Service
		Handlers *tramiteTransport.Handlers
	}
	SystemFeatures struct {
		Handlers *modules.Handlers
	}
	Audit struct {
		Handlers *audit.Handlers
	}
	KeycloakConfig struct {
		Handlers *keycloakconfig.Handlers
	}
	Branding struct {
		Handlers *branding.Handlers
	}
	OutboxStats struct {
		Handlers *outbox.StatsHandlers
	}
	LocalAuth struct {
		Handlers *localauth.Handlers
	}
	Example struct {
		Service  *exampleApp.Service
		Handlers *exampleTransport.Handlers
	}
	Atendimento struct {
		Service  *atendimentoApp.Service
		Handlers *atendimentoTransport.Handlers
	}
	Organizacao struct {
		Service  *organizacaoApp.Service
		Handlers *organizacaoTransport.Handlers
	}
}

// buildModules constrói cada módulo de negócio da plataforma. ctx é usado
// só para logs de diagnóstico do registro de módulos; o wiring em si não
// depende do estado de ativação (ver o comentário do tipo Modules).
func buildModules(ctx context.Context, deps *Dependencies) *Modules {
	_ = ctx
	auditWriter := audit.NewWriter(deps.DB)

	m := &Modules{}

	// Módulo de Integrações
	integrationsRepo := integrationsInfra.NewPostgresRepository(deps.DB)
	integrationsSvc := integrationsApp.NewService(integrationsRepo)
	m.Integrations.Service = integrationsSvc
	m.Integrations.Handlers = integrationsTransport.NewHandlers(integrationsSvc, deps.Logger)

	// Módulo Mercurio — mensageria interna. O serviço recebe o Hub como
	// HubPublisher para o fan-out em tempo real; a entrega durável usa o
	// Outbox + RabbitMQ.
	mercurioRepo := mercurioInfra.NewPostgresRepository(deps.DB)
	mercurioSvc := mercurioApp.NewService(deps.TxRunner, mercurioRepo, deps.Outbox, deps.Hub, deps.Storage, deps.Config.MinIO.Bucket, deps.Logger)
	m.Mercurio.Service = mercurioSvc
	m.Mercurio.Handlers = mercurioTransport.NewHandlers(mercurioSvc, deps.Logger)

	// Módulo Egress — barramento de eventos de domínio (Outbox/RabbitMQ já
	// existente) + webhooks de saída assíncronos. O worker de entrega e o
	// consumidor do dispatcher rodam só no cmd/worker (ver worker.go); a
	// API só monta as rotas administrativas.
	egressTargetRepo := egressInfra.NewTargetRepository(deps.DB, deps.ConfigCipher)
	egressDeliveryRepo := egressInfra.NewDeliveryRepository(deps.DB)
	egressPushRepo := egressInfra.NewPushSubscriptionRepository(deps.DB)
	egressDeliverer := egressInfra.NewHTTPDeliverer(deps.Config.Egress.Timeout, deps.Config.Egress.AllowedHosts)
	egressSvc := egressApp.NewService(egressTargetRepo, egressDeliveryRepo, egressPushRepo, egressDeliverer, deps.Config.Egress.VAPIDPublicKey, egressApp.Config{
		MaxAttempts:  deps.Config.Egress.MaxAttempts,
		BatchSize:    deps.Config.Egress.BatchSize,
		PollInterval: deps.Config.Egress.PollInterval,
	}, deps.Logger)
	m.Egress.Service = egressSvc
	m.Egress.Handlers = egressTransport.NewHandlers(egressSvc, deps.Logger)

	// Módulo Blog — publicações internas. Como o Exemplo/Mercúrio, grava o
	// evento de domínio (blog.post.published) no Outbox na mesma transação
	// da publicação; nenhum consumidor próprio (o Egress pega via fila #).
	blogRepo := blogInfra.NewPostgresRepository(deps.DB)
	blogSvc := blogApp.NewService(deps.TxRunner, blogRepo, blogRepo, deps.Outbox, deps.Storage, deps.Config.MinIO.Bucket, deps.Logger)
	m.Blog.Service = blogSvc
	m.Blog.Handlers = blogTransport.NewHandlers(blogSvc, deps.Logger)

	// Módulo Catalog — central de serviços do site público (pivô Aurora
	// v3). Leitura 100% pública; publicar emite catalog.service.published
	// no Outbox (consumível pelo Egress).
	catalogRepo := catalogInfra.NewPostgresRepository(deps.DB)
	catalogSvc := catalogApp.NewService(deps.TxRunner, catalogRepo, deps.Outbox, deps.Storage, deps.Config.MinIO.Bucket, deps.Logger)
	m.Catalog.Service = catalogSvc
	m.Catalog.Handlers = catalogTransport.NewHandlers(catalogSvc, deps.Logger)

	// Módulo Contact — formulário de contato público. Não tem client SMTP
	// próprio: emite contact.message.submitted no Outbox e reaproveita o
	// Egress já existente para notificar por e-mail/webhook.
	contactRepo := contactInfra.NewPostgresRepository(deps.DB)
	contactSvc := contactApp.NewService(deps.TxRunner, contactRepo, deps.Outbox, deps.Logger)
	m.Contact.Service = contactSvc
	m.Contact.Handlers = contactTransport.NewHandlers(contactSvc, deps.TrustedProxies, deps.Logger)

	// Módulo Directory — diretório de pessoas (busca, departamentos
	// curados a partir de grupo do AD, perfil estendido self-service).
	// Depende de users (DependsOn no catálogo): não por import Go — a
	// resolução de identidade e a listagem de pessoas são SQL direto
	// contra a tabela "users" compartilhada, nunca um import do pacote
	// usersApp/usersInfra (nenhum módulo de negócio desta plataforma
	// importa outro).
	directoryRepo := directoryInfra.NewPostgresRepository(deps.DB)
	directorySvc := directoryApp.NewService(deps.TxRunner, directoryRepo, deps.Outbox, deps.Logger)
	m.Directory.Service = directorySvc
	m.Directory.Handlers = directoryTransport.NewHandlers(directorySvc, deps.Logger, deps.Config.MaxPageSize)

	// Módulo Calendar — agenda corporativa (eventos self-service + salas
	// curadas). Emite calendar.event.created no Outbox na criação, mesmo
	// blueprint que example/directory/blog já usam.
	calendarRepo := calendarInfra.NewPostgresRepository(deps.DB)
	calendarSvc := calendarApp.NewService(deps.TxRunner, calendarRepo, deps.Outbox, deps.Logger)
	m.Calendar.Service = calendarSvc
	m.Calendar.Handlers = calendarTransport.NewHandlers(calendarSvc, deps.Logger)

	// Módulo Files — arquivos/drive sobre o MinIO já usado internamente
	// (blog/catalog/mercurio). Emite files.object.uploaded no Outbox na
	// confirmação de upload, mesmo blueprint de example/directory/
	// calendar; entra também no GC de upload órfão (ver
	// internal/app/storage_gc.go).
	filesRepo := filesInfra.NewPostgresRepository(deps.DB)
	filesSvc := filesApp.NewService(deps.TxRunner, filesRepo, deps.Outbox, deps.Storage, deps.Config.MinIO.Bucket, deps.Logger)
	m.Files.Service = filesSvc
	m.Files.Handlers = filesTransport.NewHandlers(filesSvc, deps.Logger)

	// Módulo Wiki — base de conhecimento multi-editor (qualquer
	// autenticado cria E edita qualquer página visível), com
	// wiki_revisions como rede de segurança — mesmo blueprint de
	// blog.Service.Update, sem Outbox (nenhum consumidor externo precisa
	// saber de uma edição de wiki hoje).
	wikiRepo := wikiInfra.NewPostgresRepository(deps.DB)
	wikiSvc := wikiApp.NewService(deps.TxRunner, wikiRepo)
	m.Wiki.Service = wikiSvc
	m.Wiki.Handlers = wikiTransport.NewHandlers(wikiSvc, deps.Logger)

	// Módulo Search — busca global cross-módulo (§3.5). Sem tabela/estado
	// próprio: cada Provider chama de volta o Service JÁ CONSTRUÍDO do
	// módulo de origem (ver internal/app/search_providers.go), então
	// precisa vir DEPOIS de todos eles no wiring.
	searchSvc := searchApp.NewService([]searchDomain.Provider{
		&directorySearchProvider{svc: directorySvc},
		&blogSearchProvider{svc: blogSvc},
		&wikiSearchProvider{svc: wikiSvc},
		&filesSearchProvider{svc: filesSvc},
		&catalogSearchProvider{svc: catalogSvc},
	}, deps.Logger)
	m.Search.Service = searchSvc
	m.Search.Handlers = searchTransport.NewHandlers(searchSvc, deps.ModuleRegistry.Enabled, deps.Logger)

	// Módulo Signum — motor de assinatura eletrônica reutilizável (§3.7).
	// Sem outboxWriter próprio de outros módulos: emite
	// signum.envelope.signed/refused pro Outbox compartilhado
	// (deps.Outbox), igual todo mundo. kcEnv é o fallback de .env pra
	// IssuerURL quando a config dinâmica (deps.KeycloakCfg) nunca foi
	// configurada — ver infrastructure.VerifyKeycloakPassword.
	signumRepo := signumInfra.NewPostgresRepository(deps.DB)
	signumSvc := signumApp.NewService(deps.TxRunner, signumRepo, deps.Outbox, deps.KeycloakCfg, deps.Config.Keycloak, deps.Logger)
	m.Signum.Service = signumSvc
	m.Signum.Handlers = signumTransport.NewHandlers(signumSvc, deps.Logger)

	// Módulo Trâmite — sub-fases A+B+C (§3.6.6): processo numerado +
	// documento (produzido no sistema ou anexado sobre o MinIO) +
	// assinatura eletrônica delegada ao Signum via tramiteSignatureOpener
	// (internal/app/tramite_signum.go — o único ponto onde os pacotes Go
	// de tramite e signum se encontram) + tramitação entre unidades com
	// histórico imutável, publicando tramite.processo.tramitado/
	// concluido/arquivado no Outbox compartilhado (deps.Outbox, mesmo
	// writer que signum já usa — cada módulo só muda o eventType).
	// Entra também no GC de upload órfão (storage_gc.go).
	tramiteRepo := tramiteInfra.NewPostgresRepository(deps.DB)
	tramiteSvc := tramiteApp.NewService(deps.TxRunner, tramiteRepo, deps.Storage, deps.Config.MinIO.Bucket,
		tramiteSignatureOpener{svc: signumSvc}, deps.Outbox, deps.Logger)
	m.Tramite.Service = tramiteSvc
	m.Tramite.Handlers = tramiteTransport.NewHandlers(tramiteSvc, deps.Logger)

	// Módulo de Usuários — construído DEPOIS do Trâmite de propósito:
	// LGPD self-service (§5.1) usa tramiteSvc via dois adapters
	// (internal/app/users_lgpd.go, mesmo racional de
	// tramiteSignatureOpener) — nenhum outro módulo desta lista depende
	// de m.Users existir mais cedo, então a reordenação é segura.
	usersRepo := usersInfra.NewPostgresRepository(deps.DB)
	usersSvc := usersApp.NewService(usersRepo, auditWriter,
		[]usersApp.ExportContributor{tramiteExportContributor{svc: tramiteSvc}},
		tramiteLGPDOpener{svc: tramiteSvc}, deps.Logger)
	m.Users.Handlers = usersTransport.NewHandlers(usersSvc, deps.Logger, deps.Config.MaxPageSize)

	// Registro central de módulos (system_features) — GET /api/system/features
	// e o CRUD administrativo.
	m.SystemFeatures.Handlers = modules.NewHandlers(deps.ModuleRegistry, auditWriter, deps.Logger)

	// Consulta da trilha de auditoria (módulo audit — núcleo).
	m.Audit.Handlers = audit.NewHandlers(audit.NewReader(deps.DB), deps.Logger)

	// Configuração dinâmica do Keycloak (menu Configurações > Identidade)
	m.KeycloakConfig.Handlers = keycloakconfig.NewHandlers(deps.KeycloakCfg, deps.Verifier, deps.Config.Keycloak, auditWriter, deps.Logger)

	// Branding white-label da organização (Configurações) — leitura autenticada,
	// escrita restrita a branding:manage (aurora-admin).
	m.Branding.Handlers = branding.NewHandlers(branding.NewStore(deps.DB), auditWriter, deps.Logger)

	// Estatísticas do Outbox (painel de Monitoramento)
	m.OutboxStats.Handlers = outbox.NewStatsHandlers(deps.OutboxStats, deps.Logger)

	localAuthStore := localauth.NewPostgresStore(deps.DB)
	m.LocalAuth.Handlers = localauth.NewHandlers(localAuthStore, deps.LocalSigner, auditWriter, deps.Logger)

	// Módulo Exemplo (Template genérico para novos módulos)
	exampleRepo := exampleInfra.NewPostgresRepository(deps.DB)
	exampleSvc := exampleApp.NewService(deps.TxRunner, exampleRepo, deps.Outbox, deps.Logger)
	m.Example.Service = exampleSvc
	m.Example.Handlers = exampleTransport.NewHandlers(exampleSvc, deps.Logger)

	atendimentoRepo := atendimentoInfra.NewPostgresRepository(deps.DB)
	atendimentoSvc := atendimentoApp.NewService(atendimentoRepo, deps.ModuleRegistry)
	m.Atendimento.Service = atendimentoSvc
	m.Atendimento.Handlers = atendimentoTransport.NewHandlers(atendimentoSvc, deps.Logger, deps.Config.MaxPageSize)

	organizacaoRepo := organizacaoInfra.NewPostgresRepository(deps.DB)
	organizacaoSvc := organizacaoApp.NewService(organizacaoRepo)
	m.Organizacao.Service = organizacaoSvc
	m.Organizacao.Handlers = organizacaoTransport.NewHandlers(organizacaoSvc, deps.Logger)

	return m
}
