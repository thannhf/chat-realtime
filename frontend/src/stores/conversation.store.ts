import { create } from "zustand";
import { conversationApi } from "../api/conversation.api";
import type { Conversation, ConversationUpdatedPayload } from "../types/conversation";
import { useAuthStore } from "./auth.store";

interface ConversationState {
    conversations: Conversation[];
    selectedConversation: Conversation | null;
    loading: boolean;
    conversationDetail: Conversation | null,

    fetchConversations: () => Promise<void>;
    setSelectedConversation: (
        conversation: Conversation
    ) => void;

    clearSelectedConversation: () => void;
    addConversation: (conversation: Conversation) => void;
    updateConversationPreview: (payload: ConversationUpdatedPayload) => void;
    clearUnreadCount: (conversationID: number) => void;
    setConversationDetail: (conversation: Conversation | null) => void;
}

export const useConversationStore = create<ConversationState>((set) => ({
    conversations: [],
    selectedConversation: null,
    loading: false,
    conversationDetail: null,

    fetchConversations: async () => {
        set({ loading: true });

        try {
            const conversations = await conversationApi.getConversations();

            set({
                conversations,
            });
        } finally {
            set({
                loading: false,
            });
        }
    },

    setSelectedConversation: (
        conversation: Conversation,
    ) => {
        set({
            selectedConversation: conversation,
        });
    },

    clearSelectedConversation: () => {
        set({
            selectedConversation: null,
        });
    },

    addConversation: (
        conversation: Conversation,
    ) => {
        set((state) => ({
            conversations: [
                conversation,
                ...state.conversations,
            ],
        }));
    },


    updateConversationPreview: (
        payload: ConversationUpdatedPayload
    ) => {
        set((state) => {
            const conversations = state.conversations
                .map((conversation) => {

                    if (conversation.id !== payload.conversation_id) {
                        return conversation;
                    }
                    const currentUserId = useAuthStore.getState().userId;

                    const isMine = payload.last_message.sender_id === currentUserId;

                    const isCurrentConversation =
                        state.selectedConversation?.id ===
                        payload.conversation_id;

                    return {
                        ...conversation,
                        last_message: payload.last_message,
                        updated_at: payload.updated_at,

                        unread_count: isCurrentConversation || isMine
                            ? 0
                            : conversation.unread_count + 1,
                    };
                }).sort(
                    (a, b) =>
                        new Date(b.updated_at).getTime() -
                        new Date(a.updated_at).getTime()
                );

            return {
                conversations,
            };
        });
    },

    clearUnreadCount: (conversationID) => {
        set((state) => ({
            conversations: state.conversations.map((conversation) => conversation.id === conversationID ? {...conversation, unread_count: 0} : conversation)
        }))
    },

    setConversationDetail: (conversation: Conversation | null) => {
        set({
            conversationDetail: conversation,
        })
    }
}));