import { useAuthStore } from "../../stores/auth.store";
import type { Message } from "../../types/message";

interface MessageBubbleProps {
    message: Message;
}

function MessageBubble({ message }: MessageBubbleProps) {
    const userId = useAuthStore(
        (state) => state.userId
    );

    const isMine = message.sender_id === userId;

    const renderStatus = () => {
        if (!isMine) {
            return null;
        }

        if (message.pending) {
            return "Sending...";
        }

        switch (message.status) {
            case 1:
                return "✓";

            case 2:
                return "✓✓";

            case 3:
                return `✓✓ Seen ${new Date(
                    message.created_at
                ).toLocaleTimeString(
                    "vi-VN",
                    {
                        hour: "2-digit",
                        minute: "2-digit",
                    }
                )}`;

            default:
                return null;
        }
    };

    return (
        <div
            className={`flex ${
                isMine
                    ? "justify-end"
                    : "justify-start"
            } mb-2`}
        >
            <div
                className={`max-w-md rounded-2xl px-4 py-2 ${
                    isMine
                        ? "bg-blue-500 text-white"
                        : "bg-gray-200 text-black"
                }`}
            >
                <div>
                    {message.content}
                </div>

                {isMine && (
                    <div className="text-xs opacity-70 mt-1">
                        {renderStatus()}
                    </div>
                )}
            </div>
        </div>
    );
}

export default MessageBubble;
