import { useQuery } from '@tanstack/react-query';
import { useParams, Link } from 'react-router-dom';
import { useEffect } from 'react';
import { ArrowLeft, BookOpen, Star, Clock, CheckCircle, Play } from 'lucide-react';
import { courseApi, sectionApi, lessonApi, reviewApi } from '../../api/courseApi';

export default function CourseDetailPage() {
  const { slug } = useParams<{ slug: string }>();

  const { data: course, isLoading } = useQuery({
    queryKey: ['storefront-course', slug],
    queryFn: () => courseApi.getStorefront(slug!),
    enabled: !!slug,
  });

  const { data: sections } = useQuery({
    queryKey: ['course-sections-public', course?.id],
    queryFn: () => sectionApi.listByCourse(course!.id),
    enabled: !!course?.id,
  });

  const { data: reviewsData } = useQuery({
    queryKey: ['course-reviews', course?.id],
    queryFn: () => reviewApi.listPublic(course!.id, { limit: 10 }),
    enabled: !!course?.id,
  });

  // Inject JSON-LD structured data for SEO
  useEffect(() => {
    if (!course) return;
    const jsonLd = {
      '@context': 'https://schema.org',
      '@type': 'Course',
      name: course.title,
      description: course.description || '',
      provider: {
        '@type': 'Organization',
        name: 'MyCourses',
      },
      offers: {
        '@type': 'Offer',
        price: (course.price_cents / 100).toFixed(2),
        priceCurrency: course.currency,
      },
      aggregateRating: reviewsData && reviewsData.reviewCount > 0 ? {
        '@type': 'AggregateRating',
        ratingValue: reviewsData.averageRating,
        reviewCount: reviewsData.reviewCount,
      } : undefined,
    };
    const script = document.createElement('script');
    script.type = 'application/ld+json';
    script.text = JSON.stringify(jsonLd);
    document.head.appendChild(script);
    return () => { document.head.removeChild(script); };
  }, [course, reviewsData]);

  if (isLoading) return <div className="min-h-screen flex items-center justify-center text-gray-500">Loading...</div>;
  if (!course) return <div className="min-h-screen flex items-center justify-center text-gray-500">Course not found</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-gray-900 text-white">
        <div className="max-w-5xl mx-auto px-6 py-8">
          <Link to="/courses" className="inline-flex items-center gap-2 text-gray-400 hover:text-white text-sm mb-4">
            <ArrowLeft className="w-4 h-4" /> Back to courses
          </Link>
          <div className="flex flex-col md:flex-row gap-8">
            <div className="flex-1">
              <h1 className="text-3xl font-bold mb-3">{course.title}</h1>
              <p className="text-gray-300 mb-4">{course.description}</p>
              <div className="flex items-center gap-4 text-sm text-gray-400">
                {reviewsData && reviewsData.reviewCount > 0 && (
                  <span className="flex items-center gap-1">
                    <Star className="w-4 h-4 text-yellow-400 fill-yellow-400" />
                    {reviewsData.averageRating.toFixed(1)} ({reviewsData.reviewCount} reviews)
                  </span>
                )}
                {course.category && <span className="px-2 py-0.5 bg-gray-800 rounded text-xs">{course.category}</span>}
              </div>
            </div>
            <div className="md:w-80 bg-gray-800 rounded-lg p-6">
              <div className="text-3xl font-bold mb-4">
                {course.price_cents === 0 ? 'Free' : `$${(course.price_cents / 100).toFixed(2)}`}
              </div>
              <Link
                to={`/checkout/${course.slug}`}
                className="block w-full text-center px-4 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium mb-3"
              >
                Enroll Now
              </Link>
              <div className="text-xs text-gray-400 space-y-2">
                <div className="flex items-center gap-2"><Play className="w-3 h-3" /> Full lifetime access</div>
                <div className="flex items-center gap-2"><CheckCircle className="w-3 h-3" /> Certificate of completion</div>
                <div className="flex items-center gap-2"><Clock className="w-3 h-3" /> Learn at your own pace</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Curriculum */}
      <div className="max-w-5xl mx-auto px-6 py-12">
        <h2 className="text-2xl font-bold text-gray-800 mb-6">Course Curriculum</h2>
        {sections && sections.length > 0 ? (
          <div className="space-y-4">
            {sections.map((section, idx) => (
              <div key={section.id} className="bg-white rounded-lg shadow-sm border border-gray-200">
                <div className="px-6 py-4 border-b border-gray-100">
                  <h3 className="font-semibold text-gray-800">{idx + 1}. {section.title}</h3>
                </div>
                <LessonsList sectionId={section.id} />
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-500">Curriculum coming soon</p>
        )}
      </div>

      {/* Reviews */}
      {reviewsData && reviewsData.reviewCount > 0 && (
        <div className="max-w-5xl mx-auto px-6 pb-12">
          <h2 className="text-2xl font-bold text-gray-800 mb-6">Student Reviews</h2>
          <div className="space-y-4">
            {reviewsData.reviews.map(review => (
              <div key={review.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                <div className="flex items-center gap-3 mb-2">
                  <div className="flex items-center gap-1">
                    {[1,2,3,4,5].map(i => (
                      <Star key={i} className={`w-4 h-4 ${i <= review.rating ? 'text-yellow-400 fill-yellow-400' : 'text-gray-200'}`} />
                    ))}
                  </div>
                  <span className="text-sm text-gray-500">{new Date(review.created_at).toLocaleDateString()}</span>
                </div>
                {review.comment && <p className="text-gray-700">{review.comment}</p>}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function LessonsList({ sectionId }: { sectionId: string }) {
  const { data: lessons } = useQuery({
    queryKey: ['section-lessons-public', sectionId],
    queryFn: () => lessonApi.listBySection(sectionId),
  });

  if (!lessons || lessons.length === 0) return null;

  return (
    <div className="divide-y divide-gray-50">
      {lessons.map((lesson, idx) => (
        <div key={lesson.id} className="flex items-center gap-3 px-6 py-3">
          {lesson.is_preview ? (
            <Play className="w-4 h-4 text-green-500" />
          ) : (
            <BookOpen className="w-4 h-4 text-gray-300" />
          )}
          <span className="text-sm text-gray-700 flex-1">{idx + 1}. {lesson.title}</span>
          {lesson.duration_sec > 0 && (
            <span className="text-xs text-gray-400">{Math.floor(lesson.duration_sec / 60)} min</span>
          )}
          {lesson.is_preview && (
            <span className="px-1.5 py-0.5 bg-green-50 text-green-600 text-xs rounded font-medium">Free Preview</span>
          )}
        </div>
      ))}
    </div>
  );
}
