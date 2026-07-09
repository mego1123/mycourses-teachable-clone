import { useState, useEffect, useRef } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, CheckCircle, Lock, Play, FileText, ChevronRight } from 'lucide-react';
import { courseApi, sectionApi, lessonApi, progressApi } from '../../api/courseApi';
import { useCourseProgress } from '../../hooks/useCourseProgress';

export default function CoursePlayerPage() {
  const { courseId, lessonId } = useParams<{ courseId: string; lessonId?: string }>();
  const navigate = useNavigate();
  const videoRef = useRef<HTMLVideoElement>(null);
  const [currentLessonId, setCurrentLessonId] = useState(lessonId || '');

  const { reportProgress, startHeartbeat, stopHeartbeat } = useCourseProgress(currentLessonId, courseId);

  const { data: course } = useQuery({
    queryKey: ['player-course', courseId],
    queryFn: () => courseApi.get(courseId!),
    enabled: !!courseId,
  });

  const { data: sections } = useQuery({
    queryKey: ['player-sections', courseId],
    queryFn: () => sectionApi.listByCourse(courseId!),
    enabled: !!courseId,
  });

  const { data: progress } = useQuery({
    queryKey: ['player-progress', courseId],
    queryFn: () => progressApi.getByCourse(courseId!),
    enabled: !!courseId,
  });

  // Find all lessons across sections
  
  // Find current lesson
  
  // Get lessons for each section
  const [sectionLessons, setSectionLessons] = useState<Record<string, any[]>>({});
  useEffect(() => {
    if (sections) {
      sections.forEach(async s => {
        const lessons = await lessonApi.listBySection(s.id);
        setSectionLessons(prev => ({ ...prev, [s.id]: lessons }));
      });
    }
  }, [sections]);

  // Set first lesson if none selected
  useEffect(() => {
    if (!currentLessonId && sectionLessons) {
      const firstSection = Object.keys(sectionLessons)[0];
      if (firstSection && sectionLessons[firstSection]?.length > 0) {
        setCurrentLessonId(sectionLessons[firstSection][0].id);
      }
    }
  }, [sectionLessons, currentLessonId]);

  const currentLesson = Object.values(sectionLessons).flat().find(l => l.id === currentLessonId);


  const completionPct = progress?.completionPercentage || 0;

  return (
    <div className="min-h-screen bg-gray-900 flex">
      {/* Sidebar — course curriculum */}
      <aside className="w-80 bg-gray-800 text-gray-300 flex flex-col h-screen sticky top-0 overflow-y-auto">
        <div className="p-4 border-b border-gray-700">
          <Link to="/learn/my-courses" className="flex items-center gap-2 text-sm text-gray-400 hover:text-white mb-3">
            <ArrowLeft className="w-4 h-4" /> My Courses
          </Link>
          <h2 className="font-semibold text-white text-sm">{course?.title || 'Loading...'}</h2>
          <div className="mt-2">
            <div className="flex items-center justify-between text-xs text-gray-400 mb-1">
              <span>{completionPct}% complete</span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-1.5">
              <div className="bg-indigo-500 h-1.5 rounded-full transition-all" style={{ width: `${completionPct}%` }} />
            </div>
          </div>
        </div>

        <div className="flex-1 p-2">
          {sections?.map((section, idx) => (
            <div key={section.id} className="mb-4">
              <p className="px-3 py-2 text-xs font-semibold text-gray-400 uppercase tracking-wider">
                {idx + 1}. {section.title}
              </p>
              {(sectionLessons[section.id] || []).map((lesson, lIdx) => {
                const lessonProgress = progress?.progress?.find((p: any) => p.lesson_id === lesson.id);
                const isCompleted = lessonProgress?.completed;
                const isCurrent = lesson.id === currentLessonId;
                return (
                  <button
                    key={lesson.id}
                    onClick={() => {
                      setCurrentLessonId(lesson.id);
                      navigate(`/learn/course/${courseId}/${lesson.id}`, { replace: true });
                    }}
                    className={`w-full flex items-center gap-2 px-3 py-2 rounded text-sm text-left transition-colors ${
                      isCurrent ? 'bg-indigo-600 text-white' : 'text-gray-300 hover:bg-gray-700'
                    }`}
                  >
                    {isCompleted ? (
                      <CheckCircle className="w-4 h-4 text-green-400 flex-shrink-0" />
                    ) : lesson.is_preview ? (
                      <Play className="w-4 h-4 text-green-400 flex-shrink-0" />
                    ) : (
                      <Lock className="w-4 h-4 text-gray-500 flex-shrink-0" />
                    )}
                    <span className="truncate">{lIdx + 1}. {lesson.title}</span>
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      </aside>

      {/* Main — lesson content */}
      <div className="flex-1 flex flex-col h-screen overflow-y-auto">
        {currentLesson ? (
          <>
            {/* Video player area */}
            <div className="bg-black aspect-video flex items-center justify-center">
              {currentLesson.type === 'video' ? (
                <video
                  ref={videoRef}
                  controls
                  className="w-full h-full"
                  onPlay={() => startHeartbeat(
                    () => videoRef.current?.currentTime || 0,
                    () => videoRef.current?.duration || 0
                  )}
                  onPause={stopHeartbeat}
                  onEnded={() => {
                    stopHeartbeat();
                    reportProgress(
                      videoRef.current?.duration || 0,
                      videoRef.current?.duration || 0,
                      true
                    );
                  }}
                >
                  {/* Video source will be set from Cloudflare Stream signed URL */}
                  {/* For now, a placeholder source */}
                  <source src="" type="video/mp4" />
                  Your browser does not support the video tag.
                </video>
              ) : (
                <div className="text-center text-gray-400 p-8">
                  <FileText className="w-16 h-16 mx-auto mb-3 opacity-50" />
                  <p className="text-sm">Text lesson</p>
                  <p className="text-xs text-gray-500 mt-1">{currentLesson.title}</p>
                </div>
              )}
            </div>

            {/* Lesson info + actions */}
            <div className="p-8">
              <h1 className="text-2xl font-bold text-white mb-4">{currentLesson.title}</h1>
              {currentLesson.content && (
                <div className="prose prose-invert max-w-none text-gray-300 mb-6">
                  <p>{currentLesson.content}</p>
                </div>
              )}

              <div className="flex gap-3">
                <button
                  onClick={() => reportProgress(0, 0, true)}
                  className="inline-flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm font-medium"
                >
                  <CheckCircle className="w-4 h-4" /> Mark as Complete
                </button>

                {/* Next lesson button */}
                {(() => {
                  const allFlat = Object.values(sectionLessons).flat();
                  const currentIdx = allFlat.findIndex(l => l.id === currentLessonId);
                  const nextLesson = allFlat[currentIdx + 1];
                  if (nextLesson) {
                    return (
                      <button
                        onClick={() => {
                          setCurrentLessonId(nextLesson.id);
                          navigate(`/learn/course/${courseId}/${nextLesson.id}`, { replace: true });
                        }}
                        className="inline-flex items-center gap-2 px-4 py-2 bg-gray-700 text-white rounded-lg hover:bg-gray-600 text-sm font-medium"
                      >
                        Next Lesson <ChevronRight className="w-4 h-4" />
                      </button>
                    );
                  }
                  return null;
                })()}
              </div>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-500">
            Select a lesson to start learning
          </div>
        )}
      </div>
    </div>
  );
}
