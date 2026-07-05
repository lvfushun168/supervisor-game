# OC 督学专注游戏 PRD

## 1. 项目定位

本项目是一款个人定制的本地网页专注游戏。交付形态为一个 Go 单 exe 本地服务，启动后用户通过浏览器访问 `localhost` 使用。前端使用 Vite + Vue 3，后端使用 Gin + GORM，数据库固定使用 MySQL 8.0。

用户在学习或工作时开启倒计时，系统按随机间隔触发巡查。巡查时前端采集摄像头截图并提交给本地后端，后端调用管理员配置的大模型视觉接口进行判定，再根据判定结果播放对应动作视频。系统通过警告、案底、结算、等级和档案记录制造“被监督”的沉浸感。

本项目是单机定制版，不做 SaaS、对象存储、多租户、复杂账号体系或商业运营后台。管理员路由只用于开发者或项目维护者修改配置后重新打包/交付，不面向终端客户运营。

## 2. 技术目标

- Go 单 exe 本地服务。
- Gin 提供 HTTP API。
- GORM 连接 MySQL 8.0。
- 数据库类型固定为 MySQL 8.0，但 MySQL 地址、端口、库名、账号和密码必须可配置。
- Vite + Vue 3 构建用户端和隐藏管理端。
- 前端构建产物嵌入 Go exe。
- 素材使用本地相对路径，随交付包一起提供。
- 管理员使用固定 appkey 进入，不做账号注册、密码找回、session 过期或锁定。

## 3. 目标用户

- 需要自习、备考、居家办公的人。
- 喜欢 OC 角色陪伴、训诫、监督、打卡玩法的人。
- 希望获得“仪式感”和“被看着”的专注体验，而不是严肃考试监控。

## 4. 版本范围

### 4.1 本版包含

- 用户端网页游戏。
- 主工作界面。
- 场景选择。
- 倒计时模式。
- 无限劳动模式。
- 随机巡查。
- 摄像头授权和截图。
- 服务器侧大模型视觉判定。
- 动作视频播放。
- 静态/动态场景背景。
- 环境音。
- 警告、案底、失败、结算。
- 等级、货币、今日统计、历史记录。
- 简化任务和徽章。
- 隐藏管理员路由。
- 场景配置。
- 动作配置。
- 巡查规则配置。
- 大模型配置。
- 固定 appkey 管理员入口。
- MySQL 数据库存档。

### 4.2 本版不包含

- 多用户账号体系。
- 多管理员 CRUD。
- session 过期、登录锁定、权限分级。
- 数据库类型切换，数据库类型固定为 MySQL 8.0。
- 对象存储。
- 商业支付。
- 真实排行榜。
- 战友系统。
- 邀请裂变。
- 在线人数统计。
- 多租户隔离。
- 移动端深度适配。
- TTS 音色训练。
- 本地视觉模型训练。
- 100% 行为识别准确率承诺。

## 5. 素材策略

素材采用本地文件交付，不使用对象存储。

推荐目录：

```text
assets/
  scenes/
    study_room.jpg
    study_room.mp4
  audio/
    library_noise.mp3
  actions/
    study_room/
      patrol_enter.mp4
      patrol_normal.mp4
      patrol_phone.mp4
      patrol_sleeping.mp4
      patrol_absent.mp4
      finish_success.mp4
      fail.mp4
```

配置中保存素材相对路径，例如：

```text
assets/actions/study_room/patrol_phone.mp4
```

素材可以随 exe 放在同级目录。若素材确定后不再频繁替换，也可以在后续构建中嵌入 exe；第一版优先使用外置 `assets/`，方便替换视频和背景。

动作视频包由朋友提供，是“台词 + 语音 + 角色动作 + 表情 + 音效 + 字幕”的完整视频文件。后台不单独拆分配置台词、语音或字幕。

## 6. 页面结构

### 6.1 主工作界面

第一屏就是可用游戏界面，不做营销页。

主界面元素：

- 场景背景：静态图片或动态视频。
- 角色演出层：播放动作视频。
- 计时器：倒计时或正计时。
- 状态文案：待命、劳动中、巡查预告、巡查中、结算。
- 警告条：例如 `0/3`。
- 今日案底。
- 等级和等级进度。
- 货币数量。
- 当前场景。
- 主按钮：开始、暂停、继续、结束。
- 功能入口：配置、档案、任务、全屏。

### 6.2 场景选择

用户开始劳动前选择场景。

