import { api } from '@/lib/api'

// Mock fetch globally
global.fetch = jest.fn()

describe('API Client', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    localStorage.clear()
  })

  describe('Authentication', () => {
    it('should register a new user', async () => {
      const mockResponse = {
        user: {
          id: 1,
          email: 'test@example.com',
          role: 'professor',
          created_at: '2024-01-01T00:00:00Z',
        },
      }

      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      const result = await api.register('test@example.com', 'password123', 'professor')

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/auth/register'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({
            email: 'test@example.com',
            password: 'password123',
            role: 'professor',
          }),
        })
      )
      expect(result.user.email).toBe('test@example.com')
    })

    it('should login and store token', async () => {
      const mockResponse = {
        token: 'test-token-123',
        user: {
          id: 1,
          email: 'test@example.com',
          role: 'professor',
          created_at: '2024-01-01T00:00:00Z',
        },
      }

      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      await api.login('test@example.com', 'password123')

      expect(localStorage.getItem('token')).toBe('test-token-123')
      expect(localStorage.getItem('user')).toBeTruthy()
    })

    it('should logout and clear storage', () => {
      localStorage.setItem('token', 'test-token')
      localStorage.setItem('user', '{}')

      api.logout()

      expect(localStorage.getItem('token')).toBeNull()
      expect(localStorage.getItem('user')).toBeNull()
    })

    it('should get current user from localStorage', () => {
      const user = {
        id: 1,
        email: 'test@example.com',
        role: 'professor' as const,
        created_at: '2024-01-01T00:00:00Z',
      }
      localStorage.setItem('user', JSON.stringify(user))

      const currentUser = api.getCurrentUser()

      expect(currentUser).toEqual(user)
    })

    it('should return null if no user in localStorage', () => {
      expect(api.getCurrentUser()).toBeNull()
    })
  })

  describe('Courses', () => {
    it('should fetch courses with auth token', async () => {
      localStorage.setItem('token', 'test-token')
      const mockCourses = [
        { id: 1, name: 'CS 101', created_at: '2024-01-01', assignment_count: 2 },
      ]

      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ courses: mockCourses }),
      })

      const result = await api.getCourses()

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/courses'),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer test-token',
          }),
        })
      )
      expect(result.courses).toEqual(mockCourses)
    })

    it('should create a course', async () => {
      localStorage.setItem('token', 'test-token')
      const mockResponse = {
        course_id: 1,
        name: 'CS 101',
        created_by: 1,
        created_at: '2024-01-01T00:00:00Z',
      }

      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      const result = await api.createCourse('CS 101')

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/courses'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'CS 101' }),
        })
      )
      expect(result.course_id).toBe(1)
    })
  })

  describe('Assignments', () => {
    it('should get assignments by course', async () => {
      localStorage.setItem('token', 'test-token')
      const mockAssignments = [
        {
          id: 1,
          course_id: 1,
          title: 'Lab 1',
          description: 'First lab',
          status: 'draft',
          created_at: '2024-01-01T00:00:00Z',
        },
      ]

      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ assignments: mockAssignments }),
      })

      const result = await api.getAssignmentsByCourse(1)

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/assignments/course/1'),
        expect.any(Object)
      )
      expect(result.assignments).toEqual(mockAssignments)
    })

    it('should create an assignment', async () => {
      localStorage.setItem('token', 'test-token')
      const mockResponse = {
        assignment_id: 1,
        course_id: 1,
        title: 'Lab 1',
        description: 'First lab',
        status: 'draft',
        created_at: '2024-01-01T00:00:00Z',
      }

      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      })

      const result = await api.createAssignment(1, 'Lab 1', 'First lab')

      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining('/assignments'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({
            course_id: 1,
            title: 'Lab 1',
            description: 'First lab',
          }),
        })
      )
      expect(result.assignment_id).toBe(1)
    })
  })

  describe('Error Handling', () => {
    it('should throw error on API failure', async () => {
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: 'Bad request' }),
      })

      await expect(api.getCourses()).rejects.toThrow('Bad request')
    })

    it('should handle network errors', async () => {
      ;(fetch as jest.Mock).mockRejectedValueOnce(new Error('Network error'))

      await expect(api.getCourses()).rejects.toThrow('Network error')
    })

    it('should handle 401 unauthorized', async () => {
      localStorage.setItem('token', 'invalid-token')
      ;(fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ error: 'Unauthorized' }),
      })

      await expect(api.getCourses()).rejects.toThrow('Unauthorized')
    })
  })
})

