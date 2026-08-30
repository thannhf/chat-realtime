import AppHeader from "./AppHeader"
import ConversationSidebar from "../conversation/ConversationSidebar"
import ChatWindow from "../chat/ChatWindow"
import MessageInput from "../chat/MessageInput"
import { useConversationStore } from "../../stores/conversation.store"
import { useWebSocketStore } from "../../stores/websocket.store"


function ChatLayout() {
    const selectedConversation = useConversationStore(
        (state) => state.selectedConversation
    );

    const sendMessage = useWebSocketStore(
        (state) => state.sendMessage
    );

    const connected = useWebSocketStore(
        state => state.connected
    )

    return (
        <div className="flex h-screen flex-col">
            <AppHeader />

            <div className="flex flex-1 min-h-0">
                <ConversationSidebar />
                <ChatWindow />
            </div>
            <MessageInput disabled={!selectedConversation || !connected} onSend={(message) => {
                if (!selectedConversation) return;
                sendMessage(
                    selectedConversation.id,
                    message,
                );
            }} />
        </div>
    )
}

export default ChatLayout