<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

type AnyRecord = Record<string, any>

const isAdmin = window.location.pathname === '/__admin'
const adminKey = ref(new URLSearchParams(window.location.search).get('appkey') || localStorage.getItem('adminAppKey') || '')
const activeTab = ref('status')
const loading = ref(false)
const message = ref('')
const error = ref('')

const tabs = [
  ['status', '运行状态'],
  ['runtime', '运行时预览'],
  ['characters', '角色'],
  ['scenes', '场景'],
  ['actions', '动作'],
  ['model', '大模型'],
  ['patrol', '巡查规则'],
  ['mysql', 'MySQL'],
]

const state = reactive<AnyRecord>({
  status: null,
  runtime: null,
  characters: [],
  scenes: [],
  actions: [],
  model: null,
  patrol: null,
  mysql: null,
  selectedCharacter: null,
  selectedScene: null,
  selectedAction: null,
  modelInput: {},
  mysqlInput: {},
})

const adminHeaders = computed(() => ({
  'Content-Type': 'application/json',
  'X-App-Key': adminKey.value,
}))

onMounted(() => {
  if (isAdmin && adminKey.value) {
    localStorage.setItem('adminAppKey', adminKey.value)
    void loadAll()
  }
})

async function api(path: string, options: RequestInit = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      ...adminHeaders.value,
      ...(options.headers || {}),
    },
  })
  const data = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(data?.error?.message || data?.error?.code || '请求失败')
  }
  return data
}

