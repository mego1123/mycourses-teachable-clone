import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { Save, Plus, Trash2, GripVertical, FileText, Video } from 'lucide-react';
import { courseApi, sectionApi, lessonApi } from '../../api/courseApi';

export default function CourseEditorPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'details' | 'curriculum' | 'pricing'>('details');
  const [newSectionTitle, setNewSectionTitle] = useState('');
  const [newLessonTitle, setNewLessonTitle] = useState<Record<string, string>>({});

  const { data: course, isLoading } = useQuery({
    queryKey: ['course', id],
    queryFn: () => courseApi.get(id!),
    enabled: !!id,
  });

  const { data: sections } = useQuery({
    queryKey: ['course-sections', id],
    queryFn: () => sectionApi.listByCourse(id!),
    enabled: !!id,
  });

  const updateMutation = useMutation({
    mutationFn: (data: any) => courseApi.update(id!, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['course', id] }),
  });

  const createSectionMutation = useMutation({
    mutationFn: (title: string) => sectionApi.create(id!, { title, sortOrder: (sections?.length || 0) }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['course-sections', id] });
      setNewSectionTitle('');
    },
  });

  const deleteSectionMutation = useMutation({
    mutationFn: (sectionId: string) => sectionApi.delete(sectionId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['course-sections', id] }),
  });

  const createLessonMutation = useMutation({
    mutationFn: ({ sectionId, title }: { sectionId: string; title: string }) =>
      lessonApi.create(sectionId, { title, type: 'video', sortOrder: 0 }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['course-lessons', id] });
      setNewLessonTitle({});
    },
  });

  if (isLoading) return <div className="text-center py-12 text-gray-500">Loading course...</div>;
  if (!course) return <div className="text-center py-12 text-gray-500">Course not found</div>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800">{course.title}</h1>
          <p className="text-sm text-gray-500 mt-1">
            <span className={`px-2 py-0.5 rounded text-xs font-medium ${
              course.status === 'published' ? 'bg-green-100 text-green-700' :
              course.status === 'draft' ? 'bg-gray-100 text-gray-600' : 'bg-red-100 text-red-700'
            }`}>{course.status}</span>
            <span className="ml-2">/{course.slug}</span>
          </p>
        </div>
        <button
          onClick={() => navigate('/studio/courses')}
          className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm font-medium"
        >
          ← Back to Courses
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-gray-200">
        {(['details', 'curriculum', 'pricing'] as const).map(tab => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium border-b-2 transition-colors capitalize ${
              activeTab === tab
                ? 'border-indigo-600 text-indigo-600'
                : 'border-transparent text-gray-500 hover:text-gray-700'
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      {/* Details tab */}
      {activeTab === 'details' && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 max-w-2xl">
          <DetailsForm course={course} onSave={(data) => updateMutation.mutate(data)} />
        </div>
      )}

      {/* Curriculum tab */}
      {activeTab === 'curriculum' && (
        <div className="space-y-4">
          {/* Create section */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 flex gap-3">
            <input
              type="text"
              value={newSectionTitle}
              onChange={e => setNewSectionTitle(e.target.value)}
              placeholder="New section title..."
              className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
              onKeyDown={e => { if (e.key === 'Enter' && newSectionTitle) createSectionMutation.mutate(newSectionTitle); }}
            />
            <button
              onClick={() => createSectionMutation.mutate(newSectionTitle)}
              disabled={!newSectionTitle}
              className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
            >
              <Plus className="w-4 h-4" /> Add Section
            </button>
          </div>

          {/* Sections list */}
          {(sections || []).map((section, idx) => (
            <div key={section.id} className="bg-white rounded-lg shadow-sm border border-gray-200">
              <div className="px-4 py-3 border-b border-gray-100 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <GripVertical className="w-4 h-4 text-gray-300" />
                  <span className="font-medium text-gray-800">{idx + 1}. {section.title}</span>
                </div>
                <button
                  onClick={() => deleteSectionMutation.mutate(section.id)}
                  className="p-1 text-red-500 hover:bg-red-50 rounded"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
              <div className="p-4">
                <LessonsList sectionId={section.id} />
                <div className="flex gap-2 mt-3">
                  <input
                    type="text"
                    value={newLessonTitle[section.id] || ''}
                    onChange={e => setNewLessonTitle({ ...newLessonTitle, [section.id]: e.target.value })}
                    placeholder="New lesson title..."
                    className="flex-1 px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
                    onKeyDown={e => {
                      if (e.key === 'Enter' && newLessonTitle[section.id]) {
                        createLessonMutation.mutate({ sectionId: section.id, title: newLessonTitle[section.id] });
                      }
                    }}
                  />
                  <button
                    onClick={() => newLessonTitle[section.id] && createLessonMutation.mutate({ sectionId: section.id, title: newLessonTitle[section.id] })}
                    disabled={!newLessonTitle[section.id]}
                    className="inline-flex items-center gap-1 px-3 py-1.5 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 text-sm font-medium"
                  >
                    <Plus className="w-3 h-3" /> Add Lesson
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Pricing tab */}
      {activeTab === 'pricing' && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 max-w-md">
          <PricingForm course={course} onSave={(data) => updateMutation.mutate(data)} />
        </div>
      )}
    </div>
  );
}

function DetailsForm({ course, onSave }: { course: any; onSave: (data: any) => void }) {
  const [title, setTitle] = useState(course.title);
  const [slug, setSlug] = useState(course.slug);
  const [description, setDescription] = useState(course.description || '');
  const [category, setCategory] = useState(course.category || '');

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
        <input type="text" value={title} onChange={e => setTitle(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Slug</label>
        <input type="text" value={slug} onChange={e => setSlug(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <textarea value={description} onChange={e => setDescription(e.target.value)} rows={4}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Category</label>
        <input type="text" value={category} onChange={e => setCategory(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
      </div>
      <button onClick={() => onSave({ title, slug, description, category })}
        className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium">
        <Save className="w-4 h-4" /> Save Changes
      </button>
    </div>
  );
}

function PricingForm({ course, onSave }: { course: any; onSave: (data: any) => void }) {
  const [priceCents, setPriceCents] = useState(course.price_cents || 0);
  const [currency, setCurrency] = useState(course.currency || 'usd');

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Price (in cents)</label>
        <input type="number" value={priceCents} onChange={e => setPriceCents(Number(e.target.value))}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
        <p className="text-xs text-gray-500 mt-1">${(priceCents / 100).toFixed(2)} {currency.toUpperCase()}</p>
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Currency</label>
        <select value={currency} onChange={e => setCurrency(e.target.value)}
          className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500">
          <option value="usd">USD</option>
          <option value="eur">EUR</option>
          <option value="gbp">GBP</option>
        </select>
      </div>
      <button onClick={() => onSave({ priceCents, currency })}
        className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium">
        <Save className="w-4 h-4" /> Save Pricing
      </button>
    </div>
  );
}

function LessonsList({ sectionId }: { sectionId: string }) {
  const queryClient = useQueryClient();
  const { data: lessons } = useQuery({
    queryKey: ['section-lessons', sectionId],
    queryFn: () => lessonApi.listBySection(sectionId),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => lessonApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['section-lessons', sectionId] }),
  });

  if (!lessons || lessons.length === 0) {
    return <p className="text-sm text-gray-400 py-2">No lessons yet</p>;
  }

  return (
    <div className="space-y-1">
      {lessons.map((lesson, idx) => (
        <div key={lesson.id} className="flex items-center justify-between px-3 py-2 rounded-lg hover:bg-gray-50">
          <div className="flex items-center gap-2">
            <GripVertical className="w-3 h-3 text-gray-300" />
            {lesson.type === 'video' ? <Video className="w-4 h-4 text-gray-400" /> : <FileText className="w-4 h-4 text-gray-400" />}
            <span className="text-sm text-gray-700">{idx + 1}. {lesson.title}</span>
            {lesson.is_preview && <span className="px-1.5 py-0.5 bg-blue-50 text-blue-600 text-xs rounded">Preview</span>}
          </div>
          <button onClick={() => deleteMutation.mutate(lesson.id)}
            className="p-1 text-red-400 hover:bg-red-50 rounded">
            <Trash2 className="w-3 h-3" />
          </button>
        </div>
      ))}
    </div>
  );
}

// Need to import queryClient for the LessonsList deleteMutation
// (useQueryClient is already imported at the top)

