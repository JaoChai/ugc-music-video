import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { lazy, Suspense, type ComponentType } from 'react'

// Lazy load pages for better performance
const Dashboard = lazy(() => import('@/features/dashboard/pages/DashboardPage'))
const Login = lazy(() => import('@/features/auth/pages/LoginPage'))
const Register = lazy(() => import('@/features/auth/pages/RegisterPage'))
const JobList = lazy(() => import('@/features/job/pages/JobListPage'))
const JobDetail = lazy(() => import('@/features/job/pages/JobDetailPage'))
const CreateJob = lazy(() => import('@/features/job/pages/CreateJobPage'))
const Settings = lazy(() => import('@/features/settings/pages/SettingsPage'))

// Loading component
function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
    </div>
  )
}

// Wrap lazy components with Suspense
function withSuspense(Component: ComponentType) {
  return (
    <Suspense fallback={<PageLoader />}>
      <Component />
    </Suspense>
  )
}

const router = createBrowserRouter([
  {
    path: '/',
    element: withSuspense(Dashboard),
  },
  {
    path: '/login',
    element: withSuspense(Login),
  },
  {
    path: '/register',
    element: withSuspense(Register),
  },
  {
    path: '/jobs',
    element: withSuspense(JobList),
  },
  {
    path: '/jobs/create',
    element: withSuspense(CreateJob),
  },
  {
    path: '/jobs/:id',
    element: withSuspense(JobDetail),
  },
  {
    path: '/settings',
    element: withSuspense(Settings),
  },
])

export function AppRouter() {
  return <RouterProvider router={router} />
}
