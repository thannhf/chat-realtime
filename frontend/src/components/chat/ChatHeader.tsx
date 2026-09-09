import { useEffect } from "react";
import { useConversationStore } from "../../stores/conversation.store";
import { useAuthStore } from "../../stores/auth.store";
import { usePresenceStore } from "../../stores/presence.store";
import { presenceApi } from "../../api/presence.api";

function formatLastSeen(lastSeen: string | null): string {
    if (!lastSeen) {
        return "Offline";
    }

    const date = new Date(lastSeen);
    const now = new Date();

    const diffMs = now.getTime() - date.getTime();
    const diffSeconds = Math.floor(diffMs / 1000);

    if (diffSeconds < 60) {
        return "Last seen just now";
    }

    const diffMinutes = Math.floor(diffSeconds / 60);

    if (diffMinutes < 60) {
        return `Last seen ${diffMinutes} minute${
            diffMinutes > 1 ? "s" : ""
        } ago`;
    }

    const diffHours = Math.floor(diffMinutes / 60);

    if (diffHours < 24) {
        return `Last seen ${diffHours} hour${
            diffHours > 1 ? "s" : ""
        } ago`;
    }

    const diffDays = Math.floor(diffHours / 24);

    if (diffDays === 1) {
        return "Last seen yesterday";
    }

    if (diffDays < 7) {
        return `Last seen ${diffDays} days ago`;
    }

    return `Last seen ${date.toLocaleDateString()}`;
}

export default function ChatHeader() {
    const conversationDetail = useConversationStore(
        (state) => state.conversationDetail
    );

    const currentUserId = useAuthStore(
        (state) => state.userId
    );

    const setLastSeen = usePresenceStore(
        (state) => state.setLastSeen
    );

    const isGroup = conversationDetail?.type === 2;

    const opponentUserId =
        conversationDetail && !isGroup
            ? conversationDetail.members?.find(
                  (member) =>
                      member.user_id !== currentUserId
              )?.user_id
            : undefined;

    const opponentPresence = usePresenceStore(
        (state) =>
            opponentUserId
                ? state.presence[opponentUserId]
                : undefined
    );

    const opponentLastSeen = usePresenceStore(
        (state) =>
            opponentUserId
                ? state.lastSeen[opponentUserId] ?? null
                : null
    );

    useEffect(() => {
        if (!opponentUserId) {
            return;
        }

        if (opponentPresence === "online") {
            return;
        }

        const fetchLastSeen = async () => {
            try {
                const response =
                    await presenceApi.getLastSeen(
                        opponentUserId
                    );

                setLastSeen(
                    opponentUserId,
                    response.last_seen
                );
            } catch (error) {
                console.error(
                    "Failed to fetch last seen:",
                    error
                );
            }
        };

        fetchLastSeen();
    }, [
        opponentUserId,
        opponentPresence,
        setLastSeen,
    ]);

    if (!conversationDetail) {
        return (
            <div className="border-b p-4">
                Loading...
            </div>
        );
    }

    const isOnline =
        opponentPresence === "online";

    return (
        <div className="flex items-center justify-between border-b p-4">
            <div className="flex items-center gap-3">
                {/* Avatar */}
                <div className="relative">
                    <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-300">
                        {isGroup ? "👥" : "👤"}
                    </div>

                    {/* Online indicator */}
                    {!isGroup && isOnline && (
                        <span
                            className="
                                absolute
                                bottom-0
                                right-0
                                h-3
                                w-3
                                rounded-full
                                border-2
                                border-white
                                bg-green-500
                            "
                        />
                    )}
                </div>

                {/* Conversation information */}
                <div>
                    <h2 className="font-semibold">
                        {isGroup
                            ? conversationDetail.name ??
                              "Group"
                            : "Direct Conversation"}
                    </h2>

                    {/* Direct conversation */}
                    {!isGroup && opponentUserId && (
                        <p
                            className={`text-sm ${
                                isOnline
                                    ? "text-green-500"
                                    : "text-gray-500"
                            }`}
                        >
                            {isOnline
                                ? "Online"
                                : formatLastSeen(
                                      opponentLastSeen
                                  )}
                        </p>
                    )}

                    {/* Group conversation */}
                    {isGroup && (
                        <p className="text-sm text-gray-500">
                            {conversationDetail.members
                                ?.length ?? 0}{" "}
                            members
                        </p>
                    )}
                </div>
            </div>

            {/* More button */}
            <div>
                <button
                    type="button"
                    className="rounded p-2 hover:bg-gray-100"
                >
                    ⋮
                </button>
            </div>
        </div>
    );
}
