import { defineStore } from 'pinia'
import { useSessionWs } from '~/composables/session/useSessionWs'
import type {
  AnswerAcceptedPayload,
  AnswerRevealPayload,
  ConnectionStatus,
  JoinedLobbyPayload,
  LeaderboardEntryView,
  LobbyUpdatedPayload,
  QuestionOpenedPayload,
  QuestionProgressPayload,
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
const PLAYER_NICKNAME_STORAGE_KEY = 'quiz:nickname'

const ERROR_MESSAGE_DICTIONARY: Record<string, string> = {
  unauthorized: 'Недостаточно прав для выполнения действия',
  forbidden: 'Доступ запрещен',
  session_not_found: 'Сессия не найдена',
  room_not_found: 'Комната не найдена',
  nickname_taken: 'Ник уже занят в этой комнате',
  participant_not_found: 'Игрок не найден',
  invalid_participant_token: 'Токен переподключения недействителен',
  invalid_state_transition: 'Сейчас это действие недоступно',
  game_already_started: 'Игра уже запущена',
  game_already_finished: 'Игра уже завершена',
  question_not_active: 'Время ответа на вопрос истекло',
  answer_already_submitted: 'Ответ уже отправлен',
  invalid_answer_payload: 'Некорректный формат ответа',
  option_not_found: 'Вариант ответа не найден',
  option_not_in_question: 'Вариант не принадлежит текущему вопросу',
  selection_count_invalid: 'Выбрано неверное количество вариантов',
  internal_error: 'Внутренняя ошибка сервиса',
}

const REDIRECT_TO_JOIN_ERROR_CODES = new Set<string>([
  'room_not_found',
  'participant_not_found',
  'invalid_participant_token',
  'session_not_found',
])

type ReconnectContext = {
  roomCode: string
  participantToken: string
}

function getErrorMessage(payload: WsErrorPayload): string {
  if (ERROR_MESSAGE_DICTIONARY[payload.code]) {
    return ERROR_MESSAGE_DICTIONARY[payload.code]
  }

  if (payload.message.trim().length > 0) {
    return payload.message
  }

  return 'Неизвестная ошибка сессии'
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
  const answeredCount = ref<number | null>(null)
  const totalPlayers = ref<number | null>(null)

  const leaderboardTop = ref<LeaderboardEntryView[]>([])
  const myScore = ref<number | null>(null)
  const myRank = ref<number | null>(null)

  const selectedOptionIds = ref<string[]>([])
  const hasSubmittedAnswer = ref(false)
  const isSubmittingAnswer = ref(false)
  const answerSubmitError = ref<string | null>(null)
  const lastAnswerAcceptedAt = ref<string | null>(null)

  const lastError = ref<string | null>(null)
  const lastErrorCode = ref<string | null>(null)
  const shouldReturnToJoin = ref(false)

  const connectionStatus = ref<ConnectionStatus>('idle')
  const reconnectNotice = ref<string | null>(null)

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

  const isConnected = computed(() => connectionStatus.value === 'connected')
  const isReconnecting = computed(() => connectionStatus.value === 'reconnecting')
  const isRuntimePhase = computed(() => phase.value !== 'lobby')
  const currentQuestionNumber = computed(() => {
    if (questionIndex.value === null) {
      return null
    }

    return questionIndex.value + 1
  })

  const canSubmitAnswer = computed(() => {
    if (role.value !== 'player' || phase.value !== 'question_open') {
      return false
    }

    if (!currentQuestion.value) {
      return false
    }

    if (isSubmittingAnswer.value || hasSubmittedAnswer.value) {
      return false
    }

    if (currentQuestion.value.selection_type === 'single') {
      return selectedOptionIds.value.length === 1
    }

    return selectedOptionIds.value.length > 0
  })

  const getReconnectContext = (): ReconnectContext | null => {
    if (roomCode.value && participantToken.value) {
      return {
        roomCode: roomCode.value,
        participantToken: participantToken.value,
      }
    }

    if (!import.meta.client) {
      return null
    }

    const storedRoomCode = localStorage.getItem(PLAYER_ROOM_CODE_STORAGE_KEY)
    const storedToken = localStorage.getItem(PLAYER_TOKEN_STORAGE_KEY)

    if (!storedRoomCode || !storedToken) {
      return null
    }

    return {
      roomCode: storedRoomCode,
      participantToken: storedToken,
    }
  }

  const clearAnswerUi = () => {
    selectedOptionIds.value = []
    hasSubmittedAnswer.value = false
    isSubmittingAnswer.value = false
    answerSubmitError.value = null
    lastAnswerAcceptedAt.value = null
  }

  const clearRuntimeView = () => {
    currentQuestion.value = null
    questionIndex.value = null
    totalQuestions.value = null
    deadlineAt.value = null
    revealUntil.value = null
    answeredCount.value = null
    totalPlayers.value = null
    leaderboardTop.value = []
    myScore.value = null
    myRank.value = null
    clearAnswerUi()
  }

  const clearSessionErrors = () => {
    lastError.value = null
    lastErrorCode.value = null
    shouldReturnToJoin.value = false
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
    playersCount.value = 0
    connectionStatus.value = 'idle'
    reconnectNotice.value = null
    clearRuntimeView()
    clearSessionErrors()
  }

  const persistPlayerAuth = () => {
    if (!import.meta.client || !participantToken.value || !roomCode.value || !participantId.value) {
      return
    }

    localStorage.setItem(PLAYER_TOKEN_STORAGE_KEY, participantToken.value)
    localStorage.setItem(PLAYER_ROOM_CODE_STORAGE_KEY, roomCode.value)
    localStorage.setItem(PLAYER_PARTICIPANT_ID_STORAGE_KEY, participantId.value)

    if (nickname.value) {
      localStorage.setItem(PLAYER_NICKNAME_STORAGE_KEY, nickname.value)
    }
  }

  const clearPlayerAuth = () => {
    if (!import.meta.client) {
      return
    }

    localStorage.removeItem(PLAYER_TOKEN_STORAGE_KEY)
    localStorage.removeItem(PLAYER_ROOM_CODE_STORAGE_KEY)
    localStorage.removeItem(PLAYER_PARTICIPANT_ID_STORAGE_KEY)
    localStorage.removeItem(PLAYER_NICKNAME_STORAGE_KEY)
  }

  const restorePlayerNickname = () => {
    if (!import.meta.client || nickname.value) {
      return
    }

    const persistedNickname = localStorage.getItem(PLAYER_NICKNAME_STORAGE_KEY)
    if (persistedNickname && persistedNickname.trim().length > 0) {
      nickname.value = persistedNickname.trim()
    }
  }

  const sendHostConnect = () => {
    if (!sessionId.value) {
      throw new Error('Missing session id')
    }

    hostWs.send('host_connect', { session_id: sessionId.value })
  }

  const sendPlayerReconnect = () => {
    const reconnectContext = getReconnectContext()
    if (!reconnectContext) {
      throw new Error('Reconnect context is not available')
    }

    roomCode.value = reconnectContext.roomCode
    participantToken.value = reconnectContext.participantToken

    playerWs.send('player_reconnect', {
      room_code: reconnectContext.roomCode,
      participant_token: reconnectContext.participantToken,
    })
  }

  const hostConnect = async (targetSessionId: string) => {
    role.value = 'host'
    sessionId.value = targetSessionId
    clearSessionErrors()
    reconnectNotice.value = null

    await hostWs.connect()
    sendHostConnect()
  }

  const playerJoin = async (targetRoomCode: string, targetNickname: string) => {
    role.value = 'player'
    roomCode.value = targetRoomCode
    nickname.value = targetNickname
    clearSessionErrors()
    reconnectNotice.value = null

    await playerWs.connect()
    playerWs.send('player_join', {
      room_code: targetRoomCode,
      nickname: targetNickname,
    })
  }

  const playerReconnect = async (targetRoomCode?: string, targetToken?: string) => {
    role.value = 'player'
    clearSessionErrors()
    reconnectNotice.value = null

    const storedReconnectContext = getReconnectContext()

    const reconnectContext: ReconnectContext | null =
      targetRoomCode || targetToken
        ? {
            roomCode: targetRoomCode ?? storedReconnectContext?.roomCode ?? '',
            participantToken: targetToken ?? storedReconnectContext?.participantToken ?? '',
          }
        : storedReconnectContext

    if (!reconnectContext || !reconnectContext.roomCode || !reconnectContext.participantToken) {
      throw new Error('Reconnect context is not available')
    }

    roomCode.value = reconnectContext.roomCode
    participantToken.value = reconnectContext.participantToken
    restorePlayerNickname()

    await playerWs.connect()
    sendPlayerReconnect()
  }

  const startGame = () => {
    if (!sessionId.value) {
      throw new Error('Missing session id')
    }

    hostWs.send('start_game', { session_id: sessionId.value })
  }

  const finishGame = () => {
    if (!sessionId.value) {
      throw new Error('Missing session id')
    }

    hostWs.send('finish_game', { session_id: sessionId.value })
  }

  const replaceSelectedOptions = (optionIds: string[]) => {
    const question = currentQuestion.value
    if (!question || hasSubmittedAnswer.value) {
      return
    }

    const uniqueIds = [...new Set(optionIds)]
    if (question.selection_type === 'single') {
      selectedOptionIds.value = uniqueIds.slice(0, 1)
      return
    }

    selectedOptionIds.value = uniqueIds
  }

  const toggleSelectedOption = (optionId: string) => {
    const question = currentQuestion.value
    if (!question || hasSubmittedAnswer.value) {
      return
    }

    if (question.selection_type === 'single') {
      selectedOptionIds.value = [optionId]
      return
    }

    if (selectedOptionIds.value.includes(optionId)) {
      selectedOptionIds.value = selectedOptionIds.value.filter((id) => id !== optionId)
      return
    }

    selectedOptionIds.value = [...selectedOptionIds.value, optionId]
  }

  const submitCurrentAnswer = () => {
    if (!currentQuestion.value) {
      throw new Error('No active question')
    }

    if (!canSubmitAnswer.value) {
      return
    }

    answerSubmitError.value = null
    isSubmittingAnswer.value = true

    playerWs.send('submit_answer', {
      question_id: currentQuestion.value.id,
      selected_option_ids: selectedOptionIds.value,
    })
  }

  const disconnect = () => {
    activeWs.value?.disconnect()
    connectionStatus.value = 'disconnected'
  }

  const onConnected = (wsRole: SessionRole, previousStatus: ConnectionStatus) => {
    if (previousStatus !== 'reconnecting' && previousStatus !== 'disconnected') {
      return
    }

    if (role.value !== wsRole) {
      return
    }

    try {
      if (wsRole === 'host') {
        sendHostConnect()
      } else {
        sendPlayerReconnect()
      }

      reconnectNotice.value = 'Соединение восстановлено'
    } catch (error) {
      if (error instanceof Error) {
        lastError.value = error.message
      }
    }
  }

  watch(
    () => hostWs.status.value,
    (status, previousStatus) => {
      if (role.value === 'host') {
        connectionStatus.value = status
      }

      if (status === 'connected') {
        onConnected('host', previousStatus)
      }
    },
  )

  watch(
    () => playerWs.status.value,
    (status, previousStatus) => {
      if (role.value === 'player') {
        connectionStatus.value = status
      }

      if (status === 'connected') {
        onConnected('player', previousStatus)
      }
    },
  )

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

    if (payload.status !== 'question_open') {
      clearAnswerUi()
    }

    if (payload.status === 'finished') {
      clearPlayerAuth()
    }
  }

  const onJoinedLobby = (payload: JoinedLobbyPayload) => {
    participantId.value = payload.participant_id
    participantToken.value = payload.participant_token
    nickname.value = payload.nickname
    roomCode.value = payload.room_code
    phase.value = payload.status
    shouldReturnToJoin.value = false
    persistPlayerAuth()
  }

  const onLobbyUpdated = (payload: LobbyUpdatedPayload) => {
    playersCount.value = payload.players_count
  }

  const onQuestionOpened = (payload: QuestionOpenedPayload) => {
    phase.value = 'question_open'
    questionIndex.value = payload.question_index
    totalQuestions.value = payload.total_questions
    currentQuestion.value = payload.question
    deadlineAt.value = payload.deadline_at
    revealUntil.value = null
    answeredCount.value = null
    totalPlayers.value = null
    clearAnswerUi()
  }

  const onQuestionProgress = (payload: QuestionProgressPayload) => {
    answeredCount.value = payload.answered_count
    totalPlayers.value = payload.total_players
  }

  const onAnswerAccepted = (payload: AnswerAcceptedPayload) => {
    hasSubmittedAnswer.value = true
    isSubmittingAnswer.value = false
    answerSubmitError.value = null
    lastAnswerAcceptedAt.value = payload.accepted_at
  }

  const onAnswerReveal = (payload: AnswerRevealPayload) => {
    phase.value = 'answer_reveal'
    revealUntil.value = payload.reveal_until
    leaderboardTop.value = payload.leaderboard_top
    myScore.value = payload.total_score
    myRank.value = payload.your_rank
    hasSubmittedAnswer.value = true
    isSubmittingAnswer.value = false
    selectedOptionIds.value = payload.your_selected_option_ids
  }

  const onQuestionRevealHost = (payload: QuestionRevealHostPayload) => {
    phase.value = 'answer_reveal'
    revealUntil.value = payload.reveal_until
    leaderboardTop.value = payload.leaderboard_top
    answeredCount.value = payload.answered_count
    totalPlayers.value = payload.total_players
    playersCount.value = payload.total_players
    clearAnswerUi()
  }

  const onSessionFinished = (payload: SessionFinishedPayload) => {
    phase.value = 'finished'
    currentQuestion.value = null
    deadlineAt.value = null
    revealUntil.value = null
    leaderboardTop.value = payload.leaderboard_top
    clearAnswerUi()
    clearPlayerAuth()
  }

  const onSocketError = (payload: WsErrorPayload) => {
    const message = getErrorMessage(payload)
    lastErrorCode.value = payload.code
    lastError.value = message

    if (payload.code === 'answer_already_submitted') {
      hasSubmittedAnswer.value = true
      isSubmittingAnswer.value = false
      answerSubmitError.value = null
      return
    }

    if (payload.code === 'question_not_active') {
      hasSubmittedAnswer.value = true
      isSubmittingAnswer.value = false
      answerSubmitError.value = message
      return
    }

    if (payload.code === 'invalid_answer_payload' || payload.code === 'selection_count_invalid') {
      isSubmittingAnswer.value = false
      answerSubmitError.value = message
      return
    }

    if (REDIRECT_TO_JOIN_ERROR_CODES.has(payload.code) && role.value === 'player') {
      shouldReturnToJoin.value = true
      participantId.value = null
      participantToken.value = null
      clearPlayerAuth()
    }

    isSubmittingAnswer.value = false
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
        onLobbyUpdated(message.payload as LobbyUpdatedPayload)
        break
      case 'question_opened':
        onQuestionOpened(message.payload as QuestionOpenedPayload)
        break
      case 'question_progress':
        onQuestionProgress(message.payload as QuestionProgressPayload)
        break
      case 'answer_accepted':
        onAnswerAccepted(message.payload as AnswerAcceptedPayload)
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
    answeredCount,
    totalPlayers,
    leaderboardTop,
    myScore,
    myRank,
    selectedOptionIds,
    hasSubmittedAnswer,
    isSubmittingAnswer,
    answerSubmitError,
    lastAnswerAcceptedAt,
    lastError,
    lastErrorCode,
    shouldReturnToJoin,
    connectionStatus,
    reconnectNotice,
    isConnected,
    isReconnecting,
    isRuntimePhase,
    currentQuestionNumber,
    canSubmitAnswer,
    hostConnect,
    playerJoin,
    playerReconnect,
    startGame,
    finishGame,
    replaceSelectedOptions,
    toggleSelectedOption,
    submitCurrentAnswer,
    disconnect,
    clearSessionErrors,
    clearAnswerUi,
    reset,
  }
})
