<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref } from 'vue'

type AnyRecord = Record<string, any>

const isAdmin = window.location.pathname === '/__admin'
const adminKey = ref(new URLSearchParams(window.location.search).get('appkey') || localStorage.getItem('adminAppKey') || '')
const activeTab = ref('status')
const loading = ref(false)
const message = ref('')
const error = ref('')

type WorkStatus = 'idle' | 'working' | 'paused' | 'finished' | 'failed' | 'patrolWarning' | 'patrolActive' | 'patrolResult'

const routePath = ref(window.location.pathname)
const userLoading = ref(false)
const userMessage = ref('')
const userError = ref('')
const autoFinishing = ref(false)
let timerID: number | undefined

const user = reactive<AnyRecord>({
  runtime: null,
  scenes: [],
  settings: null,
  selectedSceneKey: '',
  status: 'idle' as WorkStatus,
  sessionId: 0,
  plannedDurationSeconds: 1500,
  accumulatedSeconds: 0,
  lastTickAt: 0,
  settlement: null,
  patrol: {
    nextRemainingMs: 0,
    lastTickAt: 0,
    count: 0,
    warningCount: 0,
    violationCount: 0,
    status: '',
    reason: '',
    action: null,
    failed: false,
    finishReason: '',
    captureErrorCode: '',
  },
})

const cameraPreview = ref<HTMLVideoElement | null>(null)
const captureCanvas = ref<HTMLCanvasElement | null>(null)
const actionVideo = ref<HTMLVideoElement | null>(null)
let cameraStream: MediaStream | null = null
let patrolRunning = false

const currentScene = computed(() => user.scenes.find((scene: AnyRecord) => scene.sceneKey === user.selectedSceneKey) || user.scenes[0] || null)
const currentSetting = computed(() => user.settings || {})
const isTimingStatus = computed(() => ['working', 'patrolWarning', 'patrolActive', 'patrolResult'].includes(user.status))
const currentElapsedSeconds = computed(() => {
  const ticking = isTimingStatus.value && user.lastTickAt ? Math.floor((Date.now() - user.lastTickAt) / 1000) : 0
  return user.accumulatedSeconds + ticking
})
const timerSeconds = computed(() => {
  if (currentSetting.value.mode === 'infinite') return currentElapsedSeconds.value
  return Math.max(user.plannedDurationSeconds - currentElapsedSeconds.value, 0)
})
const timerText = computed(() => formatSeconds(timerSeconds.value))
const statusText = computed(() => {
  switch (user.status) {
    case 'working':
      return '劳动中'
    case 'patrolWarning':
      return '巡查预告'
    case 'patrolActive':
      return '巡查中'
    case 'patrolResult':
      return '巡查结果'
    case 'paused':
      return '已暂停'
    case 'finished':
      return '已结算'
    case 'failed':
      return '异常结束'
    default:
      return '待命中，等待开始劳动'
  }
})
const routeName = computed(() => {
  if (routePath.value === '/settings') return 'settings'
  if (routePath.value === '/settlement') return 'settlement'
  return 'home'
})

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
  } else if (!isAdmin) {
    window.addEventListener('popstate', syncRoute)
    window.addEventListener('beforeunload', finishOnUnload)
    void loadUserApp()
    timerID = window.setInterval(tickTimer, 250)
  }
})

