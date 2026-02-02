import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link } from 'react-router-dom'
import { Button, Input } from '@/components/ui'
import { AuthLayout } from '../components/AuthLayout'
import { useAuth } from '../hooks/useAuth'
import { loginSchema, type LoginFormData } from '../schemas/auth.schema'

export default function LoginPage() {
  const { login, isLoggingIn, loginError } = useAuth()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      await login(data)
    } catch {
      // Error is handled by useAuth hook
    }
  }

  return (
    <AuthLayout title="เข้าสู่ระบบ" subtitle="ยินดีต้อนรับกลับ! กรุณาเข้าสู่ระบบเพื่อดำเนินการต่อ">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {loginError && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {loginError}
          </div>
        )}

        <Input
          id="email"
          type="email"
          label="อีเมล"
          placeholder="you@example.com"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />

        <Input
          id="password"
          type="password"
          label="รหัสผ่าน"
          placeholder="กรอกรหัสผ่านของคุณ"
          autoComplete="current-password"
          error={errors.password?.message}
          {...register('password')}
        />

        <Button type="submit" className="w-full" isLoading={isLoggingIn}>
          เข้าสู่ระบบ
        </Button>
      </form>

      <div className="mt-6 text-center">
        <p className="text-gray-600 text-sm">
          ยังไม่มีบัญชี?{' '}
          <Link to="/register" className="text-zinc-900 hover:text-zinc-700 font-medium hover:underline">
            สมัครสมาชิก
          </Link>
        </p>
      </div>
    </AuthLayout>
  )
}
