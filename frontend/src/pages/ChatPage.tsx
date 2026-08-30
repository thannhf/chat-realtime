
import { useEffect } from "react";
import ChatLayout from "../components/layout/ChatLayout";
import { useConversationStore } from "../stores/conversation.store";
import { useMessageStore } from "../stores/message.store";
import { useAuthStore } from "../stores/auth.store";
import { useWebSocketStore } from "../stores/websocket.store";

export default function ChatPage() {
    // =========================
    // Conversation
    // =========================

    const fetchConversations = useConversationStore(
        (state) => state.fetchConversations
    );

    const selectedConversation = useConversationStore(
        (state) => state.selectedConversation
    );

    const setConversationDetail = useConversationStore(
        (state) => state.setConversationDetail
    );

    // =========================
    // Message
    // =========================

    const fetchHistory = useMessageStore(
        (state) => state.fetchHistory
    );

    const clearMessages = useMessageStore(
        (state) => state.clearMessages
    );

    // =========================
    // Auth
    // =========================

    const token = useAuthStore(
        (state) => state.accessToken
    );

    const hydrated = useAuthStore(
        (state) => state.hydrated
    );

    // =========================
    // WebSocket
    // =========================

    const connect = useWebSocketStore(
        (state) => state.connect
    );

    const disconnect = useWebSocketStore(
        (state) => state.disconnect
    );

    // =========================
    // Connect WebSocket
    // =========================

    useEffect(() => {
        if (!hydrated) return;
        if (!token) return;

        connect(token);
        fetchConversations();

        return () => {
            disconnect();
        };
    }, [
        hydrated,
        token,
        connect,
        disconnect,
        fetchConversations,
    ]);

    // =========================
    // Conversation changed
    // =========================

    useEffect(() => {
        if (!selectedConversation) {
            clearMessages();
            setConversationDetail(null);
            return;
        }

        // Fetch message history
        fetchHistory(
            selectedConversation.id
        );

        // Use selected conversation as detail
        setConversationDetail(
            selectedConversation
        );
    }, [
        selectedConversation,
        fetchHistory,
        clearMessages,
        setConversationDetail,
    ]);

    // =========================
    // Loading
    // =========================

    if (!hydrated) {
        return (
            <div className="flex h-screen items-center justify-center">
                Loading...
            </div>
        );
    }

    return <ChatLayout />;
}