onUnmounted(() => {
  window.removeEventListener('popstate', syncRoute)
  window.removeEventListener('beforeunload', finishOnUnload)
  if (timerID) window.clearInterval(timerID)
  stopCamera()
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

async function userRun(task: () => Promise<void>, ok = '') {
  userLoading.value = true
  userError.value = ''
  userMessage.value = ''
  try {
    await task()
    userMessage.value = ok
  } catch (err) {
    userError.value = err instanceof Error ? err.message : String(err)
  } finally {
    userLoading.value = false
  }
}

async function loadUserApp() {
  await userRun(async () => {
    const [runtime, scenes, settings] = await Promise.all([
      api('/api/runtime/config'),
      api('/api/scenes'),
      api('/api/settings'),
    ])
    user.runtime = runtime
    user.scenes = scenes.items || []
    user.settings = settings
    const cached = localStorage.getItem('userSettings')
    if (cached) localStorage.removeItem('userSettings')
    user.selectedSceneKey = user.scenes.find((scene: AnyRecord) => scene.sceneKey === runtime?.character?.defaultSceneKey)?.sceneKey || user.scenes[0]?.sceneKey || ''
    user.plannedDurationSeconds = plannedDurationFromSettings()
    const abandoned = localStorage.getItem('openSessionAbandoned')
    if (abandoned) {
      userMessage.value = '上一局已回到待命；如曾刷新或关闭页面，后端会在下一次开始时标记异常结束。'
      localStorage.removeItem('openSessionAbandoned')
    }
  })
}

function syncRoute() {
  routePath.value = window.location.pathname
}

function go(path: string) {
  window.history.pushState({}, '', path)
  syncRoute()
}

function plannedDurationFromSettings() {
  if (!user.settings) return 1500
  if (user.settings.mode === 'infinite') return 0
  if (user.settings.mode === 'custom') return Number(user.settings.customDurationSeconds || 1500)
  return 1500
}

async function startWork() {
  if (!user.selectedSceneKey) {
    userError.value = '暂无可用场景，请联系管理员配置'
    return
  }
  await userRun(async () => {
    user.plannedDurationSeconds = plannedDurationFromSettings()
    user.accumulatedSeconds = 0
    user.lastTickAt = Date.now()
    const result = await api('/api/session/start', {
      method: 'POST',
      body: JSON.stringify({
        sceneKey: user.selectedSceneKey,
        mode: user.settings.mode,
        plannedDurationSeconds: user.plannedDurationSeconds,
        userConfig: {
          patrolFrequency: user.settings.patrolFrequency,
          cameraEnabled: user.settings.cameraEnabled,
        },
      }),
    })
    user.sessionId = result.session.id
    user.status = result.session.status
    user.patrol.count = result.session.patrolCount || 0
    user.patrol.warningCount = result.session.warningCount || 0
    user.patrol.violationCount = result.session.violationCount || 0
    resetPatrolState()
    scheduleNextPatrol()
    user.settlement = null
    localStorage.setItem('openSessionAbandoned', String(user.sessionId))
    if (user.settings.cameraEnabled) void ensureCamera()
  })
}

async function pauseWork() {
  if (!user.sessionId) return
  await userRun(async () => {
    user.accumulatedSeconds = currentElapsedSeconds.value
    user.lastTickAt = 0
    freezePatrolTimer()
    const result = await api('/api/session/pause', {
      method: 'POST',
      body: JSON.stringify({ sessionId: user.sessionId }),
    })
    user.status = result.session.status
  })
}

async function resumeWork() {
  if (!user.sessionId) return
  await userRun(async () => {
    const result = await api('/api/session/resume', {
      method: 'POST',
      body: JSON.stringify({ sessionId: user.sessionId }),
    })
    user.status = result.session.status
    user.lastTickAt = Date.now()
    resumePatrolTimer()
  })
}

async function stopWork() {
  if (!window.confirm('确认主动结束本次劳动吗？')) return
  await finishWork('user_stop')
}

async function finishWork(finishReason: string) {
  if (!user.sessionId) return
  const actualFocusSeconds = currentElapsedSeconds.value
  user.accumulatedSeconds = actualFocusSeconds
  user.lastTickAt = 0
  await userRun(async () => {
    const result = await api('/api/session/finish', {
      method: 'POST',
      body: JSON.stringify({
        sessionId: user.sessionId,
        finishReason,
        actualFocusSeconds,
      }),
    })
    user.settlement = result.settlement
    user.status = result.settlement.result === 'failed' || result.settlement.result === 'abandoned' ? 'failed' : 'finished'
    resetPatrolState()
    stopCamera()
    localStorage.removeItem('openSessionAbandoned')
    go('/settlement')
  })
  autoFinishing.value = false
}

function tickTimer() {
  if (!isTimingStatus.value) return
  if (!autoFinishing.value && user.settings?.mode !== 'infinite' && user.plannedDurationSeconds > 0 && currentElapsedSeconds.value >= user.plannedDurationSeconds) {
    autoFinishing.value = true
    void finishWork('countdown_complete')
    return
  }
  tickPatrolTimer()
}

function patrolRange() {
  const frequency = user.settings?.patrolFrequency || 'normal'
  return user.runtime?.patrolRule?.[frequency] || user.runtime?.patrolRule?.normal || { minSeconds: 60, maxSeconds: 240 }
}

function scheduleNextPatrol() {
  const range = patrolRange()
  const min = Math.max(1, Number(range.minSeconds || 60))
  const max = Math.max(min, Number(range.maxSeconds || min))
  user.patrol.nextRemainingMs = Math.floor((min + Math.random() * (max - min)) * 1000)
  user.patrol.lastTickAt = Date.now()
}

function resetPatrolState() {
  user.patrol.nextRemainingMs = 0
  user.patrol.lastTickAt = 0
  user.patrol.status = ''
  user.patrol.reason = ''
  user.patrol.action = null
  user.patrol.failed = false
  user.patrol.finishReason = ''
  user.patrol.captureErrorCode = ''
  patrolRunning = false
}

function freezePatrolTimer() {
  if (!user.patrol.lastTickAt) return
  user.patrol.nextRemainingMs = Math.max(0, user.patrol.nextRemainingMs - (Date.now() - user.patrol.lastTickAt))
  user.patrol.lastTickAt = 0
}

function resumePatrolTimer() {
  if (!user.patrol.nextRemainingMs) scheduleNextPatrol()
  user.patrol.lastTickAt = Date.now()
}

function tickPatrolTimer() {
  if (user.status !== 'working' || patrolRunning || !user.sessionId) return
  if (!user.patrol.nextRemainingMs) scheduleNextPatrol()
  const now = Date.now()
  const elapsed = user.patrol.lastTickAt ? now - user.patrol.lastTickAt : 0
  user.patrol.lastTickAt = now
  user.patrol.nextRemainingMs = Math.max(0, user.patrol.nextRemainingMs - elapsed)
  if (user.patrol.nextRemainingMs <= 0) {
    void runPatrol()
  }
}

async function runPatrol() {
  if (patrolRunning || !user.sessionId) return
  patrolRunning = true
  user.patrol.captureErrorCode = ''
  try {
    user.status = 'patrolWarning'
    await wait(1200)
    user.status = 'patrolActive'
    const capture = await captureFrame()
    const result = await api('/api/patrol/check', {
      method: 'POST',
      body: JSON.stringify({
        sessionId: user.sessionId,
        sceneKey: user.selectedSceneKey,
        imageBase64: capture.imageBase64,
        cameraEnabled: capture.cameraEnabled,
        manualViolation: false,
        captureErrorCode: capture.captureErrorCode,
      }),
    })
    user.patrol.status = result.status
    user.patrol.reason = result.reason
    user.patrol.action = result.action
    user.patrol.count = result.sessionSummary?.patrolCount || user.patrol.count
    user.patrol.warningCount = result.sessionSummary?.warningCount || 0
    user.patrol.violationCount = result.sessionSummary?.violationCount || 0
    user.patrol.failed = Boolean(result.sessionSummary?.failed)
    user.patrol.finishReason = result.sessionSummary?.finishReason || ''
    user.status = 'patrolResult'
    await playPatrolAction()
    if (user.patrol.failed) {
      await finishAfterPatrolFailure()
      return
    }
    user.status = 'working'
    scheduleNextPatrol()
  } catch (err) {
    userError.value = err instanceof Error ? err.message : String(err)
    user.status = 'working'
    scheduleNextPatrol()
  } finally {
    patrolRunning = false
  }
}

async function finishAfterPatrolFailure() {
  user.accumulatedSeconds = currentElapsedSeconds.value
  user.lastTickAt = 0
  const reason = user.patrol.finishReason || 'max_warning'
  const result = await api('/api/session/finish', {
    method: 'POST',
    body: JSON.stringify({
      sessionId: user.sessionId,
      finishReason: reason,
      actualFocusSeconds: user.accumulatedSeconds,
    }),
  })
  user.settlement = result.settlement
  user.status = 'failed'
  localStorage.removeItem('openSessionAbandoned')
  resetPatrolState()
  stopCamera()
  go('/settlement')
}

async function playPatrolAction() {
  await nextTick()
  const video = actionVideo.value
  if (!video || !user.patrol.action?.videoUrl) {
    await wait(1400)
    return
  }
  await new Promise<void>((resolve) => {
    const cleanup = () => {
      video.removeEventListener('ended', onDone)
      video.removeEventListener('error', onError)
    }
    const onDone = () => {
      cleanup()
      resolve()
    }
    const onError = () => {
      cleanup()
      userError.value = '动作视频加载失败，请检查管理员配置。'
      resolve()
    }
    video.volume = Number(user.settings?.actionVolume ?? 0.8)
    video.addEventListener('ended', onDone, { once: true })
    video.addEventListener('error', onError, { once: true })
    void video.play().catch(() => onError())
  })
}

async function ensureCamera() {
  if (cameraStream) return cameraStream
  if (!navigator.mediaDevices?.getUserMedia) throw new Error('DEVICE_NOT_FOUND')
  const constraints: MediaStreamConstraints = {
    video: user.settings?.cameraDeviceId ? { deviceId: { exact: user.settings.cameraDeviceId } } : true,
    audio: false,
  }
  cameraStream = await navigator.mediaDevices.getUserMedia(constraints)
  if (cameraPreview.value) {
    cameraPreview.value.srcObject = cameraStream
    await cameraPreview.value.play().catch(() => undefined)
  }
  return cameraStream
}

function stopCamera() {
  if (cameraStream) {
    cameraStream.getTracks().forEach((track) => track.stop())
    cameraStream = null
  }
  if (cameraPreview.value) cameraPreview.value.srcObject = null
}

async function captureFrame() {
  if (!user.settings?.cameraEnabled) {
    return { imageBase64: '', cameraEnabled: false, captureErrorCode: '' }
  }
  try {
    await ensureCamera()
    const video = cameraPreview.value
    const canvas = captureCanvas.value
    if (!video || !canvas || video.videoWidth <= 0 || video.videoHeight <= 0) {
      return { imageBase64: '', cameraEnabled: true, captureErrorCode: 'CAPTURE_FAILED' }
    }
    const maxWidth = Number(user.runtime?.vision?.maxImageWidth || 1024)
    const scale = Math.min(1, maxWidth / video.videoWidth)
    canvas.width = Math.max(1, Math.round(video.videoWidth * scale))
    canvas.height = Math.max(1, Math.round(video.videoHeight * scale))
    const context = canvas.getContext('2d')
    if (!context) return { imageBase64: '', cameraEnabled: true, captureErrorCode: 'CAPTURE_FAILED' }
    context.drawImage(video, 0, 0, canvas.width, canvas.height)
    return { imageBase64: canvas.toDataURL('image/jpeg', 0.82), cameraEnabled: true, captureErrorCode: '' }
  } catch (err) {
    const name = err instanceof DOMException ? err.name : ''
    const code = name === 'NotAllowedError' || name === 'SecurityError' ? 'PERMISSION_DENIED' : name === 'NotFoundError' ? 'DEVICE_NOT_FOUND' : 'CAPTURE_FAILED'
    return { imageBase64: '', cameraEnabled: true, captureErrorCode: code }
  }
}

function wait(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms))
}

