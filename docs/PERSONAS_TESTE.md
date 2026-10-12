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

### Cenário 3: Trâmite de Processos entre Unidades de Rondonópolis
1. Faça login como **`ana.protocolo`** (CIA Poguba).
2. Acesse `/tramite` e crie um novo processo eletrônico (ex.: *"Solicitação de Manutenção de Equipamentos de Triagem"*).
3. Tramite o processo direcionando-o para a unidade **Complexo Operacional - Distrito Industrial (`POLO-IND`)**.
4. Faça login como **`marcos.ops`** (Polo Industrial), acesse `/tramite`:
   - O processo aparece na caixa de entrada da unidade.
   - Marcos pode despachar, anexar parecer ou concluir o processo.

---

### Cenário 4: Governança, IAM e Permissões Administrativas
1. Faça login como **`carlos.diretoria`**.
2. Acesse o menu de gestão e confirme permissões de visualização corporativa, governança e catálogo de serviços.
