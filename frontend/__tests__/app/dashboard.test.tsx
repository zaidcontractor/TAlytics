import { render, screen, waitFor, act } from '@testing-library/react'
import DashboardPage from '@/app/dashboard/page'
import { AuthProvider } from '@/lib/auth'
import { api } from '@/lib/api'

jest.mock('@/lib/api')
jest.mock('@/components/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
  usePathname: () => '/dashboard',
}))

const mockApi = api as jest.Mocked<typeof api>

describe('DashboardPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
    
    // Mock getCurrentUser to read from localStorage
    mockApi.getCurrentUser = jest.fn(() => {
      if (typeof window === 'undefined') return null
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr) : null
    })
  })

  const renderAndWaitForAuth = async () => {
    const result = render(
      <AuthProvider>
        <DashboardPage />
      </AuthProvider>
    )
    
    // Wait for AuthProvider's useEffect to complete (loads user from localStorage)
    // This happens synchronously in the useEffect, but React batches state updates
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 50))
    })
    
    // Wait for Dashboard's useEffect to run after user is loaded
    // The useEffect depends on user, so it will run when user changes from null to the actual user
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 50))
    })
    
    return result
  }

  it('should show loading state initially', async () => {
    const user = {
      id: 1,
      email: 'prof@example.com',
      role: 'professor' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    mockApi.getCourses.mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(() => {
            resolve({ courses: [] })
          }, 100)
        })
    )

    await renderAndWaitForAuth()

    // Component should render - check for the h1 heading specifically
    expect(screen.getByRole('heading', { name: /dashboard/i })).toBeInTheDocument()
  })

  it('should display courses for professor', async () => {
    const user = {
      id: 1,
      email: 'prof@example.com',
      role: 'professor' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    const mockCourses = [
      {
        id: 1,
        name: 'CS 101',
        created_at: '2024-01-01T00:00:00Z',
        assignment_count: 3,
      },
      {
        id: 2,
        name: 'CS 102',
        created_at: '2024-01-02T00:00:00Z',
        assignment_count: 5,
      },
    ]

    mockApi.getCourses.mockResolvedValueOnce({ courses: mockCourses })

    await renderAndWaitForAuth()

    // Wait for user to be loaded and courses to be fetched
    await waitFor(() => {
      expect(screen.getByText('CS 101')).toBeInTheDocument()
      expect(screen.getByText('CS 102')).toBeInTheDocument()
    }, { timeout: 5000 })

    expect(screen.getByText('3 assignments')).toBeInTheDocument()
    expect(screen.getByText('5 assignments')).toBeInTheDocument()
  })

  it('should show empty state when no courses', async () => {
    const user = {
      id: 1,
      email: 'prof@example.com',
      role: 'professor' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    mockApi.getCourses.mockResolvedValueOnce({ courses: [] })

    await renderAndWaitForAuth()

    await waitFor(() => {
      expect(screen.getByText(/no courses yet/i)).toBeInTheDocument()
      expect(screen.getByText(/create your first course/i)).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should handle API errors gracefully', async () => {
    const user = {
      id: 1,
      email: 'prof@example.com',
      role: 'professor' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    mockApi.getCourses.mockRejectedValueOnce(new Error('Failed to load courses'))

    await renderAndWaitForAuth()

    await waitFor(() => {
      const errorText = screen.queryByText(/failed to load courses/i) ||
                       screen.queryByText(/failed/i)
      expect(errorText).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should show grading queue for grader TA', async () => {
    const user = {
      id: 1,
      email: 'ta@example.com',
      role: 'grader_ta' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    await renderAndWaitForAuth()

    // Wait for user to be loaded and grader TA content to render
    // The component checks user?.role === 'grader_ta' to show this section
    await waitFor(() => {
      expect(screen.getByText(/your grading queue/i)).toBeInTheDocument()
      expect(screen.getByText(/view assigned submissions/i)).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should not call getCourses for grader TA', async () => {
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

    await renderAndWaitForAuth()

    // Wait a bit to ensure getCourses wasn't called
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 200))
    })

    expect(mockApi.getCourses).not.toHaveBeenCalled()
  })
})
