import { api } from '@/lib/axios'
import type { User } from '@/types'

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

export interface AuthResponse {
  token: string
  record: User
}

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await api.post('/api/collections/users/auth-with-password', {
      identity: data.email,
      password: data.password,
    })
    return response.data
  },

  register: async (data: RegisterRequest): Promise<User> => {
    const response = await api.post('/api/collections/users/records', {
      email: data.email,
      password: data.password,
      passwordConfirm: data.passwordConfirm,
      name: data.name || '',
    })
    return response.data
  },

  getCurrentUser: async (): Promise<User> => {
    const response = await api.get('/api/collections/users/auth-refresh')
    return response.data.record
  },

  logout: async (): Promise<void> => {
    // PocketBase doesn't have a logout endpoint, we just clear local storage
    localStorage.removeItem('auth_token')
  },
}
