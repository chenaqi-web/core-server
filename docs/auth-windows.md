# Windows 下的身份认证开发

本文档是身份认证模块在 Windows PowerShell 环境下的权威开发指南。
Gateway 文档应链接到此处，而不是重复编写相同的操作流程。

## 适用范围与 Git

所有命令都应在对应仓库的根目录中执行。身份认证相关工作应在
`feat/auth-system` 分支上进行；不要直接修改 `master` 或 `main`。

```powershell
git fetch origin
git switch feat/auth-system
git status --short --untracked-files=all
git log -3 --oneline --decorate
```

在确认并理解当前工作区状态之前，不要执行数据库迁移、启动服务或修改本地配置。
绝对不要提交密码、SMTP 授权码、JWT 密钥、Token、Cookie 或验证码。

## 隔离的 MySQL 环境

实际使用的身份认证表是 ``user``，迁移文件如下：

```text
docs/sql/auth_user_precheck.sql
docs/sql/auth_user_migration.sql
```

请使用专门的开发或测试数据库，绝对不要使用生产数据库，也不要使用包含未知用户数据的数据库。
选择一个本地数据库名称，并且只有在确认目标服务器可用于开发测试、其中数据可随时丢弃后，才创建该数据库：

```powershell
$AuthDatabase = "replace_with_isolated_auth_database"
mysql -u root -p -e ('CREATE DATABASE IF NOT EXISTS `{0}` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;' -f $AuthDatabase)
mysql -u root -p -D $AuthDatabase -e 'SELECT DATABASE();'
mysql -u root -p -D $AuthDatabase -e 'SHOW CREATE TABLE `user`;'
```

将每个仓库 `conf/config.yaml` 中的 `Mysql` 配置指向这个隔离数据库，供本地开发使用。
不要提交数据库凭据或仅适用于本地的数据库名称。执行迁移之前，先运行只读的预检查脚本，
并手动检查所有结果集：

```powershell
Get-Content .\docs\sql\auth_user_precheck.sql | mysql -u root -p -D $AuthDatabase
```

检查时必须确认：数据库表和字段与当前源码一致；不存在重复的用户名或邮箱；不存在空邮箱；
不存在值为 NULL 的个人资料字段；不存在超长字段值；不存在非法角色；
并且已经了解当前 `status` 和 `auth_version` 的状态。

即使这两个字段尚不存在，也可以安全地执行预检查，因为脚本查询的是
`information_schema`，而不是直接从 `user` 表中读取不存在的字段。

只有完成上述检查后，并且仅在隔离数据库中，才能执行迁移：

```powershell
Get-Content .\docs\sql\auth_user_migration.sql | mysql -u root -p -D $AuthDatabase
Get-Content .\docs\sql\auth_user_precheck.sql | mysql -u root -p -D $AuthDatabase
mysql -u root -p -D $AuthDatabase -e 'SHOW CREATE TABLE `user`;'
```

迁移脚本不会删除用户、覆盖旧数据或自动修复数据。
当旧数据与新结构不兼容时，迁移脚本会在严格 SQL 模式下主动失败；
请手动处理这些不兼容数据，然后再重新执行迁移。

## Docker Redis

使用专门的 Redis 容器，并选择一个非默认的宿主机端口。
生成一个本地 Redis 密码，在 Core 的本地 YAML 配置中填写对应的 `Redis` 配置，
不要提交这个密码。

```powershell
$RedisContainer = "core-auth-dev-redis"
$RedisPort = 16379
$RedisPassword = -join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
docker run --detach --name $RedisContainer --publish "${RedisPort}:6379" redis:7.4-alpine redis-server --requirepass $RedisPassword
docker ps --filter "name=$RedisContainer"
docker exec $RedisContainer redis-cli --no-auth-warning -a $RedisPassword ping
```

本地测试结束后，只停止并删除这个指定名称的测试容器：

```powershell
docker stop $RedisContainer
docker rm $RedisContainer
```

## 环境变量

Core 从进程环境变量中读取下面四个值，而不是从 YAML 中读取。
请分别为 Access Token 和 Refresh Token 创建两个不同的 32 字节密钥，
并且绝对不要输出或提交这些密钥。

```powershell
function New-AuthSecret {
    $bytes = [byte[]]::new(32)
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    [Convert]::ToBase64String($bytes)
}

$env:JWT_ACCESS_SECRET = New-AuthSecret
$env:JWT_REFRESH_SECRET = New-AuthSecret
if ($env:JWT_ACCESS_SECRET -eq $env:JWT_REFRESH_SECRET) { throw "JWT 密钥必须不同" }

$env:QQ_SMTP_USERNAME = Read-Host -Prompt "QQ SMTP 账号"
$env:QQ_SMTP_AUTH_CODE = Read-Host -Prompt "QQ SMTP 授权码" -AsSecureString
$env:QQ_SMTP_AUTH_CODE = [System.Net.NetworkCredential]::new('', $env:QQ_SMTP_AUTH_CODE).Password
```

