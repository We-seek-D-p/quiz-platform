<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'
import ProgressBar from 'primevue/progressbar'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { useGameSessionStore } from '~/stores/gameSession'

definePageMeta({
  layout: 'game',
  title: 'Игра',
})

const route = useRoute()
const router = useRouter()
const toast = useToast()
const sessionStore = useGameSessionStore()

const isBootstrapping = ref(true)
const currentQuestion = computed(() => sessionStore.currentQuestion)

const selectionTypeLabel = computed(() => {
  if (!sessionStore.currentQuestion) {
    return ''
  }

  return currentQuestion.value.selection_type === 'multiple'
    ? 'Выберите несколько вариантов'
    : 'Выберите один вариант'
})

const timerProgress = ref(0)
const timerLabel = ref('--')

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

const tryAutoConnect = async () => {
  const roomFromQuery = typeof route.query.room_code === 'string' ? route.query.room_code.trim() : ''
  const nicknameFromQuery = typeof route.query.nickname === 'string' ? route.query.nickname.trim() : ''

  try {
    await sessionStore.playerReconnect(roomFromQuery || undefined)
    return
  } catch {
    if (!roomFromQuery || !nicknameFromQuery) {
      await router.replace('/')
      return
    }

    try {
      await sessionStore.playerJoin(roomFromQuery, nicknameFromQuery)
    } catch (error: unknown) {
      const detail = error instanceof Error ? error.message : 'Попробуйте снова'
      toast.add({
        group: 'global',
        severity: 'error',
        summary: 'Не удалось подключиться к игре',
        detail,
        life: 3500,
      })
      await router.replace({ path: '/', query: roomFromQuery ? { room_code: roomFromQuery } : {} })
    }
  }
}

const retryConnection = async () => {
  if (sessionStore.connectionStatus === 'connected' || sessionStore.connectionStatus === 'connecting') {
    return
  }

  await tryAutoConnect()
}

const toggleOption = (optionId: string) => {
  sessionStore.toggleSelectedOption(optionId)
}

const submitAnswer = () => {
  try {
    sessionStore.submitCurrentAnswer()
  } catch (error: unknown) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Ошибка отправки',
      detail: error instanceof Error ? error.message : 'Не удалось отправить ответ',
      life: 3000,
    })
  }
}

watch(
  () => sessionStore.shouldReturnToJoin,
  async (value) => {
    if (!value) {
      return
    }

    const query: Record<string, string> = {}
    if (sessionStore.roomCode) {
      query.room_code = sessionStore.roomCode
    }

    await router.replace({ path: '/', query })
  },
)

watch(
  () => [sessionStore.phase, sessionStore.deadlineAt, sessionStore.revealUntil, currentQuestion.value?.time_limit_seconds] as const,
  () => {
    startTimer()
  },
)

onMounted(async () => {
  await tryAutoConnect()
  startTimer()
  isBootstrapping.value = false
})

onBeforeUnmount(() => {
  clearTimer()
})

useHead({
  title: 'Игра',
})
</script>

