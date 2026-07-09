import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { ArrowLeft, Tag, Check } from 'lucide-react';
import { courseApi, enrollmentApi, couponApi } from '../../api/courseApi';
import api from '../../api/clientBase';

export default function CheckoutPage() {
  const { courseSlug } = useParams<{ courseSlug: string }>();
  const navigate = useNavigate();
  const [couponCode, setCouponCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState<any>(null);
  const [couponError, setCouponError] = useState('');

  const { data: course, isLoading } = useQuery({
    queryKey: ['checkout-course', courseSlug],
    queryFn: () => courseApi.getStorefront(courseSlug!),
    enabled: !!courseSlug,
  });

  const validateCouponMutation = useMutation({
    mutationFn: (code: string) => couponApi.validate(code),
    onSuccess: (coupon) => {
      setAppliedCoupon(coupon);
      setCouponError('');
    },
    onError: () => {
      setAppliedCoupon(null);
      setCouponError('Invalid or expired coupon');
    },
  });

  const enrollMutation = useMutation({
    mutationFn: async () => {
      // For free courses, enroll directly
      if (finalPrice === 0) {
        return enrollmentApi.create(course!.id);
      }
      // For paid courses, create a Stripe Checkout Session
      // The backend will redirect to Stripe-hosted checkout
      const response = await api.post('/storefront/checkout', {
        courseId: course!.id,
        couponCode: appliedCoupon?.code,
      });
      if (response.data.checkoutUrl) {
        window.location.href = response.data.checkoutUrl;
        return;
      }
      // Fallback: direct enrollment (for testing without Stripe)
      return enrollmentApi.create(course!.id);
    },
    onSuccess: (data) => {
      if (data && data.id) {
        navigate(`/learn/course/${course!.id}`);
      }
    },
  });

  if (isLoading) return <div className="min-h-screen flex items-center justify-center text-gray-500">Loading...</div>;
  if (!course) return <div className="min-h-screen flex items-center justify-center text-gray-500">Course not found</div>;

  const originalPrice = course.price_cents;
  let discount = 0;
  if (appliedCoupon) {
    if (appliedCoupon.discount_type === 'percent') {
      discount = Math.round(originalPrice * appliedCoupon.discount_value / 100);
    } else {
      discount = appliedCoupon.discount_value;
    }
  }
  const finalPrice = Math.max(0, originalPrice - discount);

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-2xl mx-auto px-6">
        <Link to={`/courses/${course.slug}`} className="inline-flex items-center gap-2 text-gray-500 hover:text-gray-700 text-sm mb-6">
          <ArrowLeft className="w-4 h-4" /> Back to course
        </Link>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <div className="p-6 border-b border-gray-200">
            <h1 className="text-xl font-bold text-gray-800 mb-2">Checkout</h1>
            <p className="text-sm text-gray-500">{course.title}</p>
          </div>

          <div className="p-6 space-y-4">
            {/* Price summary */}
            <div className="flex items-center justify-between">
              <span className="text-gray-600">Original Price</span>
              <span className="text-gray-800 font-medium">${(originalPrice / 100).toFixed(2)}</span>
            </div>

            {appliedCoupon && (
              <div className="flex items-center justify-between text-green-600">
                <span className="flex items-center gap-1">
                  <Tag className="w-4 h-4" /> Discount ({appliedCoupon.code})
                </span>
                <span>-${(discount / 100).toFixed(2)}</span>
              </div>
            )}

            <div className="flex items-center justify-between pt-4 border-t border-gray-200">
              <span className="font-semibold text-gray-800">Total</span>
              <span className="text-2xl font-bold text-gray-800">${(finalPrice / 100).toFixed(2)}</span>
            </div>

            {/* Coupon input */}
            <div className="pt-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">Coupon Code</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={couponCode}
                  onChange={e => setCouponCode(e.target.value.toUpperCase())}
                  placeholder="Enter coupon code"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
                />
                <button
                  onClick={() => couponCode && validateCouponMutation.mutate(couponCode)}
                  disabled={!couponCode}
                  className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 disabled:opacity-50 text-sm font-medium"
                >
                  Apply
                </button>
              </div>
              {couponError && <p className="text-sm text-red-500 mt-1">{couponError}</p>}
              {appliedCoupon && (
                <p className="text-sm text-green-600 mt-1 flex items-center gap-1">
                  <Check className="w-4 h-4" /> Coupon applied: {appliedCoupon.code}
                </p>
              )}
            </div>
          </div>

          <div className="p-6 bg-gray-50 border-t border-gray-200">
            {finalPrice === 0 ? (
              <button
                onClick={() => enrollMutation.mutate()}
                disabled={enrollMutation.isPending}
                className="w-full px-4 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 font-medium disabled:opacity-50"
              >
                {enrollMutation.isPending ? 'Enrolling...' : 'Enroll for Free'}
              </button>
            ) : (
              <button
                onClick={() => {
                  // TODO: Redirect to Stripe Checkout in Phase 5
                  alert('Stripe checkout integration coming in Phase 5. For now, enrolling directly.');
                  enrollMutation.mutate();
                }}
                disabled={enrollMutation.isPending}
                className="w-full px-4 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium disabled:opacity-50"
              >
                {enrollMutation.isPending ? 'Processing...' : `Pay $${(finalPrice / 100).toFixed(2)} & Enroll`}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
