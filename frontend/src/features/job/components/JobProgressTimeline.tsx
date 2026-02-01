import { cn } from '@/lib/utils'
import { Check, Loader2 } from 'lucide-react'
import type { JobStatus } from '../types'
import { STATUS_ORDER, STATUS_DISPLAY_NAMES } from '../types'

interface JobProgressTimelineProps {
  currentStatus: JobStatus
  className?: string
}

// Get the index of a status in the progression
function getStatusIndex(status: JobStatus): number {
  const index = STATUS_ORDER.indexOf(status)
  // For failed status, return -1 to indicate it's not in normal progression
  return index
}

export function JobProgressTimeline({ currentStatus, className }: JobProgressTimelineProps) {
  const currentIndex = getStatusIndex(currentStatus)
  const isFailed = currentStatus === 'failed'

  // Timeline steps (excluding pending and completed, they're handled separately)
  const timelineSteps: { status: JobStatus; label: string }[] = [
    { status: 'analyzing', label: STATUS_DISPLAY_NAMES.analyzing },
    { status: 'generating_music', label: STATUS_DISPLAY_NAMES.generating_music },
    { status: 'selecting_song', label: STATUS_DISPLAY_NAMES.selecting_song },
    { status: 'generating_image', label: STATUS_DISPLAY_NAMES.generating_image },
    { status: 'processing_video', label: STATUS_DISPLAY_NAMES.processing_video },
    { status: 'uploading', label: STATUS_DISPLAY_NAMES.uploading },
    { status: 'completed', label: STATUS_DISPLAY_NAMES.completed },
  ]

  return (
    <div className={cn('flow-root', className)}>
      <ul className="-mb-8">
        {timelineSteps.map((step, stepIdx) => {
          const stepIndex = getStatusIndex(step.status)
          const isCompleted = !isFailed && currentIndex > stepIndex
          const isCurrent = currentStatus === step.status
          const isFuture = !isFailed && currentIndex < stepIndex
          const isFailedStep = isFailed && currentIndex === stepIndex

          return (
            <li key={step.status}>
              <div className="relative pb-8">
                {/* Connector line */}
                {stepIdx !== timelineSteps.length - 1 && (
                  <span
                    className={cn(
                      'absolute left-4 top-4 -ml-px h-full w-0.5',
                      isCompleted ? 'bg-green-500' : 'bg-gray-200'
                    )}
                    aria-hidden="true"
                  />
                )}

                <div className="relative flex items-start space-x-3">
                  {/* Icon */}
                  <div className="relative">
                    {isCompleted ? (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-green-500">
                        <Check className="h-5 w-5 text-white" aria-hidden="true" />
                      </span>
                    ) : isCurrent ? (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-500 ring-4 ring-blue-100">
                        <Loader2 className="h-5 w-5 text-white animate-spin" aria-hidden="true" />
                      </span>
                    ) : isFailedStep || (isFailed && stepIndex > currentIndex) ? (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200">
                        <span className="h-2.5 w-2.5 rounded-full bg-gray-400" />
                      </span>
                    ) : isFuture ? (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200">
                        <span className="h-2.5 w-2.5 rounded-full bg-gray-400" />
                      </span>
                    ) : (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-green-500">
                        <Check className="h-5 w-5 text-white" aria-hidden="true" />
                      </span>
                    )}
                  </div>

                  {/* Label */}
                  <div className="min-w-0 flex-1 pt-1.5">
                    <p
                      className={cn(
                        'text-sm font-medium',
                        isCompleted && 'text-green-600',
                        isCurrent && 'text-blue-600',
                        (isFuture || isFailed) && 'text-gray-500'
                      )}
                    >
                      {step.label}
                    </p>
                  </div>
                </div>
              </div>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
