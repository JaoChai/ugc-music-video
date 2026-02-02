import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '@/stores/auth.store'
import { settingsApi, type UpdateProfileInput } from '../api/settings.api'
import type { AxiosError } from 'axios'

interface ApiErrorResponse {
  message?: string
  data?: Record<string, { message: string }>
}

const getErrorMessage = (error: unknown): string => {
  if (!error) return 'An unexpected error occurred'

  const axiosError = error as AxiosError<ApiErrorResponse>

  if (axiosError.response?.data?.message) {
    return axiosError.response.data.message
  }

  if (axiosError.response?.data?.data) {
    const fieldErrors = axiosError.response.data.data
    const firstError = Object.values(fieldErrors)[0]
    if (firstError?.message) {
      return firstError.message
    }
  }

  if (axiosError.response?.status === 400) {
    return 'Invalid input data'
  }

  if (axiosError.response?.status === 401) {
    return 'Not authenticated'
  }

  if (axiosError.response?.status === 403) {
    return 'You do not have permission to update this profile'
  }

  return 'An unexpected error occurred. Please try again.'
}

export function useUpdateProfileMutation() {
  const queryClient = useQueryClient()
  const { setUser } = useAuthStore()

  const mutation = useMutation({
    mutationFn: (data: UpdateProfileInput) => settingsApi.updateProfile(data),
    onSuccess: (updatedUser) => {
      // Update the user in the store
      setUser(updatedUser)
      // Invalidate user queries
      queryClient.invalidateQueries({ queryKey: ['user'] })
    },
  })

  return {
    updateProfile: mutation.mutateAsync,
    isUpdating: mutation.isPending,
    isSuccess: mutation.isSuccess,
    error: mutation.error ? getErrorMessage(mutation.error) : null,
    reset: mutation.reset,
  }
}

export function useTestOpenRouterConnection() {
  const mutation = useMutation({
    mutationFn: () => settingsApi.testOpenRouterConnection(),
  })

  return {
    testConnection: mutation.mutateAsync,
    isTesting: mutation.isPending,
    result: mutation.data,
    error: mutation.error ? getErrorMessage(mutation.error) : null,
    reset: mutation.reset,
  }
}

export function useTestKIEConnection() {
  const mutation = useMutation({
    mutationFn: () => settingsApi.testKIEConnection(),
  })

  return {
    testConnection: mutation.mutateAsync,
    isTesting: mutation.isPending,
    result: mutation.data,
    error: mutation.error ? getErrorMessage(mutation.error) : null,
    reset: mutation.reset,
  }
}
