export interface Message {
    id: number;
    conversation_id: number;
    sender_id: string;
    content: string;
    message_type: string;
    status: number;
    created_at: string;
    client_msg_id?: string;
    pending?: boolean;
    seen_by?: string[];
}