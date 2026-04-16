import type { QuestionTransport, OptionTransport } from './management';

export interface OptionDraft extends Omit<OptionTransport, 'order'> {
  localId: string;
}

export interface QuestionDraft extends Omit<QuestionTransport, 'options' | 'order'> {
  localId: string;
  options: OptionDraft[];
  isExpanded?: boolean;
}

export interface QuizDraft {
  id?: string;
  title: string;
  description: string;
  questions: QuestionDraft[];
}