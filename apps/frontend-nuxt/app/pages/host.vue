<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'
import ProgressBar from 'primevue/progressbar'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { useAuthStore } from '~/stores/auth'
import { useGameSessionStore } from '~/stores/gameSession'

definePageMeta({
  middleware: 'auth',
  layout: 'game',
  title: 'Хост сессии',
})

const route = useRoute()
const router = useRouter()
const toast = useToast()
const authStore = useAuthStore()
const sessionStore = useGameSessionStore()

const isBootstrapping = ref(true)
const currentQuestion = computed(() => sessionStore.currentQuestion)

const timerProgress = ref(0)
const timerLabel = ref('--')

const sessionIdFromQuery = computed(() => {
  return typeof route.query.session_id === 'string' ? route.query.session_id.trim() : ''
})

let timerInterval: ReturnType<typeof setInterval> | null = null

const clearTimer = () => {
  if (!timerInterval) {
    return
  }

  clearInterval(timerInterval)
  timerInterval = null
}

const recomputeTimer = () => {
  const countdownTarget =
    sessionStore.phase === 'question_open'
      ? sessionStore.deadlineAt
      : sessionStore.phase === 'answer_reveal'
        ? sessionStore.revealUntil
        : null

  if (!countdownTarget) {
    timerProgress.value = 0
    timerLabel.value = '--'
    return
  }

  const endMs = new Date(countdownTarget).getTime()
  const nowMs = Date.now()
  const remainingMs = Math.max(0, endMs - nowMs)
  const remainingSec = Math.ceil(remainingMs / 1000)
  timerLabel.value = `${remainingSec}s`

  if (sessionStore.phase === 'question_open' && currentQuestion.value?.time_limit_seconds) {
    const total = Math.max(1, currentQuestion.value.time_limit_seconds)
    timerProgress.value = Math.min(100, Math.max(0, (remainingSec / total) * 100))
    return
  }

  if (sessionStore.phase === 'answer_reveal') {
    const revealWindowMs = Math.max(1, sessionStore.revealDurationSec) * 1000
    timerProgress.value = Math.min(100, Math.max(0, (remainingMs / revealWindowMs) * 100))
    return
  }

  timerProgress.value = 0
}

const startTimer = () => {
  clearTimer()
  recomputeTimer()

  timerInterval = setInterval(() => {
    recomputeTimer()
  }, 300)
}

const connectHost = async () => {
  if (!sessionIdFromQuery.value) {
    return
  }

  try {
    await sessionStore.hostConnect(sessionIdFromQuery.value)
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось подключиться к сессии',
      detail: error instanceof Error ? error.message : 'Попробуйте снова',
      life: 3500,
    })
  }
}

const retryConnection = async () => {
  if (!sessionIdFromQuery.value) {
    return
  }

  await connectHost()
}

const copyRoomCode = async () => {
  if (!sessionStore.roomCode) {
    return
  }

  try {
    await navigator.clipboard.writeText(sessionStore.roomCode)
    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Скопировано',
      detail: `Код ${sessionStore.roomCode} скопирован`,
      life: 2200,
    })
  } catch {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Не удалось скопировать',
      detail: 'Скопируйте код вручную',
      life: 2500,
    })
  }
}

const runStartGame = () => {
  try {
    sessionStore.startGame()
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось начать игру',
      detail: error instanceof Error ? error.message : 'Попробуйте снова',
      life: 3000,
    })
  }
}

const runFinishGame = () => {
  try {
    sessionStore.finishGame()
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось завершить игру',
      detail: error instanceof Error ? error.message : 'Попробуйте снова',
      life: 3000,
    })
  }
}

watch(
  () => [sessionStore.phase, sessionStore.deadlineAt, sessionStore.revealUntil, currentQuestion.value?.time_limit_seconds] as const,
  () => {
    startTimer()
  },
)

onMounted(async () => {
  await authStore.initializeSession()

  if (sessionStore.role && sessionStore.role !== 'host') {
    sessionStore.reset()
  }

  if (sessionIdFromQuery.value) {
    await connectHost()
  }

  startTimer()
  isBootstrapping.value = false
})

onBeforeUnmount(() => {
  clearTimer()
})

useHead({
  title: 'Хост сессии',
})
</script>

