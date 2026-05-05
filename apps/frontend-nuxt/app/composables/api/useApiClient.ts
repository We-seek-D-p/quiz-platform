type RequestMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'

type ApiRequestOptions = {
  method?: RequestMethod
  body?: BodyInit | Record<string, unknown> | null
  accessToken?: string | null
  headers?: HeadersInit
}

export class ApiHttpError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiHttpError'
    this.status = status
  }
}

const parseErrorMessage = (error: unknown): string => {
  const data = (error as { data?: { detail?: string } })?.data
  if (typeof data?.detail === 'string' && data.detail.trim().length > 0) {
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

export const useApiClient = () => {
  const request = async <T>(path: string, options: ApiRequestOptions = {}): Promise<T> => {
    const headers = new Headers(options.headers)

    if (
      !headers.has('Content-Type') &&
      options.body !== undefined &&
      !(options.body instanceof FormData)
    ) {
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
      throw new ApiHttpError(parseErrorMessage(error), status)
    }
  }

  return {
    request,
  }
}
