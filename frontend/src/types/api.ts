// Formatos compartilhados que espelham os DTOs do backend do Projeto Aurora.
// Mantidos manualmente em sincronia com docs/openapi.yaml — qualquer novo
// campo exposto pela API precisa ser refletido aqui para o frontend
// enxergá-lo com tipagem.

export interface User {
  id: string;
  username: string;
  email?: string;
  display_name: string;
  active: boolean;
  created_at: string;
  last_seen_at?: string;
}

// Dashboard personalizável (Kernel, §6) — a "key" é opaca pro backend;
// o catálogo de widgets (quais existem, a qual módulo cada um pertence)
// vive só no frontend (ver components/dashboard/widgets.tsx).
export interface DashboardWidgetPref {
  key: string;
  visible: boolean;
}

export interface DashboardPrefsResponse {
  widgets: DashboardWidgetPref[];
}

export type IntegrationStatus = "unknown" | "online" | "offline" | "degraded" | "disabled";

export interface Integration {
  id: string;
  key: string;
  name: string;
  type: string;
  enabled: boolean;
  status: IntegrationStatus;
  last_check_at?: string;
  last_success_at?: string;
  last_error?: string;
}

export type JobStatus = "queued" | "processing" | "completed" | "failed" | "dead_letter";

export interface TestJobResponse {
  job_id: string;
  status: JobStatus;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

// Envelope de paginação por cursor (Mercurio, Egress, auditoria).
export interface CursorPage<T> {
  items: T[];
  next_cursor?: string;
  has_more: boolean;
}

// GET /api/v1/system/features — manifesto público de módulos.
export type SystemFeatures = Record<string, boolean>;

// --- Mercurio (mensageria interna) -------------------------------------
// Módulo Directory (diretório de pessoas) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md §3.1.
export interface DirectoryDepartamento {
  id: string;
  nome: string;
  descricao: string;
  ad_groups: string[];
  ordem: number;
}

// Perfil estendido self-service (GET/PATCH /api/v1/directory/me) — cargo/
// skills/bio/ramal, nunca nome/e-mail/departamento (esses vêm do AD).
export interface DirectoryProfile {
  cargo: string;
  skills: string[];
  bio: string;
  ramal: string;
}

// Uma linha do diretório — mesmo formato pra listagem e card completo
// (GET /api/v1/directory e GET /api/v1/directory/{id}): nada aqui é
// sensível como o e-mail de UserListItem, é o que a pessoa já decidiu
// tornar navegável no diretório interno.
export interface DirectoryPerson {
  id: string;
  username: string;
  display_name: string;
  cargo: string;
  skills: string[];
  bio: string;
  ramal: string;
  departamentos: string[];
}

// Módulo Calendar (agenda corporativa) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md §3.2.
export interface CalendarRoom {
  id: string;
  nome: string;
  localizacao: string;
  capacidade: number;
}

// Reserva de sala não é uma entidade separada: um evento COM room_id já
// É a reserva (ver domain.Event no backend).
export interface CalendarEvent {
  id: string;
  title: string;
  description: string;
  location: string;
  starts_at: string;
  ends_at: string;
  all_day: boolean;
  room_id?: string;
  audience_groups: string[];
  created_by_name: string;
  // Computado pelo backend (autor OU calendar:manage) — nunca reimplementar
  // essa regra no cliente, ver o comentário de toEventResponse no backend.
  can_edit: boolean;
  created_at: string;
  updated_at: string;
}

// Módulo Files (arquivos/drive) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md §3.3.
export interface FilesFolder {
  id: string;
  name: string;
  parent_id?: string;
  audience_groups: string[];
  created_by_name: string;
  // Computado pelo backend (autor OU files:manage) — mesma razão de
  // CalendarEvent.can_edit.
  can_delete: boolean;
  created_at: string;
}

export interface FilesObject {
  id: string;
  folder_id?: string;
  name: string;
  content_type: string;
  size_bytes: number;
  uploaded_by_name: string;
  can_delete: boolean;
  created_at: string;
}

// Mesma forma de storage.Ticket no backend (ver
// lib/api/presignedUpload.ts) — devolvido por POST /files/uploads.
export interface FilesUploadTicket {
  object_key: string;
  upload_url: string;
  upload_fields?: Record<string, string> | null;
  expires_in_seconds: number;
}

// Módulo Wiki (base de conhecimento) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md
// §3.4. Diferencial frente a Files/Calendar: NÃO existe can_edit — qualquer
// autenticado que enxerga a página pode editá-la (a colaboração é o ponto);
// só can_delete é computado (autor OU wiki:manage).
export interface WikiPage {
  id: string;
  slug: string;
  title: string;
  body: string;
  parent_id?: string;
  audience_groups: string[];
  created_by_name: string;
  last_edited_by_name: string;
  can_delete: boolean;
  created_at: string;
  updated_at: string;
}

export interface WikiRevision {
  id: string;
  title: string;
  body: string;
  edited_by_name: string;
  created_at: string;
}

// Módulo Search (busca global) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md
// §3.5. Sem paginação nem card own: é uma prévia (⌘K), agrupada por tipo
// pelo backend (o frontend nunca reimplementa o agrupamento). "files" vem
// sempre sem url — sem rota endereçável por arquivo (mesmo corte de
// escopo de §3.3) —, tratado à parte: aciona o download pelo id.
export type SearchResultType = "directory" | "blog" | "wiki" | "files" | "catalog";

export interface SearchResult {
  type: SearchResultType;
  id: string;
  title: string;
  snippet?: string;
  url?: string;
}

export interface SearchResponse {
  results: Partial<Record<SearchResultType, SearchResult[]>>;
}

export type MercurioRoomKind = "broadcast" | "department" | "dm" | "channel";

// Quem pode ENVIAR mensagem na sala — "admins" na sala broadcast por
// padrão (qualquer um com acesso ainda LÊ; isto só afeta postar).
export type MercurioPostPolicy = "all" | "admins";

export interface MercurioRoom {
  id: string;
  slug: string;
  kind: MercurioRoomKind;
  name: string;
  ad_groups?: string[];
  post_policy: MercurioPostPolicy;
  // Só significativo pra kind="channel" (Onda 5, ADR 018) — omitido
  // (false) pros demais kinds.
  is_private?: boolean;
  created_at: string;
  // Só vem preenchido em GET /rooms (ListRoomsForUser) — mensagem mais
  // recente da sala, ou a criação se nunca teve mensagem. Base da
  // ordenação "mais recente primeiro" estilo WhatsApp (ADR 019); a lista
  // já chega ordenada do backend, este campo é só informativo.
  last_activity_at?: string;
}

// GET /api/v1/mercurio/contacts — pessoas que já compartilham uma sala
// com o chamador; base do picker de "convidar pro canal". NUNCA o
// diretório de usuários (ADR 018).
export interface MercurioContact {
  subject: string;
  username: string;
}

// GET /api/v1/users/roster — roster estilo MSN/Spark (todo mundo ativo,
// exceto quem chama), aberto a QUALQUER autenticado. Nunca traz e-mail
// (ver domain.RosterEntry no backend) — só o suficiente pra listar e
// iniciar uma DM.
export interface UserRosterEntry {
  subject: string;
  username: string;
  display_name?: string;
}

// Anexo já resolvido numa mensagem (URL pré-assinada de leitura, curta).
export interface MercurioAttachment {
  filename: string;
  content_type: string;
  size: number;
  url: string;
}

export interface MercurioMessage {
  id: string;
  room_id: string;
  sender_subject: string;
  sender_username: string;
  body: string;
  // Assunto (estilo Zulip, ADR 017) — "" ou ausente = sem tópico. Nunca
  // preenchido em sala kind="dm".
  topic?: string;
  attachment?: MercurioAttachment;
  edited_at?: string;
  deleted?: boolean;
  created_at: string;
}

// GET /api/v1/mercurio/rooms/{slug}/topics — tópicos com mensagem na
// sala, mais recém-ativo primeiro.
export interface MercurioTopic {
  topic: string;
  message_count: number;
  last_message_at: string;
}

// GET /api/v1/mercurio/rooms/{slug}/messages/{id}/edits — corpo anterior a
// cada edição (histórico), mais recente primeiro.
export interface MercurioMessageEdit {
  old_body: string;
  edited_by_subject: string;
  created_at: string;
}

// GET /api/v1/mercurio/mentions — "fui mencionado", mais recente primeiro.
export interface MercurioMention {
  message_id: string;
  room_slug: string;
  room_name: string;
  room_kind: MercurioRoomKind;
  body: string;
  sender_username: string;
  created_at: string;
}

// POST /api/v1/mercurio/rooms/{slug}/attachments — URL pré-assinada de upload.
export interface MercurioUploadTicket {
  object_key: string;
  upload_url: string;
  upload_fields: Record<string, string>;
  expires_in_seconds: number;
}

// Referência de anexo enviada no corpo de POST .../messages.
export interface MercurioAttachmentRef {
  object_key: string;
  filename: string;
  content_type: string;
  size: number;
}

// GET /api/v1/mercurio/unread — resumo de não-lidos para o badge da nav.
export interface MercurioUnread {
  rooms: Record<string, number>; // room_id -> contagem (só salas com pendência)
  total: number;
}

// Corpo de POST/PUT /api/v1/mercurio/admin/rooms — sala departamental.
// post_policy omitido equivale a "all".
export interface MercurioRoomInput {
  name: string;
  ad_groups: string[];
  post_policy?: MercurioPostPolicy;
}

// GET /api/v1/mercurio/presence — quem está com o Mercúrio aberto agora
// (nesta réplica da API — ver ADR 017).
export interface MercurioPresence {
  online: string[];
}

// --- Egress (webhooks de saída) -------------------------------------
export type EgressTargetKind = "generic" | "n8n" | "zabbix" | "grafana" | "push";

export interface EgressPushVapidKey {
  enabled: boolean;
  public_key?: string;
}

export interface EgressTarget {
  id: string;
  key: string;
  name: string;
  kind: EgressTargetKind;
  url: string;
  event_types: string[];
  extra_headers: Record<string, string>;
  // Corpo em Go html/template (sandboxed) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md
  // §5.3. "" (default) preserva o corpo moldado por kind (JSON). Campos
  // disponíveis: {{.Type}}, {{.Source}}, {{.OccurredAt}} (time.Time — aceita
  // métodos como {{.OccurredAt.Format "02/01/2006"}}), {{.CorrelationID}},
  // {{.Payload.<campo>}}.
  body_template: string;
  enabled: boolean;
  secret_set: boolean;
}

export type EgressDeliveryStatus =
  | "pending"
  | "delivering"
  | "delivered"
  | "failed"
  | "dead_letter";

export interface EgressDelivery {
  id: string;
  target_id: string;
  event_id: string;
  event_type: string;
  status: EgressDeliveryStatus;
  attempts: number;
  last_status_code?: number;
  last_error?: string;
  next_attempt_at: string;
  created_at: string;
}

// POST /api/v1/egress/targets/{id}/test — disparo sintético ao vivo.
export interface EgressTestResult {
  delivered: boolean;
  status_code: number;
  retryable: boolean;
  error?: string;
  signature_sent: boolean;
}

// --- Blog (publicações internas) -------------------------------------
// "em_revisao" — workflow editorial (§5.2 do roadmap): rascunho enviado
// pra revisão, aguardando alguém com blog:approve (que não seja o
// próprio autor) aprovar ou pedir ajuste.
export type BlogPostStatus = "draft" | "em_revisao" | "published";

// Item de listagem (GET /api/v1/blog/posts e /blog/admin/posts) — sem corpo.
export interface BlogPostSummary {
  id: string;
  slug: string;
  title: string;
  summary: string;
  status: BlogPostStatus;
  tags: string[];
  cover_url?: string; // URL pré-assinada de leitura da capa (1 h)
  // Destaque manual do admin — GET /blog/posts/pinned devolve este post
  // (se publicado e visível), IGNORANDO a ordem por data (ADR 021).
  pinned: boolean;
  // Vazio = visível a todo autenticado; não-vazio = só quem tem algum
  // desses grupos do AD.
  audience_groups: string[];
  // Categoria — nunca obrigatória; ausente/undefined = sem categoria.
  // Name/slug/color vêm resolvidos pelo backend (cache curto server-side,
  // ver Service.categoryByID) para a UI não precisar de um segundo fetch
  // só pra pintar o badge.
  category_id?: string;
  category_slug?: string;
  category_name?: string;
  category_color?: string;
  requires_ack: boolean;
  publish_at?: string; // agendamento (só relevante em rascunho)
  author_name: string;
  published_at?: string;
  created_at: string;
  // Estimativa (contagem de palavras do corpo / 200 por minuto).
  reading_minutes: number;
  // Feedback de quem pediu ajuste (status volta a "draft") — "" fora
  // desse caso. Limpo de novo ao reenviar pra revisão.
  review_note?: string;
}

// Post completo (GET /api/v1/blog/posts/{slug} e /blog/admin/posts/{id}).
export interface BlogPost extends BlogPostSummary {
  body: string;
  updated_at: string;
  my_acked_at?: string; // preenchido só em GET /blog/posts/{slug}
}

// Corpo de criação/edição (POST/PUT /api/v1/blog/admin/posts).
export interface BlogPostInput {
  title: string;
  summary?: string;
  body: string;
  tags?: string[];
  slug?: string;
  cover_object_key?: string;
  audience_groups?: string[];
  // null explícito = remover a categoria; ausente = não mexer (mesma
  // convenção de cover_object_key); string = a nova categoria.
  category_id?: string | null;
  requires_ack?: boolean;
  publish_at?: string | null;
}

// --- Categorias do Blog -------------------------------------------------
// GET /api/v1/blog/categories (leitura aberta a todo autenticado) e
// POST/PUT/DELETE /api/v1/blog/admin/categories (blog:manage).
export interface BlogCategory {
  id: string;
  slug: string;
  name: string;
  description: string;
  color: string; // "#RRGGBB"
  created_at: string;
  updated_at: string;
}

export interface BlogCategoryInput {
  name: string;
  slug?: string; // "" = derivado do nome
  description?: string;
  color?: string; // "" = default do backend
}

// GET /api/v1/blog/admin/posts/{id}/acks — blog:manage.
export interface BlogAckSummary {
  confirmed: number;
  acks: { user_subject: string; user_name: string; acked_at: string }[];
}

// GET/POST /api/v1/blog/posts/{slug}/reactions.
export interface BlogReactionSummary {
  counts: Record<string, number>;
  mine: string[];
}

// GET/POST /api/v1/blog/posts/{slug}/comments.
export interface BlogComment {
  id: string;
  author_name: string;
  body: string;
  mine: boolean;
  created_at: string;
}

// GET /api/v1/blog/admin/posts/{id}/views — blog:manage (Onda 5).
export interface BlogViewStats {
  unique_readers: number;
  total_views: number;
}

// Resposta de POST /api/v1/blog/admin/digest/run — blog:manage (Onda 5).
export interface BlogDigestResult {
  days: number;
  since: string;
  posts: { slug: string; title: string; summary: string; author_name: string; published_at: string }[];
  dispatched: boolean;
}

// POST /api/v1/blog/admin/posts/cover — URL pré-assinada de upload da capa.
export interface BlogCoverUploadTicket {
  object_key: string;
  upload_url: string;
  upload_fields: Record<string, string>;
  expires_in_seconds: number;
}

// POST /api/v1/blog/admin/posts/content-image — URL pré-assinada de
// upload de uma imagem inserida no CORPO do post (não a capa). Ao
// contrário de BlogCoverUploadTicket, inclui public_url — a URL estável
// (nunca expira) a embutir direto no markdown salvo, já que o objeto vai
// pra um prefixo de leitura pública (blog/content/*).
export interface BlogContentImageUploadTicket {
  object_key: string;
  upload_url: string;
  upload_fields: Record<string, string>;
  expires_in_seconds: number;
  public_url: string;
}

// GET /api/v1/blog/admin/posts/{id}/revisions — snapshot do conteúdo
// ANTES de cada edição, mais recente primeiro.
export interface BlogPostRevision {
  id: string;
  title: string;
  summary: string;
  body: string;
  tags: string[];
  edited_by_name: string;
  created_at: string;
}

// --- Catalog (central de serviços do site público, pivô Aurora v3) -----
// Espelha application.View do backend (internal/modules/catalog). A
// leitura pública (Server Component, sem sessão) usa lib/catalog/types.ts
// à parte — este arquivo é só o que a UI administrativa (catalog:manage,
// via apiClient/BFF) precisa, mesma separação que o Blog já tinha entre
// lib/blog/types.ts (público) e as interfaces Blog* aqui.
export type CatalogServiceStatus = "draft" | "published";

export interface CatalogResponsibility {
  side: "provider" | "client";
  text: string;
}

export interface CatalogStep {
  number: string;
  title: string;
  description: string;
}

export interface CatalogFeatureGroup {
  title: string;
  items: string[];
}

export interface CatalogFAQ {
  question: string;
  answer: string;
}

export interface CatalogGalleryImage {
  url: string;
  alt: string;
  caption: string;
  // Achado de auditoria: só vem preenchido nas respostas ADMINISTRATIVAS
  // (GET .../admin/services e .../admin/services/{id}) — a leitura
  // pública nunca expõe a chave crua do objeto, só a URL já resolvida.
  // Precisa estar aqui pro formulário de edição poder reenviar a galeria
  // existente sem perder a referência de cada imagem já salva.
  object_key?: string;
}

// GET /api/v1/catalog/admin/services e /admin/services/{id} (catalog:manage).
export interface CatalogService {
  id: string;
  slug: string;
  name: string;
  summary: string;
  category: string;
  icon: string;
  tagline: string;
  description: string;
  hero_image_url?: string;
  hero_image_alt?: string;
  meta_title?: string;
  meta_description?: string;
  responsibilities: CatalogResponsibility[];
  steps: CatalogStep[];
  feature_groups: CatalogFeatureGroup[];
  faqs: CatalogFAQ[];
  gallery: CatalogGalleryImage[];
  status: CatalogServiceStatus;
  display_order: number;
  published_at?: string;
  created_at: string;
  updated_at: string;
}

// Corpo de criação/edição (POST/PUT /api/v1/catalog/admin/services).
// gallery aqui usa object_key (chave crua do MinIO), diferente de
// CatalogGalleryImage.url (já resolvida) — é o que o backend espera de
// volta ao salvar.
export interface CatalogServiceInput {
  name: string;
  slug?: string;
  summary?: string;
  category?: string;
  icon?: string;
  tagline?: string;
  description?: string;
  meta_title?: string;
  meta_description?: string;
  // ausente = manter a imagem atual (PUT); "" = remover; senão define.
  hero_image_key?: string;
  hero_image_alt?: string;
  display_order?: number;
  responsibilities?: CatalogResponsibility[];
  steps?: CatalogStep[];
  feature_groups?: CatalogFeatureGroup[];
  faqs?: CatalogFAQ[];
  gallery?: { object_key: string; alt: string; caption: string }[];
}

// POST /api/v1/catalog/admin/services/images — mesmo shape de
// BlogCoverUploadTicket (mesmo storage.Ticket do backend).
export interface CatalogImageUploadTicket {
  object_key: string;
  upload_url: string;
  upload_fields: Record<string, string>;
  expires_in_seconds: number;
}

// --- Contact (formulário de contato público, pivô Aurora v3) -----------
// GET /api/v1/contact/admin/messages (contact:read).
export interface ContactMessage {
  id: string;
  name: string;
  email: string;
  phone?: string;
  service_slug?: string;
  message: string;
  ip_address?: string;
  created_at: string;
}

// --- Auditoria -----------------------------------------------------
// GET /api/v1/audit/logs (audit:read) — trilha imutável, paginada por cursor.
export interface AuditLogEntry {
  id: string;
  user_id?: string;
  action: string;
  resource_type?: string;
  resource_id?: string;
  metadata: unknown;
  correlation_id?: string;
  ip_address?: string;
  created_at: string;
}

// GET /api/v1/admin/feature-flags (restrito a aurora-admin) — ver docs/openapi.yaml.
// title/locked/depends_on espelham modules.FeatureState (registry.go) —
// antes ausentes daqui, o backend já mandava os três desde sempre, mas
// FeatureFlagsPanel.tsx não tinha como enxergá-los pelo tipo (achado de
// auditoria: toggle de módulo locked/com dependente não avisava nada).
export interface FeatureFlag {
  key: string;
  title?: string;
  enabled: boolean;
  locked?: boolean;
  description?: string;
  depends_on?: string[];
}

// GET/PUT /api/v1/admin/keycloak (restrito a aurora-admin) — ver docs/openapi.yaml.
// "source" indica de onde vieram os valores: "database" (já salvo pelo
// menu Configurações > Keycloak — o que está em uso agora), "environment"
// (nunca foi salvo por esta tela; vem de variável de ambiente do
// processo) ou "unset". Nunca inclui um client secret em texto plano.
export interface KeycloakSettingsStatus {
  source: "database" | "environment" | "unset";
  issuer_url: string;
  realm: string;
  client_id: string;
  client_secret_set: boolean;
  audience: string;
  frontend_client_id: string;
  frontend_client_secret_set: boolean;
  updated_at?: string;
  updated_by?: string;
}

// POST /api/v1/admin/keycloak/test — resultado de um teste de conexão,
// nunca persiste nada.
export interface KeycloakTestResult {
  status: "ok" | "warning" | "failed";
  discovery_ok: boolean;
  discovery_message: string;
  credentials_checked: boolean;
  credentials_ok: boolean;
  credentials_message: string;
}

// --- Signum (assinatura eletrônica) — ver docs/ROADMAP_INTRANET_ENTERPRISE.md
// §3.7. Nunca dono do conteúdo assinado — o hash de integridade é o que
// o widget confirma, nunca o documento em si.
export type SignumEnvelopeStatus = "pendente" | "assinado" | "recusado" | "cancelado";
export type SignumSignatureStatus = "pendente" | "assinado" | "recusado";
export type SignumMode = "sequencial" | "paralelo";

export interface SignumSignature {
  id: string;
  signatario_user_id: string;
  signatario_nome?: string;
  cargo_no_momento?: string;
  ordem: number;
  status: SignumSignatureStatus;
  assinado_em?: string;
  // Computado pelo backend (comparado contra o viewer autenticado) —
  // nunca reimplementar essa regra no cliente, mesma razão de
  // FilesFolder.can_delete. Ausente/false na resposta pública do
  // verificador.
  is_me?: boolean;
}

export interface SignumEnvelope {
  id: string;
  owner_module: string;
  owner_resource_type: string;
  owner_resource_id: string;
  conteudo_hash: string;
  modo: SignumMode;
  status: SignumEnvelopeStatus;
  signatures: SignumSignature[];
  created_at: string;
}

// GET /api/v1/signum/verificar/{codigo} — rota pública (§3.7.3), nunca
// expõe o conteúdo do documento.
export interface SignumVerificationResult {
  status: SignumEnvelopeStatus;
  conteudo_hash: string;
  modo: SignumMode;
  signatures: SignumSignature[];
  created_at: string;
}

// --- Trâmite (processos administrativos numerados) — sub-fases A, B e
// C, ver docs/ROADMAP_INTRANET_ENTERPRISE.md §3.6.6. Sigilo com ACL
// granular (sub-fase D) ainda não existe.
export type TramiteNivelAcesso = "publico" | "restrito" | "sigiloso";
export type TramiteProcessoStatus = "aberto" | "tramitando" | "concluido" | "arquivado";
export type TramiteDocumentoStatus = "minuta" | "pendente_assinatura" | "assinado" | "cancelado";
// 'recebido' e 'despacho' reservados no enum do backend, sem caso de uso
// ainda (ver a nota de corte da sub-fase C no roadmap).
export type TramiteMovimentacaoAcao = "tramitado" | "recebido" | "despacho" | "concluido" | "reaberto" | "arquivado";

export interface TramiteTipoDocumento {
  id: string;
  nome: string;
  exige_assinatura: boolean;
  numeracao_propria: boolean;
  ativo: boolean;
  created_at: string;
}

// CanManage é computado no backend (autor OU tramite:manage) — nunca
// reimplementar essa regra no cliente.
export interface TramiteProcesso {
  id: string;
  numero: string;
  ano: number;
  assunto: string;
  unidade_origem?: string;
  // unidade_atual — a unidade responsável pelo processo NESTE MOMENTO
  // (sub-fase C); igual a unidade_origem até a primeira tramitação.
  unidade_atual?: string;
  nivel_acesso: TramiteNivelAcesso;
  status: TramiteProcessoStatus;
  created_by_name: string;
  can_manage: boolean;
  created_at: string;
  closed_at?: string;
}

export interface TramiteProcessoListResponse {
  items: TramiteProcesso[];
  meta: { page: number; page_size: number; total_items: number; total_pages: number };
}

// TramiteMovimentacao — linha do histórico IMUTÁVEL de um processo
// (append-only garantido por trigger no Postgres, não só convenção).
export interface TramiteMovimentacao {
  id: string;
  unidade_origem?: string;
  unidade_destino?: string;
  usuario_nome: string;
  acao: TramiteMovimentacaoAcao;
  despacho_texto?: string;
  created_at: string;
}

// TramiteAcessoSigiloso — uma concessão de acesso ao CONTEÚDO de um
// processo sigiloso (§3.6.5). Sem uma linha aqui, nem admin
// (tramite:manage) vê os documentos de um processo sigiloso — só a
// metadata (número, assunto, status).
export interface TramiteAcessoSigiloso {
  id: string;
  user_id: string;
  username: string;
  concedido_por_nome: string;
  concedido_em: string;
}

export interface TramiteDocumento {
  id: string;
  processo_id: string;
  tipo_documento_id: string;
  numero_sequencial: number;
  titulo: string;
  conteudo?: string;
  has_storage_key: boolean;
  content_type?: string;
  size_bytes?: number;
  nivel_acesso: TramiteNivelAcesso;
  status: TramiteDocumentoStatus;
  // signum_envelope_id só existe a partir de pendente_assinatura — usado
  // pra montar <SignatureWidget envelopeId={...} /> (componente do
  // Signum, embutido — ver §3.7.4).
  signum_envelope_id?: string;
  created_by_name: string;
  can_manage: boolean;
  created_at: string;
}

// ============================================================
// Tipos do Módulo de Organização (Localidades, Setores, Perfis RBAC)
// ============================================================

export interface Setor {
  id: string;
  localidade_id: string;
  nome: string;
  slug: string;
  tipo: "TECNICO" | "RECEPCAO" | "GERENCIA" | "ADMINISTRATIVO";
  grupo_ad: string;
  descricao: string;
  ativo: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface Localidade {
  id: string;
  nome: string;
  slug: string;
  tipo: "CRAS" | "CREAS" | "CENTRO_POP" | "CASA_MULHER" | "CASA_ABRIGO" | "CONSELHO_TUTELAR" | "GESTAO_SEMPRAS" | "OUTROS";
  grupo_ad: string;
  endereco: string;
  telefone: string;
  bairro: string;
  ativo: boolean;
  setores?: Setor[];
  created_at?: string;
  updated_at?: string;
}

export interface Perfil {
  id: string;
  nome: string;
  slug: string;
  descricao: string;
  grupo_ad: string;
  nivel: "admin" | "gestao" | "gerencia" | "tecnico" | "recepcao";
  permissoes: string[];
  ativo: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface UsuarioLotacao {
  user_id: string;
  username: string;
  display_name: string;
  perfis: Perfil[];
  localidades: Localidade[];
  setores: Setor[];
}

export interface UserAccessSummary {
  is_admin: boolean;
  perfis: Perfil[];
  localidades: Localidade[];
  setores: Setor[];
  allowed_slugs: string[];
  active_unit: string;
  can_change_unit: boolean;
}
