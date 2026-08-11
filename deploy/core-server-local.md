# Core Server 本地容器部署

容器使用 `conf/config.docker.local.yaml` 配置文件，Kafka 已关闭。MySQL 和 Redis 通过 Docker 网络别名 `mysql`、`redis` 访问。

请在 `core-server` 目录中执行以下命令。

```powershell
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"

docker network create renai-local

docker network connect --alias mysql renai-local <MYSQL_CONTAINER_NAME>
docker network connect --alias redis renai-local <REDIS_CONTAINER_NAME>

docker build -f deploy/Dockerfile -t renai-core-server:local .

docker run -d `
  --name renai-core-server `
  --network renai-local `
  -p 8081:8081 `
  --restart unless-stopped `
  renai-core-server:local

docker logs -f renai-core-server
```

将 `<MYSQL_CONTAINER_NAME>` 和 `<REDIS_CONTAINER_NAME>` 替换为 `docker ps` 查询到的容器名称。

## 线上配置

同一个 Dockerfile 可以使用不同的 YAML 配置文件构建镜像。在 `conf/` 下创建线上配置文件，再通过 `CONFIG_FILE` 指定：

```powershell
docker build -f deploy/Dockerfile `
  --build-arg CONFIG_FILE=conf/config.docker.prod.yaml `
  -t renai-core-server:prod .
```
