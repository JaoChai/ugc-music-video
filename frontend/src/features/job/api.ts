import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/axios'
import type { ApiResponse } from '@/types'
import type { Job, CreateJobRequest } from './types'

// Query keys
export const jobKeys = {
  all: ['jobs'] as const,
  lists: () => [...jobKeys.all, 'list'] as const,
  list: (filters: Record<string, unknown>) => [...jobKeys.lists(), filters] as const,
  details: () => [...jobKeys.all, 'detail'] as const,
  detail: (id: string) => [...jobKeys.details(), id] as const,
}

// API functions
async function fetchJobs(params: {
  page?: number
  perPage?: number
}): Promise<{ jobs: Job[]; meta: ApiResponse<Job[]>['meta'] }> {
  const response = await api.get<ApiResponse<Job[]>>('/api/v1/jobs', {
    params: { page: params.page || 1, per_page: params.perPage || 10 },
  })

  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to fetch jobs')
  }

  return {
    jobs: response.data.data,
    meta: response.data.meta,
  }
}

async function fetchJob(id: string): Promise<Job> {
  const response = await api.get<ApiResponse<Job>>(`/api/v1/jobs/${id}`)

  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to fetch job')
  }

  return response.data.data
}

async function createJob(data: CreateJobRequest): Promise<Job> {
  const response = await api.post<ApiResponse<Job>>('/api/v1/jobs', data)

  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to create job')
  }

  return response.data.data
}

async function cancelJob(id: string): Promise<void> {
  const response = await api.delete<ApiResponse<null>>(`/api/v1/jobs/${id}`)

  if (!response.data.success) {
    throw new Error(response.data.error?.message || 'Failed to cancel job')
  }
}

// Hooks
export function useJobs(params: {
  page?: number
  perPage?: number
} = {}) {
  return useQuery({
    queryKey: jobKeys.list(params),
    queryFn: () => fetchJobs(params),
  })
}

export function useJob(id: string, options?: { refetchInterval?: number | false }) {
  return useQuery({
    queryKey: jobKeys.detail(id),
    queryFn: () => fetchJob(id),
    enabled: !!id,
    refetchInterval: options?.refetchInterval,
  })
}

export function useCreateJob() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createJob,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() })
    },
  })
}

export function useCancelJob() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: cancelJob,
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() })
      queryClient.invalidateQueries({ queryKey: jobKeys.detail(id) })
    },
  })
}
