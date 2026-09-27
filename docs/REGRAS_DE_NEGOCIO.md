# Estrutura, funcionamento e regras de negócio

Referência funcional da plataforma. Serve para quem **implanta e administra**
(como modelar a organização, quem pode o quê) e para quem **desenvolve**
(quais regras cada módulo garante). Tudo aqui reflete o código em
`backend/internal`. Como programar um módulo está em
[GUIDE_DESENVOLVEDOR.md](GUIDE_DESENVOLVEDOR.md) e
[GUIDE_MODULOS.md](GUIDE_MODULOS.md). Como implantar está em
[DEPLOY.md](DEPLOY.md).

---

## 1. Como a plataforma funciona

### 1.1. O caminho de uma requisição

1. **Login.** O login principal é pelo Keycloak (OIDC), que pode federar o
   Active Directory. Se o Keycloak ficar indisponível, há o login local de
   contingência. O token traz quem é a pessoa e os **grupos** dela.
2. **Identidade efetiva.** A cada requisição, o IAM da API:
   - cria ou atualiza o usuário na primeira vez que ele entra
     (provisionamento automático, com nome, e-mail e grupos);
   - soma as **permissões** de todos os perfis que a pessoa recebeu, seja
     por **lotação manual** ou por **mapeamento de grupo**;
   - monta a lista de **escopos** (onde a pessoa está lotada).

   O resultado fica em cache por poucos segundos. Quando um administrador
   muda perfis, lotações ou mapeamentos, o cache é invalidado em todas as
   instâncias e a mudança vale **na hora**.
3. **Módulo.** A rota exige a permissão necessária. Depois, o próprio
   módulo aplica as regras dele: unidade, sigilo, dono, ACL.
4. **Efeitos.** Toda alteração grava, na mesma transação:
   - o **registro de auditoria**, imutável e encadeado por hash;
   - o **evento de domínio** no outbox.

   O worker publica esse evento no RabbitMQ, e daí ele segue para o tempo
   real (WebSocket), para os webhooks (Egress) e para a busca.

### 1.2. Módulos (Kernel)

| Tipo | Módulos | Pode desligar? |
|---|---|---|
| Núcleo | IAM & Usuários, Auditoria | Não |
| Plug-ins | Mercúrio, Egress, Blog, Catálogo, Contato, Diretório, Agenda, Arquivos, Wiki, Busca, Signum, Trâmite, Módulo Modelo | Sim, em Configurações → Módulos (`modules:manage`) |

- Desligar um módulo tem efeito imediato, sem reiniciar:
  - as rotas dele respondem `404 MODULE_DISABLED`;
  - ele some do menu, do site público, do sitemap e da busca;
  - as assinaturas de tempo real dele caem;
  - os workers e consumidores de fila dele param.
- **Dependências:** o Trâmite depende do Signum. Não é possível desligar o
  Signum com o Trâmite ligado, nem ligar o Trâmite com o Signum desligado.
  A tela mostra esse grafo.
- **Módulos públicos** (Catálogo, Contato, Diretório, Agenda e Signum) têm
  páginas no site institucional, que funcionam sem login.

### 1.3. Garantias transversais

