import { defineStore } from 'pinia'
import { useSessionWs } from '~/composables/session/useSessionWs'
import type {
  AnswerRevealPayload,
  ConnectionStatus,
  JoinedLobbyPayload,
  LeaderboardEntryView,
  QuestionOpenedPayload,
  QuestionRevealHostPayload,
  QuizQuestionView,
  SessionFinishedPayload,
  SessionPhase,
  SessionRole,
  SessionSnapshotPayload,
  WsEnvelope,
  WsErrorPayload,
} from '~/types/session-ws'

const PLAYER_TOKEN_STORAGE_KEY = 'quiz:player_token'
const PLAYER_ROOM_CODE_STORAGE_KEY = 'quiz:room_code'
const PLAYER_PARTICIPANT_ID_STORAGE_KEY = 'quiz:participant_id'

function getErrorMessage(payload: WsErrorPayload): string {
  const defaultMessage = payload.message || 'Unexpected session error'

  const dictionary: Record<string, string> = {
    room_not_found: 'Игра не найдена',
    nickname_taken: 'Имя уже занято',
    game_already_finished: 'Игра уже закончена',
    invalid_participant_token: 'Reconnect token is invalid',
    answer_already_submitted: 'Вы уже ответили',
    question_not_active: 'На вопрос больше нельзя ответить',
    selection_count_invalid: 'Неправильное количество выбранных ответов',
  }

  return dictionary[payload.code] ?? defaultMessage
}