function finishOnUnload() {
  if (!user.sessionId || (!isTimingStatus.value && user.status !== 'paused')) return
  const payload = JSON.stringify({
    sessionId: user.sessionId,
    finishReason: 'page_unload',
    actualFocusSeconds: currentElapsedSeconds.value,
  })
  navigator.sendBeacon('/api/session/finish', new Blob([payload], { type: 'application/json' }))
}

async function saveUserSettings() {
  await userRun(async () => {
    const saved = await api('/api/settings', {
      method: 'PUT',
      body: JSON.stringify(user.settings),
    })
    user.settings = saved
    user.plannedDurationSeconds = plannedDurationFromSettings()
    localStorage.setItem('userSettings', JSON.stringify(saved))
    if (!saved.cameraEnabled) stopCamera()
  }, '配置已保存')
}

function selectDuration(seconds: number) {
  user.settings.mode = 'custom'
  user.settings.customDurationSeconds = seconds
  user.plannedDurationSeconds = seconds
}

async function enterFullscreen() {
  if (!document.fullscreenElement) {
    await document.documentElement.requestFullscreen()
  } else {
    await document.exitFullscreen()
  }
}

function showPlaceholder(name: string) {
  userMessage.value = `${name} 会在后续里程碑接入。`
}

function formatSeconds(total: number) {
  const safe = Math.max(0, Math.floor(total))
  const minutes = Math.floor(safe / 60)
  const seconds = safe % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
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
  <main v-if="!isAdmin" class="user-shell" :class="`filter-${currentSetting.screenFilter || 'normal'}`">
    <section v-if="routeName === 'home'" class="work-layout">
      <section
        class="stage"
        :style="currentScene?.backgroundUrl ? { backgroundImage: `linear-gradient(rgba(11,14,18,.38), rgba(11,14,18,.72)), url('/${currentScene.backgroundUrl}')` } : undefined"
      >
        <video
          v-if="currentScene?.backgroundVideoUrl"
          class="scene-video"
          :src="'/' + currentScene.backgroundVideoUrl"
          autoplay
          muted
          loop
          playsinline
        ></video>
        <video
          v-if="user.status === 'patrolResult' && user.patrol.action?.videoUrl"
          ref="actionVideo"
          class="action-video"
          :src="'/' + user.patrol.action.videoUrl"
          :poster="user.patrol.action.posterUrl ? '/' + user.patrol.action.posterUrl : undefined"
          playsinline
        ></video>
        <div class="stage-overlay">
          <div class="scene-label">{{ currentScene?.name || '暂无场景' }}</div>
          <div class="timer">{{ timerText }}</div>
          <p class="status">{{ statusText }}</p>
          <p v-if="user.patrol.reason" class="patrol-reason">{{ user.patrol.reason }}</p>
          <div class="actions">
            <button v-if="user.status === 'idle' || user.status === 'finished' || user.status === 'failed'" type="button" @click="startWork">开始劳动</button>
            <button v-if="user.status === 'working'" type="button" @click="pauseWork">暂停</button>
            <button v-if="user.status === 'paused'" type="button" @click="resumeWork">继续</button>
            <button v-if="user.status === 'working' || user.status === 'paused'" type="button" class="danger-btn" @click="stopWork">结束</button>
          </div>
        </div>
      </section>

      <aside class="hud">
        <h1>Supervisor Game</h1>
        <label>当前场景
          <select v-model="user.selectedSceneKey" :disabled="user.status === 'working' || user.status === 'paused'">
            <option v-for="scene in user.scenes" :key="scene.sceneKey" :value="scene.sceneKey">{{ scene.name }}</option>
          </select>
        </label>
        <dl>
          <div><dt>警告</dt><dd>{{ user.patrol.warningCount }}/{{ user.runtime?.patrolRule?.maxWarnings || 3 }}</dd></div>
          <div><dt>案底</dt><dd>{{ user.patrol.violationCount }}/{{ user.runtime?.patrolRule?.maxViolations || 3 }}</dd></div>
          <div><dt>巡查</dt><dd>{{ user.patrol.count }} 次</dd></div>
          <div><dt>下次巡查</dt><dd>{{ user.status === 'working' ? formatSeconds(user.patrol.nextRemainingMs / 1000) : '--:--' }}</dd></div>
          <div><dt>徽章等级</dt><dd>Lv.{{ user.settlement?.levelAfter || 1 }}</dd></div>
          <div><dt>货币</dt><dd>{{ user.settlement?.currencyAfter || 0 }}</dd></div>
          <div><dt>巡查结果</dt><dd>{{ user.patrol.status || '待命' }}</dd></div>
        </dl>
        <div class="camera-box">
          <video ref="cameraPreview" autoplay muted playsinline></video>
          <canvas ref="captureCanvas" hidden></canvas>
          <span>{{ user.settings?.cameraEnabled ? '摄像头预览' : '摄像头关闭' }}</span>
        </div>
        <nav class="quick-nav">
          <button type="button" class="secondary" @click="go('/settings')">配置</button>
          <button type="button" class="secondary" @click="showPlaceholder('档案')">档案</button>
          <button type="button" class="secondary" @click="showPlaceholder('任务')">任务</button>
          <button type="button" class="secondary" @click="enterFullscreen">全屏</button>
        </nav>
      </aside>
    </section>

    <section v-else-if="routeName === 'settings'" class="user-panel">
      <div class="panel-title">
        <h1>用户配置</h1>
        <button type="button" class="secondary" @click="go('/')">返回</button>
      </div>
      <form v-if="user.settings" class="settings-grid" @submit.prevent="saveUserSettings">
        <label>劳动模式
          <select v-model="user.settings.mode">
            <option value="pomodoro">番茄钟</option>
            <option value="custom">自定义时长</option>
            <option value="infinite">无限模式</option>
          </select>
        </label>
        <label>自定义时长
          <input v-model.number="user.settings.customDurationSeconds" min="300" step="60" type="number" />
        </label>
        <div class="duration-row">
          <button type="button" class="secondary" @click="selectDuration(900)">15 分钟</button>
          <button type="button" class="secondary" @click="selectDuration(1500)">25 分钟</button>
          <button type="button" class="secondary" @click="selectDuration(2700)">45 分钟</button>
          <button type="button" class="secondary" @click="selectDuration(3600)">60 分钟</button>
        </div>
        <label>巡查频率
          <select v-model="user.settings.patrolFrequency">
            <option value="slow">慢速</option>
            <option value="normal">正常</option>
            <option value="high">高压</option>
          </select>
        </label>
        <label>背景音
          <input v-model="user.settings.backgroundAudioKey" />
        </label>
        <label>背景音量
          <input v-model.number="user.settings.backgroundVolume" min="0" max="1" step="0.05" type="range" />
        </label>
        <label>动作音量
          <input v-model.number="user.settings.actionVolume" min="0" max="1" step="0.05" type="range" />
        </label>
        <label>UI 音量
          <input v-model.number="user.settings.uiVolume" min="0" max="1" step="0.05" type="range" />
        </label>
        <label>画面滤镜
          <select v-model="user.settings.screenFilter">
            <option value="normal">正常</option>
            <option value="grayscale">黑白</option>
            <option value="dark">暗色压迫</option>
          </select>
        </label>
        <label class="check"><input v-model="user.settings.quietPatrolEnabled" type="checkbox" /> 轻声巡查</label>
        <label class="check"><input v-model="user.settings.cameraEnabled" type="checkbox" /> 摄像头开启</label>
        <label>摄像头设备
          <input v-model="user.settings.cameraDeviceId" placeholder="默认设备" />
        </label>
        <button type="submit">保存配置</button>
      </form>
    </section>

    <section v-else class="user-panel settlement-panel">
      <div class="panel-title">
        <h1>结算</h1>
        <button type="button" class="secondary" @click="go('/')">返回主界面</button>
      </div>
      <div v-if="user.settlement" class="settlement-grid">
        <div><span>结果</span><strong>{{ user.settlement.result }}</strong></div>
        <div><span>专注时长</span><strong>{{ formatSeconds(user.settlement.actualFocusSeconds) }}</strong></div>
        <div><span>巡查次数</span><strong>{{ user.settlement.patrolCount }}</strong></div>
        <div><span>违规次数</span><strong>{{ user.settlement.violationCount }}</strong></div>
        <div><span>获得货币</span><strong>{{ user.settlement.earnedCurrency }}</strong></div>
        <div><span>等级变化</span><strong>Lv.{{ user.settlement.levelBefore }} → Lv.{{ user.settlement.levelAfter }}</strong></div>
      </div>
      <video v-if="user.settlement?.settlementAction?.videoUrl" :src="'/' + user.settlement.settlementAction.videoUrl" controls></video>
      <p v-else class="notice">暂无结算信息。</p>
    </section>

    <p v-if="userLoading" class="notice floating-notice">处理中...</p>
    <p v-if="userMessage" class="notice success floating-notice">{{ userMessage }}</p>
    <p v-if="userError" class="notice danger floating-notice">{{ userError }}</p>
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
