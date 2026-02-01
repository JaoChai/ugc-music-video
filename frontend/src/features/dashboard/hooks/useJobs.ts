import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api'
import type { Job, JobsResponse, CreateJobRequest } from '@/features/job/types'

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
    queryFn: async (): Promise<JobsResponse> => {
      const { data } = await api.get<JobsResponse>('/api/collections/jobs/records', {
        params: { page, perPage, sort: '-created' },
      })
      return data
    },
  })
}

// Fetch single job by ID
export function useJobQuery(id: string) {
  return useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: async (): Promise<Job> => {
      const { data } = await api.get<Job>(`/api/collections/jobs/records/${id}`)
      return data
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
      const { data } = await api.get<JobsResponse>('/api/collections/jobs/records', {
        params: { perPage: 500 },
      })

      const items = data.items || []
      const total = items.length
      const completed = items.filter((job) => job.status === 'completed').length
      const inProgress = items.filter((job) =>
        !['completed', 'failed', 'pending'].includes(job.status)
      ).length
      const failed = items.filter((job) => job.status === 'failed').length
      const pending = items.filter((job) => job.status === 'pending').length

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
      const { data } = await api.get<JobsResponse>('/api/collections/jobs/records', {
        params: { page: 1, perPage: limit, sort: '-created' },
      })
      return data.items || []
    },
  })
}

// Create job mutation
export function useCreateJobMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (request: CreateJobRequest): Promise<Job> => {
      const { data } = await api.post<Job>('/api/collections/jobs/records', request)
      return data
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
    mutationFn: async (id: string): Promise<Job> => {
      const { data } = await api.patch<Job>(`/api/collections/jobs/records/${id}`, {
        status: 'failed',
        error_message: 'Cancelled by user',
      })
      return data
    },
    onSuccess: (data) => {
      // Update the specific job in cache
      queryClient.setQueryData(jobKeys.detail(data.id), data)
      // Invalidate lists and stats
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() })
      queryClient.invalidateQueries({ queryKey: jobKeys.stats() })
    },
  })
}
