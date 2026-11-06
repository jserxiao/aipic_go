# AI 绘图小程序后端 — Go 项目进度文档

> **项目路径**: `d:/my/aipic_go`
> **对应小程序项目**: `d:/my/aipic`
> **最后更新**: 2026-11-03

---

## 一、项目概述

基于小程序项目 `aipic` 的接口需求，使用 Go 语言构建后端 API 服务。

### 技术栈

| 类别 | 技术选型 |
|------|---------|
| 语言 | Go 1.21+ |
| Web 框架 | Gin |
| 数据库 | SQLite (modernc.org/sqlite，纯 Go 无需 CGO) |
| 认证 | JWT (golang-jwt/jwt/v5)，区分 access/refresh token 类型 |
| 日志 | 请求日志中间件 + Go 标准库 log |
| 数据库驱动 | modernc.org/sqlite |
| UUID | github.com/google/uuid |

### 统一响应格式

```json
{
  "code": 0,
  "data": {},
  "message": "success"
}
```

---

## 二、项目目录结构

```
aipic_go/
├── main.go                          # 程序入口
├── go.mod / go.sum                  # Go 模块定义
├── .gitignore                       # Git 忽略配置
├── PROGRESS.md                      # 本文档（进度跟踪）
│
├── internal/
│   ├── config/
│   │   └── config.go                # 全局配置（含生产环境安全检查）
│   │
│   ├── model/
│   │   └── models.go                # 数据模型定义
│   │
│   ├── store/
│   │   ├── db.go                    # 数据库初始化 & 迁移
│   │   └── queries.go               # 数据库操作函数（含事务）
│   │
│   ├── middleware/
│   │   ├── middleware.go            # 认证、CORS 中间件
│   │   ├── ratelimit.go             # 限流中间件（令牌桶）
│   │   └── logger.go                # 请求日志 + RequestID 中间件
│   │
│   ├── service/
│   │   ├── auth.go                  # 认证服务（微信登录、JWT）
│   │   ├── user.go                  # 用户信息服务
│   │   ├── generate.go              # 图片生成服务（异步任务）
│   │   ├── history.go               # 历史记录服务
│   │   ├── upload.go                # 文件上传服务（magic bytes 校验）
│   │   └── payment.go               # 支付服务（充值、VIP、事务确认）
│   │
│   ├── handler/
│   │   ├── auth.go                  # 认证处理器
│   │   ├── user.go                  # 用户处理器
│   │   ├── generate.go              # 生成处理器（含越权校验）
│   │   ├── upload.go                # 上传处理器
│   │   └── payment.go               # 支付处理器（回调验签）
│   │
│   ├── router/
│   │   └── router.go                # 路由配置（含限流）
│   │
│   └── utils/
│       ├── response.go              # 统一响应封装
│       ├── jwt.go                   # JWT 工具（区分 access/refresh）
│       └── id.go                    # ID 生成工具
```

---

## 三、接口清单与实现进度

### 3.1 认证模块

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 微信登录 | POST | `/api/auth/login` | ✅ 已完成 | 接收 wx code，返回 token + 用户信息；严格限流 |
| 刷新 Token | POST | `/api/auth/refresh` | ✅ 已完成 | 用 refreshToken 换取新 token；严格限流 |
| 退出登录 | POST | `/api/auth/logout` | ✅ 已完成 | 客户端清除 token |

### 3.2 用户模块

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 获取用户信息 | GET | `/api/user/info` | ✅ 已完成 | 返回用户信息（含积分、VIP） |
| 更新头像 | PUT | `/api/user/avatar` | ✅ 已完成 | 传入 avatar URL，含长度校验 |

### 3.3 图片生成模块

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 提交生成任务 | POST | `/api/generate/submit` | ✅ 已完成 | 异步任务，含参数白名单校验、先扣积分后创建 |
| 查询任务进度 | GET | `/api/generate/task/:taskId` | ✅ 已完成 | 带越权校验，只能查自己的任务 |
| 取消生成任务 | POST | `/api/generate/task/:taskId/cancel` | ✅ 已完成 | 带越权校验，只能取消自己的任务 |

