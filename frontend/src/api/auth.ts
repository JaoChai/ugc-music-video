import { api } from '@/lib/axios'
import type {
  LoginInput,
  LoginResponse,
  RegisterInput,
  RefreshResponse,
  User,
  ApiResponse,
} from './types'

export async function login(input: LoginInput): Promise<LoginResponse> {
  const response = await api.post<ApiResponse<LoginResponse>>(
    '/api/v1/auth/login',
    input
  )
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Login failed')
  }
  return response.data.data
}

export async function register(input: RegisterInput): Promise<User> {
  const response = await api.post<ApiResponse<User>>(
    '/api/v1/auth/register',
    input
  )
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Registration failed')
  }
  return response.data.data
}

export async function refresh(): Promise<RefreshResponse> {
  const response = await api.post<ApiResponse<RefreshResponse>>(
    '/api/v1/auth/refresh'
  )
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Token refresh failed')
  }
  return response.data.data
}

export async function getMe(): Promise<User> {
  const response = await api.get<ApiResponse<User>>('/api/v1/auth/me')
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to get user')
  }
  return response.data.data
}
