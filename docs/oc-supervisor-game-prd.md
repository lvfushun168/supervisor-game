# OC 督学专注游戏 PRD

## 1. 项目定位

本项目是一款个人定制网页专注游戏。用户侧以轻量网页形式运行，可通过本机 localhost 打开；管理侧提供一个不对外公开的隐藏路由，用于配置角色动作视频包、场景、巡查规则、大模型接口、数据库连接和服务器参数。用户在学习或工作时开启倒计时，角色会随机巡查，通过动作视频包、摄像头判定和结算反馈制造“被监督”的沉浸感。

本版本定位为完整交付版，不再区分 MVP。基础玩法一次性交付完整；识别能力强依赖管理员配置的大模型接口。若未配置可用大模型，巡查判定不得降级为随机或本地规则，应直接报错并提示用户联系管理员。

## 2. 目标用户

- 需要自习、备考、居家办公的人。
- 喜欢 OC 角色陪伴、训诫、监督、打卡玩法的人。
- 希望获得“仪式感”和“被看着”的专注体验，而非严肃考试监控。

## 3. 版本范围

### 3.1 完整版包含

- 用户端网页游戏，可通过 localhost 或部署地址运行。
- 主工作界面。
- 倒计时/无限劳动模式。
- 随机巡查演出。
- 角色动作视频包接入。
- 场景选择。
- 场景静态/动态背景配置。
- 警告、案底、结算。
- 配置页。
- 档案页。
- 任务与徽章。
- 服务器数据库存档。
- 全屏按钮。
- PWA 安装提示。
- 摄像头授权、截图和巡查帧采集。
- 可配置大模型视觉判定。
- 隐藏管理员路由。
- 动作 key-value 配置后台。
- 场景 key-value 配置后台。
- 动作视频包/判定 JSON 配置后台。
- 大模型接口配置后台。
- 数据库连接配置。
- 管理员增删改查。

### 3.2 完整版不包含

- 真实排行榜。
- 战友系统。
- 邀请裂变。
- 真实商城支付。
- 在线人数统计。
- 多用户社交系统。
- 商业支付系统。
- TTS 音色训练。
- 100% 行为识别准确率承诺。

## 4. 页面结构

### 4.1 主工作界面

主界面由背景图、角色演出层、右侧 HUD、底部/侧边操作按钮组成。

功能元素：

- 场景背景：由当前场景配置决定，可为静态图片或动态视频背景。
- 角色演出层：巡查时出现角色立绘或动画。
- 计时器：显示当前劳动剩余时间或累计时间。
- 状态文案：显示“待命”“劳动中”“巡查中”“暂停”“结算”等状态。
- 警告条：显示当前警告次数，例如 0/3。
- 今日案底：显示今日违规次数。
- 徽章等级：显示当前等级。
- 货币：显示本地虚拟货币。
- 当前场景：显示当前选择的场景名称。
- 主按钮：开始劳动、暂停、继续、结束劳动。
- 功能入口：配置、档案、任务、全屏。

### 4.1.1 场景选择

用户开始劳动前需要选择场景。

场景选择规则：

- 仅展示管理员启用的场景。
- 默认选中管理员设置的默认场景。
- 场景切换会更换背景、环境音、动作集合和模型结果映射。
- 劳动进行中不可切换场景。
- 若当前没有可用场景，用户端提示“暂无可用场景，请联系管理员配置”。

### 4.2 配置页

配置页用于提供“系统感”，同时满足基础可调参数。

配置项：

- 劳动模式：
  - 番茄钟模式：默认 25 分钟。
  - 自定义时长：15/25/45/60 分钟。
  - 无限模式：正计时，不自动结束。
- 巡查频率：
  - 慢速：3-8 分钟一次。
  - 正常：1-4 分钟一次。
  - 高压：30 秒-2 分钟一次。
- 背景音乐/白噪：
  - 静音。
  - 图书馆。
  - 雨声。
  - 风声。
  - 客户自定义音频。
- 音量：
  - 背景音量。
  - 动作视频音量。
  - UI 音效音量。
- 巡查轻声模式：
  - 开启后，正常巡查优先使用轻声版动作视频包。
  - 违规时恢复正常训诫动作视频包。
- 画面滤镜：
  - 正常。
  - 黑白/旧胶片。
  - 暗色压迫。
- 摄像头：
  - 开启/关闭摄像头巡查。
  - 摄像头预览。
  - 设备选择，预算不足时可只用浏览器默认摄像头。
