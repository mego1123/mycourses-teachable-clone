import { useQuery } from '@tanstack/react-query';
import { DollarSign, TrendingUp, Users } from 'lucide-react';
import { enrollmentApi } from '../../api/courseApi';

export default function SalesPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['creator-enrollments-all'],
    queryFn: () => enrollmentApi.listByCreator({ limit: 100 }),
  });

  const enrollments = data?.enrollments || [];
  const totalRevenue = enrollments
    .filter(e => e.status === 'active' || e.status === 'completed')
    .reduce((sum, e) => sum + e.price_paid_cents, 0);
  const totalRefunded = enrollments
    .filter(e => e.status === 'refunded')
    .reduce((sum, e) => sum + e.price_paid_cents, 0);
  const netRevenue = totalRevenue - totalRefunded;

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Sales</h1>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center">
              <DollarSign className="w-5 h-5 text-white" />
            </div>
          </div>
          <p className="text-2xl font-bold text-gray-800">${(totalRevenue / 100).toFixed(2)}</p>
          <p className="text-sm text-gray-500 mt-1">Gross Revenue</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-red-500 rounded-lg flex items-center justify-center">
              <TrendingUp className="w-5 h-5 text-white" />
            </div>
          </div>
          <p className="text-2xl font-bold text-gray-800">${(totalRefunded / 100).toFixed(2)}</p>
          <p className="text-sm text-gray-500 mt-1">Refunded</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-indigo-500 rounded-lg flex items-center justify-center">
              <Users className="w-5 h-5 text-white" />
            </div>
          </div>
          <p className="text-2xl font-bold text-gray-800">${(netRevenue / 100).toFixed(2)}</p>
          <p className="text-sm text-gray-500 mt-1">Net Revenue</p>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="font-semibold text-gray-800">Transactions</h3>
        </div>
        {isLoading ? (
          <div className="text-center py-12 text-gray-500">Loading...</div>
        ) : enrollments.length === 0 ? (
          <div className="text-center py-12 text-gray-500">No sales yet</div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Course ID</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {enrollments.map(e => (
                <tr key={e.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm text-gray-600">
                    {new Date(e.enrolled_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 text-gray-800 font-medium">
                    ${(e.price_paid_cents / 100).toFixed(2)}
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      e.status === 'active' ? 'bg-green-100 text-green-700' :
                      e.status === 'completed' ? 'bg-blue-100 text-blue-700' :
                      e.status === 'refunded' ? 'bg-red-100 text-red-700' :
                      'bg-gray-100 text-gray-600'
                    }`}>{e.status}</span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-400 font-mono">
                    {e.course_id.slice(0, 8)}...
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
