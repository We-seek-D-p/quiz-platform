type RequestMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'

type ApiRequestOptions = {
  method?: RequestMethod
  body?: BodyInit | Record<string, unknown> | null
  accessToken?: string | null
  headers?: HeadersInit
}

export class ApiHttpError extends Error {
  status: number
  code?: string

  constructor(message: string, status: number, code?: string) {
    super(message)
    this.name = 'ApiHttpError'
    this.status = status
    this.code = code
  }
}

type ApiErrorPayload = {
  code?: unknown
  message?: unknown
  detail?: unknown
}

function asRecord(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return null
  }

  return value as Record<string, unknown>
}

function getErrorPayload(error: unknown): ApiErrorPayload {
  const source = asRecord(error)
  if (!source) {
    return {}
  }

  const directData = asRecord(source.data)
  if (directData) {
    return directData as ApiErrorPayload
  }

  const response = asRecord(source.response)
  if (response) {
    const responseData = asRecord(response._data)
    if (responseData) {
      return responseData as ApiErrorPayload
    }
  }

  return {}
}

const parseErrorMessage = (error: unknown): string => {
  const data = getErrorPayload(error)

  if (typeof data.message === 'string' && data.message.trim().length > 0) {
    return data.message
  }

  if (typeof data.detail === 'string' && data.detail.trim().length > 0) {
    return data.detail
  }

  const statusMessage = (error as { statusMessage?: string })?.statusMessage
  if (typeof statusMessage === 'string' && statusMessage.trim().length > 0) {
    return statusMessage
  }

  const message = (error as { message?: string })?.message
  if (typeof message === 'string' && message.trim().length > 0) {
    return message
  }

  return 'Request failed'
}

const parseErrorCode = (error: unknown): string | undefined => {
  const data = getErrorPayload(error)
  if (typeof data.code === 'string' && data.code.trim().length > 0) {
    return data.code
  }

  return undefined
}

export const useApiClient = () => {
  const request = async <T>(path: string, options: ApiRequestOptions = {}): Promise<T> => {
    const headers = new Headers(options.headers)

    if (!headers.has('Content-Type') && options.body !== undefined && !(options.body instanceof FormData)) {
      headers.set('Content-Type', 'application/json')
    }

    if (options.accessToken) {
      headers.set('Authorization', `Bearer ${options.accessToken}`)
    }

    try {
      const fetchOptions = {
        method: options.method ?? 'GET',
        headers,
        credentials: 'include',
        ...(options.body !== undefined ? { body: options.body } : {}),
      }

      return await $fetch<T>(path, fetchOptions)
    } catch (error: unknown) {
      const status = (error as { status?: number })?.status ?? 500
      throw new ApiHttpError(parseErrorMessage(error), status, parseErrorCode(error))
    }
  }

  return {
    request,
  }
}
