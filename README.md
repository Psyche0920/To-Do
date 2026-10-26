# PreZero To-Do：本地操作向导

当前项目包含 Go HTTP API 和 Docker Compose 管理的 PostgreSQL。任务列表接口目前返回代码中的示例数据；Go 尚未连接数据库，Nuxt 前端尚未实现。`tasks` 表结构、合法任务插入及仅空格标题拒绝测试已由学习者手动验证，结果见第 10.6 节；其余示例与扩展测试不代表已经执行。

## 1. 运行位置与前置条件

需要 Go（当前 `go.mod` 声明为 1.27.3）、Docker Desktop 和 Docker Compose。

除特别说明外，下面命令都在项目根目录运行：

```bash
cd /Users/hanweijia/To-Do
```

根目录包含 `compose.yaml`、`.env.example` 和 `.gitignore`。Go 模块目录是 `backend`，入口文件是 `backend/cmd/api/main.go`。

## 2. 准备本地环境变量

第一次配置、且尚无 `.env` 时，运行：

```bash
cp .env.example .env
```

如果已经有 `.env`，跳过复制，避免覆盖自己的配置。在编辑器中检查这三个变量并设置本地密码：

- `POSTGRES_USER`：数据库用户名，本向导使用 `wehan`。
- `POSTGRES_PASSWORD`：本地数据库密码。
- `POSTGRES_DB`：数据库名，本向导使用 `wehan`。

`.env` 已通过 `.gitignore` 排除，不应提交到 Git。`.env.example` 作为配置模板保留。

这些变量用于首次初始化数据库；已经存在数据卷时，修改 `.env` 不会自动更改现有数据库的用户、密码或数据库名。

## 3. 启动 Docker Desktop

在 macOS 上运行：

```bash
open -a Docker
```

等待 Docker Desktop 启动完成，然后检查后台服务：

```bash
docker info
```

能显示服务端信息后，再继续启动数据库。

## 4. 验证配置并启动 PostgreSQL

先检查 Compose 配置：

```bash
docker compose config -q
```

没有错误输出表示配置检查通过。

启动数据库，并查看容器状态：

```bash
docker compose up -d
docker compose ps
```

第一次运行会下载 PostgreSQL 镜像、创建网络与数据卷，并初始化数据库。容器显示 `Started` 或 `Up` 不代表初始化已经完成。

检查数据库是否准备好接受连接：

```bash
docker compose exec db pg_isready -U wehan -d wehan
```

预期输出包含：

```text
/var/run/postgresql:5432 - accepting connections
```

如果首次启动时显示 `no response`，等待几秒后重复检查。如果持续失败，查看日志：

```bash
docker compose logs --tail=50 db
```

如果 `.env` 中的用户名或数据库名不是 `wehan`，对应修改命令中的 `-U` 和 `-d` 参数。

`pg_isready` 只确认数据库服务是否就绪；即使指定的用户名或数据库名不正确，也可能报告服务已就绪。它不证明身份认证成功，不代表 Go 已连接数据库，也不验证任务数据表是否存在。下面可用实际 SQL 查询确认当前用户和数据库：

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "SELECT current_user, current_database();"
```

预期当前用户和数据库名均为 `wehan`。

## 5. 启动 Go API

在项目根目录的一个终端中运行：

```bash
go -C backend run ./cmd/api
```

`-C backend` 让 Go 从包含 `go.mod` 的模块目录执行命令。等价写法是在 `backend` 目录运行 `go run ./cmd/api`。

终端会显示：

```text
API server listening on :8080
```

让这个终端保持运行。另开一个终端执行接口测试。

## 6. 测试接口

健康检查：

```bash
curl -i http://localhost:8080/healthz
```

预期状态码为 `200 OK`，正文为：

```json
{"status":"ok"}
```

任务列表：

```bash
curl -i http://localhost:8080/api/tasks
```

预期状态码为 `200 OK`，响应头包含 `Content-Type: application/json`。当前示例正文为：

```json
[{"title":"hackthon","completed":true},{"title":"sleep","completed":false}]
```

`curl` 默认发送 GET 请求，`-i` 用于同时显示响应头。GET 是请求方法，不是 URL 路径的一部分；任务列表路径是 `/api/tasks`，不是 `/get/api/tasks`。

当前 `/healthz` 只确认 Go 服务能够响应，不检查 PostgreSQL。任务列表数据来自 `listTasksHandler` 中的 `[]Task`，还未从数据库读取。

## 7. 修改代码、停止与再次启动

Go API 不会自动加载保存后的代码。在运行 API 的终端按 **Control + C**，停止旧服务，再从项目根目录运行：

```bash
go -C backend run ./cmd/api
```

只停止数据库容器、保留数据：

```bash
docker compose stop
```

再次启动数据库：

```bash
docker compose up -d
```

如果需要移除容器和 Compose 网络，可以运行：

```bash
docker compose down
```

命名数据卷仍会保留。不要在日常停止操作中加 `-v`；`docker compose down -v` 会删除本项目的命名卷及其中的数据库数据。

## 8. 端口与数据卷对应关系

当前 Go 在 Mac 上运行，PostgreSQL 在容器中运行：

```text
浏览器或 curl → localhost:8080 → Go API
Go 将来连接数据库 → localhost:5432 → Docker 转发 → PostgreSQL 容器的 5432
```

Compose 中的 `"5432:5432"` 是 **Mac 端口:容器端口** 的转发规则；Go 使用的 `localhost:5432` 是 **目标主机:目标端口**。它们用于不同配置。

```yaml
volumes:
  - pgdata:/var/lib/postgresql/data
