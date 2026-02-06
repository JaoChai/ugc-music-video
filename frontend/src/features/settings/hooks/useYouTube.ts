import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { settingsApi } from '../api/settings.api'
import { apiKeysKeys } from './useApiKeys'

export function useYouTubeConnect() {
  const [isConnecting, setIsConnecting] = useState(false)

  const connect = async () => {
    setIsConnecting(true)
    try {
      const authURL = await settingsApi.getYouTubeConnectURL()
      window.location.href = authURL
    } catch {
      setIsConnecting(false)
      throw new Error('Failed to initiate YouTube connection')
    }
  }

  return { connect, isConnecting }
}

export function useDisconnectYouTube() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: settingsApi.disconnectYouTube,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: apiKeysKeys.all })
    },
  })
}
