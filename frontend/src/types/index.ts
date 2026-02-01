// Common types for the application

export interface User {
  id: string
  email: string
  name: string
  avatar?: string
  openrouter_model?: string
  created: string
  updated: string
}

export interface ApiResponse<T> {
  data: T
  message?: string
}

export interface PaginatedResponse<T> {
  items: T[]
  page: number
  perPage: number
  totalItems: number
  totalPages: number
}

export interface ApiError {
  message: string
  code?: string
  data?: Record<string, unknown>
}
