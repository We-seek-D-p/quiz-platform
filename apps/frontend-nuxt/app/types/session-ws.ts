export type SessionPhase = 'lobby' | 'question_open' | 'answer_reveal' | 'finished'

export type SessionRole = 'host' | 'player'

export type SessionWsMode = SessionRole

export type ConnectionStatus = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'disconnected'

export interface WsEnvelope<TPayload = unknown> {
  type: string
  payload: TPayload
}

export interface QuizOptionView {
  id: string
  text: string
}

export interface QuizQuestionView {
  id: string
  text: string
  selection_type: 'single' | 'multiple'
  options: QuizOptionView[]
}

export interface LeaderboardEntryView {
  nickname: string
  score: number
  rank: number
}

export interface SessionSnapshotPayload {
  status: SessionPhase
  session_id: string
  room_code: string
  players_count: number
  question_index?: number
  total_questions?: number
  question?: QuizQuestionView
  deadline_at?: string
  reveal_until?: string
  leaderboard_top?: LeaderboardEntryView[]
}

export interface JoinedLobbyPayload {
  participant_id: string
  participant_token: string
  nickname: string
  room_code: string
  status: SessionPhase
}

export interface QuestionOpenedPayload {
  question_index: number
  total_questions: number
  question: QuizQuestionView
  deadline_at: string
}

export interface AnswerRevealPayload {
  question_id: string
  correct_option_ids: string[]
  your_selected_option_ids: string[]
  your_result: 'correct' | 'wrong'
  score_delta: number
  total_score: number
  your_rank: number
  leaderboard_top: LeaderboardEntryView[]
  reveal_duration_sec: number
  reveal_until: string
}

export interface QuestionRevealHostPayload {
  question_id: string
  correct_option_ids: string[]
  answered_count: number
  total_players: number
  leaderboard_top: LeaderboardEntryView[]
  reveal_duration_sec: number
  reveal_until: string
}

export interface SessionFinishedPayload {
  leaderboard_top: LeaderboardEntryView[]
}

export interface WsErrorPayload {
  code: string
  message: string
}
