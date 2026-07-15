# Redis 的镜像下载

Redis 和数据库类似，目前也不需要额外安装插件，所以直接使用 docker run 启动容器即可。


如果你想开启持久化，可以加上这个参数  --appendonly yes 表示开启 AOF 持久化。

```bash
docker run -d \
--name redis8.8 \
-p 6379:6379 \
-e TZ=Asia/Shanghai \
redis:8.8
```