import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Card, CardContent, Button, Input } from '@/components/ui'
import { ArrowLeft, Plus, Briefcase, Search, X, Eye, XCircle } from 'lucide-react'
import { useJobs, useCancelJob } from '../api'
import { JobStatusBadge } from '../components'
import type { JobStatus } from '../types'
import { TERMINAL_STATUSES } from '../types'

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: '', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'analyzing', label: 'Analyzing' },
  { value: 'generating_music', label: 'Generating Music' },
  { value: 'selecting_song', label: 'Selecting Song' },
  { value: 'generating_image', label: 'Generating Image' },
  { value: 'processing_video', label: 'Processing Video' },
  { value: 'uploading', label: 'Uploading' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
]

export default function JobListPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [searchInput, setSearchInput] = useState('')

  const { data, isLoading, error } = useJobs({
    page,
    perPage: 10,
  })

  const cancelJob = useCancelJob()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setSearchQuery(searchInput)
    setPage(1)
  }

  const clearSearch = () => {
    setSearchInput('')
    setSearchQuery('')
    setPage(1)
  }

  const handleCancelJob = async (id: string, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    if (confirm('Are you sure you want to cancel this job?')) {
      await cancelJob.mutateAsync(id)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="min-h-screen bg-gray-100">
      {/* Header */}
      <div className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link to="/" className="text-gray-600 hover:text-gray-900">
                <ArrowLeft className="h-5 w-5" />
              </Link>
              <h1 className="text-2xl font-bold text-gray-900">Jobs</h1>
            </div>
            <Link to="/jobs/create">
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Job
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div className="flex flex-col sm:flex-row gap-4">
          {/* Search */}
          <form onSubmit={handleSearch} className="flex-1 flex gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder="Search by concept..."
                value={searchInput}
                onChange={(e) => setSearchInput(e.target.value)}
                className="pl-10"
              />
              {searchInput && (
                <button
                  type="button"
                  onClick={clearSearch}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  <X className="h-4 w-4" />
                </button>
              )}
            </div>
            <Button type="submit" variant="outline">
              Search
            </Button>
          </form>

          {/* Status filter */}
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value)
              setPage(1)
            }}
            className="block rounded-lg border border-gray-300 px-4 py-2 text-gray-900 focus:border-blue-500 focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 focus:outline-none"
          >
            {STATUS_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-8">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        ) : error ? (
          <Card>
            <CardContent className="py-12 text-center">
              <p className="text-red-600">Error loading jobs. Please try again.</p>
            </CardContent>
          </Card>
        ) : !data?.jobs?.length ? (
          <Card>
            <CardContent className="py-12 text-center">
              <Briefcase className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No jobs found</h3>
              <p className="text-gray-600 mb-4">
                {searchQuery || statusFilter
                  ? 'Try adjusting your filters.'
                  : 'Get started by creating your first job.'}
              </p>
              {!searchQuery && !statusFilter && (
                <Link to="/jobs/create">
                  <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    Create Job
                  </Button>
                </Link>
              )}
            </CardContent>
          </Card>
        ) : (
          <>
            {/* Table view */}
            <Card>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50 border-b border-gray-200">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Concept
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Status
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Created
                      </th>
                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {data.jobs.map((job) => (
                      <tr
                        key={job.id}
                        className="hover:bg-gray-50 cursor-pointer"
                        onClick={() => navigate(`/jobs/${job.id}`)}
                      >
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="bg-gray-100 p-2 rounded-lg">
                              <Briefcase className="h-4 w-4 text-gray-600" />
                            </div>
                            <div className="max-w-md">
                              <p className="font-medium text-gray-900 truncate">{job.concept}</p>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <JobStatusBadge status={job.status} />
                        </td>
                        <td className="px-6 py-4 text-sm text-gray-500">
                          {formatDate(job.created_at)}
                        </td>
                        <td className="px-6 py-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={(e) => {
                                e.stopPropagation()
                                navigate(`/jobs/${job.id}`)
                              }}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                            {!TERMINAL_STATUSES.includes(job.status as JobStatus) && (
                              <Button
                                size="sm"
                                variant="ghost"
                                onClick={(e) => handleCancelJob(job.id, e)}
                                disabled={cancelJob.isPending}
                                className="text-red-600 hover:text-red-700 hover:bg-red-50"
                              >
                                <XCircle className="h-4 w-4" />
                              </Button>
                            )}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>

            {/* Pagination */}
            {data.meta && data.meta.total_pages > 1 && (
              <div className="mt-4 flex items-center justify-between">
                <p className="text-sm text-gray-700">
                  Showing {(page - 1) * 10 + 1} to {Math.min(page * 10, data.meta.total)} of{' '}
                  {data.meta.total} results
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.min(data.meta!.total_pages, p + 1))}
                    disabled={page === data.meta.total_pages}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  )
}
