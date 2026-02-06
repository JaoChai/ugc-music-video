import { cn } from '@/lib/utils'
import { Check, Loader2, XCircle } from 'lucide-react'
import type { JobStatus } from '../types'
import { STATUS_ORDER, STATUS_DISPLAY_NAMES } from '../types'

interface JobProgressTimelineProps {
  currentStatus: JobStatus
  failedAtStatus?: JobStatus
  className?: string
}

// Get the index of a status in the progression
function getStatusIndex(status: JobStatus): number {
  return STATUS_ORDER.indexOf(status)
}

export function JobProgressTimeline({ currentStatus, failedAtStatus, className }: JobProgressTimelineProps) {
  const currentIndex = getStatusIndex(currentStatus)
  const isFailed = currentStatus === 'failed'
  const failedAtIndex = failedAtStatus ? getStatusIndex(failedAtStatus) : -1

  // Timeline steps (excluding pending, they're handled separately)
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

          // Determine step state
          let isCompleted = false
          let isCurrent = false
          let isFailedStep = false
          let isFuture = false

          if (isFailed && failedAtIndex >= 0) {
            // Failed job: show checkmarks up to the failed step, red X at failed step, gray after
            if (stepIndex < failedAtIndex) {
              isCompleted = true
            } else if (stepIndex === failedAtIndex) {
              isFailedStep = true
            } else {
              isFuture = true
            }
          } else if (isFailed) {
            // Failed job without failedAtStatus: all gray
            isFuture = true
          } else {
            // Normal progression
            isCompleted = currentIndex > stepIndex
            isCurrent = currentStatus === step.status && currentStatus !== 'completed'
            isFuture = currentIndex < stepIndex
          }

          return (
            <li key={step.status}>
              <div className="relative pb-8">
                {/* Connector line */}
                {stepIdx !== timelineSteps.length - 1 && (
                  <span
                    className={cn(
                      'absolute left-4 top-4 -ml-px h-full w-0.5',
                      isCompleted ? 'bg-green-500' :
                      isFailedStep ? 'bg-red-500' :
                      'bg-gray-200'
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
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-zinc-900 ring-4 ring-zinc-100">
                        <Loader2 className="h-5 w-5 text-white animate-spin" aria-hidden="true" />
                      </span>
                    ) : isFailedStep ? (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-red-500">
                        <XCircle className="h-5 w-5 text-white" aria-hidden="true" />
                      </span>
                    ) : isFuture ? (
                      <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200">
                        <span className="h-2.5 w-2.5 rounded-full bg-gray-400" />
                      </span>
                    ) : (
                      // Terminal completed state for 'completed' step
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
                        (isCompleted || (currentStatus === 'completed' && step.status === 'completed')) && 'text-green-600',
                        isCurrent && 'text-zinc-900',
                        isFailedStep && 'text-red-600',
                        isFuture && 'text-gray-500'
                      )}
                    >
                      {step.label}
                      {isFailedStep && ' — ล้มเหลว'}
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
