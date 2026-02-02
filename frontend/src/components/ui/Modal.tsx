import {
  type ReactNode,
  useEffect,
  useCallback,
  type MouseEvent,
} from 'react'
import { createPortal } from 'react-dom'
import { cn } from '@/lib/utils'

export interface ModalProps {
  isOpen: boolean
  onClose: () => void
  children: ReactNode
  className?: string
}

function Modal({ isOpen, onClose, children, className }: ModalProps) {
  const handleEscape = useCallback(
    (event: globalThis.KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    },
    [onClose]
  )

  useEffect(() => {
    if (isOpen) {
      document.addEventListener('keydown', handleEscape)
      document.body.style.overflow = 'hidden'
    }

    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.body.style.overflow = 'unset'
    }
  }, [isOpen, handleEscape])

  const handleOverlayClick = (event: MouseEvent<HTMLDivElement>) => {
    if (event.target === event.currentTarget) {
      onClose()
    }
  }

  if (!isOpen) return null

  return createPortal(
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
    >
      {/* Overlay */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
        onClick={handleOverlayClick}
        aria-hidden="true"
      />

      {/* Modal Content */}
      <div
        className={cn(
          'relative z-50 w-full max-w-lg rounded-lg bg-white shadow-xl',
          'animate-in fade-in-0 zoom-in-95 duration-200',
          className
        )}
      >
        {children}
      </div>
    </div>,
    document.body
  )
}

export interface ModalHeaderProps {
  children: ReactNode
  onClose?: () => void
  className?: string
}

function ModalHeader({ children, onClose, className }: ModalHeaderProps) {
  return (
    <div
      className={cn(
        'flex items-center justify-between border-b border-gray-200 px-6 py-4',
        className
      )}
    >
      <h2 className="text-lg font-semibold text-gray-900">{children}</h2>
      {onClose && (
        <button
          onClick={onClose}
          className="rounded-md p-1 text-gray-400 hover:bg-zinc-100 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-zinc-500"
          aria-label="ปิด"
        >
          <svg
            className="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      )}
    </div>
  )
}

export interface ModalContentProps {
  children: ReactNode
  className?: string
}

function ModalContent({ children, className }: ModalContentProps) {
  return <div className={cn('px-6 py-4', className)}>{children}</div>
}

export interface ModalFooterProps {
  children: ReactNode
  className?: string
}

function ModalFooter({ children, className }: ModalFooterProps) {
  return (
    <div
      className={cn(
        'flex items-center justify-end gap-3 border-t border-gray-200 bg-gray-50 px-6 py-4',
        className
      )}
    >
      {children}
    </div>
  )
}

export { Modal, ModalHeader, ModalContent, ModalFooter }
