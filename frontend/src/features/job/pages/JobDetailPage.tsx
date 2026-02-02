import { Link, useParams } from 'react-router-dom'
import { Card, CardHeader, CardContent, Button } from '@/components/ui'
import {
  ArrowLeft,
  XCircle,
  AlertTriangle,
  Music,
  Image as ImageIcon,
  Video,
  Calendar,
  Clock,
} from 'lucide-react'
import { useJob, useCancelJob } from '../api'
import { JobStatusBadge, JobProgressTimeline, VideoPlayer, DownloadButton } from '../components'
import { TERMINAL_STATUSES } from '../types'

export default function JobDetailPage() {
  const { id } = useParams<{ id: string }>()
  const cancelJob = useCancelJob()

  // Fetch job data
  const { data: job, isLoading, error } = useJob(id!)

  // Determine if we should auto-refresh based on current status
  const shouldRefresh = job && !TERMINAL_STATUSES.includes(job.status)

  // Use a separate query for auto-refresh to avoid conditional hook calls
  const { data: refreshedJob } = useJob(id!, {
    refetchInterval: shouldRefresh ? 5000 : false,
  })

  // Use the most recent data available
  const currentJob = refreshedJob || job

  const handleCancel = async () => {
    if (!id) return
    if (confirm('Are you sure you want to cancel this job?')) {
      await cancelJob.mutateAsync(id)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    )
  }

  if (error || !currentJob) {
    return (
      <div className="min-h-screen bg-gray-100">
        <div className="bg-white shadow-sm border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <div className="flex items-center gap-4">
              <Link to="/jobs" className="text-gray-600 hover:text-gray-900">
                <ArrowLeft className="h-5 w-5" />
              </Link>
              <h1 className="text-2xl font-bold text-gray-900">Job Not Found</h1>
            </div>
          </div>
        </div>
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Card>
            <CardContent className="py-12 text-center">
              <AlertTriangle className="h-12 w-12 text-yellow-500 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Job not found</h3>
              <p className="text-gray-600 mb-4">
                The job you're looking for doesn't exist or has been deleted.
              </p>
              <Link to="/jobs">
                <Button>Back to Jobs</Button>
              </Link>
            </CardContent>
          </Card>
        </main>
      </div>
    )
  }

  const isTerminal = TERMINAL_STATUSES.includes(currentJob.status)

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <div className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link to="/jobs" className="text-gray-600 hover:text-gray-900">
                <ArrowLeft className="h-5 w-5" />
              </Link>
              <div>
                <h1 className="text-2xl font-bold text-gray-900">Job Details</h1>
                <p className="text-sm text-gray-500 font-mono">{currentJob.id}</p>
              </div>
            </div>
            {!isTerminal && (
              <Button
                variant="destructive"
                onClick={handleCancel}
                disabled={cancelJob.isPending}
                isLoading={cancelJob.isPending}
              >
                <XCircle className="h-4 w-4 mr-2" />
                Cancel Job
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Column */}
          <div className="lg:col-span-2 space-y-6">
            {/* Job Info Header */}
            <Card>
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h2 className="text-lg font-semibold text-gray-900 mb-2">Concept</h2>
                    <p className="text-gray-700 whitespace-pre-wrap">{currentJob.concept}</p>
                  </div>
                  <JobStatusBadge status={currentJob.status} />
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center gap-6 text-sm text-gray-500">
                  <div className="flex items-center gap-1.5">
                    <Calendar className="h-4 w-4" />
                    <span>Created: {formatDate(currentJob.created_at)}</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <Clock className="h-4 w-4" />
                    <span>Updated: {formatDate(currentJob.updated_at)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Error Section */}
            {currentJob.status === 'failed' && currentJob.error_message && (
              <Card className="border-red-200 bg-red-50">
                <CardHeader>
                  <div className="flex items-center gap-2 text-red-700">
                    <AlertTriangle className="h-5 w-5" />
                    <h2 className="text-lg font-semibold">Error</h2>
                  </div>
                </CardHeader>
                <CardContent>
                  <p className="text-red-700">{currentJob.error_message}</p>
                </CardContent>
              </Card>
            )}

            {/* Song Prompt Section */}
            {currentJob.song_prompt && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Music className="h-5 w-5 text-purple-600" />
                    <h2 className="text-lg font-semibold text-gray-900">Song Prompt</h2>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <dt className="text-sm text-gray-500">Title</dt>
                      <dd className="text-gray-900 font-medium">{currentJob.song_prompt.title}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-gray-500">Style</dt>
                      <dd className="text-gray-900">{currentJob.song_prompt.style}</dd>
                    </div>
                  </div>

                  {currentJob.song_prompt.prompt && (
                    <div>
                      <dt className="text-sm text-gray-500 mb-1">Prompt</dt>
                      <dd className="text-gray-700 text-sm whitespace-pre-wrap bg-gray-50 p-3 rounded-lg">
                        {currentJob.song_prompt.prompt}
                      </dd>
                    </div>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Generated Songs Section */}
            {currentJob.generated_songs && currentJob.generated_songs.length > 0 && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <Music className="h-5 w-5 text-purple-600" />
                    <h2 className="text-lg font-semibold text-gray-900">Generated Songs</h2>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  {currentJob.generated_songs.map((song) => (
                    <div key={song.id} className="border rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-medium text-gray-900">{song.title}</span>
                        {currentJob.selected_song_id === song.id && (
                          <span className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded-full">
                            Selected
                          </span>
                        )}
                      </div>
                      {song.audio_url && (
                        <audio controls className="w-full" src={song.audio_url}>
                          Your browser does not support the audio element.
                        </audio>
                      )}
                    </div>
                  ))}
                </CardContent>
              </Card>
            )}

            {/* Image Section */}
            {currentJob.image_url && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <ImageIcon className="h-5 w-5 text-green-600" />
                    <h2 className="text-lg font-semibold text-gray-900">Generated Image</h2>
                  </div>
                </CardHeader>
                <CardContent>
                  <img
                    src={currentJob.image_url}
                    alt="Generated image"
                    className="w-full rounded-lg"
                  />
                </CardContent>
              </Card>
            )}

            {/* Video Section */}
            {currentJob.video_url && (
              <Card>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Video className="h-5 w-5 text-blue-600" />
                      <h2 className="text-lg font-semibold text-gray-900">Generated Video</h2>
                    </div>
                    <DownloadButton
                      url={currentJob.video_url}
                      filename={`ugc-video-${currentJob.id}.mp4`}
                    />
                  </div>
                </CardHeader>
                <CardContent>
                  <VideoPlayer
                    src={currentJob.video_url}
                    poster={currentJob.image_url}
                  />
                </CardContent>
              </Card>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Progress Timeline */}
            <Card>
              <CardHeader>
                <h3 className="font-semibold text-gray-900">Progress</h3>
                {!isTerminal && (
                  <p className="text-xs text-gray-500 mt-1">
                    Auto-refreshing every 5 seconds...
                  </p>
                )}
              </CardHeader>
              <CardContent>
                <JobProgressTimeline currentStatus={currentJob.status} />
              </CardContent>
            </Card>

            {/* Job Info */}
            <Card>
              <CardHeader>
                <h3 className="font-semibold text-gray-900">Job Information</h3>
              </CardHeader>
              <CardContent>
                <dl className="space-y-4">
                  <div>
                    <dt className="text-sm text-gray-500">Job ID</dt>
                    <dd className="text-gray-900 font-mono text-sm break-all">{currentJob.id}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-gray-500">Status</dt>
                    <dd className="mt-1">
                      <JobStatusBadge status={currentJob.status} />
                    </dd>
                  </div>
                  {currentJob.llm_model && (
                    <div>
                      <dt className="text-sm text-gray-500">Model</dt>
                      <dd className="text-gray-900 font-mono text-sm">{currentJob.llm_model}</dd>
                    </div>
                  )}
                  <div>
                    <dt className="text-sm text-gray-500">Created</dt>
                    <dd className="text-gray-900 text-sm">{formatDate(currentJob.created_at)}</dd>
                  </div>
                  <div>
                    <dt className="text-sm text-gray-500">Last Updated</dt>
                    <dd className="text-gray-900 text-sm">{formatDate(currentJob.updated_at)}</dd>
                  </div>
                </dl>
              </CardContent>
            </Card>

            {/* Quick Actions */}
            <Card>
              <CardHeader>
                <h3 className="font-semibold text-gray-900">Actions</h3>
              </CardHeader>
              <CardContent className="space-y-2">
                <Link to="/jobs" className="block">
                  <Button variant="outline" className="w-full">
                    <ArrowLeft className="h-4 w-4 mr-2" />
                    Back to Jobs
                  </Button>
                </Link>
                <Link to="/jobs/create" className="block">
                  <Button variant="outline" className="w-full">
                    Create New Job
                  </Button>
                </Link>
                {currentJob.video_url && (
                  <DownloadButton
                    url={currentJob.video_url}
                    filename={`ugc-video-${currentJob.id}.mp4`}
                  />
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </main>
    </div>
  )
}
