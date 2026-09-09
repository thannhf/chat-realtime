import MessageBubble from "./MessageBubble";
import type { Message } from "../../types/message";
import { useMessageStore } from "../../stores/message.store";
import type { RefObject } from "react";

interface Props {
    messages: Message[];
    bottomRef: RefObject<HTMLDivElement | null>;
}

export default function MessageList({messages, bottomRef}: Props) {
    const loadingHistory = useMessageStore(
        (state) => state.loadingHistory 
    )
    const hasMore = useMessageStore(
        (state) => state.hasMore
    )

    return (
        <div className="flex flex-col gap-2 p-4">
            {loadingHistory && (
                <div className="flex justify-center py-4">
                    <div className="h-6 w-6 animate-spin rounded-full border-2 border-gray-300 border-t-blue-500" />
                </div>
            )}
            {!hasMore && (
                <div className="py-4 text-center text-xs text-gray-400">
                    Beginning of conversation 
                </div>
            )}
            {messages.map((message) => (
                <MessageBubble
                    key={message.id}
                    message={message}
                />
            ))}
            <div ref={bottomRef} />
        </div>
    );
}