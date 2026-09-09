import apiClient from "../libs/axios";

export type PresenceStatus = "online" | "offline"

interface PresenceResponse {
    data: Record<string, PresenceStatus>;
}

export interface LastSeenResponse {
    user_id: string;
    last_seen: string | null;
}

export const presenceApi = {
    getUsersPresence: async (userIds: string[]): Promise<Record<string, PresenceStatus>> => {
        const response = await apiClient.get<PresenceResponse>(
            "/api/v1/chat/presence", {
                params: {
                    user_ids: userIds.join(","),
                },
            },
        );
        return response.data.data;
    },

    getLastSeen: async (userId: string): Promise<LastSeenResponse> => {
        const response = await apiClient.get(
            `/api/v1/chat/last-seen/${userId}`
        );

        return response.data;
    }
}