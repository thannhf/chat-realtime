import apiClient from "../libs/axios"

export const readApi = {
    markAsRead: async (conversationId: number, messageID: number) => {
        const response = await apiClient.patch(`/api/v1/chat/conversations/${conversationId}/read`, {
            message_id: messageID,
        });

        return response.data;
    },
};