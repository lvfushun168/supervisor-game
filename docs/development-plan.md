# OC 督学专注游戏详细开发计划

## 1. 开发口径

本开发计划基于 [oc-supervisor-game-prd.md](./oc-supervisor-game-prd.md)，目标是交付一个单机定制版本地网页专注游戏。

技术栈固定为：

- 后端：Go + Gin + GORM。
- 数据库：MySQL 8.0。
- 前端：Vite + Vue 3。
- 交付：Go 单 exe 本地服务，前端构建产物嵌入 exe。
- 素材：本地 `assets/` 相对路径随包交付。
- 管理端：固定 appkey 进入隐藏路由，不做复杂账号体系。

开发原则：

- 先完成可运行闭环，再补体验和配置细节。
- 所有配置优先简单表单化，避免做成复杂运营后台。
- 用户端第一屏必须是游戏界面。
- 不做对象存储、多租户、多管理员、支付、排行、社交。
- 所有暂时无法实现、依赖外部素材、依赖客户确认或技术条件不确定的内容，必须在文档、代码注释、Issue 或提交说明中使用 `TODO:` 标记。

`TODO:` 标记格式：

```text
TODO: [原因] [暂定处理方式] [后续触发条件]
```

示例：

```text
TODO: 客户暂未提供失败动作视频，当前先使用 patrol_suspicious.mp4 兜底，收到素材后替换为 fail.mp4。
```

## 2. 总体里程碑

| 里程碑 | 目标 | 结果 |
| --- | --- | --- |
| M0 | 工程基础和配置规范 | 项目能启动、构建、连接配置清晰 |
| M1 | 数据库和后端基础能力 | GORM models、迁移、种子数据、基础 API |
| M2 | 管理端配置闭环 | appkey 管理端可配置场景、动作、模型、巡查规则、MySQL |
| M3 | 用户端主界面和计时状态机 | 用户能开始、暂停、继续、结束劳动 |
| M4 | 巡查和动作播放闭环 | 随机巡查、截图、后端判定、动作视频播放 |
| M5 | 成长、档案、任务和结算 | 历史记录、等级、货币、任务、结算页可用 |
| M6 | 打包、验收和交付整理 | 单 exe + assets + 文档可交付 |

## 3. M0：工程基础和配置规范

### 3.1 目标

建立开发人员可以稳定启动、构建、调试的基础工程，并明确本地配置方式。

### 3.2 后端任务

- 确认 Go module 名称和目录结构。
- 保留 `cmd` 或根目录 `main.go` 启动入口。
- 建立 `internal/config`，负责读取服务配置。
- 支持 `.env` 文件读取。
- 支持环境变量覆盖配置。
- 增加以下基础配置项：
  - `APP_ENV`
  - `APP_ADDR`
  - `APP_KEY`
  - `DB_DSN`
  - `ASSETS_DIR`
  - `CONFIG_ENCRYPTION_KEY`
- 建立统一错误响应结构。
- 建立 `/api/health`。
- 建立静态资源托管：
  - 前端 `dist` 嵌入 exe。
  - `assets/` 使用本地目录服务。
- 增加基础日志输出：
  - 服务启动地址。
  - 数据库连接状态。
  - assets 目录。
  - 当前运行模式。

### 3.2.1 数据库配置自举规则

MySQL 是本项目唯一主存储。由于管理端保存的 MySQL 连接配置也存放在 MySQL 中，第一版按以下规则处理首次启动和后续修改：

- 首次启动必须通过 `DB_DSN` 或 `.env` 提供可连接的 MySQL DSN。
- 服务启动成功并完成迁移后，才允许在管理端保存 `mysql_configs`。
- 服务启动时先使用 `DB_DSN` 作为 bootstrap 连接读取配置表；若存在已启用的 `mysql_configs`，再切换为该连接作为运行期主连接。
- 如果 bootstrap 数据库中不存在已启用的 `mysql_configs`，则直接使用 `DB_DSN` 作为运行期主连接。
- 管理端修改 MySQL 连接配置后，第一版只要求提示“重启服务后生效”，不做运行时热切换。
- 如果已启用的 `mysql_configs` 连接失败，服务应回退到 bootstrap `DB_DSN` 并在日志和 `/api/admin/status` 中暴露明确状态。
- `.env.example` 必须明确写出首次启动需要准备 MySQL 8.0，并创建目标数据库。

`CONFIG_ENCRYPTION_KEY` 用于本地加密 API Key 和数据库密码。第一版可使用 AES-GCM；如果暂未实现加密，必须使用 `TODO:` 标记，并保证密钥字段不会通过用户端 API 下发。

### 3.3 前端任务

- 确认 Vite + Vue 3 基础结构。
- 增加路由基础结构：
  - `/`
  - `/settings`
  - `/profile`
  - `/tasks`
  - `/settlement`
  - `/__admin`
- 增加 API client 基础封装。
- 增加统一错误提示组件。
- 增加全局样式基础变量。
- 配置 Vite dev server 代理 `/api` 到 Go 后端。

### 3.4 配置文件

- `.env.example` 必须包含可运行示例。
- README 必须说明：
  - 如何启动 MySQL。
  - 如何启动 Go 服务。
  - 如何启动前端开发服务器。
  - 如何构建单 exe。

