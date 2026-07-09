import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Ticket } from 'lucide-react';
import { couponApi } from '../../api/courseApi';

export default function CouponsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const [newCoupon, setNewCoupon] = useState({
    code: '', discountType: 'percent' as 'percent' | 'fixed', discountValue: 10,
    currency: 'usd', usageLimit: 0,
  });
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['creator-coupons'],
    queryFn: () => couponApi.list({ limit: 100 }),
  });

  const createMutation = useMutation({
    mutationFn: (data: typeof newCoupon) => couponApi.create({
      ...data,
      usageLimit: data.usageLimit > 0 ? data.usageLimit : undefined,
    }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['creator-coupons'] });
      setShowCreate(false);
      setNewCoupon({ code: '', discountType: 'percent', discountValue: 10, currency: 'usd', usageLimit: 0 });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => couponApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['creator-coupons'] }),
  });

  const coupons = data?.coupons || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800">Coupons</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 text-sm font-medium"
        >
          <Plus className="w-4 h-4" /> New Coupon
        </button>
      </div>

      {showCreate && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
          <h3 className="font-semibold text-gray-800 mb-4">Create Coupon</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Code</label>
              <input type="text" value={newCoupon.code} onChange={e => setNewCoupon({ ...newCoupon, code: e.target.value.toUpperCase() })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" placeholder="SAVE20" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Discount Type</label>
              <select value={newCoupon.discountType} onChange={e => setNewCoupon({ ...newCoupon, discountType: e.target.value as any })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500">
                <option value="percent">Percentage</option>
                <option value="fixed">Fixed Amount</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {newCoupon.discountType === 'percent' ? 'Discount (%)' : 'Discount (cents)'}
              </label>
              <input type="number" value={newCoupon.discountValue} onChange={e => setNewCoupon({ ...newCoupon, discountValue: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Usage Limit (0 = unlimited)</label>
              <input type="number" value={newCoupon.usageLimit} onChange={e => setNewCoupon({ ...newCoupon, usageLimit: Number(e.target.value) })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500" />
            </div>
          </div>
          <div className="flex gap-3 mt-4">
            <button onClick={() => createMutation.mutate(newCoupon)} disabled={!newCoupon.code}
              className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium">
              Create
            </button>
            <button onClick={() => setShowCreate(false)}
              className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 text-sm font-medium">
              Cancel
            </button>
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">Loading...</div>
      ) : coupons.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <Ticket className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">No coupons yet</p>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Code</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Discount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Used</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {coupons.map(coupon => (
                <tr key={coupon.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 font-mono font-medium text-gray-800">{coupon.code}</td>
                  <td className="px-6 py-4 text-gray-600">
                    {coupon.discount_type === 'percent' ? `${coupon.discount_value}%` : `$${(coupon.discount_value / 100).toFixed(2)}`}
                  </td>
                  <td className="px-6 py-4 text-gray-600">
                    {coupon.used_count}{coupon.usage_limit ? ` / ${coupon.usage_limit}` : ''}
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${coupon.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'}`}>
                      {coupon.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button onClick={() => deleteMutation.mutate(coupon.id)}
                      className="p-1.5 text-red-500 hover:bg-red-50 rounded">
                      <Trash2 className="w-4 h-4" />
                    </button>
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
