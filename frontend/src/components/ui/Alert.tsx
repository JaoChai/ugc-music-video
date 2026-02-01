import { type HTMLAttributes, type ReactNode } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const alertVariants = cva(
  'relative w-full rounded-lg border p-4',
  {
    variants: {
      variant: {
        default: 'bg-white border-gray-200 text-gray-900',
        destructive: 'bg-red-50 border-red-200 text-red-900',
        success: 'bg-green-50 border-green-200 text-green-900',
        warning: 'bg-yellow-50 border-yellow-200 text-yellow-900',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

const alertIconVariants = cva('h-5 w-5', {
  variants: {
    variant: {
      default: 'text-gray-600',
      destructive: 'text-red-600',
      success: 'text-green-600',
      warning: 'text-yellow-600',
    },
  },
  defaultVariants: {
    variant: 'default',
  },
})

export interface AlertProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof alertVariants> {
  icon?: ReactNode
}

function Alert({ className, variant, icon, children, ...props }: AlertProps) {
  const defaultIcons: Record<string, ReactNode> = {
    default: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
    destructive: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
    success: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    ),
    warning: (
      <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
    ),
  }

  const displayIcon = icon ?? defaultIcons[variant || 'default']

  return (
    <div
      role="alert"
      className={cn(alertVariants({ variant }), className)}
      {...props}
    >
      <div className="flex gap-3">
        {displayIcon && (
          <div className={cn(alertIconVariants({ variant }), 'flex-shrink-0')}>
            {displayIcon}
          </div>
        )}
        <div className="flex-1">{children}</div>
      </div>
    </div>
  )
}

export interface AlertTitleProps extends HTMLAttributes<HTMLHeadingElement> {}

function AlertTitle({ className, ...props }: AlertTitleProps) {
  return (
    <h5
      className={cn('mb-1 font-medium leading-none tracking-tight', className)}
      {...props}
    />
  )
}

export interface AlertDescriptionProps extends HTMLAttributes<HTMLParagraphElement> {}

function AlertDescription({ className, ...props }: AlertDescriptionProps) {
  return (
    <div
      className={cn('text-sm [&_p]:leading-relaxed', className)}
      {...props}
    />
  )
}

export { Alert, AlertTitle, AlertDescription, alertVariants }
