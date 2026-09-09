import { useEffect } from "react";
import { useConversationStore } from "../../stores/conversation.store";
import ConversationList from "./ConversationList";

function ConversationSidebar() {
    const fetchConversations = useConversationStore(
        (state) => state.fetchConversations
    );

    const loading = useConversationStore(
        (state) => state.loading
    );

    useEffect(() => {
        fetchConversations();
    }, [fetchConversations]);

    if (loading) {
        return <div>Loading...</div>;
    }

    return (
        <aside className="w-80 border-r">
            <div className="p-4">
                <input
                    className="w-full rounded-lg border px-4 py-2"
                    placeholder="Search conversations..."
                />
            </div>

            <ConversationList />
        </aside>
    );
}

export default ConversationSidebar;