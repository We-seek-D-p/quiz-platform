export interface OptionTransport {
  id?: string;
  text: string;
  isCorrect: boolean;
  order: number;
}

export interface QuestionTransport {
  id?: string;
  text: string;
  type: 'single' | 'multiple';
  order: number;
  options: OptionTransport[];
}

export interface QuizTransport {
  id: string;
  title: string;
  description?: string;
  questions: QuestionTransport[];
  createdAt: string;
  updatedAt: string;
}


export type CreateQuizDto = Omit<QuizTransport, 'id' | 'createdAt' | 'updatedAt'>;
export type UpdateQuizDto = Partial<CreateQuizDto>;