- 数据维护：
  - 清空今日记录。
  - 清空全部本地记录。

### 4.3 档案页

档案页用于展示用户专注成果和违规记录，是低成本但很撑成品感的核心页面。

展示内容：

- 今日专注时长。
- 今日巡查次数。
- 今日违规次数。
- 今日结算状态。
- 最近记录列表。
- 近 7 日或 14 日专注时长图表。
- 徽章等级进度。
- 角色档案。

角色档案字段：

- 角色姓名。
- 角色称号。
- 角色身份。
- 配音信息。
- 角色简介。
- 代表动作视频。

历史记录字段：

- 日期时间。
- 劳动时长。
- 巡查次数。
- 违规次数。
- 结束原因：光荣下班、主动撤离、禁闭失败。

### 4.4 任务页

任务页展示服务器配置的日常任务。

任务示例：

- 完成一次劳动：奖励 10 货币。
- 连续专注 10 分钟：奖励 5 货币。
- 今日无案底完成劳动：奖励 20 货币。
- 开启摄像头完成一次劳动：奖励 10 货币。
- 查看角色档案：奖励 3 货币。

任务逻辑：

- 每日任务按本地日期刷新。
- 同一任务每日只可领取一次。
- 任务完成后进入“可领取”状态。
- 点击领取后增加本地货币。

### 4.5 结算页

劳动结束后显示结算页。

结算类型：

- 光荣下班：倒计时正常结束，未触发失败条件。
- 主动撤离：用户主动结束。
- 禁闭失败：警告达到上限。

结算内容：

- 本次劳动时长。
- 巡查次数。
- 违规次数。
- 获得货币。
- 徽章进度变化。
- 结算评价动作视频。

### 4.6 隐藏管理员路由

管理员路由不在用户界面展示入口，建议使用以下形式之一：

- `/__admin`
- `/ops`
- `/manage`
- 带密钥参数的入口，例如 `/__admin?token=xxxx`

管理员路由用于配置游戏运行所需的全部动态资源和服务参数。该路由只给开发者、朋友或项目管理员使用，不对终端用户说明。

管理员登录：

- 输入管理员密码或访问 token。
- 登录成功后写入管理员 session。
- 管理员 session 过期时间默认 24 小时。
- 连续错误 5 次后锁定 10 分钟。

管理员首页模块：

- 动作视频包配置。
- 场景配置。
- 角色配置。
- 巡查规则配置。
- 大模型接口配置。
- 数据库配置。
- 运行时预览。
- 配置导入导出。

### 4.7 场景配置 CRUD

场景是动作配置的上一级。用户进入游戏时先选择场景，场景决定背景、巡查气氛、可用动作集合、默认规则和结果映射。

示例 sceneKey：

- `study_room`
- `classroom`
- `office`
- `library`
- `dormitory`
- `interrogation_room`

场景配置字段：

