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
    <AuthLayout title="สร้างบัญชี" subtitle="สมัครสมาชิกเพื่อเริ่มต้นใช้งาน UGC Platform">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {registerError && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {registerError}
          </div>
        )}

        <Input
          id="name"
          type="text"
          label="ชื่อ (ไม่บังคับ)"
          placeholder="ชื่อของคุณ"
          autoComplete="name"
          error={errors.name?.message}
          {...register('name')}
        />

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
          placeholder="สร้างรหัสผ่าน (อย่างน้อย 8 ตัวอักษร)"
          autoComplete="new-password"
          error={errors.password?.message}
          {...register('password')}
        />

        <Input
          id="confirmPassword"
          type="password"
          label="ยืนยันรหัสผ่าน"
          placeholder="กรอกรหัสผ่านอีกครั้ง"
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />

        <Button type="submit" className="w-full" isLoading={isRegistering}>
          สมัครสมาชิก
        </Button>
      </form>

      <div className="mt-6 text-center">
        <p className="text-gray-600 text-sm">
          มีบัญชีอยู่แล้ว?{' '}
          <Link to="/login" className="text-zinc-900 hover:text-zinc-700 font-medium hover:underline">
            เข้าสู่ระบบ
          </Link>
        </p>
      </div>
    </AuthLayout>
  )
}
