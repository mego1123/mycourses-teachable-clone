// Course platform types — used by creator studio and learner pages.

export interface Course {
  id: string;
  tenant_id: string;
  title: string;
  description: string;
  slug: string;
  price_cents: number;
  currency: string;
  status: 'draft' | 'published' | 'archived';
  thumbnail_url?: string | null;
  intro_video_url?: string | null;
  category?: string | null;
  drip_enabled: boolean;
  published_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Section {
  id: string;
  course_id: string;
  title: string;
  description: string;
  sort_order: number;
  drip_offset_days: number;
  created_at: string;
  updated_at: string;
}

export interface Lesson {
  id: string;
  section_id: string;
  course_id: string;
  title: string;
  type: 'video' | 'text' | 'pdf' | 'quiz';
  content: string;
  media_asset_id?: string | null;
  sort_order: number;
  is_preview: boolean;
  duration_sec: number;
  created_at: string;
  updated_at: string;
}

export interface MediaAsset {
  id: string;
  tenant_id: string;
  kind: 'video' | 'pdf' | 'image';
  title: string;
  cf_stream_id?: string | null;
  r2_key?: string | null;
  r2_url?: string | null;
  size_bytes: number;
  mime_type: string;
  duration_sec: number;
  status: 'processing' | 'ready' | 'failed' | 'deleted';
  created_at: string;
  updated_at: string;
}

export interface Enrollment {
  id: string;
  course_id: string;
  tenant_id: string;
  user_id: string;
  status: 'active' | 'completed' | 'refunded' | 'disputed' | 'chargeback_lost';
  price_paid_cents: number;
  currency: string;
  coupon_id?: string | null;
  stripe_session_id?: string | null;
  stripe_charge_id?: string | null;
  enrolled_at: string;
  completed_at?: string | null;
  refunded_at?: string | null;
  refund_reason?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CourseProgress {
  id: string;
  enrollment_id: string;
  lesson_id: string;
  user_id: string;
  completed: boolean;
  completed_at?: string | null;
  video_position_sec: number;
  last_viewed_at: string;
  created_at: string;
  updated_at: string;
}

export interface CourseCoupon {
  id: string;
  tenant_id: string;
  code: string;
  discount_type: 'percent' | 'fixed';
  discount_value: number;
  currency: string;
  course_id?: string | null;
  expires_at?: string | null;
  usage_limit?: number | null;
  used_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Review {
  id: string;
  course_id: string;
  user_id: string;
  tenant_id: string;
  rating: number;
  comment: string;
  is_hidden: boolean;
  created_at: string;
  updated_at: string;
}

export interface Payout {
  id: string;
  tenant_id: string;
  stripe_transfer_id?: string | null;
  stripe_payout_id?: string | null;
  amount_cents: number;
  currency: string;
  status: 'pending' | 'paid' | 'failed' | 'cancelled';
  failure_reason?: string | null;
  initiated_at: string;
  completed_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CustomDomain {
  id: string;
  domain: string;
  tenant_id: string;
  status: 'pending' | 'active' | 'failed';
  dns_verified: boolean;
  ssl_status: 'pending' | 'active' | 'failed';
  cf_hostname_id?: string | null;
  verification_records: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface Certificate {
  id: string;
  enrollment_id: string;
  user_id: string;
  course_id: string;
  tenant_id: string;
  certificate_number: string;
  verification_token: string;
  learner_name: string;
  course_title: string;
  creator_name: string;
  issued_at: string;
  status: 'active' | 'revoked';
  revoked_reason?: string | null;
  revoked_at?: string | null;
  created_at: string;
}

export interface CreatorProfile {
  id: string;
  tenant_id: string;
  bio: string;
  website_url?: string | null;
  social_links: Record<string, string>;
  avatar_url?: string | null;
  banner_url?: string | null;
  created_at: string;
  updated_at: string;
}

// API response wrappers
export interface PaginatedResponse<T> {
  courses?: T[];
  enrollments?: T[];
  coupons?: T[];
  reviews?: T[];
  payouts?: T[];
  total: number;
  page: number;
  limit: number;
}

export interface AverageRating {
  avg_rating: number;
  review_count: number;
}
