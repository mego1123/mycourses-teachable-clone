import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { Award, CheckCircle, XCircle } from 'lucide-react';
import { certificateApi } from '../../api/courseApi';

export default function CertificateVerifyPage() {
  const { token } = useParams<{ token: string }>();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['verify-certificate', token],
    queryFn: () => certificateApi.verify(token!),
    enabled: !!token,
  });

  if (isLoading) {
    return <div className="min-h-screen flex items-center justify-center text-gray-500">Verifying certificate...</div>;
  }

  if (isError || !data?.valid) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-lg border border-gray-200 p-12 text-center max-w-md">
          <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-800 mb-2">Certificate Not Found</h1>
          <p className="text-gray-500">This certificate is either invalid, revoked, or does not exist.</p>
        </div>
      </div>
    );
  }

  const cert = data.certificate;

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12">
      <div className="bg-white rounded-lg shadow-lg border border-gray-200 p-12 max-w-2xl w-full mx-4">
        <div className="text-center mb-8">
          <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
          <h1 className="text-3xl font-bold text-gray-800 mb-2">Certificate Verified</h1>
          <p className="text-gray-500">This certificate is valid and authentic</p>
        </div>

        <div className="border-4 border-double border-indigo-200 rounded-lg p-8">
          <div className="text-center">
            <Award className="w-16 h-16 text-yellow-500 mx-auto mb-4" />
            <p className="text-sm text-gray-500 uppercase tracking-wider mb-2">Certificate of Completion</p>
            <p className="text-xs text-gray-400 mb-6">{cert.certificate_number}</p>

            <p className="text-gray-600 mb-2">This certifies that</p>
            <h2 className="text-2xl font-bold text-gray-800 mb-4">{cert.learner_name}</h2>
            <p className="text-gray-600 mb-2">has successfully completed</p>
            <h3 className="text-xl font-semibold text-indigo-600 mb-4">{cert.course_title}</h3>
            <p className="text-gray-600 mb-6">offered by {cert.creator_name}</p>

            <div className="flex items-center justify-between pt-6 border-t border-gray-200">
              <div className="text-left">
                <p className="text-xs text-gray-400">Issue Date</p>
                <p className="text-sm font-medium text-gray-700">
                  {new Date(cert.issued_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
                </p>
              </div>
              <div className="text-right">
                <p className="text-xs text-gray-400">Verification</p>
                <p className="text-sm font-mono text-gray-500">{cert.verification_token.slice(0, 16)}...</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
