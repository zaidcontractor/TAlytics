import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import GradingPage from '@/app/grading/page'
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
  usePathname: () => '/grading',
}))

const mockApi = api as jest.Mocked<typeof api>

describe('GradingPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })

  const renderWithAuth = async () => {
    const result = render(
      <AuthProvider>
        <GradingPage />
      </AuthProvider>
    )
    // Wait for auth to initialize
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100))
    })
    return result
  }

  it('should load and display assigned submissions', async () => {
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

    const mockSubmissions = [
      {
        id: 1,
        assignment_id: 1,
        student_identifier: 'student001',
        text: 'def factorial(n): return 1',
        file_path: '',
        graded_status: 'pending',
        assigned_ta_id: 1,
        created_at: '2024-01-01T00:00:00Z',
        assignment_title: 'Lab 1',
        course_name: 'CS 101',
      },
      {
        id: 2,
        assignment_id: 1,
        student_identifier: 'student002',
        text: 'def factorial(n): return 2',
        file_path: '',
        graded_status: 'in_progress',
        assigned_ta_id: 1,
        created_at: '2024-01-01T00:00:00Z',
        assignment_title: 'Lab 1',
        course_name: 'CS 101',
      },
    ]

    mockApi.getAssignedSubmissions.mockResolvedValueOnce({
      submissions: mockSubmissions,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText('student001')).toBeInTheDocument()
      expect(screen.getByText('student002')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should show empty state when no submissions', async () => {
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

    mockApi.getAssignedSubmissions.mockResolvedValueOnce({ submissions: [] })

    await renderWithAuth()

    await waitFor(() => {
      expect(
        screen.getByText(/no submissions assigned to you yet/i)
      ).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should display submission details when selected', async () => {
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

    const mockSubmissions = [
      {
        id: 1,
        assignment_id: 1,
        student_identifier: 'student001',
        text: 'def factorial(n): return 1 if n == 0 else n * factorial(n-1)',
        file_path: '',
        graded_status: 'pending',
        assigned_ta_id: 1,
        created_at: '2024-01-01T00:00:00Z',
        assignment_title: 'Lab 1',
        course_name: 'CS 101',
      },
    ]

    mockApi.getAssignedSubmissions.mockResolvedValueOnce({
      submissions: mockSubmissions,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText('student001')).toBeInTheDocument()
    }, { timeout: 5000 })

    const submissionButton = screen.getByText('student001')
    await act(async () => {
      await userEvent.click(submissionButton)
    })

    await waitFor(() => {
      expect(
        screen.getByText(/def factorial/i)
      ).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should display correct status badges', async () => {
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

    const mockSubmissions = [
      {
        id: 1,
        assignment_id: 1,
        student_identifier: 'student001',
        text: 'code',
        file_path: '',
        graded_status: 'graded',
        assigned_ta_id: 1,
        created_at: '2024-01-01T00:00:00Z',
        assignment_title: 'Lab 1',
        course_name: 'CS 101',
      },
      {
        id: 2,
        assignment_id: 1,
        student_identifier: 'student002',
        text: 'code',
        file_path: '',
        graded_status: 'regrade_required',
        assigned_ta_id: 1,
        created_at: '2024-01-01T00:00:00Z',
        assignment_title: 'Lab 1',
        course_name: 'CS 101',
      },
    ]

    mockApi.getAssignedSubmissions.mockResolvedValueOnce({
      submissions: mockSubmissions,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText('graded')).toBeInTheDocument()
      expect(screen.getByText('regrade_required')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should handle API errors', async () => {
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

    mockApi.getAssignedSubmissions.mockRejectedValueOnce(
      new Error('Failed to load submissions')
    )

    await renderWithAuth()

    await waitFor(() => {
      const errorText = screen.queryByText(/failed to load submissions/i) ||
                       screen.queryByText(/failed/i)
      expect(errorText).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should handle null submissions array', async () => {
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

    mockApi.getAssignedSubmissions.mockResolvedValueOnce({
      submissions: null as any,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(
        screen.getByText(/no submissions assigned to you yet/i)
      ).toBeInTheDocument()
    }, { timeout: 5000 })
  })
})
