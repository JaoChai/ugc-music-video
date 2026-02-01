// User types
export interface User {
  id: string
  email: string
  name: string | null
  openrouter_model: string
  created_at: string
  updated_at: string
}

// Auth types
export interface LoginInput {
  email: string
  password: string
}

export interface RegisterInput {
  email: string
  password: string
  name?: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RefreshResponse {
  token: string
}

// Job types
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

export interface SongPrompt {
  prompt: string
  style: string
  title: string
  model: string
  instrumental: boolean
}

export interface GeneratedSong {
  id: string
  audio_url: string
  title: string
  duration: number
}

export interface ImagePrompt {
  prompt: string
  aspect_ratio: string
  resolution: string
}

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

export interface CreateJobInput {
  concept: string
  model?: string
}

// API response types
export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface ApiError {
  code: number
  message: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: ApiError
  meta?: PaginationMeta
}
