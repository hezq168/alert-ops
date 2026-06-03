# alert-ops

一个基于 **Gin + Vue3** 的告警运营管理平台，提供完整的 RBAC 权限管理和多通道告警转发功能。

![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![Vue Version](https://img.shields.io/badge/Vue-3.5+-4FC08D?logo=vue.js)
![License](https://img.shields.io/badge/license-MIT-green)
[![Swagger](https://img.shields.io/badge/docs-Swagger-85EA2D?logo=swagger)](http://localhost:8080/swagger/index.html)

## 📖 项目简介

alert-ops 是一个前后端分离的告警运营管理系统，专注于告警接入、规则匹配、多通道转发与 AI 辅助分析。项目采用 Go + Gin 框架构建后端 API，Vue3 + Element Plus 构建前端界面，提供完善的用户权限管理和告警全生命周期管理功能。

**核心特性**：
- ✅ 完整的 RBAC 权限模型
- ✅ 动态菜单和路由生成
- ✅ JWT 认证机制
- ✅ 按钮级权限控制
- ✅ 多告警源 Webhook 接入（Alertmanager）
- ✅ 灵活的规则引擎（默认/时间抑制/AI 分析）
- ✅ 多通道消息发送（飞书/钉钉/自定义 Webhook）
- ✅ AI 辅助告警分析与建议
- ✅ 工作时间感知的告警抑制
- ✅ 告警流水全程追溯

---

## 🏗️ 项目架构

```
alert-ops/
├── cmd/                        # 应用程序入口
│   └── server/                 # 主服务程序
├── conf/                       # 配置文件
│   └── config.yaml             # 应用配置
├── docs/                       # Swagger 文档
├── internal/                   # 内部应用代码
│   ├── alert/                  # 告警模块
│   │   ├── adapter/            # 告警源适配器（Alertmanager 等）
│   │   ├── handler/            # HTTP 处理器层
│   │   ├── repo/               # 数据访问层（接口 + 实现）
│   │   └── service/            # 业务逻辑层（接口 + 实现）
│   ├── api/                    # 路由定义
│   ├── captcha/                # 验证码功能
│   ├── config/                 # 配置管理
│   ├── middleware/             # 中间件（JWT、CORS、权限）
│   ├── model/                  # 数据模型定义
│   ├── repo/                   # 全局数据访问（数据库初始化）
│   ├── scheduler/              # 定时任务
│   ├── user/                   # 用户模块
│   │   ├── handler/            # 用户/角色 HTTP 处理器
│   │   ├── repo/               # 用户/角色数据访问层
│   │   └── service/            # 用户/角色业务逻辑层
│   └── util/                   # 工具函数
├── pkg/                        # 可复用的公共包
│   ├── errno/                  # 错误处理
│   ├── gorm_logger/            # GORM 日志
│   ├── jwt/                    # JWT 认证
│   ├── password/               # 密码加密
│   └── response/               # 统一响应格式化
├── scripts/                    # 部署脚本（Dockerfile、docker-compose）
├── web/                        # 前端项目（Vue3 + Element Plus + Vite）
│   └── src/
│       ├── api/                # 前端 API 接口层
│       ├── assets/             # 静态资源
│       ├── components/         # 公共组件
│       ├── directives/         # 自定义指令（权限控制）
│       ├── router/             # 路由配置 + 动态路由
│       ├── stores/             # Pinia 状态管理
│       ├── utils/              # 工具函数
│       └── views/              # 页面组件
│           ├── alert/          # 告警管理页面（6 个）
│           └── system/         # 系统管理页面（3 个）
├── go.mod                      # Go 模块定义
└── README.md                   # 项目说明文档
```

### 分层架构设计

项目采用 **Handler → Service → Repo** 三层架构，每层通过接口解耦：

```
┌─────────────────────────────────────────────────┐
│  Handler 层    │  参数校验、请求响应、Swagger 注释  │
├─────────────────────────────────────────────────┤
│  Service 层    │  业务逻辑编排、接口抽象            │
├─────────────────────────────────────────────────┤
│  Repo 层       │  数据库 CRUD、接口抽象            │
└─────────────────────────────────────────────────┘
```

- **Repo 层**：接口 + 小写结构体实现，使用全局 `repo.DB`
- **Service 层**：接口 + 小写结构体实现，构造函数返回接口类型
- **Handler 层**：通过 Service 接口依赖注入，使用 `pkg/response` 统一响应

---

## 🛠️ 技术栈

### 后端技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| [Gin](https://github.com/gin-gonic/gin) | v1.12.0 | 高性能 Go Web 框架 |
| [GORM](https://gorm.io/) | v1.31.1 | Go 语言 ORM 库 |
| MySQL | 8.0+ | 关系型数据库 |
| [JWT](https://github.com/golang-jwt/jwt) | v5.3.1 | JSON Web Token 认证 |
| [Viper](https://github.com/spf13/viper) | v1.21.0 | 配置管理库 |
| [Zap](https://github.com/uber-go/zap) | v1.28.0 | 高性能日志库 |
| [Lumberjack](https://github.com/natefinch/lumberjack) | v2.2.1 | 日志滚动切割 |
| [Swagger](https://swaggo.github.io/swag/) | v1.16.6 | API 文档生成 |
| [bcrypt](https://golang.org/x/crypto) | v0.51.0 | 密码加密 |

### 前端技术栈

| 技术 | 版本 | 说明 |
|------|------|------|
| [Vue](https://vuejs.org/) | 3.5+ | 渐进式 JavaScript 框架 |
| [Element Plus](https://element-plus.org/) | 2.14+ | Vue 3 组件库 |
| [Pinia](https://pinia.vuejs.org/) | 3.0+ | Vue 3 官方状态管理 |
| [Vue Router](https://router.vuejs.org/) | 5.0+ | Vue 路由管理 |
| [Axios](https://axios-http.com/) | 1.16+ | HTTP 请求库 |
| [Vite](https://vitejs.dev/) | 8.0+ | 新一代前端构建工具 |
| [Monaco Editor](https://microsoft.github.io/monaco-editor/) | 4.7+ | 代码编辑器（模板编辑） |

---

## ✨ 核心功能

### 1. 用户管理
- ✅ 用户注册（支持验证码）
- ✅ 用户登录（JWT 认证）
- ✅ 用户列表（分页查询）
- ✅ 用户信息修改
- ✅ 用户删除
- ✅ 用户状态管理（启用/禁用）
- ✅ 密码修改

### 2. 角色管理
- ✅ 角色创建
- ✅ 角色列表查询（分页）
- ✅ 角色删除
- ✅ 角色权限分配
- ✅ 用户角色分配

### 3. 权限管理
- ✅ 权限创建（菜单权限、按钮权限）
- ✅ 权限列表查询
- ✅ 权限更新
- ✅ 权限删除
- ✅ 权限树形结构

### 4. 认证与授权
- ✅ JWT Token 认证
- ✅ 动态路由生成（基于用户权限）
- ✅ 动态菜单渲染
- ✅ 按钮级权限控制（`v-permission` 指令）
- ✅ 路由守卫权限校验
- ✅ API 级权限中间件

### 5. 告警源管理
- ✅ 告警源 CRUD
- ✅ 支持 Alertmanager Webhook 接入
- ✅ 告警源启用/禁用
- ✅ 自定义配置（JSON 格式）

### 6. 规则引擎
- ✅ 转发规则 CRUD
- ✅ **默认规则**：直接模板渲染 → 通道发送
- ✅ **时间规则**：工作时间判断 → 非工作时间自动抑制
- ✅ **AI 规则**：AI 分析告警 → 模板渲染（含 AI 建议）→ 通道发送
- ✅ 按标签匹配（match_labels）
- ✅ 规则优先级排序

### 7. 消息模板
- ✅ 模板 CRUD
- ✅ 支持变量替换（Go template 语法）
- ✅ 按通道类型和消息类型分类

### 8. 发送通道
- ✅ 通道 CRUD
- ✅ 飞书通道（交互式卡片 + HMAC-SHA256 签名）
- ✅ 钉钉通道（Markdown 消息 + 签名 + @提醒）
- ✅ 自定义 Webhook 通道（JSON POST）
- ✅ 通道启用/禁用

### 9. 告警流水
- ✅ 告警记录全程追溯
- ✅ 按状态/级别/时间范围筛选
- ✅ 分页查询
- ✅ 原始数据与格式化消息对照

### 10. AI 辅助分析
- ✅ 支持 OpenAI 兼容 API（OpenAI / CodeBuddy 等）
- ✅ AI 配置页面化管理，支持热切换（无需重启）
- ✅ 连接测试功能（一键验证配置可用性）
- ✅ AI 分析告警严重程度与根因
- ✅ AI 建议嵌入发送消息

### 11. 告警抑制
- ✅ 非工作时间告警自动抑制
- ✅ 定时任务（每 5 分钟）检查并发送
- ✅ 工作时间按规则配置（每条规则独立设置 `work_time_start` / `work_time_end`）

### 12. 系统功能
- ✅ 图片验证码（自实现）
- ✅ 配置热加载（Viper + fsnotify）
- ✅ 跨域支持（CORS）
- ✅ API 文档（Swagger）
- ✅ 前后端一体部署（后端托管 dist）

---

## 🚀 快速开始

### 环境要求

- **后端**：Go 1.25.0+
- **前端**：Node.js 18+
- **数据库**：MySQL 8.0+
- **包管理**：npm 或 yarn

---

### 后端启动

#### 1. 克隆项目

```bash
git clone <repository-url>
cd alert-ops
```

#### 2. 安装依赖

```bash
go mod download
```

#### 3. 配置数据库

创建 MySQL 数据库：

```sql
CREATE DATABASE alert_app CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

修改 `conf/config.yaml` 中的数据库配置：

```yaml
mysql:
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "your_password"
  dbname: "alert_app"
  max_idle_conns: 10
  max_open_conns: 100
  max_lifetime: "1h"
```

#### 4. 启动服务

```bash
# 使用默认配置文件
go run cmd/server/main.go

# 或指定配置文件
go run cmd/server/main.go /path/to/config.yaml
```

服务默认运行在 `http://localhost:8080`

#### 5. 默认管理员账号

```
用户名：admin
密码：admin123
```

拥有全部菜单和按钮权限。

#### 6. 访问 Swagger 文档

在开发模式下访问：`http://localhost:8080/swagger/index.html`

---

### 前端启动

#### 1. 进入前端目录

```bash
cd web
```

#### 2. 安装依赖

```bash
npm install
# 或
yarn install
```

#### 3. 启动开发服务器

```bash
npm run dev
# 或
yarn dev
```

前端默认运行在 `http://localhost:3000`

#### 4. 构建生产版本

```bash
npm run build
# 或
yarn build
```

构建产物在 `web/dist` 目录，可由后端静态文件服务托管。

---

## 📝 配置文件说明

配置文件位于 `conf/config.yaml`：

```yaml
app:
  name: "alert_app"
  mode: "dev"          # 运行模式：dev | prod | test
  port: 8080           # 服务端口
  version: "v0.0.1"
  start_time: "2026-01-01"
  machine_id: 1

log:
  level: "debug"       # 日志级别：debug | info | warn | error
  filename: "logs/alert_app.log"
  max_size: 200        # 单文件大小（MB）
  max_age: 30          # 保留天数
  max_backups: 7       # 最多保留文件数

mysql:
  host: "127.0.0.1"
  port: 3306
  username: "root"
  password: "123456"
  dbname: "alert_app"
  max_idle_conns: 10
  max_open_conns: 100
  max_lifetime: "1h"

jwt:
  secret: "your-secret-key"   # JWT 签名密钥
  expire: 24                  # Token 过期时间（小时）
```

> **说明**：AI 配置已迁移至数据库管理，通过管理界面动态配置（系统管理 → AI配置），无需在配置文件中设置。工作时间由每条告警规则独立配置。

---

## 🔧 API 接口说明

### 认证相关（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| GET | `/api/v1/auth/captcha` | 获取验证码 |

### Webhook 接收（公开）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/webhook/alertmanager/:slug` | 接收 Alertmanager Webhook |

### 用户相关（需要认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/user/info` | 获取当前用户信息 |
| GET | `/api/v1/user/menus` | 获取用户菜单 |
| POST | `/api/v1/user/change-password` | 修改密码 |
| GET | `/api/v1/users?page=1&page_size=10` | 获取用户列表（分页） |
| PUT | `/api/v1/users/:id` | 更新用户 |
| DELETE | `/api/v1/users/:id` | 删除用户 |
| PUT | `/api/v1/users/:id/status` | 更新用户状态 |
| POST | `/api/v1/users/:id/roles` | 分配用户角色 |
| DELETE | `/api/v1/users/:id/roles` | 删除用户角色 |
| GET | `/api/v1/users/:id/roles` | 获取用户角色 |

### 角色相关（需要认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/roles?page=1&page_size=10` | 获取角色列表（分页） |
| POST | `/api/v1/roles` | 创建角色 |
| DELETE | `/api/v1/roles/:id` | 删除角色 |
| GET | `/api/v1/roles/:id/permissions` | 获取角色权限 |
| POST | `/api/v1/roles/:id/permissions` | 分配角色权限 |

### 权限相关（需要认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/permissions` | 获取权限列表 |
| POST | `/api/v1/permissions` | 创建权限 |
| PUT | `/api/v1/permissions/:id` | 更新权限 |
| DELETE | `/api/v1/permissions/:id` | 删除权限 |

### 告警源管理（需要认证 + 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/alert-sources` | 告警源列表（分页） |
| POST | `/api/v1/alert-sources` | 创建告警源 |
| GET | `/api/v1/alert-sources/:id` | 告警源详情 |
| PUT | `/api/v1/alert-sources/:id` | 更新告警源 |
| DELETE | `/api/v1/alert-sources/:id` | 删除告警源 |

### 转发规则（需要认证 + 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/alert-rules` | 规则列表（按 source_id） |
| POST | `/api/v1/alert-rules` | 创建规则 |
| GET | `/api/v1/alert-rules/:id` | 规则详情 |
| PUT | `/api/v1/alert-rules/:id` | 更新规则 |
| DELETE | `/api/v1/alert-rules/:id` | 删除规则 |

### 消息模板（需要认证 + 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/alert-templates` | 模板列表（按 source_id） |
| POST | `/api/v1/alert-templates` | 创建模板 |
| GET | `/api/v1/alert-templates/:id` | 模板详情 |
| PUT | `/api/v1/alert-templates/:id` | 更新模板 |
| DELETE | `/api/v1/alert-templates/:id` | 删除模板 |

### 发送通道（需要认证 + 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/alert-channels` | 通道列表（按 source_id） |
| POST | `/api/v1/alert-channels` | 创建通道 |
| GET | `/api/v1/alert-channels/:id` | 通道详情 |
| PUT | `/api/v1/alert-channels/:id` | 更新通道 |
| DELETE | `/api/v1/alert-channels/:id` | 删除通道 |

### 告警流水（需要认证 + 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/alert-records` | 告警流水列表（分页+筛选） |
| POST | `/api/v1/alert-records/:id/analyze` | AI 分析单条告警记录 |

### AI 配置（需要认证 + 权限）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/ai-config` | 获取 AI 配置 |
| PUT | `/api/v1/ai-config` | 更新 AI 配置（热切换） |
| POST | `/api/v1/ai-config/test` | 测试 AI 连接（使用已保存配置） |

---

## 📊 数据模型

### 用户权限模块（5 张表）

| 模型 | 表名 | 说明 |
|------|------|------|
| User | users | 用户表 |
| Role | roles | 角色表 |
| UserRole | user_roles | 用户-角色关联表 |
| Permission | permissions | 权限表（菜单/按钮） |
| RolePermission | role_permissions | 角色-权限关联表 |

### 告警模块（8 张表）

| 模型 | 表名 | 说明 |
|------|------|------|
| AlertSource | alert_sources | 告警源 |
| AlertRule | alert_rules | 转发规则 |
| RuleChannel | rule_channels | 规则-通道关联表 |
| AlertTemplate | alert_templates | 消息模板 |
| AlertChannel | alert_channels | 发送通道 |
| AlertRecord | alert_records | 告警流水记录 |
| SuppressedAlert | suppressed_alerts | 抑制告警队列 |
| AIConfig | ai_configs | AI 配置（单记录，ID=1） |

---

## 🔄 告警处理流程

```
外部 Alertmanager → Webhook 接收 → 适配器解析 → 规则引擎
    │
    ├── 加载告警源下的启用规则
    ├── 按 match_labels 匹配规则
    ├── 按优先级排序执行规则
    │   ├── default 规则 → 模板渲染 → 多通道发送
    │   ├── time 规则 → 工作时间判断 → 直接发送 / 抑制到上班
    │   └── ai 规则 → AI 分析 → 模板渲染(含AI建议) → 通道发送
    └── 无匹配规则 → 默认模板 → 发送到所有通道
```

**支持的通道类型**：
- **飞书 (feishu)**：交互式卡片消息 + HMAC-SHA256 签名验证
- **钉钉 (dingtalk)**：Markdown 消息 + 签名验证 + @提醒（自动处理换行兼容）
- **自定义 Webhook (webhook)**：JSON POST 请求

**支持的 AI 提供商**：
- OpenAI 兼容 API（支持任何兼容 OpenAI 接口的服务）

---

## 🔐 权限控制

项目采用 **RBAC（Role-Based Access Control）** 权限模型，实现细粒度的权限管理。

### 权限模型

```
用户（User） ←→ 角色（Role） ←→ 权限（Permission）
```

### 权限类型

| 类型 | 说明 | 是否生成路由 | 是否显示菜单 |
|------|------|--------------|--------------|
| `menu` | 菜单权限 | ✅ 是 | ✅ 是 |
| `button` | 按钮权限 | ❌ 否 | ❌ 否 |

### 前端权限控制

- **动态路由生成**：登录后根据用户权限菜单动态添加路由
- **动态菜单渲染**：根据用户权限构建菜单树
- **按钮权限控制**：`v-permission` 指令控制按钮显隐
- **路由守卫**：路由跳转前校验用户权限

### 后端权限控制

- **JWT 认证中间件**：验证 Token 有效性
- **API 权限中间件**：基于 Permission 表的 `api_path` + `api_method` 匹配校验
- **公开路由白名单**：登录、注册、验证码、Webhook 接收

---

## ⏰ 定时任务

| 任务 | 说明 | 频率 |
|------|------|------|
| SuppressedSender | 检查并发送被抑制的告警（仅工作时间，通过每条规则的 `work_time_start/end` 配置） | 每 5 分钟 |

---

## 🛠️ 开发常用命令

```bash
# 生成 Swagger 文档
go run github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go -o docs

# 运行后端（开发模式）
go run cmd/server/main.go

# 运行前端（开发模式）
cd web && npm run dev

# 后端测试
go test ./...

# 代码格式化
go fmt ./...
go vet ./...

# 构建生产版本
go build -o alert-ops cmd/server/main.go
```

---

## 📦 部署说明

### 后端部署

#### 1. 编译二进制文件

```bash
# 本地编译
go build -o alert-ops cmd/server/main.go

# Linux 交叉编译
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o alert-ops cmd/server/main.go
```

#### 2. 上传配置文件和二进制文件到服务器

```bash
scp alert-ops user@server:/app/
scp conf/config.yaml user@server:/app/conf/
```

#### 3. 运行服务

```bash
cd /app
./alert-ops ./conf/config.yaml
```

建议使用 `systemd` 或 `supervisor` 管理进程。

---

### 前端部署

#### 1. 构建生产版本

```bash
cd web
npm run build
```

#### 2. 部署到 Web 服务器

**方案 A：由后端托管**

构建后的文件在 `web/dist` 目录，后端已配置静态文件服务：

```go
// 静态文件服务 - 托管前端构建产物
r.Static("/assets", "./web/dist/assets")
r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

// SPA 回退：非 API 路由返回 index.html
r.NoRoute(func(c *gin.Context) {
  c.File("./web/dist/index.html")
})
```

**方案 B：使用 Nginx 托管**

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/web/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
    }
}
```

### Docker 部署

项目包含 `Dockerfile` 和 `docker-compose.yml`，可直接使用：

```bash
cd scripts
docker-compose up -d
```

> **说明**：`docker-compose.yml` 中 MySQL 数据库名默认为 `alert_app`，与 `conf/config.yaml` 中的 `mysql.dbname` 保持一致。

---

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...

# 生成 Swagger 文档
go run github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go -o docs
```

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 贡献流程

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

---

## 📮 联系方式

如有问题或建议，请提交 Issue 或联系项目维护者。

---

## 📚 参考资料

- [Gin 文档](https://gin-gonic.com/zh-cn/docs/)
- [Vue 3 文档](https://cn.vuejs.org/)
- [Element Plus 文档](https://element-plus.org/zh-CN/)
- [GORM 文档](https://gorm.io/zh_CN/docs/)
- [Pinia 文档](https://pinia.vuejs.org/zh/)

---

**注意**：本项目已完成 RBAC 权限管理与告警运营核心功能，可作为告警管理平台直接使用。
