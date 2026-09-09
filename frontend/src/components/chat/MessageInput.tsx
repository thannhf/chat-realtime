import { useEffect, useRef, useState } from "react";

interface MessageInputProps {
    onSend: (message: string) => void;
    disabled?: boolean;
}

function MessageInput({onSend, disabled = false}: MessageInputProps) {
    const [message, setMessage] = useState("");

    const inputRef = useRef<HTMLTextAreaElement | null>(null);

    useEffect(() => {
        if(!disabled) {
            inputRef.current?.focus();
        }
    }, [disabled]);

    const handleSend = () => {
        const text = message.trim();

        if(!text || disabled) return;
        onSend(text);
        setMessage("");
    }

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if(e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    return (
        <div className="flex items-center gap-2 border-t p-4">
            <textarea
                ref={inputRef}
                value={message}
                disabled={disabled}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={disabled ? "Select a conversation first..." : "Type a message..."}
                className="flex-1 resize-none rounded-lg border px-4 py-2 focus:outline-none disabled:bg-gray-100"
            />

            <button
                disabled={disabled}
                onClick={handleSend}
                className="rounded-lg bg-blue-500 px-4 py-2 text-white disabled:bg-gray-300 disabled:cursor-not-allowed"
            >
                Send
            </button>

        </div>
    )
}

export default MessageInput