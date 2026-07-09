import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { BookOpen, Award, Clock } from 'lucide-react';
import { enrollmentApi, certificateApi } from '../../api/courseApi';

export default function MyCoursesPage() {
  const { data: enrollmentsData, isLoading } = useQuery({
    queryKey: ['my-enrollments'],
    queryFn: () => enrollmentApi.listMine({ limit: 100 }),
  });

  const { data: certificates } = useQuery({
    queryKey: ['my-certificates'],
    queryFn: () => certificateApi.listMine(),
  });

  const enrollments = enrollmentsData?.enrollments || [];

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-5xl mx-auto px-6">
        <h1 className="text-2xl font-bold text-gray-800 mb-6">My Courses</h1>

        {isLoading ? (
          <div className="text-center py-12 text-gray-500">Loading...</div>
        ) : enrollments.length === 0 ? (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
            <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-3" />
            <p className="text-gray-500 mb-4">You haven't enrolled in any courses yet</p>
            <Link to="/courses" className="text-indigo-600 hover:text-indigo-700 font-medium">Browse courses →</Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {enrollments.map(enrollment => (
              <div key={enrollment.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 bg-indigo-50 rounded-lg flex items-center justify-center">
                      <BookOpen className="w-6 h-6 text-indigo-500" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-800">Course</p>
                      <p className="text-xs text-gray-400 font-mono">{enrollment.course_id.slice(0, 8)}...</p>
                    </div>
                  </div>
                  <span className={`px-2 py-1 rounded text-xs font-medium ${
                    enrollment.status === 'active' ? 'bg-green-100 text-green-700' :
                    enrollment.status === 'completed' ? 'bg-blue-100 text-blue-700' :
                    'bg-gray-100 text-gray-600'
                  }`}>{enrollment.status}</span>
                </div>

                <div className="flex items-center gap-4 text-sm text-gray-500 mb-4">
                  <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" /> Enrolled {new Date(enrollment.enrolled_at).toLocaleDateString()}
                  </span>
                </div>

                <Link
                  to={`/learn/course/${enrollment.course_id}`}
                  className="block w-full text-center px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium"
                >
                  {enrollment.status === 'completed' ? 'Review Course' : 'Continue Learning'}
                </Link>
              </div>
            ))}
          </div>
        )}

        {/* Certificates */}
        {certificates && certificates.length > 0 && (
          <div className="mt-12">
            <h2 className="text-xl font-bold text-gray-800 mb-4">My Certificates</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {certificates.map(cert => (
                <div key={cert.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 flex items-center gap-4">
                  <Award className="w-10 h-10 text-yellow-500" />
                  <div className="flex-1">
                    <p className="font-medium text-gray-800">{cert.course_title}</p>
                    <p className="text-sm text-gray-500">{cert.certificate_number}</p>
                  </div>
                  <Link
                    to={`/certificates/verify/${cert.verification_token}`}
                    className="text-indigo-600 hover:text-indigo-700 text-sm font-medium"
                  >
                    View
                  </Link>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