### 3.5 验收标准

- 执行 `go test ./...` 成功。
- 执行 `npm run build` 成功。
- 执行 `go build -o bin/supervisor-game .` 成功。
- 本地访问 `/api/health` 返回正常 JSON。
- 未配置数据库时能给出清晰状态，不崩溃。
- `.env.example` 能指导开发人员完成启动。

验收人签名_____

## 4. M1：数据库和后端基础能力

### 4.1 目标

建立 MySQL 8.0 数据模型、迁移能力、默认种子数据和后端基础服务层。

### 4.2 数据表

实现以下 GORM models：

- `AppSetting`
- `Character`
- `UserSetting`
- `Scene`
- `ActionConfig`
- `ModelConfig`
- `PatrolRule`
- `MySQLConfig`
- `WorkSession`
- `PatrolRecord`
- `DailyStat`
- `Task`
- `TaskRecord`
- `Badge`
- `UserProgress`

### 4.3 字段要求

`app_settings`：

- `id`
- `setting_key`
- `setting_value_json`
- `description`
- `created_at`
- `updated_at`

唯一约束：

- `setting_key`

`characters`：

- `id`
- `character_key`
- `name`
- `enabled`
- `description`
- `avatar_url`
- `profile_json`
- `voice_style`
- `default_scene_key`
- `metadata_json`
- `created_at`
- `updated_at`

唯一约束：

- `character_key`

`user_settings`：

- `id`
- `mode`
- `custom_duration_seconds`
- `patrol_frequency`
- `background_audio_key`
- `background_volume`
- `action_volume`
- `ui_volume`
- `quiet_patrol_enabled`
- `screen_filter`
- `camera_enabled`
- `camera_device_id`
- `metadata_json`
- `created_at`
- `updated_at`

`scenes`：

- `id`
- `scene_key`
- `name`
- `enabled`
- `description`
- `background_type`
- `background_url`
- `background_video_url`
- `background_poster_url`
- `ambient_audio_url`
- `default_action_key`
- `available_action_keys_json`
- `model_result_action_map_json`
- `metadata_json`
- `created_at`
- `updated_at`

`action_configs`：

- `id`
- `scene_key`
- `action_key`
- `name`
- `enabled`
- `priority`
- `video_url`
- `poster_url`
- `duration_ms`
- `next_action_key`
- `model_result_map_json`
- `local_rule_map_json`
- `metadata_json`
- `created_at`
- `updated_at`

`model_configs`：

- `id`
- `name`
- `provider`
- `enabled`
- `base_url`
- `api_key_encrypted`
- `model`
- `timeout_ms`
- `max_image_width`
- `temperature`
- `prompt`
- `response_schema_json`
- `retry_count`
- `created_at`
- `updated_at`

`mysql_configs`：

- `id`
- `host`
- `port`
- `database_name`
- `username`
- `password_encrypted`
- `charset`
- `timezone`
- `max_open_conns`
- `max_idle_conns`
- `enabled`
- `last_tested_at`
- `last_test_result`
- `last_test_error`
- `created_at`
- `updated_at`

`patrol_rules`：

- `id`
- `slow_min_seconds`
- `slow_max_seconds`
- `normal_min_seconds`
- `normal_max_seconds`
- `high_min_seconds`
- `high_max_seconds`
- `max_warnings`
- `max_violations`
- `suspicious_adds_warning`
- `violation_direct_fail`
- `camera_off_strategy`
- `capture_failed_strategy`
- `model_timeout_retry_count`
- `user_error_message`
- `created_at`
- `updated_at`

`work_sessions`：

- `id`
- `scene_key`
- `mode`
- `planned_duration_seconds`
- `started_at`
- `ended_at`
- `actual_focus_seconds`
- `patrol_count`
- `warning_count`
- `violation_count`
- `result`
- `finish_reason`
- `earned_currency`
- `created_at`
- `updated_at`

`patrol_records`：

- `id`
- `session_id`
- `scene_key`
- `triggered_at`
- `status`
- `confidence`
- `reason`
- `action_key`
- `warning_delta`
- `violation_delta`
- `model_raw_json`
- `error_code`
- `created_at`

`daily_stats`：

- `id`
- `stat_date`
- `focus_seconds`
- `session_count`
- `patrol_count`
- `warning_count`
- `violation_count`
- `earned_currency`
- `last_result`
- `created_at`
- `updated_at`

唯一约束：

- `stat_date`

`tasks`：

- `id`
- `task_key`
- `name`
- `description`
- `task_type`
- `target_value`
- `reward_currency`
- `enabled`
- `sort_order`
- `metadata_json`
- `created_at`
- `updated_at`

唯一约束：

- `task_key`

`task_records`：

- `id`
- `task_key`
- `record_date`
- `progress_value`
- `status`
- `claimed_at`
- `created_at`
- `updated_at`

唯一约束：

- `task_key + record_date`

`badges`：

- `id`
- `badge_key`
- `name`
- `description`
- `enabled`
- `unlocked`
- `unlocked_at`
- `metadata_json`
- `created_at`
- `updated_at`

唯一约束：

- `badge_key`

`user_progress`：

- `id`
- `level`
- `total_focus_seconds`
- `currency`
- `current_streak_days`
- `longest_streak_days`
- `last_focus_date`
- `created_at`
- `updated_at`

