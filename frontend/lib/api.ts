// API client for TAlytics backend
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface User {
  id: number;
  email: string;
  role: 'professor' | 'head_ta' | 'grader_ta';
  created_at: string;
}

export interface Course {
  id: number;
  name: string;
  created_at: string;
  assignment_count?: number;
}

export interface Assignment {
  id: number;
  course_id: number;
  title: string;
  description: string;
  status: 'draft' | 'open' | 'grading' | 'completed';
  created_at: string;
  rubric?: Rubric;
}

export interface Rubric {
  id: number;
  assignment_id: number;
  json_blob: string;
  max_points: number;
  created_at: string;
}

export interface Submission {
  id: number;
  assignment_id: number;
  student_identifier: string;
  text: string;
  file_path: string;
  graded_status: 'pending' | 'in_progress' | 'graded' | 'regrade_required';
  assigned_ta_id?: number;
  created_at: string;
  assignment_title?: string;
  course_name?: string;
  rubric_json?: string;
  rubric_max_points?: number;
}

export interface Grade {
  id: number;
  submission_id: number;
  score: number;
  feedback: string;
  rubric_breakdown: string;
  graded_by: number;
  created_at: string;
}

export interface AnomalyReport {
  assignment_id: number;
  total_grades: number;
  average_score: number;
  standard_deviation: number;
  ta_severity_issues: TASeverityAnomaly[];
  outlier_grades: OutlierAnomaly[];
  criterion_issues: CriterionAnomaly[];
  regrade_risks: RegradeRisk[];
  generated_at: string;
}

export interface TASeverityAnomaly {
  ta_id: number;
  ta_email: string;
  average_score: number;
  grades_count: number;
  deviation: number;
  severity: 'too_harsh' | 'too_lenient';
}

export interface OutlierAnomaly {
  submission_id: number;
  student_identifier: string;
  score: number;
  z_score: number;
  graded_by: number;
  grader_email: string;
}

export interface CriterionAnomaly {
  criterion_name: string;
  average_score: number;
  standard_deviation: number;
  inconsistent_submission_ids: number[];
}

export interface RegradeRisk {
  submission_id: number;
  student_identifier: string;
  score: number;
  risk_score: number;
  risk_factors: string[];
  graded_by: number;
  grader_email: string;
}

// API Client
class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(error.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  // Auth endpoints
  async register(email: string, password: string, role: string) {
    return this.request<{ user: User }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, role }),
    });
  }

  async login(email: string, password: string) {
    const response = await this.request<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    
    if (typeof window !== 'undefined' && response.token) {
      localStorage.setItem('token', response.token);
      localStorage.setItem('user', JSON.stringify(response.user));
    }
    
    return response;
  }

  logout() {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
    }
  }

  getCurrentUser(): User | null {
    if (typeof window === 'undefined') return null;
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  }

  // Course endpoints
  async getCourses() {
    return this.request<{ courses: Course[] }>('/courses');
  }

  async createCourse(name: string) {
    return this.request<{ course_id: number; name: string; created_by: number; created_at: string }>('/courses', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  // Assignment endpoints
  async getAssignmentsByCourse(courseId: number) {
    return this.request<{ assignments: Assignment[] }>(`/assignments/course/${courseId}`);
  }

  async getAssignment(id: number) {
    return this.request<Assignment>(`/assignments/${id}`);
  }

  async createAssignment(courseId: number, title: string, description?: string) {
    return this.request<{ assignment_id: number; course_id: number; title: string; description: string; status: string; created_at: string }>('/assignments', {
      method: 'POST',
      body: JSON.stringify({ course_id: courseId, title, description }),
    });
  }

  async notifyTAs(assignmentId: number) {
    return this.request<{ assignment_id: number; total_submissions: number; distribution: Record<number, number>; message: string }>(`/assignments/${assignmentId}/notify-tas`, {
      method: 'POST',
    });
  }

  // Rubric endpoints
  async createRubric(assignmentId: number, jsonBlob: string, maxPoints: number) {
    return this.request<{ rubric_id: number; assignment_id: number; max_points: number; message: string }>('/rubrics', {
      method: 'POST',
      body: JSON.stringify({ assignment_id: assignmentId, json_blob: jsonBlob, max_points: maxPoints }),
    });
  }

  async uploadRubricPDF(assignmentId: number, file: File) {
    const formData = new FormData();
    formData.append('assignment_id', assignmentId.toString());
    formData.append('file', file);

    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    
    const headers: Record<string, string> = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(`${this.baseUrl}/rubrics/upload`, {
      method: 'POST',
      headers,
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(error.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  // Submission endpoints
  async getAssignedSubmissions() {
    return this.request<{ submissions: Submission[] }>('/submissions/assigned');
  }

  async uploadSubmission(assignmentId: number, studentIdentifier: string, text?: string, file?: File) {
    if (file) {
      const formData = new FormData();
      formData.append('assignment_id', assignmentId.toString());
      formData.append('student_identifier', studentIdentifier);
      formData.append('file', file);

      const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
      
      const headers: Record<string, string> = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch(`${this.baseUrl}/submissions/upload`, {
        method: 'POST',
        headers,
        body: formData,
      });

      if (!response.ok) {
        const error = await response.json().catch(() => ({ error: response.statusText }));
        throw new Error(error.error || `HTTP error! status: ${response.status}`);
      }

      return response.json();
    } else {
      return this.request<{ submission_id: number; assignment_id: number; student_identifier: string; graded_status: string; created_at: string }>('/submissions/upload', {
        method: 'POST',
        body: JSON.stringify({ assignment_id: assignmentId, student_identifier: studentIdentifier, text }),
      });
    }
  }

  // Grading endpoints
  async gradeSubmission(submissionId: number, options?: {
    score?: number;
    feedback?: string;
    rubricBreakdown?: string;
    useClaudeRecommendation?: boolean;
  }) {
    return this.request<{ message: string; grade_id?: number; claude_recommendation?: any }>('/grade', {
      method: 'POST',
      body: JSON.stringify({
        submission_id: submissionId,
        ...options,
      }),
    });
  }

  async batchGrade(assignmentId: number, autoApprove: boolean = false) {
    return this.request<{ assignment_id: number; total_graded: number; average_score: number; graded_submission_ids: number[]; message: string }>('/grade/batch', {
      method: 'POST',
      body: JSON.stringify({ assignment_id: assignmentId, auto_approve: autoApprove }),
    });
  }

  // Anomaly endpoints
  async analyzeAnomalies(assignmentId: number) {
    return this.request<{ message: string; report_id: number; summary: AnomalyReport }>(`/analyze/${assignmentId}`, {
      method: 'POST',
    });
  }

  async getAnomalies(assignmentId: number) {
    return this.request<{ report: AnomalyReport; status: string }>(`/anomalies/${assignmentId}`);
  }
}

export const api = new ApiClient();

