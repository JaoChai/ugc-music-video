import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'

export type StatsCardVariant = 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'gray'

interface StatsCardProps {
  icon: LucideIcon
  title: string
  value: number | string
  variant?: StatsCardVariant
  isLoading?: boolean
}

const variantStyles: Record<StatsCardVariant, { bg: string; iconColor: string }> = {
  blue: {
    bg: 'bg-zinc-100',
    iconColor: 'text-zinc-600',
  },
  green: {
    bg: 'bg-green-100',
    iconColor: 'text-green-600',
  },
  yellow: {
    bg: 'bg-yellow-100',
    iconColor: 'text-yellow-600',
  },
  red: {
    bg: 'bg-red-100',
    iconColor: 'text-red-600',
  },
  purple: {
    bg: 'bg-purple-100',
    iconColor: 'text-purple-600',
  },
  gray: {
    bg: 'bg-gray-100',
    iconColor: 'text-gray-600',
  },
}

export function StatsCard({
  icon: Icon,
  title,
  value,
  variant = 'blue',
  isLoading = false,
}: StatsCardProps) {
  const styles = variantStyles[variant]

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
      <div className="flex items-center gap-4">
        <div className={cn('p-3 rounded-lg', styles.bg)}>
          <Icon className={cn('h-6 w-6', styles.iconColor)} />
        </div>
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          {isLoading ? (
            <div className="h-8 w-16 bg-gray-200 animate-pulse rounded mt-1" />
          ) : (
            <p className="text-2xl font-bold text-gray-900">{value}</p>
          )}
        </div>
      </div>
    </div>
  )
}
