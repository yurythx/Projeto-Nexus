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

## Comandos úteis

| Comando | O que faz |
|---|---|
| `make prod-ps` | estado dos containers |
| `make prod-logs` | logs de todos os serviços |
| `make prod-seed-admin` | cria/reseta o admin local |
| `make prod-down` | para a stack (mantém os volumes) |
