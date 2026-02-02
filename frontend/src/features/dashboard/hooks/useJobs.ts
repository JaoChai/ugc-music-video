import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/axios'
import type { Job, CreateJobRequest } from '@/features/job/types'

// API Response types (matching Go backend)
interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: number
    message: string
  }
  meta?: {
    page: number
    per_page: number
    total: number
    total_pages: number
  }
}

// Query keys
export const jobKeys = {
  all: ['jobs'] as const,
  lists: () => [...jobKeys.all, 'list'] as const,
  list: (page: number, perPage: number) => [...jobKeys.lists(), { page, perPage }] as const,
  details: () => [...jobKeys.all, 'detail'] as const,
  detail: (id: string) => [...jobKeys.details(), id] as const,
  stats: () => [...jobKeys.all, 'stats'] as const,
}

// Fetch jobs list with pagination
export function useJobsQuery(page: number = 1, perPage: number = 10) {
  return useQuery({
    queryKey: jobKeys.list(page, perPage),
    queryFn: async () => {
      const response = await api.get<ApiResponse<Job[]>>('/api/v1/jobs', {
        params: { page, per_page: perPage },
      })

      if (!response.data.success || !response.data.data) {
        throw new Error(response.data.error?.message || 'Failed to fetch jobs')
      }

      return {
        jobs: response.data.data,
        meta: response.data.meta,
      }
    },
  })
}

// Fetch single job by ID
export function useJobQuery(id: string) {
  return useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: async (): Promise<Job> => {
      const response = await api.get<ApiResponse<Job>>(`/api/v1/jobs/${id}`)

      if (!response.data.success || !response.data.data) {
        throw new Error(response.data.error?.message || 'Failed to fetch job')
      }

      return response.data.data
    },
    enabled: !!id,
  })
}

// Fetch job stats for dashboard
export function useJobStatsQuery() {
  return useQuery({
    queryKey: jobKeys.stats(),
    queryFn: async () => {
      // Fetch all jobs to calculate stats
      const response = await api.get<ApiResponse<Job[]>>('/api/v1/jobs', {
        params: { per_page: 500 },
      })

      if (!response.data.success || !response.data.data) {
        throw new Error(response.data.error?.message || 'Failed to fetch jobs')
      }

      const jobs = response.data.data
      const total = jobs.length
      const completed = jobs.filter((job) => job.status === 'completed').length
      const inProgress = jobs.filter((job) =>
        !['completed', 'failed', 'pending'].includes(job.status)
      ).length
      const failed = jobs.filter((job) => job.status === 'failed').length
      const pending = jobs.filter((job) => job.status === 'pending').length

      return {
        total,
        completed,
        inProgress,
        failed,
        pending,
      }
    },
  })
}

// Fetch recent jobs for dashboard
export function useRecentJobsQuery(limit: number = 5) {
  return useQuery({
    queryKey: [...jobKeys.lists(), 'recent', limit],
    queryFn: async (): Promise<Job[]> => {
      const response = await api.get<ApiResponse<Job[]>>('/api/v1/jobs', {
        params: { page: 1, per_page: limit },
      })

      if (!response.data.success || !response.data.data) {
        throw new Error(response.data.error?.message || 'Failed to fetch jobs')
      }

      return response.data.data
    },
  })
}

// Create job mutation
export function useCreateJobMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (request: CreateJobRequest): Promise<Job> => {
      const response = await api.post<ApiResponse<Job>>('/api/v1/jobs', request)

      if (!response.data.success || !response.data.data) {
        throw new Error(response.data.error?.message || 'Failed to create job')
      }

      return response.data.data
    },
    onSuccess: () => {
      // Invalidate all job queries to refetch
      queryClient.invalidateQueries({ queryKey: jobKeys.all })
    },
  })
}

// Cancel job mutation
export function useCancelJobMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string): Promise<void> => {
      const response = await api.delete<ApiResponse<null>>(`/api/v1/jobs/${id}`)

      if (!response.data.success) {
        throw new Error(response.data.error?.message || 'Failed to cancel job')
      }
    },
    onSuccess: (_data, id) => {
      // Invalidate the specific job and lists
      queryClient.invalidateQueries({ queryKey: jobKeys.detail(id) })
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() })
      queryClient.invalidateQueries({ queryKey: jobKeys.stats() })
    },
  })
}
