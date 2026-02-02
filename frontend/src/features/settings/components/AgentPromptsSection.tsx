import { Bot } from 'lucide-react'
import { useAgentPrompts } from '../hooks/useAgentPrompts'
import { AgentPromptCard } from './AgentPromptCard'
import type { AgentType } from '@/types'

interface AgentConfig {
  type: AgentType
  title: string
  description: string
}

const agentConfigs: AgentConfig[] = [
  {
    type: 'song_concept',
    title: 'Song Concept Agent',
    description: 'วิเคราะห์ concept และสร้าง prompt สำหรับ Suno AI',
  },
  {
    type: 'song_selector',
    title: 'Song Selector Agent',
    description: 'เลือกเพลงที่ดีที่สุดจากตัวเลือกที่ Suno สร้าง',
  },
  {
    type: 'image_concept',
    title: 'Image Concept Agent',
    description: 'สร้าง prompt รูปภาพสำหรับ NanoBanana',
  },
]

export function AgentPromptsSection() {
  const { data, isLoading } = useAgentPrompts()

  const getCustomPrompt = (type: AgentType): string | null => {
    if (!data) return null
    switch (type) {
      case 'song_concept':
        return data.prompts.song_concept_prompt
      case 'song_selector':
        return data.prompts.song_selector_prompt
      case 'image_concept':
        return data.prompts.image_concept_prompt
      default:
        return null
    }
  }

  const getDefaultPrompt = (type: AgentType): string => {
    if (!data) return ''
    switch (type) {
      case 'song_concept':
        return data.defaults.song_concept
      case 'song_selector':
        return data.defaults.song_selector
      case 'image_concept':
        return data.defaults.image_concept
      default:
        return ''
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <Bot className="h-5 w-5 text-gray-500" />
        <h2 className="text-lg font-semibold text-gray-900">AI Prompts</h2>
      </div>
      <p className="text-sm text-gray-500">
        กำหนด system prompt ของ AI agent แต่ละตัวเพื่อปรับแต่งผลลัพธ์ตามที่ต้องการ
      </p>

      <div className="space-y-3">
        {agentConfigs.map((agent) => (
          <AgentPromptCard
            key={agent.type}
            type={agent.type}
            title={agent.title}
            description={agent.description}
            customPrompt={getCustomPrompt(agent.type)}
            defaultPrompt={getDefaultPrompt(agent.type)}
            isLoading={isLoading}
          />
        ))}
      </div>
    </div>
  )
}
