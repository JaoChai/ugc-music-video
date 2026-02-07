import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  _hasHydrated: boolean
  setUser: (user: User | null) => void
  setToken: (token: string | null) => void
  login: (user: User, token: string) => void
  logout: () => void
  isAdmin: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      _hasHydrated: false,
      setUser: (user) => set({ user, isAuthenticated: !!user }),
      setToken: (token) => set({ token }),
      login: (user, token) => {
        set({ user, token, isAuthenticated: true })
      },
      logout: () => {
        set({ user: null, token: null, isAuthenticated: false })
      },
      isAdmin: () => get().user?.role === 'admin',
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)

// Zustand v5: register hydration callback AFTER store is created
// Using onRehydrateStorage with useAuthStore.setState inside the callback
// can fail in v5 because the store variable may not be assigned yet
useAuthStore.persist.onFinishHydration(() => {
  localStorage.removeItem('auth_token')
  useAuthStore.setState({ _hasHydrated: true })
})

// If already hydrated (e.g. synchronous storage), set immediately
if (useAuthStore.persist.hasHydrated()) {
  useAuthStore.setState({ _hasHydrated: true })
}
