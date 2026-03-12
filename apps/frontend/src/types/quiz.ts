export type SelectionType = 'single' | 'multiple';

export interface QuizListItem {
    id: string;
    ownerId: string;
    title: string;
    description: string;
    isPublic: boolean;
    questionCount: number;
    createdAt: string;
    updatedAt: string;
}

export interface QuestionOption {
    id: string;
    text: string;
    orderIndex: number;
}

export interface QuizQuestion {
    id: string;
    quizId: string;
    text: string;
    selectionType: SelectionType;
    timeLimitSeconds: number;
    orderIndex: number;
    options: QuestionOption[];
    correctOptionIds: string[];
}

export interface QuizDetails {
    id: string;
    ownerId: string;
    title: string;
    description: string;
    isPublic: boolean;
    questions: QuizQuestion[];
    createdAt: string;
    updatedAt: string;
}

export interface PublicQuizCard {
    id: string;
    title: string;
    description: string;
    authorNickname: string;
    questionCount: number;
    updatedAt: string;
}
