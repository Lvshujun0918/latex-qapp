# Plan: AI错题本（Ionic + Vue + Capacitor + Go/GORM + LaTeX/PDF）

## 目标
构建 Android 优先的原生学习应用：学生可离线记录错题（LaTeX），联网后自动/手动同步到服务端；服务端提供 AI 解析、学习推荐与统计，并支持基于 LaTeX 生成错题本 PDF。

## 技术栈与约束（已冻结）
- 前端：Ionic Framework（Vue）+ Capacitor + TypeScript + Pinia
- 后端：Go + Gin + GORM + JWT（Bearer）
- 存储：本地 SQLite（离线）+ 云端 MySQL/PostgreSQL（自建）
- AI：第三方大模型 API（可替换 OpenAI/Claude）
- 文档能力：LaTeX 题目存储与服务端 PDF 编译导出
- 平台：Android 优先（后续可扩展 iOS）

## 分阶段规划

### Phase 0：需求与数据规范
1. 冻结 MVP 范围：
   - 错题录入（大模型视觉识别后存储 LaTeX）
   - AI 解析与建议
   - 离线可用 + 同步
   - LaTeX 记录 + 错题本 PDF 导出
2. 定义统一数据字典：
   - 题目基础信息、LaTeX 源码、知识点、AI 解析结构、同步字段
3. 明确安全策略：
   - 仅 Bearer Token；移动端本地安全存储 Token

### Phase 1：前端工程初始化（Ionic）
1. 初始化 Ionic Vue + Capacitor Android 工程
2. 搭建目录结构：
   - `src/views`：页面
   - `src/stores`：Pinia 状态
   - `src/services`：API/DB/同步服务
   - `src/types`：类型定义
3. 集成核心能力：
   - Camera（拍照）
   - Filesystem（本地文件）
   - Network（网络状态）
4. 本地数据库层：
   - SQLite 初始化
   - 错题、AI 解析、同步日志本地 CRUD

### Phase 2：后端工程初始化（Go + Gin + GORM）
1. 建立分层架构：
   - `internal/handler`、`internal/service`、`internal/repo`、`internal/model`、`internal/middleware`
2. GORM 模型定义与迁移
3. 认证模块（JWT）：
   - 注册/登录/刷新 token/用户信息
4. 基础接口：
   - 错题 CRUD
   - AI 分析触发与查询
   - 同步 push/pull
   - 学习统计

### Phase 3：LaTeX 题库能力
1. 题目模型扩展：
   - `latex_source`、`latex_version`、`render_status`
2. 录入链路：
   - 图片视觉识别 → AI/人工修订为 LaTeX
3. 存储策略：
   - 仅存储 LaTeX 与版本，支持回溯
4. 检索能力：
   - 学科/知识点/难度/掌握度维度筛选

### Phase 4：错题本 PDF 生成（LaTeX 编译）
1. 设计模板：
   - 封面、目录、题目、解析、错因、复习建议、统计附录
2. 服务端生成流程：
   - 选择题集 → 生成 `.tex` → 调用 `xelatex` 编译 → 返回 PDF URL
3. 任务化处理：
   - 异步生成、状态查询、失败重试
4. 异常兜底：
   - 单题 LaTeX 编译失败时跳过并记录，不阻断整本导出

### Phase 5：离线同步与冲突处理
1. 本地字段：`is_synced`、`local_version`、`is_deleted`
2. 同步策略：
   - 自动实时（网络可用时）+ 手动同步按钮
3. 冲突策略：
   - 默认 Last-Write-Wins（保留冲突日志）
4. 数据一致性：
   - 增量同步与断点续传

### Phase 6：测试、部署与发布
1. 测试：
   - 后端接口测试
   - 前端关键流程（录题、解析、同步、导出）
2. 部署：
   - Docker 化 Go 服务 + 数据库 + LaTeX 运行环境
3. 发布：
   - Android 签名打包与灰度分发
4. 监控：
   - AI 调用成功率、PDF 失败率、同步失败率

## 核心数据模型（建议）
- `users`：用户信息
- `error_records`：错题主表（LaTeX、分类、难度）
- `ai_analyses`：AI 解析结果（结构化）
- `knowledge_points`：知识点字典
- `error_record_tags`：错题与知识点关联
- `sync_logs`：同步事件日志
- `pdf_jobs`：PDF 生成任务

## API 规划（MVP）
- 认证：
   - `POST /api/auth/register`（用户名 + 密码）
   - `POST /api/auth/login`（用户名 + 密码）
  - `POST /api/auth/register`
  - `POST /api/auth/login`
  - `POST /api/auth/refresh`
  - `GET /api/auth/me`
- 错题：
  - `POST /api/records`
  - `GET /api/records`
  - `GET /api/records/:id`
  - `PUT /api/records/:id`
  - `DELETE /api/records/:id`
- AI：
  - `POST /api/ai/analyze`
  - `GET /api/ai/analysis/:recordId`
- 同步：
  - `POST /api/sync/push`
  - `POST /api/sync/pull`
- PDF：
  - `POST /api/pdf/export`
  - `GET /api/pdf/jobs/:jobId`

## 目录建议

### 前端
- `frontend/src/views`
- `frontend/src/stores`
- `frontend/src/services/api.ts`
- `frontend/src/services/db.ts`
- `frontend/src/services/sync.ts`
- `frontend/src/services/pdf.ts`

### 后端
- `backend/cmd/server/main.go`
- `backend/internal/model`（GORM 模型）
- `backend/internal/repo`（GORM 仓储）
- `backend/internal/service`（业务逻辑）
- `backend/internal/handler`（HTTP 接口）
- `backend/internal/service/pdf_service.go`
- `backend/templates/latex`（LaTeX 模板）

## 里程碑与验收
1. M1（基础可跑）
   - Ionic App 启动成功，Go + GORM 服务启动成功
2. M2（核心闭环）
   - 录题 → AI 解析 → 本地保存 → 同步到云端
3. M3（文档能力）
   - LaTeX 题目存储可用，PDF 导出可下载
4. M4（发布就绪）
   - Android 安装包可稳定运行，核心指标可监控

## 风险与对策
- LaTeX 编译环境复杂：
  - 统一在服务端容器内处理，固定 TeX Live 版本
- OCR/公式识别误差：
  - 保留人工修订 LaTeX 流程
- 弱网下同步失败：
  - 重试队列 + 幂等接口 + 增量同步
- AI 成本波动：
  - 题目哈希缓存 + 可配置重算策略

## 下一步执行顺序
1. 初始化 `frontend`（Ionic Vue + Capacitor）
2. 初始化 `backend`（Go + Gin + GORM）
3. 创建 GORM 模型与迁移
4. 打通认证 + 错题 CRUD + 本地 SQLite
5. 接入 AI 分析
6. 接入 LaTeX 模板与 PDF 导出
7. 完成同步策略与端到端联调
