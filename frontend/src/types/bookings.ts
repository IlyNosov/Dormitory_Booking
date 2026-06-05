export type Room = number;

export interface Booking {
  id: string;
  start: string;
  end: string;
  room: Room;
  title: string;
  telegramId?: string;
  userEmail?: string;
  isPrivate: boolean;
  description?: string;
  canManage?: boolean;
}

export interface User {
  email: string;
  isAdmin: boolean;
  isSuperAdmin: boolean;
}

export interface RoomInfo {
  number: number;
  name: string;
  isActive: boolean;
  weekdayOpen: number;
  weekdayClose: number;
  friSatOpen: number;
  friSatClose: number;
  sunOpen: number;
  sunClose: number;
}

export interface AdminEntry {
  email: string;
  addedBy: string;
  createdAt: string;
}

export type RoomFilter = "all" | Room;
export type VisFilter = "all" | "public" | "private";
export type ViewMode = "cards" | "table";

export type CreateBookingPayload = {
  start: string;
  end: string;
  room: Room;
  title: string;
  isPrivate: boolean;
  description?: string;
  telegram_id?: string;
};

export type UpdateBookingPayload = {
  start: string;
  end: string;
  room: Room;
  title: string;
  isPrivate: boolean;
  description?: string;
  telegram_id?: string;
};

// Backward-compat alias
export type Bookings = Booking;
