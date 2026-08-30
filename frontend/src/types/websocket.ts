export interface WebSocketEvent<T = unknown> {
    type: string;
    payload: T;
}