### 4.4 后端服务

- 建立 repository/service 分层。
- 建立 AutoMigrate。
- 建立 seed 逻辑：
  - 默认角色。
  - 默认场景 `study_room`。
  - 默认动作 key。
  - 默认巡查规则。
  - 默认等级配置。
  - 默认任务配置。
- seed 必须幂等，多次执行不能重复创建。
- JSON 字段统一以 string 存储，读写时校验合法 JSON。
- 密钥字段第一版可以使用可逆加密或本地简单加密；如果暂未实现加密，必须写 `TODO:` 并确保不下发前端。
- 枚举字段必须在 service 层校验，不能只依赖前端：
  - `mode`：`pomodoro`、`custom`、`infinite`
  - `patrol_frequency`：`slow`、`normal`、`high`
  - `session.result`：`success`、`left`、`failed`、`abandoned`
  - `finish_reason`：`countdown_complete`、`user_stop`、`max_warning`、`max_violation`、`page_unload`
  - `task_records.status`：`pending`、`claimable`、`claimed`
  - `camera_off_strategy` 和 `capture_failed_strategy`：`normal`、`suspicious`、`violation`、`uncertain`

### 4.5 API

实现基础 API：

- `GET /api/health`
- `GET /api/runtime/config`
- `GET /api/scenes`
- `GET /api/settings`
- `PUT /api/settings`

`GET /api/runtime/config` 响应示例：

```json
{
  "app": {
    "env": "development",
    "assetsBaseUrl": "/assets/",
    "serverTime": "2026-07-06T10:00:00+08:00"
  },
  "patrolRule": {
    "slow": { "minSeconds": 180, "maxSeconds": 480 },
    "normal": { "minSeconds": 60, "maxSeconds": 240 },
    "high": { "minSeconds": 30, "maxSeconds": 120 },
    "maxWarnings": 3,
    "maxViolations": 3,
    "cameraOffStrategy": "suspicious",
    "captureFailedStrategy": "uncertain",
    "userErrorMessage": "巡查系统暂不可用，请联系管理员处理。"
  },
  "character": {
    "characterKey": "default_oc",
    "name": "督学员"
  },
  "userSetting": {
    "mode": "pomodoro",
    "customDurationSeconds": 1500,
    "patrolFrequency": "normal",
    "backgroundVolume": 0.4,
    "actionVolume": 0.8,
    "uiVolume": 0.6,
    "quietPatrolEnabled": false,
    "screenFilter": "normal",
    "cameraEnabled": true
  }
}
```

`GET /api/settings` 返回当前 `user_settings`。`PUT /api/settings` 保存用户配置，请求字段与 `userSetting` 相同；服务端负责填充默认值并校验枚举和音量范围。

### 4.6 验收标准

- MySQL 8.0 空库启动后能自动建表。
- seed 后至少有一个可用场景和一组动作配置。
- `GET /api/runtime/config` 返回用户端所需配置，且不包含 API Key、数据库密码、appkey。
- `GET /api/scenes` 只返回启用场景。
- JSON 字段非法时保存失败并返回明确错误。
- 重复启动服务不会重复插入默认数据。

验收人签名:LFS

## 5. M2：管理端配置闭环

### 5.1 目标

开发隐藏管理端，使维护者可以用 appkey 修改本地配置，并保存到 MySQL。

### 5.2 鉴权

- 管理端入口：`/__admin?appkey=xxx`。
- 管理员 API 支持：
  - Header：`X-App-Key: xxx`
  - Query：`?appkey=xxx`
- appkey 来自 `APP_KEY` 或配置文件。
- 不做用户登录页。
- 不做账号注册。
- 不做 session 过期。
- appkey 错误时返回 `APPKEY_INVALID`。

### 5.3 管理端页面

管理端至少包含以下页面或 Tab：

- 运行状态。
- 运行时配置预览。
- 角色配置。
- 场景配置。
- 动作配置。
- 大模型配置。
- 巡查规则。
- MySQL 连接配置。

运行状态至少展示：

- 服务版本和启动时间。
- 当前监听地址。
- 当前生效的 MySQL 来源：`mysql_configs` 或 `DB_DSN`。
- 当前 MySQL 连接状态和最近错误。
- assets 目录和可访问状态。
- 当前启用场景、大模型和巡查规则摘要。

运行时配置预览调用 `GET /api/admin/runtime-config`，返回用户端最终可见配置加管理端诊断信息，但仍不得返回明文 API Key、数据库密码或 appkey。

### 5.4 场景配置

后端 API：

- `GET /api/admin/scenes`
- `POST /api/admin/scenes`
- `PUT /api/admin/scenes/:id`
- `DELETE /api/admin/scenes/:id`

前端能力：

- 列表展示。
- 新增场景。
- 编辑场景。
- 删除场景。
- 启用/禁用。
- 复制场景。
- 编辑背景路径。
- 编辑环境音路径。
- 编辑可用动作 key。
- 编辑模型结果映射 JSON。
- 编辑 metadata JSON。

校验：

- `sceneKey` 必填。
- `sceneKey` 只允许英文、数字、下划线。
- `sceneKey` 唯一。
- 启用场景必须配置背景。
- 启用场景必须至少绑定一个启用动作。

