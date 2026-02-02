import { useSyncExternalStore } from 'react'
import { useAuthStore } from '@/stores/auth.store'

/**
 * Hook to safely subscribe to Zustand persist hydration state.
 * Uses useSyncExternalStore to ensure proper React lifecycle integration.
 *
 * This solves the issue where onRehydrateStorage setState calls happen
 * outside React's lifecycle and don't trigger re-renders.
 */
export function useHasHydrated(): boolean {
  return useSyncExternalStore(
    // subscribe: called when component mounts, returns unsubscribe function
    (onStoreChange) => {
      // If already hydrated, no need to subscribe
      if (useAuthStore.persist.hasHydrated()) {
        return () => {}
      }

      // Subscribe to hydration completion
      return useAuthStore.persist.onFinishHydration(onStoreChange)
    },
    // getSnapshot: returns current hydration state (client)
    () => useAuthStore.persist.hasHydrated(),
    // getServerSnapshot: returns false during SSR
    () => false
  )
}