`QQ_SMTP_AUTH_CODE` 指的是 QQ SMTP 授权码，不是 QQ 登录密码。
自动化测试使用的是伪造的邮件发送器，不会发送真实邮件。

## Proto 与 Wire

当前生成的源码中记录的版本为 `protoc-gen-go v1.36.11` 和
`protoc-gen-go-grpc v1.6.2`。本项目用于保证可复现生成结果的基准版本为
`protoc v3.19.4`、`protoc-gen-go v1.36.11` 和
`protoc-gen-go-grpc v1.6.2`。

生成代码前，先检查工具版本：

```powershell
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

Core 会生成 `docs/proto` 目录中的所有 proto 文件；
Gateway 通过现有的 Make 目标生成身份认证 proto。
项目安装的 Wire 模块版本为 `github.com/google/wire v0.7.0`。

```powershell
# core-server
$ProtoFiles = Get-ChildItem .\docs\proto\*.proto | ForEach-Object FullName
& protoc --proto_path=./docs/proto --go_out=. --go_opt=module=backend/core-server --go-grpc_out=. --go-grpc_opt=module=backend/core-server $ProtoFiles
go run github.com/google/wire/cmd/wire ./cmd/server

# gateway
protoc --proto_path=./docs/proto --go_out=. --go_opt=module=backend/gateway --go-grpc_out=. --go-grpc_opt=module=backend/gateway ./docs/proto/auth.proto
go run github.com/google/wire/cmd/wire ./cmd/server
```

绝对不要手动修改 `*.pb.go`、`*_grpc.pb.go` 或 `wire_gen.go`。
运行生成器后，检查 `git diff --check` 和 `git status --short`；
如果出现任何不符合预期的生成差异，应立即停止并调查原因。

## 测试与质量检查

身份认证单元测试是安全的，因为它们使用 `miniredis`、`sqlmock`、
伪造的 gRPC 客户端和伪造的 SMTP 传输层。
这些测试不依赖 Core、Gateway、MySQL、Redis、Agent、Ollama、Kafka 或 QQ SMTP。

```powershell
# core-server 身份认证单元测试
go test ./internal/config -run '^TestAuthConfig' -count=1
go test ./internal/infras/auth -count=1
go test ./internal/infras/cache -run 'Test(AuthStore|GenerateEmailCode|EmailCode)' -count=1
go test ./internal/infras/mail -run '^TestQQMailSender' -count=1
go test ./internal/infras/repo -run '^TestUserRepo' -count=1
go test ./internal/application -run '^TestAuthService' -count=1
go test ./internal/rpc -run '^TestAuthRPC' -count=1

# 重复运行对并发敏感的 Core 测试
go test ./internal/application -run 'TestAuthService(ConcurrentLoginOnlyOneSucceeds|ConcurrentRefreshOnlyOneSucceeds|ResetAuthVersionInvalidatesOldTokensWhenRedisCleanupFails)' -count=10
go test ./internal/infras/cache -run 'Test(AuthStoreConcurrent|EmailCode(Concurrent|Late))' -count=10

# gateway 身份认证单元测试
go test ./internal/config -count=1
go test ./internal/facade -run '^TestGatewayRoutesKeepHealthAndAddAuthEndpoints$' -count=1
go test ./internal/facade/controller -run 'Test(AuthController|RefreshCookieManager)' -count=1
go test ./internal/facade/middleware -run '^Test(AuthMiddleware|BearerToken)' -count=1
go test ./internal/client/rpc/core-rpc/authpb -count=1

# 重复运行 Gateway 身份认证测试
go test ./internal/facade ./internal/facade/controller ./internal/facade/middleware -count=10

# 分别在每个仓库的根目录中执行
go test ./... -run '^$'
go vet ./...
go mod verify
go mod tidy -diff
git diff --check
```

不要把 `go test ./... -run '^$'` 当作完整测试。
在此仓库中，完整测试套件包含一些现有测试，它们可能会连接已经配置的
MySQL、Redis、gRPC、Agent、Ollama、Kafka 或 AI Chat 服务。

只有在确认所有外部依赖都已经隔离后，才可以运行完整测试套件。
不要仅仅为了让身份认证检查通过，就修改这些与身份认证无关的测试。

## 启动顺序与地址

依次启动：隔离的 MySQL、指定名称的 Docker Redis 容器、Core、Gateway。

Core 从 `core-server/conf/config.yaml` 的 `server.addr` 获取 gRPC 监听地址。
Gateway 从 `gateway/conf/config.yaml` 的 `server.addr` 获取 HTTP 监听地址，
并从 `rpc.core_server_addr` 获取 Core 服务地址。
`storage.base_url` 与身份认证无关，不要为了身份认证修改它。

```powershell
# 完成隔离配置并设置环境变量后，在 core-server 中运行
go run ./cmd/server