### 3.4 历史记录模块

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 获取历史记录 | GET | `/api/generate/history` | ✅ 已完成 | 分页查询，pageSize 上限 100 |
| 删除单条记录 | DELETE | `/api/generate/history/:recordId` | ✅ 已完成 | 带越权校验 |
| 清空所有记录 | DELETE | `/api/generate/history` | ✅ 已完成 | 清空当前用户所有记录 |

### 3.5 图片上传模块

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 上传图片 | POST | `/api/upload/image` | ✅ 已完成 | 含扩展名 + magic bytes 双重校验 |

### 3.6 支付模块

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 创建充值订单 | POST | `/api/payment/recharge/create` | ✅ 已完成 | 金额上限 10000 元 |
| 创建 VIP 订单 | POST | `/api/payment/vip/create` | ✅ 已完成 | 套餐白名单校验 |
| 查询订单状态 | GET | `/api/payment/order/:orderId` | ✅ 已完成 | 带越权校验，只能查自己的订单 |
| 微信支付回调 | POST | `/api/payment/notify` | ✅ 已完成 | 开发模式模拟回调，生产模式需验签 |

### 3.7 其他

| 接口 | 方法 | 路径 | 状态 | 说明 |
|------|------|------|------|------|
| 健康检查 | GET | `/health` | ✅ 已完成 | 服务健康检查 |
| 静态文件 | GET | `/uploads/*` | ✅ 已完成 | 上传文件访问 |

---

## 四、安全审计记录（v1.1.0）

> 以下为安全审计发现并修复的问题清单

### ✅ 已修复安全问题

| # | 严重度 | 问题 | 修复方案 |
|---|--------|------|---------|
| 1 | 🔴 严重 | JWT access/refresh token 无类型区分，可用 access 刷新或用 refresh 访问 API | 在 Claims 中增加 `type` 字段，`ParseToken` 只接受 access，`ParseRefreshToken` 只接受 refresh |
| 2 | 🔴 严重 | 支付回调接口 `/payment/confirm/:orderId` 完全无鉴权，任何人可伪造支付成功 | 改为 `/payment/notify`，开发模式仅用于测试，生产模式必须验签 |
| 3 | 🔴 严重 | JWT 默认密钥硬编码 `aipic_default_secret_key` | 生产环境（非 DevMode）强制检查：密钥不能为空、不能为默认值、长度≥32 |
| 4 | 🔴 严重 | 确认支付非原子操作：先更新订单状态再增加积分，中间可能失败 | 新增 `ConfirmPaymentInTransaction`，事务内完成订单更新+积分增加+VIP更新 |
| 5 | 🟠 高 | 生成任务扣积分与创建任务非原子，任务创建失败后积分不退还 | 调整顺序：先扣积分（事务安全），再创建任务；失败时退还积分 |
| 6 | 🟠 高 | 登录 code 无长度限制 | 增加 code 长度校验（≤256 字符）和 refreshToken 长度校验（≤2048） |
| 7 | 🟠 高 | 登录接口无频率限制，可被刷接口 | 新增 `StrictRateLimit`：登录/刷新限流每分钟 10 次/IP |
| 8 | 🟠 高 | 充值金额无上限，可创建巨额订单 | 新增金额校验：最低 1 元，最高 10000 元 |
| 9 | 🟠 高 | CORS 配置 `Origin: *` 不安全 | 生产模式只允许配置的 `CORS_ORIGIN`，开发模式才用 `*` |
| 10 | 🟠 高 | 生成任务查询/取消无越权校验 | `GetTaskProgress` 和 `CancelGenerateTask` 增加 userID 校验 |
| 11 | 🟠 高 | 订单查询无越权校验 | `QueryOrderStatus` 改为按 orderID + userID 联合查询 |
| 12 | 🟠 高 | 删除历史记录无越权校验 | `DeleteGenerateRecord` 增加 userID 校验 |
| 13 | 🟡 中 | JWT 未校验签名算法，可能被 alg=none 攻击 | `parseTokenClaims` 中检查 `token.Method` 必须为 HMAC 类型 |
| 14 | 🟡 中 | 上传文件仅校验扩展名，可上传伪装文件 | 新增 `http.DetectContentType` 校验真实文件头（magic bytes） |
| 15 | 🟡 中 | 全局无限流中间件 | 新增 `NormalRateLimit`：普通接口每分钟 120 次/IP |
| 16 | 🟡 中 | 生产模式未禁止 mock 登录 | 配置层强制检查：生产模式必须配置 WX_APP_ID/SECRET/MCH 信息 |
| 17 | 🟡 中 | token 长度无上限，可能超长 token 攻击 | Auth 中间件限制 token 长度 ≤2048 |
| 18 | 🟡 中 | VIP 到期时间未考虑已有 VIP 的叠加续费 | `calculateVipExpireDate` 检查当前 VIP 是否未过期，未过期则在此基础上续费 |
| 19 | 🟡 中 | 微信登录返回 body 无大小限制 | `getWxOpenID` 使用 `io.LimitReader` 限制 1MB |
| 20 | 🟡 中 | 头像 URL 无长度限制 | `UpdateAvatar` 校验 URL 长度 ≤2048 |
| 21 | 🟢 低 | 缺少请求日志和 RequestID | 新增 `RequestLogger` 中间件，生成 UUID 并输出请求日志 |
| 22 | 🟢 低 | 缺少清空历史记录接口 | 新增 `DELETE /generate/history` 接口 |
| 23 | 🟢 低 | ratio/model 参数未做白名单校验 | `SubmitTask` 增加 `isValidRatio` 和 `isValidModel` 校验 |
| 24 | 🟢 低 | `GetUserID` fallback 不安全 | 移除 fallback 到 WxAppID 的逻辑，改为返回空字符串 |
| 25 | 🟢 低 | 分页 pageSize 无上限 | `GetHistory` 校验 pageSize ≤100 |

