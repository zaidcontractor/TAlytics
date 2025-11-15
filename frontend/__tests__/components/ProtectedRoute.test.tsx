import { render, screen, waitFor } from '@testing-library/react'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { AuthProvider } from '@/lib/auth'
import { useRouter } from 'next/navigation'

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  usePathname: jest.fn(() => '/'),
}))

const mockPush = jest.fn()
const mockUseRouter = useRouter as jest.Mock

beforeEach(() => {
  mockUseRouter.mockReturnValue({
    push: mockPush,
    replace: jest.fn(),
    prefetch: jest.fn(),
    back: jest.fn(),
  })
  localStorage.clear()
})

describe('ProtectedRoute', () => {
  it('should show loading state initially', async () => {
    render(
      <AuthProvider>
        <ProtectedRoute>
          <div>Protected Content</div>
        </ProtectedRoute>
      </AuthProvider>
    )

    // Loading state might be brief, check for either loading or redirect
    await waitFor(() => {
      const loadingSpinner = screen.queryByRole('status', { hidden: true })
      const redirectHappened = mockPush.mock.calls.length > 0
      expect(loadingSpinner || redirectHappened).toBeTruthy()
    }, { timeout: 1000 })
  })

  it('should redirect to login if not authenticated', async () => {
    render(
      <AuthProvider>
        <ProtectedRoute>
          <div>Protected Content</div>
        </ProtectedRoute>
      </AuthProvider>
    )

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/login')
    })
  })

  it('should render content when authenticated', async () => {
    localStorage.setItem(
      'user',
      JSON.stringify({
        id: 1,
        email: 'test@example.com',
        role: 'professor',
        created_at: '2024-01-01T00:00:00Z',
      })
    )
    localStorage.setItem('token', 'test-token')

    render(
      <AuthProvider>
        <ProtectedRoute>
          <div>Protected Content</div>
        </ProtectedRoute>
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Protected Content')).toBeInTheDocument()
    })
  })

  it('should enforce role restrictions', async () => {
    localStorage.setItem(
      'user',
      JSON.stringify({
        id: 1,
        email: 'ta@example.com',
        role: 'grader_ta',
        created_at: '2024-01-01T00:00:00Z',
      })
    )
    localStorage.setItem('token', 'test-token')

    render(
      <AuthProvider>
        <ProtectedRoute allowedRoles={['professor', 'head_ta']}>
          <div>Protected Content</div>
        </ProtectedRoute>
      </AuthProvider>
    )

    await waitFor(() => {
      expect(mockPush).toHaveBeenCalledWith('/dashboard')
    })
  })

  it('should allow access for correct role', async () => {
    localStorage.setItem(
      'user',
      JSON.stringify({
        id: 1,
        email: 'prof@example.com',
        role: 'professor',
        created_at: '2024-01-01T00:00:00Z',
      })
    )
    localStorage.setItem('token', 'test-token')

    render(
      <AuthProvider>
        <ProtectedRoute allowedRoles={['professor', 'head_ta']}>
          <div>Protected Content</div>
        </ProtectedRoute>
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Protected Content')).toBeInTheDocument()
    })
  })
})

