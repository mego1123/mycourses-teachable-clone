// API client namespaces for the course platform.
// Uses the existing axios instance from client.ts (with auth + tenant headers).

import api from './clientBase';
import type {
  Course, Section, Lesson, Enrollment, CourseProgress,
  CourseCoupon, Review, Payout, CustomDomain, Certificate,
} from '../types/course';

// =============================================================================
// COURSE API
// =============================================================================

export const courseApi = {
  // Storefront (public)
  listStorefront: (params?: { page?: number; limit?: number }) =>
    api.get<{ courses: Course[]; page: number; limit: number }>('/storefront/courses', { params }).then(r => r.data),

  getStorefront: (slug: string) =>
    api.get<Course>(`/storefront/courses/${slug}`).then(r => r.data),

  listMarketplace: (params?: { page?: number; limit?: number; category?: string }) =>
    api.get<{ courses: Course[]; page: number; limit: number }>('/storefront/marketplace', { params }).then(r => r.data),

  // Creator studio
  listByCreator: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get<{ courses: Course[]; total: number; page: number; limit: number }>('/creator/courses', { params }).then(r => r.data),

  get: (id: string) =>
    api.get<Course>(`/creator/courses/${id}`).then(r => r.data),

  create: (data: { title: string; slug: string; description?: string; priceCents?: number; currency?: string; category?: string }) =>
    api.post<Course>('/creator/courses', data).then(r => r.data),

  update: (id: string, data: Partial<{ title: string; slug: string; description: string; priceCents: number; currency: string; category: string }>) =>
    api.put<Course>(`/creator/courses/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    api.delete(`/creator/courses/${id}`).then(r => r.data),

  publish: (id: string) =>
    api.post<Course>(`/creator/courses/${id}/publish`).then(r => r.data),

  unpublish: (id: string) =>
    api.post<Course>(`/creator/courses/${id}/unpublish`).then(r => r.data),
};

// =============================================================================
// SECTION API
// =============================================================================

export const sectionApi = {
  listByCourse: (courseId: string) =>
    api.get<Section[]>(`/creator/courses/${courseId}/sections`).then(r => r.data),

  create: (courseId: string, data: { title: string; description?: string; sortOrder?: number; dripOffsetDays?: number }) =>
    api.post<Section>(`/creator/courses/${courseId}/sections`, data).then(r => r.data),

  update: (id: string, data: { title: string; description?: string; sortOrder?: number; dripOffsetDays?: number }) =>
    api.put<Section>(`/creator/sections/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    api.delete(`/creator/sections/${id}`).then(r => r.data),
};

// =============================================================================
// LESSON API
// =============================================================================

export const lessonApi = {
  listBySection: (sectionId: string) =>
    api.get<Lesson[]>(`/creator/sections/${sectionId}/lessons`).then(r => r.data),

  listByCourse: (courseId: string) =>
    api.get<Lesson[]>(`/creator/courses/${courseId}/lessons`).then(r => r.data),

  create: (sectionId: string, data: { title: string; type?: string; content?: string; sortOrder?: number; isPreview?: boolean; durationSec?: number }) =>
    api.post<Lesson>(`/creator/sections/${sectionId}/lessons`, data).then(r => r.data),

  update: (id: string, data: { title: string; type?: string; content?: string; sortOrder?: number; isPreview?: boolean; durationSec?: number }) =>
    api.put<Lesson>(`/creator/lessons/${id}`, data).then(r => r.data),

  delete: (id: string) =>
    api.delete(`/creator/lessons/${id}`).then(r => r.data),
};

// =============================================================================
// ENROLLMENT API
// =============================================================================

export const enrollmentApi = {
  listMine: (params?: { page?: number; limit?: number }) =>
    api.get<{ enrollments: Enrollment[]; page: number; limit: number }>('/learner/enrollments', { params }).then(r => r.data),

  listByCreator: (params?: { page?: number; limit?: number; status?: string }) =>
    api.get<{ enrollments: Enrollment[]; page: number; limit: number }>('/creator/enrollments', { params }).then(r => r.data),

  create: (courseId: string) =>
    api.post<Enrollment>(`/learner/enrollments/${courseId}`).then(r => r.data),
};

// =============================================================================
// PROGRESS API
// =============================================================================

export const progressApi = {
  update: (lessonId: string, data: { videoPositionSec: number; completed: boolean }) =>
    api.post<CourseProgress>(`/learner/progress/${lessonId}`, data).then(r => r.data),

  getByCourse: (courseId: string) =>
    api.get<{ progress: CourseProgress[]; completionPercentage: number }>(`/learner/progress/course/${courseId}`).then(r => r.data),
};

// =============================================================================
// COUPON API
// =============================================================================

export const couponApi = {
  list: (params?: { page?: number; limit?: number }) =>
    api.get<{ coupons: CourseCoupon[]; page: number; limit: number }>('/creator/coupons', { params }).then(r => r.data),

  create: (data: { code: string; discountType: 'percent' | 'fixed'; discountValue: number; currency?: string; courseId?: string; expiresAt?: string; usageLimit?: number }) =>
    api.post<CourseCoupon>('/creator/coupons', data).then(r => r.data),

  delete: (id: string) =>
    api.delete(`/creator/coupons/${id}`).then(r => r.data),

  validate: (code: string) =>
    api.post<CourseCoupon>('/storefront/coupons/validate', { code }).then(r => r.data),
};

// =============================================================================
// REVIEW API
// =============================================================================

export const reviewApi = {
  listPublic: (courseId: string, params?: { page?: number; limit?: number }) =>
    api.get<{ reviews: Review[]; averageRating: number; reviewCount: number }>(`/storefront/courses/${courseId}/reviews`, { params }).then(r => r.data),

  hide: (id: string) =>
    api.post(`/creator/reviews/${id}/hide`).then(r => r.data),
};

// =============================================================================
// PAYOUT API
// =============================================================================

export const payoutApi = {
  list: (params?: { page?: number; limit?: number }) =>
    api.get<Payout[]>('/creator/payouts', { params }).then(r => r.data),
};

// =============================================================================
// CUSTOM DOMAIN API
// =============================================================================

export const customDomainApi = {
  list: () =>
    api.get<CustomDomain[]>('/creator/custom-domains').then(r => r.data),

  create: (domain: string) =>
    api.post<CustomDomain>('/creator/custom-domains', { domain }).then(r => r.data),

  delete: (id: string) =>
    api.delete(`/creator/custom-domains/${id}`).then(r => r.data),
};

// =============================================================================
// CERTIFICATE API
// =============================================================================

export const certificateApi = {
  listMine: () =>
    api.get<Certificate[]>('/learner/certificates').then(r => r.data),

  verify: (token: string) =>
    api.get<{ valid: boolean; certificate: Certificate }>(`/public/certificates/verify/${token}`).then(r => r.data),
};
