# 010 — Revisão da plataforma e meta de 100% de cobertura por pacote

- **Status:** Aceito
- **Data:** 2026-09-26
- **Escopo:** todos os pacotes de `internal/platform/*`, a composição (`internal/app`), `pkg/httputil`, o CI

## Contexto

O ADR 009 levou cada plug-in a 100% de cobertura. O núcleo em que todos eles
se apoiam não tinha a mesma exigência. Esse núcleo é:

- o Kernel;
- o outbox e a mensageria;
- a LGPD;
- a autenticação;
- os limitadores;
- o armazenamento;
- a composição de dependências.

Parte dele nunca rodava contra a infraestrutura de verdade. Um erro ali
atinge todos os módulos de uma vez.

Aplicamos a este núcleo o mesmo método usado nos plug-ins. Para cada pacote:

1. ler a responsabilidade dele;
2. confrontá-la com o código;
3. corrigir o que não se sustentava;
4. cobrir cada caminho com teste, inclusive as falhas de banco, broker e armazenamento.

## Decisão

### 10.1. Correções por pacote

| Pacote | Principais correções |
|---|---|
| **kernel** | O backoff do `LISTEN` volta ao mínimo quando a conexão se restabelece, e o erro é logado.<br>O supervisor zera o backoff depois de uma execução longa. |
| **outbox** | Cada evento tem o seu próprio backoff (`next_attempt_at`, migration 000123) e há um teto de 20 tentativas.<br>O lote para na primeira falha de publicação e é abortado se não conseguir registrar o status.<br>Nova rota `POST /monitoring/outbox/requeue` para reprocessar eventos com falha. Ela exige a permissão nova `monitoring:manage`, é auditada e aparece como botão no painel de Monitoramento. |
| **lgpd** | A capacidade `PersonalData` passa a ser implementada pelos plug-ins que guardam dados pessoais (hoje, o Diretório).<br>O pacote de exportação traz os dados de cada módulo; se algum falha, a resposta é 500, nunca um pacote incompleto.<br>A eliminação anonimiza os dados dos módulos na mesma transação e tenta de novo até 5 vezes (migration 000124). |
| **messaging** | O retry de uma mensagem volta só para a fila que falhou. Antes, ele era republicado no exchange e reentregue a todas as filas inscritas.<br>Corrigida uma corrida na reconexão. Os intervalos de backoff passam a ser fixados por conexão.<br>Uma `RABBITMQ_URL` malformada agora falha na hora, sem esperar os 45 s de tolerância do boot. |
| **storage** | `Get` agora devolve `ErrObjectNotFound` na hora para um objeto inexistente, em vez de o erro aparecer no meio da leitura.<br>Uma retenção WORM zero ou negativa é recusada. |
| **idempotency** | Uma chave abandonada (a requisição morreu no meio) pode ser retomada depois de 5 minutos.<br>A query string entra no hash da requisição. |
| **localauth** | O teto do bloqueio progressivo deixa de depender do formato de texto de `time.Duration`. |
| **netguard** | Passam a ser bloqueadas as formas IPv6 que embutem um IPv4 privado: IPv4-compatible, NAT64 local, Teredo e 6to4. |
| **transparency** | O XML dos dados abertos é escapado.<br>Uma falha de contagem responde 500 em vez de publicar um zero falso.<br>`?dias` fora de 1–365 responde 400. |
| **keycloakconfig** | O segredo salvo nunca é enviado a outro issuer, nem no teste de conexão nem ao salvar. |
| **branding** | O logo aceita só `https://` ou um caminho local de fato (`//host` e `/\host` são recusados).<br>Tokens de cor com contraste abaixo do WCAG AA (4,5:1) são recusados, conforme exige o e-MAG. |
| **ws** | Derrubar um tópico avisa os clientes (`topic.dropped`), e o Mercúrio mostra o aviso na tela.<br>O intervalo de ping/pong passa a ser configurável. |
| **logging** | A máscara de PII também vale para erros e `fmt.Stringer`. |
| **metrics** | Os gauges do pool do Postgres são registrados uma única vez. Um segundo registro só troca o pool observado, em vez de entrar em pânico. |
| **telemetry** | Uma falha ao criar o exporter OTLP desliga o tracing e registra o erro, sem derrubar a API. |
| **httputil** | `Query` agora corta por caractere, não por byte. Cortar no meio de um "ç" gerava UTF-8 inválido, que o Postgres rejeitava com um 500. |
| **app** | `NewDependencies` foi separado em `config.Load` e `build`. A declaração da topologia do RabbitMQ virou uma função testável.<br>`supervised` fixa o seu backoff quando é criado.<br>O servidor de métricas do worker espera o listener encerrar. |

### 10.2. Código removido

Removido o código sem uso em produção, que só inflava a superfície:

- o pacote `jobs`, morto desde o Microkernel;
- o limitador de taxa em Postgres;
- o circuit breaker (`resilience`) e as métricas sem uso, junto com a dependência `gobreaker`;
- o `InMemoryLimiter` do httpserver, usado só em teste.

### 10.3. Infraestrutura de teste real

O CI roda, além do Postgres:

- **RabbitMQ** (serviço do job), para confirms, retry, DLQ e reconexão, e para o Relay do outbox contra o broker de verdade;
- **MinIO** (subido num passo do job), para operações de objeto, URLs pré-assinadas e object-lock/WORM.

As variáveis de ambiente são `TEST_RABBITMQ_URL`, `TEST_MINIO_ENDPOINT`,
`TEST_MINIO_ACCESS_KEY` e `TEST_MINIO_SECRET_KEY`. Sem elas, os testes
correspondentes são pulados.

`internal/app` monta as dependências reais, via `build`, contra essa
infraestrutura e prova duas coisas:

- cada dependência obrigatória falha rápido, com a etapa no erro: Postgres, Redis, assinador, OIDC, cifra, proxies, RabbitMQ, MinIO, Kernel e topologia;
- as opcionais só geram aviso: o bucket do MinIO e a configuração persistida do Keycloak.

### 10.4. Meta de cobertura

`scripts/coverage-gate.sh` passa a exigir 100% em:

- cada plug-in de `internal/modules/*`, como no ADR 009;
- cada pacote de `internal/platform/*`, menos os dublês de teste `dbtest` e `storagetest`;
- `internal/app`;
- cada pacote de `pkg/*`.

O perfil vem de `go test -race -p 1 -coverpkg=./internal/...,./pkg/...`.

## Consequências

- Uma regressão no núcleo agora quebra o CI com a linha descoberta, como já acontecia com os plug-ins.
- As costuras de teste são campos ou variáveis de pacote documentados: intervalos, `begin` e `subscribe` injetáveis, `newExporter`, `lookupNetIP`. Nenhuma muda o comportamento em produção.
- Os testes de infraestrutura dependem de serviços que o CI provê. Localmente eles são pulados, a menos que as variáveis estejam definidas (ver `GUIDE_DESENVOLVEDOR.md`).
