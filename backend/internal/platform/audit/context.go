package audit

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/logging"
)

// FromRequest devolve uma Entry já preenchida com a proveniência da
// requisição HTTP: ator (id interno, subject, roles), IP de origem, user
// agent, contexto organizacional (lotações efetivas) e correlation id. O
// chamador só completa Action/Resource/Before/After/Metadata.
//
// O IP vem de r.RemoteAddr, que o middleware httpserver.TrustedRealIP já
// reescreveu com o IP real do cliente quando a conexão chega por um proxy
// confiável.
func FromRequest(r *http.Request) Entry {
	e := FromContext(r.Context())
	e.IPAddress = r.RemoteAddr
	e.UserAgent = r.UserAgent()
	return e
}

// FromContext preenche ator, contexto organizacional e correlation id a
// partir do context (sem IP/user agent) — para casos de uso da camada de
// aplicação que recebem só o ctx.
func FromContext(ctx context.Context) Entry {
	var e Entry
	if cid := logging.CorrelationID(ctx); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			e.CorrelationID = &id
		}
	}
	if identity, ok := auth.IdentityFromContext(ctx); ok {
		e.ActorSubject = identity.Subject
		e.ActorRoles = identity.Roles
		if actor, _, ok := auth.ActorUUID(ctx); ok {
			e.ActorID = actor
		}
		if len(identity.Scopes) > 0 {
			e.EntityContext = map[string]any{"scopes": identity.Scopes}
		}
	}
	return e
}

// Meta é açúcar para montar uma Entry a partir do contexto com ação,
// recurso e diff numa só chamada.
func Meta(ctx context.Context, action, resourceType, resourceID string, before, after any) Entry {
	e := FromContext(ctx)
	e.Action = action
	e.ResourceType = resourceType
	e.ResourceID = resourceID
	e.Before = before
	e.After = after
	return e
}
