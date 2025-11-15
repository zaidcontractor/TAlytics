'use client';

import { useEffect, useState } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Layout } from '@/components/Layout';
import { api, Submission } from '@/lib/api';

export default function GradingPage() {
  const [submissions, setSubmissions] = useState<Submission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedSubmission, setSelectedSubmission] = useState<Submission | null>(null);

  useEffect(() => {
    loadSubmissions();
  }, []);

  const loadSubmissions = async () => {
    try {
      const response = await api.getAssignedSubmissions();
      setSubmissions(response.submissions || []);
    } catch (err: any) {
      setError(err.message || 'Failed to load submissions');
      setSubmissions([]); // Set empty array on error
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'graded':
        return 'bg-green-100 text-green-800';
      case 'in_progress':
        return 'bg-yellow-100 text-yellow-800';
      case 'regrade_required':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <ProtectedRoute allowedRoles={['grader_ta', 'professor', 'head_ta']}>
      <Layout>
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-gray-900">Grading Workspace</h1>
            <p className="mt-2 text-sm text-gray-600">
              Review and grade assigned submissions
            </p>
          </div>

          {error && (
            <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {loading ? (
            <div className="flex justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          ) : !submissions || submissions.length === 0 ? (
            <div className="bg-white shadow rounded-lg p-12 text-center">
              <p className="text-gray-500">No submissions assigned to you yet</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-1">
                <div className="bg-white shadow rounded-lg">
                  <div className="p-4 border-b">
                    <h2 className="text-lg font-semibold text-gray-900">
                      Submissions ({submissions.length})
                    </h2>
                  </div>
                  <div className="divide-y">
                    {submissions.map((submission) => (
                      <button
                        key={submission.id}
                        onClick={() => setSelectedSubmission(submission)}
                        className={`w-full text-left p-4 hover:bg-gray-50 transition-colors ${
                          selectedSubmission?.id === submission.id ? 'bg-blue-50' : ''
                        }`}
                      >
                        <div className="flex justify-between items-start">
                          <div>
                            <p className="font-medium text-gray-900">
                              {submission.student_identifier}
                            </p>
                            <p className="text-sm text-gray-500 mt-1">
                              {submission.assignment_title}
                            </p>
                          </div>
                          <span
                            className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(
                              submission.graded_status
                            )}`}
                          >
                            {submission.graded_status}
                          </span>
                        </div>
                      </button>
                    ))}
                  </div>
                </div>
              </div>

              <div className="lg:col-span-2">
                {selectedSubmission ? (
                  <div className="bg-white shadow rounded-lg">
                    <div className="p-6">
                      <h2 className="text-xl font-semibold text-gray-900 mb-4">
                        {selectedSubmission.student_identifier}
                      </h2>
                      <div className="mb-4">
                        <h3 className="text-sm font-medium text-gray-700 mb-2">Submission</h3>
                        <div className="bg-gray-50 rounded-md p-4">
                          <pre className="whitespace-pre-wrap text-sm text-gray-900">
                            {selectedSubmission.text || 'No text submission'}
                          </pre>
                        </div>
                      </div>
                      <div className="border-t pt-4">
                        <p className="text-sm text-gray-600">
                          This is a simplified view. The full grading interface with Claude chat
                          and rubric panel will be implemented next.
                        </p>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="bg-white shadow rounded-lg p-12 text-center">
                    <p className="text-gray-500">Select a submission to view details</p>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </Layout>
    </ProtectedRoute>
  );
}

