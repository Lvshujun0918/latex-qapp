# 数据库设计：AI错题本（Go + GORM）

## 1. 设计目标
- 支持学生端错题离线记录与云端同步
- 支持题目 LaTeX 存储与版本回溯
- 支持 AI 结构化解析、学习统计、知识点标签
- 支持错题本 PDF 异步生成任务
- 兼容 GORM，默认以 MySQL/PostgreSQL 为云端主库（字段设计尽量跨库）

## 2. 命名与通用约定
- 主键统一：`id`（`BIGINT` 或 `UUID`，MVP 可先 BIGINT 自增）
- 时间字段统一：`created_at`、`updated_at`、`deleted_at`（软删）
- 业务时间戳：`local_version`（客户端最后修改时间）、`server_version`（服务端版本号）
- JSON 字段：MySQL 用 `JSON`，PostgreSQL 用 `JSONB`
- 字符集：UTF-8（需支持中英文与 LaTeX 特殊字符）

---

## 3. 核心表结构

## 3.1 users（用户表）
用途：登录、鉴权、账号状态。

字段建议：
- `id` BIGINT PK
- `username` VARCHAR(64) NOT NULL UNIQUE
- `password_hash` VARCHAR(255) NOT NULL
- `display_name` VARCHAR(64) NOT NULL
- `avatar_url` VARCHAR(512) NULL
- `status` VARCHAR(16) NOT NULL DEFAULT 'active'  // active, disabled
- `last_login_at` DATETIME/TIMESTAMP NULL
- `created_at` DATETIME/TIMESTAMP NOT NULL
- `updated_at` DATETIME/TIMESTAMP NOT NULL
- `deleted_at` DATETIME/TIMESTAMP NULL

索引：
- `uk_users_username`
- `idx_users_status`

---

## 3.2 refresh_tokens（刷新令牌表）
用途：JWT 刷新机制，支持多设备登录。

字段建议：
- `id` BIGINT PK
- `user_id` BIGINT NOT NULL FK -> users.id
- `token_hash` VARCHAR(255) NOT NULL UNIQUE
- `device_id` VARCHAR(128) NULL
- `device_name` VARCHAR(128) NULL
- `expires_at` DATETIME/TIMESTAMP NOT NULL
- `revoked_at` DATETIME/TIMESTAMP NULL
- `created_at` DATETIME/TIMESTAMP NOT NULL
- `updated_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `uk_refresh_tokens_token_hash`
- `idx_refresh_tokens_user_id`
- `idx_refresh_tokens_expires_at`

---

## 3.3 error_records（错题主表）
用途：错题核心实体，保存 LaTeX、分类与同步元数据。

字段建议：
- `id` BIGINT PK
- `user_id` BIGINT NOT NULL FK -> users.id
- `subject` VARCHAR(32) NOT NULL  // math, physics...
- `grade_level` VARCHAR(32) NULL
- `question_type` VARCHAR(32) NULL  // choice, fill_blank, proof...
- `difficulty` TINYINT/SMALLINT NOT NULL DEFAULT 3  // 1-5
- `title` VARCHAR(255) NULL
- `source_type` VARCHAR(16) NOT NULL DEFAULT 'manual' // manual, camera, import
- `latex_source` LONGTEXT/TEXT NOT NULL
- `latex_version` INT NOT NULL DEFAULT 1
- `latex_render_status` VARCHAR(16) NOT NULL DEFAULT 'pending' // pending, ok, failed
- `latex_render_error` TEXT NULL
- `mistake_reason` TEXT NULL // 学生自述错因
- `is_favorite` BOOLEAN NOT NULL DEFAULT FALSE
- `mastery_level` TINYINT/SMALLINT NOT NULL DEFAULT 0 // 0-100
- `last_reviewed_at` DATETIME/TIMESTAMP NULL
- `review_count` INT NOT NULL DEFAULT 0

同步与冲突字段：
- `client_record_id` VARCHAR(64) NOT NULL // 客户端 UUID，便于离线映射
- `is_synced` BOOLEAN NOT NULL DEFAULT FALSE
- `is_deleted` BOOLEAN NOT NULL DEFAULT FALSE
- `local_version` BIGINT NOT NULL DEFAULT 0 // 客户端毫秒时间戳
- `server_version` BIGINT NOT NULL DEFAULT 0 // 服务端版本号（自增或时间戳）
- `sync_status` VARCHAR(16) NOT NULL DEFAULT 'pending' // pending, synced, conflict, failed

通用字段：
- `created_at` DATETIME/TIMESTAMP NOT NULL
- `updated_at` DATETIME/TIMESTAMP NOT NULL
- `deleted_at` DATETIME/TIMESTAMP NULL

索引：
- `idx_error_records_user_id_created_at`
- `idx_error_records_user_subject`
- `idx_error_records_user_sync_status`
- `idx_error_records_client_record_id`（建议唯一：`uk_error_records_user_client_record_id`）
- `idx_error_records_latex_render_status`

---

## 3.4 knowledge_points（知识点字典）
用途：标准知识点库。

字段建议：
- `id` BIGINT PK
- `subject` VARCHAR(32) NOT NULL
- `name` VARCHAR(128) NOT NULL
- `parent_id` BIGINT NULL FK -> knowledge_points.id
- `path` VARCHAR(512) NULL // 层级路径，如 数学/函数/二次函数
- `description` TEXT NULL
- `sort_order` INT NOT NULL DEFAULT 0
- `is_active` BOOLEAN NOT NULL DEFAULT TRUE
- `created_at` DATETIME/TIMESTAMP NOT NULL
- `updated_at` DATETIME/TIMESTAMP NOT NULL
- `deleted_at` DATETIME/TIMESTAMP NULL

索引：
- `uk_kp_subject_name`（subject + name）
- `idx_kp_parent_id`

---

## 3.5 error_record_tags（错题-知识点关联表）
用途：多对多关联。

字段建议：
- `id` BIGINT PK
- `error_record_id` BIGINT NOT NULL FK -> error_records.id
- `knowledge_point_id` BIGINT NOT NULL FK -> knowledge_points.id
- `source` VARCHAR(16) NOT NULL DEFAULT 'ai' // ai, manual
- `confidence` DECIMAL(5,2) NULL // AI 打标置信度
- `created_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `uk_ert_record_kp`（error_record_id + knowledge_point_id）
- `idx_ert_kp_id`

