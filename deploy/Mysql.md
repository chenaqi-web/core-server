# Mysql的镜像下载

下面是一个简单的mysql下载命令，暂无挂载

```bash
docker run -d \
--name mysql8.4.7 \
-p 3306:3306 \
-e MYSQL_ROOT_PASSWORD=123456 \
-e TZ=Asia/Shanghai \
mysql:8.4.7
```