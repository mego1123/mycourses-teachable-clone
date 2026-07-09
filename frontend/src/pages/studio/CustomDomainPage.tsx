import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Globe, CheckCircle, XCircle, Clock } from 'lucide-react';
import { customDomainApi } from '../../api/courseApi';

export default function CustomDomainPage() {
  const [newDomain, setNewDomain] = useState('');
  const queryClient = useQueryClient();

  const { data: domains, isLoading } = useQuery({
    queryKey: ['custom-domains'],
    queryFn: () => customDomainApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (domain: string) => customDomainApi.create(domain),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['custom-domains'] });
      setNewDomain('');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => customDomainApi.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['custom-domains'] }),
  });

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-800 mb-6">Custom Domain</h1>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <h3 className="font-semibold text-gray-800 mb-2">Add Custom Domain</h3>
        <p className="text-sm text-gray-500 mb-4">
          Point your custom domain (e.g. academy.yourname.com) to your storefront.
          We'll automatically provision an SSL certificate.
        </p>
        <div className="flex gap-3">
          <input
            type="text"
            value={newDomain}
            onChange={e => setNewDomain(e.target.value)}
            placeholder="academy.yourname.com"
            className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500"
          />
          <button
            onClick={() => createMutation.mutate(newDomain)}
            disabled={!newDomain}
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 disabled:opacity-50 text-sm font-medium"
          >
            <Plus className="w-4 h-4" /> Add Domain
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">Loading...</div>
      ) : !domains || domains.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <Globe className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">No custom domains configured</p>
        </div>
      ) : (
        <div className="space-y-4">
          {domains.map(domain => (
            <div key={domain.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Globe className="w-5 h-5 text-gray-400" />
                  <div>
                    <p className="font-medium text-gray-800">{domain.domain}</p>
                    <div className="flex items-center gap-3 mt-1 text-sm">
                      <StatusBadge label="DNS" verified={domain.dns_verified} status={domain.status} />
                      <StatusBadge label="SSL" verified={domain.ssl_status === 'active'} status={domain.ssl_status} />
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => deleteMutation.mutate(domain.id)}
                  className="p-2 text-red-500 hover:bg-red-50 rounded"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>

              {domain.status === 'pending' && (
                <div className="mt-4 p-4 bg-yellow-50 rounded-lg border border-yellow-200">
                  <p className="text-sm text-yellow-800 font-medium mb-2">DNS Configuration Required</p>
                  <p className="text-sm text-yellow-700 mb-2">
                    Add a CNAME record pointing to <code className="bg-yellow-100 px-1 rounded">mycourses.com</code>:
                  </p>
                  <div className="bg-yellow-100 rounded p-3 font-mono text-sm text-yellow-900">
                    <div>Name: {domain.domain.split('.')[0]}</div>
                    <div>Type: CNAME</div>
                    <div>Value: mycourses.com</div>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function StatusBadge({ label, verified, status }: { label: string; verified: boolean; status: string }) {
  const icon = verified ? <CheckCircle className="w-3 h-3" /> : status === 'pending' ? <Clock className="w-3 h-3" /> : <XCircle className="w-3 h-3" />;
  const color = verified ? 'bg-green-100 text-green-700' : status === 'pending' ? 'bg-yellow-100 text-yellow-700' : 'bg-red-100 text-red-700';
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium ${color}`}>
      {icon} {label}
    </span>
  );
}
