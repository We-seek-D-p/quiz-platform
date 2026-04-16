export type QuestionSelectionType = 'single' | 'multiple'

export interface QuizCreate {
  title: string
  description: string
}

export interface QuizUpdate {
  title?: string
  description?: string
}

export interface QuizPublic {
  id: string
  title: string
  description: string
  created_at: string
  updated_at: string
}

export interface OptionCreate {
  text: string
  order_index: number
  is_correct: boolean
}

export interface OptionUpdate {
  id?: string | null
  text?: string
  order_index?: number
  is_correct?: boolean
}

export interface OptionPublic {
  id: string
  text: string
  order_index: number
  is_correct: boolean
}

export interface QuestionCreate {
  quiz_id: string
  text: string
  selection_type: QuestionSelectionType
  time_limit_seconds: number
  order_index: number
  options: OptionCreate[]
}

export interface QuestionUpdate {
  text?: string
  selection_type?: QuestionSelectionType
  time_limit_seconds?: number
  order_index?: number
  options?: OptionUpdate[]
}

export interface QuestionPublic {
  id: string
  quiz_id: string
  text: string
  selection_type: QuestionSelectionType
  time_limit_seconds: number
  order_index: number
  options: OptionPublic[]
}
