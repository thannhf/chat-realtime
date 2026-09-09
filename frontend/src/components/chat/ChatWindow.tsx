import { useEffect, useRef, useState } from "react";
import { useConversationStore } from "../../stores/conversation.store";
import { useMessageStore } from "../../stores/message.store";
import MessageList from "./MessageList";
import { readApi } from "../../api/read.api";
import ChatHeader from "./ChatHeader";
import { conversationApi } from "../../api/conversation.api";

export default function ChatWindow() {
    const [hasNewMessage, setHasNewMessage] = useState(false);
    const containerRef = useRef<HTMLDivElement | null>(null);
    const isAtBottomRef = useRef(true);
    const bottomRef = useRef<HTMLDivElement>(null);

    const messages = useMessageStore(
        (state) => state.messages
    );

    const selectedConversation = useConversationStore(
        state => state.selectedConversation
    );

    const setConversationDetail = useConversationStore(
        state => state.setConversationDetail
    )

    const loadMoreHistory = useMessageStore(
        (state) => state.loadMoreHistory
    );

    const hasMore = useMessageStore(
        (state) => state.hasMore
    );

    const loadingHistory = useMessageStore(
        (state) => state.loadingHistory
    );

    const handleScroll = async () => {
        const container = containerRef.current;

        if(!container) return;

        const distanceFromBottom =
        container.scrollHeight -
        container.scrollTop -
        container.clientHeight;

        const atBottom = distanceFromBottom < 80

        isAtBottomRef.current = atBottom;

        if (atBottom) {
            setHasNewMessage(false);
        }

        if(container.scrollTop <= 20 && hasMore && !loadingHistory && selectedConversation) {
            const previousHeight = container.scrollHeight
            

            // console.log("At bottom:", isAtBottomRef.current);

            await loadMoreHistory(selectedConversation.id)

            requestAnimationFrame(() => {
                if(!containerRef.current) return;
                const newHeight = containerRef.current.scrollHeight;

                container.scrollTop += newHeight - previousHeight;
            })
        }
    }

    useEffect(() => {
        const c = containerRef.current;
        if (!c) return;
    }, [messages]);

    useEffect(() => {
        if(!selectedConversation){
            setConversationDetail(null);
            return;
        }

        conversationApi
            .getDetail(selectedConversation.id)
            .then((data)=>{
                setConversationDetail(data);
            })
            .catch((err)=>{
                console.error(
                    "Get conversation detail failed",
                    err
                );
            });
    }, [
        selectedConversation,
        setConversationDetail
    ]);

    useEffect(() => {
        if(!selectedConversation) return;
        if(messages.length === 0) return;

        const lastMessage = messages[messages.length - 1];

        readApi
            .markAsRead(
                selectedConversation.id,
                lastMessage.id
            )
            .catch(err=>{
                console.error(
                    "Mark as read failed",
                    err
                );
            });
    }, [
        selectedConversation,
        messages
    ]);

    useEffect(() => {
        if(!isAtBottomRef.current) {
            setHasNewMessage(true);
            return;
        }
        bottomRef.current?.scrollIntoView({
            behavior: "smooth",
        });
    }, [messages]);

    if(!selectedConversation) {
        return (
            <div className="flex flex-1 items-center justify-center">
                <div className="text-center text-gray-500">
                    <div className="text-5xl mb-4">
                        💬
                    </div>
                    <h2 className="text-xl font-semibold">
                        Select a conversation 
                    </h2>
                    <p className="mt-2">
                        Choose a chat to start messaging
                    </p>
                </div>
            </div>
        )
    }

    return (
        <div className="relative flex flex-1 min-h-0 flex-col">
            <ChatHeader />

            <div ref={containerRef} onScroll={handleScroll} className="min-h-0 flex-1 overflow-y-auto">
                <MessageList messages={messages} bottomRef={bottomRef} />
            </div>

            {hasNewMessage && (
                <button onClick={() => {
                    bottomRef.current?.scrollIntoView({
                        behavior: "smooth"
                    });

                    setHasNewMessage(false);
                }} className="absolute bottom-6 left-1/2 -translate-x-1/2 rounded-full bg-blue-500 px-4 py-2 text-white shadow-lg">
                    ↓ New messages
                </button>
            )}
        </div>
    );
}