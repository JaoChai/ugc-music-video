import { api } from '@/lib/axios'
import type { User, APIKeysStatus, UpdateAPIKeysInput } from '@/types'

export interface UpdateProfileInput {
  name?: string
  openrouter_model?: string
}

export const settingsApi = {
  updateProfile: async (data: UpdateProfileInput): Promise<User> => {
    // Get current user ID from auth storage
    const authStorage = localStorage.getItem('auth-storage')
    if (!authStorage) {
      throw new Error('Not authenticated')
    }

    const { state } = JSON.parse(authStorage)
    const userId = state?.user?.id

    if (!userId) {
      throw new Error('User ID not found')
    }

    const response = await api.patch(`/api/collections/users/records/${userId}`, data)
    return response.data
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
}
