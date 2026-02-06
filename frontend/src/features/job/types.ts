// Job types for the UGC video generation pipeline

export type JobStatus =
  | 'pending'
  | 'analyzing'
  | 'generating_music'
  | 'selecting_song'
  | 'generating_image'
  | 'processing_video'
  | 'uploading'
  | 'uploading_youtube'
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
  youtube_url?: string
  youtube_video_id?: string
  youtube_error?: string
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
  'uploading_youtube',
  'completed',
]

// Terminal statuses - job is done (success or failure)
export const TERMINAL_STATUSES: JobStatus[] = ['completed', 'failed']

// Infer which stage a job failed at based on populated fields
export function inferFailedAtStatus(job: Job): JobStatus {
  if (job.youtube_url || job.youtube_error) return 'uploading_youtube'
  if (job.video_url) return 'uploading'
  if (job.image_url) return 'processing_video'
  if (job.image_prompt) return 'generating_image'
  if (job.generated_songs?.length) return 'selecting_song'
  if (job.song_prompt) return 'generating_music'
  return 'analyzing'
}

// Status display names
export const STATUS_DISPLAY_NAMES: Record<JobStatus, string> = {
  pending: 'รอดำเนินการ',
  analyzing: 'กำลังวิเคราะห์',
  generating_music: 'กำลังสร้างเพลง',
  selecting_song: 'กำลังเลือกเพลง',
  generating_image: 'กำลังสร้างภาพ',
  processing_video: 'กำลังประมวลผลวิดีโอ',
  uploading: 'กำลังอัปโหลด',
  uploading_youtube: 'กำลังอัปโหลด YouTube',
  completed: 'เสร็จสิ้น',
  failed: 'ล้มเหลว',
}
