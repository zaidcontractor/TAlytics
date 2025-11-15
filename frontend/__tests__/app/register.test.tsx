import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import RegisterPage from '@/app/register/page'
import { AuthProvider } from '@/lib/auth'
import { api } from '@/lib/api'

jest.mock('@/lib/api')
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
  }),
}))

const mockApi = api as jest.Mocked<typeof api>

describe('RegisterPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })

  it('should render registration form', () => {
    render(
      <AuthProvider>
        <RegisterPage />
      </AuthProvider>
    )

    expect(screen.getByLabelText(/email address/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/role/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create account/i })).toBeInTheDocument()
  })

  it('should validate password length', async () => {
    render(
      <AuthProvider>
        <RegisterPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await userEvent.type(emailInput, 'test@example.com')
    await userEvent.type(passwordInput, 'short')
    await userEvent.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/password must be at least 8 characters/i)).toBeInTheDocument()
    }, { timeout: 2000 })
  })

  it('should successfully register a new user', async () => {
    mockApi.register.mockResolvedValueOnce({
      user: {
        id: 1,
        email: 'test@example.com',
        role: 'professor',
        created_at: '2024-01-01T00:00:00Z',
      },
    })
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
        <RegisterPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const roleSelect = screen.getByLabelText(/role/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await userEvent.type(emailInput, 'test@example.com')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.selectOptions(roleSelect, 'professor')
    await userEvent.click(submitButton)

    await waitFor(() => {
      expect(mockApi.register).toHaveBeenCalledWith(
        'test@example.com',
        'password123',
        'professor'
      )
    })
  })

  it('should show error on registration failure', async () => {
    mockApi.register.mockRejectedValueOnce(new Error('User with this email already exists'))

    render(
      <AuthProvider>
        <RegisterPage />
      </AuthProvider>
    )

    const emailInput = screen.getByLabelText(/email address/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /create account/i })

    await userEvent.type(emailInput, 'existing@example.com')
    await userEvent.type(passwordInput, 'password123')
    await userEvent.click(submitButton)

    // Wait for error to be displayed - the error is set in catch block
    await waitFor(() => {
      // The error message from the catch block: err.message || 'Registration failed. Please try again.'
      const errorText = screen.queryByText(/user with this email already exists/i) ||
                       screen.queryByText(/registration failed/i) ||
                       screen.queryByText(/please try again/i)
      expect(errorText).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('should allow role selection', async () => {
    render(
      <AuthProvider>
        <RegisterPage />
      </AuthProvider>
    )

    const roleSelect = screen.getByLabelText(/role/i)

    expect(roleSelect).toHaveValue('grader_ta') // Default

    await userEvent.selectOptions(roleSelect, 'professor')
    expect(roleSelect).toHaveValue('professor')

    await userEvent.selectOptions(roleSelect, 'head_ta')
    expect(roleSelect).toHaveValue('head_ta')
  })
})