```

这行把名为 `pgdata` 的 Docker 卷挂载到容器内的 PostgreSQL 数据目录。Compose 文件底部的 `volumes: pgdata:` 声明该卷。PostgreSQL 通过这个目录读写数据，Docker 管理卷的存储与生命周期。

## 9. 常见问题

| 现象 | 原因与处理 |
| --- | --- |
| `failed to connect to the docker API` 或 `docker.sock ... no such file` | Docker Desktop 后台服务未就绪。启动 Docker Desktop，等待 `docker info` 能连接服务端。 |
| 容器已启动，但 `pg_isready` 返回 `no response` | 首次初始化可能尚未完成。稍等后重试；持续失败时查看 `docker compose logs --tail=50 db`。 |
| `FATAL: role "todo" does not exist` | 本地配置的用户和数据库均为 `wehan`；连接命令使用 `-U wehan -d wehan`。不要仅凭 `pg_isready` 成功判断用户存在。 |
| `go: cannot find main module` | 当前目录没有 `go.mod`。从根目录使用 `go -C backend run ./cmd/api`，或进入 `backend` 再运行。 |
| `404 Not Found` | 检查请求路径是否与已注册路由一致，修改代码后重启 Go 服务。 |
| `405 Method Not Allowed`，且 `Allow: Get` | 请求方法区分大小写。代码应注册 `GET /api/tasks`；保存后重启旧服务。 |
| `address already in use` 或端口已被占用 | 找到并停止先前启动的服务，避免重复启动。 |

## 10. 创建 tasks 表并测试数据库读写

本节由学习者手动操作，所有命令都在 `To-Do` 根目录运行。先确认 PostgreSQL 已经启动并接受连接。

### 10.1 准备建表脚本

在编辑器中创建或检查 `backend/schema.sql`：

```sql
CREATE TABLE tasks (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(200) NOT NULL CHECK (length(btrim(title)) > 0),
    completed BOOLEAN NOT NULL DEFAULT FALSE
);
```

注意列名必须一致：`btrim(title)` 引用已声明的 `title` 列。如果误写成 `btrim(titile)`，PostgreSQL 会在建表时报告该列不存在，建表语句失败。这是列名拼写错误，与插入任务时约束检查失败是两种不同情况。脚本末尾以 `);` 结束。

| 定义 | 作用 |
| --- | --- |
| `CREATE TABLE tasks` | 创建任务表。 |
| `id BIGINT GENERATED ALWAYS AS IDENTITY` | 用整数保存任务编号，插入时由数据库自动生成。 |
| `PRIMARY KEY` | 保证任务编号唯一且非空，用于定位任务。 |
| `title VARCHAR(200)` | 标题最多 200 个字符。 |
| `NOT NULL` | 不允许字段值为 SQL 的 `NULL`；`NULL` 与空字符串不同。 |
| `CHECK (length(btrim(title)) > 0)` | 去掉标题首尾的普通空格后，检查字符数是否大于零。 |
| `completed BOOLEAN NOT NULL DEFAULT FALSE` | 完成状态使用布尔类型；未提供时使用 `false`。 |

`btrim` 在这里仅用于检查，不会自动修改最终保存的标题。例如 `'  Read Go  '` 可以通过检查，原来的首尾空格仍会被保存。这个表达式不等于过滤所有种类的空白字符。

### 10.2 执行建表脚本

```bash
docker compose exec -T db psql -U wehan -d wehan -v ON_ERROR_STOP=1 < backend/schema.sql
```

| 命令部分 | 作用 |
| --- | --- |
| `docker compose exec` | 在正在运行的服务容器中执行命令。 |
| `-T` | 不分配交互终端，适合从文件重定向输入。 |
| `db` | Compose 中数据库的服务名。 |
| `psql` | PostgreSQL 命令行客户端。 |
| `-U wehan` | 使用数据库用户 `wehan`。 |
| `-d wehan` | 连接名为 `wehan` 的数据库。 |
| `-v ON_ERROR_STOP=1` | 脚本遇到 SQL 错误时停止，并返回失败状态。 |
| `< backend/schema.sql` | 由 Mac 上的 shell 读取本地 SQL 文件，并传给容器中的客户端。 |

预期输出：

```text
CREATE TABLE
```

建表脚本只需要执行一次。如果提示 `relation "tasks" already exists`，先检查已有表结构，不要为了重跑脚本直接删除已有表：

```bash
docker compose exec db psql -U wehan -d wehan -c "\d tasks"
```

如果曾修改建表脚本，已有表不会自动跟着改变；需要先核对表结构，再决定后续迁移方式。

### 10.3 插入一条任务

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "INSERT INTO tasks (title) VALUES ('Learn PostgreSQL') RETURNING id, title, completed;"
```

