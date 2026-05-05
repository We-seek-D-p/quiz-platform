export type QuestionSelectionType = 'single' | 'multiple'

export type QuizCreate = {
  title: string
  description: string
}

export type QuizUpdate = {
  title?: string
  description?: string
}

export type QuizPublic = {
  id: string
  title: string
  description: string
  created_at: string
  updated_at: string
}

export type OptionCreate = {
  text: string
  order_index: number
  is_correct: boolean
}

export type OptionUpdate = {
  id?: string | null
  text?: string
  order_index?: number
  is_correct?: boolean
}

export type OptionPublic = {
  id: string
  text: string
  order_index: number
  is_correct: boolean
}

export type QuestionCreate = {
  text: string
  selection_type: QuestionSelectionType
  time_limit_seconds: number
  order_index: number
  options: OptionCreate[]
}

export type QuestionUpdate = {
  text?: string
  selection_type?: QuestionSelectionType
  time_limit_seconds?: number
  order_index?: number
  options?: OptionUpdate[]
}

export type QuestionPublic = {
  id: string
  text: string
  selection_type: QuestionSelectionType
  time_limit_seconds: number
  order_index: number
  options: OptionPublic[]
}