### 5.5 动作配置

后端 API：

- `GET /api/admin/actions`
- `POST /api/admin/actions`
- `PUT /api/admin/actions/:id`
- `DELETE /api/admin/actions/:id`

前端能力：

- 按场景筛选动作。
- 新增动作。
- 编辑动作。
- 删除动作。
- 启用/禁用。
- 复制动作。
- 编辑视频路径。
- 编辑 poster 路径。
- 编辑时长。
- 编辑 nextActionKey。
- 编辑 modelResultMap。
- 编辑 metadata JSON。
- 本地预览视频。

校验：

- `actionKey` 必填。
- 同一场景下 `actionKey` 唯一。
- 启用动作必须配置视频路径。
- `durationMs` 必须大于 0。
- JSON 字段必须合法。

### 5.6 大模型配置

后端 API：

- `GET /api/admin/model-config`
- `PUT /api/admin/model-config`
- `POST /api/admin/model-config/test`

前端能力：

- 编辑 provider。
- 编辑 baseUrl。
- 编辑 apiKey。
- 编辑 model。
- 编辑 timeoutMs。
- 编辑 maxImageWidth。
- 编辑 temperature。
- 编辑 prompt。
- 编辑 retryCount。
- 上传或选择测试图片并测试返回。

注意：

- API Key 默认脱敏显示。
- 保存时允许覆盖 API Key。
- 不得通过用户端 API 下发 API Key。
- TODO: 如果第一版没有图片测试上传能力，必须保留测试按钮位置并标记原因。

### 5.7 角色配置

后端 API：

- `GET /api/admin/characters`
- `POST /api/admin/characters`
- `PUT /api/admin/characters/:id`
- `DELETE /api/admin/characters/:id`

前端能力：

- 列表展示。
- 新增角色。
- 编辑角色基础信息。
- 启用/禁用。
- 配置头像路径。
- 编辑角色档案 JSON。
- 设置默认场景。
- 编辑 metadata JSON。

校验：

- `characterKey` 必填，只允许英文、数字、下划线。
- `characterKey` 唯一。
- 启用角色必须配置名称。
- `defaultSceneKey` 如果填写，必须指向存在的场景。

### 5.8 巡查规则配置

后端 API：

- `GET /api/admin/patrol-rule`
- `PUT /api/admin/patrol-rule`

配置项：

- 慢速巡查最小/最大秒数。
- 正常巡查最小/最大秒数。
- 高压巡查最小/最大秒数。
- 最大警告数。
- 最大案底数。
- 可疑是否计入警告。
- 摄像头关闭策略。
- 截图失败策略。
- 大模型超时重试次数。
- 错误提示文案。

校验：

- 每档巡查 `minSeconds` 必须大于 0。
- 每档巡查 `maxSeconds` 必须大于等于 `minSeconds`。
- 最大警告数和最大案底数必须大于 0。
- 摄像头关闭策略、截图失败策略只允许 `normal`、`suspicious`、`violation`、`uncertain`。

### 5.9 MySQL 连接配置

后端 API：

- `GET /api/admin/mysql-config`
- `PUT /api/admin/mysql-config`
- `POST /api/admin/mysql-config/test`
- `POST /api/admin/mysql-config/migrate`

前端能力：

- 查看当前连接状态。
- 编辑 host。
- 编辑 port。
- 编辑 databaseName。
- 编辑 username。
- 编辑 password。
- 编辑 charset。
- 编辑 timezone。
- 编辑 maxOpenConns。
- 编辑 maxIdleConns。
- 测试连接。
- 执行迁移。

规则：

- 数据库类型固定为 MySQL 8.0。
- 密码默认脱敏显示。
- 修改连接后允许提示重启服务生效。
- 第一版不要求无重启热切换。
- 测试连接只验证目标 MySQL 是否可连接和认证是否成功，不自动切换当前服务连接。
- 执行迁移只允许对当前生效连接执行；如果要对新配置执行迁移，必须先保存配置并重启。
- 保存配置后，`GET /api/admin/status` 应提示待重启状态。

### 5.10 验收标准

- 使用正确 appkey 可以进入管理端。
- 错误 appkey 无法访问管理员 API。
- 可以查看运行状态和运行时配置预览。
- 可以新增、编辑、删除、启用、禁用角色。
- 可以新增、编辑、删除、启用、禁用场景。
- 可以新增、编辑、删除、启用、禁用动作。
- 可以编辑并保存大模型配置。
- 可以编辑并保存巡查规则。
- 可以编辑并测试 MySQL 连接。
- 修改 MySQL 连接后能看到“重启后生效”提示。
- 管理端保存后刷新页面配置不丢失。
- 用户端 runtime config 能读取管理端保存的启用配置。

验收人签名：LFS

## 6. M3：用户端主界面和计时状态机

### 6.1 目标

完成用户端基础游戏界面和劳动状态流转，使用户不依赖巡查即可完整开始和结束一局。

### 6.2 页面

实现页面：

- `/` 主工作界面。
- `/settings` 用户配置。
- `/settlement` 结算页。

主工作界面包含：

