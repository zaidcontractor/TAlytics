import { render, screen, waitFor, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import CoursesPage from '@/app/courses/page'
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
  usePathname: () => '/courses',
}))

const mockApi = api as jest.Mocked<typeof api>

describe('CoursesPage', () => {
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
        <CoursesPage />
      </AuthProvider>
    )
    // Wait for auth to initialize
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 100))
    })
    return result
  }

  it('should load and display courses', async () => {
    const mockCourses = [
      {
        id: 1,
        name: 'CS 101',
        created_at: '2024-01-01T00:00:00Z',
        assignment_count: 2,
      },
    ]

    mockApi.getCourses.mockResolvedValueOnce({ courses: mockCourses })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText('CS 101')).toBeInTheDocument()
    }, { timeout: 5000 })
  })

  it('should open create course modal', async () => {
    mockApi.getCourses.mockResolvedValueOnce({ courses: [] })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /new course/i })).toBeInTheDocument()
    }, { timeout: 5000 })

    const createButton = screen.getByRole('button', { name: /new course/i })
    await act(async () => {
      await userEvent.click(createButton)
    })

    await waitFor(() => {
      expect(screen.getByText(/create new course/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/course name/i)).toBeInTheDocument()
    })
  })

  it('should create a new course', async () => {
    mockApi.getCourses
      .mockResolvedValueOnce({ courses: [] })
      .mockResolvedValueOnce({
        courses: [
          {
            id: 1,
            name: 'CS 101',
            created_at: '2024-01-01T00:00:00Z',
            assignment_count: 0,
          },
        ],
      })

    mockApi.createCourse.mockResolvedValueOnce({
      course_id: 1,
      name: 'CS 101',
      created_by: 1,
      created_at: '2024-01-01T00:00:00Z',
    })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /new course/i })).toBeInTheDocument()
    }, { timeout: 5000 })

    // Open modal
    const createButton = screen.getByRole('button', { name: /new course/i })
    await act(async () => {
      await userEvent.click(createButton)
    })

    await waitFor(() => {
      expect(screen.getByLabelText(/course name/i)).toBeInTheDocument()
    })

    // Fill form
    const nameInput = screen.getByLabelText(/course name/i)
    await act(async () => {
      await userEvent.type(nameInput, 'CS 101')
    })

    // Submit - find the submit button in the modal
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
      // Fallback
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
      expect(mockApi.createCourse).toHaveBeenCalledWith('CS 101')
    }, { timeout: 5000 })
  })

  it('should handle create course error', async () => {
    mockApi.getCourses.mockResolvedValueOnce({ courses: [] })
    mockApi.createCourse.mockRejectedValueOnce(new Error('Course creation failed'))

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /new course/i })).toBeInTheDocument()
    }, { timeout: 5000 })

    const createButton = screen.getByRole('button', { name: /new course/i })
    await act(async () => {
      await userEvent.click(createButton)
    })

    await waitFor(() => {
      expect(screen.getByLabelText(/course name/i)).toBeInTheDocument()
    })

    const nameInput = screen.getByLabelText(/course name/i)
    await act(async () => {
      await userEvent.type(nameInput, 'CS 101')
    })

    // Submit - find the submit button in the modal
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
      // Fallback
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
      const errorText = screen.queryByText(/course creation failed/i) ||
                       screen.queryByText(/failed/i)
      expect(errorText).toBeInTheDocument()
    }, { timeout: 10000 })
  })

  it('should close modal on cancel', async () => {
    mockApi.getCourses.mockResolvedValueOnce({ courses: [] })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /new course/i })).toBeInTheDocument()
    }, { timeout: 5000 })

    const createButton = screen.getByRole('button', { name: /new course/i })
    await act(async () => {
      await userEvent.click(createButton)
    })

    await waitFor(() => {
      expect(screen.getByText(/create new course/i)).toBeInTheDocument()
    })

    const cancelButton = screen.getByRole('button', { name: /cancel/i })
    await act(async () => {
      await userEvent.click(cancelButton)
    })

    await waitFor(() => {
      expect(screen.queryByText(/create new course/i)).not.toBeInTheDocument()
    })
  })

  it('should handle null courses array', async () => {
    mockApi.getCourses.mockResolvedValueOnce({ courses: null as any })

    await renderWithAuth()

    await waitFor(() => {
      expect(screen.getByText(/no courses yet/i)).toBeInTheDocument()
    }, { timeout: 5000 })
  })
})