---

## 3.6 ai_analyses（AI 解析结果）
用途：保存每次 AI 分析结果，支持历史版本与成本统计。

字段建议：
- `id` BIGINT PK
- `error_record_id` BIGINT NOT NULL FK -> error_records.id
- `user_id` BIGINT NOT NULL FK -> users.id
- `provider` VARCHAR(32) NOT NULL // openai, anthropic...
- `model_name` VARCHAR(64) NOT NULL
- `prompt_version` VARCHAR(32) NULL
- `input_payload` JSON/JSONB NULL
- `analysis_text` LONGTEXT/TEXT NOT NULL
- `structured_result` JSON/JSONB NULL // steps, key_points, suggestions
- `token_input` INT NULL
- `token_output` INT NULL
- `cost_amount` DECIMAL(10,4) NULL
- `status` VARCHAR(16) NOT NULL DEFAULT 'success' // success, failed
- `error_message` TEXT NULL
- `created_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `idx_ai_analyses_record_id`
- `idx_ai_analyses_user_id_created_at`
- `idx_ai_analyses_provider_model`

---

## 3.7 review_logs（复习日志）
用途：记录每次复习行为，用于统计与推荐。

字段建议：
- `id` BIGINT PK
- `user_id` BIGINT NOT NULL FK -> users.id
- `error_record_id` BIGINT NOT NULL FK -> error_records.id
- `action` VARCHAR(32) NOT NULL // viewed, retried, solved, unsolved
- `result_score` TINYINT/SMALLINT NULL // 0-100
- `duration_seconds` INT NULL
- `note` TEXT NULL
- `created_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `idx_review_logs_user_created_at`
- `idx_review_logs_record_id`

---

## 3.8 sync_logs（同步日志）
用途：记录每次 push/pull 事件、冲突与错误，便于排障。