- 当前场景背景。
- 角色演出层占位。
- 计时器。
- 状态文案。
- 警告条。
- 今日案底。
- 等级进度。
- 货币。
- 当前场景。
- 开始按钮。
- 暂停按钮。
- 继续按钮。
- 结束按钮。
- 配置入口。
- 档案入口。
- 任务入口。
- 全屏按钮。

### 6.3 状态机

实现状态：

- `idle`
- `working`
- `paused`
- `finished`
- `failed`

本阶段先预留但不完整实现：

- `patrolWarning`
- `patrolActive`
- `patrolResult`

如果巡查状态暂未接入，代码中必须标记：

```text
TODO: 巡查状态将在 M4 接入，当前只保留状态枚举和 UI 占位。
```

### 6.4 Session API

后端 API：

- `POST /api/session/start`
- `POST /api/session/pause`
- `POST /api/session/resume`
- `POST /api/session/finish`

`start` 请求字段：

- `sceneKey`
- `mode`
- `plannedDurationSeconds`
- `userConfig`

`finish` 请求字段：

- `sessionId`
- `finishReason`
- `actualFocusSeconds`

请求和响应示例：

```json
{
  "sceneKey": "study_room",
  "mode": "pomodoro",
  "plannedDurationSeconds": 1500,
  "userConfig": {
    "patrolFrequency": "normal",
    "cameraEnabled": true
  }
}
```

`POST /api/session/start` 成功响应：

```json
{
  "session": {
    "id": 1,
    "sceneKey": "study_room",
    "mode": "pomodoro",
    "plannedDurationSeconds": 1500,
    "startedAt": "2026-07-06T10:00:00+08:00",
    "status": "working",
    "warningCount": 0,
    "violationCount": 0
  }
}
```

`POST /api/session/pause`、`POST /api/session/resume` 请求：

```json
{
  "sessionId": 1
}
```

`POST /api/session/finish` 请求：

```json
{
  "sessionId": 1,
  "finishReason": "user_stop",
  "actualFocusSeconds": 620
}
```

`finishReason` 只允许：

- `countdown_complete`
- `user_stop`
- `max_warning`
- `max_violation`
- `page_unload`

`POST /api/session/finish` 成功响应：

```json
{
  "settlement": {
    "sessionId": 1,
    "result": "left",
    "actualFocusSeconds": 620,
    "patrolCount": 2,
    "warningCount": 1,
    "violationCount": 0,
    "earnedCurrency": 0,
    "levelBefore": 1,
    "levelAfter": 1,
    "currencyAfter": 30,
    "settlementAction": {
      "actionKey": "finish_success",
      "videoUrl": "assets/actions/study_room/finish_success.mp4",
      "posterUrl": ""
    }
  }
}
```

### 6.5 计时规则

- 番茄钟模式倒计时。
- 无限模式正计时。
- 暂停时停止计时。
- 继续时恢复计时。
- 主动结束时弹确认。
- 倒计时归零自动结算。
- 页面刷新后第一版可以回到 `idle` 并提示上一局异常结束。
- 页面刷新、关闭或前端无法恢复运行态时，前端应尽力调用 `finishReason=page_unload`；如果调用失败，下一次启动时后端将最近一条未结束 session 标记为 `abandoned`，但不进入正常结算页。
- `abandoned` session 只记录实际已知时长，不发放完成奖励，不触发“完成一次劳动”任务。

### 6.6 用户配置

实现配置项：

- 劳动模式。
- 自定义时长。
- 巡查频率。
- 背景音量。
- 动作视频音量。
- UI 音效音量。
- 轻声巡查。
- 画面滤镜。
- 摄像头开关。

用户配置 API：

- `GET /api/settings`
- `PUT /api/settings`

`PUT /api/settings` 请求示例：

```json
{
  "mode": "pomodoro",
  "customDurationSeconds": 1500,
  "patrolFrequency": "normal",
  "backgroundAudioKey": "library",
  "backgroundVolume": 0.4,
  "actionVolume": 0.8,
  "uiVolume": 0.6,
  "quietPatrolEnabled": false,
  "screenFilter": "normal",
  "cameraEnabled": true,
  "cameraDeviceId": ""
}
```

校验：

- 音量范围为 `0` 到 `1`。
- 自定义时长必须大于等于 300 秒。
- `screenFilter` 第一版只允许 `normal`、`grayscale`、`dark`。
- 前端可以用 localStorage 缓存最近设置，但页面加载后必须以服务端返回为准。

### 6.7 验收标准

- 用户能选择启用场景。
- 用户能开始一局劳动。
- 倒计时显示准确。
- 无限模式能正计时。
- 暂停后计时停止。
- 继续后计时恢复。
- 主动结束后进入结算页。
- 倒计时结束后进入成功结算页。
- session 记录保存到 MySQL。
- 用户配置刷新后仍保留。
- 页面刷新造成的未完成 session 能被标记为 `abandoned` 或 `page_unload`。
- 全屏按钮可用。

验收人签名：LFS

## 7. M4：巡查和动作播放闭环

### 7.1 目标

完成项目核心闭环：随机巡查、摄像头截图、后端判定、结果映射、动作视频播放、警告案底更新。

### 7.2 巡查调度

前端根据用户选择的巡查频率和后台规则随机生成下一次巡查时间。

规则：

