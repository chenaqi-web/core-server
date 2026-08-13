# Production Middleware

Create `.env` in this directory from `.env.example`, then replace every placeholder with a unique secret. Keep `.env` only on the server.

```bash
cp .env.example .env
chmod 600 .env
mkdir -p /home/docker/mysql /home/docker/redis /home/docker/chroma
docker compose --env-file .env -f docker-compose-prod.yml config
docker compose --env-file .env -f docker-compose-prod.yml up -d
docker compose --env-file .env -f docker-compose-prod.yml ps
```

MySQL, Redis, and Chroma are reachable only from containers attached to `chenaqi-net`. Run the application containers on that same network with `mysql`, `redis`, and `chroma` as their respective hosts.

For agent-server, copy `agent-server/deploy/.env.production.example` to a server-only environment file and set `CHROMA_AUTH_TOKEN` to the exact value used by this Compose file. Start it with `--env-file` and attach it to `chenaqi-net`.

The SQL scripts are imported only when the MySQL data directory is empty. If `/home/docker/mysql` already contains data, import any missing scripts manually.
