import { api } from '@/lib/axios'
import type { User, APIKeysStatus, UpdateAPIKeysInput } from '@/types'

export interface UpdateProfileInput {
  name?: string
  openrouter_model?: string
}

export interface TestConnectionResponse {
  success: boolean
  message: string
}

export const settingsApi = {
  updateProfile: async (data: UpdateProfileInput): Promise<User> => {
    const response = await api.patch<{ data: User }>('/api/v1/auth/profile', data)
    return response.data.data
  },

  getAPIKeysStatus: async (): Promise<APIKeysStatus> => {
    const response = await api.get<{ data: APIKeysStatus }>('/api/v1/auth/api-keys')
    return response.data.data
  },

  updateAPIKeys: async (data: UpdateAPIKeysInput): Promise<APIKeysStatus> => {
    const response = await api.put<{ data: APIKeysStatus }>('/api/v1/auth/api-keys', data)
    return response.data.data
  },

  deleteAPIKeys: async (): Promise<void> => {
    await api.delete('/api/v1/auth/api-keys')
  },

  testOpenRouterConnection: async (): Promise<TestConnectionResponse> => {
    const response = await api.post<{ data: TestConnectionResponse }>('/api/v1/auth/test-openrouter')
    return response.data.data
  },

  testKIEConnection: async (): Promise<TestConnectionResponse> => {
    const response = await api.post<{ data: TestConnectionResponse }>('/api/v1/auth/test-kie')
    return response.data.data
  },
}
