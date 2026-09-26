# Deploy em servidor (Docker Compose)

Sobe a stack completa em modo produção (`APP_ENV=production`) usando só o
`docker-compose.yml`: PostgreSQL, RabbitMQ, Redis, MinIO, migrations, API,
worker e frontend.

## Pré-requisitos

- Docker Engine + Docker Compose v2, `git`, `openssl` e `make`.
- ~4 GB de RAM (o build do frontend é a etapa mais pesada).
- Portas livres no host (padrão do script): **3010** frontend, **8010** API
  e WebSocket, **9010** MinIO (URLs pré-assinadas), **9011** console MinIO.
  Elas diferem das portas de dev (3002/8002/9002/9003) para não colidir com
  outras stacks na mesma máquina.

## Primeira vez

```bash
git clone https://github.com/yurythx/Projeto-Nexus.git
cd Projeto-Nexus
./scripts/deploy.sh 192.168.1.42    # IP/DNS que o navegador usa
make prod-seed-admin                 # imprime a senha do admin UMA vez
```

O script:

1. gera `.env` a partir do `.env.example` com segredos aleatórios
   (`openssl rand`), `APP_ENV=production` e as URLs públicas apontando para
   o host informado;
2. gera `secrets/local_auth_private_key.pem` (login local RS256);
3. faz `docker compose up -d --build` e espera todos os serviços ficarem
   `healthy`.

Acesse `http://<host>:3010` e entre com `admin` e a senha impressa.

## Atualizar

Envie as mudanças por git e, no servidor:

```bash
make deploy        # git pull --ff-only + rebuild + restart
```

O `.env` e a chave existentes são reaproveitados; migrations novas rodam
sozinhas (serviço `migrate`) antes da API e do worker.

## Cuidados

- **Não apague nem regenere o `.env`** depois que houver dados: a senha do
  banco fica gravada no volume e `CONFIG_ENCRYPTION_KEY` cifra segredos
  salvos pela tela de Configurações.
- Mudou host ou porta pública? Ajuste `FRONTEND_URL`, `API_PUBLIC_URL`,
  `WEBSOCKET_PUBLIC_URL` e `MINIO_PUBLIC_URL` no `.env` e rode `make deploy`
  — as `NEXT_PUBLIC_*` são embutidas no build do frontend.
- Sem Keycloak configurado (`KEYCLOAK_ISSUER_URL` vazio) só o login local
  fica disponível — esperado num ambiente de teste.
- A stack é servida em HTTP puro; para expor fora da rede interna, coloque
  um proxy reverso com TLS na frente e ajuste as URLs e `TRUSTED_PROXIES`.

## Ambiente de teste: Keycloak + dados fictícios

Enquanto não há um Keycloak oficial, `make demo-keycloak` (depois do
primeiro deploy) sobe um **Keycloak de teste** (`docker-compose.keycloak.yml`,
porta **8180**) já com o realm `nexus` e carrega no Postgres uma estrutura
organizacional fictícia:

| Entidade | Unidades | Departamentos |
|---|---|---|
| Prefeitura Municipal (PREF) | Sede Administrativa | RH, TI, Contabilidade, Compras, Administração, Governo |
| Secretaria de Saúde (SEMSA) | PSF Sagrada Família, PSF Conjunto, PSF Marechal Rondon, UPA 24h | padrão* |
| Secretaria de Educação (SEMED) | Escolas Maria Elza, Marechal Dutra, Silvestre, Elizabete | padrão* |
| SEMPRAS | Sede, CRAS Conjunto, CRAS Ana Carla, CRAS Alfredo, CREAS, Centro POP | padrão* |

\* Direção, Administração, Atendimento, Recursos Humanos, Almoxarifado.

- **50 usuários por unidade**, divididos igualmente entre os departamentos
  (754 no total, com as contas `teste.admin`, `teste.auditor`, `teste.iam`
  e `teste.conteudo`). Lista em `deploy/demo/usuarios.csv`; senha comum em
  `DEMO_USER_PASSWORD` no `.env` do servidor.
- Cada departamento é um grupo no Keycloak (ex.: `SEMSA-UPA-ADM`) mapeado
  para o perfil **Servidor** com lotação no departamento; a Administração
  de cada unidade também recebe **Protocolo** na unidade.
- Os usuários já existem no Nexus antes do 1º login (Diretório com cargo e
  ramal); o login pelo Keycloak só os atualiza.
- Console do Keycloak: `http://<host>:8180/admin` (`admin` /
  `KEYCLOAK_ADMIN_PASSWORD` do `.env`).

Mudar a estrutura: edite `scripts/demo-data/generate.mjs`, rode
`make demo-generate` e faça commit dos arquivos gerados. O SQL é reaplicável
(`make demo-seed`), mas o realm só é importado quando ainda não existe —
para reimportar, `docker compose rm -sf keycloak && docker volume rm
projeto-nexus_keycloak_data` e `make deploy`.

Trocar pelo Keycloak oficial: aponte `KEYCLOAK_ISSUER_URL` (e o client
secret) para ele, tire `docker-compose.keycloak.yml` do `COMPOSE_FILE` no
`.env` e rode `make deploy`.

## Comandos úteis

| Comando | O que faz |
|---|---|
| `make prod-ps` | estado dos containers |
| `make prod-logs` | logs de todos os serviços |
| `make prod-seed-admin` | cria/reseta o admin local |
| `make prod-down` | para a stack (mantém os volumes) |
