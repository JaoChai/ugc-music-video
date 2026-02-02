import { api } from '@/lib/axios'
import type {
  ApiResponse,
  AgentPromptsResponse,
  AgentPrompts,
  AgentType,
  UpdateAgentPromptInput,
} from '@/types'

export const promptsApi = {
  getPrompts: async (): Promise<AgentPromptsResponse> => {
    const response = await api.get<ApiResponse<AgentPromptsResponse>>(
      '/api/v1/auth/prompts'
    )
    if (!response.data.data) {
      throw new Error('Failed to get prompts')
    }
    return response.data.data
  },

  updatePrompt: async (data: UpdateAgentPromptInput): Promise<AgentPrompts> => {
    const response = await api.put<ApiResponse<AgentPrompts>>(
      '/api/v1/auth/prompts',
      data
    )
    if (!response.data.data) {
      throw new Error('Failed to update prompt')
    }
    return response.data.data
  },

  resetPrompt: async (agentType: AgentType): Promise<AgentPrompts> => {
    const response = await api.delete<ApiResponse<AgentPrompts>>(
      `/api/v1/auth/prompts/${agentType}`
    )
    if (!response.data.data) {
      throw new Error('Failed to reset prompt')
    }
    return response.data.data
  },
}
