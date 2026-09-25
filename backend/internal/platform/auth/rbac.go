package auth

// Roles conhecidos pela plataforma. O Keycloak é a fonte da verdade
// para a atribuição de roles a usuários — estas constantes existem só para
// que as checagens de autorização no código não espalhem strings literais
// (evitando erro de digitação silencioso).
const (
	RoleUser               RoleName = "aurora-user"
	RoleAdmin              RoleName = "aurora-admin"
	RoleIntegrationManager RoleName = "aurora-integration-manager"
	RoleAuditor            RoleName = "aurora-auditor"
)

type RoleName = string

// Permission é uma capacidade granular verificada por RequirePermission,
// para handlers em que "o chamador tem um destes roles" não é específico
// o bastante. O mapeamento abaixo é deliberadamente simples — uma única
// tabela estática role -> permissões — e é o ponto de extensão caso a
// plataforma precise no futuro de regras por recurso ou baseadas em
// atributos (attribute-based access control).
type Permission string

const (
	PermUsersRead          Permission = "users:read"
	PermUsersManage        Permission = "users:manage"
	PermIntegrationsRead   Permission = "integrations:read"
	PermIntegrationsTest   Permission = "integrations:test"
	PermIntegrationsManage Permission = "integrations:manage"
	PermAuditRead          Permission = "audit:read"
	// PermBlogManage autoriza criar/editar/publicar/remover posts do
	// módulo Blog. Como PermModulesManage/PermKeycloakManage, não é
	// concedida a nenhum role em rolePermissions — só o aurora-admin a
	// possui (via HasPermission). A leitura dos posts publicados exige
	// apenas autenticação, não permissão. Um futuro papel "aurora-editor"
	// entraria aqui.
	PermBlogManage Permission = "blog:manage"
	// PermBlogApprove autoriza aprovar ou pedir ajuste num post em
	// revisão (§5.2 do roadmap — workflow editorial). Deliberadamente
	// SEPARADA de PermBlogManage: hoje as duas só existem via o atalho
	// de aurora-admin (nenhuma concedida em rolePermissions), mas o
	// desenho já deixa pronto um futuro papel "revisor" (blog:approve,
	// sem blog:manage) sem exigir mudança de código — mesmo racional que
	// já levou PermModulesManage/PermKeycloakManage a serem permissões
	// próprias em vez de reaproveitar uma existente. A checagem de que
	// quem aprova não pode ser o autor do post é feita à parte, no
	// application.Service (Approve/RequestChanges) — esta permissão só
	// resolve "pode revisar ALGUM post", nunca "pode revisar ESTE post".
	PermBlogApprove Permission = "blog:approve"
	// PermBrandingManage autoriza editar o branding white-label da
	// organização (nome, logo, favicon, contatos de suporte) em
	// Configurações. Admin-only — é uma configuração global, uma só para
	// todo o tenant. Ler o branding (GET /api/v1/branding) exige só
	// autenticação.
	PermBrandingManage Permission = "branding:manage"
	// PermMercurioManage autoriza a administração das salas departamentais
	// do Mercúrio (criar/editar/remover e mapear grupos do AD). Também
	// admin-only — uma sala mal mapeada expõe uma conversa a quem não
	// deveria vê-la. Participar do chat e listar as próprias salas não
	// exige esta permissão, só autenticação.
	PermMercurioManage Permission = "mercurio:manage"
	// PermModulesManage não é concedida a nenhum role em rolePermissions
	// abaixo — só o aurora-admin a possui, através do atalho em
	// HasPermission. Ativar/desativar um módulo em produção afeta todo
	// mundo imediatamente (rotas somem, consumidores param), então é
	// deliberadamente restrito ao papel mais privilegiado, sem meio-termo
	// por role. Ver internal/platform/modules.
	PermModulesManage Permission = "modules:manage"
	// PermKeycloakManage, pelo mesmo motivo de PermModulesManage
	// acima, também não é concedida a nenhum role em rolePermissions — só
	// o aurora-admin (via HasPermission). Ver a configuração dinâmica do
	// Keycloak em internal/platform/keycloakconfig — errar aqui derruba
	// a autenticação de TODA a plataforma, então esta permissão é
	// deliberadamente mais restrita do que PermIntegrationsManage.
	PermKeycloakManage Permission = "keycloak:manage"
	// PermCatalogManage autoriza criar/editar/publicar/remover serviços do
	// módulo Catalog (a central de serviços do site público). Como
	// PermBlogManage, não é concedida a nenhum role em rolePermissions —
	// só o aurora-admin a possui (via HasPermission). A leitura do
	// catálogo é 100% pública (nem autenticação exige).
	PermCatalogManage Permission = "catalog:manage"
	// PermContactRead autoriza consultar as mensagens recebidas pelo
	// formulário de contato público (módulo Contact). Só o aurora-admin
	// (via HasPermission) — a mensagem carrega PII (nome, e-mail,
	// telefone) de um visitante externo, então segue o mesmo racional de
	// minimização de PermUsersRead (gap G-10).
	PermContactRead Permission = "contact:read"
	// PermDirectoryManage autoriza criar/editar/remover os departamentos
	// curados (mapeamento de grupo do AD) do módulo Directory. Como
	// PermBlogManage/PermCatalogManage, não é concedida a nenhum role em
	// rolePermissions — só o aurora-admin (via HasPermission). Ler o
	// diretório e editar o próprio perfil estendido não exige esta
	// permissão, só autenticação — só a curadoria de departamento é
	// restrita.
	PermDirectoryManage Permission = "directory:manage"
	// PermCalendarManage autoriza CRUD das salas (recurso físico curado)
	// do módulo Calendar. Eventos são self-service (qualquer autenticado
	// cria o próprio; editar/apagar o de outra pessoa também exige esta
	// permissão — ver application.Service.Update/Delete) — só a curadoria
	// de sala e a moderação alheia passam por aqui.
	PermCalendarManage Permission = "calendar:manage"
	// PermFilesManage autoriza apagar pasta/arquivo de OUTRA pessoa no
	// módulo Files (moderação) — criar pasta, enviar e apagar o PRÓPRIO
	// arquivo/pasta não exige nada além de sessão válida (self-service,
	// mesmo racional de PermCalendarManage).
	PermFilesManage Permission = "files:manage"
	// PermWikiManage autoriza apagar página de OUTRA pessoa no módulo
	// wiki (moderação) — criar página e EDITAR qualquer página (mesmo a
	// de outro autor: é o diferencial do módulo) não exigem nada além de
	// sessão válida. Só apagar o trabalho alheio passa por aqui — mesmo
	// racional de PermFilesManage/PermCalendarManage.
	PermWikiManage Permission = "wiki:manage"
	// PermSignumManage autoriza cancelar um envelope pendente e ver
	// qualquer envelope (não só os que o próprio admin assina) — abrir
	// um envelope (OpenEnvelope) é uma chamada Go interna do módulo
	// dono, não uma rota HTTP, então não passa por aqui; assinar/recusar
	// a própria assinatura também não exige esta permissão, só ser um
	// dos signatários designados.
	PermSignumManage Permission = "signum:manage"
	// PermTramiteCreate autoriza abrir um processo novo (§3.6.6) — ao
	// contrário de blog/wiki/files/calendar, tramite NÃO é self-service
	// por padrão: exige esta permissão. Editar/apagar documento dentro
	// de um processo próprio não exige nada além de ter aberto o
	// processo (ou tramite:manage) — ver PermTramiteManage.
	PermTramiteCreate Permission = "tramite:create"
	// PermTramiteManage autoriza administrar o catálogo de tipos de
	// documento e agir sobre processo/documento de OUTRA pessoa
	// (concluir, arquivar, editar/cancelar documento) — mesmo racional
	// de PermFilesManage/PermWikiManage.
	PermTramiteManage Permission = "tramite:manage"
	// PermTramiteRoute autoriza tramitar (encaminhar) um processo entre
	// unidades (§3.6.6 sub-fase C) — DELIBERADAMENTE não tem o mesmo
	// carve-out de autoria que Concluir/Arquivar (autor OU
	// tramite:manage): encaminhar pra outra unidade é uma ação
	// organizacional de despacho, não uma edição do próprio recurso, daí
	// uma permissão dedicada em vez de "quem abriu o processo pode
	// tramitar pra qualquer lugar".
	PermTramiteRoute Permission = "tramite:route"
)

