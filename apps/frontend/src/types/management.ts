export interface OptionTransport {
  id?: string;
  text: string;
  is_correct: boolean;
  order_index: number;
}

export interface QuestionTransport {
  quiz_id: string;
  text: string;
  selection_type: 'single' | 'multiple';
  time_limit_seconds: number;
  order_index: number;
  options: OptionTransport[];
}

export interface QuizSummary {
  id: string;
  title: string;
}