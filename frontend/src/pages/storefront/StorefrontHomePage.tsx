import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { BookOpen } from 'lucide-react';
import { courseApi } from '../../api/courseApi';

export default function StorefrontHomePage() {
  const { data, isLoading } = useQuery({
    queryKey: ['storefront-courses'],
    queryFn: () => courseApi.listStorefront({ limit: 50 }),
  });

  const courses = data?.courses || [];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero */}
      <div className="bg-gradient-to-br from-indigo-600 to-purple-700 text-white py-20">
        <div className="max-w-4xl mx-auto text-center px-6">
          <h1 className="text-4xl md:text-5xl font-bold mb-4">Master New Skills</h1>
          <p className="text-lg text-indigo-100">Browse our catalog of expert-led courses</p>
        </div>
      </div>

      {/* Course grid */}
      <div className="max-w-7xl mx-auto px-6 py-12">
        <h2 className="text-2xl font-bold text-gray-800 mb-6">All Courses</h2>
        {isLoading ? (
          <div className="text-center py-12 text-gray-500">Loading courses...</div>
        ) : courses.length === 0 ? (
          <div className="text-center py-12">
            <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-3" />
            <p className="text-gray-500">No courses published yet</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {courses.map(course => (
              <Link
                key={course.id}
                to={`/courses/${course.slug}`}
                className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
              >
                <div className="h-40 bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center">
                  <BookOpen className="w-12 h-12 text-gray-300" />
                </div>
                <div className="p-5">
                  <h3 className="font-semibold text-gray-800 mb-1 line-clamp-2">{course.title}</h3>
                  <p className="text-sm text-gray-500 line-clamp-2 mb-3">{course.description || 'No description'}</p>
                  <div className="flex items-center justify-between">
                    <span className="text-lg font-bold text-gray-800">
                      {course.price_cents === 0 ? 'Free' : `$${(course.price_cents / 100).toFixed(2)}`}
                    </span>
                    {course.category && (
                      <span className="px-2 py-1 bg-indigo-50 text-indigo-600 text-xs rounded font-medium">
                        {course.category}
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
