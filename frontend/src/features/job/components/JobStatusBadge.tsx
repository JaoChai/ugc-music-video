import { cn } from '@/lib/utils'
import type { JobStatus } from '../types'
import { STATUS_DISPLAY_NAMES } from '../types'

interface JobStatusBadgeProps {
  status: JobStatus
  className?: string
}

const statusStyles: Record<JobStatus, string> = {
  pending: 'bg-gray-100 text-gray-800',
  analyzing: 'bg-zinc-100 text-zinc-800',
  generating_music: 'bg-zinc-100 text-zinc-800',
  selecting_song: 'bg-zinc-100 text-zinc-800',
  generating_image: 'bg-zinc-100 text-zinc-800',
  processing_video: 'bg-zinc-100 text-zinc-800',
  uploading: 'bg-zinc-100 text-zinc-800',
  uploading_youtube: 'bg-red-100 text-red-800',
  completed: 'bg-green-100 text-green-800',
  failed: 'bg-red-100 text-red-800',
}

// Statuses that should show animation
const animatedStatuses: JobStatus[] = [
  'analyzing',
  'generating_music',
  'selecting_song',
  'generating_image',
  'processing_video',
  'uploading',
  'uploading_youtube',
]

export function JobStatusBadge({ status, className }: JobStatusBadgeProps) {
  const isAnimated = animatedStatuses.includes(status)

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium',
        statusStyles[status],
        className
      )}
    >
      {isAnimated && (
        <span className="relative flex h-2 w-2">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-current opacity-75"></span>
          <span className="relative inline-flex rounded-full h-2 w-2 bg-current"></span>
        </span>
      )}
      {STATUS_DISPLAY_NAMES[status]}
    </span>
  )
}
