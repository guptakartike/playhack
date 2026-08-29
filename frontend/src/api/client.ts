import type {
  User,
  Facility,
  Court,
  SlotWithAvailability,
  Booking,
  BookingDetail,
  WaitlistEntry,
  RequestOTPResponse,
  VerifyOTPResponse,
  ApiError,
} from './types';

const API_BASE = '/api';

export class ApiException extends Error {
  status: number;
  data?: ApiError;

  constructor(status: number, message: string, data?: ApiError) {
    super(message);
    this.status = status;
    this.data = data;
    this.name = 'ApiException';
  }
}

export function getStoredToken(): string | null {
  return localStorage.getItem('token');
}

export function setStoredToken(token: string | null): void {
  if (token) {
    localStorage.setItem('token', token);
  } else {
    localStorage.removeItem('token');
  }
}

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {});
  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const token = getStoredToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  if (response.status === 204) {
    return {} as T;
  }

  const contentType = response.headers.get('Content-Type') || '';
  const isJson = contentType.includes('application/json');
  const data = isJson ? await response.json() : await response.text();

  if (!response.ok) {
    const errorMessage =
      (isJson && (data as ApiError)?.error) ||
      `Request failed with status ${response.status}`;

    if (response.status === 401) {
      // Clear token if expired or invalid
      setStoredToken(null);
    }

    throw new ApiException(response.status, errorMessage, isJson ? (data as ApiError) : undefined);
  }

  return data as T;
}

// Auth API
export const authApi = {
  requestOTP: (email: string) =>
    apiFetch<RequestOTPResponse>('/auth/request-otp', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  verifyOTP: (email: string, code: string) =>
    apiFetch<VerifyOTPResponse>('/auth/verify-otp', {
      method: 'POST',
      body: JSON.stringify({ email, code }),
    }),

  getMe: () => apiFetch<User>('/me'),
};

// Facility & Slot Discovery API
export const facilityApi = {
  listFacilities: () => apiFetch<Facility[]>('/facilities'),

  listCourts: (facilityId: string) =>
    apiFetch<Court[]>(`/facilities/${facilityId}/courts`),

  listSlots: (courtId: string, dateStr?: string) => {
    const query = dateStr ? `?date=${encodeURIComponent(dateStr)}` : '';
    return apiFetch<SlotWithAvailability[]>(`/courts/${courtId}/slots${query}`);
  },
};

// Bookings API
export const bookingApi = {
  createBooking: (slotId: string, playerCount: number = 1) =>
    apiFetch<Booking>('/bookings', {
      method: 'POST',
      body: JSON.stringify({ slot_id: slotId, player_count: playerCount }),
    }),

  getMyBookings: () => apiFetch<BookingDetail[]>('/bookings/mine'),

  cancelBooking: (bookingId: string) =>
    apiFetch<{ message: string }>(`/bookings/${bookingId}`, {
      method: 'DELETE',
    }),
};

// Waitlist API
export const waitlistApi = {
  joinWaitlist: (slotId: string) =>
    apiFetch<WaitlistEntry>(`/slots/${slotId}/waitlist`, {
      method: 'POST',
    }),
};
