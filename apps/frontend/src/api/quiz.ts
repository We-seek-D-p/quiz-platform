import type {
  QuestionCreate,
  QuestionPublic,
  QuestionUpdate,
  QuizCreate,
  QuizPublic,
  QuizUpdate,
} from '@/types'

const MANAGEMENT_API_PREFIX = '/api/v1'

export type ManagementRequestOptions = {
  accessToken?: string | null
  userId?: string | null
  getAccessToken?: () => string | null
  refreshAccessToken?: () => Promise<boolean>
}

export type ApiErrorPayload = {
  detail?: string
}

const buildRequestInit = (
  init: RequestInit = {},
  options: ManagementRequestOptions = {},
  accessToken?: string | null,
): RequestInit => {
  const headers = new Headers(init.headers)

  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  if (options.userId) {
    headers.set('X-User-Id', options.userId)
  }

  return {
    ...init,
    headers,
    credentials: 'include',
  }
}

const request = async (
  path: string,
  method: 'GET' | 'POST' | 'PATCH' | 'DELETE',
  options: ManagementRequestOptions = {},
  body?: unknown,
): Promise<Response> => {
  const init: RequestInit = {
    method,
    ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
  }

  const resolveAccessToken = (): string | null | undefined => {
    if (options.getAccessToken) {
      return options.getAccessToken()
    }

    return options.accessToken
  }

  const send = async (): Promise<Response> => {
    return fetch(
      `${MANAGEMENT_API_PREFIX}${path}`,
      buildRequestInit(init, options, resolveAccessToken()),
    )
  }

  const response = await send()
  if (response.status !== 401 || !options.refreshAccessToken) {
    return response
  }

  const refreshed = await options.refreshAccessToken()
  if (!refreshed) {
    return response
  }

  return send()
}

export const parseManagementErrorMessage = async (response: Response): Promise<string> => {
  try {
    const payload = (await response.json()) as ApiErrorPayload
    if (typeof payload.detail === 'string' && payload.detail.trim().length > 0) {
      return payload.detail
    }
  } catch {
    return `${response.status} ${response.statusText}`.trim()
  }

  return `${response.status} ${response.statusText}`.trim()
}

export const createQuizRequest = async (
  payload: QuizCreate,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request('/quizzes/', 'POST', options, payload)
}

export const getQuizzesRequest = async (
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request('/quizzes/', 'GET', options)
}

export const getQuizRequest = async (
  quizId: string,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}`, 'GET', options)
}

export const updateQuizRequest = async (
  quizId: string,
  payload: QuizUpdate,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}`, 'PATCH', options, payload)
}

export const deleteQuizRequest = async (
  quizId: string,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}`, 'DELETE', options)
}

export const getQuizQuestionsRequest = async (
  quizId: string,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}/questions/`, 'GET', options)
}

export const getQuestionRequest = async (
  quizId: string,
  questionId: string,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}/questions/${questionId}`, 'GET', options)
}

export const createQuestionRequest = async (
  quizId: string,
  payload: QuestionCreate,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}/questions/`, 'POST', options, payload)
}

export const updateQuestionRequest = async (
  quizId: string,
  questionId: string,
  payload: QuestionUpdate,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}/questions/${questionId}`, 'PATCH', options, payload)
}

export const deleteQuestionRequest = async (
  quizId: string,
  questionId: string,
  options: ManagementRequestOptions,
): Promise<Response> => {
  return request(`/quizzes/${quizId}/questions/${questionId}`, 'DELETE', options)
}

export const parseQuizPublic = async (response: Response): Promise<QuizPublic> => {
  return (await response.json()) as QuizPublic
}

export const parseQuizPublicList = async (response: Response): Promise<QuizPublic[]> => {
  return (await response.json()) as QuizPublic[]
}

export const parseQuestionPublic = async (
  response: Response,
): Promise<QuestionPublic> => {
  return (await response.json()) as QuestionPublic
}

export const parseQuestionPublicList = async (
  response: Response,
): Promise<QuestionPublic[]> => {
  return (await response.json()) as QuestionPublic[]
}
