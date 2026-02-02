import { useState, useEffect } from 'react'
import { useAuthStore } from '@/stores/auth.store'

/**
 * Hook to safely subscribe to Zustand persist hydration state.
 *
 * Uses the official Zustand v5 pattern with empty dependency array
 * to properly handle hydration state synchronization.
 */
export function useHasHydrated(): boolean {
  const [hydrated, setHydrated] = useState(false)

  useEffect(() => {
    // Subscribe to rehydration start (for manual rehydration cases)
    const unsubHydrate = useAuthStore.persist.onHydrate(() => {
      setHydrated(false)
    })

    // Subscribe to hydration completion
    const unsubFinishHydration = useAuthStore.persist.onFinishHydration(() => {
      setHydrated(true)
    })

    // Set current hydration status AFTER subscribing
    // This handles the case where hydration already completed
    setHydrated(useAuthStore.persist.hasHydrated())

    return () => {
      unsubHydrate()
      unsubFinishHydration()
    }
  }, [])

  return hydrated
}
