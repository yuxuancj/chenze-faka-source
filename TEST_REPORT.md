# 晨泽发卡系统 - 测试覆盖率报告

## 第二轮覆盖率提升（中间件/模型/路由/Service层补全）

**生成时间**: 2026-08-07  
**测试框架**: Go testing + testify/assert + SQLite 内存数据库 + httptest  
**测试文件总数**: 54 个 `*_test.go` 文件

---

## 一、整体覆盖率总览

| 包 | 覆盖率 | 状态 |
|----|--------|------|
| `cmd` | 60.3% | ✅ |
| `internal/controller` | 68.9% | ✅ |
| `internal/middleware` | **100.0%** | ✅ |
| `internal/model` | **100.0%** | ✅ |
| `internal/pkg/captcha` | **100.0%** | ✅ |
| `internal/pkg/crypto` | 88.1% | ✅ |
| `internal/pkg/database` | 79.5% | ✅ |
| `internal/pkg/response` | **100.0%** | ✅ |
| `internal/pkg/utils` | 98.0% | ✅ |
| `internal/router` | 93.8% | ✅ |
| `internal/service` | 88.9% | ✅ |
| **整体** | **83.1%** | ✅ ≥ 80% |

---

## 二、核心模块覆盖率详情

### 1. 中间件（100%）

| 文件 | 函数 | 覆盖率 |
|------|------|--------|
| auth_middleware.go | NewAuthMiddleware | 100% |
| | AuthRequired | 100% |
| | AdminRequired | 100% |
| install_middleware.go | NewInstallMiddleware | 100% |
| | Handle | 100% |
| | isInstalled | 100% |
| | isAllowedWhenNotInstalled | 100% |
| | isAPIRequest | 100% |
| license_middleware.go | NewLicenseMiddleware | 100% |
| | Handle | 100% |
| | isLicensePublicPath | 100% |

### 2. 模型（100%）

所有 `TableName()` 方法、`OrderStatusText()`、`ToVO()` 均已覆盖。

### 3. 路由（93.8%）

- `Setup()` — 路由注册测试 ✅
- `serveEmbeddedFile()` — 文件服务测试 ✅
- `detectContentType()` — MIME 类型检测 ✅

### 4. Service 层（88.9%）

| 服务 | 关键函数覆盖率 |
|------|--------------|
| auth_service.go | Register 92.9%, Login 90.0%, ParseToken 88.9% |
| order_service.go | CreateOrder 95.7%, HandlePaymentCallback 95.7%, VerifyNotify 100% |
| product_service.go | Create 91.7%, Update 95.0%, ListOnShelfGrouped 94.1% |
| card_service.go | ImportCards 68.8%, GetAvailableCards 90.9%, ExportCards 95.8% |
| category_service.go | Create 87.5%, Update 92.9%, GetAll 100% |
| payment_service.go | Create 87.5%, Update 93.8%, GetActive 100% |
| email_service.go | Create 88.9%, Update 95.0%, TestConnection 100% |
| node_service.go | Create 87.5%, Update 92.9%, GetBestNode 100% |
| license_service.go | QuickVerify 90.9%, IsVerified 100%, GetCache 100% |
| dashboard_service.go | GetStats 100%, GetOrderStatusCounts 100% |
| log_service.go | WriteOperation 100%, WriteOrder 100%, ListOperation 89.5% |
| upload_service.go | SaveFile 77.3%, GetFile 100%, determineFileType 100% |
| upgrade_service.go | GetVersion 100%, CheckUpdate 100%, UploadPackage 100% |

### 5. Controller 层（68.9%）

| 控制器 | 关键端点覆盖率 |
|--------|--------------|
| auth_controller.go | Login 100%, Register 100%, GetCaptcha 100%, GetProfile 100% |
| order_controller.go | Create 100%, Query 100%, Notify 100%, Return 100% |
| card_controller.go | Import 84.6%, List 77.8%, Delete 75.0% |
| product_controller.go | List 75.0%, OnShelf 60.0%, GetByID 77.8% |
| admin_controller.go | SystemStatus 100%, GetSettings 100%, OrderStatusCounts 60% |

### 6. pkg 工具包

