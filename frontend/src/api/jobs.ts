import { api } from '@/lib/axios'
import type {
  Job,
  CreateJobInput,
  PaginationMeta,
  ApiResponse,
} from './types'

export async function createJob(input: CreateJobInput): Promise<Job> {
  const response = await api.post<ApiResponse<Job>>('/api/jobs', input)
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to create job')
  }
  return response.data.data
}

export async function getJobs(
  page: number = 1,
  perPage: number = 10
): Promise<{ jobs: Job[]; meta: PaginationMeta }> {
  const response = await api.get<ApiResponse<Job[]>>('/api/jobs', {
    params: { page, per_page: perPage },
  })
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to get jobs')
  }
  return {
    jobs: response.data.data,
    meta: response.data.meta || {
      page,
      per_page: perPage,
      total: response.data.data.length,
      total_pages: 1,
    },
  }
}

export async function getJob(id: string): Promise<Job> {
  const response = await api.get<ApiResponse<Job>>(`/api/jobs/${id}`)
  if (!response.data.success || !response.data.data) {
    throw new Error(response.data.error?.message || 'Failed to get job')
  }
  return response.data.data
}

export async function cancelJob(id: string): Promise<void> {
  const response = await api.post<ApiResponse<null>>(`/api/jobs/${id}/cancel`)
  if (!response.data.success) {
    throw new Error(response.data.error?.message || 'Failed to cancel job')
  }
}