<template>
  <section class="host-runtime">
    <Card class="host-runtime__card">
      <template #content>
        <div class="host-runtime__header">
          <div class="host-runtime__status">
            <Tag
              :severity="sessionStore.isConnected ? 'success' : sessionStore.isReconnecting ? 'warn' : 'danger'"
              :value="sessionStore.isConnected ? 'Подключено' : sessionStore.isReconnecting ? 'Переподключение' : 'Не в сети'"
            />
            <span v-if="sessionStore.roomCode" class="host-runtime__room">{{ sessionStore.roomCode }}</span>
          </div>

          <div class="host-runtime__header-actions">
            <span class="host-runtime__timer">{{ timerLabel }}</span>
            <Button
              v-if="sessionStore.roomCode"
              label="Скопировать код"
              icon="pi pi-copy"
              text
              @click="copyRoomCode"
            />
          </div>
        </div>

        <ProgressBar :value="timerProgress" :show-value="false" class="host-runtime__progress" />

        <Message v-if="sessionStore.lastError" severity="warn" :closable="false">{{ sessionStore.lastError }}</Message>
        <Message v-if="sessionStore.reconnectNotice" severity="success" :closable="false">
          {{ sessionStore.reconnectNotice }}
        </Message>

        <div v-if="isBootstrapping" class="host-runtime__state">
          <p>Подготавливаем сессию...</p>
        </div>

        <div v-else-if="!sessionIdFromQuery" class="host-runtime__state">
          <h1 class="host-runtime__title">Сессия не выбрана</h1>
          <p class="host-runtime__subtitle">Перейдите в список квизов и создайте сессию для запуска игры.</p>
          <Button label="К списку квизов" icon="pi pi-list" @click="router.push('/quizzes')" />
        </div>

        <div
          v-else-if="sessionStore.connectionStatus === 'disconnected'"
          class="host-runtime__state"
        >
          <p>Соединение с Session Service потеряно.</p>
          <Button label="Переподключиться" icon="pi pi-refresh" @click="retryConnection" />
        </div>

        <div v-else-if="sessionStore.phase === 'lobby'" class="host-runtime__state">
          <h1 class="host-runtime__title">Лобби</h1>
          <p class="host-runtime__subtitle">Игроков подключено: {{ sessionStore.playersCount }}</p>
          <div class="host-runtime__actions">
            <Button
              label="Начать игру"
              icon="pi pi-play"
              :disabled="!sessionStore.isConnected"
              @click="runStartGame"
            />
          </div>
        </div>

        <div v-else-if="sessionStore.phase === 'question_open'" class="host-runtime__state">
          <h1 class="host-runtime__title" v-if="currentQuestion">{{ currentQuestion.text }}</h1>
          <p class="host-runtime__subtitle">
            Вопрос {{ sessionStore.currentQuestionNumber }}
            <span v-if="sessionStore.totalQuestions">/ {{ sessionStore.totalQuestions }}</span>
          </p>

          <p class="host-runtime__meta" v-if="sessionStore.answeredCount !== null">
            Ответов: {{ sessionStore.answeredCount }} / {{ sessionStore.totalPlayers ?? sessionStore.playersCount }}
          </p>

          <div class="host-runtime__actions">
            <Button label="Завершить игру" severity="danger" icon="pi pi-stop" @click="runFinishGame" />
          </div>
        </div>

        <div v-else-if="sessionStore.phase === 'answer_reveal'" class="host-runtime__state">
          <h1 class="host-runtime__title">Промежуточный рейтинг</h1>
          <p class="host-runtime__subtitle">Следующий вопрос откроется автоматически</p>

          <ol class="host-runtime__leaderboard" v-if="sessionStore.leaderboardTop.length > 0">
            <li v-for="entry in sessionStore.leaderboardTop" :key="`${entry.nickname}-${entry.rank}`">
              <span>{{ entry.rank }}. {{ entry.nickname }}</span>
              <strong>{{ entry.score }}</strong>
            </li>
          </ol>

          <div class="host-runtime__actions">
            <Button label="Завершить игру" severity="danger" icon="pi pi-stop" @click="runFinishGame" />
          </div>
        </div>

        <div v-else-if="sessionStore.phase === 'finished'" class="host-runtime__state">
          <h1 class="host-runtime__title">Игра завершена</h1>
          <p class="host-runtime__subtitle">Финальный leaderboard</p>

          <ol class="host-runtime__leaderboard" v-if="sessionStore.leaderboardTop.length > 0">
            <li v-for="entry in sessionStore.leaderboardTop" :key="`${entry.nickname}-${entry.rank}`">
              <span>{{ entry.rank }}. {{ entry.nickname }}</span>
              <strong>{{ entry.score }}</strong>
            </li>
          </ol>

          <div class="host-runtime__actions">
            <Button label="Новая сессия" icon="pi pi-plus" @click="router.push('/quizzes')" />
          </div>
        </div>
      </template>
    </Card>
  </section>
</template>

<style scoped>
.host-runtime {
  display: grid;
  min-height: calc(100dvh - 1.5rem);
  place-items: center;
}

.host-runtime__card {
  width: min(100%, 56rem);
  border-radius: 1.25rem;
}

.host-runtime__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.75rem;
}

.host-runtime__status {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.host-runtime__room {
  font-size: clamp(1.2rem, 2.5vw, 1.8rem);
  font-weight: 800;
  letter-spacing: 0.08em;
}

.host-runtime__header-actions {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.host-runtime__timer {
  font-weight: 700;
}

.host-runtime__progress {
  margin: 0.75rem 0 1rem;
  height: 0.625rem;
}

.host-runtime__state {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.host-runtime__title {
  margin: 0;
  font-size: clamp(1.4rem, 2.2vw, 2rem);
  line-height: 1.25;
}

.host-runtime__subtitle,
.host-runtime__meta {
  margin: 0;
  color: var(--app-color-text-muted);
}

.host-runtime__actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.host-runtime__leaderboard {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.host-runtime__leaderboard li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid var(--app-color-border);
  border-radius: 0.75rem;
  padding: 0.65rem 0.75rem;
}
</style>
