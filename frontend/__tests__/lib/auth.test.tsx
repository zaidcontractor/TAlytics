import { render, screen, waitFor, act } from '@testing-library/react'
import { AuthProvider, useAuth } from '@/lib/auth'
import { api } from '@/lib/api'

jest.mock('@/lib/api')
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}))

const mockApi = api as jest.Mocked<typeof api>

function TestComponent() {
  const { user, isAuthenticated, loading } = useAuth()
  return (
    <div>
      {loading && <div>Loading...</div>}
      {isAuthenticated && <div>Authenticated: {user?.email}</div>}
      {!isAuthenticated && !loading && <div>Not authenticated</div>}
    </div>
  )
}

describe('AuthProvider', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
    
    // Mock getCurrentUser to read from localStorage
    mockApi.getCurrentUser = jest.fn(() => {
      if (typeof window === 'undefined') return null
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr) : null
    })
    
    // Mock logout to actually clear localStorage
    mockApi.logout = jest.fn(() => {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
      }
    })
  })

  it('should load user from localStorage on mount', async () => {
    const user = {
      id: 1,
      email: 'test@example.com',
      role: 'professor' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    // Wait for AuthProvider's useEffect to load user from localStorage
    await waitFor(() => {
      expect(screen.getByText(/authenticated: test@example.com/i)).toBeInTheDocument()
    }, { timeout: 2000 })
  })

  it('should show not authenticated when no user', async () => {
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByText(/not authenticated/i)).toBeInTheDocument()
    }, { timeout: 2000 })
  })

  it('should login and update user state', async () => {
    const mockResponse = {
      token: 'test-token',
      user: {
        id: 1,
        email: 'test@example.com',
        role: 'professor' as const,
        created_at: '2024-01-01T00:00:00Z',
      },
    }

    mockApi.login.mockResolvedValueOnce(mockResponse)

    function LoginTest() {
      const { login } = useAuth()

      const handleLogin = async () => {
        await login('test@example.com', 'password123')
      }

      return (
        <div>
          <button onClick={handleLogin}>Login</button>
        </div>
      )
    }

    render(
      <AuthProvider>
        <LoginTest />
        <TestComponent />
      </AuthProvider>
    )

    const loginButton = screen.getByText('Login')
    await act(async () => {
      loginButton.click()
    })

    await waitFor(() => {
      expect(mockApi.login).toHaveBeenCalledWith('test@example.com', 'password123')
    })
  })

  it('should logout and clear user state', async () => {
    const user = {
      id: 1,
      email: 'test@example.com',
      role: 'professor' as const,
      created_at: '2024-01-01T00:00:00Z',
    }
    localStorage.setItem('user', JSON.stringify(user))
    localStorage.setItem('token', 'test-token')

    function LogoutTest() {
      const { logout } = useAuth()
      return <button onClick={logout}>Logout</button>
    }

    render(
      <AuthProvider>
        <LogoutTest />
        <TestComponent />
      </AuthProvider>
    )

    // Wait for auth to initialize
    await waitFor(() => {
      expect(screen.getByText(/authenticated/i)).toBeInTheDocument()
    }, { timeout: 2000 })

    const logoutButton = screen.getByText('Logout')
    
    // Click logout - api.logout() clears localStorage synchronously
    await act(async () => {
      logoutButton.click()
    })

    // api.logout() clears localStorage immediately
    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })
})
