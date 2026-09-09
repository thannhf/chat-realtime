import { create } from "zustand";
import { messageApi } from "../api/message.api";
import type { Message } from "../types/message";
import { userUserStore } from "./user.store";

interface MessageState {
    messages: Message[];
    loading: boolean;
    nextCursor: number | null;
    hasMore: boolean;
    loadingHistory: boolean;

    fetchHistory: (conversationId: number) => Promise<void>;
    clearMessages: () => void;
    addMessage: (message: Message) => void;
    addOptimisticMessage: (message: Message) => void;
    replaceMessage: (clientMsgId: string, message: Message) => void;
    loadMoreHistory: (conversationId: number) => Promise<void>;
    markMessageAsRead: (
        conversationId: number,
        lastReadMessageId: number,
        readerUserId: string,
        currentUserId: string,
    ) => void;
    updateMessageStatus: (messageId: number, status: number) => void;
    fetchSeenByUsers: () => Promise<void>;
}

export const useMessageStore = create<MessageState>((set, get) => ({
    messages: [],
    loading: false,
    nextCursor: null,
    hasMore: true,
    loadingHistory: false,

    fetchHistory: async (conversationId) => {
        console.log("FETCH HISTORY CALLED", conversationId);
        set({loading: true});

        try {
            const response = await messageApi.getHistory(conversationId);
            // console.log(
            //     "FETCH HISTORY",
            //     messages
            // );

            set({
                messages: response.data,
                nextCursor: response.next_cursor,
                hasMore: response.next_cursor !== null,
            });

            await get().fetchSeenByUsers();
        } finally {
            set({loading: false});
        }
    },

    clearMessages: () => {
        set({
            messages: [],
            nextCursor: null,
            hasMore: true,
        });
    },

    addMessage: (message) => {
        // console.log(
        //     "ADD REALTIME MESSAGE",
        //     message
        // );
        set((state) => ({
            messages: [
                ...state.messages,
                message,
            ],
        }));

        void get().fetchSeenByUsers();
    },

    addOptimisticMessage: (message) => {
        set((state) => ({
            messages:[
                ...state.messages,
                message,
            ],
        }));
    },

    replaceMessage: (clientMsgId, message) => {
        set((state) => ({
            messages: state.messages.map((item) => item.client_msg_id === clientMsgId ? {...message, pending: false} : item)
        }));
    },

    loadMoreHistory: async (ConversationSidebar) => {
        const state = useMessageStore.getState();

        if(!state.hasMore || state.loadingHistory) return;
        set({loadingHistory: true});

        try {
            const response = await messageApi.getHistory(ConversationSidebar, state.nextCursor ?? undefined);

           if (response.data.length === 0) {
                set({
                    hasMore: false,
                    nextCursor: null,
                    loadingHistory: false,
                });
                return;
            }

            set((state) => ({
                messages: [
                    ...response.data,
                    ...state.messages,
                ],
                nextCursor: response.next_cursor,
                hasMore: response.next_cursor !== null,
            }));

            await get().fetchSeenByUsers();

            set({
                loadingHistory: false,
            });
        } catch (err) {
            set({loadingHistory: false});
            throw err;
        }
    },

    markMessageAsRead: (
        conversationId,
        lastReadMessageId,
        readerUserId,
        currentUserId,
    ) => {
        if (readerUserId === currentUserId) {
            return;
        }

        set((state) => ({
            messages: state.messages.map((message) => {
                if (
                    message.conversation_id !== conversationId ||
                    message.sender_id !== currentUserId ||
                    message.id > lastReadMessageId
                ) {
                    return message;
                }

                const seenBy = message.seen_by ?? [];

                if (seenBy.includes(readerUserId)) {
                    return message;
                }

                return {
                    ...message,
                    status: 3,
                    seen_by: [...seenBy, readerUserId],
                };
            }),
        }));
    },

    updateMessageStatus: (messageId, status) => {
        set((state) => {
            // console.log("UPDATE STATUS:", {
            //     messageId,
            //     status,
            //     messages: state.messages.map((m) => ({
            //         id: m.id,
            //         status: m.status,
            //     })),
            // });

            return {
                messages: state.messages.map((message) =>
                    message.id === messageId
                        ? {
                            ...message,
                            status,
                        }
                        : message
                ),
            };
        });
    },

    fetchSeenByUsers: async () => {
        const messages = get().messages;
        const userIds = [
            ...new Set(
                messages.flatMap((message) => message.seen_by ?? []),
            ),
        ];
        // console.log("FETCH SEEN USER IDS:", userIds);

        if (userIds.length === 0) {
            return;
        }

        await userUserStore.getState().fetchUsers(userIds);
    },
}));