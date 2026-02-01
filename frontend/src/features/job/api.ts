import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api'
import type { Job, JobsResponse, CreateJobRequest } from './types'

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
  status?: string
  search?: string
}): Promise<JobsResponse> {
  const searchParams = new URLSearchParams()
  if (params.page) searchParams.set('page', params.page.toString())
  if (params.perPage) searchParams.set('perPage', params.perPage.toString())
  if (params.status) searchParams.set('filter', `status="${params.status}"`)
  if (params.search) {
    const existingFilter = searchParams.get('filter')
    const searchFilter = `concept~"${params.search}"`
    searchParams.set('filter', existingFilter ? `${existingFilter}&&${searchFilter}` : searchFilter)
  }

  const response = await api.get(`/api/collections/jobs/records?${searchParams.toString()}`)
  return response.data
}

async function fetchJob(id: string): Promise<Job> {
  const response = await api.get(`/api/collections/jobs/records/${id}`)
  return response.data
}

async function createJob(data: CreateJobRequest): Promise<Job> {
  const response = await api.post('/api/collections/jobs/records', data)
  return response.data
}

async function cancelJob(id: string): Promise<Job> {
  const response = await api.patch(`/api/collections/jobs/records/${id}`, {
    status: 'failed',
    error_message: 'Cancelled by user',
  })
  return response.data
}

// Hooks
export function useJobs(params: {
  page?: number
  perPage?: number
  status?: string
  search?: string
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
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: jobKeys.lists() })
      queryClient.setQueryData(jobKeys.detail(data.id), data)
    },
  })
}
