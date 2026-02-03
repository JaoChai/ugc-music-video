import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { adminApi } from '../api/admin.api'
import type { UpdateSystemPromptInput } from '@/types'

export const systemPromptKeys = {
  all: ['system-prompts'] as const,
}

export function useSystemPrompts() {
  return useQuery({
    queryKey: systemPromptKeys.all,
    queryFn: adminApi.getSystemPrompts,
  })
}

export function useUpdateSystemPromptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateSystemPromptInput) =>
      adminApi.updateSystemPrompt(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: systemPromptKeys.all })
      // Also invalidate user prompts since defaults may have changed
      queryClient.invalidateQueries({ queryKey: ['prompts'] })
    },
  })
}
