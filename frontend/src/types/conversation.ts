import type { Message } from "./message";

export const ConversationType = {
    Direct: 1,
    Group: 2,
} as const;

export type ConversationType = (typeof ConversationType)[keyof typeof ConversationType]

export interface ConversationMember {
    id: number;
    conversation_id: number;
    user_id: string;
    role: string;
    joined_at: string;
}

export interface Conversation {
    id: number;
    name?: string | null;
    type: ConversationType;
    created_at: string;
    updated_at: string;
    members?: ConversationMember[];

    avatar?: string;
    last_message?: Message;
    unread_count: number;
}

export interface ConversationUpdatedPayload {
    conversation_id: number;
    last_message: Message;
    updated_at: string;
}

export interface CreateConversationRequest {
    type: ConversationType;
    member_ids: string[];
    name?: string;
}

export interface CreateConversationResponse {
    message: string;
    conversation: Conversation;
}

export type GetConversationsResponse = Conversation[];