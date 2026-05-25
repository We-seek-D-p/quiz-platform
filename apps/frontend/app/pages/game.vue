<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'
import { useToast } from 'primevue/usetoast'
import SessionConnectionBanner from '~/components/session/SessionConnectionBanner.vue'
import SessionLeaderboard from '~/components/session/SessionLeaderboard.vue'
import SessionTimerBar from '~/components/session/SessionTimerBar.vue'
import { usePhaseTimer } from '~/composables/session/usePhaseTimer'
import { useGameSessionStore } from '~/stores/gameSession'

definePageMeta({
  layout: 'game',
  title: 'Игра',
})

const route = useRoute()
const router = useRouter()
const toast = useToast()
const sessionStore = useGameSessionStore()

const PLAYER_TOKEN_STORAGE_KEY = 'quiz:player_token'
const PLAYER_ROOM_CODE_STORAGE_KEY = 'quiz:room_code'

const isBootstrapping = ref(true)
const currentQuestion = computed(() => sessionStore.currentQuestion)
const missingJoinContext = ref(false)

const selectionTypeLabel = computed(() => {
  if (!sessionStore.currentQuestion) {
    return ''
  }

  return currentQuestion.value.selection_type === 'multiple' ? 'Выберите несколько вариантов' : 'Выберите один вариант'
})

const { timerLabel, timerProgress } = usePhaseTimer({
  phase: toRef(sessionStore, 'phase'),
  deadlineAt: toRef(sessionStore, 'deadlineAt'),
  revealUntil: toRef(sessionStore, 'revealUntil'),
  questionTimeLimitSeconds: computed(() => currentQuestion.value?.time_limit_seconds ?? null),
  revealDurationSec: toRef(sessionStore, 'revealDurationSec'),
})

const tryAutoConnect = async () => {
  const roomFromQuery = typeof route.query.room_code === 'string' ? route.query.room_code.trim() : ''
  const nicknameFromQuery = typeof route.query.nickname === 'string' ? route.query.nickname.trim() : ''
  const hasStoredReconnect = import.meta.client
    ? Boolean(localStorage.getItem(PLAYER_TOKEN_STORAGE_KEY) && localStorage.getItem(PLAYER_ROOM_CODE_STORAGE_KEY))
    : false

  if (!roomFromQuery && !hasStoredReconnect) {
    missingJoinContext.value = true
    return
  }

  try {
    await sessionStore.playerReconnect(roomFromQuery || undefined)
    missingJoinContext.value = false
    return
  } catch {
    if (!roomFromQuery || !nicknameFromQuery) {
      missingJoinContext.value = true
      await router.replace('/')
      return
    }

    try {
      await sessionStore.playerJoin(roomFromQuery, nicknameFromQuery)
      missingJoinContext.value = false
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

const returnToJoin = async () => {
  const query: Record<string, string> = {}
  if (sessionStore.roomCode) {
    query.room_code = sessionStore.roomCode
  }

  await router.replace({ path: '/', query })
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

watch(() => sessionStore.lastError, (newError) => {
  if (newError) {
    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Ошибка соединения',
      detail: typeof newError === 'string' ? newError : 'Ошибка подключения к серверу',
      life: 5000
    })
  }
})

watch(() => sessionStore.reconnectNotice, (notice) => {
  if (notice) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Соединение...',
      detail: notice,
      life: 3000
    })
  }
})

onMounted(async () => {
  await tryAutoConnect()
  if (missingJoinContext.value) {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Нет данных для входа',
      detail: 'Введите код комнаты и никнейм для подключения',
      life: 3200,
    })
  }
  isBootstrapping.value = false
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
          <SessionConnectionBanner
            :status="sessionStore.connectionStatus"
            :room-code="sessionStore.roomCode"
            room-prefix="Комната: "
          />
        </div>

        <SessionTimerBar
          v-if="sessionStore.phase !== 'lobby'"
          :label="timerLabel"
          :progress="timerProgress"
          class="game-screen__progress"
        />

        <div v-if="isBootstrapping" class="game-screen__state">
          <p>Подключаемся к игровой сессии...</p>
        </div>

        <div v-else-if="missingJoinContext" class="game-screen__state">
          <p>Нет данных для входа в игру.</p>
          <Button label="Перейти к входу" icon="pi pi-arrow-left" @click="returnToJoin" />
        </div>

        <div
          v-else-if="sessionStore.connectionStatus === 'disconnected' && !sessionStore.shouldReturnToJoin"
          class="game-screen__state"
        >
          <p>Соединение потеряно. Попробуйте подключиться снова.</p>
          <div class="game-screen__actions">
            <Button label="Переподключиться" icon="pi pi-refresh" @click="retryConnection" />
            <Button label="К входу" text icon="pi pi-arrow-left" @click="returnToJoin" />
          </div>
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

          <SessionLeaderboard :entries="sessionStore.leaderboardTop" />
        </div>

        <div v-else-if="sessionStore.phase === 'finished'" class="game-screen__state">
          <h1 class="game-screen__title">Игра завершена</h1>
          <p class="game-screen__subtitle">Финальный рейтинг</p>

          <SessionLeaderboard :entries="sessionStore.leaderboardTop" />

          <div class="game-screen__actions">
            <Button
              v-if="sessionStore.roomCode"
              label="Сыграть еще"
              icon="pi pi-refresh"
              @click="router.replace({ path: '/', query: { room_code: sessionStore.roomCode } })"
            />
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

.game-screen__progress {
  margin-top: 0.75rem;
  margin-bottom: 1rem;
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

@media (max-width: 768px) {
  .game-screen__options {
    grid-template-columns: 1fr;
  }
}
</style>
