# Personas Corporativas de Teste — Projeto Nexus (Rondonópolis - MT)

> **Ambiente:** Servidor de Teste (`https://192.168.1.42`)  
> **Método de Autenticação:** Login Local (NextAuth Credentials via `POST /api/v1/auth/login`)  
> **Senha Padrão de Todas as Personas:** `Nexus@Rondonopolis2026!`

---

## 1. Tabela de Personas e Lotações

Todas as personas estão ativas, com hashes criptográficos **Argon2id**, perfis cadastrados no Diretório e permissões resolvidas conforme o [ADR 013 (Permissão com Escopo e Herança)](adr/013-permissao-com-escopo-e-heranca.md) e [ADR 030](adr/030-blog-autoria-dedicada-publico-alvo-e-estrutura-rondonopolis.md):

| Usuário | Nome Completo | Cargo & Função | Unidade (Rondonópolis) | Departamento | Perfil IAM | Papel no Sistema |
|---|---|---|---|---|---|---|
| `admin` | Administrador Local | Administrador da Plataforma & TI | Sede Centro (`SEDE-CENTRO`) | DEV | `administrador` | Acesso irrestrito a configurações e auditoria. |
| `carlos.diretoria` | Carlos Mendes | Diretor Executivo & Relações Institucionais | Sede Centro (`SEDE-CENTRO`) | Diretoria Executiva (`DIR-EXEC`) | `administrador` + `gestor-iam` | Liderança institucional e gestão de identidades. |
| `mariana.mkt` | Mariana Silva | Coordenadora de Comunicação & Marketing | Sede Centro (`SEDE-CENTRO`) | Marketing (`MKT`) | `gestor-conteudo` | Autora de publicações no Blog (`blog:manage`) e Wiki. |
| `lucas.tech` | Lucas Albuquerque | Tech Lead & Engenheiro de Software P&D | Hub de Inovação (`HUB-AURORA`) | Engenharia / DEV (`DEV`) | `servidor` + `gestor-conteudo` (local) | Redator técnico e receptor de comunicados do Hub. |
| `marcos.ops` | Marcos Roberto Souza | Supervisor de Operações & Logística | Complexo Industrial (`POLO-IND`) | Operações (`OPS`) | `servidor` + `protocolo` | Operações de fábrica/armazém e trâmite logístico. |
| `ana.protocolo` | Ana Paula Ferreira | Analista de Protocolo Central & Atendimento | CIA Sagrada Família (`UNID-POGUBA`) | Protocolo Central (`PROT`) | `protocolo` | Triagem de mensagens públicas e abertura de processos. |
| `rodrigo.agro` | Rodrigo Camargo | Supervisor Operacional Intermodal | Terminal Ferroviário (`TERM-AGRO`) | Operações Intermodais (`INTERMODAL`) | `servidor` | Operador do anel viário e leitor institucional. |

---

## 2. Cenários Práticos de Teste e Validação

### Cenário 1: Validação do Público-Alvo no Blog (ADR 014 / ADR 030)
1. Faça login como **`mariana.mkt`** (senha: `Nexus@Rondonopolis2026!`).
2. Acesse `/gestao/blog/novo`.
3. Crie uma matéria técnica (ex.: *"Novo Lab de IoT & Sensores no Hub Aurora"*).
4. No seletor de **Público-Alvo**, expanda o componente e marque exclusivamente a unidade **Hub de Inovação & P&D - Vila Aurora**. Salve e publique.
5. Em outra janela anônima, faça login como **`lucas.tech`** e acesse `/blog`: o artigo estará visível no feed.
6. Em seguida, faça login como **`marcos.ops`** e acesse `/blog`: confirme que o artigo **não aparece** no feed nem na busca, respeitando o isolamento do público-alvo.

---

