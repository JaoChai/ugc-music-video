import { api } from '@/lib/axios'
import type { User, ApiResponse } from '@/types'

export interface LoginResponse {
  token: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  passwordConfirm: string
  name?: string
}

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await api.post<ApiResponse<LoginResponse>>('/api/v1/auth/login', {
      email: data.email,
      password: data.password,
    })
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Login failed')
    }
    return response.data.data
  },

  register: async (data: RegisterRequest): Promise<User> => {
    const response = await api.post<ApiResponse<User>>('/api/v1/auth/register', {
      email: data.email,
      password: data.password,
      name: data.name || undefined,
    })
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Registration failed')
    }
    return response.data.data
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await api.get<ApiResponse<User>>('/api/v1/auth/me')
    if (!response.data.success || !response.data.data) {
      throw new Error(response.data.error?.message || 'Failed to get user')
    }
    return response.data.data
  },

  logout: async (): Promise<void> => {
    localStorage.removeItem('auth_token')
  },
}
