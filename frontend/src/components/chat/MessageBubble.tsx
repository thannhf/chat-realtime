import { useAuthStore } from "../../stores/auth.store";
import { userUserStore } from "../../stores/user.store";
import type { Message } from "../../types/message";

interface MessageBubbleProps {
    message: Message;
}

function MessageBubble({ message }: MessageBubbleProps) {
    const userId = useAuthStore((state) => state.userId);
    const users = userUserStore((state) => state.users);

    const isMine = message.sender_id === userId;

    const seenByUsers = (message.seen_by ?? [])
        .map((userId) => users[userId])
        .filter(Boolean);

    // console.log("MESSAGE:", message);
    // console.log("SEEN BY:", message.seen_by);
    // console.log("USERS:", users);
    // console.log("SEEN BY USERS:", seenByUsers);

    const renderStatus = () => {
        if (!isMine) return null;

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
                ).toLocaleTimeString("vi-VN", {
                    hour: "2-digit",
                    minute: "2-digit",
                })}`;

            default:
                return null;
        }
    };

    return (
        <div
            className={`flex ${
                isMine ? "justify-end" : "justify-start"
            } mb-2`}
        >
            <div
                className={`max-w-md rounded-2xl px-4 py-2 ${
                    isMine
                        ? "bg-blue-500 text-white"
                        : "bg-gray-200 text-black"
                }`}
            >
                <div>{message.content}</div>

                {isMine && (
                    <>
                        <div className="text-xs opacity-70 mt-1">
                            {renderStatus()}
                        </div>

                        {seenByUsers.length > 0 && (
                            <div className="flex items-center mt-1">
                                {seenByUsers.map((user) => (
                                    <div
                                        key={user.id}
                                        className="w-5 h-5 rounded-full overflow-hidden border border-white"
                                        title={user.username}
                                    >
                                        {user.avatar_url ? (
                                            <img
                                                src={user.avatar_url}
                                                alt={user.username}
                                                className="w-full h-full object-cover"
                                            />
                                        ) : (
                                            <div className="w-full h-full flex items-center justify-center bg-gray-400 text-white text-[10px] font-medium">
                                                {user.username
                                                    .charAt(0)
                                                    .toUpperCase()}
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        )}
                    </>
                )}
            </div>
        </div>
    );
}

export default MessageBubble;
