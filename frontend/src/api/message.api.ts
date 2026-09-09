import apiClient from "../libs/axios";
import type { Message } from "../types/message";

export interface HistoryResponse {
    data: Message[];
    next_cursor: number | null;
}

export const messageApi = {
    getHistory: async (conversationId: number, cursor?: number): Promise<HistoryResponse> => {
        const response = await apiClient.get<HistoryResponse>(
            `/api/v1/chat/history/${conversationId}`, 
            {
                params: {
                    last_message_id: cursor,
                    limit: 50,
                },
            },
        );

        return response.data;
    },
}