- 慢速：默认 180-480 秒。
- 正常：默认 60-240 秒。
- 高压：默认 30-120 秒。
- 巡查结束后重新生成下一次巡查时间。
- 暂停时暂停巡查倒计时。
- 继续时恢复巡查倒计时。

### 7.3 巡查状态

实现完整状态：

- `patrolWarning`
- `patrolActive`
- `patrolResult`

流程：

- `working` 到达巡查时间。
- 进入 `patrolWarning`。
- 播放预告效果。
- 进入 `patrolActive`。
- 采集摄像头帧。
- 调用 `/api/patrol/check`。
- 进入 `patrolResult`。
- 播放动作视频。
- 更新统计。
- 返回 `working` 或进入 `failed`。

### 7.4 摄像头截图

前端实现：

- 请求摄像头权限。
- 显示摄像头预览。
- 使用 canvas 截取当前帧。
- 将截图压缩为 JPEG base64。
- 按 `maxImageWidth` 压缩图片。
- 截图失败时按巡查规则处理。

浏览器权限失败：

- 如果规则为放行，返回 `normal`。
- 如果规则为可疑，返回 `suspicious`。
- 如果规则为违规，返回 `violation`。

### 7.5 Patrol API

后端 API：

- `POST /api/patrol/check`

请求字段：

- `sessionId`
- `sceneKey`
- `imageBase64`
- `cameraEnabled`
- `manualViolation`
- `captureErrorCode`

响应字段：

- `status`
- `confidence`
- `reason`
- `action`
- `warningDelta`
- `violationDelta`
- `sessionSummary`

请求示例：

```json
{
  "sessionId": 1,
  "sceneKey": "study_room",
  "imageBase64": "data:image/jpeg;base64,/9j/...",
  "cameraEnabled": true,
  "manualViolation": false,
  "captureErrorCode": ""
}
```

摄像头关闭或截图失败时：

- `cameraEnabled=false` 且 `imageBase64` 为空，由后端按 `cameraOffStrategy` 生成状态并写入巡查记录。
- 截图失败时传 `captureErrorCode`，例如 `PERMISSION_DENIED`、`DEVICE_NOT_FOUND`、`CAPTURE_FAILED`，由后端按 `captureFailedStrategy` 生成状态并写入巡查记录。
- 如果后台策略要求调用模型，但 `imageBase64` 为空，后端必须返回 `CAMERA_FRAME_MISSING`，不得调用模型。

成功响应示例：

```json
{
  "status": "using_phone",
  "confidence": 0.91,
  "reason": "检测到用户手持手机且视线离开屏幕。",
  "action": {
    "actionKey": "patrol_phone",
    "name": "抓到玩手机",
    "videoUrl": "assets/actions/study_room/patrol_phone.mp4",
    "posterUrl": "assets/actions/study_room/patrol_phone.jpg",
    "durationMs": 6500
  },
  "warningDelta": 1,
  "violationDelta": 1,
  "sessionSummary": {
    "sessionId": 1,
    "patrolCount": 3,
    "warningCount": 2,
    "violationCount": 1,
    "failed": false,
    "finishReason": ""
  }
}
```

失败响应示例：

```json
{
  "error": {
    "code": "MODEL_CONFIG_MISSING",
    "message": "巡查系统暂不可用，请联系管理员处理。"
  }
}
```

落库规则：

- 每次调用 `/api/patrol/check` 都应写入 `patrol_records`，包括模型错误、摄像头错误和动作缺失。
- 模型错误不增加警告或案底，除非后台规则明确配置为 `suspicious` 或 `violation`。
- 动作缺失时返回错误并记录 `ACTION_CONFIG_MISSING`，不更新 session 的警告和案底。
- 达到失败条件时，后端响应 `sessionSummary.failed=true` 和对应 `finishReason`，前端随后进入失败结算。

### 7.6 大模型调用

后端实现：

- 读取启用的大模型配置。
- 未配置时返回 `MODEL_CONFIG_MISSING`。
- 构造 openai-compatible vision 请求。
- 设置超时。
- 设置重试次数。
- 要求模型返回 JSON。
- 解析并校验 JSON。
- 只信任白名单字段。
- 保存原始返回 JSON 到 `patrol_records.model_raw_json`。

返回 JSON 校验：

- `status` 必须在允许枚举中。
- `confidence` 必须是数字。
- `reason` 必须是字符串。
- `actionKey` 如果存在，必须能映射到启用动作。
- 如果模型返回额外字段，只允许保存到 `model_raw_json`，不得直接参与业务判断。

### 7.7 动作映射和播放

后端映射：

- 优先使用模型返回的合法 `actionKey`。
- 否则使用当前场景 `modelResultActionMap`。
- 再查找 `sceneKey + actionKey` 动作。
- 动作不存在时返回 `ACTION_CONFIG_MISSING`。

前端播放：

- 使用动作 `videoUrl` 播放本地视频。
- 支持 poster。
- 支持动作视频音量。
- 视频结束后返回工作状态。
- 如果视频加载失败，提示配置错误。

### 7.8 警告和案底

默认规则：

- `normal`：不增加。
- `suspicious`：警告 +1。
- `violation`：警告 +1，案底 +1。
- `using_phone`：警告 +1，案底 +1。
- `sleeping`：警告 +1，案底 +1。
- `absent`：警告 +1，案底 +1。
- `uncertain`：默认警告 +1。

