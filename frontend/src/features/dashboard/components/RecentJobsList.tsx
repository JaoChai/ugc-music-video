import { Link } from 'react-router-dom'
import { Briefcase, ChevronRight, Clock, AlertCircle, CheckCircle2, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { Job, JobStatus } from '@/features/job/types'
import { STATUS_DISPLAY_NAMES } from '@/features/job/types'

interface RecentJobsListProps {
  jobs: Job[]
  isLoading?: boolean
}

// Status badge styling
const statusStyles: Record<JobStatus, { bg: string; text: string; icon: typeof CheckCircle2 }> = {
  pending: { bg: 'bg-gray-100', text: 'text-gray-700', icon: Clock },
  analyzing: { bg: 'bg-zinc-100', text: 'text-zinc-700', icon: Loader2 },
  generating_music: { bg: 'bg-purple-100', text: 'text-purple-700', icon: Loader2 },
  selecting_song: { bg: 'bg-purple-100', text: 'text-purple-700', icon: Loader2 },
  generating_image: { bg: 'bg-indigo-100', text: 'text-indigo-700', icon: Loader2 },
  processing_video: { bg: 'bg-cyan-100', text: 'text-cyan-700', icon: Loader2 },
  uploading: { bg: 'bg-teal-100', text: 'text-teal-700', icon: Loader2 },
  uploading_youtube: { bg: 'bg-red-100', text: 'text-red-700', icon: Loader2 },
  completed: { bg: 'bg-green-100', text: 'text-green-700', icon: CheckCircle2 },
  failed: { bg: 'bg-red-100', text: 'text-red-700', icon: AlertCircle },
}

function StatusBadge({ status }: { status: JobStatus }) {
  const style = statusStyles[status] || statusStyles.pending
  const Icon = style.icon
  const isLoading = ['analyzing', 'generating_music', 'selecting_song', 'generating_image', 'processing_video', 'uploading', 'uploading_youtube'].includes(status)

  return (
    <span className={cn('inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium', style.bg, style.text)}>
      <Icon className={cn('h-3 w-3', isLoading && 'animate-spin')} />
      {STATUS_DISPLAY_NAMES[status] || status}
    </span>
  )
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'เมื่อสักครู่'
  if (diffMins < 60) return `${diffMins} นาทีที่แล้ว`
  if (diffHours < 24) return `${diffHours} ชั่วโมงที่แล้ว`
  if (diffDays < 7) return `${diffDays} วันที่แล้ว`

  return date.toLocaleDateString('th-TH', { month: 'short', day: 'numeric' })
}

function truncateText(text: string, maxLength: number = 50): string {
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength).trim() + '...'
}

function JobListSkeleton() {
  return (
    <div className="divide-y divide-gray-100">
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="p-4 animate-pulse">
          <div className="flex items-center gap-4">
            <div className="h-10 w-10 bg-gray-200 rounded-lg" />
            <div className="flex-1">
              <div className="h-4 bg-gray-200 rounded w-3/4 mb-2" />
              <div className="h-3 bg-gray-100 rounded w-1/4" />
            </div>
            <div className="h-6 w-20 bg-gray-200 rounded-full" />
          </div>
        </div>
      ))}
    </div>
  )
}

function EmptyState() {
  return (
    <div className="py-12 text-center">
      <Briefcase className="h-12 w-12 text-gray-300 mx-auto mb-4" />
      <h3 className="text-sm font-medium text-gray-900 mb-1">ยังไม่มีงาน</h3>
      <p className="text-sm text-gray-500">สร้างงานแรกของคุณเพื่อเริ่มต้น</p>
    </div>
  )
}

export function RecentJobsList({ jobs, isLoading }: RecentJobsListProps) {
  if (isLoading) {
    return <JobListSkeleton />
  }

  if (!jobs || jobs.length === 0) {
    return <EmptyState />
  }

  return (
    <div className="divide-y divide-gray-100">
      {jobs.map((job) => (
        <Link
          key={job.id}
          to={`/jobs/${job.id}`}
          className="flex items-center gap-4 p-4 hover:bg-gray-50 transition-colors group"
        >
          <div className="flex-shrink-0 bg-gray-100 p-2.5 rounded-lg group-hover:bg-gray-200 transition-colors">
            <Briefcase className="h-5 w-5 text-gray-600" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-900 truncate">
              {truncateText(job.concept)}
            </p>
            <p className="text-xs text-gray-500 mt-0.5">
              {formatDate(job.created_at)}
            </p>
          </div>
          <div className="flex items-center gap-3">
            <StatusBadge status={job.status} />
            <ChevronRight className="h-4 w-4 text-gray-400 group-hover:text-gray-600 transition-colors" />
          </div>
        </Link>
      ))}
    </div>
  )
}
