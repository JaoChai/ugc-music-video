import { useState, useEffect } from 'react'
import { useAuthStore } from '@/stores/auth.store'

/**
 * Hook to safely subscribe to Zustand persist hydration state.
 *
 * Uses useState + useEffect pattern to handle the race condition where
 * hydration may complete between initial render and effect execution.
 */
export function useHasHydrated(): boolean {
  const [hasHydrated, setHasHydrated] = useState(
    useAuthStore.persist.hasHydrated()
  )

  useEffect(() => {
    // Skip subscription if already hydrated
    if (hasHydrated) return

    const unsubscribe = useAuthStore.persist.onFinishHydration(() => {
      setHasHydrated(true)
    })

    // Handle race condition: hydration may have completed between
    // useState initialization and this effect running.
    // Use queueMicrotask to avoid synchronous setState in effect body.
    if (useAuthStore.persist.hasHydrated()) {
      queueMicrotask(() => setHasHydrated(true))
    }

    return unsubscribe
  }, [hasHydrated])

  return hasHydrated
}
