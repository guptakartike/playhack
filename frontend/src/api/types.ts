export interface User {
  id: string;
  college_email: string;
  name?: string;
  role: string;
  created_at: string;
}

export interface Facility {
  id: string;
  name: string;
  sport_type: string;
}

export interface Court {
  id: string;
  label: string;
  max_players: number;
}

export interface SlotWithAvailability {
  id: string;
  start_time: string;
  end_time: string;
  available: boolean;
}

export interface Booking {
  id: string;
  slot_id: string;
  user_id: string;
  player_count: number;
  status: 'confirmed' | 'cancelled' | string;
  created_at: string;
}

export interface BookingDetail {
  id: string;
  slot_id: string;
  user_id: string;
  facility_name: string;
  sport_type: string;
  court_label: string;
  start_time: string;
  end_time: string;
  player_count: number;
  status: 'confirmed' | 'cancelled' | string;
  created_at: string;
}

export interface WaitlistEntry {
  id: string;
  slot_id: string;
  user_id: string;
  status: 'waiting' | 'notified' | string;
  created_at: string;
  notified_at?: string;
}

export interface NotificationPayload {
  slot_id: string;
  court_label: string;
  facility_name: string;
  start_time: string;
  message: string;
}

export interface RequestOTPResponse {
  message: string;
  code?: string;
}

export interface VerifyOTPResponse {
  token: string;
  user: User;
}

export interface ApiError {
  error: string;
}
