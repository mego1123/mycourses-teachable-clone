import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { BookOpen, DollarSign, Users, TrendingUp, Plus } from 'lucide-react';
import { courseApi, enrollmentApi } from '../../api/courseApi';

export default function CreatorDashboardPage() {
  const { data: coursesData } = useQuery({
    queryKey: ['creator-courses'],
    queryFn: () => courseApi.listByCreator({ limit: 100 }),
  });

  const { data: enrollmentsData } = useQuery({
    queryKey: ['creator-enrollments'],
    queryFn: () => enrollmentApi.listByCreator({ limit: 100 }),
  });

  const courses = coursesData?.courses || [];
  const enrollments = enrollmentsData?.enrollments || [];
  const totalRevenue = enrollments
    .filter(e => e.status === 'active' || e.status === 'completed')
    .reduce((sum, e) => sum + e.price_paid_cents, 0);
  const totalEnrollments = enrollments.length;
  const publishedCourses = courses.filter(c => c.status === 'published').length;

  return (
    <div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard icon={DollarSign} label="Total Revenue" value={`$${(totalRevenue / 100).toFixed(2)}`} color="bg-green-500" />
        <StatCard icon={Users} label="Total Enrollments" value={totalEnrollments.toString()} color="bg-blue-500" />
        <StatCard icon={BookOpen} label="Published Courses" value={publishedCourses.toString()} color="bg-indigo-500" />
        <StatCard icon={TrendingUp} label="Total Courses" value={courses.length.toString()} color="bg-purple-500" />
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h3 className="font-semibold text-gray-800">Your Courses</h3>
          <Link to="/studio/courses" className="text-sm text-indigo-600 hover:text-indigo-700 font-medium">View all →</Link>
        </div>
        <div className="divide-y divide-gray-100">
          {courses.length === 0 ? (
            <div className="px-6 py-12 text-center">
              <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-3" />
              <p className="text-gray-500 mb-4">No courses yet</p>
              <Link to="/studio/courses" className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium">
                <Plus className="w-4 h-4" /> Create your first course
              </Link>
            </div>
          ) : (
            courses.slice(0, 5).map((course) => (
              <Link key={course.id} to={`/studio/courses/${course.id}/edit`} className="flex items-center justify-between px-6 py-4 hover:bg-gray-50">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
                    <BookOpen className="w-5 h-5 text-gray-400" />
                  </div>
                  <div>
                    <p className="font-medium text-gray-800">{course.title}</p>
                    <p className="text-sm text-gray-500">${(course.price_cents / 100).toFixed(2)}</p>
                  </div>
                </div>
                <span className={`px-2 py-1 rounded text-xs font-medium ${
                  course.status === 'published' ? 'bg-green-100 text-green-700' :
                  course.status === 'draft' ? 'bg-gray-100 text-gray-600' : 'bg-red-100 text-red-700'
                }`}>{course.status}</span>
              </Link>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function StatCard({ icon: Icon, label, value, color }: { icon: any; label: string; value: string; color: string }) {
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div className="flex items-center justify-between mb-4">
        <div className={`w-10 h-10 ${color} rounded-lg flex items-center justify-center`}>
          <Icon className="w-5 h-5 text-white" />
        </div>
      </div>
      <p className="text-2xl font-bold text-gray-800">{value}</p>
      <p className="text-sm text-gray-500 mt-1">{label}</p>
    </div>
  );
}