export const useGameSessionStore = defineStore('game-session', () => {
  const role = ref<SessionRole | null>(null)
  const phase = ref<SessionPhase>('lobby')

  const sessionId = ref<string | null>(null)
  const roomCode = ref<string | null>(null)
  const participantId = ref<string | null>(null)
  const participantToken = ref<string | null>(null)
  const nickname = ref<string | null>(null)

  const currentQuestion = ref<QuizQuestionView | null>(null)
  const questionIndex = ref<number | null>(null)
  const totalQuestions = ref<number | null>(null)
  const deadlineAt = ref<string | null>(null)
  const revealUntil = ref<string | null>(null)
  const playersCount = ref(0)

  const leaderboardTop = ref<LeaderboardEntryView[]>([])
  const myScore = ref<number | null>(null)
  const myRank = ref<number | null>(null)

  const lastError = ref<string | null>(null)
  const connectionStatus = ref<ConnectionStatus>('idle')

  const hostWs = useSessionWs({
    mode: 'host',
    onMessage: (message) => handleMessage(message),
    onError: (message) => {
      lastError.value = message
    },
  })

  const playerWs = useSessionWs({
    mode: 'player',
    onMessage: (message) => handleMessage(message),
    onError: (message) => {
      lastError.value = message
    },
  })

  const activeWs = computed(() => {
    if (role.value === 'host') {
      return hostWs
    }
    if (role.value === 'player') {
      return playerWs
    }

    return null
  })

  watch(
    () => hostWs.status.value,
    (status) => {
      if (role.value === 'host') {
        connectionStatus.value = status
      }
    },
  )

  watch(
    () => playerWs.status.value,
    (status) => {
      if (role.value === 'player') {
        connectionStatus.value = status
      }
    },
  )

  const clearTransientState = () => {
    currentQuestion.value = null
    questionIndex.value = null
    totalQuestions.value = null
    deadlineAt.value = null
    revealUntil.value = null
    leaderboardTop.value = []
    myScore.value = null
    myRank.value = null
    playersCount.value = 0
  }

  const reset = () => {
    hostWs.disconnect()
    playerWs.disconnect()

    role.value = null
    phase.value = 'lobby'

    sessionId.value = null
    roomCode.value = null
    participantId.value = null
    participantToken.value = null
    nickname.value = null
    clearTransientState()
    connectionStatus.value = 'idle'
    lastError.value = null
  }

  const persistPlayerAuth = () => {
    if (!import.meta.client || !participantToken.value || !roomCode.value || !participantId.value) {
      return
    }

    localStorage.setItem(PLAYER_TOKEN_STORAGE_KEY, participantToken.value)
    localStorage.setItem(PLAYER_ROOM_CODE_STORAGE_KEY, roomCode.value)
    localStorage.setItem(PLAYER_PARTICIPANT_ID_STORAGE_KEY, participantId.value)
  }

  const clearPlayerAuth = () => {
    if (!import.meta.client) {
      return
    }

    localStorage.removeItem(PLAYER_TOKEN_STORAGE_KEY)
    localStorage.removeItem(PLAYER_ROOM_CODE_STORAGE_KEY)
    localStorage.removeItem(PLAYER_PARTICIPANT_ID_STORAGE_KEY)
  }

  const hostConnect = async (targetSessionId: string) => {
    role.value = 'host'
    sessionId.value = targetSessionId
    lastError.value = null

    await hostWs.connect()
    hostWs.send('host_connect', { session_id: targetSessionId })
  }

  const playerJoin = async (targetRoomCode: string, targetNickname: string) => {
    role.value = 'player'
    roomCode.value = targetRoomCode
    nickname.value = targetNickname
    lastError.value = null

    await playerWs.connect()
    playerWs.send('player_join', {
      room_code: targetRoomCode,
      nickname: targetNickname,
    })
  }

  const playerReconnect = async (targetRoomCode?: string, targetToken?: string) => {
    role.value = 'player'
    lastError.value = null

    const reconnectRoomCode = targetRoomCode ?? roomCode.value ?? (import.meta.client ? localStorage.getItem(PLAYER_ROOM_CODE_STORAGE_KEY) : null)
    const reconnectToken = targetToken ?? participantToken.value ?? (import.meta.client ? localStorage.getItem(PLAYER_TOKEN_STORAGE_KEY) : null)

    if (!reconnectRoomCode || !reconnectToken) {
      throw new Error('Reconnect context is not available')
    }

    roomCode.value = reconnectRoomCode
    participantToken.value = reconnectToken

    await playerWs.connect()
    playerWs.send('player_reconnect', {
      room_code: reconnectRoomCode,
      participant_token: reconnectToken,
    })
  }

  const startGame = () => {
    if (!sessionId.value) {
      throw new Error('Missing session id')
    }

    hostWs.send('start_game', { session_id: sessionId.value })
  }

  const submitAnswer = (questionId: string, selectedOptionIds: string[]) => {
    playerWs.send('submit_answer', {
      question_id: questionId,
      selected_option_ids: selectedOptionIds,
    })
  }

  const finishGame = () => {
    if (!sessionId.value) {
      throw new Error('Missing session id')
    }

    hostWs.send('finish_game', { session_id: sessionId.value })
  }

  const disconnect = () => {
    activeWs.value?.disconnect()
    connectionStatus.value = 'disconnected'
  }

  const applySessionSnapshot = (payload: SessionSnapshotPayload) => {
    phase.value = payload.status
    sessionId.value = payload.session_id
    roomCode.value = payload.room_code
    playersCount.value = payload.players_count

    questionIndex.value = payload.question_index ?? null
    totalQuestions.value = payload.total_questions ?? null
    currentQuestion.value = payload.question ?? null
    deadlineAt.value = payload.deadline_at ?? null
    revealUntil.value = payload.reveal_until ?? null
    leaderboardTop.value = payload.leaderboard_top ?? []
  }

  const onJoinedLobby = (payload: JoinedLobbyPayload) => {
    participantId.value = payload.participant_id
    participantToken.value = payload.participant_token
    nickname.value = payload.nickname
    roomCode.value = payload.room_code
    phase.value = payload.status
    persistPlayerAuth()
  }

  const onQuestionOpened = (payload: QuestionOpenedPayload) => {
    phase.value = 'question_open'
    questionIndex.value = payload.question_index
    totalQuestions.value = payload.total_questions
    currentQuestion.value = payload.question
    deadlineAt.value = payload.deadline_at
    revealUntil.value = null
  }

  const onAnswerReveal = (payload: AnswerRevealPayload) => {
    phase.value = 'answer_reveal'
    revealUntil.value = payload.reveal_until
    leaderboardTop.value = payload.leaderboard_top
    myScore.value = payload.total_score
    myRank.value = payload.your_rank
  }

  const onQuestionRevealHost = (payload: QuestionRevealHostPayload) => {
    phase.value = 'answer_reveal'
    revealUntil.value = payload.reveal_until
    leaderboardTop.value = payload.leaderboard_top
    playersCount.value = payload.total_players
  }

  const onSessionFinished = (payload: SessionFinishedPayload) => {
    phase.value = 'finished'
    currentQuestion.value = null
    deadlineAt.value = null
    revealUntil.value = null
    leaderboardTop.value = payload.leaderboard_top
    clearPlayerAuth()
  }

  const onSocketError = (payload: WsErrorPayload) => {
    lastError.value = getErrorMessage(payload)
  }

  const handleMessage = (message: WsEnvelope) => {
    switch (message.type) {
      case 'session_snapshot':
        applySessionSnapshot(message.payload as SessionSnapshotPayload)
        break
      case 'joined_lobby':
        onJoinedLobby(message.payload as JoinedLobbyPayload)
        break
      case 'lobby_updated':
        playersCount.value = Number((message.payload as { players_count?: number }).players_count ?? playersCount.value)
        break
      case 'question_opened':
        onQuestionOpened(message.payload as QuestionOpenedPayload)
        break
      case 'answer_reveal':
        onAnswerReveal(message.payload as AnswerRevealPayload)
        break
      case 'question_reveal_host':
        onQuestionRevealHost(message.payload as QuestionRevealHostPayload)
        break
      case 'session_finished':
        onSessionFinished(message.payload as SessionFinishedPayload)
        break
      case 'error':
        onSocketError(message.payload as WsErrorPayload)
        break
      default:
        break
    }
  }

  return {
    role,
    phase,
    sessionId,
    roomCode,
    participantId,
    participantToken,
    nickname,
    currentQuestion,
    questionIndex,
    totalQuestions,
    deadlineAt,
    revealUntil,
    playersCount,
    leaderboardTop,
    myScore,
    myRank,
    lastError,
    connectionStatus,
    hostConnect,
    playerJoin,
    playerReconnect,
    startGame,
    submitAnswer,
    finishGame,
    disconnect,
    reset,
  }
})
