# 013 — Permissão com escopo e herança

- **Status:** Aceito (fases 1 e 2 em implementação; fase 3 depois)
- **Data:** 2026-09-27
- **Escopo:** `internal/platform/auth`, `internal/platform/iam`, os módulos
  com permissões de gestão, o frontend (hooks de permissão e formulários de
  dono)

## Contexto

Hoje a identidade carrega as permissões **somadas** de todos os perfis e,
separada, a lista de escopos. `auth.HasPermission(perm)` ignora o escopo.
Por isso só o Trâmite, os Arquivos e o Mercúrio restringem por lotação
([REGRAS_DE_NEGOCIO.md §5](../REGRAS_DE_NEGOCIO.md#5-onde-o-escopo-vale-e-onde-não-vale)).
Lotar alguém como gestor de conteúdo "na UPA" libera Blog e Catálogo para a
prefeitura inteira.

As plataformas de referência avaliam acesso como **(quem, ação,
recurso)**, com o recurso posicionado na organização e a concessão valendo
do escopo para baixo: Google Cloud, Azure, SAP e Salesforce.

## Proposta

### 1. Concessões em vez de soma

A identidade passa a carregar **concessões**: cada uma tem perfil,
permissões e escopo. O resolvedor já lê isso linha a linha e hoje descarta
a associação. `Permissions`, a soma, continua existindo só para exibição e
menus.

### 2. Cobertura (herança para baixo)

| Concessão | Cobre recursos… |
|---|---|
| sem escopo (global) | de toda a plataforma, inclusive os institucionais (sem dono) |
| entidade E | de E e de todas as unidades e departamentos de E |
| unidade U | de U, das **subunidades** de U (árvore `parent_id`) e dos departamentos delas |
| departamento D | de D |

A árvore de subunidades é expandida no resolvedor e fica no cache junto
com a identidade, então checar não custa consulta.

### 3. API de autorização

- `auth.Can(identidade, perm, alvo)`: o alvo é a posição do recurso
  (entidade, unidade, departamento). Vale se alguma concessão com `perm`
  cobre o alvo.
- `auth.CanGlobal(identidade, perm)`: só concessão sem escopo. Vale para
  recursos institucionais e para as permissões de plataforma.
- `auth.HasPermission` passa a ser "tem em algum escopo". Serve para
  mostrar o menu e para criar recursos no próprio escopo, **nunca** para
  decidir sobre um recurso existente.

### 4. Duas classes de permissão

**Plataforma (só valem com concessão global):** `modules:manage`,
`keycloak:manage`, `branding:manage`, `monitoring:*`, `egress:manage`,
`audit:verify`, `signum:manage`, `example:manage`, `iam:manage` e
`users:*`. A administração delegada do IAM fica para a fase 3.

**Com escopo (valem onde foram concedidas):**

| Permissão | Posição do recurso | O que muda no banco |
|---|---|---|
| `tramite:manage` | unidade de origem ou atual do processo | nada |
| `catalog:manage` | unidade responsável do serviço | nada (coluna existe) |
| `blog:manage` | dono do post | nova coluna de dono |
| `wiki:manage` | dono da página, herdado da página-mãe | nova coluna de dono |
| `calendar:manage` | dono da sala (moderação dos eventos nela) | nova coluna de dono |
| `mercurio:manage` | departamento da sala | nada |
| `directory:manage` | lotação exibida da pessoa | nada |
| `files:manage`, `contact:*`, `audit:read` | — | continuam globais nesta proposta (fase 3) |

**Dono na criação:**
- quem tem concessão global escolhe qualquer dono, ou nenhum, o que torna
  o recurso **institucional**;
- quem tem concessão com escopo só escolhe dentro dele, e o padrão é a
  própria unidade.

**Leitura não muda:** Blog, Wiki, Catálogo e Agenda continuam visíveis
como hoje. Segmentar a leitura por público-alvo é outra funcionalidade.

### 5. Compatibilidade

- **Conteúdo existente** fica **institucional** (sem dono): só quem tem
  concessão global o gerencia.
- **Concessões com escopo que hoje "vazam"** (gestão ou plataforma
  concedidas com escopo) passam a valer só no escopo, ou a não valer, no
  caso das de plataforma. Antes de publicar, um **relatório de impacto**
  (`make iam-scope-report`) lista cada lotação e mapeamento afetado,
  indicando quem perde o quê.
- **Ambiente de teste:** os grupos `NEXUS-*` são globais e não mudam. O
  Protocolo já é restrito por unidade.

## Fases

1. **Núcleo:** concessões na identidade (resolvedor, cache, `/me`), `Can`
   e `CanGlobal`, expansão das subunidades, permissões de plataforma só
   globais, relatório de impacto e testes da matriz de cobertura.
2. **Módulos:**
   - Trâmite (`tramite:manage` por unidade, reunindo o `InUnidade` atual
     na mesma regra);
   - Catálogo, Blog, Wiki, Agenda, Mercúrio e Diretório, com a coluna de
     dono, os formulários e testes HTTP de "gestor da unidade A não
     gerencia B, o global gerencia tudo".
3. **Depois:**
   - administração delegada do IAM (gestor que cadastra usuários e
     lotações só da sua entidade);
   - auditoria filtrada por escopo;
   - contato encaminhado por setor;
   - Arquivos com dono organizacional.

## Decisões tomadas (2026-09-27)

1. Conteúdo existente sem dono é **institucional**: só a gestão global o
   gerencia.
2. Concessão na unidade **cobre as subunidades**.
3. Administração delegada do IAM fica na **fase 3**.
4. A **leitura continua aberta**. O escopo restringe só a gestão.
