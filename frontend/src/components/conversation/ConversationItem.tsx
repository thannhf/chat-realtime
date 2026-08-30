import type { Conversation } from "../../types/conversation";
import { useConversationStore } from "../../stores/conversation.store";
import { readApi } from "../../api/read.api";

interface Props {
    conversation: Conversation;
}

export default function ConversationItem({ conversation }: Props) {
    const selectConversation = useConversationStore(
        (state) => state.setSelectedConversation
    );

    const clearUnreadCount = useConversationStore(
        (state) => state.clearUnreadCount
    );

    const handleSelectConversation = async () => {
        selectConversation(conversation);

        if (conversation.unread_count > 0) {
            try {
                if (!conversation.last_message) {
                    return;
                }

                await readApi.markAsRead(
                    conversation.id,
                    conversation.last_message.id
                );

                clearUnreadCount(conversation.id);
            } catch (error) {
                console.error(error);
            }
        }
    };

    return (
        <div
            onClick={handleSelectConversation}
            className="cursor-pointer border-b p-4 hover:bg-gray-100"
        >
            <div className="flex items-start justify-between">
                <div className="flex flex-col overflow-hidden">
                    <h3 className="font-medium">
                        {conversation.name ?? "Direct Conversation"}
                    </h3>

                    <p className="truncate text-sm text-gray-500">
                        {conversation.last_message?.content ??
                            (conversation.type === 1
                                ? "Direct"
                                : "Group")}
                    </p>
                </div>

                {conversation.last_message && (
                    <span className="ml-2 text-xs text-gray-400">
                        {new Date(
                            conversation.last_message.created_at
                        ).toLocaleTimeString([], {
                            hour: "2-digit",
                            minute: "2-digit",
                        })}
                    </span>
                )}

                {conversation.unread_count > 0 && (
                    <div className="flex min-w-5 items-center justify-center rounded-full bg-blue-500 px-2 py-0.5 text-xs font-medium text-white">
                        {conversation.unread_count > 99
                            ? "99+"
                            : conversation.unread_count}
                    </div>
                )}
            </div>
        </div>
    );
}