```json
{
  "sceneKey": "study_room",
  "name": "自习室",
  "enabled": true,
  "description": "适合考研、自习、刷题场景。",
  "backgroundType": "image",
  "backgroundUrl": "https://cdn.example.com/scenes/study_room.jpg",
  "backgroundVideoUrl": "",
  "backgroundPosterUrl": "",
  "ambientAudioUrl": "https://cdn.example.com/audio/library_noise.mp3",
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

CRUD 功能：

- 新增场景。
- 编辑场景。
- 删除场景。
- 启用/禁用场景。
- 复制场景。
- 上传或填写静态背景 URL。
- 上传或填写动态背景视频 URL。
- 上传或填写动态背景 poster。
- 配置环境音。
- 配置场景下可用动作。
- 配置模型结果到动作 key 的映射。
- 预览场景。

校验规则：

- `sceneKey` 必填，只允许英文、数字、下划线。
- `sceneKey` 不可重复。
- 启用场景必须至少配置静态背景或动态背景。
- 若 `backgroundType=image`，必须配置 `backgroundUrl`。
- 若 `backgroundType=video`，必须配置 `backgroundVideoUrl`，建议配置 `backgroundPosterUrl`。
- 启用场景必须至少绑定一个启用动作。
- 场景绑定的动作 key 必须存在。

### 4.8 动作配置 CRUD

动作配置挂在场景之下。每一个动作以 `sceneKey + actionKey` 作为唯一组合，动作视频包、判定 JSON、触发条件作为 value。同一个 `actionKey` 可以在不同场景下对应不同动作视频包。

动作视频包是一个完整视频文件，由朋友提供。台词、语音、角色动作、表情、音效、局部字幕等都应封装在该视频文件中，管理员后台不再单独拆分配置台词、语音或素材资源。

示例 actionKey：

- `idle`
- `start_work`
- `patrol_enter`
- `patrol_normal`
- `patrol_suspicious`
- `patrol_violation`
- `patrol_absent`
- `patrol_sleeping`
- `patrol_phone`
- `praise`
- `warning`
- `fail`
- `finish_success`
- `finish_quit`
- `exit`

动作配置字段：

```json
{
  "sceneKey": "study_room",
  "actionKey": "patrol_phone",
  "name": "抓到玩手机",
  "enabled": true,
  "priority": 100,
  "videoUrl": "https://cdn.example.com/actions/patrol_phone.mp4",
  "posterUrl": "https://cdn.example.com/actions/patrol_phone.jpg",
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

CRUD 功能：

- 新增动作。
- 编辑动作。
- 删除动作。
- 启用/禁用动作。
- 复制动作。
- 调整优先级。
- 上传或填写动作视频 URL。
- 上传或填写视频 poster。
- 编辑动作元数据 JSON。
- 配置模型结果到动作的映射。
- 点击预览动作。

校验规则：

- `actionKey` 必填，只允许英文、数字、下划线。
- 同一场景下 `actionKey` 不可重复。
- `sceneKey` 必须存在且启用。
- 启用动作必须配置动作视频 URL。
- `metadata` 必须是合法 JSON。
- `durationMs` 必须大于 0。

### 4.9 大模型接口配置

管理员可配置一个或多个大模型视觉接口。运行时按启用状态和优先级选择。

配置字段：

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

支持功能：

- 新增接口。
- 编辑接口。
- 删除接口。
- 启用/禁用接口。
- 测试连接。
- 上传测试图片并查看返回 JSON。
- 配置 prompt。
- 配置返回 JSON schema。
- 配置超时时间。
- 配置失败重试次数。
- 配置模型结果和动作 key 的映射。

安全要求：

- API Key 不在用户端暴露。
- 用户端只调用服务器判定接口。
- 服务器再调用大模型供应商。
- 管理员页面中 API Key 默认脱敏显示。

### 4.10 数据库配置

数据库使用服务器侧数据库，不使用用户本地 localStorage 作为主存储。

管理员可配置：

- 数据库类型：SQLite、PostgreSQL、MySQL，默认 SQLite。
- 数据库连接地址。
- 用户名。
- 密码。
- 数据库名。
- 表前缀。
- 连接池大小。
- 备份开关。
- 备份路径。

配置原则：

- 若使用 SQLite，适合单客户轻量部署。
- 若使用 PostgreSQL/MySQL，适合需要长期稳定运行和多人访问的部署。
- 数据库密码不得下发给前端。
- 数据库连接测试只在服务器执行。

数据库配置功能：

- 查看当前数据库状态。
- 测试连接。
- 初始化表结构。
- 迁移表结构。
- 导出备份。
- 导入备份。
- 清理测试数据。

## 5. 核心状态机

游戏状态：

- idle：待命。
- working：劳动中。
- paused：暂停。
- patrolWarning：巡查预告。
- patrolActive：巡查中。
- patrolResult：巡查结果展示。
- finished：已完成。
- failed：失败。

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

- 点击开始后进入 working。
- working 状态开始计时，并安排下一次巡查。
- 到达巡查时间后进入 patrolWarning。
- patrolWarning 播放脚步/敲门/门开音效。
- patrolActive 显示角色，执行判定。
- patrolResult 根据判定播放动作视频包并显示结果。
- 结果展示结束后回到 working。
- 警告达到上限进入 failed。
- 倒计时归零进入 finished。
- 用户主动结束进入 finished，类型为主动撤离。

## 6. 巡查逻辑

### 6.1 巡查触发

每次进入 working 后，根据巡查频率随机生成下一次巡查时间。

示例：

```text
慢速：180-480 秒
正常：60-240 秒
高压：30-120 秒
```

同一轮巡查结束后重新生成下一次巡查时间。

### 6.2 巡查演出

巡查分为三个阶段：

- 预告：播放当前场景的预告动作视频或执行屏幕暗化。
- 入场：播放 `patrol_enter` 动作视频包。
- 判定：根据大模型结果播放对应动作视频包。

可配置动作视频包：

- 入场动作视频。
- 正常巡查视频。
- 可疑巡查视频。
- 玩手机视频。
- 打瞌睡视频。
- 离岗视频。
- 表扬视频。
- 失败视频。
- 结算视频。

### 6.3 判定方式

完整版采用“服务器大模型强校验”的方式。

判定来源：

- 摄像头是否开启。
- 用户是否处于暂停。
- 用户是否主动点击“认罪/背叛组织”。
- 摄像头画面是否成功取到帧。
- 管理员配置的大模型视觉接口。
- 管理员配置的错误处理策略。

基础判定类型：

- normal：正常。
- suspicious：可疑。
- violation：违规。
- using_phone：玩手机。
- sleeping：打瞌睡。
- absent：离岗。
- uncertain：不确定。

判定流程：

```text
巡查触发
  -> 读取当前 sceneKey
  -> 前端采集摄像头当前帧
  -> 前端将 sceneKey、sessionId、截图帧发送到服务器 /api/patrol/check
  -> 服务器读取当前启用的大模型配置
  -> 若无可用大模型配置，返回 MODEL_CONFIG_MISSING
  -> 调用大模型视觉接口
  -> 校验大模型返回 JSON
  -> 根据当前场景的 modelResultActionMap 映射 actionKey
  -> 读取 sceneKey + actionKey 对应动作配置
  -> 返回判定结果和动作配置
  -> 前端播放对应动作视频包
  -> 写入巡查记录
```

错误处理策略：

- 摄像头关闭：按管理员配置处理，可放行、可疑或直接违规。
- 截图失败：返回 `uncertain` 或 `absent`，由管理员配置。
- 大模型未配置：直接报错，用户端提示“巡查系统未配置，请联系管理员”。
- 大模型接口超时：直接报错或按管理员配置重试，不使用随机判定替代。
- 大模型返回非法 JSON：记录错误并提示联系管理员，不使用随机判定替代。
- 用户点击“背叛组织”：立即判 `violation`。

用户端错误提示：

```text
巡查系统暂不可用。
原因：大模型接口未配置或不可用。
请联系管理员处理。
```

### 6.3.1 大模型返回 JSON

大模型必须返回 JSON，服务器负责解析和校验。

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

服务器只信任白名单字段：

- `status`
- `confidence`
- `reason`
- `objects`
- `actionKey`

若模型返回 `actionKey`，仍需检查该动作是否存在且启用；否则按 `status` 映射默认动作。

若当前场景没有对应动作配置：

- 服务器返回 `ACTION_CONFIG_MISSING`。
- 用户端提示“当前场景动作配置缺失，请联系管理员”。
- 后台记录缺失的 `sceneKey`、`status`、`actionKey`。

### 6.4 警告与案底

警告规则：

- suspicious：增加 1 次警告，可不计入案底。
- violation：增加 1 次案底，同时增加 1 次警告。
- normal：不增加。

失败规则：

- 警告达到 3 次，进入禁闭失败。
- 案底达到 3 次，也可进入禁闭失败。

可配置项：

- 最大警告数，默认 3。
- 最大案底数，默认 3。
- 可疑是否计入警告，默认是。
- 违规是否直接失败，默认否。

## 7. 计时逻辑

### 7.1 番茄钟模式

- 用户选择时长。
- 点击开始后倒计时。
- 暂停时停止计时。
- 继续时恢复计时。
- 倒计时归零后自动结算。

### 7.2 无限模式

- 点击开始后正计时。
- 用户主动结束时结算。
- 可设置最低有效时长，例如 5 分钟。

### 7.3 有效专注时长

有效专注时长用于徽章、任务、历史统计。

规则：

- paused 状态不计入。
- failed 状态可按比例扣减，例如只记录 50%。
- 主动撤离可扣减固定 5 分钟，或按实际时长记录但不给额外奖励。

## 8. 奖励系统

### 8.1 货币

货币为本地虚拟数值，不接支付。

获取方式：

- 完成劳动。
- 完成任务。
- 无违规奖励。
- 连续天数奖励。

消耗方式：

- 完整版默认可暂不做真实消耗，只记录获得数量。
- 可扩展为兑换背景、装饰、特殊动作视频包。

### 8.2 徽章

徽章根据累计有效专注时长解锁。

示例等级：

- 1 级：初始。
- 2 级：累计 1 小时。
- 3 级：累计 10 小时。
- 4 级：累计 50 小时。
- 5 级：累计 100 小时。

徽章只升不降。

## 9. 数据存储

使用服务器数据库保存。前端可使用 localStorage 缓存非敏感配置，但不作为主存储。

保存内容：

- 用户配置。
- 今日统计。
- 历史记录。
- 货币数量。
- 任务领取状态。
- 徽章进度。
- 动作配置。
- 素材配置。
- 大模型配置。
- 数据库配置。
- 巡查记录。
- 最近一次劳动状态。

每日刷新逻辑：

- 根据服务器日期判断是否新的一天。
- 新的一天重置今日统计和每日任务。
- 历史记录保留。

### 9.1 数据表建议

建议数据表：

- `admin_users`：管理员账号。
- `app_settings`：全局配置。
- `characters`：角色基础信息。
- `scenes`：场景配置。
- `action_configs`：动作 key-value 配置。
- `model_configs`：大模型接口配置。
- `database_configs`：数据库连接配置记录。
- `sessions`：劳动局记录。
- `patrol_records`：巡查记录。
- `daily_stats`：每日统计。
- `tasks`：任务配置。
- `task_records`：任务完成记录。
- `badges`：徽章配置。
- `user_progress`：用户进度。

### 9.2 scenes 表字段

核心字段：

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

### 9.3 action_configs 表字段

核心字段：

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

- `scene_key + action_key` 唯一。

### 9.4 model_configs 表字段

核心字段：

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
- `created_at`
- `updated_at`

### 9.5 服务器 API

用户端 API：

- `GET /api/runtime/config`：获取用户端运行配置。
- `GET /api/scenes`：获取可用场景列表。
- `POST /api/session/start`：开始劳动。
- `POST /api/session/pause`：暂停劳动。
- `POST /api/session/resume`：继续劳动。
- `POST /api/session/finish`：结束劳动。
- `POST /api/patrol/check`：提交巡查帧并获取判定结果。
- `GET /api/profile/stats`：获取档案统计。
- `GET /api/tasks/today`：获取今日任务。
- `POST /api/tasks/:id/claim`：领取任务奖励。

管理员 API：

- `POST /api/admin/login`
- `GET /api/admin/scenes`
- `POST /api/admin/scenes`
- `PUT /api/admin/scenes/:id`
- `DELETE /api/admin/scenes/:id`
- `POST /api/admin/scenes/:id/preview`
- `GET /api/admin/actions`
- `POST /api/admin/actions`
- `PUT /api/admin/actions/:id`
- `DELETE /api/admin/actions/:id`
- `POST /api/admin/actions/:id/preview`
- `GET /api/admin/model-configs`
- `POST /api/admin/model-configs`
- `PUT /api/admin/model-configs/:id`
- `DELETE /api/admin/model-configs/:id`
- `POST /api/admin/model-configs/:id/test`
- `GET /api/admin/database`
- `PUT /api/admin/database`
- `POST /api/admin/database/test`
- `POST /api/admin/database/migrate`
- `POST /api/admin/export`
- `POST /api/admin/import`

## 10. 素材规格

### 10.1 场景背景

背景：

- 推荐 1920x1080 或 2560x1440。
- 静态背景：JPG/PNG/WebP。
- 动态背景：MP4/WebM，建议提供 poster 图片。

### 10.2 动作视频包

动作视频包由朋友提供，是“台词 + 语音 + 角色动作 + 表情 + 音效 + 字幕”的一体化视频文件。

推荐格式：

- MP4/H.264，优先。
- WebM，备选。
- 建议 1080p。
- 建议每个视频 3-10 秒。
- 如需透明背景，另行约定 WebM alpha 或绿幕/纯色抠像方案。

建议动作视频包：

- `idle`：待命。
- `start_work`：开始劳动。
- `patrol_enter`：巡查入场。
- `patrol_normal`：正常巡查。
- `patrol_suspicious`：可疑。
- `patrol_phone`：玩手机。
- `patrol_sleeping`：打瞌睡。
- `patrol_absent`：离岗。
- `praise`：表扬。
- `warning`：警告。
- `fail`：禁闭/失败。
- `finish_success`：光荣下班。
- `finish_quit`：主动撤离。
- `exit`：退场。

### 10.3 环境音

推荐格式：

- MP3/OGG/WAV。

环境音类型：

- 背景音乐。
- 白噪。
- 场景环境声。

角色台词、配音、角色动作音效不在后台单独配置，默认封装在动作视频包中。

## 11. 可配置结构

配置由隐藏管理员路由维护，保存到服务器数据库。运行时前端通过 `/api/runtime/config` 获取已启用配置。

示例结构：

```json
{
  "character": {
    "name": "角色名",
    "title": "角色称号",
    "bio": "角色简介"
  },
  "scenes": {
    "study_room": {
      "name": "自习室",
      "backgroundType": "image",
      "backgroundUrl": "assets/scenes/study_room.jpg",
      "backgroundVideoUrl": "",
      "ambientAudioUrl": "assets/audio/library_noise.mp3",
      "modelResultActionMap": {
        "normal": "patrol_normal",
        "using_phone": "patrol_phone",
        "sleeping": "patrol_sleeping",
        "absent": "patrol_absent",
        "uncertain": "patrol_suspicious"
      },
      "actions": {
        "patrol_enter": {
          "name": "巡查入场",
          "videoUrl": "assets/actions/study_room/patrol_enter.mp4",
          "durationMs": 5000,
          "nextActionKey": "patrol_normal",
          "metadata": {
            "screenFilter": "dark"
          }
        },
        "patrol_phone": {
          "name": "抓到玩手机",
          "videoUrl": "assets/actions/study_room/patrol_phone.mp4",
          "durationMs": 6500,
          "modelResultMap": ["using_phone"],
          "metadata": {
            "mood": "angry"
          }
        }
      }
    }
  },
  "model": {
    "enabled": true,
    "provider": "openai-compatible",
    "baseUrl": "https://api.example.com/v1",
    "model": "vision-model-name"
  }
}
```

配置读取优先级：

```text
服务器数据库配置
  -> 管理员导入配置
  -> 默认内置配置
```

前端不得保存以下敏感配置：

- 大模型 API Key。
- 数据库密码。
- 管理员密码。
- 服务端内部 token。

## 12. 交互细节

### 12.1 开始劳动

点击开始：

- 检查配置。
- 初始化本轮数据。
- 播放开始动作视频包或开始音效。
- 进入 working。
- 安排第一次巡查。

### 12.2 暂停

点击暂停：

- 进入 paused。
- 停止计时。
- 暂停巡查倒计时。
- 背景音降低或暂停。

### 12.3 继续

点击继续：

- 返回 working。
- 恢复计时。
- 恢复巡查。

### 12.4 主动结束

点击结束：

- 弹确认。
- 确认后进入结算页。
- 结算类型为主动撤离。

### 12.5 违规按钮

可放一个风格化按钮，例如“认罪”“背叛组织”“我分心了”。

点击后：

- 立即触发一次违规。
- 增加案底和警告。
- 播放训诫动作视频包。

这个按钮可以替代一部分真实识别需求，让用户主动参与游戏。

## 13. 视觉风格要求

- 第一屏必须是可用游戏界面，不做营销页。
- HUD 要有仪表盘感。
- 文案要有角色口吻。
- 页面可采用档案、表格、印章、编号、徽章等视觉元素。
- 按钮状态要明显：可用、禁用、危险、已选中。
- 巡查时应有明显气氛变化：暗化、震动、门声、角色出现。

## 14. 验收标准

### 14.1 功能验收

- 可以正常开始、暂停、继续、结束劳动。
- 用户开始劳动前可以选择场景。
- 场景切换后能正确更换静态/动态背景和环境音。
- 倒计时准确。
- 巡查能按随机间隔触发。
- 巡查能显示角色、播放音效、给出结果。
- 巡查能播放管理员配置的动作视频。
- 管理员能新增、编辑、删除、启用、禁用动作配置。
- 管理员能新增、编辑、删除、启用、禁用场景配置。
- 管理员能配置场景静态图片背景和动态视频背景。
- 管理员能配置场景下可用动作集合。
- 管理员能配置动作 key 与动作视频包、判定 JSON 的映射。
- 管理员能配置场景下模型结果到动作 key 的映射。
- 管理员能配置大模型接口并测试连接。
- 管理员能上传测试图片并查看大模型返回 JSON。
- 未配置可用大模型时，巡查不得随机放行，应提示联系管理员。
- 场景缺失对应动作配置时，应提示联系管理员并记录后台错误。
- 管理员能配置数据库并测试连接。
- 服务器能保存劳动局、巡查记录、档案统计和任务记录。
- 警告/案底能正确增加。
- 达到失败条件后能进入失败结算。
- 正常完成后能进入成功结算。
- 配置项刷新页面后仍保留。
- 档案页能显示今日和历史记录。
- 任务能完成和领取奖励。
- 徽章进度能随有效时长增长。

### 14.2 素材验收

- 背景图显示正常。
- 角色立绘无明显拉伸或遮挡。
- 动作视频包能正常播放。
- 视频中的台词、语音、角色动作与对应事件匹配。

### 14.3 兼容验收

- Chrome 桌面端可用。
- Edge 桌面端可用，视预算。
- Safari/iPhone 不作为默认强验收项，除非另行约定。

## 15. 报价边界建议

本 PRD 对应完整交付版。

完整版包含：

- 用户端网页游戏。
- 服务器 API。
- 服务器数据库。
- 隐藏管理员路由。
- 场景 key-value 配置 CRUD。
- 静态/动态场景背景配置。
- 动作 key-value 配置 CRUD。
- 动作视频包/判定 JSON 配置。
- 大模型接口配置和测试。
- 数据库配置和测试。
- 基础摄像头授权/预览。
- 随机巡查。
- 素材接入。
- 一套角色。
- 一套背景。
- 一套动作视频包接入。

仍需单独约定或额外收费项：

- 多个客户/多租户隔离。
- 真正商业级账号体系。
- YOLO/MediaPipe 本地检测模型训练。
- TTS 接入或音色克隆。
- 排行榜/战友/邀请系统。
- 在线人数统计。
- 商城支付。
- 多角色切换。
- 多套主题。
- 移动端深度适配。
- 超出约定轮数的改稿。

## 16. 推荐交付话术

本版本是角色定制专注游戏完整交付版，主打沉浸式巡查、专注计时、档案记录、角色动作视频和可配置大模型判定。用户侧保持简单，管理员侧通过隐藏路由维护动作、素材、大模型和数据库配置。识别效果取决于配置的大模型能力、摄像头角度、光线和用户环境，不承诺 100% 准确。

## 17. 对标游戏截图留存

以下截图用于内部需求分析和功能拆解，不作为美术照抄依据。实际交付应替换为客户自有角色、背景、UI 风格与文案。

### 17.1 主工作台 / HUD

截图文件：

![主工作台 HUD](/Users/lvfushun/Documents/Codex/2026-07-04/ni-zhi-d/outputs/benchmark-redwatch-01-main-hud.png)

观察点：

- 第一屏就是游戏主界面，不是介绍页。
- 背景图占满全屏，右侧悬浮督学台承担主要操作。
- 督学台包含货币、今日案底、徽章、回合编号、入口按钮、警告条、计时器、在线人数、继续/结束按钮。
- 本项目为个人定制版，不实现真实在线人数、榜单、战友、征召等社区功能。若需要“在线人数”氛围文案，只能作为固定装饰文案或服务器配置数值，不暗示真实社交系统。
- 左上角显示站点名和用户署名，增强“系统在线”的感觉。
- 底部有 PWA 安装提示，强化“装到桌面像应用”的体验。

### 17.2 配置页

截图文件：

![配置页](/Users/lvfushun/Documents/Codex/2026-07-04/ni-zhi-d/outputs/benchmark-redwatch-02-settings.png)

观察点：

- 配置页被包装成“表格/条例/档案”风格，不像普通设置面板。
- 低成本配置项很多，但大部分都是前端状态：
  - 劳动模式。
  - 巡查频率。
  - 白噪。
  - 音量。
  - 轻声巡查。
  - 画面滤镜。
  - 摄像头开关。
  - 摄像头设备选择。
  - 重置。
  - 试机/校准。
  - 安装到桌面。
- 配置页是“撑产品感”的关键模块，用户侧保留常用配置，管理员侧提供完整配置。

### 17.3 档案页

截图文件：

![档案页](/Users/lvfushun/Documents/Codex/2026-07-04/ni-zhi-d/outputs/benchmark-redwatch-03-stats.png)

观察点：

- 档案页负责长期留存和成长感。
- 展示徽章等级、下一等级进度、今日工作台、最近归档、专注时长、巡查次数、违纪次数、近十四日图表。
- 还有督学官角色档案，包括姓名、称号、职务、配音、背景故事和代表动作。
- 这部分不依赖 AI，完整版使用服务器数据库保存。

### 17.4 任务弹窗

截图文件：

![任务弹窗](/Users/lvfushun/Documents/Codex/2026-07-04/ni-zhi-d/outputs/benchmark-redwatch-04-tasks-modal.png)

观察点：

- 任务是叠加弹窗，不打断主界面。
- 功能定位是每日补给、新兵探索、赚取货币/券类道具。
- 完整版任务由服务器配置，用户侧只负责展示和领取。
- 任务页的价值是制造“每天回来做一下”的理由。
- 对标游戏中的社交、征召、榜单类入口不纳入个人定制版。

## 18. 对标游戏主线业务逻辑

对标游戏的主线不是复杂战斗或关卡，而是一个“专注计时 + 随机巡查 + 违规惩罚 + 成长记录”的循环。个人定制版保留这个核心循环和管理配置能力，不做社区互动线。

### 18.1 用户主流程

```text
进入页面
  -> 选择/确认配置
  -> 点击开始劳动
  -> 进入专注计时
  -> 系统随机安排巡查
  -> 巡查演出触发
  -> 摄像头/规则/AI 进行判定
  -> 输出正常、可疑、违规、离岗等结果
  -> 更新警告、案底、货币、任务进度
  -> 回到专注计时
  -> 倒计时结束或用户主动结束
  -> 结算
  -> 写入档案
  -> 推进徽章/任务/历史统计
```

### 18.2 核心闭环

核心闭环由四个系统组成：

```text
计时系统
  提供基础目标：坚持到下班。

巡查系统
  提供不确定性：不知道什么时候会被查。

惩罚系统
  提供压力：警告、案底、失败。

成长系统
  提供长期动机：货币、任务、徽章、档案。
```

玩家的心理体验来自“不确定巡查”和“角色反馈”，而不完全来自识别技术。完整版通过服务器配置动作和大模型，让演出、音效、记录和识别形成完整闭环。

### 18.3 主要业务对象

```text
Session / 劳动局
  id
  startTime
  endTime
  mode
  plannedDuration
  actualFocusSeconds
  patrolCount
  warningCount
  violationCount
  result

Patrol / 巡查
  id
  sessionId
  triggerTime
  result
  reason
  lineId
  snapshotEnabled

Profile / 档案
  totalFocusSeconds
  badgeLevel
  currency
  dailyStats
  history

Task / 任务
  id
  date
  condition
  status
  reward

Config / 配置
  workMode
  patrolFrequency
  volumes
  noiseType
  cameraEnabled
  visualFilter
```

### 18.4 主状态流

```text
idle 待命
  用户还没开始劳动。

working 劳动中
  计时器运行，巡查倒计时运行。

patrolWarning 巡查预告
  播放脚步、敲门、暗化、震动等气氛演出。

patrolActive 巡查中
  角色出现，摄像头取帧并提交服务器大模型判定。

patrolResult 巡查结果
  展示正常/可疑/违规，播放对应动作视频包，更新数据。

paused 暂停
  停止计时和巡查。

finished 结算
  写入历史，发放奖励，推进徽章。

failed 禁闭/失败
  警告或案底达到上限，强制结算。
```

### 18.5 巡查判定分层

对标游戏可以理解为三层能力：

```text
第一层：演出层
  门开、脚步、角色入场、台词、配音、音效都封装在动作视频包中。
  这是最能被用户感知的部分。

第二层：配置层
  场景、动作、结果映射、错误提示、管理员配置。
  完整版作为后台能力实现。

第三层：智能层
  判断是否玩手机、离岗、打瞌睡。
  通过管理员配置的大模型接口实现。
```

因此完整版应实现第一层和第二层，并提供第三层的大模型配置能力。

### 18.6 个人定制版应删除的对标功能

以下功能属于社区/运营能力，不适合个人定制版：

- 荣誉榜/排行榜。
- 战友系统。
- 征召/邀请。
- 真实在线人数。
- 登录账号。
- 云端档案。
- 跨设备同步。
- 商城支付。
- 服务端任务发放。

这些功能会引入社交关系、运营规则和售后成本，和本项目“个人定制督学工具”的定位冲突。

### 18.7 为什么功能看起来多但实现可控

对标游戏很多功能是“数值和文案包装”：

- 军功币：服务器保存的数字。
- 今日案底：服务器保存的计数。
- 徽章：累计时长阈值。
- 任务：服务器配置，本地触发条件上报。
- 档案：历史记录列表。
- 白噪和音量：音频播放控制。
- 滤镜：CSS filter。
- 全屏：浏览器 Fullscreen API。
- PWA 安装提示：manifest + beforeinstallprompt。

真正高成本的是：

- 真实行为识别。
- 多端兼容。
- 云端同步。
- 社交关系。
- 排行榜。
- 支付商城。
- 后台配置。

完整版应把显眼的包装做足，同时用隐藏管理员路由解决后续换动作、换视频、换模型、换数据库的问题。
