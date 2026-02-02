import { useState, useEffect } from 'react'
import { Button, Input } from '@/components/ui'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card'
import { useAPIKeysStatus, useUpdateAPIKeysMutation } from '../hooks/useApiKeys'

export function APIKeysSection() {
  const { data: keysStatus, isLoading } = useAPIKeysStatus()
  const { updateAPIKeys, isUpdating, isSuccess, error, reset } = useUpdateAPIKeysMutation()

  const [openRouterKey, setOpenRouterKey] = useState('')
  const [kieKey, setKieKey] = useState('')
  const [showOpenRouterKey, setShowOpenRouterKey] = useState(false)
  const [showKieKey, setShowKieKey] = useState(false)

  // Reset success state after 3 seconds
  useEffect(() => {
    if (isSuccess) {
      const timer = setTimeout(() => {
        reset()
        setOpenRouterKey('')
        setKieKey('')
      }, 3000)
      return () => clearTimeout(timer)
    }
  }, [isSuccess, reset])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const updateData: { openrouter_api_key?: string; kie_api_key?: string } = {}

    if (openRouterKey) {
      updateData.openrouter_api_key = openRouterKey
    }

    if (kieKey) {
      updateData.kie_api_key = kieKey
    }

    if (Object.keys(updateData).length > 0) {
      try {
        await updateAPIKeys(updateData)
      } catch {
        // Error is handled by the hook's error state
      }
    }
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Keys</CardTitle>
          <CardDescription>กำลังโหลด...</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>API Keys</CardTitle>
        <CardDescription>
          กำหนด API keys สำหรับใช้งานบริการภายนอก API keys จะถูกเข้ารหัสและเก็บอย่างปลอดภัย
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Success Message */}
          {isSuccess && (
            <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm">
              บันทึก API keys สำเร็จแล้ว!
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
              {error}
            </div>
          )}

          {/* OpenRouter API Key */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-gray-700">
                OpenRouter API Key
              </label>
              {keysStatus?.has_openrouter_key ? (
                <span className="inline-flex items-center gap-1 text-sm text-green-600">
                  <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  เชื่อมต่อแล้ว
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 text-sm text-amber-600">
                  <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                  </svg>
                  จำเป็นต้องใส่
                </span>
              )}
            </div>
            <div className="relative">
              <Input
                type={showOpenRouterKey ? 'text' : 'password'}
                value={openRouterKey}
                onChange={(e) => setOpenRouterKey(e.target.value)}
                placeholder={keysStatus?.has_openrouter_key ? '••••••••••••' : 'sk-or-v1-...'}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowOpenRouterKey(!showOpenRouterKey)}
                aria-label={showOpenRouterKey ? 'ซ่อน API Key' : 'แสดง API Key'}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                {showOpenRouterKey ? (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                )}
              </button>
            </div>
            <p className="text-sm text-gray-500">
              ใช้สำหรับ LLM (วิเคราะห์คอนเซ็ปต์, สร้างเนื้อเพลง, เลือกเพลง) -{' '}
              <a
                href="https://openrouter.ai/keys"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline"
              >
                รับ API Key
              </a>
            </p>
          </div>

          {/* KIE API Key */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-gray-700">
                KIE API Key (Suno/NanoBanana)
              </label>
              {keysStatus?.has_kie_key ? (
                <span className="inline-flex items-center gap-1 text-sm text-green-600">
                  <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                  </svg>
                  เชื่อมต่อแล้ว
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 text-sm text-amber-600">
                  <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                  </svg>
                  จำเป็นต้องใส่
                </span>
              )}
            </div>
            <div className="relative">
              <Input
                type={showKieKey ? 'text' : 'password'}
                value={kieKey}
                onChange={(e) => setKieKey(e.target.value)}
                placeholder={keysStatus?.has_kie_key ? '••••••••••••' : 'kie-...'}
                className="pr-10"
              />
              <button
                type="button"
                onClick={() => setShowKieKey(!showKieKey)}
                aria-label={showKieKey ? 'ซ่อน API Key' : 'แสดง API Key'}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                {showKieKey ? (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                  </svg>
                ) : (
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                )}
              </button>
            </div>
            <p className="text-sm text-gray-500">
              ใช้สำหรับสร้างเพลง (Suno) และรูปภาพ (NanoBanana) -{' '}
              <a
                href="https://kie.ai"
                target="_blank"
                rel="noopener noreferrer"
                className="text-blue-600 hover:underline"
              >
                รับ API Key
              </a>
            </p>
          </div>

          {/* Save Button */}
          <div className="pt-2">
            <Button
              type="submit"
              isLoading={isUpdating}
              disabled={isUpdating || (!openRouterKey && !kieKey)}
            >
              บันทึก API Keys
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
