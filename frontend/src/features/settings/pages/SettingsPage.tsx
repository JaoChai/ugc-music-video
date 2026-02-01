import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Input } from '@/components/ui'
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card'
import { DashboardLayout } from '@/features/dashboard/components/DashboardLayout'
import { useAuthStore } from '@/stores/auth.store'
import { useUpdateProfileMutation } from '../hooks/useSettings'
import type { UpdateProfileInput } from '../api/settings.api'

interface SettingsFormData {
  name: string
  openrouter_model: string
}

export default function SettingsPage() {
  const navigate = useNavigate()
  const { user, isAuthenticated } = useAuthStore()
  const { updateProfile, isUpdating, isSuccess, error, reset } = useUpdateProfileMutation()

  const {
    register,
    handleSubmit,
    formState: { errors, isDirty },
    reset: resetForm,
  } = useForm<SettingsFormData>({
    defaultValues: {
      name: user?.name || '',
      openrouter_model: user?.openrouter_model || '',
    },
  })

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login')
    }
  }, [isAuthenticated, navigate])

  // Reset form when user data changes
  useEffect(() => {
    if (user) {
      resetForm({
        name: user.name || '',
        openrouter_model: user.openrouter_model || '',
      })
    }
  }, [user, resetForm])

  // Reset success state after 3 seconds
  useEffect(() => {
    if (isSuccess) {
      const timer = setTimeout(() => {
        reset()
      }, 3000)
      return () => clearTimeout(timer)
    }
  }, [isSuccess, reset])

  const onSubmit = async (data: SettingsFormData) => {
    try {
      const updateData: UpdateProfileInput = {}

      if (data.name !== user?.name) {
        updateData.name = data.name
      }

      if (data.openrouter_model !== (user?.openrouter_model || '')) {
        updateData.openrouter_model = data.openrouter_model
      }

      if (Object.keys(updateData).length > 0) {
        await updateProfile(updateData)
      }
    } catch {
      // Error is handled by the hook
    }
  }

  if (!isAuthenticated || !user) {
    return null
  }

  return (
    <DashboardLayout>
      <div className="max-w-2xl mx-auto">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
          <p className="mt-1 text-gray-600">Manage your account settings and preferences.</p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* Success Message */}
          {isSuccess && (
            <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg text-sm">
              Profile updated successfully!
            </div>
          )}

          {/* Error Message */}
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
              {error}
            </div>
          )}

          {/* Profile Section */}
          <Card>
            <CardHeader>
              <CardTitle>Profile</CardTitle>
              <CardDescription>Your personal information.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Input
                id="email"
                type="email"
                label="Email"
                value={user.email}
                disabled
                className="bg-gray-50 cursor-not-allowed"
              />

              <Input
                id="name"
                type="text"
                label="Name"
                placeholder="Enter your name"
                error={errors.name?.message}
                {...register('name')}
              />
            </CardContent>
          </Card>

          {/* Preferences Section */}
          <Card>
            <CardHeader>
              <CardTitle>Preferences</CardTitle>
              <CardDescription>Configure your default settings.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Input
                id="openrouter_model"
                type="text"
                label="Default OpenRouter Model"
                placeholder="anthropic/claude-3.5-sonnet"
                error={errors.openrouter_model?.message}
                {...register('openrouter_model')}
              />
              <p className="text-sm text-gray-500">
                This model will be used as the default for new jobs. You can override it per job.
              </p>
            </CardContent>
          </Card>

          {/* Save Button */}
          <div className="flex justify-end">
            <Button
              type="submit"
              isLoading={isUpdating}
              disabled={!isDirty || isUpdating}
            >
              Save Changes
            </Button>
          </div>
        </form>
      </div>
    </DashboardLayout>
  )
}
