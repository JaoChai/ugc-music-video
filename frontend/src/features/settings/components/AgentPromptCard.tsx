import { useState, useMemo } from 'react'
import { ChevronDown, ChevronRight, RotateCcw } from 'lucide-react'
import { Button, Card, Textarea, Badge, Spinner } from '@/components/ui'
import {
  useUpdatePromptMutation,
  useResetPromptMutation,
} from '../hooks/useAgentPrompts'
import type { AgentType } from '@/types'

interface AgentPromptCardProps {
  type: AgentType
  title: string
  description: string
  customPrompt: string | null
  defaultPrompt: string
  isLoading?: boolean
}

const MAX_PROMPT_LENGTH = 10000

export function AgentPromptCard({
  type,
  title,
  description,
  customPrompt,
  defaultPrompt,
  isLoading = false,
}: AgentPromptCardProps) {
  const [isExpanded, setIsExpanded] = useState(false)
  const [showDefault, setShowDefault] = useState(false)
  const [editValue, setEditValue] = useState<string | null>(null)

  const updateMutation = useUpdatePromptMutation()
  const resetMutation = useResetPromptMutation()

  const isCustom = customPrompt !== null && customPrompt !== ''
  const currentPrompt = isCustom ? customPrompt : defaultPrompt

  // Use edited value if user has started editing, otherwise use current prompt
  const displayValue = editValue !== null ? editValue : currentPrompt
  const hasChanges = useMemo(
    () => editValue !== null && editValue !== currentPrompt,
    [editValue, currentPrompt]
  )

  const handleSave = async () => {
    if (editValue === null) return
    try {
      await updateMutation.mutateAsync({
        agent_type: type,
        prompt: editValue,
      })
      setEditValue(null) // Reset to use new server value
    } catch {
      // Error handled by mutation
    }
  }

  const handleReset = async () => {
    try {
      await resetMutation.mutateAsync(type)
      setShowDefault(false)
      setEditValue(null) // Reset to use new server value
    } catch {
      // Error handled by mutation
    }
  }

  const handleCancel = () => {
    setEditValue(null) // Discard changes, use server value
  }

  const handleChange = (value: string) => {
    setEditValue(value)
  }

  if (isLoading) {
    return (
      <Card className="p-4">
        <div className="flex items-center gap-3">
          <Spinner className="h-5 w-5" />
          <span className="text-gray-500">Loading...</span>
        </div>
      </Card>
    )
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
        <Badge variant={isCustom ? 'default' : 'secondary'}>
          {isCustom ? 'กำหนดเอง' : 'ค่าเริ่มต้น'}
        </Badge>
      </button>

      {isExpanded && (
        <div className="border-t p-4 space-y-4">
          <div>
            <Textarea
              value={showDefault ? defaultPrompt : displayValue}
              onChange={(e) => handleChange(e.target.value)}
              disabled={showDefault}
              rows={10}
              className="font-mono text-sm"
              placeholder="ใส่ system prompt ที่ต้องการ..."
            />
            <div className="flex justify-between items-center mt-2 text-sm text-gray-500">
              <span>
                {showDefault ? defaultPrompt.length : displayValue.length} /{' '}
                {MAX_PROMPT_LENGTH.toLocaleString()} ตัวอักษร
              </span>
              {displayValue.length > MAX_PROMPT_LENGTH && (
                <span className="text-red-500">เกินจำนวนที่กำหนด</span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
              <input
                type="checkbox"
                checked={showDefault}
                onChange={(e) => setShowDefault(e.target.checked)}
                className="rounded border-gray-300"
              />
              ดู prompt เริ่มต้น
            </label>
          </div>

          <div className="flex justify-between items-center pt-2 border-t">
            <div>
              {isCustom && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleReset}
                  disabled={resetMutation.isPending}
                >
                  <RotateCcw className="h-4 w-4 mr-2" />
                  {resetMutation.isPending ? 'กำลังรีเซ็ต...' : 'รีเซ็ตเป็นค่าเริ่มต้น'}
                </Button>
              )}
            </div>
            <div className="flex gap-2">
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
                      displayValue.length > MAX_PROMPT_LENGTH
                    }
                  >
                    {updateMutation.isPending ? 'กำลังบันทึก...' : 'บันทึก'}
                  </Button>
                </>
              )}
            </div>
          </div>

          {(updateMutation.isError || resetMutation.isError) && (
            <div className="text-sm text-red-500">
              เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง
            </div>
          )}
        </div>
      )}
    </Card>
  )
}
