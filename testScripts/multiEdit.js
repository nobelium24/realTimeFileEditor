// client-multi-edits.js
const io = require('socket.io-client');

// Mock JWT for testing — replace with a real one when ready
const MOCK_JWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAcmFtLmNvbSIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJpc3MiOiJub2JlbGl1bTI0IiwiaWF0IjoxNzUzMTY0ODEzLCJleHAiOjE3NTMyNTEyMTN9.Q8aSgNNIHRhWuHkNXCeF3sz2R1UoSX7JEbCGD15kUbg";

// Document ID to test with (replace with actual existing ID in DB)
const DOCUMENT_ID = "45236245-348f-45e7-81fa-6432d6a34362";
const localUrl = "http://localhost:9091"

const socket = io(`${localUrl}/ws`, {
    extraHeaders: {
        // Authorization: `Bearer ${TOKEN}`,
        Authorization: MOCK_JWT
    },
    transports: ['websocket'],
    forceNew: true,
    reconnectionAttempts: 5
});

socket.on("connect", () => {
    console.log("Connected to WebSocket server with ID:", socket.id);

    // Join the document room
    socket.emit("join", DOCUMENT_ID);

    // Start emitting edits every 2 seconds
    let editCount = 1;
    const interval = setInterval(() => {
        const newContent = `Edit #${editCount} made at ${new Date().toISOString()}`;
        console.log("Sending edit:", newContent);

        socket.emit("edit", {
            Title: "Another one",
            ID: DOCUMENT_ID,
            Content: newContent,
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