失败：

- 警告达到上限进入 `failed`。
- 案底达到上限进入 `failed`。

### 7.9 验收标准

- 巡查能按随机间隔触发。
- 暂停时不会触发巡查。
- 巡查能请求摄像头权限。
- 摄像头截图能提交到后端。
- 未配置大模型时返回明确错误，不随机放行。
- 大模型返回正常结果时能映射动作。
- 动作视频能正确播放。
- 动作缺失时提示联系管理员并记录错误。
- 警告和案底按规则增加。
- 达到失败条件后进入失败结算。
- 巡查记录保存到 MySQL。

验收人签名_____

## 8. M5：成长、档案、任务和结算

### 8.1 目标

完成长期留存相关功能，让用户能看到劳动成果、历史记录、等级成长和每日任务。

### 8.2 结算页

结算类型：

- 光荣下班。
- 主动撤离。
- 禁闭失败。

展示内容：

- 本次劳动时长。
- 巡查次数。
- 违规次数。
- 获得货币。
- 等级进度变化。
- 结算评价文案。
- 结算动作视频。

### 8.3 档案页

API：

- `GET /api/profile/stats`

响应示例：

```json
{
  "today": {
    "date": "2026-07-06",
    "focusSeconds": 3600,
    "sessionCount": 2,
    "patrolCount": 5,
    "warningCount": 1,
    "violationCount": 0,
    "earnedCurrency": 30,
    "lastResult": "success"
  },
  "history": [
    {
      "sessionId": 1,
      "startedAt": "2026-07-06T10:00:00+08:00",
      "actualFocusSeconds": 1500,
      "patrolCount": 3,
      "violationCount": 0,
      "result": "success",
      "finishReason": "countdown_complete"
    }
  ],
  "chart": [
    { "date": "2026-07-06", "focusSeconds": 3600 }
  ],
  "progress": {
    "level": 2,
    "totalFocusSeconds": 3900,
    "currentLevelSeconds": 3600,
    "nextLevelSeconds": 36000,
    "currency": 30,
    "currentStreakDays": 1
  },
  "character": {
    "characterKey": "default_oc",
    "name": "督学员",
    "profile": {}
  }
}
```

展示：

- 今日专注时长。
- 今日巡查次数。
- 今日违规次数。
- 今日结算状态。
- 最近劳动记录。
- 近 7 日或 14 日专注时长图表。
- 等级进度。
- 角色档案。

### 8.4 等级

规则：

- 根据累计有效专注时长升级。
- 只升不降。
- 默认等级：
  - 1 级：初始。
  - 2 级：累计 1 小时。
  - 3 级：累计 10 小时。
  - 4 级：累计 50 小时。
  - 5 级：累计 100 小时。

### 8.5 货币

获取方式：

- 完成劳动。
- 完成任务。
- 无违规奖励。
- 连续天数奖励。

第一版不做消耗。

第一版默认奖励规则：

- 倒计时正常完成：`10` 货币。
- 无限模式达到最低有效时长后主动结束：每满 10 分钟 `5` 货币，最多按 60 分钟计算。
- 无案底完成劳动：额外 `20` 货币。
- 禁闭失败：不发放完成奖励，可保留任务以外的已获得货币。
- 主动撤离：记录有效时长，但不发放完成奖励。
- `abandoned`：不发放奖励。

TODO: 如果客户要求货币消费功能，另行增加商店或兑换配置，本版不实现。

### 8.6 任务

API：

- `GET /api/tasks/today`
- `POST /api/tasks/:id/claim`

默认任务：

- 完成一次劳动：`complete_session`，奖励 10。
- 连续专注 10 分钟：`focus_10_minutes`，奖励 5。
- 今日无案底完成劳动：`clean_session`，奖励 20。
- 开启摄像头完成一次劳动：`camera_session`，奖励 10。
- 查看角色档案：`view_profile`，奖励 3。

规则：

- 每日按服务器日期刷新。
- 同一任务每日只能领取一次。
- 完成后进入可领取状态。
- 领取后增加货币。

任务进度更新时机：

- session 结束后更新 `complete_session`、`focus_10_minutes`、`clean_session`、`camera_session`。
- 用户访问档案页后更新 `view_profile`。
- 任务记录以 `task_key + record_date` 唯一，重复触发只更新进度，不重复创建。

`GET /api/tasks/today` 响应示例：

```json
{
  "tasks": [
    {
      "taskKey": "complete_session",
      "name": "完成一次劳动",
      "progressValue": 1,
      "targetValue": 1,
      "rewardCurrency": 10,
      "status": "claimable"
    }
  ]
}
```

`POST /api/tasks/:id/claim` 成功响应：

```json
{
  "taskKey": "complete_session",
  "status": "claimed",
  "rewardCurrency": 10,
  "currencyAfter": 40
}
```

### 8.7 每日统计

每日统计更新时机：

- session 结束。
- 巡查结束。
- 任务领取。

统计字段：

- 日期。
- 专注秒数。
- session 数。
- 巡查次数。
- 警告次数。
- 违规次数。
- 获得货币。

每日刷新规则：

