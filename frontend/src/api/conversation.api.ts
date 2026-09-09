import apiClient from "../libs/axios";
import type {
    Conversation,
    CreateConversationRequest,
    CreateConversationResponse,
} from "../types/conversation";

interface ConversationResponse {
    data: Conversation[];
}

export const conversationApi = {
    getConversations: async (): Promise<Conversation[]> => {
        const response = await apiClient.get<ConversationResponse>(
            "/api/v1/chat/get_conversations",
        );
        return response.data.data;
    },

    createConversation: async (data: CreateConversationRequest): Promise<CreateConversationResponse> => {
        const response = await apiClient.post<CreateConversationResponse>(
            "/api/v1/chat/conversations",
            data,
        );

        return response.data;
    },

    getDetail: async (conversationId: number): Promise<Conversation> => {
        const response = await apiClient.get(`/api/v1/chat/conversations/${conversationId}`);

        return response.data.data
    }
};