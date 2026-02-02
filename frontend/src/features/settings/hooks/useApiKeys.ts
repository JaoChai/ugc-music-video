import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import axios, { AxiosError } from 'axios'
import { settingsApi } from '../api/settings.api'
import type { UpdateAPIKeysInput } from '@/types'

interface ApiErrorResponse {
  message?: string
}

export const apiKeysKeys = {
  all: ['api-keys'] as const,
  status: () => [...apiKeysKeys.all, 'status'] as const,
}

export function useAPIKeysStatus() {
  return useQuery({
    queryKey: apiKeysKeys.status(),
    queryFn: settingsApi.getAPIKeysStatus,
  })
}

export function useUpdateAPIKeysMutation() {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (data: UpdateAPIKeysInput) => settingsApi.updateAPIKeys(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: apiKeysKeys.all })
    },
  })

  const getErrorMessage = (error: unknown): string => {
    if (!error) return 'เกิดข้อผิดพลาด'

    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError<ApiErrorResponse>

      if (axiosError.response?.data?.message) {
        return axiosError.response.data.message
      }

      if (axiosError.response?.status === 400) {
        return 'ข้อมูลไม่ถูกต้อง'
      }

      if (axiosError.response?.status === 401) {
        return 'กรุณาเข้าสู่ระบบใหม่'
      }
    }

    if (error instanceof Error) {
      return error.message
    }

    return 'เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง'
  }

  return {
    updateAPIKeys: mutation.mutateAsync,
    isUpdating: mutation.isPending,
    isSuccess: mutation.isSuccess,
    error: mutation.error ? getErrorMessage(mutation.error) : null,
    reset: mutation.reset,
  }
}

export function useDeleteAPIKeysMutation() {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => settingsApi.deleteAPIKeys(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: apiKeysKeys.all })
    },
  })

  return {
    deleteAPIKeys: mutation.mutateAsync,
    isDeleting: mutation.isPending,
  }
}
