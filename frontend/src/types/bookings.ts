export type Room = 21 | 132 | 256;

export type Bookings = {
    id: string;
    start: string;
    end: string;
    room: Room;
    title: string;
    telegramId: string;
    isPrivate: boolean;
    description?: string;
    canManage?: boolean;
};

export type RoomFilter = "all" | Room;
export type VisFilter = "all" | "public" | "private";
export type ViewMode = "cards" | "table";

export type CreateBookingPayload = {
    start: string;
    end: string;
    room: Room;
    title: string;
    telegramId: string;
    isPrivate: boolean;
    description?: string;
};