| 模块 | 覆盖率 |
|------|--------|
| crypto (AES) | 88.1% |
| captcha (验证码) | 100% |
| response (响应封装) | 100% |
| utils (工具函数) | 98.0% |
| database (数据库) | 79.5% |

---

## 三、未覆盖代码说明

| 模块 | 未覆盖函数 | 原因说明 |
|------|-----------|----------|
| database.go | `WipeTables` (45.5%) | `DROP TABLE` 在 SQLite 内存数据库中行为不同；MySQL 特定分支无法在单元测试中覆盖 |
| crypto/aes.go | `AesEncrypt` (83.3%) | 加密块大小相关边界条件难以在测试中精确触发 |
| email_service.go | `send` (32.3%) | 真实 SMTP 网络发送涉及 TCP 连接、TLS 握手，无法在单元测试中完整覆盖 |
| node_service.go | `Ping` (已覆盖) | Mock HTTP 服务器已覆盖成功/失败/超时路径 |
| admin_controller.go | `Dashboard`/`OrderStatusCounts` (60%) | 聚合查询的复杂 SQL 分支 |
| cmd/main.go | `main` (0%) | 程序入口，涉及进程生命周期管理 |

---

## 四、测试文件清单

### 新增测试文件（第二轮）

| 文件 | 类型 |
|------|------|
| `internal/middleware/auth_middleware_test.go` | 中间件单元测试 |
| `internal/middleware/install_middleware_test.go` | 中间件单元测试 |
| `internal/middleware/license_middleware_test.go` | 中间件单元测试 |
| `internal/model/model_test.go` | 模型单元测试 |
| `internal/model/upload_test.go` | 模型单元测试 |
| `internal/router/router_test.go` | 路由集成测试 |
| `internal/service/log_service_test.go` | Service 单元测试 |
| `internal/service/dashboard_service_test.go` | Service 单元测试 |
| `internal/service/category_service_supplement_test.go` | Service 补充测试 |
| `internal/service/payment_service_test.go` | Service 单元测试 |
| `internal/service/node_service_test.go` | Service 单元测试 |
| `internal/service/email_service_test.go` | Service 单元测试 |
| `internal/service/upload_service_test.go` | Service 单元测试 |
| `internal/service/license_service_test.go` | Service 单元测试 |
| `internal/service/upgrade_service_test.go` | Service 单元测试 |
| `internal/service/email_service_supplement_test.go` | Service 补充测试 |
| `internal/service/node_service_supplement_test.go` | Service 补充测试 |
| `internal/service/license_service_gap_test.go` | Service 补充测试 |
| `internal/controller/admin_controller_supplement_test.go` | Controller 集成测试 |
| `internal/controller/admin_controller_gap_test.go` | Controller 集成测试 |
| `internal/controller/product_controller_supplement_test.go` | Controller 补充测试 |
| `internal/pkg/database/database_supplement_test.go` | pkg 单元测试 |
| `cmd/main_test.go` | 入口单元测试 |

---

## 五、执行结果

```bash
$ go test ./... -coverprofile=coverage.out -covermode=atomic

ok   chenze-faka/cmd                      5.066s   coverage: 60.3%
ok   chenze-faka/internal/controller      1.094s   coverage: 68.9%
ok   chenze-faka/internal/middleware      0.061s   coverage: 100.0%
ok   chenze-faka/internal/model           0.017s   coverage: 100.0%
ok   chenze-faka/internal/pkg/captcha     0.061s   coverage: 100.0%
ok   chenze-faka/internal/pkg/crypto      0.016s   coverage: 88.1%
ok   chenze-faka/internal/pkg/database    0.097s   coverage: 79.5%
ok   chenze-faka/internal/pkg/response   0.036s   coverage: 100.0%
ok   chenze-faka/internal/pkg/utils       0.022s   coverage: 98.0%
ok   chenze-faka/internal/router          0.082s   coverage: 93.8%
ok   chenze-faka/internal/service        21.485s  coverage: 88.9%
```

**全部通过 ✓ | 无跳过 ✓ | 无 panic ✓ | 整体覆盖率 83.1% ≥ 80% ✓**