<template>
  <section class="game-screen">
    <Card class="game-screen__card">
      <template #content>
        <div class="game-screen__header">
          <div class="game-screen__header-left">
            <Tag
              :severity="sessionStore.isConnected ? 'success' : sessionStore.isReconnecting ? 'warn' : 'danger'"
              :value="sessionStore.isConnected ? 'Подключено' : sessionStore.isReconnecting ? 'Переподключение' : 'Не в сети'"
            />
            <span class="game-screen__room">Комната: {{ sessionStore.roomCode ?? '--------' }}</span>
          </div>

          <span class="game-screen__timer">{{ timerLabel }}</span>
        </div>

        <ProgressBar :value="timerProgress" :show-value="false" class="game-screen__progress" />

        <Message v-if="sessionStore.lastError" severity="warn" :closable="false">
          {{ sessionStore.lastError }}
        </Message>

        <Message v-if="sessionStore.reconnectNotice" severity="success" :closable="false">
          {{ sessionStore.reconnectNotice }}
        </Message>

        <div v-if="isBootstrapping" class="game-screen__state">
          <p>Подключаемся к игровой сессии...</p>
        </div>

        <div
          v-else-if="sessionStore.connectionStatus === 'disconnected' && !sessionStore.shouldReturnToJoin"
          class="game-screen__state"
        >
          <p>Соединение потеряно. Попробуйте подключиться снова.</p>
          <Button label="Переподключиться" icon="pi pi-refresh" @click="retryConnection" />
        </div>

        <div v-else-if="sessionStore.phase === 'lobby'" class="game-screen__state">
          <h1 class="game-screen__title">Лобби</h1>
          <p class="game-screen__subtitle">Ожидайте начала игры от хоста</p>
          <p class="game-screen__meta">Игроков в комнате: {{ sessionStore.playersCount }}</p>
        </div>

        <div v-else-if="sessionStore.phase === 'question_open' && currentQuestion" class="game-screen__question">
          <div class="game-screen__question-head">
            <p class="game-screen__meta">
              Вопрос {{ sessionStore.currentQuestionNumber }}
              <span v-if="sessionStore.totalQuestions">/ {{ sessionStore.totalQuestions }}</span>
            </p>
            <p class="game-screen__subtitle">{{ selectionTypeLabel }}</p>
          </div>

          <h1 class="game-screen__title">{{ currentQuestion.text }}</h1>

          <div class="game-screen__options">
            <button
              v-for="option in currentQuestion.options"
              :key="option.id"
              type="button"
              class="option-btn"
              :class="{ 'option-btn--active': sessionStore.selectedOptionIds.includes(option.id) }"
              :disabled="sessionStore.hasSubmittedAnswer"
              @click="toggleOption(option.id)"
            >
              <span>{{ option.text }}</span>
              <i v-if="sessionStore.selectedOptionIds.includes(option.id)" class="pi pi-check" />
            </button>
          </div>

          <Message v-if="sessionStore.answerSubmitError" severity="error" :closable="false">
            {{ sessionStore.answerSubmitError }}
          </Message>

          <div class="game-screen__actions">
            <Button
              label="Ответить"
              icon="pi pi-send"
              :disabled="!sessionStore.canSubmitAnswer"
              :loading="sessionStore.isSubmittingAnswer"
              @click="submitAnswer"
            />
            <Tag v-if="sessionStore.hasSubmittedAnswer" severity="success" value="Ответ принят" />
          </div>
        </div>

        <div v-else-if="sessionStore.phase === 'answer_reveal'" class="game-screen__state">
          <h1 class="game-screen__title">Ответы раскрыты</h1>
          <p class="game-screen__subtitle">Сейчас откроется следующий вопрос</p>
          <p class="game-screen__meta" v-if="sessionStore.myScore !== null">
            Ваш счет: {{ sessionStore.myScore }} · Место: {{ sessionStore.myRank ?? '-' }}
          </p>

          <ol class="game-screen__leaderboard" v-if="sessionStore.leaderboardTop.length > 0">
            <li v-for="entry in sessionStore.leaderboardTop" :key="`${entry.nickname}-${entry.rank}`">
              <span>{{ entry.rank }}. {{ entry.nickname }}</span>
              <strong>{{ entry.score }}</strong>
            </li>
          </ol>
        </div>

        <div v-else-if="sessionStore.phase === 'finished'" class="game-screen__state">
          <h1 class="game-screen__title">Игра завершена</h1>
          <p class="game-screen__subtitle">Финальный рейтинг</p>

          <ol class="game-screen__leaderboard" v-if="sessionStore.leaderboardTop.length > 0">
            <li v-for="entry in sessionStore.leaderboardTop" :key="`${entry.nickname}-${entry.rank}`">
              <span>{{ entry.rank }}. {{ entry.nickname }}</span>
              <strong>{{ entry.score }}</strong>
            </li>
          </ol>

          <div class="game-screen__actions">
            <Button label="Вернуться на главную" outlined @click="router.replace('/')" />
          </div>
        </div>
      </template>
    </Card>
  </section>
</template>

<style scoped>
.game-screen {
  display: grid;
  min-height: calc(100dvh - 1.5rem);
  place-items: center;
}

.game-screen__card {
  width: min(100%, 52rem);
  border-radius: 1.25rem;
}

.game-screen__header {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
}

.game-screen__header-left {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.game-screen__room {
  font-weight: 700;
  letter-spacing: 0.02em;
}

.game-screen__timer {
  font-weight: 700;
}

.game-screen__progress {
  margin-top: 0.75rem;
  margin-bottom: 1rem;
  height: 0.625rem;
}

.game-screen__state,
.game-screen__question {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.game-screen__question-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.game-screen__title {
  margin: 0;
  font-size: clamp(1.5rem, 2.6vw, 2.2rem);
  line-height: 1.2;
}

.game-screen__subtitle,
.game-screen__meta {
  margin: 0;
  color: var(--app-color-text-muted);
}

.game-screen__options {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.option-btn {
  border: 1px solid var(--app-color-border);
  border-radius: 0.875rem;
  padding: 0.9rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--p-content-background);
  color: inherit;
  font: inherit;
  cursor: pointer;
  text-align: left;
  transition:
    border-color var(--app-transition-fast),
    background-color var(--app-transition-fast);
}

.option-btn--active {
  border-color: var(--app-color-primary);
  background: color-mix(in srgb, var(--app-color-primary) 12%, transparent);
}

.option-btn:disabled {
  cursor: default;
  opacity: 0.7;
}

.game-screen__actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.game-screen__leaderboard {
  margin: 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.game-screen__leaderboard li {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border: 1px solid var(--app-color-border);
  border-radius: 0.75rem;
  padding: 0.65rem 0.75rem;
}

@media (max-width: 768px) {
  .game-screen__options {
    grid-template-columns: 1fr;
  }
}
</style>
