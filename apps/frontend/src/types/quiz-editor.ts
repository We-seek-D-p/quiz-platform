export interface OptionDraft {
  localId: string
  id?: string
  text: string
  is_correct: boolean
  order_index: number
}

export interface QuestionDraft {
  localId: string
  id?: string
  text: string
  selection_type: 'single' | 'multiple'
  time_limit_seconds: number
  order_index: number
  options: OptionDraft[]
}

export interface QuizDraft {
  id?: string
  title: string
  description: string
  questions: QuestionDraft[]
}
