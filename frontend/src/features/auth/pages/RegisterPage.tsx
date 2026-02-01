import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'
import { Button, Input } from '@/components/ui'
import { AuthLayout } from '../components/AuthLayout'
import { useAuth } from '../hooks/useAuth'
import { registerSchema, type RegisterFormData } from '../schemas/auth.schema'

export default function RegisterPage() {
  const { register: registerUser, isRegistering, registerError } = useAuth()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      password: '',
      confirmPassword: '',
      name: '',
    },
  })

  const onSubmit = async (data: RegisterFormData) => {
    try {
      await registerUser({
        email: data.email,
        password: data.password,
        passwordConfirm: data.confirmPassword,
        name: data.name,
      })
    } catch {
      // Error is handled by useAuth hook
    }
  }

  return (
    <AuthLayout title="Create Account" subtitle="Sign up to get started with UGC Platform.">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {registerError && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {registerError}
          </div>
        )}

        <Input
          id="name"
          type="text"
          label="Name (optional)"
          placeholder="Your name"
          autoComplete="name"
          error={errors.name?.message}
          {...register('name')}
        />

        <Input
          id="email"
          type="email"
          label="Email"
          placeholder="you@example.com"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />

        <Input
          id="password"
          type="password"
          label="Password"
          placeholder="Create a password (min. 8 characters)"
          autoComplete="new-password"
          error={errors.password?.message}
          {...register('password')}
        />

        <Input
          id="confirmPassword"
          type="password"
          label="Confirm Password"
          placeholder="Confirm your password"
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />

        <Button type="submit" className="w-full" isLoading={isRegistering}>
          Sign Up
        </Button>
      </form>

      <div className="mt-6 text-center">
        <p className="text-gray-600 text-sm">
          Already have an account?{' '}
          <Link to="/login" className="text-blue-600 hover:text-blue-700 font-medium hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </AuthLayout>
  )
}
