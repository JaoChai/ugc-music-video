import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth.store'
import { authApi, type LoginRequest, type RegisterRequest } from '../api/auth.api'
import type { AxiosError } from 'axios'

interface ApiErrorResponse {
  message?: string
  data?: Record<string, { message: string }>
}

interface LocationState {
  from?: string
}

export function useAuth() {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const { login: storeLogin, logout: storeLogout, isAuthenticated } = useAuthStore()

  // Get the intended destination from location state (set by PrivateRoute)
  const from = (location.state as LocationState)?.from || '/'

  const loginMutation = useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: (response) => {
      storeLogin(response.user, response.token)
      queryClient.invalidateQueries({ queryKey: ['user'] })
      // Redirect to intended destination or home
      navigate(from, { replace: true })
    },
  })

  const registerMutation = useMutation({
    mutationFn: (data: RegisterRequest) => authApi.register(data),
    onSuccess: () => {
      navigate('/login')
    },
  })

  const logoutMutation = useMutation({
    mutationFn: () => authApi.logout(),
    onSuccess: () => {
      storeLogout()
      queryClient.clear()
      navigate('/login')
    },
  })

  const currentUserQuery = useQuery({
    queryKey: ['user', 'current'],
    queryFn: () => authApi.getCurrentUser(),
    enabled: isAuthenticated,
    retry: false,
  })

  const login = async (data: LoginRequest) => {
    return loginMutation.mutateAsync(data)
  }

  const register = async (data: RegisterRequest) => {
    return registerMutation.mutateAsync(data)
  }

  const logout = async () => {
    return logoutMutation.mutateAsync()
  }

  const getErrorMessage = (error: unknown): string => {
    if (!error) return 'An unexpected error occurred'

    const axiosError = error as AxiosError<ApiErrorResponse>

    if (axiosError.response?.data?.message) {
      return axiosError.response.data.message
    }

    if (axiosError.response?.data?.data) {
      const fieldErrors = axiosError.response.data.data
      const firstError = Object.values(fieldErrors)[0]
      if (firstError?.message) {
        return firstError.message
      }
    }

    if (axiosError.response?.status === 400) {
      return 'Invalid email or password'
    }

    if (axiosError.response?.status === 401) {
      return 'Invalid credentials'
    }

    return 'An unexpected error occurred. Please try again.'
  }

  return {
    login,
    register,
    logout,
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    loginError: loginMutation.error ? getErrorMessage(loginMutation.error) : null,
    registerError: registerMutation.error ? getErrorMessage(registerMutation.error) : null,
    currentUser: currentUserQuery.data,
    isLoadingUser: currentUserQuery.isLoading,
    isAuthenticated,
  }
}
