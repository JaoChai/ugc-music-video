// Common types for the application

export interface User {
  id: string
  email: string
  name: string | null
  avatar?: string
  openrouter_model?: string
  created_at: string
  updated_at: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: ApiError
  meta?: PaginationMeta
}

export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  items: T[]
  page: number
  perPage: number
  totalItems: number
  totalPages: number
}

export interface ApiError {
  code: number
  message: string
  details?: Record<string, string>
}

export interface APIKeysStatus {
  has_openrouter_key: boolean
  has_kie_key: boolean
}

export interface UpdateAPIKeysInput {
  openrouter_api_key?: string
  kie_api_key?: string
}
