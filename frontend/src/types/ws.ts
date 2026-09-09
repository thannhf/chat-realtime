import type { Message } from "./message";
import type { ConversationUpdatedPayload } from "./conversation";

export type WSEvent =
    | {
        type: "message";
        data: MessageEventData;
    }
    | {
        type: "conversation_updated";
        data: ConversationUpdatedPayload;
    }
    | {
        type: "read_receipt";
        data: ReadReceiptPayload;
    }
    | {
        type: "typing";
        data: unknown;
    }
    | {
        type: "presence";
        data: unknown;
    }
    | {
        type: "message_status";
        data: MessageStatusPayload;
    };

export interface MessageEventData extends Message {
    client_msg_id?: string;
}

export interface ReadReceiptPayload {
    conversation_id: number;
    user_id: string;
    last_read_message_id: number;
}

export interface MessageStatusPayload {
    message_id: number;
    conversation_id: number;
    user_id: string;
    status: number;
}

export interface PresencePayload {
    user_id: string;
    status: "online" | "offline";
}