### Cenário 2: Validação do Diretório Telefônico e Setores (`/diretorio`)
1. Acesse `/diretorio` com qualquer um dos usuários.
2. Na aba **Pessoas**, teste os filtros por unidade:
   - Filtre por **Vila Aurora**: aparece *Lucas Albuquerque* com telefone `(66) 3422-5500` e ramal `5508`.
   - Filtre por **Distrito Industrial**: aparece *Marcos Roberto Souza* com telefone `(66) 3439-8800` e ramal `8813`.
   - Filtre por **Sagrada Família**: aparece *Ana Paula Ferreira* com telefone `(66) 3425-4100` e ramal `4118`.
3. Na aba **Setores**, confira a visualização das 5 unidades e 19 departamentos organizados por entidade.

---

### Cenário 3: Trâmite de Processos entre Unidades de Rondonópolis (`/tramite`)

Os 4 processos eletrônicos abaixo já estão ativos e circulando no sistema com documentos juntados e histórico auditável:

| Processo | Tipo | Origem | Lotação Atual | Assunto | Documentos / Autores |
|---|---|---|---|---|---|
| **`000001/2026`** | Contratação | Hub Aurora (`HUB-AURORA`) | Sede Centro (`SEDE-CENTRO`) | Expansão de Infraestrutura de Servidores para o Laboratório AgroTech | • *Memorando nº 12/2026* (Lucas Albuquerque)<br>• *Despacho Decisório nº 44/2026* (Carlos Mendes) |
| **`000002/2026`** | Requerimento | Polo Industrial (`POLO-IND`) | Terminal Agro (`TERM-AGRO`) | Revisão e Manutenção Preventiva da Frota Bitrem (Safra 2026/2027) | • *Laudo de Vistoria nº 08/2026* (Marcos Souza)<br>• *Parecer Balança nº 19/2026* (Rodrigo Camargo) |
| **`000003/2026`** | Requerimento | CIA Poguba (`UNID-POGUBA`) | Hub Aurora (`HUB-AURORA`) | Cadastro e Integração de Cooperativa Agropecuária no Portal Nexus | • *Formulário de Triagem* (Ana Paula Ferreira) |
| **`000004/2026`** | Memorando | Sede Centro (`SEDE-CENTRO`) | Sede Centro (`SEDE-CENTRO`) | Aprovação de Campanha Institucional e Stand da Nexus na Exposul 2026 | • *Plano de Mídia Exposul* (Mariana Silva)<br>• *Homologação Presidência* (Carlos Mendes) — **Concluído** |

#### Como validar no navegador:
1. Faça login como **`lucas.tech`** e acesse `/tramite`:
   - Veja o processo **`000003/2026`** (enviado por Ana da Poguba) aguardando atendimento técnico no Hub Aurora.
   - Abra o processo e confira o formulário de triagem e o histórico de movimentações.
2. Faça login como **`rodrigo.agro`** e acesse `/tramite`:
   - Veja o processo **`000002/2026`** tramitado por Marcos (Polo Industrial) para o Terminal Agro.
3. Faça login como **`carlos.diretoria`**:
   - Acesse `/tramite` e veja o processo **`000001/2026`** na Sede Centro, com o despacho decisório de aprovação de verba.

---

### Cenário 4: Agenda Corporativa & Reserva de Salas em Rondonópolis (`/agenda`)

Foram configuradas **7 salas de reunião corporativas** distribuídas entre as unidades de Rondonópolis:
- **Sede Centro:** *Auditório Marechal Rondon* (120 lugares) e *Sala Executiva Pantanal* (16 lugares)
- **Hub Vila Aurora:** *Espaço Inovação & Scrum Hub* (24 lugares) e *Lab de Pesquisa & IoT AgroTech* (12 lugares)
- **Distrito Industrial:** *Sala de Operações & Logística Polo* (18 lugares)
- **CIA Poguba:** *Sala de Treinamento CIA Poguba* (30 lugares)
- **Terminal Agro:** *Sala de Briefing do Terminal Agro* (10 lugares)

