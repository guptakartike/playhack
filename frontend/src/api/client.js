const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export function getStoredToken() {
  return localStorage.getItem('token') || '';
}

export function setStoredToken(token) {
  localStorage.setItem('token', token);
}

export function clearStoredToken() {
  localStorage.removeItem('token');
}

export async function fetchMyBookings() {
  const token = getStoredToken();
  if (!token) {
    throw new Error('UNAUTHORIZED');
  }

  const response = await fetch(`${API_BASE_URL}/bookings/mine`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (response.status === 401) {
    clearStoredToken();
    throw new Error('UNAUTHORIZED');
  }

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to fetch bookings');
  }

  return response.json();
}

export async function cancelBookingApi(bookingId) {
  const token = getStoredToken();
  if (!token) {
    throw new Error('UNAUTHORIZED');
  }

  const response = await fetch(`${API_BASE_URL}/bookings/${bookingId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (response.status === 401) {
    clearStoredToken();
    throw new Error('UNAUTHORIZED');
  }

  if (response.status === 403) {
    throw new Error('Booking belongs to another user');
  }

  if (response.status === 404) {
    throw new Error('Booking not found');
  }

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.error || 'Failed to cancel booking');
  }

  return response.json();
}