规则：

- 只展示启用场景。
- 默认选中后台配置的默认场景。
- 场景决定背景、环境音、动作集合和模型结果映射。
- 劳动中不可切换场景。
- 若没有启用场景，提示“暂无可用场景，请联系管理员配置”。

### 6.3 用户配置页

用户配置页用于调整游戏体验。

配置项：

- 劳动模式：番茄钟、自定义时长、无限模式。
- 时长：15/25/45/60 分钟。
- 巡查频率：慢速、正常、高压。
- 背景音：静音、图书馆、雨声、风声、自定义配置音频。
- 音量：背景音量、动作视频音量、UI 音效音量。
- 轻声巡查：优先使用轻声版动作。
- 画面滤镜：正常、黑白、暗色压迫。
- 摄像头：开启/关闭、预览、设备选择。
- 数据维护：清空今日记录、清空全部本地记录。

用户配置保存到服务器数据库。前端可用 localStorage 做非敏感缓存，但不作为主存储。

### 6.4 档案页

展示用户专注成果和历史记录。

内容：

- 今日专注时长。
- 今日巡查次数。
- 今日违规次数。
- 今日结算状态。
- 最近劳动记录。
- 近 7 日或 14 日专注时长图表。
- 等级进度。
- 角色档案。

历史记录字段：

- 日期时间。
- 劳动时长。
- 巡查次数。
- 违规次数。
- 结束原因：光荣下班、主动撤离、禁闭失败。

### 6.5 任务页

任务系统做轻量版本，优先内置几条常用任务。

示例：

- 完成一次劳动：奖励 10 货币。
- 连续专注 10 分钟：奖励 5 货币。
- 今日无案底完成劳动：奖励 20 货币。
- 开启摄像头完成一次劳动：奖励 10 货币。
- 查看角色档案：奖励 3 货币。

规则：

- 每日按服务器日期刷新。
- 同一任务每日只可领取一次。
- 完成后进入可领取状态。
- 领取后增加货币。

### 6.6 结算页

劳动结束后显示结算。

结算类型：

- 光荣下班：倒计时正常结束。
- 主动撤离：用户主动结束。
- 禁闭失败：警告或案底达到上限。

结算内容：

- 本次劳动时长。
- 巡查次数。
- 违规次数。
- 获得货币。
- 等级进度变化。
- 结算动作视频。

### 6.7 隐藏管理员路由

管理员路由使用：

```text
/__admin?appkey=配置值
```

说明：

- appkey 固定配置在服务端环境变量或配置文件中。
- 不做账号注册。
- 不做多管理员。
- 不做登录过期。
- 不做错误锁定。
- 只用于开发者或维护者修改配置。

管理员功能：

- 场景配置。
- 动作配置。
- 巡查规则配置。
- 大模型配置。
- MySQL 连接配置。
- 角色配置。
- 运行时配置预览。

后台可以简单、直接、表单化，不追求复杂运营后台体验。

## 7. 核心状态机

游戏状态：

- `idle`：待命。
- `working`：劳动中。
- `paused`：暂停。
- `patrolWarning`：巡查预告。
- `patrolActive`：巡查中。
- `patrolResult`：巡查结果展示。
- `finished`：已完成。
- `failed`：失败。

状态流转：

```text
idle
  -> working
  -> patrolWarning
  -> patrolActive
  -> patrolResult
  -> working
  -> finished

working -> paused -> working
working -> failed
working -> finished
working -> idle
```

规则：

- 点击开始后进入 `working`。
- `working` 状态计时，并安排下一次巡查。
- 到达巡查时间后进入 `patrolWarning`。
- `patrolWarning` 播放预告音效或暗化屏幕。
- `patrolActive` 显示角色，采集摄像头帧并提交后端判定。
- `patrolResult` 根据判定播放动作视频并更新统计。
- 结果展示结束后回到 `working`。
- 警告或案底达到上限进入 `failed`。
- 倒计时归零进入 `finished`。
- 用户主动结束进入 `finished`，类型为主动撤离。

## 8. 巡查逻辑

### 8.1 巡查触发

每次进入 `working` 后，根据巡查频率随机生成下一次巡查时间。

默认区间：

```text
慢速：180-480 秒
正常：60-240 秒
高压：30-120 秒
```

同一轮巡查结束后重新生成下一次巡查时间。

### 8.2 巡查流程