输入只包含标题；数据库应自动生成 `id`，并将 `completed` 设为 `false`。预期输出类似：

```text
 id |      title       | completed
----+------------------+-----------
  1 | Learn PostgreSQL | f
```

具体编号取决于当前数据库状态，不保证一定是 1。每次重复执行这条命令，都会再插入一条任务。

### 10.4 查询已保存的数据

```bash
docker compose exec db psql -U wehan -d wehan -c "SELECT id, title, completed FROM tasks ORDER BY id;"
```

预期能查到刚才插入的任务。`f` 和 `t` 是 `psql` 显示布尔值 `false` 和 `true` 的方式。

### 10.5 验证标题约束失败时的行为

先记录当前总行数：

```bash
docker compose exec db psql -U wehan -d wehan -c "SELECT COUNT(*) FROM tasks;"
```

尝试插入空标题：

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "INSERT INTO tasks (title) VALUES ('');"
```

再尝试只包含普通空格的标题：

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "INSERT INTO tasks (title) VALUES ('   ');"
```

这两条命令预期都失败，错误信息包含：

```text
violates check constraint
```

这是期望的测试结果：数据库拒绝无效任务，不会保存这次插入的记录。再次查询总行数：

```bash
docker compose exec db psql -U wehan -d wehan -c "SELECT COUNT(*) FROM tasks;"
```

在没有其他写入的情况下，行数应与失败测试前一致。本次验证前后均为 **2**。后续重新测试时，以当时实际记录数为基准，不要求始终等于 2。

如果错误详情显示 `Failing row contains (3, ..., f)`，那是数据库展示被拒绝的记录，不是插入成功。预期的约束错误无需修复；应核对任务总行数没有增加。

约束也适用于 `UPDATE`：如果更新后的标题违反检查，这次更新语句失败，原来的记录保持不变。若在显式事务中发生错误，该事务还需要回滚；本节每条命令都是独立执行的。

| 标题输入 | 预期结果 | 原因 |
| --- | --- | --- |
| `'Read Go'` | 成功 | 非空且长度不超过 200。 |
| `''` | 失败 | `CHECK` 检查结果为假。 |
| `'   '` | 失败 | 去掉普通空格后长度为零。 |
| `NULL` | 失败 | 违反 `NOT NULL`。 |
| 超过 200 个字符 | 失败 | 超出 `VARCHAR(200)` 长度限制。 |

失败的插入可能已经消耗一个自动编号，因此后续 `id` 不一定连续；编号有间隔不代表无效记录已保存。

### 10.6 已验证结果与进度

以下结果来自 2026-10-26 学习者提供的终端输出，以及学习者对失败后记录数的确认；更新文档时没有再次执行数据库写入。