// rolePermissions concede ao aurora-admin toda permissão implicitamente
// (verificado à parte em HasPermission) e dá aos demais roles o conjunto
// mínimo implicado pelo próprio nome — ex.: um "aurora-auditor" só pode ler
// (audit, users, integrations), nunca escrever.
//
// Gap G-10 da auditoria de conformidade: RoleUser tinha PermUsersRead, o
// que — combinado com GET /api/v1/users devolvendo o e-mail de cada
// usuário — deixava QUALQUER servidor autenticado ler o diretório inteiro
// com PII (LGPD art. 6º III, minimização). RoleUser agora não tem
// nenhuma permissão de diretório: um usuário comum enxerga só a si mesmo
// via GET /api/v1/me (que exige apenas autenticação, não permissão).
// Listar/consultar terceiros passa a exigir aurora-auditor ou
// aurora-admin.
var rolePermissions = map[RoleName][]Permission{
	RoleIntegrationManager: {
		PermIntegrationsRead,
		PermIntegrationsTest,
		PermIntegrationsManage,
	},
	RoleAuditor: {
		PermAuditRead,
		PermUsersRead,
		PermIntegrationsRead,
	},
}

// HasPermission reporta se os roles de identity concedem permission.
// aurora-admin sempre tem toda permissão, independente do mapa acima.
func HasPermission(identity Identity, permission Permission) bool {
	if identity.IsAdmin() || identity.HasRole(RoleAdmin) {
		return true
	}
	for _, role := range identity.Roles {
		for _, p := range rolePermissions[role] {
			if p == permission {
				return true
			}
		}
	}
	return false
}
