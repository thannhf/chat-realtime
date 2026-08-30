
import { create } from "zustand";

export type PresenceStatus = "online" | "offline";

interface PresenceState {
    // Realtime presence
    presence: Record<string, PresenceStatus>;

    // Last seen timestamp
    lastSeen: Record<string, string | null>;

    // Presence
    setPresence: (
        userId: string,
        status: PresenceStatus
    ) => void;

    getPresence: (
        userId: string
    ) => PresenceStatus;

    setMultiplePresence: (
        presence: Record<string, PresenceStatus>
    ) => void;

    // Last seen
    setLastSeen: (
        userId: string,
        lastSeen: string | null
    ) => void;

    setMultipleLastSeen: (
        lastSeen: Record<string, string | null>
    ) => void;

    getLastSeen: (
        userId: string
    ) => string | null;
}

export const usePresenceStore = create<PresenceState>(
    (set, get) => ({
        presence: {},
        lastSeen: {},

        setPresence: (userId, status) => {
            set((state) => ({
                presence: {
                    ...state.presence,
                    [userId]: status,
                },
            }));
        },

        getPresence: (userId) => {
            return get().presence[userId] ?? "offline";
        },

        setMultiplePresence: (presence) => {
            set((state) => ({
                presence: {
                    ...state.presence,
                    ...presence,
                },
            }));
        },

        setLastSeen: (userId, lastSeen) => {
            set((state) => ({
                lastSeen: {
                    ...state.lastSeen,
                    [userId]: lastSeen,
                },
            }));
        },

        setMultipleLastSeen: (lastSeen) => {
            set((state) => ({
                lastSeen: {
                    ...state.lastSeen,
                    ...lastSeen,
                },
            }));
        },

        getLastSeen: (userId) => {
            return get().lastSeen[userId] ?? null;
        },
    })
);
