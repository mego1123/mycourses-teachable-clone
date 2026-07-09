import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { DollarSign, TrendingUp, Wallet, Clock } from 'lucide-react';
import api from '../../api/clientBase';
import type { Payout } from '../../types/course';

export default function PayoutsPage() {
  const queryClient = useQueryClient();

  const { data: payoutsData, isLoading } = useQuery({
    queryKey: ['creator-payouts'],
    queryFn: () => api.get<{ payouts: Payout[]; totalPaid: number }>('/creator/payouts').then(r => r.data),
  });

  const { data: connectStatus } = useQuery({
    queryKey: ['connect-status'],
    queryFn: () => api.get('/creator/connect/status').then(r => r.data),
  });

  const requestPayoutMutation = useMutation({
    mutationFn: () => api.post('/creator/payouts/request').then(r => r.data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['creator-payouts'] }),
  });

  const payouts = payoutsData?.payouts || [];
  const totalPaid = payoutsData?.totalPaid || 0;
  const isConnectActive = connectStatus?.status === 'active';

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Payouts</h1>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center">
              <DollarSign className="w-5 h-5 text-white" />
            </div>
          </div>
          <p className="text-2xl font-bold text-gray-800">${(totalPaid / 100).toFixed(2)}</p>
          <p className="text-sm text-gray-500 mt-1">Total Paid Out</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-indigo-500 rounded-lg flex items-center justify-center">
              <Wallet className="w-5 h-5 text-white" />
            </div>
          </div>
          <p className="text-2xl font-bold text-gray-800">
            ${payouts.filter(p => p.status === 'pending').reduce((sum, p) => sum + p.amount_cents, 0) / 100}
          </p>
          <p className="text-sm text-gray-500 mt-1">Pending</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="w-10 h-10 bg-yellow-500 rounded-lg flex items-center justify-center">
              <TrendingUp className="w-5 h-5 text-white" />
            </div>
          </div>
          <p className="text-2xl font-bold text-gray-800">
            ${payouts.filter(p => p.status === 'failed').reduce((sum, p) => sum + p.amount_cents, 0) / 100}
          </p>
          <p className="text-sm text-gray-500 mt-1">Failed</p>
        </div>
      </div>

      {/* Request payout */}
      {isConnectActive ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
          <h3 className="font-semibold text-gray-800 mb-2">Request Manual Payout</h3>
          <p className="text-sm text-gray-500 mb-4">
            Transfer your available balance to your bank account. This will initiate a Stripe transfer.
          </p>
          <button
            onClick={() => requestPayoutMutation.mutate()}
            disabled={requestPayoutMutation.isPending}
            className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
          >
            {requestPayoutMutation.isPending ? 'Processing...' : 'Request Payout'}
          </button>
        </div>
      ) : (
        <div className="bg-yellow-50 rounded-lg border border-yellow-200 p-6 mb-6">
          <p className="text-sm text-yellow-800">
            Complete Stripe Connect onboarding to enable payouts.{' '}
            <a href="/studio/connect" className="font-medium underline">Set up Stripe →</a>
          </p>
        </div>
      )}

      {/* Payout history */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="font-semibold text-gray-800">Payout History</h3>
        </div>
        {isLoading ? (
          <div className="text-center py-12 text-gray-500">Loading...</div>
        ) : payouts.length === 0 ? (
          <div className="text-center py-12">
            <Clock className="w-12 h-12 text-gray-300 mx-auto mb-3" />
            <p className="text-gray-500">No payouts yet</p>
          </div>
        ) : (
          <table className="w-full">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Date</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Transfer ID</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {payouts.map(payout => (
                <tr key={payout.id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm text-gray-600">
                    {new Date(payout.initiated_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 font-medium text-gray-800">
                    ${(payout.amount_cents / 100).toFixed(2)}
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2 py-1 rounded text-xs font-medium ${
                      payout.status === 'paid' ? 'bg-green-100 text-green-700' :
                      payout.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                      payout.status === 'failed' ? 'bg-red-100 text-red-700' :
                      'bg-gray-100 text-gray-600'
                    }`}>{payout.status}</span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-400 font-mono">
                    {payout.stripe_transfer_id ? payout.stripe_transfer_id.slice(0, 12) + '...' : '—'}
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
