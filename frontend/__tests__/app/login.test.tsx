import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LoginPage from '@/app/login/page'
import { AuthProvider } from '@/lib/auth'
import { api } from '@/lib/api'

jest.mock('@/lib/api')
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}))

const mockApi = api as jest.Mocked<typeof api>

describe('LoginPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })

  it('should render login form', () => {
    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    )

    expect(screen.getByLabelText(/email address/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('should show error on invalid credentials', async () => {
    mockApi.login.mockRejectedValueOnce(new Error('Invalid email or password'))

    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await userEvent.type(emailInput, 'wrong@example.com')
    await userEvent.type(passwordInput, 'wrongpassword')
    await userEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/invalid email or password/i)).toBeInTheDocument()
    })
  })

  it('should successfully login with valid credentials', async () => {
    mockApi.login.mockResolvedValueOnce({
      token: 'test-token',
      user: {
        id: 1,
        email: 'test@example.com',
        role: 'professor',
        created_at: '2024-01-01T00:00:00Z',
      },
    })

    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await userEvent.type(emailInput, 'test@example.com')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    await waitFor(() => {
      expect(mockApi.login).toHaveBeenCalledWith('test@example.com', 'password123')
    })
  })

  it('should validate email format', async () => {
    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await userEvent.type(emailInput, 'invalid-email')
    await userEvent.click(submitButton)

    // HTML5 validation should prevent submission
    expect(emailInput).toBeInvalid()
  })

  it('should require password', async () => {
    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await userEvent.type(emailInput, 'test@example.com')
    await userEvent.click(submitButton)

    // HTML5 validation should prevent submission
    const passwordInput = screen.getByLabelText(/password/i)
    expect(passwordInput).toBeInvalid()
  })

  it('should show loading state during login', async () => {
    mockApi.login.mockImplementation(
      () =>
        new Promise((resolve) => {
          setTimeout(() => {
            resolve({
              token: 'test-token',
              user: {
                id: 1,
                email: 'test@example.com',
                role: 'professor',
                created_at: '2024-01-01T00:00:00Z',
              },
            })
          }, 100)
        })
    )

    render(
      <AuthProvider>
        <LoginPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await userEvent.type(emailInput, 'test@example.com')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    expect(screen.getByText(/signing in/i)).toBeInTheDocument()
    expect(submitButton).toBeDisabled()
  })
})

