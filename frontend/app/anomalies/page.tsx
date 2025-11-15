'use client';

import { useEffect, useState } from 'react';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Layout } from '@/components/Layout';
import { api, AnomalyReport } from '@/lib/api';

export default function AnomaliesPage() {
  const [reports, setReports] = useState<AnomalyReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    // In a real implementation, we'd fetch anomaly reports
    // For now, this is a placeholder
    setLoading(false);
  }, []);

  return (
    <ProtectedRoute allowedRoles={['professor', 'head_ta']}>
      <Layout>
        <div className="px-4 py-6 sm:px-0">
          <div className="mb-6">
            <h1 className="text-3xl font-bold text-gray-900">Anomaly Detection</h1>
            <p className="mt-2 text-sm text-gray-600">
              Review grading anomalies and statistical analysis
            </p>
          </div>

          {loading ? (
            <div className="flex justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          ) : (
            <div className="bg-white shadow rounded-lg p-6">
              <p className="text-gray-500">
                Anomaly detection dashboard will display:
              </p>
              <ul className="mt-4 list-disc list-inside text-sm text-gray-600 space-y-2">
                <li>TA severity deviation analysis</li>
                <li>Statistical outlier identification</li>
                <li>Criterion inconsistency reports</li>
                <li>Regrade risk predictions</li>
                <li>Visual charts and graphs</li>
              </ul>
            </div>
          )}
        </div>
      </Layout>
    </ProtectedRoute>
  );
}

