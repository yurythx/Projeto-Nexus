# Registros de Decisão de Arquitetura (ADR)

Toda mudança de contrato de API, RBAC, auditoria ou middleware exige um ADR
(ver checklist no template de PR).

| ADR | Decisão | Status |
|---|---|---|
| [002](002-enterprise-resilience-and-governance.md) | Resiliência: idempotência, circuit breaker, DLQ | Aceito (contexto histórico: parte dos módulos citados foi substituída pelo Microkernel) |
| [003](003-local-auth-rsa-hardening.md) | Login local RS256 e bloqueio de conta | Aceito |
| [004](004-outbox-listen-notify-e-rbac-modular.md) | Outbox por LISTEN/NOTIFY e RBAC por módulo | Aceito (o exemplo `demands` é histórico; o RBAC modular vive hoje nos Manifests do Kernel) |
| [005](005-roadmap-conformidade-governamental.md) | Roadmap de conformidade SGD/MGI | Em execução |
| [006](006-rfc7807-problem-details.md) | Erros RFC 7807 por negociação de conteúdo | Aceito |
| [007](007-excecao-csp-style-src-vlibras.md) | Exceção de CSP `style-src` (VLibras) | Aceito, revisão trimestral |
| [008](008-kernel-grafo-de-modulos-e-estrategia-de-testes.md) | Grafo de módulos, invariantes do IAM, proveniência da auditoria, LGPD federado e estratégia de testes | Aceito |
| [009](009-revisao-dos-plugins-e-meta-de-cobertura.md) | Revisão das regras de negócio de cada plug-in e meta de 100% de cobertura por módulo no CI | Aceito |
| [010](010-revisao-da-plataforma-e-cobertura-por-pacote.md) | Revisão de cada pacote da plataforma e da composição; RabbitMQ e MinIO reais no CI; meta de 100% por pacote | Aceito |