| 测试项 | 观察到的结果 | 状态 |
| --- | --- | --- |
| 检查已有表：`psql ... -c "\d tasks"` | `id` 为自动编号主键，`title` 为最多 200 字符且非空，`completed` 默认 `false`，存在 `tasks_title_check`。 | 通过 |
| 插入合法标题并 `RETURNING id, title, completed` | 返回 `id = 2`、`title = Learn PostgreSQL`、`completed = f`，并显示 `INSERT 0 1`。 | 通过 |
| 约束失败测试前查询 `COUNT(*)` | 总行数为 `2`。 | 通过 |
| 插入仅包含普通空格的标题 `'   '` | 被 `tasks_title_check` 拒绝，错误详情展示待插入编号为 `3` 的记录。 | 通过：预期拒绝 |
| 失败后再次查询 `COUNT(*)` | 学习者确认仍为 `2`，无效记录没有保存。 | 通过 |

表已经存在时，重复执行 `CREATE TABLE tasks` 返回 `relation "tasks" already exists`；这是重复建表提示。本次通过检查现有表结构完成核对，没有删除或重建表。

本步核心验收完成，按此前的 18 个学习里程碑计数为 **7/18，约 39%**。

第 10.4 节的完整列表查询命令保留为后续操作指南；当前记录不单独声称已核验其输出。空字符串 `''`、`NULL`、超过 200 字符标题及无效 `UPDATE` 仍属于扩展测试，本次没有记录它们的执行结果。

数据库里的任务不会立刻出现在 `/api/tasks` 中，因为 Go 处理函数仍返回代码中的示例列表。

## 11. 添加 PostgreSQL 的 Go 驱动

这是第 8 个学习里程碑的第一个小步骤。Go 标准库 `database/sql` 提供统一的数据库接口；`pgx/stdlib` 为该接口提供 PostgreSQL 驱动。

### 11.1 下载并记录依赖

先进入包含 `go.mod` 的 `backend` 目录，再执行：

```bash
cd /Users/hanweijia/To-Do/backend
go get github.com/jackc/pgx/v5/stdlib
```

这条命令会下载驱动及所需依赖，并更新：

- `backend/go.mod`：记录模块依赖和版本。
- `backend/go.sum`：记录依赖校验信息。

它的用途类似 Python 项目添加数据库驱动依赖；此时还没有创建数据库连接。

### 11.2 验证驱动版本

仍在 `backend` 目录运行：

```bash
go list -m github.com/jackc/pgx/v5
```

本次实际输出：

```text
github.com/jackc/pgx/v5 v5.11.0
```

2026-10-26 学习者已手动运行以上命令并提供成功输出；本地 `go.mod` 也已记录 `github.com/jackc/pgx/v5 v5.11.0`。

当前 `go.mod` 中的驱动带有 `// indirect` 标记；此阶段应用代码还未导入它。后续加入实际导入后，再通过 `go mod tidy` 整理依赖，目前无需手动修改标记。

驱动安装已验证，但不能据此判断 Go 已连接 PostgreSQL；Go 代码仍返回示例任务列表。完整的第 8 步还需要连接验证和数据库读取验证，因此当前仍按 **7/18，约 39%** 记录；第 8 步全部通过后约为 **44%**。

## 12. 下一步

在 Go 中配置数据库连接并验证能连上 PostgreSQL，然后让 `/api/tasks` 从 `tasks` 表读取数据。目前尚未实现新增、修改、删除或数据库持久化 API。

## 官方参考

- [PostgreSQL 官方 Docker 镜像与 Compose 示例](https://github.com/docker-library/docs/blob/master/postgres/README.md)
- [Docker Compose 端口配置](https://docs.docker.com/reference/compose-file/services/#ports)
- [PostgreSQL 17：pg_isready](https://www.postgresql.org/docs/17/app-pg-isready.html)

- [PostgreSQL 17：创建数据表](https://www.postgresql.org/docs/17/tutorial-table.html)
- [PostgreSQL 17：Identity Columns](https://www.postgresql.org/docs/17/ddl-identity-columns.html)
- [PostgreSQL 17：数据约束](https://www.postgresql.org/docs/17/ddl-constraints.html)

- [pgx 官方仓库](https://github.com/jackc/pgx)
- [pgx 的 database/sql 适配层](https://github.com/jackc/pgx/blob/master/stdlib/sql.go)
