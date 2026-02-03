import { useState } from 'react'
import { Shield, ChevronDown, ChevronRight, Save } from 'lucide-react'
import { Button, Textarea, Badge, Spinner } from '@/components/ui'
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from '@/components/ui/Card'
import {
  useSystemPrompts,
  useUpdateSystemPromptMutation,
} from '../hooks/useSystemPrompts'
import type { SystemPrompt, AgentType } from '@/types'

const MAX_PROMPT_LENGTH = 15000

interface PromptEditorProps {
  prompt: SystemPrompt
  title: string
  description: string
}

function PromptEditor({ prompt, title, description }: PromptEditorProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [editValue, setEditValue] = useState<string | null>(null)
  const updateMutation = useUpdateSystemPromptMutation()

  const displayValue = editValue !== null ? editValue : prompt.prompt_content
  const hasChanges =
    editValue !== null && editValue !== prompt.prompt_content

  const handleSave = async () => {
    if (editValue === null) return
    try {
      await updateMutation.mutateAsync({
        prompt_type: prompt.prompt_type as AgentType,
        prompt_content: editValue,
      })
      setEditValue(null)
    } catch {
      // Error handled by mutation
    }
  }

  const handleCancel = () => {
    setEditValue(null)
  }

  return (
    <Card className="overflow-hidden">
      <button
        type="button"
        className="w-full p-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center gap-3">
          {isExpanded ? (
            <ChevronDown className="h-5 w-5 text-gray-500" />
          ) : (
            <ChevronRight className="h-5 w-5 text-gray-500" />
          )}
          <div className="text-left">
            <div className="font-medium text-gray-900">{title}</div>
            <div className="text-sm text-gray-500">{description}</div>
          </div>
        </div>
        <Badge variant="default">System Default</Badge>
      </button>

      {isExpanded && (
        <div className="border-t p-4 space-y-4">
          <Textarea
            value={displayValue}
            onChange={(e) => setEditValue(e.target.value)}
            rows={12}
            className="font-mono text-sm"
          />
          <div className="flex justify-between items-center text-sm text-gray-500">
            <span>
              {displayValue.length} / {MAX_PROMPT_LENGTH.toLocaleString()}{' '}
              ตัวอักษร
            </span>
            {prompt.updated_by && (
              <span>
                อัปเดตล่าสุด:{' '}
                {new Date(prompt.updated_at).toLocaleDateString('th-TH')}
              </span>
            )}
          </div>

          {displayValue.length > MAX_PROMPT_LENGTH && (
            <div className="text-sm text-red-500">
              เกินจำนวนตัวอักษรที่กำหนด
            </div>
          )}

          {displayValue.length < 100 && (
            <div className="text-sm text-red-500">
              Prompt ต้องมีอย่างน้อย 100 ตัวอักษร
            </div>
          )}

          <div className="flex justify-end gap-2 pt-2 border-t">
            {hasChanges && (
              <>
                <Button variant="outline" size="sm" onClick={handleCancel}>
                  ยกเลิก
                </Button>
                <Button
                  size="sm"
                  onClick={handleSave}
                  disabled={
                    updateMutation.isPending ||
                    displayValue.length > MAX_PROMPT_LENGTH ||
                    displayValue.length < 100
                  }
                >
                  <Save className="h-4 w-4 mr-2" />
                  {updateMutation.isPending ? 'กำลังบันทึก...' : 'บันทึก'}
                </Button>
              </>
            )}
          </div>

          {updateMutation.isError && (
            <div className="text-sm text-red-500">
              เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง
            </div>
          )}
        </div>
      )}
    </Card>
  )
}

export function SystemDefaultsSection() {
  const { data, isLoading, isError } = useSystemPrompts()

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <Spinner className="h-6 w-6" />
        </CardContent>
      </Card>
    )
  }

  if (isError || !data) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-red-500">
          ไม่สามารถโหลด System Prompts ได้
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-amber-500" />
          <CardTitle>AI Prompts</CardTitle>
        </div>
        <CardDescription>
          กำหนด system prompt ของ AI agent แต่ละตัวเพื่อปรับแต่งผลลัพธ์ตามที่ต้องการ
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <PromptEditor
          prompt={data.song_concept}
          title="Song Concept Prompt"
          description="Default prompt สำหรับสร้าง concept เพลง"
        />
        <PromptEditor
          prompt={data.song_selector}
          title="Song Selector Prompt"
          description="Default prompt สำหรับเลือกเพลงที่ดีที่สุด"
        />
        <PromptEditor
          prompt={data.image_concept}
          title="Image Concept Prompt"
          description="Default prompt สำหรับสร้าง concept รูปภาพ"
        />
      </CardContent>
    </Card>
  )
}
