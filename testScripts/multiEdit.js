// client-multi-edits.js
import { io } from "socket.io-client";
import { v4 as uuidv4 } from "uuid";

// Mock JWT for testing — replace with a real one when ready
const MOCK_JWT = "your-mock-jwt-token-here";

// Document ID to test with (replace with actual existing ID in DB)
const DOCUMENT_ID = "your-document-id-here";

const socket = io("ws://localhost:3001", {
    path: "/ws",
    auth: {
        token: MOCK_JWT,
    },
});

socket.on("connect", () => {
    console.log("Connected to WebSocket server with ID:", socket.id);

    // Join the document room
    socket.emit("join_document", {
        documentId: DOCUMENT_ID,
    });

    // Start emitting edits every 2 seconds
    let editCount = 1;
    const interval = setInterval(() => {
        const newContent = `Edit #${editCount} made at ${new Date().toISOString()}`;
        console.log("Sending edit:", newContent);

        socket.emit("edit_document", {
            documentId: DOCUMENT_ID,
            content: newContent,
        });

        editCount++;

        if (editCount > 5) {
            clearInterval(interval);
            console.log("✅ Finished sending test edits. Disconnecting...");
            setTimeout(() => socket.disconnect(), 1000);
        }
    }, 2000);
});

socket.on("document_updated", (payload) => {
    console.log("📩 Received document update:", payload);
});

socket.on("connect_error", (err) => {
    console.error("❌ Connection error:", err.message);
});

socket.on("disconnect", () => {
    console.log("Disconnected from WebSocket server.");
});