async function run(task: () => Promise<void>, ok = '已保存') {
  loading.value = true
  error.value = ''
  message.value = ''
  try {
    await task()
    message.value = ok
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

async function loadAll() {
  await run(async () => {
    await Promise.all([
      loadStatus(),
      loadRuntime(),
      loadCharacters(),
      loadScenes(),
      loadActions(),
      loadModel(),
      loadPatrol(),
      loadMySQL(),
    ])
  }, '已加载')
}

async function loadStatus() {
  state.status = await api('/api/admin/status')
}

async function loadRuntime() {
  state.runtime = await api('/api/admin/runtime-config')
}

async function loadCharacters() {
  const data = await api('/api/admin/characters')
  state.characters = data.items || []
  if (!state.selectedCharacter) state.selectedCharacter = emptyCharacter()
}

async function loadScenes() {
  const data = await api('/api/admin/scenes')
  state.scenes = data.items || []
  if (!state.selectedScene) state.selectedScene = emptyScene()
}

async function loadActions() {
  const data = await api('/api/admin/actions')
  state.actions = data.items || []
  if (!state.selectedAction) state.selectedAction = emptyAction()
}

async function loadModel() {
  state.model = await api('/api/admin/model-config')
  state.modelInput = { ...state.model, apiKey: '', enabled: Boolean(state.model.enabled) }
}

async function loadPatrol() {
  state.patrol = await api('/api/admin/patrol-rule')
}

async function loadMySQL() {
  state.mysql = await api('/api/admin/mysql-config')
  state.mysqlInput = { ...state.mysql, password: '', enabled: Boolean(state.mysql.enabled) }
}

function activateAdmin() {
  localStorage.setItem('adminAppKey', adminKey.value)
  void loadAll()
}

function selectCharacter(item: AnyRecord) {
  state.selectedCharacter = { ...item }
}

function newCharacter(base?: AnyRecord) {
  state.selectedCharacter = base ? { ...base, id: 0, characterKey: `${base.characterKey}_copy` } : emptyCharacter()
}

function selectScene(item: AnyRecord) {
  state.selectedScene = { ...item }
}

function newScene(base?: AnyRecord) {
  state.selectedScene = base ? { ...base, id: 0, sceneKey: `${base.sceneKey}_copy`, enabled: false } : emptyScene()
}

function selectAction(item: AnyRecord) {
  state.selectedAction = { ...item }
}

function newAction(base?: AnyRecord) {
  state.selectedAction = base ? { ...base, id: 0, actionKey: `${base.actionKey}_copy`, enabled: false } : emptyAction()
}

async function saveCharacter() {
  await run(async () => {
    ensureJSON(state.selectedCharacter.profileJson, '角色档案 JSON')
    ensureJSON(state.selectedCharacter.metadataJson, 'metadata JSON')
    const body = JSON.stringify(state.selectedCharacter)
    if (state.selectedCharacter.id) {
      await api(`/api/admin/characters/${state.selectedCharacter.id}`, { method: 'PUT', body })
    } else {
      await api('/api/admin/characters', { method: 'POST', body })
    }
    await loadCharacters()
    await loadRuntime()
  })
}

async function deleteCharacter(id: number) {
  await run(async () => {
    await api(`/api/admin/characters/${id}`, { method: 'DELETE' })
    state.selectedCharacter = emptyCharacter()
    await loadCharacters()
  }, '已删除')
}

async function saveScene() {
  await run(async () => {
    ensureJSON(state.selectedScene.availableActionKeysJson, '可用动作 JSON')
    ensureJSON(state.selectedScene.modelResultActionMapJson, '结果映射 JSON')
    ensureJSON(state.selectedScene.metadataJson, 'metadata JSON')
    const body = JSON.stringify(state.selectedScene)
    if (state.selectedScene.id) {
      await api(`/api/admin/scenes/${state.selectedScene.id}`, { method: 'PUT', body })
    } else {
      await api('/api/admin/scenes', { method: 'POST', body })
    }
    await loadScenes()
    await loadRuntime()
  })
}

async function deleteScene(id: number) {
  await run(async () => {
    await api(`/api/admin/scenes/${id}`, { method: 'DELETE' })
    state.selectedScene = emptyScene()
    await loadScenes()
  }, '已删除')
}

async function saveAction() {
  await run(async () => {
    ensureJSON(state.selectedAction.modelResultMapJson, '模型映射 JSON')
    ensureJSON(state.selectedAction.localRuleMapJson, '本地规则 JSON')
    ensureJSON(state.selectedAction.metadataJson, 'metadata JSON')
    const body = JSON.stringify(state.selectedAction)
    if (state.selectedAction.id) {
      await api(`/api/admin/actions/${state.selectedAction.id}`, { method: 'PUT', body })
    } else {
      await api('/api/admin/actions', { method: 'POST', body })
    }
    await loadActions()
    await loadScenes()
  })
}

async function deleteAction(id: number) {
  await run(async () => {
    await api(`/api/admin/actions/${id}`, { method: 'DELETE' })
    state.selectedAction = emptyAction()
    await loadActions()
  }, '已删除')
}

async function saveModel() {
  await run(async () => {
    ensureJSON(state.modelInput.responseSchemaJson, '响应 schema JSON')
    await api('/api/admin/model-config', {
      method: 'PUT',
      body: JSON.stringify(state.modelInput),
    })
    await loadModel()
  })
}

async function testModel() {
  await run(async () => {
    const data = await api('/api/admin/model-config/test', { method: 'POST', body: '{}' })
    message.value = data.message || data.status
  }, '')
}

async function savePatrol() {
  await run(async () => {
    await api('/api/admin/patrol-rule', {
      method: 'PUT',
      body: JSON.stringify(state.patrol),
    })
    await loadPatrol()
    await loadRuntime()
  })
}

async function saveMySQL() {
  await run(async () => {
    await api('/api/admin/mysql-config', {
      method: 'PUT',
      body: JSON.stringify(state.mysqlInput),
    })
    await loadMySQL()
    await loadStatus()
  })
}

async function testMySQL() {
  await run(async () => {
    await api('/api/admin/mysql-config/test', {
      method: 'POST',
      body: JSON.stringify(state.mysqlInput),
    })
    await loadMySQL()
  }, '测试完成')
}

async function migrateMySQL() {
  await run(async () => {
    await api('/api/admin/mysql-config/migrate', { method: 'POST', body: '{}' })
  }, '迁移完成')
}

function ensureJSON(value: string, label: string) {
  JSON.parse(value || '{}')
  return label
}

function emptyCharacter() {
  return {
    id: 0,
    characterKey: '',
    name: '',
    enabled: false,
    description: '',
    avatarUrl: '',
    profileJson: '{}',
    voiceStyle: '',
    defaultSceneKey: '',
    metadataJson: '{}',
  }
}

function emptyScene() {
  return {
    id: 0,
    sceneKey: '',
    name: '',
    enabled: false,
    description: '',
    backgroundType: 'image',
    backgroundUrl: '',
    backgroundVideoUrl: '',
    backgroundPosterUrl: '',
    ambientAudioUrl: '',
    defaultActionKey: '',
    availableActionKeysJson: '[]',
    modelResultActionMapJson: '{}',
    metadataJson: '{}',
  }
}

function emptyAction() {
  return {
    id: 0,
    sceneKey: 'study_room',
    actionKey: '',
    name: '',
    enabled: false,
    priority: 0,
    videoUrl: '',
    posterUrl: '',
    durationMs: 8000,
    nextActionKey: '',
    modelResultMapJson: '{}',
    localRuleMapJson: '{}',
    metadataJson: '{}',
  }
}
</script>

<template>
  <main v-if="!isAdmin" class="app-shell">
    <section class="stage">
      <div class="scene-label">Default Scene</div>
      <div class="timer">25:00</div>
      <p class="status">待命中，等待开始劳动</p>
      <div class="actions">
        <button type="button">开始劳动</button>
        <button type="button" class="secondary">配置</button>
      </div>
    </section>

    <aside class="hud">
      <h1>Supervisor Game</h1>
      <dl>
        <div><dt>警告</dt><dd>0/3</dd></div>
        <div><dt>今日案底</dt><dd>0</dd></div>
        <div><dt>徽章等级</dt><dd>Lv.1</dd></div>
        <div><dt>货币</dt><dd>0</dd></div>
      </dl>
    </aside>
  </main>

  <main v-else class="admin-shell">
    <header class="admin-topbar">
      <div>
        <p class="eyebrow">Hidden Admin</p>
        <h1>管理端配置闭环</h1>
      </div>
      <div class="appkey-box">
        <input v-model="adminKey" type="password" placeholder="APP_KEY" />
        <button type="button" @click="activateAdmin">进入</button>
      </div>
    </header>

    <nav class="admin-tabs" aria-label="管理端标签">
      <button
        v-for="[key, label] in tabs"
        :key="key"
        type="button"
        :class="{ active: activeTab === key }"
        @click="activeTab = key"
      >
        {{ label }}
      </button>
    </nav>

    <p v-if="loading" class="notice">处理中...</p>
    <p v-if="message" class="notice success">{{ message }}</p>
    <p v-if="error" class="notice danger">{{ error }}</p>

    <section v-if="activeTab === 'status'" class="admin-panel">
      <div class="panel-title">
        <h2>运行状态</h2>
        <button type="button" @click="loadStatus">刷新</button>
      </div>
      <pre>{{ JSON.stringify(state.status, null, 2) }}</pre>
    </section>

    <section v-if="activeTab === 'runtime'" class="admin-panel">
      <div class="panel-title">
        <h2>运行时配置预览</h2>
        <button type="button" @click="loadRuntime">刷新</button>
      </div>
      <pre>{{ JSON.stringify(state.runtime, null, 2) }}</pre>
    </section>

    <section v-if="activeTab === 'characters'" class="admin-panel split-panel">
      <div>
        <div class="panel-title">
          <h2>角色配置</h2>
          <button type="button" @click="newCharacter()">新增</button>
        </div>
        <table>
          <thead><tr><th>Key</th><th>名称</th><th>启用</th><th></th></tr></thead>
          <tbody>
            <tr v-for="item in state.characters" :key="item.id">
              <td>{{ item.characterKey }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.enabled ? '是' : '否' }}</td>
              <td class="row-actions">
                <button type="button" @click="selectCharacter(item)">编辑</button>
                <button type="button" @click="newCharacter(item)">复制</button>
                <button type="button" class="danger-btn" @click="deleteCharacter(item.id)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <form class="edit-form" @submit.prevent="saveCharacter">
        <label>characterKey<input v-model="state.selectedCharacter.characterKey" /></label>
        <label>名称<input v-model="state.selectedCharacter.name" /></label>
        <label class="check"><input v-model="state.selectedCharacter.enabled" type="checkbox" /> 启用</label>
        <label>头像路径<input v-model="state.selectedCharacter.avatarUrl" /></label>
        <label>默认场景<input v-model="state.selectedCharacter.defaultSceneKey" /></label>
        <label>语音风格<input v-model="state.selectedCharacter.voiceStyle" /></label>
        <label>描述<textarea v-model="state.selectedCharacter.description"></textarea></label>
        <label>角色档案 JSON<textarea v-model="state.selectedCharacter.profileJson"></textarea></label>
        <label>metadata JSON<textarea v-model="state.selectedCharacter.metadataJson"></textarea></label>
        <button type="submit">保存角色</button>
      </form>
    </section>

    <section v-if="activeTab === 'scenes'" class="admin-panel split-panel">
      <div>
        <div class="panel-title">
          <h2>场景配置</h2>
          <button type="button" @click="newScene()">新增</button>
        </div>
        <table>
          <thead><tr><th>Key</th><th>名称</th><th>启用</th><th></th></tr></thead>
          <tbody>
            <tr v-for="item in state.scenes" :key="item.id">
              <td>{{ item.sceneKey }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.enabled ? '是' : '否' }}</td>
              <td class="row-actions">
                <button type="button" @click="selectScene(item)">编辑</button>
                <button type="button" @click="newScene(item)">复制</button>
                <button type="button" class="danger-btn" @click="deleteScene(item.id)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <form class="edit-form" @submit.prevent="saveScene">
        <label>sceneKey<input v-model="state.selectedScene.sceneKey" /></label>
        <label>名称<input v-model="state.selectedScene.name" /></label>
        <label class="check"><input v-model="state.selectedScene.enabled" type="checkbox" /> 启用</label>
        <label>背景类型<input v-model="state.selectedScene.backgroundType" /></label>
        <label>背景图片<input v-model="state.selectedScene.backgroundUrl" /></label>
        <label>背景视频<input v-model="state.selectedScene.backgroundVideoUrl" /></label>
        <label>poster<input v-model="state.selectedScene.backgroundPosterUrl" /></label>
        <label>环境音<input v-model="state.selectedScene.ambientAudioUrl" /></label>
        <label>默认动作<input v-model="state.selectedScene.defaultActionKey" /></label>
        <label>描述<textarea v-model="state.selectedScene.description"></textarea></label>
        <label>可用动作 JSON<textarea v-model="state.selectedScene.availableActionKeysJson"></textarea></label>
        <label>结果映射 JSON<textarea v-model="state.selectedScene.modelResultActionMapJson"></textarea></label>
        <label>metadata JSON<textarea v-model="state.selectedScene.metadataJson"></textarea></label>
        <button type="submit">保存场景</button>
      </form>
    </section>

    <section v-if="activeTab === 'actions'" class="admin-panel split-panel">
      <div>
        <div class="panel-title">
          <h2>动作配置</h2>
          <button type="button" @click="newAction()">新增</button>
        </div>
        <table>
          <thead><tr><th>场景</th><th>Key</th><th>名称</th><th>启用</th><th></th></tr></thead>
          <tbody>
            <tr v-for="item in state.actions" :key="item.id">
              <td>{{ item.sceneKey }}</td>
              <td>{{ item.actionKey }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.enabled ? '是' : '否' }}</td>
              <td class="row-actions">
                <button type="button" @click="selectAction(item)">编辑</button>
                <button type="button" @click="newAction(item)">复制</button>
                <button type="button" class="danger-btn" @click="deleteAction(item.id)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <form class="edit-form" @submit.prevent="saveAction">
        <label>sceneKey<input v-model="state.selectedAction.sceneKey" /></label>
        <label>actionKey<input v-model="state.selectedAction.actionKey" /></label>
        <label>名称<input v-model="state.selectedAction.name" /></label>
        <label class="check"><input v-model="state.selectedAction.enabled" type="checkbox" /> 启用</label>
        <label>优先级<input v-model.number="state.selectedAction.priority" type="number" /></label>
        <label>视频路径<input v-model="state.selectedAction.videoUrl" /></label>
        <label>poster<input v-model="state.selectedAction.posterUrl" /></label>
        <label>时长 ms<input v-model.number="state.selectedAction.durationMs" type="number" /></label>
        <label>nextActionKey<input v-model="state.selectedAction.nextActionKey" /></label>
        <label>模型映射 JSON<textarea v-model="state.selectedAction.modelResultMapJson"></textarea></label>
        <label>本地规则 JSON<textarea v-model="state.selectedAction.localRuleMapJson"></textarea></label>
        <label>metadata JSON<textarea v-model="state.selectedAction.metadataJson"></textarea></label>
        <video v-if="state.selectedAction.videoUrl" :src="'/' + state.selectedAction.videoUrl" controls></video>
        <button type="submit">保存动作</button>
      </form>
    </section>

    <section v-if="activeTab === 'model'" class="admin-panel">
      <div class="panel-title">
        <h2>大模型配置</h2>
        <button type="button" @click="testModel">测试</button>
      </div>
      <form class="edit-form wide-form" @submit.prevent="saveModel">
        <label>名称<input v-model="state.modelInput.name" /></label>
        <label>provider<input v-model="state.modelInput.provider" /></label>
        <label class="check"><input v-model="state.modelInput.enabled" type="checkbox" /> 启用</label>
        <label>baseUrl<input v-model="state.modelInput.baseUrl" /></label>
        <label>apiKey<input v-model="state.modelInput.apiKey" type="password" :placeholder="state.modelInput.apiKeyMasked || '留空保留旧值'" /></label>
        <label>model<input v-model="state.modelInput.model" /></label>
        <label>timeoutMs<input v-model.number="state.modelInput.timeoutMs" type="number" /></label>
        <label>maxImageWidth<input v-model.number="state.modelInput.maxImageWidth" type="number" /></label>
        <label>temperature<input v-model.number="state.modelInput.temperature" type="number" step="0.1" /></label>
        <label>retryCount<input v-model.number="state.modelInput.retryCount" type="number" /></label>
        <label>prompt<textarea v-model="state.modelInput.prompt"></textarea></label>
        <label>responseSchema JSON<textarea v-model="state.modelInput.responseSchemaJson"></textarea></label>
        <button type="submit">保存模型配置</button>
      </form>
    </section>

    <section v-if="activeTab === 'patrol'" class="admin-panel">
      <div class="panel-title"><h2>巡查规则</h2></div>
      <form v-if="state.patrol" class="edit-form wide-form" @submit.prevent="savePatrol">
        <label>慢速最小秒<input v-model.number="state.patrol.slowMinSeconds" type="number" /></label>
        <label>慢速最大秒<input v-model.number="state.patrol.slowMaxSeconds" type="number" /></label>
        <label>正常最小秒<input v-model.number="state.patrol.normalMinSeconds" type="number" /></label>
        <label>正常最大秒<input v-model.number="state.patrol.normalMaxSeconds" type="number" /></label>
        <label>高压最小秒<input v-model.number="state.patrol.highMinSeconds" type="number" /></label>
        <label>高压最大秒<input v-model.number="state.patrol.highMaxSeconds" type="number" /></label>
        <label>最大警告<input v-model.number="state.patrol.maxWarnings" type="number" /></label>
        <label>最大案底<input v-model.number="state.patrol.maxViolations" type="number" /></label>
        <label class="check"><input v-model="state.patrol.suspiciousAddsWarning" type="checkbox" /> 可疑计入警告</label>
        <label class="check"><input v-model="state.patrol.violationDirectFail" type="checkbox" /> 违规直接失败</label>
        <label>摄像头关闭策略<input v-model="state.patrol.cameraOffStrategy" /></label>
        <label>截图失败策略<input v-model="state.patrol.captureFailedStrategy" /></label>
        <label>超时重试<input v-model.number="state.patrol.modelTimeoutRetryCount" type="number" /></label>
        <label>用户错误提示<textarea v-model="state.patrol.userErrorMessage"></textarea></label>
        <button type="submit">保存巡查规则</button>
      </form>
    </section>

    <section v-if="activeTab === 'mysql'" class="admin-panel">
      <div class="panel-title">
        <h2>MySQL 连接配置</h2>
        <div class="row-actions">
          <button type="button" @click="testMySQL">测试连接</button>
          <button type="button" @click="migrateMySQL">执行迁移</button>
        </div>
      </div>
      <form class="edit-form wide-form" @submit.prevent="saveMySQL">
        <label>host<input v-model="state.mysqlInput.host" /></label>
        <label>port<input v-model.number="state.mysqlInput.port" type="number" /></label>
        <label>databaseName<input v-model="state.mysqlInput.databaseName" /></label>
        <label>username<input v-model="state.mysqlInput.username" /></label>
        <label>password<input v-model="state.mysqlInput.password" type="password" :placeholder="state.mysqlInput.passwordMasked || '留空保留旧值'" /></label>
        <label>charset<input v-model="state.mysqlInput.charset" /></label>
        <label>timezone<input v-model="state.mysqlInput.timezone" /></label>
        <label>maxOpenConns<input v-model.number="state.mysqlInput.maxOpenConns" type="number" /></label>
        <label>maxIdleConns<input v-model.number="state.mysqlInput.maxIdleConns" type="number" /></label>
        <label class="check"><input v-model="state.mysqlInput.enabled" type="checkbox" /> 启用，重启后生效</label>
        <button type="submit">保存 MySQL 配置</button>
      </form>
      <pre>{{ JSON.stringify(state.mysql, null, 2) }}</pre>
    </section>
  </main>
</template>