- 使用服务器本地日期，格式为 `YYYY-MM-DD`。
- 第一次读取当天任务或统计时，如果不存在当天记录则创建默认记录。
- 跨天不删除历史任务记录，只创建新日期记录。
- 连续天数根据有有效专注时长的日期计算，`abandoned` 不计入连续天数。

### 8.8 验收标准

- 成功完成劳动后能进入成功结算。
- 主动结束后能进入主动撤离结算。
- 失败后能进入禁闭失败结算。
- 结算数据和数据库记录一致。
- 档案页能显示今日统计。
- 档案页能显示最近历史记录。
- 等级能随有效时长提升。
- 货币能正确增加。
- 今日任务能展示。
- 已完成任务能领取奖励。
- 同一任务每日不能重复领取。

验收人签名_____

## 9. M6：打包、验收和交付整理

### 9.1 目标

将项目整理成可交付的本地单机版本，包含 exe、assets、配置示例、启动说明和验收记录。

### 9.2 打包

交付目录建议：

```text
supervisor-game/
  supervisor-game.exe
  .env.example
  assets/
    scenes/
    audio/
    actions/
  README.md
```

Mac 开发环境可先产出 darwin 可执行文件；交付 Windows 时需要交叉编译或在 Windows 环境构建。

TODO: 如果目标客户系统确定为 Windows，需要补充 Windows 构建脚本和启动说明。

### 9.3 构建脚本

需要提供：

- 前端构建脚本。
- Go 构建脚本。
- 一键 build 命令。
- MySQL 初始化说明。
- 启动说明。

### 9.4 配置文档

README 必须说明：

- 如何配置 appkey。
- 如何配置 MySQL 地址。
- 如何准备 assets。
- 如何启动服务。
- 如何访问用户端。
- 如何访问管理端。
- 如何测试大模型配置。
- 常见错误处理。

### 9.5 最终回归

回归清单：

- 新库启动。
- 默认种子数据。
- 管理端 appkey。
- MySQL 配置测试。
- 场景配置。
- 动作配置。
- 大模型配置。
- 用户开始劳动。
- 随机巡查。
- 摄像头截图。
- 动作视频播放。
- 成功结算。
- 主动撤离结算。
- 失败结算。
- 档案统计。
- 任务领取。
- 单 exe 静态页面访问。
- assets 本地路径访问。

### 9.6 验收标准

- 交付包解压后按 README 可以启动。
- 用户端可以通过 localhost 打开。
- 管理端可以通过 appkey 打开。
- MySQL 连接配置可测试。
- 核心游戏流程完整可用。
- 不依赖对象存储。
- 不依赖外部后台服务，除大模型供应商接口外。
- 素材缺失时能给出清晰配置错误。
- 所有未完成或暂缓内容均有 `TODO:` 标记。

验收人签名_____

## 10. 开发过程约束

### 10.1 提交要求

- 每个里程碑至少一个可运行提交。
- 提交前必须执行：
  - `go test ./...`
  - `npm run build`
- 涉及数据库模型变更时必须确认 AutoMigrate 行为。
- 涉及前端交互时必须本地浏览器验证。

### 10.2 TODO 要求

以下情况必须写 `TODO:`：

- 依赖客户素材但素材未提供。
- 依赖大模型供应商但接口不可用。
- 目标浏览器兼容性暂未验证。
- 当前阶段只做占位。
- 存在已知体验问题但不影响主流程。
- 存在安全简化实现。
- 需要后续客户确认文案、角色口吻或视觉风格。

禁止使用含糊表述替代 TODO，例如：

- 后面再说。
- 以后优化。
- 暂时这样。
- 待完善。

必须改为：

```text
TODO: 当前阶段仅实现基础版本，后续在客户确认后补充具体内容。
```

### 10.3 不做事项

开发过程中不得擅自加入以下范围：

- 多用户账号。
- 多管理员权限。
- 支付。
- 排行榜。
- 战友系统。
- 在线人数。
- 对象存储。
- 多租户。
- 数据库类型切换。
- 本地视觉模型训练。
- 移动端深度适配。

如客户新增以上需求，必须单独评估并更新 PRD 和开发计划。

## 11. 风险和处理

### 11.1 大模型不稳定

风险：

- 超时。
- 返回非 JSON。
- 判断不准。

处理：

- 后端严格校验 JSON。
- 不做随机假结果。
- 明确提示联系管理员。
- 保存错误记录。

### 11.2 摄像头权限失败

风险：

- 用户拒绝权限。
- 浏览器不支持。
- 设备不可用。

处理：

- 用户端明确提示。
- 按后台规则处理为放行、可疑或违规。
- 保存巡查记录。

### 11.3 素材缺失

风险：

- 视频路径错误。
- 背景路径错误。
- 客户未提供完整动作包。

处理：

- 管理端校验路径。
- 用户端播放失败时提示配置错误。
- 缺失素材必须写 `TODO:`。

### 11.4 MySQL 配置错误

风险：

- 地址不通。
- 账号密码错误。
- 数据库不存在。

处理：

- 管理端提供测试连接。
- 启动时输出数据库状态。
- README 给出排查步骤。

## 12. 交付确认

开发完成后，应逐项完成 M0 到 M6 的验收签名。未签名里程碑视为未正式验收通过。
