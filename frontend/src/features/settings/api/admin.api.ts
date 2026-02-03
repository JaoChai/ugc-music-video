import { api } from '@/lib/axios'
import type {
  ApiResponse,
  SystemPromptsResponse,
  SystemPrompt,
  UpdateSystemPromptInput,
} from '@/types'

export const adminApi = {
  getSystemPrompts: async (): Promise<SystemPromptsResponse> => {
    const response = await api.get<ApiResponse<SystemPromptsResponse>>(
      '/api/v1/admin/system-prompts'
    )
    if (!response.data.data) {
      throw new Error('Failed to get system prompts')
    }
    return response.data.data
  },

  updateSystemPrompt: async (
    data: UpdateSystemPromptInput
  ): Promise<SystemPrompt> => {
    const response = await api.put<ApiResponse<SystemPrompt>>(
      '/api/v1/admin/system-prompts',
      data
    )
    if (!response.data.data) {
      throw new Error('Failed to update system prompt')
    }
    return response.data.data
  },
}