---

## 五、数据模型

### 5.1 数据库表

| 表名 | 说明 |
|------|------|
| `users` | 用户信息 |
| `generate_tasks` | 生成任务（异步任务跟踪） |
| `generate_records` | 生成记录（历史记录） |
| `payment_orders` | 支付订单 |

### 5.2 VIP 套餐配置

| 套餐 | 类型 | 价格 | 赠送积分 | 时长 | 说明 |
|------|------|------|---------|------|------|
| 月卡 | monthly | ¥49 | 1000 | 1月 | 每月自动续费 |
| 季卡 | quarterly | ¥139 | 3000 | 3月 | 每季自动续费 |
| 年卡 | yearly | ¥399 | 10000 | 12月 | 每年自动续费 |

### 5.3 充值规则

- 1 元 = 10 积分
- 自定义充值金额范围：1 ~ 10000 元

### 5.4 AI 模型配置

| 模型 | 标识 | 消耗积分 |
|------|------|---------|
| Qwen-Image-3.0 | qwen-image-3.0 | 18 |

---

## 六、环境配置

### 环境变量

| 变量名 | 默认值 | 必填 | 说明 |
|--------|--------|------|------|
| `PORT` | 8080 | 否 | 服务端口 |
| `DB_PATH` | ./data/aipic.db | 否 | SQLite 数据库路径 |
| `JWT_SECRET` | (开发模式有默认) | **生产必填** | JWT 签名密钥，长度≥32 |
| `JWT_EXPIRE_HOURS` | 2 | 否 | JWT 过期时间（小时） |
| `REFRESH_EXPIRE_HOURS` | 168 | 否 | RefreshToken 过期时间（小时） |
| `UPLOAD_DIR` | ./uploads | 否 | 上传文件目录 |
| `MAX_UPLOAD_SIZE` | 5242880 | 否 | 最大上传大小（字节） |
| `WX_APP_ID` | (空) | **生产必填** | 微信小程序 AppID |
| `WX_APP_SECRET` | (空) | **生产必填** | 微信小程序 AppSecret |
| `WX_MCH_ID` | (空) | **生产必填** | 微信支付商户号 |
| `WX_MCH_KEY` | (空) | **生产必填** | 微信支付 API 密钥 |
| `WX_MCH_CERT_SERIAL` | (空) | **生产必填** | 微信支付证书序列号 |
| `WX_MCH_PRIVATE_KEY_PATH` | (空) | **生产必填** | 微信支付商户私钥路径 |
| `WX_NOTIFY_URL` | (空) | **生产必填** | 微信支付回调通知地址 |
| `CORS_ORIGIN` | (空) | **生产必填** | 允许的跨域来源 |

