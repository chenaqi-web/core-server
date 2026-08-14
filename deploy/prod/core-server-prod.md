# Core Server 生产部署

本文档使用独立 Docker 容器部署 `core-server`，不加入 Docker Compose。MySQL 和 Redis 必须已经运行，并加入 `chenaqi-net` 网络。

项目目录：`/home/newweb/core-server`

## 一.首次部署

先确认共享网络已经存在：

```bash
docker network inspect chenaqi-net >/dev/null 2>&1 || docker network create --driver bridge chenaqi-net
```

构建镜像。`CONFIG_FILE` 指定生产配置，该配置会使用网络中的 `mysql` 与 `redis` 作为服务地址：

cd /home/newweb/core-server


```bash
docker build \
  -f deploy/prod/Dockerfile \
  --build-arg CONFIG_FILE=conf/config.prod.yaml \
  -t renai-core-server:prod \
  .
```

启动容器并加入现有网络：

```bash
docker run -d \
  --name core-server \
  --network chenaqi-net \
  -p 8081:8081 \
  --restart unless-stopped \
  renai-core-server:prod
```

查看启动日志：

```bash
docker logs -f core-server
```


## 二.更新部署

拉取或上传新代码后，重新构建并替换容器：

```bash
cd /home/newweb/core-server
docker build \
  -f deploy/prod/Dockerfile \
  --build-arg CONFIG_FILE=conf/config.prod.yaml \
  -t renai-core-server:prod \
  .

docker rm -f core-server

docker run -d \
  --name core-server \
  --network chenaqi-net \
  -p 8081:8081 \
  --restart unless-stopped \
  renai-core-server:prod

docker logs -f core-server
```

更新 `core-server` 容器不会删除 MySQL 和 Redis 的持久化数据。
