import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CourseDetailPage from '@/app/courses/[id]/page'
import { AuthProvider } from '@/lib/auth'
import { api } from '@/lib/api'

jest.mock('@/lib/api')
jest.mock('@/components/ProtectedRoute', () => ({
  ProtectedRoute: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    back: jest.fn(),
  }),
  useParams: () => ({ id: '1' }),
  usePathname: () => '/courses/1',
}))

const mockApi = api as jest.Mocked<typeof api>

describe('CourseDetailPage', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
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
  })

  const renderWithAuth = async () => {
    const result = render(
      <AuthProvider>
        <CourseDetailPage />
      </AuthProvider>
    )
    // Wait for auth to initialize
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100))
    })
    return result
  }

  it('should load and display assignments', async () => {
    const mockAssignments = [
      {
        id: 1,
        course_id: 1,
        title: 'Lab 1',
        description: 'First lab',
        status: 'draft',
        created_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        course_id: 1,
        title: 'Lab 2',
        description: 'Second lab',
        status: 'open',
        created_at: '2024-01-02T00:00:00Z',
      },
    ]

    mockApi.getAssignmentsByCourse.mockResolvedValueOnce({
      assignments: mockAssignments,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText('Lab 1')).toBeInTheDocument()
      expect(screen.getByText('Lab 2')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should show empty state when no assignments', async () => {
    mockApi.getAssignmentsByCourse.mockResolvedValueOnce({ assignments: [] })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText(/no assignments yet/i)).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should create a new assignment', async () => {
    mockApi.getAssignmentsByCourse
      .mockResolvedValueOnce({ assignments: [] })
      .mockResolvedValueOnce({
        assignments: [
          {
            id: 1,
            course_id: 1,
            title: 'Lab 1',
            description: 'First lab',
            status: 'draft',
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
      })

    mockApi.createAssignment.mockResolvedValueOnce({
      assignment_id: 1,
      course_id: 1,
      title: 'Lab 1',
      description: 'First lab',
      status: 'draft',
      created_at: '2024-01-01T00:00:00Z',
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /new assignment/i })).toBeInTheDocument()
    }, { timeout: 5000 })

    // Open modal
    const createButton = screen.getByRole('button', { name: /new assignment/i })
    await act(async () => {
      await userEvent.click(createButton)
    })

    await waitFor(() => {
      expect(screen.getByLabelText(/title/i)).toBeInTheDocument()
    })

    // Fill form
    const titleInput = screen.getByLabelText(/title/i)
    const descriptionInput = screen.getByLabelText(/description/i)
    await act(async () => {
      await userEvent.type(titleInput, 'Lab 1')
      await userEvent.type(descriptionInput, 'First lab')
    })

    // Submit - find the submit button in the modal by looking for button with type="submit" in the form
    await waitFor(() => {
      const forms = document.querySelectorAll('form')
      expect(forms.length).toBeGreaterThan(0)
    })
    const forms = document.querySelectorAll('form')
    const modalForm = forms[forms.length - 1] as HTMLFormElement
    const submitButton = modalForm.querySelector('button[type="submit"]') as HTMLButtonElement
    if (submitButton) {
      await act(async () => {
        await userEvent.click(submitButton)
      })
    } else {
      // Fallback: find by text "Create" that's not disabled
      const buttons = screen.getAllByRole('button')
      const createBtn = buttons.find(btn => 
        btn.textContent?.toLowerCase().trim() === 'create' &&
        !btn.disabled
      )
      if (createBtn) {
        await act(async () => {
          await userEvent.click(createBtn)
        })
      }
    }

    await waitFor(() => {
      expect(mockApi.createAssignment).toHaveBeenCalledWith(1, 'Lab 1', 'First lab')
    }, { timeout: 5000 })
  })

  it('should display assignment status badges correctly', async () => {
    const mockAssignments = [
      {
        id: 1,
        course_id: 1,
        title: 'Draft Assignment',
        description: '',
        status: 'draft',
        created_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        course_id: 1,
        title: 'Grading Assignment',
        description: '',
        status: 'grading',
        created_at: '2024-01-02T00:00:00Z',
      },
    ]

    mockApi.getAssignmentsByCourse.mockResolvedValueOnce({
      assignments: mockAssignments,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText('draft')).toBeInTheDocument()
      expect(screen.getByText('grading')).toBeInTheDocument()
    }, { timeout: 10000 })
  })

  it('should handle API errors', async () => {
    mockApi.getAssignmentsByCourse.mockRejectedValueOnce(
      new Error('Failed to load assignments')
    )

    await renderWithAuth()

    await waitFor(() => {
      const errorText = screen.queryByText(/failed to load assignments/i) ||
                       screen.queryByText(/failed/i)
      expect(errorText).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should handle null assignments array', async () => {
    mockApi.getAssignmentsByCourse.mockResolvedValueOnce({
      assignments: null as any,
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText(/no assignments yet/i)).toBeInTheDocument()
    }, { timeout: 5000 })
  })
})
