import type { QuestionSelectionType } from '~/types/management'

export type OptionDraft = {
  localId: string
  id?: string
  text: string
  is_correct: boolean
  order_index: number
}

export type QuestionDraft = {
  localId: string
  id?: string
  text: string
  selection_type: QuestionSelectionType
  time_limit_seconds: number
  order_index: number
  options: OptionDraft[]
}

export type QuizDraft = {
  id?: string
  title: string
  description: string
  questions: QuestionDraft[]
}