```text
巡查触发
  -> 读取当前 sceneKey
  -> 进入 patrolWarning
  -> 播放预告效果
  -> 进入 patrolActive
  -> 前端采集摄像头当前帧
  -> 前端提交 sceneKey、sessionId、截图帧到 /api/patrol/check
  -> 后端读取启用的大模型配置
  -> 调用大模型视觉接口
  -> 校验返回 JSON
  -> 根据 scene.modelResultActionMap 映射 actionKey
  -> 读取 sceneKey + actionKey 对应动作配置
  -> 写入巡查记录
  -> 返回判定结果和动作配置
  -> 前端播放动作视频
  -> 更新警告、案底、等级和任务进度
```

### 8.3 判定类型

基础类型：

- `normal`：正常。
- `suspicious`：可疑。
- `violation`：违规。
- `using_phone`：玩手机。
- `sleeping`：打瞌睡。
- `absent`：离岗。
- `uncertain`：不确定。

推荐映射：

```json
{
  "normal": "patrol_normal",
  "suspicious": "patrol_suspicious",
  "violation": "patrol_violation",
  "using_phone": "patrol_phone",
  "sleeping": "patrol_sleeping",
  "absent": "patrol_absent",
  "uncertain": "patrol_suspicious"
}
```

### 8.4 错误处理

- 未配置可用大模型：返回 `MODEL_CONFIG_MISSING`，不得随机放行。
- 摄像头关闭：按后台规则处理，可放行、可疑或违规。
- 截图失败：按后台规则处理，可返回 `uncertain` 或 `absent`。
- 大模型超时：返回错误或按配置重试，不使用随机判定替代。
- 大模型返回非法 JSON：返回错误并记录，不使用随机判定替代。
- 场景动作缺失：返回 `ACTION_CONFIG_MISSING`。

用户端统一提示：

```text
巡查系统暂不可用，请联系管理员处理。
```

### 8.5 大模型返回 JSON

推荐格式：

```json
{
  "status": "normal",
  "confidence": 0.86,
  "reason": "用户坐在桌前，未发现明显玩手机、离岗或睡觉行为。",
  "objects": ["person", "desk", "phone"],
  "actionKey": "patrol_normal"
}
```

后端只信任白名单字段：

- `status`
- `confidence`
- `reason`
- `objects`
- `actionKey`

若模型返回 `actionKey`，后端仍需检查该动作是否存在且启用；否则按 `status` 映射默认动作。

## 9. 警告、案底和成长

默认规则：

- `normal`：不增加警告和案底。
- `suspicious`：增加 1 次警告。
- `violation`：增加 1 次警告和 1 次案底。
- `using_phone`：增加 1 次警告和 1 次案底。
- `sleeping`：增加 1 次警告和 1 次案底。
- `absent`：增加 1 次警告和 1 次案底。
- `uncertain`：默认增加 1 次警告，可后台调整。

失败规则：

- 警告达到 3 次，进入禁闭失败。
- 案底达到 3 次，进入禁闭失败。

可配置项：

- 最大警告数，默认 3。
- 最大案底数，默认 3。
- 可疑是否计入警告，默认是。
- 违规是否直接失败，默认否。

等级规则：

- 根据累计有效专注时长升级。
- 等级只升不降。

示例：

- 1 级：初始。
- 2 级：累计 1 小时。
- 3 级：累计 10 小时。
- 4 级：累计 50 小时。
- 5 级：累计 100 小时。

货币获取：

- 完成劳动。
- 完成任务。
- 无违规奖励。
- 连续天数奖励。

第一版不做真实消耗，只记录获得数量。

## 10. 计时逻辑

### 10.1 番茄钟模式

- 用户选择时长。
- 点击开始后倒计时。
- 暂停时停止计时。
- 继续时恢复计时。
- 倒计时归零后自动结算。

### 10.2 无限模式

- 点击开始后正计时。
- 用户主动结束时结算。
- 可设置最低有效时长，例如 5 分钟。

### 10.3 有效专注时长

- `paused` 状态不计入。
- `failed` 可按比例扣减，例如只记录 50%。
- 主动撤离可按实际时长记录，但不给额外奖励。

## 11. 管理员配置

### 11.1 场景配置

场景字段：

