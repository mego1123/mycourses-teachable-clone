import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Plus, BookOpen, Trash2, Globe, Eye } from 'lucide-react';
import { courseApi } from '../../api/courseApi';
import type { Course } from '../../types/course';

export default function CoursesPage() {
  const [showCreate, setShowCreate] = useState(false);
  const [newCourse, setNewCourse] = useState({ title: '', slug: '', description: '', priceCents: 0, currency: 'usd', category: '' });
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['creator-courses'],
    queryFn: () => courseApi.listByCreator({ limit: 100 }),
  });

  const createMutation = useMutation({
    mutationFn: (data: typeof newCourse) => courseApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['creator-courses'] });
      setShowCreate(false);
      setNewCourse({ title: '', slug: '', description: '', priceCents: 0, currency: 'usd', category: '' });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => courseApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['creator-courses'] }),
  });

  const publishMutation = useMutation({
    mutationFn: (id: string) => courseApi.publish(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['creator-courses'] }),
  });

  const unpublishMutation = useMutation({
    mutationFn: (id: string) => courseApi.unpublish(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['creator-courses'] }),
  });

  const courses = data?.courses || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800">Courses</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium"
        >
          <Plus className="w-4 h-4" /> New Course
        </button>
      </div>

      {showCreate && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
          <h3 className="font-semibold text-gray-800 mb-4">Create New Course</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
              <input
                type="text"
                value={newCourse.title}
                onChange={e => setNewCourse({ ...newCourse, title: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="My Awesome Course"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Slug</label>
              <input
                type="text"
                value={newCourse.slug}
                onChange={e => setNewCourse({ ...newCourse, slug: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="my-awesome-course"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Price (cents)</label>
              <input
                type="number"
                value={newCourse.priceCents}
                onChange={e => setNewCourse({ ...newCourse, priceCents: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="9900"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
              <input
                type="text"
                value={newCourse.category}
                onChange={e => setNewCourse({ ...newCourse, category: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                placeholder="programming"
              />
            </div>
            <div className="md:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea
                value={newCourse.description}
                onChange={e => setNewCourse({ ...newCourse, description: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
                rows={3}
                placeholder="A brief description of your course"
              />
            </div>
          </div>
          <div className="flex gap-3 mt-4">
            <button
              onClick={() => createMutation.mutate(newCourse)}
              disabled={!newCourse.title || !newCourse.slug}
              className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
            >
              Create Course
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm font-medium"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">Loading courses...</div>
      ) : courses.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <BookOpen className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">No courses yet. Click "New Course" to create one.</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Course</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Price</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {courses.map((course: Course) => (
                <tr key={course.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4">
                    <Link to={`/studio/courses/${course.id}/edit`} className="font-medium text-gray-800 hover:text-indigo-600">
                      {course.title}
                    </Link>
                    <p className="text-sm text-gray-500">/{course.slug}</p>
                  </td>
                  <td className="px-6 py-4 text-gray-600">${(course.price_cents / 100).toFixed(2)}</td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      course.status === 'published' ? 'bg-green-100 text-green-700' :
                      course.status === 'draft' ? 'bg-gray-100 text-gray-600' : 'bg-red-100 text-red-700'
                    }`}>{course.status}</span>
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      {course.status === 'draft' ? (
                        <button
                          onClick={() => publishMutation.mutate(course.id)}
                          className="p-1.5 text-green-600 hover:bg-green-50 rounded"
                          title="Publish"
                        >
                          <Globe className="w-4 h-4" />
                        </button>
                      ) : (
                        <button
                          onClick={() => unpublishMutation.mutate(course.id)}
                          className="p-1.5 text-gray-500 hover:bg-gray-100 rounded"
                          title="Unpublish"
                        >
                          <Eye className="w-4 h-4" />
                        </button>
                      )}
                      <Link
                        to={`/studio/courses/${course.id}/edit`}
                        className="p-1.5 text-indigo-600 hover:bg-indigo-50 rounded"
                        title="Edit"
                      >
                        <BookOpen className="w-4 h-4" />
                      </Link>
                      <button
                        onClick={() => { if (confirm('Delete this course?')) deleteMutation.mutate(course.id); }}
                        className="p-1.5 text-red-500 hover:bg-red-50 rounded"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
