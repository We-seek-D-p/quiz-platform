<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import Checkbox from 'primevue/checkbox'
import Message from 'primevue/message'
import RadioButton from 'primevue/radiobutton'
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import SessionConnectionBanner from '~/components/session/SessionConnectionBanner.vue'
import SessionFinishedPanel from '~/components/session/SessionFinishedPanel.vue'
import SessionRevealPanel from '~/components/session/SessionRevealPanel.vue'
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
const CONNECTION_LOST_ERROR = 'Connection lost'

const isBootstrapping = ref(true)
const currentQuestion = computed(() => sessionStore.currentQuestion)
const missingJoinContext = ref(false)
const hadActiveConnection = ref(false)
const showLoader = computed(() => isBootstrapping.value || (sessionStore.isConnected && !sessionStore.isSnapshotLoaded))

const selectionTypeLabel = computed(() => {
  const question = currentQuestion.value
  if (!question) {
    return ''
  }

  return question.selection_type === 'multiple' ? 'Выберите несколько вариантов' : 'Выберите один вариант'
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

const returnToJoin = async () => {
  const query: Record<string, string> = {}
  const currentRoomCode = sessionStore.roomCode
  if (currentRoomCode) {
    query.room_code = currentRoomCode
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
    const currentRoomCode = sessionStore.roomCode
    if (currentRoomCode) {
      query.room_code = currentRoomCode
    }

    await router.replace({ path: '/', query })
  },
)

watch(
  () => sessionStore.lastError,
  (newError) => {
    if (newError === CONNECTION_LOST_ERROR) {
      return
    }

    if (newError) {
      toast.add({
        group: 'global',
        severity: 'error',
        summary: 'Ошибка соединения',
        detail: typeof newError === 'string' ? newError : 'Ошибка подключения к серверу',
        life: 5000,
      })
    }
  },
)

watch(
  () => sessionStore.connectionStatus,
  (status, previousStatus) => {
    if (status === 'connected') {
      if (previousStatus === 'reconnecting' || previousStatus === 'disconnected') {
        toast.add({
          group: 'global',
          severity: 'success',
          summary: 'Соединение восстановлено',
          detail: 'Можно продолжать игру.',
          life: 3000,
        })
      }

      hadActiveConnection.value = true
      return
    }

    if (status === 'reconnecting' && hadActiveConnection.value && previousStatus !== 'reconnecting') {
      toast.add({
        group: 'global',
        severity: 'warn',
        summary: 'Соединение потеряно',
        detail: 'Пробуем переподключиться автоматически...',
        life: 4000,
      })
      return
    }

    if (status !== 'disconnected' || previousStatus !== 'reconnecting' || !hadActiveConnection.value) {
      return
    }

    toast.add({
      group: 'global',
      severity: 'error',
      summary: 'Не удалось восстановить соединение',
      detail: 'Обновите страницу или вернитесь к входу в игру.',
      life: 5000,
    })
  },
)

onMounted(async () => {
  sessionStore.reset()
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
  <section class="grid min-h-[calc(100dvh-1.5rem)] place-items-center">
    <Card class="w-full max-w-[52rem]">
      <template #content>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <SessionConnectionBanner
            :status="sessionStore.connectionStatus"
            :room-code="sessionStore.roomCode"
            room-prefix="Комната: "
          />
        </div>

        <SessionTimerBar
          v-if="!showLoader && sessionStore.phase !== 'lobby' && sessionStore.phase !== 'finished'"
          :label="timerLabel"
          :progress="timerProgress"
          :show-progress="sessionStore.phase === 'question_open'"
          class="mt-3 mb-4"
        />

        <div v-if="showLoader" class="flex flex-col gap-4">
          <p>Подключаемся к игровой сессии...</p>
        </div>

        <div v-else-if="missingJoinContext" class="flex flex-col gap-4">
          <p>Нет данных для входа в игру.</p>
          <Button label="Перейти к входу" icon="pi pi-arrow-left" @click="returnToJoin" />
        </div>

        <div v-else-if="sessionStore.phase === 'lobby'" class="flex flex-col gap-4">
          <h1 class="m-0 text-[clamp(1.5rem,2.6vw,2.2rem)] leading-[1.2]">Лобби</h1>
          <p class="m-0 text-(--app-color-text-muted)">Ожидайте начала игры от хоста</p>
          <p class="m-0 text-(--app-color-text-muted)">Игроков в комнате: {{ sessionStore.playersCount }}</p>
        </div>

        <div v-else-if="['question_open', 'answer_reveal'].includes(sessionStore.phase) && currentQuestion" class="flex flex-col gap-4">
          <div class="flex flex-wrap items-baseline justify-between gap-2">
            <p class="m-0 text-(--app-color-text-muted)">
              Вопрос {{ sessionStore.currentQuestionNumber }}
              <span v-if="sessionStore.totalQuestions">/ {{ sessionStore.totalQuestions }}</span>
            </p>
            <p v-if="sessionStore.phase === 'question_open'" class="m-0 text-(--app-color-text-muted)">{{ selectionTypeLabel }}</p>
            <p v-else-if="sessionStore.myScore !== null" class="m-0 text-(--app-color-text-muted)">
              Ваш счет: {{ sessionStore.myScore }} · Место: {{ sessionStore.myRank ?? '-' }}
            </p>
          </div>

          <h1 class="m-0 text-[clamp(1.5rem,2.6vw,2.2rem)] leading-[1.2]">{{ currentQuestion.text }}</h1>

          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <button
              v-for="option in currentQuestion.options"
              :key="option.id"
              type="button"
              class="option-btn"
              :class="{
                'option-btn--active': sessionStore.phase === 'question_open' && sessionStore.selectedOptionIds.includes(option.id),
                'option-btn--correct': sessionStore.phase === 'answer_reveal' && sessionStore.correctOptionIds.includes(option.id),
                'option-btn--wrong': sessionStore.phase === 'answer_reveal' && sessionStore.selectedOptionIds.includes(option.id) && !sessionStore.correctOptionIds.includes(option.id),
                'option-btn--dimmed': sessionStore.phase === 'answer_reveal' && !sessionStore.correctOptionIds.includes(option.id) && !sessionStore.selectedOptionIds.includes(option.id),
              }"
              :disabled="sessionStore.hasSubmittedAnswer || sessionStore.phase === 'answer_reveal'"
              @click="toggleOption(option.id)"
            >
              <span>{{ option.text }}</span>
              <span class="option-control">
                <RadioButton
                  v-if="currentQuestion.selection_type === 'single'"
                  :model-value="sessionStore.selectedOptionIds[0] ?? null"
                  :value="option.id"
                  :disabled="sessionStore.hasSubmittedAnswer || sessionStore.phase === 'answer_reveal'"
                  tabindex="-1"
                />
                <Checkbox
                  v-else
                  :model-value="sessionStore.selectedOptionIds.includes(option.id)"
                  binary
                  :disabled="sessionStore.hasSubmittedAnswer || sessionStore.phase === 'answer_reveal'"
                  tabindex="-1"
                />
              </span>
            </button>
          </div>

          <Message v-if="sessionStore.answerSubmitError" severity="error" :closable="false">
            {{ sessionStore.answerSubmitError }}
          </Message>

          <div v-if="sessionStore.phase === 'question_open'" class="flex flex-wrap items-center gap-3">
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

        <SessionRevealPanel
          v-else-if="sessionStore.phase === 'leaderboard_reveal'"
          phase="leaderboard_reveal"
          :entries="sessionStore.leaderboardTop"
          :score="sessionStore.myScore"
          :rank="sessionStore.myRank"
        />

        <SessionFinishedPanel
          v-else-if="sessionStore.phase === 'finished'"
          :entries="sessionStore.leaderboardTop"
          :rank="sessionStore.myRank"
        />
      </template>
    </Card>
  </section>
</template>

<style scoped>
.option-btn {
  border: 1px solid var(--app-color-border);
  border-radius: var(--app-control-radius);
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

.option-btn--correct {
  border-color: var(--p-green-500);
  background: color-mix(in srgb, var(--p-green-500) 15%, transparent);
}

.option-btn--wrong {
  border-color: var(--p-red-500);
  background: color-mix(in srgb, var(--p-red-500) 15%, transparent);
}

.option-btn--dimmed {
  opacity: 0.5;
}

.option-btn:disabled {
  cursor: default;
}

.option-control {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
}

.option-control :deep(.p-checkbox),
.option-control :deep(.p-radiobutton) {
  pointer-events: none;
}
</style>