### 开发模式 vs 生产模式

**开发模式**（`WX_APP_ID` 或 `WX_APP_SECRET` 为空时自动启用）：
- 使用默认 JWT 密钥
- 任意 code 都能登录（Mock 登录）
- 自动创建测试用户（userId: u_001，1000 积分，VIP）
- 生成任务使用 Mock 异步模拟（5 秒完成）
- 支付参数返回 Mock 数据，回调可直接确认
- CORS 允许所有来源

**生产模式**（`WX_APP_ID` 和 `WX_APP_SECRET` 均已配置）：
- 强制检查 JWT_SECRET 不能为默认值、长度≥32
- 强制检查微信支付配置（MCH_ID、MCH_KEY、NOTIFY_URL）
- 调用真实微信 `jscode2session` 接口
- 支付回调必须验签
- CORS 只允许配置的来源
- Gin 设为 release 模式

### 启动方式

```bash
# 开发模式（Mock）
go run main.go

# 生产模式
WX_APP_ID=wx_xxx \
WX_APP_SECRET=xxx \
WX_MCH_ID=xxx \
WX_MCH_KEY=xxx \
WX_NOTIFY_URL=https://api.example.com/api/payment/notify \
JWT_SECRET=your_32_char_secret_key_here \
CORS_ORIGIN=https://your-frontend.com \
go run main.go

# 编译
go build -o aipic_go.exe .

# 运行编译产物
./aipic_go.exe
```

---

## 七、安全设计说明

### 7.1 JWT 双 Token 机制

- **Access Token**：有效期 2 小时，用于 API 访问
- **Refresh Token**：有效期 7 天，用于刷新 access token
- 两者在 Claims 中通过 `type` 字段区分，不可互换使用
- 签名算法强制为 HMAC（HS256），拒绝 `alg=none` 攻击

### 7.2 越权防护

所有涉及用户资源的接口均做了越权校验：
- 查询/取消生成任务：校验 `task.UserID == 当前用户`
- 查询订单状态：按 `orderID + userID` 联合查询
- 删除历史记录：`WHERE record_id = ? AND user_id = ?`

### 7.3 支付安全

- 支付回调接口 `/payment/notify` 在生产环境**必须验签**
- 支付确认使用数据库事务，原子完成：更新订单状态 → 增加积分 → 更新 VIP
- 订单状态非 `pending` 时拒绝重复确认（防重放攻击）
- 充值金额上限 10000 元

### 7.4 限流策略

| 接口类型 | 限流规则 |
|---------|---------|
| 登录/刷新 | 每分钟 10 次/IP |
| 普通接口 | 每分钟 120 次/IP |

### 7.5 文件上传安全

- 扩展名白名单：jpg/jpeg/png/gif/webp/bmp
- magic bytes 校验：`http.DetectContentType` 检查真实文件类型
- 文件大小限制：5MB
- 文件名不使用原始名称，使用时间戳+userID 生成

---

## 八、TODO（待实现 / 优化）

### P0 — 上线前必须完成

- [ ] **对接真实微信支付**：实现统一下单、支付回调验签（V3 API + 平台证书）
- [ ] **对接 AI 生图模型**：接入通义千问 `qwen-image-3.0` 模型，替换 Mock 生成逻辑
- [ ] **生产环境数据库**：从 SQLite 迁移到 MySQL/PostgreSQL
- [ ] **日志系统**：替换标准库 log 为 zap/zerolog，支持结构化日志和文件轮转
- [ ] **HTTPS**：配置 TLS 证书，生产环境必须使用 HTTPS

### P1 — 中优先级

