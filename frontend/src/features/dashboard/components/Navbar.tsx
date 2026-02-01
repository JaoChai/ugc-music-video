import { Link, useLocation } from 'react-router-dom'
import { LayoutDashboard, Briefcase, LogOut, User, ChevronDown, Settings } from 'lucide-react'
import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui'
import { useAuthStore } from '@/stores/auth.store'
import { cn } from '@/lib/utils'

interface NavLinkProps {
  to: string
  icon: typeof LayoutDashboard
  children: React.ReactNode
}

function NavLink({ to, icon: Icon, children }: NavLinkProps) {
  const location = useLocation()
  const isActive = location.pathname === to || (to !== '/' && location.pathname.startsWith(to))

  return (
    <Link
      to={to}
      className={cn(
        'flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
        isActive
          ? 'bg-blue-50 text-blue-700'
          : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
      )}
    >
      <Icon className="h-4 w-4" />
      {children}
    </Link>
  )
}

function UserDropdown() {
  const { user, logout } = useAuthStore()
  const [isOpen, setIsOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100 transition-colors"
      >
        <div className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center">
          <User className="h-4 w-4 text-blue-600" />
        </div>
        <span className="hidden sm:inline">{user?.name || user?.email || 'User'}</span>
        <ChevronDown className={cn('h-4 w-4 transition-transform', isOpen && 'rotate-180')} />
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-50">
          <div className="px-4 py-2 border-b border-gray-100">
            <p className="text-sm font-medium text-gray-900">{user?.name || 'User'}</p>
            <p className="text-xs text-gray-500 truncate">{user?.email}</p>
          </div>
          <Link
            to="/settings"
            onClick={() => setIsOpen(false)}
            className="flex items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
          >
            <Settings className="h-4 w-4" />
            Settings
          </Link>
          <button
            onClick={() => {
              setIsOpen(false)
              logout()
            }}
            className="flex items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50 w-full"
          >
            <LogOut className="h-4 w-4" />
            Logout
          </button>
        </div>
      )}
    </div>
  )
}

export function Navbar() {
  const { isAuthenticated } = useAuthStore()

  return (
    <nav className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          {/* Logo */}
          <div className="flex items-center gap-8">
            <Link to="/" className="flex items-center gap-2">
              <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <LayoutDashboard className="h-5 w-5 text-white" />
              </div>
              <span className="font-bold text-xl text-gray-900">UGC Platform</span>
            </Link>

            {/* Navigation Links */}
            {isAuthenticated && (
              <div className="hidden md:flex items-center gap-1">
                <NavLink to="/" icon={LayoutDashboard}>Dashboard</NavLink>
                <NavLink to="/jobs" icon={Briefcase}>Jobs</NavLink>
              </div>
            )}
          </div>

          {/* Right Side */}
          <div className="flex items-center gap-4">
            {isAuthenticated ? (
              <UserDropdown />
            ) : (
              <>
                <Link to="/login">
                  <Button variant="ghost" size="sm">Sign In</Button>
                </Link>
                <Link to="/register">
                  <Button size="sm">Sign Up</Button>
                </Link>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Mobile Navigation */}
      {isAuthenticated && (
        <div className="md:hidden border-t border-gray-100">
          <div className="flex items-center gap-1 px-4 py-2">
            <NavLink to="/" icon={LayoutDashboard}>Dashboard</NavLink>
            <NavLink to="/jobs" icon={Briefcase}>Jobs</NavLink>
          </div>
        </div>
      )}
    </nav>
  )
}
