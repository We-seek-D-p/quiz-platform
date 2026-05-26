import { ApiHttpError, useApiClient } from '~/composables/api/useApiClient'
import { useAuthStore } from '~/stores/auth'
import type {
  QuestionCreate,
  QuestionPublic,
  QuestionUpdate,
  QuizCreate,
  QuizPublic,
  QuizUpdate,
  SessionCreate,
  SessionPublic,
} from '~/types/management'

export const useManagementApi = () => {
  const config = useRuntimeConfig()
  const { request } = useApiClient()
  const authStore = useAuthStore()

  const managementBase = config.public.managementApiBase

  type ManagementRequestBody = BodyInit | Record<string, unknown> | null

  const authorizedRequest = async <T>(
    path: string,
    method: 'GET' | 'POST' | 'PATCH' | 'DELETE',
    body?: ManagementRequestBody,
    headers?: HeadersInit,
  ): Promise<T> => {
    const accessToken = authStore.accessToken
    if (!accessToken) {
      throw new ApiHttpError('Сессия недоступна. Выполните вход снова.', 401)
    }

    try {
      return await request<T>(path, {
        method,
        body,
        accessToken,
        headers,
      })
    } catch (error: unknown) {
      if (!(error instanceof ApiHttpError) || error.status !== 401) {
        throw error
      }

      const refreshed = await authStore.refreshAccessToken()
      if (!refreshed || !authStore.accessToken) {
        throw new ApiHttpError('Сессия истекла. Войдите снова.', 401, 'unauthorized')
      }

      try {
        return await request<T>(path, {
          method,
          body,
          accessToken: authStore.accessToken,
          headers,
        })
      } catch (retryError: unknown) {
        if (retryError instanceof ApiHttpError && retryError.status === 401) {
          authStore.clearSession()
          throw new ApiHttpError('Сессия истекла. Войдите снова.', 401, 'unauthorized')
        }

        throw retryError
      }
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

    createSession: async (payload: SessionCreate): Promise<SessionPublic> => {
      return authorizedRequest<SessionPublic>(`${managementBase}/sessions/`, 'POST', payload, {
        'Idempotency-Key': crypto.randomUUID(),
      })
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

    updateQuestion: async (quizId: string, questionId: string, payload: QuestionUpdate): Promise<QuestionPublic> => {
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
