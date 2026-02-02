import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { promptsApi } from '../api/prompts.api'
import type { AgentType, UpdateAgentPromptInput } from '@/types'

const promptKeys = {
  all: ['prompts'] as const,
}

export function useAgentPrompts() {
  return useQuery({
    queryKey: promptKeys.all,
    queryFn: promptsApi.getPrompts,
  })
}

export function useUpdatePromptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateAgentPromptInput) => promptsApi.updatePrompt(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: promptKeys.all })
    },
  })
}

export function useResetPromptMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (agentType: AgentType) => promptsApi.resetPrompt(agentType),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: promptKeys.all })
    },
  })
}
