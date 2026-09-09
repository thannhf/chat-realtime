class WebSocketService {
    private socket: WebSocket | null = null;
    private onOpen?: () => void;
    private onClose?: () => void;
    private onMessage?: (event: MessageEvent) => void;

    setOnOpen(callback: () => void) {
        this.onOpen = callback;
    }

    setOnClose(callback: () => void) {
        this.onClose = callback;
    }

    setOnMessage(callback: (event: MessageEvent) => void) {
        this.onMessage = callback;
    }

    connect(url: string) {
        if (this.socket && (this.socket.readyState === WebSocket.OPEN ||
        this.socket.readyState === WebSocket.CONNECTING)) {
            return;
        }
        this.socket = new WebSocket(url);

        this.socket.onopen = () => {
            this.onOpen?.();
        };

        this.socket.onmessage = (event) => {
            this.onMessage?.(event);
        };

        this.socket.onerror = () => {};

        this.socket.onclose = () => {
            this.onClose?.();
        };
    }

    disconnect() {
        if (!this.socket) {
            return;
        }
        if (
            this.socket.readyState === WebSocket.OPEN ||
            this.socket.readyState === WebSocket.CONNECTING
        ) {
            this.socket.close();
        }

        this.socket = null;
    }

    send(data: unknown) {
        if(!this.socket) return;
        this.socket.send(JSON.stringify(data));
    }

    getSocket() {
        return this.socket;
    }
}

export const websocketService = new WebSocketService();