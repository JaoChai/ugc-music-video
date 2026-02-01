import { api } from '@/lib/axios'
import type { User } from '@/types'

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
}