# 在另一个 PowerShell 窗口中进入 gateway 并运行
go run ./cmd/server
```

在 Gateway 根目录中，应根据实际 YAML 配置计算 HTTP 基础地址，
不要在脚本或文档中写死端口：

```powershell
$configText = Get-Content .\conf\config.yaml -Raw
$GatewayAddr = [regex]::Match($configText, '(?ms)^server:\s*\r?\n\s*addr:\s*"([^"]+)"').Groups[1].Value
if ([string]::IsNullOrWhiteSpace($GatewayAddr)) { throw "未找到 server.addr" }
$GatewayBase = "http://$GatewayAddr"
```

调用身份认证接口之前，先验证现有的健康检查路由：

```powershell
Invoke-RestMethod -Method Get -Uri "$GatewayBase/api/v1/health/ping"
```

## 手动 HTTP 流程

使用一次性的测试用户和 PowerShell Web 会话。
不要把 Token、验证码或生产环境凭据粘贴到 Shell 历史记录或文档中。

```powershell
$WebSession = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$TestEmail = "replace-with-disposable@example.invalid"
$TestUsername = "replace-with-disposable-username"
$TestPassword = Read-Host -Prompt "一次性测试密码"

# 发送注册验证码。只能从已经配置的测试邮件发送器中获取验证码。
Invoke-RestMethod -Method Post -Uri "$GatewayBase/api/v1/auth/send-email-code" -ContentType 'application/json' -Body (@{ email = $TestEmail; purpose = 'register' } | ConvertTo-Json)
$RegistrationCode = Read-Host -Prompt "注册验证码"

Invoke-RestMethod -Method Post -Uri "$GatewayBase/api/v1/auth/register" -ContentType 'application/json' -Body (@{ username = $TestUsername; email = $TestEmail; password = $TestPassword; confirm_password = $TestPassword; email_code = $RegistrationCode } | ConvertTo-Json)

$Login = Invoke-RestMethod -Method Post -Uri "$GatewayBase/api/v1/auth/login" -WebSession $WebSession -ContentType 'application/json' -Body (@{ username = $TestUsername; password = $TestPassword } | ConvertTo-Json)
$AccessToken = $Login.data.access_token

Invoke-RestMethod -Method Post -Uri "$GatewayBase/api/v1/auth/refresh" -WebSession $WebSession
Invoke-RestMethod -Method Post -Uri "$GatewayBase/api/v1/auth/logout" -WebSession $WebSession
```

登录和刷新接口返回的 JSON 中包含 Access Token，但绝不会包含 Refresh Token。

可以使用 `Authorization: Bearer $AccessToken` 调用受保护的测试路由；
不要为了完成这项检查而人为添加一个生产环境路由。

测试密码重置时，先申请用途为 `reset_password` 的验证码，
然后向 `/api/v1/auth/reset-password-by-email` 提交
`email`、`email_code`、`new_password` 和 `confirm_password`。
随后确认浏览器中的会话 Cookie 已被清除，并且只有新密码能够登录。

## Session、Cookie 与 CORS 边界

系统没有在线心跳机制。只有当账号对应的 Redis Session 存在时，
该账号才会被视为在线；关闭浏览器并不会自动退出登录。

Session 最长保留七天，并会在主动退出、通过邮箱重置密码或到期后被清除。
服务器无法可靠判断浏览器窗口是否仍处于打开状态。

刷新 Cookie 的名称为 `refresh_token`，路径为 `/api/v1/auth`，
属性为 HttpOnly、SameSite=Lax，不设置 Domain，
MaxAge/Expires 为七天。

其 Secure 标记由 Gateway 的 `auth.cookie_secure` 配置决定。
由于 Cookie 路径被限制为 `/api/v1/auth`，普通业务路由不会收到刷新 Cookie，
而刷新、退出等身份认证路由会收到该 Cookie。

每次刷新都会立即轮换 Token，旧 Token 会立刻失效。
如果服务器已经成功完成 Token 轮换，但携带新 Cookie 的响应丢失，
客户端可能会失去继续刷新的能力。

当前阶段特意没有实现 Refresh Token 宽限期、重放恢复或在线心跳。
如果 Session 卡住，可以通过重置密码清除 Session，或者等待 Session 到期。

当前阶段不会配置具体的前端 Origin，也不会修改 CORS。
未来浏览器客户端使用 Cookie 时，必须发送 `credentials: "include"`；
Gateway 必须设置 `AllowCredentials=true`，并明确列出允许的 Origin，
绝对不能使用 `*`。

## 初始管理员

系统没有创建管理员的 API。

备份数据库并在隔离数据库或经过批准的数据库中确认目标用户后，
运维人员可以使用实际的表和字段，将现有账号提升为管理员：

```sql
START TRANSACTION;

UPDATE `user`
SET role = 'admin'
WHERE email = 'replace-with-admin@example.invalid'
  AND deleted_at IS NULL;

SELECT id, name, email, role
FROM `user`
WHERE email = 'replace-with-admin@example.invalid';

COMMIT;
```

不要在公开注册参数中添加 `role`，不要创建管理员创建接口，
也不要为此操作添加注册开关。
