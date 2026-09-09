import { useConversationStore } from "../../stores/conversation.store";
import ConversationItem from "./ConversationItem";

export default function ConversationList() {
    const conversations = useConversationStore(
        (state) => state.conversations
    );

    return (
        <div>
            {conversations.map((conversation) => (
                <ConversationItem
                    key={conversation.id}
                    conversation={conversation}
                />
            ))}
        </div>
    );
}