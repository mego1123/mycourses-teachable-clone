import { useQuery, useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { CheckCircle, Clock, AlertCircle, ExternalLink } from 'lucide-react';
import api from '../../api/clientBase';

export default function ConnectOnboardingPage() {
  const navigate = useNavigate();

  const { data: status, isLoading } = useQuery({
    queryKey: ['connect-status'],
    queryFn: () => api.get('/creator/connect/status').then(r => r.data),
    refetchInterval: 5000, // Poll every 5 seconds during onboarding
  });

  const onboardMutation = useMutation({
    mutationFn: () => api.get('/creator/connect/onboarding').then(r => r.data),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
  });

  if (isLoading) {
    return <div className="text-center py-12 text-gray-500">Loading...</div>;
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Stripe Connect</h1>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-8">
        {status?.connected ? (
          <div>
            <div className="flex items-center gap-3 mb-6">
              {status.status === 'active' ? (
                <CheckCircle className="w-8 h-8 text-green-500" />
              ) : status.status === 'in_review' ? (
                <Clock className="w-8 h-8 text-yellow-500" />
              ) : (
                <AlertCircle className="w-8 h-8 text-blue-500" />
              )}
              <div>
                <h2 className="font-semibold text-gray-800">
                  {status.status === 'active' ? 'Connected & Ready' :
                   status.status === 'in_review' ? 'Under Review' :
                   'Onboarding Incomplete'}
                </h2>
                <p className="text-sm text-gray-500">Account: {status.accountId}</p>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4 mb-6">
              <StatusItem label="Charges Enabled" enabled={status.chargesEnabled} />
              <StatusItem label="Payouts Enabled" enabled={status.payoutsEnabled} />
              <StatusItem label="Details Submitted" enabled={status.detailsSubmitted} />
              <StatusItem label="Can Receive Payments" enabled={status.status === 'active'} />
            </div>

            {status.status !== 'active' && (
              <button
                onClick={() => onboardMutation.mutate()}
                disabled={onboardMutation.isPending}
                className="w-full px-4 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium disabled:opacity-50"
              >
                {onboardMutation.isPending ? 'Redirecting...' : 'Complete Onboarding'}
              </button>
            )}

            {status.status === 'active' && (
              <div className="space-y-3">
                <button
                  onClick={() => navigate('/studio/dashboard')}
                  className="w-full px-4 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 font-medium"
                >
                  You're all set! Go to Dashboard
                </button>
                <button
                  onClick={() => onboardMutation.mutate()}
                  className="w-full px-4 py-2 text-gray-600 hover:bg-gray-50 rounded-lg text-sm font-medium"
                >
                  Update Stripe Account Details
                </button>
              </div>
            )}
          </div>
        ) : (
          <div className="text-center py-8">
            <div className="w-16 h-16 bg-indigo-50 rounded-full flex items-center justify-center mx-auto mb-4">
              <ExternalLink className="w-8 h-8 text-indigo-500" />
            </div>
            <h2 className="font-semibold text-gray-800 mb-2">Connect Your Stripe Account</h2>
            <p className="text-sm text-gray-500 mb-6 max-w-md mx-auto">
              To receive payouts from course sales, you need to connect a Stripe account.
              This is where your earnings will be deposited.
            </p>
            <button
              onClick={() => onboardMutation.mutate()}
              disabled={onboardMutation.isPending}
              className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium disabled:opacity-50"
            >
              {onboardMutation.isPending ? 'Redirecting...' : 'Connect with Stripe'}
            </button>
          </div>
        )}
      </div>

      <div className="mt-6 bg-blue-50 rounded-lg p-4 text-sm text-blue-700">
        <p className="font-medium mb-1">How payouts work:</p>
        <ul className="list-disc list-inside space-y-1 text-blue-600">
          <li>When a learner buys your course, the payment is split automatically</li>
          <li>Platform commission is deducted (default 10%)</li>
          <li>Your earnings accumulate in your Stripe balance</li>
          <li>Payouts to your bank account happen weekly (configurable)</li>
        </ul>
      </div>
    </div>
  );
}

function StatusItem({ label, enabled }: { label: string; enabled: boolean }) {
  return (
    <div className={`flex items-center gap-2 p-3 rounded-lg ${enabled ? 'bg-green-50' : 'bg-gray-50'}`}>
      {enabled ? (
        <CheckCircle className="w-4 h-4 text-green-500" />
      ) : (
        <Clock className="w-4 h-4 text-gray-400" />
      )}
      <span className={`text-sm ${enabled ? 'text-green-700' : 'text-gray-500'}`}>{label}</span>
    </div>
  );
}