```json
{
  "sceneKey": "study_room",
  "name": "自习室",
  "enabled": true,
  "description": "适合考研、自习、刷题场景。",
  "backgroundType": "image",
  "backgroundUrl": "assets/scenes/study_room.jpg",
  "backgroundVideoUrl": "",
  "backgroundPosterUrl": "",
  "ambientAudioUrl": "assets/audio/library_noise.mp3",
  "defaultActionKey": "patrol_enter",
  "availableActionKeys": [
    "patrol_enter",
    "patrol_normal",
    "patrol_phone",
    "patrol_sleeping",
    "patrol_absent",
    "finish_success"
  ],
  "modelResultActionMap": {
    "normal": "patrol_normal",
    "using_phone": "patrol_phone",
    "sleeping": "patrol_sleeping",
    "absent": "patrol_absent",
    "uncertain": "patrol_suspicious"
  },
  "metadata": {
    "theme": "dark_academic",
    "hudTone": "strict"
  }
}
```

功能：

- 新增场景。
- 编辑场景。
- 删除场景。
- 启用/禁用场景。
- 复制场景。
- 填写静态背景路径。
- 填写动态背景视频路径。
- 填写环境音路径。
- 配置可用动作。
- 配置模型结果到动作 key 的映射。

校验：

- `sceneKey` 必填，只允许英文、数字、下划线。
- `sceneKey` 不可重复。
- 启用场景必须配置背景。
- 启用场景必须至少绑定一个启用动作。
- 场景绑定的动作 key 必须存在。

### 11.2 动作配置

动作以 `sceneKey + actionKey` 唯一。

字段：

```json
{
  "sceneKey": "study_room",
  "actionKey": "patrol_phone",
  "name": "抓到玩手机",
  "enabled": true,
  "priority": 100,
  "videoUrl": "assets/actions/study_room/patrol_phone.mp4",
  "posterUrl": "assets/actions/study_room/patrol_phone.jpg",
  "durationMs": 6500,
  "nextActionKey": "exit",
  "modelResultMap": ["using_phone"],
  "localRuleMap": ["manual_violation"],
  "metadata": {
    "mood": "angry",
    "cameraShake": true,
    "screenFilter": "dark"
  }
}
```

功能：

- 新增动作。
- 编辑动作。
- 删除动作。
- 启用/禁用动作。
- 复制动作。
- 调整优先级。
- 填写动作视频路径。
- 填写 poster 路径。
- 编辑 metadata JSON。
- 预览动作。

校验：

- `actionKey` 必填，只允许英文、数字、下划线。
- 同一场景下 `actionKey` 不可重复。
- `sceneKey` 必须存在。
- 启用动作必须配置动作视频路径。
- `metadata` 必须是合法 JSON。
- `durationMs` 必须大于 0。

### 11.3 大模型配置

第一版只做一个启用中的 openai-compatible 视觉接口配置。

字段：

```json
{
  "provider": "openai-compatible",
  "name": "主视觉判定接口",
  "enabled": true,
  "baseUrl": "https://api.example.com/v1",
  "apiKey": "sk-***",
  "model": "vision-model-name",
  "timeoutMs": 15000,
  "maxImageWidth": 768,
  "temperature": 0,
  "prompt": "你是专注巡查助手，请判断用户是否玩手机、离岗、打瞌睡或正常学习，只返回 JSON。",
  "responseSchema": {
    "status": "normal | using_phone | absent | sleeping | uncertain",
    "confidence": "number",
    "reason": "string"
  }
}
```

功能：

- 编辑接口配置。
- 启用/禁用接口。
- 测试连接。
- 上传测试图片并查看返回 JSON。
- 配置 prompt。
- 配置超时时间。
- 配置失败重试次数。

安全：

- API Key 不下发前端。
- 用户端只调用本地服务器判定接口。
- 本地服务器再调用大模型供应商。
- 管理员页面中 API Key 默认脱敏显示。

### 11.4 巡查规则配置

字段：

- 巡查频率区间。
- 最大警告数。
- 最大案底数。
- 可疑是否计入警告。
- 摄像头关闭策略。
- 截图失败策略。
- 大模型超时重试次数。
- 大模型错误提示文案。

### 11.5 MySQL 连接配置

数据库类型固定为 MySQL 8.0，不提供 SQLite/PostgreSQL 等数据库切换。

MySQL 连接信息必须可配置：

- 主机地址。
- 端口。
- 数据库名。
- 用户名。
- 密码。
- 字符集，默认 `utf8mb4`。
- 时区，默认 `Local`。
- 最大连接数。
- 最大空闲连接数。

配置方式：

