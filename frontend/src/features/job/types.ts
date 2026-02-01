// Job types for the UGC video generation pipeline

export type JobStatus =
  | 'pending'
  | 'analyzing'
  | 'generating_music'
  | 'selecting_song'
  | 'generating_image'
  | 'processing_video'
  | 'uploading'
  | 'completed'
  | 'failed'

export interface SongDetails {
  id: string
  title: string
  style: string
  lyrics?: string
  audio_url?: string
}

export interface Job {
  id: string
  concept: string
  model?: string
  status: JobStatus
  error_message?: string
  song?: SongDetails
  image_url?: string
  video_url?: string
  created: string
  updated: string
}

export interface CreateJobRequest {
  concept: string
  model?: string
}

export interface JobsResponse {
  page: number
  perPage: number
  totalItems: number
  totalPages: number
  items: Job[]
}

// Status progression order for timeline
export const STATUS_ORDER: JobStatus[] = [
  'pending',
  'analyzing',
  'generating_music',
  'selecting_song',
  'generating_image',
  'processing_video',
  'uploading',
  'completed',
]

// Terminal statuses - job is done (success or failure)
export const TERMINAL_STATUSES: JobStatus[] = ['completed', 'failed']

// Status display names
export const STATUS_DISPLAY_NAMES: Record<JobStatus, string> = {
  pending: 'Pending',
  analyzing: 'Analyzing',
  generating_music: 'Generating Music',
  selecting_song: 'Selecting Song',
  generating_image: 'Generating Image',
  processing_video: 'Processing Video',
  uploading: 'Uploading',
  completed: 'Completed',
  failed: 'Failed',
}
