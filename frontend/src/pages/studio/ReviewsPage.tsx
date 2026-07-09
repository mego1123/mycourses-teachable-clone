import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Star, EyeOff, MessageCircle } from 'lucide-react';
import { reviewApi } from '../../api/courseApi';

export default function ReviewsPage() {
  const queryClient = useQueryClient();

  // Note: We'd need a creator-level reviews list endpoint.
  // For now, show a placeholder that the reviews will appear here.
  const { data: reviewsData, isLoading } = useQuery({
    queryKey: ['creator-reviews'],
    queryFn: async () => {
      // TODO: Add a /creator/reviews endpoint
      // For now return empty — the API will be added in Phase 5
      return { reviews: [] as any[] };
    },
  });

  const hideMutation = useMutation({
    mutationFn: (id: string) => reviewApi.hide(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['creator-reviews'] }),
  });

  const reviews = reviewsData?.reviews || [];

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Reviews</h1>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">Loading...</div>
      ) : reviews.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <MessageCircle className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">No reviews yet</p>
          <p className="text-sm text-gray-400 mt-1">Reviews will appear here when learners start reviewing your courses</p>
        </div>
      ) : (
        <div className="space-y-4">
          {reviews.map(review => (
            <div key={review.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex items-center gap-1">
                    {[1,2,3,4,5].map(i => (
                      <Star key={i} className={`w-4 h-4 ${i <= review.rating ? 'text-yellow-400 fill-yellow-400' : 'text-gray-200'}`} />
                    ))}
                  </div>
                  <span className="text-sm text-gray-500">{new Date(review.created_at).toLocaleDateString()}</span>
                </div>
                <button
                  onClick={() => hideMutation.mutate(review.id)}
                  className="p-1.5 text-gray-400 hover:bg-gray-100 rounded"
                  title="Hide review"
                >
                  <EyeOff className="w-4 h-4" />
                </button>
              </div>
              {review.comment && <p className="text-gray-700 mt-3">{review.comment}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
