import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Button } from '@/components/ui'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card'
import { useAPIKeysStatus } from '../hooks/useApiKeys'
import { useYouTubeConnect, useDisconnectYouTube } from '../hooks/useYouTube'

export function YouTubeSection() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  const { data: apiKeysStatus } = useAPIKeysStatus()
  const { connect, isConnecting } = useYouTubeConnect()
  const disconnect = useDisconnectYouTube()

  const isConnected = apiKeysStatus?.has_youtube ?? false

  // Handle OAuth callback redirect params
  useEffect(() => {
    const youtubeParam = searchParams.get('youtube')
    if (youtubeParam === 'connected') {
      setMessage({ type: 'success', text: 'เชื่อมต่อ YouTube สำเร็จแล้ว!' })
      searchParams.delete('youtube')
      setSearchParams(searchParams, { replace: true })
    } else if (youtubeParam === 'error') {
      const reason = searchParams.get('reason') || 'unknown'
      setMessage({ type: 'error', text: `เชื่อมต่อ YouTube ล้มเหลว: ${reason}` })
      searchParams.delete('youtube')
      searchParams.delete('reason')
      setSearchParams(searchParams, { replace: true })
    }
  }, [searchParams, setSearchParams])

  // Clear message after 5 seconds
  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => setMessage(null), 5000)
      return () => clearTimeout(timer)
    }
  }, [message])

  const handleConnect = async () => {
    try {
      await connect()
    } catch {
      setMessage({ type: 'error', text: 'ไม่สามารถเริ่มการเชื่อมต่อ YouTube ได้' })
    }
  }

  const handleDisconnect = async () => {
    if (!confirm('คุณแน่ใจหรือไม่ว่าต้องการยกเลิกการเชื่อมต่อ YouTube?')) return
    try {
      await disconnect.mutateAsync()
      setMessage({ type: 'success', text: 'ยกเลิกการเชื่อมต่อ YouTube สำเร็จแล้ว' })
    } catch {
      setMessage({ type: 'error', text: 'ยกเลิกการเชื่อมต่อ YouTube ล้มเหลว' })
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>YouTube</CardTitle>
        <CardDescription>
          เชื่อมต่อ YouTube เพื่ออัปโหลดวิดีโอโดยอัตโนมัติ (Unlisted)
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {message && (
          <div
            className={`px-4 py-3 rounded-lg text-sm ${
              message.type === 'success'
                ? 'bg-green-50 border border-green-200 text-green-700'
                : 'bg-red-50 border border-red-200 text-red-700'
            }`}
          >
            {message.text}
          </div>
        )}

        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm font-medium text-gray-900">
              สถานะ: {isConnected ? 'เชื่อมต่อแล้ว' : 'ยังไม่เชื่อมต่อ'}
            </p>
            <p className="text-xs text-gray-500 mt-1">
              {isConnected
                ? 'วิดีโอใหม่จะถูกอัปโหลดไป YouTube โดยอัตโนมัติ'
                : 'เชื่อมต่อเพื่อเปิดใช้งานการอัปโหลดอัตโนมัติ'}
            </p>
          </div>

          {isConnected ? (
            <Button
              variant="destructive"
              onClick={handleDisconnect}
              isLoading={disconnect.isPending}
              disabled={disconnect.isPending}
            >
              ยกเลิกการเชื่อมต่อ
            </Button>
          ) : (
            <Button
              onClick={handleConnect}
              isLoading={isConnecting}
              disabled={isConnecting}
            >
              เชื่อมต่อ YouTube
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
