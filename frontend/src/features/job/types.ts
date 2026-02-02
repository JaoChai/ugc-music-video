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

// Matches Go backend SongPrompt struct
export interface SongPrompt {
  prompt: string
  style: string
  title: string
  model: string
  instrumental: boolean
}

// Matches Go backend GeneratedSong struct
export interface GeneratedSong {
  id: string
  audio_url: string
  title: string
  duration: number
}

// Matches Go backend ImagePrompt struct
export interface ImagePrompt {
  prompt: string
  aspect_ratio: string
  resolution: string
}

// Matches Go backend JobResponse struct
export interface Job {
  id: string
  user_id: string
  status: JobStatus
  concept: string
  llm_model: string
  song_prompt?: SongPrompt
  generated_songs?: GeneratedSong[]
  selected_song_id?: string
  image_prompt?: ImagePrompt
  audio_url?: string
  image_url?: string
  video_url?: string
  error_message?: string
  created_at: string
  updated_at: string
}

export interface CreateJobRequest {
  concept: string
  model?: string
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
