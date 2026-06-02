<script setup lang="ts">
import Button from 'primevue/button'
import Card from 'primevue/card'
import { useToast } from 'primevue/usetoast'
import SessionConnectionBanner from '~/components/session/SessionConnectionBanner.vue'
import SessionRevealPanel from '~/components/session/SessionRevealPanel.vue'
import SessionTimerBar from '~/components/session/SessionTimerBar.vue'
import { usePhaseTimer } from '~/composables/session/usePhaseTimer'
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

const CONNECTION_LOST_ERROR = 'Connection lost'

const isBootstrapping = ref(true)
const currentQuestion = computed(() => sessionStore.currentQuestion)
const isStartPending = ref(false)
const isFinishPending = ref(false)
const hadActiveConnection = ref(false)
const showLoader = computed(() => isBootstrapping.value || (sessionStore.isConnected && !sessionStore.isSnapshotLoaded))

const sessionIdFromQuery = computed(() => {
  return typeof route.query.session_id === 'string' ? route.query.session_id.trim() : ''
})

const { timerLabel, timerProgress } = usePhaseTimer({
  phase: toRef(sessionStore, 'phase'),
  deadlineAt: toRef(sessionStore, 'deadlineAt'),
  revealUntil: toRef(sessionStore, 'revealUntil'),
  questionTimeLimitSeconds: computed(() => currentQuestion.value?.time_limit_seconds ?? null),
  revealDurationSec: toRef(sessionStore, 'revealDurationSec'),
})

const connectHost = async () => {
  if (!sessionIdFromQuery.value) {
    return
  }

  try {
    await sessionStore.hostConnect(sessionIdFromQuery.value, authStore.accessToken ?? undefined)
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

const joinLink = computed(() => {
  if (!sessionStore.roomCode || !import.meta.client) {
    return ''
  }

  const url = new URL(window.location.origin)
  url.pathname = '/'
  url.searchParams.set('room_code', sessionStore.roomCode)
  return url.toString()
})

const shareJoinLink = async () => {
  if (!joinLink.value) {
    return
  }

  try {
    if (navigator.share) {
      await navigator.share({
        title: 'Подключение к игре',
        text: `Код комнаты: ${sessionStore.roomCode ?? ''}`,
        url: joinLink.value,
      })
      return
    }

    await navigator.clipboard.writeText(joinLink.value)
    toast.add({
      group: 'global',
      severity: 'success',
      summary: 'Ссылка скопирована',
      detail: 'Ссылка для игроков скопирована в буфер обмена',
      life: 2200,
    })
  } catch {
    toast.add({
      group: 'global',
      severity: 'warn',
      summary: 'Не удалось поделиться ссылкой',
      detail: 'Скопируйте ссылку вручную',
      life: 2800,
    })
  }
}

const runStartGame = async () => {
  if (!sessionStore.isConnected || isStartPending.value) {
    return
  }

  isStartPending.value = true

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
  } finally {
    isStartPending.value = false
  }
}

const runFinishGame = async () => {
  if (!sessionStore.isConnected || isFinishPending.value) {
    return
  }

  isFinishPending.value = true

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
  } finally {
    isFinishPending.value = false
  }
}

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
          detail: 'Можно продолжать управление сессией.',
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
      detail: 'Обновите страницу или вернитесь к списку квизов.',
      life: 5000,
    })
  },
)

onMounted(async () => {
  sessionStore.reset()
  await authStore.initializeSession()

  if (!sessionIdFromQuery.value) {
    await router.replace('/quizzes')
    return
  }

  if (sessionStore.role && sessionStore.role !== 'host') {
    sessionStore.reset()
  }

  await connectHost()

  isBootstrapping.value = false
})

useHead({
  title: 'Хост сессии',
})
</script>

<template>
  <section class="grid min-h-[calc(100dvh-1.5rem)] place-items-center">
    <Card class="w-full max-w-(--app-card-wide)">
      <template #content>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <SessionConnectionBanner :status="sessionStore.connectionStatus" :room-code="sessionStore.roomCode" />

          <div class="inline-flex items-center gap-1">
            <Button v-if="sessionStore.roomCode" label="Скопировать код" icon="pi pi-copy" text @click="copyRoomCode" />
            <Button
              v-if="sessionStore.roomCode"
              label="Ссылка для игроков"
              icon="pi pi-share-alt"
              text
              @click="shareJoinLink"
            />
          </div>
        </div>

        <SessionTimerBar v-if="!showLoader && sessionStore.phase !== 'lobby' && sessionStore.phase !== 'finished'" :label="timerLabel" :progress="timerProgress" class="my-3" />

        <div v-if="showLoader" class="flex flex-col gap-4">
          <p>Подготавливаем сессию...</p>
        </div>

        <div v-else-if="!sessionIdFromQuery" class="flex flex-col gap-4">
          <h1 class="m-0 text-[clamp(1.4rem,2.2vw,2rem)] leading-tight">Сессия не выбрана</h1>
          <p class="m-0 text-(--app-color-text-muted)">Перейдите в список квизов и создайте сессию для запуска игры.</p>
          <Button label="К списку квизов" icon="pi pi-list" @click="router.push('/quizzes')" />
        </div>

        <div v-else-if="sessionStore.phase === 'lobby'" class="flex flex-col gap-4">
          <h1 class="m-0 text-[clamp(1.4rem,2.2vw,2rem)] leading-tight">Лобби</h1>
          <p class="m-0 text-(--app-color-text-muted)">Игроков подключено: {{ sessionStore.playersCount }}</p>
          <div class="flex flex-wrap items-center gap-2">
            <Button
              label="Начать игру"
              icon="pi pi-play"
              :disabled="!sessionStore.isConnected || isStartPending"
              :loading="isStartPending"
              @click="runStartGame"
            />
          </div>
        </div>

        <div v-else-if="sessionStore.phase === 'question_open'" class="flex flex-col gap-4">
          <h1 class="m-0 text-[clamp(1.4rem,2.2vw,2rem)] leading-tight" v-if="currentQuestion">{{ currentQuestion.text }}</h1>
          <p class="m-0 text-(--app-color-text-muted)">
            Вопрос {{ sessionStore.currentQuestionNumber }}
            <span v-if="sessionStore.totalQuestions">/ {{ sessionStore.totalQuestions }}</span>
          </p>

          <p class="m-0 text-(--app-color-text-muted)" v-if="sessionStore.answeredCount !== null">
            Ответов: {{ sessionStore.answeredCount }} / {{ sessionStore.totalPlayers ?? sessionStore.playersCount }}
          </p>

          <div class="flex flex-wrap items-center gap-2">
            <Button
              label="Завершить игру"
              severity="danger"
              icon="pi pi-stop"
              :disabled="!sessionStore.isConnected || isFinishPending"
              :loading="isFinishPending"
              @click="runFinishGame"
            />
          </div>
        </div>

        <SessionRevealPanel
          v-else-if="sessionStore.phase === 'answer_reveal' || sessionStore.phase === 'leaderboard_reveal'"
          :phase="sessionStore.phase"
          :entries="sessionStore.leaderboardTop"
        >
          <div class="flex flex-wrap items-center gap-2">
            <Button
              label="Завершить игру"
              severity="danger"
              icon="pi pi-stop"
              :disabled="!sessionStore.isConnected || isFinishPending"
              :loading="isFinishPending"
              @click="runFinishGame"
            />
          </div>
        </SessionRevealPanel>

        <SessionRevealPanel v-else-if="sessionStore.phase === 'finished'" phase="finished" :entries="sessionStore.leaderboardTop" />
      </template>
    </Card>
  </section>
</template>