#### Eventos cadastrados para o mês:
1. **Comitê Executivo Nexus:** Amanhã das 09h às 11h30 na *Sala Executiva Pantanal* (Organizado por Carlos Mendes).
2. **Sprint Review AgroTech:** Em 2 dias das 14h às 16h no *Espaço Inovação Hub Aurora* (Restrito ao Hub Aurora — Organizado por Lucas Albuquerque).
3. **Alinhamento de Frotas e Safra:** Em 3 dias das 08h30 às 10h30 na *Sala Operações Polo* (Organizado por Marcos Souza).
4. **Workshop de Atendimento e LGPD:** Em 4 dias das 13h30 às 17h na *Sala CIA Poguba* (Organizado por Ana Paula Ferreira).
5. **Workshop Aberto Nexus:** Em 7 dias das 19h às 21h30 no *Auditório Marechal Rondon* (Público — Organizado por Mariana Silva).

#### Como validar no navegador:
1. Faça login como **`lucas.tech`** e acesse `/agenda`:
   - No calendário do mês, você verá **5 eventos**, incluindo o evento restrito *"Sprint Review: Plataforma de Telemetria e IA AgroTech"*.
2. Faça login como **`marcos.ops`** e acesse `/agenda`:
   - Você verá **4 eventos**. O evento do Hub Aurora **não aparece**, validando o filtro de público-alvo por unidade na Agenda!
3. Acesse sem login em aba anônima na Agenda pública (`/agenda` sem autenticação): apenas o evento público *"Workshop Nexus: O Futuro da Tecnologia no Campo"* no Auditório Marechal Rondon fica visível aos visitantes externos.

---

### Cenário 5: Base de Conhecimento Wiki & Isolamento de Manuais Técnicos (`/wiki`)

A Wiki corporativa conta com 6 manuais e procedimentos operacionais cadastrados para as unidades de Rondonópolis, com controle de versão e público-alvo:
- **`onboarding-colaborador`**: *Guia de Boas-Vindas & Integração do Colaborador Nexus* (Mariana Silva — Sede Centro)
- **`codigo-de-conduta`**: *Código de Ética, Integridade & Conduta Corporativa* (Carlos Mendes — Sede Centro)
- **`politica-seguranca-informacao`**: *Política de Segurança da Informação, Controle de Acessos & SOC* (Lucas Albuquerque — Hub Aurora)
- **`guia-desenvolvimento-agrotech`**: *Manual de Boas Práticas de Engenharia e Arquitetura de Software* (Restrito ao Hub Aurora — Lucas Albuquerque)
- **`manual-operacoes-polo-industrial`**: *Manual Operacional do Complexo Logístico e Segurança do Trabalho (SMS)* (Marcos Souza — Polo Industrial)
- **`protocolo-atendimento-produtor`**: *Diretrizes de Atendimento Integrado e Triagem Documental* (Ana Paula Ferreira — CIA Poguba)

#### Como validar no navegador:
1. Faça login como **`lucas.tech`** e acesse `/wiki`:
   - A árvore lateral exibe **6 páginas**, incluindo o *Manual de Boas Práticas de Engenharia e Arquitetura de Software* restrito à equipe técnica do Hub Aurora.
2. Faça login como **`marcos.ops`** e acesse `/wiki`:
   - A árvore exibe **5 páginas**. O manual confidencial de arquitetura do Hub Aurora **não aparece na árvore nem na busca**, demonstrando a aplicação do público-alvo por escopo da unidade (ADR 014 / ADR 030).
3. Faça login como **`carlos.diretoria`**:
   - Como Diretor Executivo da Sede, Carlos visualiza toda a base corporativa e o histórico de revisões.

---

### Cenário 6: Governança, IAM e Permissões Administrativas
1. Faça login como **`carlos.diretoria`**.
2. Acesse o menu de gestão e confirme permissões de visualização corporativa, governança e catálogo de serviços.
3. Compare o menu de navegação e as ações disponíveis em relação ao perfil estritamente operacional de **`rodrigo.agro`** (leitura básica e despachos locais).



