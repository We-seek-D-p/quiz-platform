import { ApiHttpError, useApiClient } from '~/composables/api/useApiClient'
import { useAuthStore } from '~/stores/auth'
import type {
  QuestionCreate,
  QuestionPublic,
  QuestionUpdate,
  QuizCreate,
  QuizPublic,
  QuizUpdate,
} from '~/types/management'

export const useManagementApi = () => {
  const config = useRuntimeConfig()
  const { request } = useApiClient()
  const authStore = useAuthStore()

  const managementBase = config.public.managementApiBase

  const authorizedRequest = async <T>(path: string, method: 'GET' | 'POST' | 'PATCH' | 'DELETE', body?: unknown): Promise<T> => {
    const accessToken = authStore.accessToken
    if (!accessToken) {
      throw new ApiHttpError('Сессия недоступна. Выполните вход снова.', 401)
    }

    try {
      return await request<T>(path, {
        method,
        body,
        accessToken,
      })
    } catch (error: unknown) {
      if (!(error instanceof ApiHttpError) || error.status !== 401) {
        throw error
      }

      const refreshed = await authStore.refreshAccessToken()
      if (!refreshed || !authStore.accessToken) {
        throw error
      }

      return request<T>(path, {
        method,
        body,
        accessToken: authStore.accessToken,
      })
    }
  }

  return {
    getQuizzes: async (): Promise<QuizPublic[]> => {
      return authorizedRequest<QuizPublic[]>(`${managementBase}/quizzes/`, 'GET')
    },

    getQuizQuestions: async (quizId: string): Promise<QuestionPublic[]> => {
      return authorizedRequest<QuestionPublic[]>(`${managementBase}/quizzes/${quizId}/questions/`, 'GET')
    },

    getQuiz: async (quizId: string): Promise<QuizPublic> => {
      return authorizedRequest<QuizPublic>(`${managementBase}/quizzes/${quizId}`, 'GET')
    },

    createQuiz: async (payload: QuizCreate): Promise<QuizPublic> => {
      return authorizedRequest<QuizPublic>(`${managementBase}/quizzes/`, 'POST', payload)
    },

    updateQuiz: async (quizId: string, payload: QuizUpdate): Promise<QuizPublic> => {
      return authorizedRequest<QuizPublic>(`${managementBase}/quizzes/${quizId}`, 'PATCH', payload)
    },

    deleteQuiz: async (quizId: string): Promise<void> => {
      await authorizedRequest<void>(`${managementBase}/quizzes/${quizId}`, 'DELETE')
    },

    createQuestion: async (quizId: string, payload: QuestionCreate): Promise<QuestionPublic> => {
      return authorizedRequest<QuestionPublic>(`${managementBase}/quizzes/${quizId}/questions/`, 'POST', payload)
    },

    updateQuestion: async (
      quizId: string,
      questionId: string,
      payload: QuestionUpdate,
    ): Promise<QuestionPublic> => {
      return authorizedRequest<QuestionPublic>(
        `${managementBase}/quizzes/${quizId}/questions/${questionId}`,
        'PATCH',
        payload,
      )
    },

    deleteQuestion: async (quizId: string, questionId: string): Promise<void> => {
      await authorizedRequest<void>(`${managementBase}/quizzes/${quizId}/questions/${questionId}`, 'DELETE')
    },
  }
}
