import { create } from "zustand";
import { websocketService } from "../services/websocket.service";
import { useMessageStore } from "./message.store";
import { useAuthStore } from "./auth.store";
import { useConversationStore } from "./conversation.store";
import type { MessageStatusPayload, PresencePayload, ReadReceiptPayload, WSEvent } from "../types/ws";
import { usePresenceStore } from "./presence.store";

interface WebSocketState {
    connected: boolean;

    connect: (token: string) => void;
    disconnect: () => void;
    send: (data: unknown) => void;
    sendMessage: (conversationId: number, message: string) => void;
    sendMessageDelivered: (conversationId: number, messageId: number) => void;
}

export const useWebSocketStore = create<WebSocketState>((set, get) => ({
    connected: false,

    connect: (token: string) => {
        const url = `ws://localhost:8000/api/v1/chat/ws?token=${token}`;
        
        websocketService.setOnOpen(() => {
            // console.log("WebSocket Connected");
            set({connected: true});
        });

        websocketService.setOnClose(() => {
            // console.log("WebSocket Disconnected");
            set({connected: false});
        });

        websocketService.setOnMessage((event) => {
            // console.log("RAW WS:", event.data);
            
            const wsEvent = JSON.parse(event.data) as WSEvent;
            
            switch (wsEvent.type) {
                case "message": {
                    console.log("WS Event", wsEvent);
                    const message = wsEvent.data;
                    const currentUserId = useAuthStore.getState().userId;

                    if(message.sender_id !== currentUserId) {
                        get().sendMessageDelivered(
                            message.conversation_id, 
                            message.id,
                        )
                    }

                    if(message.client_msg_id && message.sender_id === currentUserId) {
                        useMessageStore.getState().replaceMessage(
                            message.client_msg_id,
                            message
                        );
                    } else {
                        useMessageStore.getState().addMessage(message);
                    }
                    // useConversationStore.getState().updateConversationPreview(message);
                    break;
                }
                case "conversation_updated": {
                    useConversationStore.getState().updateConversationPreview(wsEvent.data);

                    break;
                }
                case "read_receipt": {
                    const payload = wsEvent.data as ReadReceiptPayload;
                    const currentUserId = useAuthStore.getState().userId;
                    if (!currentUserId) {
                        break;
                    }

                    useMessageStore.getState().markMessageAsRead(
                        payload.conversation_id,
                        payload.last_read_message_id,
                        payload.user_id,
                        currentUserId,
                    )
                    // console.log("READ RECEIPT: ", payload)
                    break;
                }
                case "message_status": {
                    const payload = wsEvent.data as MessageStatusPayload
                    // console.log("MESSAGE STATUS:", payload);
                    useMessageStore.getState().updateMessageStatus(
                        payload.message_id,
                        payload.status,
                    );
                    break;
                }
                case "typing":
                    break;
                case "presence": {
                    const payload = wsEvent.data as PresencePayload;
                    usePresenceStore.getState().setPresence(
                        payload.user_id,
                        payload.status,
                    );
                    console.log("PRESENCE:", payload);
                    break;
                }
                default:
                    console.warn(
                        "Unknow websocket event:", wsEvent
                    );
            }
        });
        websocketService.connect(url);
    },

    disconnect: () => {
        websocketService.disconnect();
        set({connected: false});
    },

    send: (data: unknown) => {
        websocketService.send(data);
    },

    sendMessage: (conversationId, message) => {
        const clientMsgId = crypto.randomUUID();
        const optimisticMessage = {
            id: Date.now(),
            conversation_id: conversationId,
            sender_id: "",
            content: message,
            message_type: "text",
            status: 0,
            created_at: new Date().toISOString(),
            client_msg_id: clientMsgId,
            pending: true,
        }

        useMessageStore.getState().addOptimisticMessage(
            optimisticMessage
        );

        get().send({
            type: "message",
            data: {
                conversation_id: conversationId,
                client_msg_id: clientMsgId,
                content: message,
                message_type: "text",
            },
        });
        // console.log("Send Message")
        // console.log({conversationId, message});
    },

    sendMessageDelivered: (conversationId, messageId) => {
        get().send({
            type: "message_delivered",
            data: {
                conversation_id: conversationId,
                message_id: messageId,
            },
        });
    },
}));