<div align="center">

# Cline2API

Cline API 反向代理 · 多账号轮询 · 三协议兼容 · 动态模型同步 · 桌面端

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](#构建)

**🌐 English: [English README](README.en.md)**

</div>

---

## 简介

Cline2API 是 Cline API 的反向代理服务，支持多账号轮询、OpenAI Chat / Anthropic Messages / OpenAI Responses 三协议、API Key 鉴权，内置中英文管理后台（自动跟随浏览器语言，可手动切换）。提供跨平台桌面端单文件应用（Windows / macOS / Linux），双击即用。

**开发语言**：Go（后端 + 代理 + 桌面壳），HTML/CSS/JS（管理后台前端，内嵌于二进制）。

## 核心功能

- **三协议兼容**：同时支持 `/v1/chat/completions`（OpenAI Chat）、`/v1/messages`（Anthropic Messages API）与 `/v1/responses`（OpenAI Responses API，全面兼容 Cherry Studio 等客户端的流式与输入解析）
- **多账号轮询**：自动在多个 Cline 账号间切换负载（`round_robin` / `fill` / `random` 策略）
- **中英文管理后台**：浏览器访问 `/admin/` 管理账号、API Key、模型配置、请求头、代理设置；自动跟随浏览器语言，侧栏可手动切换
- **双引擎模型动态同步**：
  - **从 Cline 同步模型**：启动时自动拉取 Cline 官方推荐模型接口（免费 / cline-pass / 推荐模型），模型变化时弹窗提示，也可在后台手动点击同步
  - **从 opencode 同步模型**：启动后后台定时（每 10 分钟）自动拉取 opencode 开放模型，自动对齐最新免费与付费模型，支持手动同步
- **正向出口代理**：
  - **Cline 出口代理**：在后台「代理配置」直接配置发往 `api.cline.bot` 的 HTTP / HTTPS / SOCKS5 代理（支持账密认证），支持一键连通性探测并即时热更新
  - **opencode 代理池**：支持配置代理列表轮询出口，并具备限流自动冷却机制
- **自定义模型**：后台可手动添加/删除模型 ID，并自由选择默认模型（未设置时自动回退到第一个免费模型）
- **API Key 鉴权**：保护代理端点，支持生成/删除多个 API Key
- **System Prompt 覆盖**：在数据目录下放 `override.md` 则自动替换系统提示词
- **账号导入/导出**：支持 OAuth 登录、手动 Token、批量文件导入，以及跨设备导出
- **请求日志**：记录每次请求的 token 用量、耗时、TPS 等指标
- **桌面端**：单文件跨平台桌面应用（Wails v2），关闭窗口即停止服务
[opencode-performance-monitor](../../Plugins/opencode-performance-monitor)
## 快速开始

### 方式一：桌面端（推荐，分享给他人）

从 [Releases](https://github.com/luawei1/cline2api/releases) 下载对应平台的可执行文件，双击运行即可。

> Windows 提示 SmartScreen「已保护你的电脑」是**未购买代码签名证书的正常现象**，
> 点击「更多信息 → 仍要运行」即可，不影响使用。

| 平台 | 文件 | 说明 |
|------|------|------|
| Windows x64 | `cline-proxy-desktop.exe` | Win10/11 自带 WebView2 |
| macOS Apple Silicon | `cline-proxy-desktop-darwin-arm64` | 需 Xcode CLT |
| macOS Intel | `cline-proxy-desktop-darwin-amd64` | 需 Xcode CLT |
| Linux x64 | `cline-proxy-desktop-linux-amd64` | 需 GTK3 + WebKit2GTK |

### 方式二：命令行

```bash
go build -o cline-proxy .
./cline-proxy              # 默认端口 3457
./cline-proxy -port 8080   # 指定端口
```

启动后访问 http://127.0.0.1:3457/admin/ 进入管理后台。

### 方式三：Docker 部署（支持 AMD64 & ARM64 多架构）

推荐使用 `docker-compose.yml` 配合数据目录挂载启动：

```yaml
services:
  cline-proxy:
    image: 2casiku/cline2api:latest
    container_name: cline-proxy
    restart: unless-stopped
    ports:
      - "3457:3457"
    volumes:
      # 推荐整目录挂载，自动持久化所有配置与账号，宿主机无需预先创建任何文件
      - ./data:/app/data
    environment:
      - PORT=3457
      - CLINE_PROXY_HOST=0.0.0.0
      - DATA_DIR=/app/data
```

```bash
docker compose up -d      # 启动容器
docker compose logs -f     # 查看日志
docker compose down        # 停止容器
```

> **注意**：使用 `DATA_DIR=/app/data` 配合 `./data:/app/data` 挂载，程序启动时会自动初始化并持久化全部配置文件；切勿直接使用单文件挂载（若宿主机文件未提前存在，Docker 会默认将其创建为文件夹导致读写失败）。

## 使用指南

### 1. 添加 Cline 账号

在管理后台 **账号管理 → 导入账号**：

- **OAuth 浏览器登录**：点击按钮启动设备授权流程，在系统浏览器中完成登录（支持已登录 Cline 的浏览器）
- **手动输入 Token**：输入已有账号的 refreshToken
- **批量文件导入**：上传 JSON 文件或粘贴文本（每行一个 token，或 JSON 数组 `[{refreshToken, email}]`）

### 2. 配置客户端

```
Base URL: http://127.0.0.1:3457/v1
API Key:  <在管理后台生成的 Key>
Model:    cline-free/glm-5.2 (或任意可用模型)
```

支持的三种接入协议：
1. **OpenAI Chat**：`/v1/chat/completions`
2. **Anthropic Claude**：`/v1/messages`
3. **OpenAI Responses**：`/v1/responses`（支持 Cherry Studio 等客户端）

### 3. 账号导出/导入（跨设备迁移）

- **导出**：账号管理页面点击「导出」按钮，下载 `cline-accounts-export.json`
- **导入**：在另一台设备上用「从文件导入」上传该文件
- 导出格式与批量导入格式完全兼容

### 4. System Prompt 覆盖

在数据目录（或 exe 所在目录）下创建 `override.md`，内容将自动替换所有客户端请求的系统提示词。

### 5. 监听地址与访问设置（局域网 / 多网卡）

默认只监听 `127.0.0.1`（仅本机可访问）。管理后台 **访问设置** 区可：

- **监听地址下拉选择**：`127.0.0.1`（仅本机）/ `0.0.0.0`（所有网卡）/ 本机检测到的 IP，保存后自动重启监听立即生效
- **管理后台密码**：默认无密码；设置后访问 `/admin/` 需输入密码登录（会话 Cookie，24 小时有效），留空保存可清除密码

命令行指定监听地址方式：
```bash
./cline-proxy -host 0.0.0.0
# 或环境变量方式
CLINE_PROXY_HOST=0.0.0.0 ./cline-proxy
```

## 可用模型与同步机制

后台提供了自动与手动的双源模型同步支持：

### 1. 从 Cline 同步模型
- **实现原理**：向 Cline 官方公开接口（`https://api.cline.bot/api/v1/ai/cline/recommended-models`）发起探测，拉取最新的免费模型（Free）、订阅模型（ClinePass）和推荐模型。
- **自动对齐**：与本地模型池进行差量比对，自动标记并加入新增模型，自动移除非官方提供的下架模型，并实时持久化。若模型发生变动，进入管理后台时会弹出通知。
- **手动触发**：在管理后台「可用模型」列表上方，点击 **「从 Cline 同步模型」** 可即时手动刷新。

### 2. 从 opencode 同步模型
- **实现原理**：向 opencode 官方模型接口拉取全部开放模型，智能识别免费模型（含 `-free` 后缀或白名单），并自动补齐 200K 上下文与输出长度设定。
- **定时同步**：开启 opencode 免费模型支持后，后台启动轻量定时器，**每 10 分钟自动在后台同步一次**，模型列表始终保持最新。
- **手动触发**：在管理后台点击 **「从 opencode 同步模型」** 即可立即触发手动更新。

## 网络出口代理设置

如果部署在需要代理访问外网的环境，可直接在 Web 管理后台设置：

- **Cline 官方出口代理**：位于后台「代理配置」卡片内，直接填入发往 `api.cline.bot` 的代理地址（支持 `http://user:pass@ip:port` 或 `socks5://user:pass@ip:port`）。点击“测试连接”可实时验证与官方接口的握手延迟，保存后即时热生效，无需重启服务。
- **opencode 出口代理**：支持多代理列表轮询，自动进行限流冷却隔离。

## 数据文件与持久化

程序查找数据文件的优先级顺序：
1. 环境变量 `DATA_DIR` / `CLINE_DATA_DIR` 指定的目录（Docker 挂载优先）
2. 可执行文件所在目录
3. 当前工作目录
4. 用户主目录 `~/.cline2api/`

| 文件 | 说明 |
|------|------|
| `.cline-accounts.json` | 账号池、API Key、模型列表配置 |
| `.cline-config.json` | 系统基础设置（轮询策略、自定义请求头、Cline 出口代理等） |
| `.cline-zen.json` | opencode 免费模型配置与代理池状态 |
| `.cline-request-logs.json` | 请求用量与日志监控 |
| `override.md` | System Prompt 覆盖（可选） |

## 构建与 CI/CD

### 本地构建桌面端
```bash
# Windows
./desktop/build.sh
# macOS
xcode-select --install && ./desktop/build.sh
# Linux
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev && ./desktop/build.sh
```

### Docker 多架构自动发布
代码推送到 `dev` 分支或发布 `v*` 标签时，GitHub Actions 会自动触发构建并推送支持 **AMD64 & ARM64** 的多架构镜像至 Docker Hub。

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 项目结构

```
├── main.go              CLI 入口（go build .）
├── desktop_main.go      桌面端入口（go build -tags desktop）
├── proxy.go             HTTP 服务、API 路由、协议转换、SSE
├── responses.go         OpenAI Responses 协议转换与流式事件推流
├── admin.go             管理后台 REST API 与系统配置持久化
├── admin_html.go        管理后台前端（内嵌单页）
├── models_sync.go       Cline 官方推荐模型拉取与自动同步
├── zen.go               opencode 模型对接、定时同步与调度
├── zen_proxy.go         opencode 专用代理池与 uTLS 模拟
├── http.go              全局 HTTP 客户端与 Cline 动态出口代理
├── pool.go              账号池、多位置数据路径查找与初始化
├── request_logs.go      请求日志统计与持久化
├── desktop/             桌面端构建脚本、文档、图标生成器
├── Dockerfile           多架构交叉编译 Docker 镜像构建
├── docker-compose.yml   Docker Compose 编排模板
└── .github/workflows/   CI 自动化流水线（多架构 Docker & 三平台 Release）
```

## 许可证

[MIT License](LICENSE) © 2026 [luawei1](https://github.com/luawei1)