- 第一优先级：管理员后台保存的 MySQL 连接配置。
- 第二优先级：启动环境变量 `DB_DSN`。
- 第三优先级：`.env` 文件。

管理员后台功能：

- 查看当前数据库连接状态。
- 编辑 MySQL 连接信息。
- 测试 MySQL 连接。
- 保存配置。
- 执行数据库迁移。

注意：

- 数据库密码不得下发给用户端。
- 管理员页面中数据库密码默认脱敏显示。
- 修改 MySQL 连接配置后，允许提示用户重启本地服务生效；第一版不要求无重启热切换。

## 12. 数据存储

使用 MySQL 8.0 作为主存储。数据库类型不切换，但连接信息需要支持配置。

保存内容：

- 用户配置。
- 今日统计。
- 历史记录。
- 货币数量。
- 任务领取状态。
- 等级进度。
- 角色配置。
- 场景配置。
- 动作配置。
- 大模型配置。
- 巡查规则。
- 劳动局记录。
- 巡查记录。

每日刷新逻辑：

- 根据服务器日期判断是否新的一天。
- 新的一天重置今日统计和每日任务。
- 历史记录保留。

### 12.1 数据表

建议表：

- `app_settings`：全局配置。
- `characters`：角色基础信息。
- `scenes`：场景配置。
- `action_configs`：动作配置。
- `model_configs`：大模型接口配置。
- `patrol_rules`：巡查规则。
- `mysql_configs`：MySQL 连接配置。
- `work_sessions`：劳动局记录。
- `patrol_records`：巡查记录。
- `daily_stats`：每日统计。
- `tasks`：任务配置。
- `task_records`：任务完成记录。
- `badges`：徽章配置。
- `user_progress`：用户进度。

不需要：

- `admin_users`
- 多数据库类型配置表
- 多租户相关表

### 12.2 scenes 表字段

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

### 12.3 action_configs 表字段

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

唯一约束：

- `scene_key + action_key`

### 12.4 model_configs 表字段

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

### 12.5 mysql_configs 表字段

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

### 12.6 work_sessions 表字段

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

### 12.7 patrol_records 表字段

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

## 13. API

### 13.1 用户端 API

- `GET /api/health`
- `GET /api/runtime/config`
- `GET /api/scenes`
- `POST /api/session/start`
- `POST /api/session/pause`
- `POST /api/session/resume`
- `POST /api/session/finish`
- `POST /api/patrol/check`
- `GET /api/profile/stats`
- `GET /api/tasks/today`
- `POST /api/tasks/:id/claim`

### 13.2 管理员 API

管理员 API 使用 appkey 校验。可以通过 header 或 query 传入：

```text
X-App-Key: xxx
```

或：

```text
?appkey=xxx
```

接口：

- `GET /api/admin/status`
- `GET /api/admin/runtime-config`
- `GET /api/admin/scenes`
- `POST /api/admin/scenes`
- `PUT /api/admin/scenes/:id`
- `DELETE /api/admin/scenes/:id`
- `GET /api/admin/actions`
- `POST /api/admin/actions`
- `PUT /api/admin/actions/:id`
- `DELETE /api/admin/actions/:id`
- `GET /api/admin/model-config`
- `PUT /api/admin/model-config`
- `POST /api/admin/model-config/test`
- `GET /api/admin/patrol-rule`
- `PUT /api/admin/patrol-rule`
- `GET /api/admin/mysql-config`
- `PUT /api/admin/mysql-config`
- `POST /api/admin/mysql-config/test`
- `POST /api/admin/mysql-config/migrate`

### 13.3 统一错误结构

```json
{
  "error": {
    "code": "MODEL_CONFIG_MISSING",
    "message": "巡查系统暂不可用，请联系管理员处理。"
  }
}
```

常用错误码：

- `APPKEY_INVALID`
- `CONFIG_INVALID`
- `SCENE_NOT_FOUND`
- `SCENE_CONFIG_MISSING`
- `ACTION_CONFIG_MISSING`
- `MODEL_CONFIG_MISSING`
- `MODEL_TIMEOUT`
- `MODEL_INVALID_RESPONSE`
- `CAMERA_FRAME_MISSING`
- `SESSION_NOT_FOUND`
- `SESSION_STATE_INVALID`

## 14. 前端实现口径

### 14.1 路由

- `/`：主工作界面。
- `/settings`：用户配置。
- `/profile`：档案。
- `/tasks`：任务。
- `/settlement`：结算。
- `/__admin`：隐藏管理端。