字段建议：
- `id` BIGINT PK
- `user_id` BIGINT NOT NULL FK -> users.id
- `device_id` VARCHAR(128) NULL
- `direction` VARCHAR(8) NOT NULL // push, pull
- `entity_type` VARCHAR(32) NOT NULL // error_record, review_log...
- `entity_id` BIGINT NULL
- `client_record_id` VARCHAR(64) NULL
- `local_version` BIGINT NULL
- `server_version` BIGINT NULL
- `status` VARCHAR(16) NOT NULL // success, conflict, failed
- `error_code` VARCHAR(64) NULL
- `error_message` TEXT NULL
- `payload` JSON/JSONB NULL
- `created_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `idx_sync_logs_user_created_at`
- `idx_sync_logs_status`
- `idx_sync_logs_client_record_id`

---

## 3.9 pdf_jobs（PDF 生成任务）
用途：错题本导出任务队列。

字段建议：
- `id` BIGINT PK
- `user_id` BIGINT NOT NULL FK -> users.id
- `job_no` VARCHAR(64) NOT NULL UNIQUE
- `template_name` VARCHAR(64) NOT NULL DEFAULT 'default'
- `title` VARCHAR(255) NOT NULL
- `filter_payload` JSON/JSONB NOT NULL // 导出筛选条件
- `record_count` INT NOT NULL DEFAULT 0
- `tex_file_url` VARCHAR(1024) NULL
- `pdf_file_url` VARCHAR(1024) NULL
- `status` VARCHAR(16) NOT NULL DEFAULT 'queued' // queued, running, success, failed, partial
- `progress` TINYINT/SMALLINT NOT NULL DEFAULT 0 // 0-100
- `error_summary` TEXT NULL
- `started_at` DATETIME/TIMESTAMP NULL
- `finished_at` DATETIME/TIMESTAMP NULL
- `expires_at` DATETIME/TIMESTAMP NULL // 下载链接过期时间
- `created_at` DATETIME/TIMESTAMP NOT NULL
- `updated_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `uk_pdf_jobs_job_no`
- `idx_pdf_jobs_user_created_at`
- `idx_pdf_jobs_status`

---

## 3.10 pdf_job_items（PDF 任务明细）
用途：记录每道题在导出任务中的渲染状态，支持“部分成功”。

字段建议：
- `id` BIGINT PK
- `pdf_job_id` BIGINT NOT NULL FK -> pdf_jobs.id
- `error_record_id` BIGINT NOT NULL FK -> error_records.id
- `seq_no` INT NOT NULL
- `render_status` VARCHAR(16) NOT NULL DEFAULT 'pending' // pending, success, failed, skipped
- `error_message` TEXT NULL
- `created_at` DATETIME/TIMESTAMP NOT NULL
- `updated_at` DATETIME/TIMESTAMP NOT NULL

索引：
- `uk_pdf_job_items_job_seq`（pdf_job_id + seq_no）
- `idx_pdf_job_items_record_id`
- `idx_pdf_job_items_render_status`

---

## 4. 表关系概览
- `users 1-N error_records`
- `users 1-N ai_analyses`
- `users 1-N review_logs`
- `users 1-N sync_logs`
- `users 1-N pdf_jobs`
- `error_records N-N knowledge_points`（经 `error_record_tags`）
- `error_records 1-N ai_analyses`
- `error_records 1-N review_logs`
- `pdf_jobs 1-N pdf_job_items`
- `error_records 1-N pdf_job_items`

---

## 5. GORM 建模建议
- 所有模型嵌入统一基类（含 `ID/CreatedAt/UpdatedAt/DeletedAt`）
- 对 JSON 字段使用自定义类型（兼容 MySQL JSON / PostgreSQL JSONB）
- 关联外键建议显式声明 `constraint:OnUpdate:CASCADE,OnDelete:RESTRICT`（关键业务数据默认限制删除）
- 高频查询字段建立复合索引（如 `user_id + created_at`）
- `client_record_id` + `user_id` 作为幂等键处理同步写入

---

## 6. 同步策略落地到数据库

关键规则：
1. 客户端每次修改错题时递增 `local_version`（建议毫秒时间戳）
2. 服务端写入成功后更新 `server_version`
3. `push` 时以 `(user_id, client_record_id)` 定位记录
4. 若客户端 `local_version` < 服务端 `server_version`：
   - 标记冲突，写 `sync_logs`，返回服务端版本给客户端
5. 默认冲突策略：Last-Write-Wins（同时保留冲突日志便于追踪）

---

## 7. 初始化迁移顺序（推荐）
1. `users`
2. `refresh_tokens`
3. `knowledge_points`
4. `error_records`
5. `error_record_tags`
6. `ai_analyses`
7. `review_logs`
8. `sync_logs`
9. `pdf_jobs`
10. `pdf_job_items`

---

## 8. MVP 最小可用子集（可先上线）
如果要快速启动，第一版可先建以下表：
- `users`
- `refresh_tokens`
- `error_records`
- `ai_analyses`
- `sync_logs`
- `pdf_jobs`

待第二迭代再补：
- `knowledge_points`
- `error_record_tags`
- `review_logs`
- `pdf_job_items`

---

## 9. 未来扩展预留
- 多角色（老师/家长）可新增 `roles`、`user_roles`
- 班级与共享题库可新增 `classes`、`class_members`、`shared_records`
- 题目去重可加 `question_fingerprint`（文本+LaTeX hash）
- AI 成本治理可加 `ai_billing_daily_stats`
