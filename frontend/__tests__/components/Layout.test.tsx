import { render, screen, waitFor } from '@testing-library/react'
import { Layout } from '@/components/Layout'
import { AuthProvider } from '@/lib/auth'

jest.mock('next/navigation', () => ({
  usePathname: () => '/dashboard',
  useRouter: () => ({
    push: jest.fn(),
  }),
}))

describe('Layout', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('should render navigation for professor', () => {
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
        <Layout>
          <div>Content</div>
        </Layout>
      </AuthProvider>
    )

    expect(screen.getByText('TAlytics')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Courses')).toBeInTheDocument()
    expect(screen.getByText('Grading')).toBeInTheDocument()
    expect(screen.getByText('Anomalies')).toBeInTheDocument()
  })

  it('should show correct navigation for grader TA', () => {
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
        <Layout>
          <div>Content</div>
        </Layout>
      </AuthProvider>
    )

    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Grading')).toBeInTheDocument()
    expect(screen.queryByText('Courses')).not.toBeInTheDocument()
    expect(screen.queryByText('Anomalies')).not.toBeInTheDocument()
  })

  it('should display user email and role', () => {
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
        <Layout>
          <div>Content</div>
        </Layout>
      </AuthProvider>
    )

    expect(screen.getByText(/prof@example.com/i)).toBeInTheDocument()
    expect(screen.getByText(/professor/i)).toBeInTheDocument()
  })

  it('should handle logout', async () => {
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
        <Layout>
          <div>Content</div>
        </Layout>
      </AuthProvider>
    )

    const logoutButton = screen.getByText('Logout')
    logoutButton.click()

    // Logout should clear storage
    await waitFor(() => {
      expect(localStorage.getItem('token')).toBeNull()
    })
  })
})

