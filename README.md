# LaTeX QApp

一个面向移动端的 AI 错题本项目：支持拍照识别题目、LaTeX 草稿生成、AI 解题解析、错题管理与统计、按题目导出 PDF。

## 目录

- [项目亮点](#项目亮点)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [环境变量说明](#环境变量说明)
- [常用开发命令](#常用开发命令)
- [API 概览](#api-概览)
- [Android 发布](#android-发布)
- [后端镜像发布](#后端镜像发布)
- [常见问题](#常见问题)
- [安全建议](#安全建议)
- [License](#license)

## 项目亮点

- 拍照录题：通过 AI 视觉模型提取题干和结构化信息。
- 错题管理：支持列表检索、详情查看、编辑、删除。
- AI 解析：支持生成答案与分步解析。
- 复习闭环：提供复习与统计页面，聚焦薄弱点。
- PDF 导出：按选择题目生成可下载 PDF（后端 LaTeX 编译）。
- 移动端体验：基于 Capacitor，可构建 Android APK。

## 技术栈

### 前端

- Vue 3 + TypeScript + Vite
- Pinia + Vue Router
- Tailwind CSS v4 + shadcn/reka-ui 组件体系
- Capacitor（Android、Camera、StatusBar 等）

### 后端

- Go 1.26 + Gin
- GORM + SQLite
- JWT 鉴权（Bearer Token）
- Qwen（视觉/文本）模型集成
- LaTeX/PDF：`latexmk + xelatex`

### CI/CD

- GitHub Actions Android Release（标签触发）
- GitHub Actions 后端镜像推送 GHCR

## 项目结构

```text
latex-qapp/
├─ frontend/                     # Vue + Capacitor 前端
│  ├─ src/
│  ├─ android/                   # Android 工程
│  ├─ capacitor.config.ts
│  └─ package.json
├─ backend/                      # Go 后端
│  ├─ cmd/server/main.go
│  ├─ internal/
│  ├─ template.tex               # PDF 模板
│  ├─ Dockerfile
│  └─ .env.example
├─ .github/workflows/
│  ├─ android-release.yml
│  └─ backend-ghcr.yml
└─ dev.ps1                       # 一键同时启动前后端（Windows）
```

## 快速开始

### 1. 环境准备

- Node.js 22+（建议与 CI 对齐）
- npm 10+
- Go 1.26+
- Windows PowerShell（如果使用 `dev.ps1`）

可选（仅当你要本地导出 PDF 时需要）：

- `latexmk`
- `xelatex`（TeX Live）

### 2. 克隆与安装

```bash
# 根目录
cd latex-qapp

# 前端依赖
cd frontend
npm ci
cd ..
```

### 3. 配置环境变量

后端：复制并编辑 `backend/.env.example`。

```bash
cd backend
cp .env.example .env
```

前端：`frontend/.env` 里至少要配置 `VITE_API_BASE_URL`，例如：

```env
VITE_API_BASE_URL=http://localhost:8080
```

### 4. 启动开发环境

#### Windows 一键启动（推荐）

```powershell
./dev.ps1
```

该脚本会同时启动：

- 前端：`npm run dev`
- 后端：`go run ./cmd/server/.`

#### 手动启动

```bash
# 终端 1：前端
cd frontend
npm run dev

# 终端 2：后端
cd backend
go run ./cmd/server/.
```

### 5. 健康检查

后端默认端口 `8080`，可访问：

- `GET /health`

## 环境变量说明

### 后端（`backend/.env`）

| 变量名 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | 后端服务端口 |
| `DATABASE_DSN` | `file:app.db?cache=shared&mode=rwc` | SQLite DSN |
| `JWT_SECRET` | `change-me-in-production` | JWT 密钥 |
| `ALLOW_ORIGIN` | `*` | CORS 允许来源 |
| `QWEN_API_KEY` | 空 | Qwen API Key |
| `QWEN_BASE_URL` | `https://dashscope.aliyuncs.com/compatible-mode/v1` | Qwen OpenAI 兼容地址 |
| `QWEN_VISION_MODEL` | `qwen-vl-max-latest` | 视觉模型 |
| `QWEN_TEXT_MODEL` | `qwen-plus` | 文本模型 |
| `TEMPLATE_TEX_PATH` | `./template.tex` | PDF LaTeX 模板路径 |
| `PDF_OUTPUT_DIR` | `./public/pdfs` | PDF 输出目录 |

### 前端（`frontend/.env`）

| 变量名 | 示例 | 说明 |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `http://localhost:8080` | 前端请求后端地址 |

## 常用开发命令

### 前端

```bash
cd frontend
npm run dev          # 本地开发
npm run build        # 生产构建
npm run preview      # 预览构建产物
npm run lint         # ESLint
npm run test:unit    # Vitest
npm run test:e2e     # Cypress
```

### 后端

```bash
cd backend
go run ./cmd/server/.   # 运行服务
go test ./...           # 运行全部测试
```
### Android（本地）

```bash
cd frontend
npm run build
npx cap sync android
cd android
./gradlew assembleRelease
```

如需注入版本元数据，可附加参数：

- `-PVERSION_NAME=1.0.0`
- `-PVERSION_CODE=10000`
- `-PGIT_SHA=abc1234`
- `-PBUILD_TIME=2026-04-10T00:00:00Z`

## API 概览

后端统一前缀：`/api`。

### 鉴权

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me`（需 JWT）

### AI

- `POST /api/ai/vision/latex`（需 JWT）
- `POST /api/ai/vision/latex/stream`（需 JWT）
- `POST /api/ai/solve`（需 JWT）
- `POST /api/ai/solve/stream`（需 JWT）

### 错题记录

- `GET /api/records`（需 JWT）
- `POST /api/records`（需 JWT）
- `GET /api/records/:id`（需 JWT）
- `PUT /api/records/:id`（需 JWT）
- `DELETE /api/records/:id`（需 JWT）

### 统计

- `GET /api/stats/overview`（需 JWT）
- `GET /api/stats/by-category`（需 JWT）
- `GET /api/stats/trending`（需 JWT）

### PDF

- `POST /api/pdf/export`（需 JWT）
- `GET /api/pdf/jobs/:jobId`（需 JWT）
- 静态文件：`/public/**`

### 鉴权方式

前端通过 `Authorization: Bearer <token>` 访问受保护接口。

### 生产部署参考

可基于 `backend/docker-compose.release.yml`：

1. 在 `backend/` 下创建 `.env.release`（参考 `.env.release.example`）。
2. 启动：

```bash
cd backend
docker compose -f docker-compose.release.yml up -d --build
```

## 常见问题

### 1) 登录后接口仍然 401/403

优先检查：

- 前端是否在请求头中携带 `Authorization: Bearer <token>`
- 后端 `JWT_SECRET` 是否与签发时一致
- `VITE_API_BASE_URL` 是否指向正确服务

### 2) PDF 导出失败（`latexmk`/`xelatex`）

- 本地开发机需安装 TeX Live 相关组件
- 确认 `TEMPLATE_TEX_PATH`、`PDF_OUTPUT_DIR` 可访问且有写权限
- Docker 部署建议直接使用仓库提供的后端 Dockerfile

### 3) CORS 错误

- 将 `ALLOW_ORIGIN` 设置为前端真实地址（开发环境常见是 Vite 地址）