### 14.2 状态管理

前端维护运行态：

- 当前状态。
- 当前 session。
- 当前场景。
- 当前配置。
- 当前计时。
- 下一次巡查时间。
- 警告数。
- 案底数。
- 当前播放动作。

服务器保存最终记录和统计数据。前端刷新后可以从服务器恢复最近一次未结束 session；第一版若恢复复杂，可刷新后回到 idle 并提示上一局异常结束。

### 14.3 摄像头截图

- 使用 `navigator.mediaDevices.getUserMedia` 获取摄像头。
- 使用 `canvas` 截取当前帧。
- 压缩为 JPEG base64。
- 提交给 `/api/patrol/check`。
- 图片最大宽度按后台模型配置压缩，例如 768。

## 15. 验收标准

### 15.1 功能验收

- 可以启动 Go exe 并通过 localhost 打开页面。
- 可以配置 MySQL 地址、端口、库名、账号和密码。
- 可以测试 MySQL 8.0 连接。
- 可以正常连接 MySQL 8.0。
- 可以正常开始、暂停、继续、结束劳动。
- 用户开始劳动前可以选择场景。
- 场景切换后能更换背景和环境音。
- 倒计时准确。
- 随机巡查能按配置区间触发。
- 巡查时能采集摄像头截图。
- 巡查能调用本地后端判定接口。
- 未配置可用大模型时不得随机放行，应提示联系管理员。
- 大模型返回结果后能映射到动作视频。
- 场景缺失对应动作时应提示联系管理员并记录错误。
- 动作视频能正常播放。
- 警告和案底能正确增加。
- 达到失败条件后进入失败结算。
- 正常完成后进入成功结算。
- 主动结束后进入主动撤离结算。
- 历史记录能保存到 MySQL。
- 档案页能显示今日和历史记录。
- 任务能完成和领取奖励。
- 等级能随有效时长增长。
- 管理员可以通过 appkey 进入隐藏后台。
- 管理员可以编辑场景、动作、大模型和巡查规则。
- 管理员可以编辑和测试 MySQL 连接配置。
- 配置刷新页面后仍保留。

### 15.2 素材验收

- 背景图或背景视频显示正常。
- 环境音可播放和调节音量。
- 动作视频包能正常播放。
- 视频中的台词、语音、角色动作与对应事件匹配。
- 素材路径使用本地相对路径可访问。

### 15.3 兼容验收

- Chrome 桌面端可用。
- Edge 桌面端尽量可用。
- Safari/iPhone 不作为默认强验收项。

## 16. 开发阶段

### 16.1 第一阶段：数据和配置

- GORM models。
- AutoMigrate。
- 默认种子配置。
- appkey 管理后台鉴权。
- 场景配置 API。
- 动作配置 API。
- 大模型配置 API。
- 巡查规则 API。
- MySQL 连接配置 API。
- runtime config API。

### 16.2 第二阶段：用户端主循环

- 主工作界面。
- 用户配置页。
- 场景选择。
- 计时状态机。
- session start/pause/resume/finish。
- 背景和环境音。
- 动作视频播放。

### 16.3 第三阶段：巡查闭环

- 随机巡查调度。
- 摄像头授权。
- 截图上传。
- `/api/patrol/check`。
- 大模型调用。
- 结果映射。
- 巡查记录。
- 警告、案底、失败。

### 16.4 第四阶段：成长和收尾

- 档案页。
- 任务页。
- 等级和货币。
- 结算页。
- 管理后台体验打磨。
- 打包文档。

## 17. 对标截图说明

原对标截图只用于内部需求分析，不作为美术照抄依据。实际交付应替换为客户自有角色、背景、UI 风格与文案。

本项目保留对标游戏的核心循环：

```text
专注计时
  -> 随机巡查
  -> 截图判定
  -> 播放角色动作
  -> 更新警告/案底/奖励
  -> 结算与成长记录
```

删除对标游戏中的社区和运营功能：

- 排行榜。
- 战友。
- 征召。
- 在线人数。
- 支付商城。
- 多端同步。

## 18. 交付口径

本版本是单机定制版，主打“本地运行、简单配置、随机巡查、截图判定、动作反馈、成长记录”。

识别效果取决于配置的大模型能力、摄像头角度、光线和用户环境，不承诺 100% 准确。若未配置大模型，系统必须明确报错，不得使用随机结果伪装识别。