| Garantia | Como |
|---|---|
| Auditoria de toda alteração | Tabela só de inclusão (o banco impede alterar ou apagar), cadeia SHA-256 verificável, exportação LAI (CSV/JSON/XML) e cópia diária imutável (WORM) |
| Consentimento LGPD | Aceite registrado com versão do termo e IP (anônimo ou da conta); direitos do titular: acesso, portabilidade e pedido de exclusão |
| Limites de abuso | Por identidade na API (600/min), por IP nas rotas públicas e no login, e bloqueio progressivo de login local (ver [DEPLOY.md](DEPLOY.md#segurança-do-login-local)) |
| Idempotência | Repetir uma alteração com a mesma chave devolve a resposta original, sem duplicar |
| Dados pessoais nos logs | Mascarados (CPF, e-mail, telefone) |

---

## 2. Estrutura organizacional

```
Entidade (ex.: Prefeitura, Secretaria de Saúde, Empresa A)
└── Unidade (ex.: UPA 24h, Escola, Filial) — pode ter unidade-mãe
    └── Departamento (ex.: Administração, RH, TI)
```

### 2.1. Entidade

É o órgão ou a organização de nível mais alto: prefeitura, secretaria,
autarquia, empresa.

- Campos: nome, sigla, slug (endereço), documento (CNPJ ou outro) e ativo.
- O slug é gerado a partir do nome quando não é informado, e é único.

### 2.2. Unidade

É o local ou a subdivisão onde as pessoas trabalham: escola, posto de
saúde, filial, sede.

- Pertence a **uma** entidade e **não pode mudar de entidade** depois de
  criada, porque as lotações gravam a entidade da unidade.
- Pode ter uma **unidade-mãe**, o que forma uma hierarquia. A mãe precisa
  ser da **mesma entidade**, e ciclos são recusados.
- Campos: nome, sigla, slug (único dentro da entidade), grupo do AD de
  referência, e-mail, telefone, endereço e ativo.
- **É o nível que o Trâmite usa.** Processos são abertos, ficam e são
  tramitados entre **unidades**.

### 2.3. Departamento

É o setor dentro de uma unidade: Administração, Atendimento, RH.

- Pertence a **uma** unidade e **não pode mudar de unidade**.
- Campos: nome, sigla, slug (único dentro da unidade), grupo do AD de
  referência, e-mail, telefone e ativo.
- **Todo departamento pertence a uma unidade.** Uma organização que só tem
  departamentos (como a Prefeitura no ambiente de teste) usa uma unidade
  "Sede" para agrupá-los.

### 2.4. Regras comuns

- **Exclusão protegida:** excluir entidade, unidade ou departamento **em
  uso** responde `409`. Nada é apagado em cascata; primeiro se remove o que
  depende, depois o item.

  **Bloqueiam a exclusão:**
  - unidades de uma entidade e departamentos de uma unidade;
  - lotações e mapeamentos de grupo que apontam para o item;
  - processos do Trâmite (origem, unidade atual e histórico).

  **Não bloqueiam; a referência é apenas limpa:**
  - unidade responsável de um serviço do Catálogo;
  - lotação exibida no Diretório;
  - departamento de uma sala do Mercúrio;
  - unidade-mãe das subunidades (elas viram unidades de topo).
- **"Ativo":** hoje só controla o que aparece no **Diretório** (setores
  públicos e busca de setores). Ver os [pontos de atenção](#8-pontos-de-atenção-comportamento-atual).
- Toda alteração é auditada com o antes e o depois.
- Quem administra: permissão `iam:manage`.

---

## 3. Perfis e permissões

### 3.1. Permissões

- Formato `recurso:ação` (ex.: `tramite:create`). Há dois curingas:
  `recurso:*` (todas as ações do recurso) e `*` (tudo).
- Cada módulo declara as próprias permissões. A lista completa, com
  descrições, sai em `GET /api/v1/iam/permissions` e aparece na tela de
  perfis.

| Módulo | Permissões |
|---|---|
| IAM | `users:read`, `users:manage`, `iam:manage`, `modules:manage`, `keycloak:manage`, `branding:manage`, `monitoring:read`, `monitoring:manage` |
| Auditoria | `audit:read`, `audit:verify` |
| Trâmite | `tramite:create`, `tramite:route`, `tramite:manage` |
| Blog / Catálogo / Wiki | `blog:manage`, `catalog:manage`, `wiki:manage` |
| Contato | `contact:read`, `contact:manage` |
| Agenda / Arquivos / Diretório | `calendar:manage`, `files:manage`, `directory:manage` |
| Mercúrio / Signum / Egress | `mercurio:manage`, `signum:manage`, `egress:manage` |
| Módulo Modelo | `example:manage` |

O que **não** exige permissão, só estar autenticado:
- ler Blog e Wiki, criar e editar páginas da Wiki;
- usar o Mercúrio, a Agenda (eventos próprios), os Arquivos (pastas
  próprias e compartilhadas) e o Diretório;
- criar envelopes no Signum e usar a Busca.

### 3.2. Perfis

Um perfil é um **conjunto nomeado de permissões**. Perfis de sistema (vêm
com a plataforma):

| Perfil | Permissões | Para quem |
|---|---|---|
| `administrador` | `*` | Administração da plataforma |
| `auditor` | `audit:read`, `audit:verify`, `users:read`, `monitoring:read` | Controle interno |
| `gestor-iam` | `users:read`, `users:manage`, `iam:manage` | Quem cadastra estrutura, usuários e lotações |
| `gestor-conteudo` | `blog:manage`, `catalog:manage`, `wiki:manage`, `contact:read`, `contact:manage` | Comunicação e ouvidoria |
| `protocolo` | `tramite:create`, `tramite:route` | Quem abre e tramita processos |
| `servidor` | *(nenhuma)* | Colaborador comum, só o que não exige permissão |

Regras:
- É possível criar **perfis customizados** com qualquer combinação de
  permissões válidas. Permissão malformada é recusada.
- Perfis de sistema não são excluídos e mantêm o slug. O `administrador`
  **sempre** mantém `*` e fica ativo, para que um erro de edição não tranque
  a plataforma.
- **Desativar um perfil** retira na hora as permissões dele de todos que o
  têm. Reativar devolve.
- A role de realm `nexus-admin` no Keycloak equivale a `*`.

### 3.3. Ninguém concede o que não tem (anti-escalada)

Criar ou editar um perfil, lotar alguém ou mapear um grupo exige que
**quem concede** tenha todas as permissões envolvidas. Retirar também exige.
Exemplo: o `gestor-iam` (que não tem `audit:read`) não cria um perfil com
`audit:read` nem lota alguém num perfil com essa permissão; recebe `403`.
Só quem tem acesso total concede ou retira `nexus-admin`.

---

## 4. Lotações e mapeamentos de grupo

Há dois jeitos de alguém receber um perfil. Os dois resultam na mesma coisa:
**perfil + escopo**.

| | Lotação manual | Mapeamento de grupo |
|---|---|---|
| O que é | "Usuário X tem o perfil P no escopo E" | "Quem está no grupo G do Keycloak/AD tem o perfil P no escopo E" |
| Onde se cadastra | Configurações → Usuários → Lotações | Configurações → Mapeamento AD |
| Quando vale | Enquanto existir | Enquanto a pessoa estiver no grupo (lido do token a cada login ou renovação) |
| Uso típico | Exceções e acúmulos | Regra geral: departamento do AD → perfil + departamento |

- **Escopo:** entidade, unidade e/ou departamento. Tudo vazio significa a
  plataforma inteira. O escopo é **completado e validado**:
  - informar só o departamento preenche a unidade e a entidade;
  - informar unidade e departamento de unidades diferentes é recusado.
- Uma pessoa pode ter **várias** lotações, por exemplo trabalhar em duas
  unidades.
- O grupo é comparado sem diferenciar maiúsculas e aceita o caminho do
  Keycloak (`/SEMSA/SEMSA-UPA/SEMSA-UPA-ADM` ou `SEMSA-UPA-ADM`).
- As permissões efetivas são a **soma** de todos os perfis de todas as
  lotações e mapeamentos.
- Mudanças valem na próxima requisição da pessoa, sem novo login.

---

## 5. Onde o escopo vale (e onde não vale)

> **Regra central:** a **permissão** diz **o que** a pessoa pode fazer, e
> vale para a plataforma inteira. O **escopo** diz **onde** ela está
> lotada, e só restringe nos módulos que o consultam.

| Módulo | O escopo restringe? | Como |
|---|---|---|
| **Trâmite** | Sim | Abrir processo só em unidade onde está lotado; tramitar, juntar documento e concluir só com o processo **na sua unidade**; ver processo restrito só se a sua unidade for a de origem ou a atual |
| **Arquivos** | Sim, por ACL | Pastas compartilhadas com `unidade`, `departamento` ou `perfil` usam as lotações da pessoa |
| **Mercúrio** | Sim | Salas de departamento: quem está no grupo ou lotado no departamento |
| **Diretório** | Só exibição | Mostra a lotação da pessoa, sem restringir acesso |
| Blog, Catálogo, Wiki, Agenda, Contato, Signum, Egress, IAM, Auditoria, Módulos, Branding | **Não** | Quem tem a permissão pode usá-la em **tudo**, qualquer que seja o escopo da lotação |

Consequência prática: lotar alguém como `gestor-conteudo` **na UPA** dá
`blog:manage` e `catalog:manage` **para a plataforma inteira**, não só para
a UPA. O escopo só faz diferença para Trâmite, Arquivos e Mercúrio.

---

## 6. Como modelar uma implantação

1. **Entidades:** uma por órgão (ou empresa).
2. **Unidades:** os locais de trabalho, que no Trâmite são quem "recebe"
   processos. Use unidade-mãe se houver hierarquia.
3. **Departamentos:** os setores de cada unidade.
4. **Grupos no Keycloak/AD:** idealmente um por departamento, com nome
   único (ex.: `SEMSA-UPA-ADM`), e grupos de papel (`NEXUS-ADMINS`,
   `NEXUS-AUDITORES`…).
5. **Mapeamentos:**
   - cada grupo de departamento → `servidor` no departamento;
   - quem protocola (ex.: a Administração da unidade) → `protocolo` na
     unidade;
   - grupos de papel → perfis de gestão, sem escopo.
6. **Exceções:** lotações manuais.

O ambiente de teste segue exatamente essa receita
(`scripts/demo-data/generate.mjs`), com 4 entidades, 15 unidades e 76
departamentos.

**Várias organizações independentes** (ex.: duas empresas): a estrutura se
modela sem código, com uma entidade por empresa. Mas a plataforma **não
isola** as empresas entre si. Blog, Wiki, Catálogo, Agenda pública, a sala
"Geral" do Mercúrio, o Diretório, a identidade visual e o site são
compartilhados, e o administrador vê tudo. Para empresas independentes, use
**uma implantação por empresa**.

---

## 7. Regras por módulo

### Trâmite (processos administrativos)
- **Tipos:** Contratação, Memorando, Ofício, Requerimento, Outros. Tipo
  desativado não abre processo novo.
- **Numeração:** `NNNNNN/AAAA`, sequencial por ano.
- **Estados:** `aberto` → `em_tramitacao` → `concluido` → `arquivado`.
  Reabrir (`tramite:manage`) volta para `em_tramitacao`. Só se arquiva o
  que está **concluído**.
- **Abrir:** exige `tramite:create` e lotação na unidade de origem.
- **Tramitar:** exige `tramite:route`, o processo na sua unidade e um
  destino diferente da unidade atual.
- **Sigilo:**

  | Sigilo | Quem vê |
  |---|---|
  | público | qualquer autenticado |
  | restrito | autor, credenciados, lotados na unidade de origem ou na atual, e `tramite:manage` |
  | sigiloso | **somente** o autor e os credenciados nominalmente (nem a unidade nem o gestor) |

  Credenciar ou revogar acesso nominal só existe para processo não público
  e dá **leitura**.
- **Documentos:**
  - nascem em rascunho;
  - são editáveis só pela unidade que está com o processo;
  - anexos vão para o MinIO por URL pré-assinada;
  - **pedido de assinatura** abre um envelope no Signum, e o documento
    assinado não pode mais ser editado;
  - não se conclui processo com assinatura pendente.
- Processo concluído ou arquivado não recebe alterações. Cada movimento
  (abertura, tramitação, conclusão, arquivamento, reabertura) fica no
  histórico.

### Signum (assinatura eletrônica)
- Qualquer autenticado abre um envelope com o **hash SHA-256** do
  documento e de 1 a 50 signatários ativos, em ordem sequencial ou livre.
- **Cerimônia:**
  1. o signatário pede um **desafio** (uso único, expira em
     `SIGNUM_CHALLENGE_TTL`, padrão 5 min);
  2. confirma o hash do documento;
  3. **reautentica com a senha**: conta local contra o banco, conta do
     Keycloak via Keycloak/AD, pelo client `nexus-backend`.
- **Sequencial:** ninguém assina antes dos anteriores.
- **Estados:** `pending` → `completed` (todos assinaram) | `refused`
  (alguém recusou com motivo) | `cancelled` (dono ou `signum:manage`
  cancelou). Envelope não pendente não aceita ação.
- Tentativas de senha erradas são limitadas por usuário.
- **Verificação pública** pelo id do envelope, sem login.
- Veem o envelope: o dono, os signatários e `signum:manage`.

### Arquivos
- Pastas em árvore e arquivos no MinIO. Upload e download são feitos
  **direto** pelo navegador, por URL pré-assinada (15 min); o limite é
  `UPLOAD_MAX_FILE_MB` (100 MB).
- **Acesso a uma pasta:**
  - `files:manage` tem acesso total;
  - o **dono** da pasta, ou de qualquer pasta acima dela, tem acesso total;
  - a **ACL** de qualquer nível acima (herança) dá leitura, e escrita se
    `can_write`.
- Sujeitos da ACL: `everyone`, `user`, `perfil`, `ad_group`, `unidade`,
  `departamento`.
- Alterar a ACL, renomear ou excluir a pasta exige ser dono ou ter
  `files:manage`.
- Nomes são únicos dentro da pasta. Pasta com conteúdo só é excluída com
  confirmação recursiva. Não se move uma pasta para dentro de si mesma.

### Mercúrio (chat)
- **Salas:**
  - **global** (todos);
  - **de departamento**: acesso por grupo do AD ou lotação no
    departamento, e `mercurio:manage`;
  - **direta** (duas pessoas; não se abre conversa consigo mesmo).
- Criar ou arquivar sala e moderar mensagens: `mercurio:manage`. Sala
  arquivada é só leitura.
- Mensagens de 1 a 4.000 caracteres. O autor **edita nos primeiros 15
  minutos** e pode excluir. Tudo chega em **tempo real**.
- Não lidas por sala; "marcar como lida" zera o contador.

### Agenda
- Eventos com visibilidade:
  - **público:** aparece no site;
  - **interno:** todos os autenticados;
  - **privado:** só o organizador.
- Duração de até 31 dias.
- Salas (`calendar:manage`) com capacidade e recursos. **Sem conflito:**
  uma sala não aceita duas reservas no mesmo horário. Sala inativa não
  aceita reservas, e sala com reservas futuras não é excluída (desative).
- Só o organizador ou `calendar:manage` altera. Evento cancelado não se
  altera, cria-se outro. Cancelar libera a sala.

### Blog
- `blog:manage` cria, edita, publica, arquiva e exclui.
- **Estados:** `draft` → `published` → `archived`, com volta a rascunho.
  Não se publica sem texto.
- Leitores veem só o publicado. Rascunho e arquivado ficam só para
  gestores. Slug único. Tipos: notícia ou comunicado.

### Wiki
- Qualquer autenticado cria e edita. Excluir exige `wiki:manage`.
- **Controle de concorrência:** a edição informa a versão aberta. Se outra
  pessoa salvou antes, a edição é recusada (`409`) em vez de sobrescrever.
- Histórico de revisões com restauração. Árvore de páginas: não se exclui
  página com subpáginas, e uma página não fica dentro de si mesma.

### Catálogo de Serviços (Carta de Serviços)
- `catalog:manage` cria e publica.
- **Publicar exige** resumo e ao menos um **canal de atendimento** (Lei
  13.460/2017).
- Publicado aparece no site sem login. Publicado **não é excluído**
  (arquive antes). A unidade responsável precisa existir.

### Contato (ouvidoria)
- Formulário público sem login.
- Exige **consentimento LGPD**. Há um campo-armadilha contra robôs e um
  limite de **5 mensagens por hora por IP**.
- Ler: `contact:read` (contém dados pessoais). Triar: `contact:manage`.
- Estados: `new`, `in_progress`, `answered`, `archived`. Pode haver um
  responsável, que precisa ser um usuário ativo.

### Diretório
- Pessoas e setores. Os dados de identidade (nome, e-mail, grupos) vêm do
  Keycloak/AD.
- Cada pessoa edita o próprio cargo, telefone, ramal e bio. A **lotação
  exibida** e o perfil de terceiros exigem `directory:manage`.
- **Setores públicos** no site listam só entidades, unidades e
  departamentos **ativos**.

### Busca Global
- Busca em Blog, Wiki, Arquivos, Catálogo, Diretório e Trâmite, nos
  módulos ativos, **respeitando as regras de cada um**: rascunho não
  aparece, a ACL das pastas vale, o sigilo do processo vale.
- Mínimo de 2 caracteres. Acha por **prefixo** ("vacina" acha
  "vacinação") e aceita aspas, `-termo` e `or`.

### Egress (webhooks de saída)
- `egress:manage` cadastra destinos:
  - tipos webhook, n8n, Zabbix, Grafana;
  - padrões de evento (ex.: `tramite.*`);
  - segredo HMAC opcional, guardado cifrado e nunca devolvido.
- **Anti-SSRF:** destinos em rede interna, loopback, endereço de metadados
  de nuvem ou nome de container são recusados. Em produção não dá para
  desligar a proteção.
- Entregas assíncronas, com **retentativa e espera crescente** até
  `EGRESS_MAX_ATTEMPTS` (8). Há reenvio manual e teste de conexão.

### Auditoria
- `audit:read` consulta (com filtros) e exporta (LAI: CSV, JSON, XML).
  `audit:verify` confere a cadeia de hash. A própria exportação é auditada.

### LGPD
- **Aceite do termo:** anônimo (identificador do navegador) ou da conta,
  com versão e IP.
- **Direitos do titular:** "meus dados" (acesso e portabilidade, incluindo
  os dados pessoais dos módulos) e pedido de exclusão, processado com
  retentativas.

### Transparência
- Endpoints públicos, sem login: dados da plataforma, módulos ativos,
  conjuntos de dados e estatística de ações da auditoria.

---

## 8. Pontos de atenção (comportamento atual)

Comportamentos do código hoje que merecem decisão de quem administra:

1. **O escopo não limita permissões fora de Trâmite, Arquivos e
   Mercúrio** (seção 5). Perfis de gestão com escopo valem para a
   plataforma toda.
2. **"Ativo" em entidade, unidade e departamento só afeta o Diretório.**
   Lotações numa unidade desativada **continuam dando acesso**, e o
   Trâmite **aceita** abrir processo e tramitar para uma unidade
   desativada. Recomendação: recusar unidade ou entidade inativa como
   origem ou destino no Trâmite e ignorar lotações nelas no IAM.
3. **Processo sigiloso não é visto nem por `tramite:manage`.** É
   intencional (só autor e credenciados), mas quem administra precisa
   saber.
4. **Reabrir processo exige `tramite:manage`,** que nenhum perfil de
   sistema tem além do administrador. O Protocolo conclui e arquiva, mas
   não reabre.
5. **Sem isolamento entre entidades** (seção 6): uma implantação serve a
   uma organização (que pode ter várias entidades), não a várias
   organizações independentes.
6. **Excluir uma unidade-mãe não é bloqueado pelas subunidades:** elas
   perdem a mãe e viram unidades de topo, sem aviso. Se a hierarquia
   importa, reatribua as subunidades antes.