- [ ] **Redis 任务队列**：使用 Redis 替代内存中的任务进度管理，支持多实例部署
- [ ] **OSS/CDN 文件存储**：上传图片替换为阿里云 OSS / 腾讯云 COS
- [ ] **配置文件**：使用 yaml/viper 管理配置，替代纯环境变量
- [ ] **Token 黑名单**：退出登录时将 token 加入 Redis 黑名单
- [ ] **敏感操作日志**：记录积分变动、支付操作等审计日志

### P2 — 低优先级

- [ ] **多图生成支持**：一次生成 4 张图片
- [ ] **图生图模式**：参考图 + 文字描述生成
- [ ] **社区分享功能**：生成结果分享到社区
- [ ] **更多风格预设**：油画、水彩、动漫等风格
- [ ] **API 文档自动生成**：使用 Swagger/OpenAPI 生成接口文档
- [ ] **单元测试**：编写 handler 和 service 层的测试用例
- [ ] **Docker 部署**：提供 Dockerfile 和 docker-compose 配置

---

## 九、迭代记录

### v1.1.0 — 2026-11-03（安全审计修复）

**安全修复**：
- JWT 区分 access/refresh token 类型，防止类型混用
- 支付回调改为 `/payment/notify`，生产模式必须验签
- 生产环境强制检查 JWT 密钥（不能为默认值，长度≥32）
- 支付确认改为事务操作（原子完成订单更新+积分增加+VIP更新）
- 生成任务先扣积分后创建，失败时退还积分
- 全部资源接口增加越权校验（任务、订单、历史记录）
- 新增限流中间件：登录严格限流，普通接口全局限流
- CORS 生产环境收紧，只允许配置的来源
- JWT 强制 HMAC 签名算法，拒绝 alg=none 攻击
- 文件上传增加 magic bytes 真实类型校验
- token 长度限制，防止超长 token 攻击
- VIP 续费支持叠加（已有 VIP 未过期时在到期时间基础上续费）

**新增功能**：
- 请求日志中间件 + RequestID（UUID）
- 清空历史记录接口 `DELETE /generate/history`
- ratio/model 参数白名单校验
- 分页 pageSize 上限 100

**配置变更**：
- 新增环境变量：`WX_MCH_ID`、`WX_MCH_KEY`、`WX_MCH_CERT_SERIAL`、`WX_MCH_PRIVATE_KEY_PATH`、`WX_NOTIFY_URL`、`CORS_ORIGIN`
- JWT 默认过期时间从 24 小时缩短为 2 小时
- 生产环境启动时强制检查所有必填配置

### v1.0.0 — 2026-11-03（初始版本）

- 项目骨架搭建（Gin + SQLite + JWT）
- 实现全部接口（认证、用户、生成、历史、上传、支付）
- 数据库自动迁移
- 开发模式 Mock 支持
- 统一响应格式
- 中间件（认证、跨域）
- 静态文件服务

---

## 十、修改日志

> 所有后续修改记录在此，格式：`日期 — 修改内容`

- **2026-11-03 v1.2.0**: 小程序开发版联调支持
  - 新增 `.env` 文件支持（`main.go` 中 `loadEnvFile` 自动加载）
  - 升级请求日志中间件，完整打印每个接口的请求体和响应体（JSON 美化格式）
  - 新增 `POST /api/payment/mock-confirm` 接口（仅开发模式，模拟支付确认）
  - 修改小程序 `config/index.ts`：关闭 Mock 模式，`TARO_APP_API_BASE` 指向 `http://localhost:8080/api`
  - 修改小程序 `services/payment.ts`：检测 mock 支付参数自动走模拟确认流程
  - 修改小程序 `pages/mine/index.tsx`：VIP 支付同样支持 mock 确认
  - 修改小程序 `store/generateStore.ts`：非 Mock 模式从服务端加载历史记录
  - 全部 10 组接口联调测试通过
- **2026-11-03 v1.1.0**: 安全审计修复，修复 25 项安全和逻辑问题，新增限流/日志/越权校验/事务安全
- **2026-11-03 v1.0.0**: 初始创建项目，实现全部接口，修复生成记录重复保存问题
