// client.js
const io = require('socket.io-client');
// import { v4 as uuidv4 } from 'uuid';

const DOC_ID = "45236245-348f-45e7-81fa-6432d6a34362"; // simulate a random doc ID
const USER_ID = "20b89c43-2009-4220-b032-23c758820cb2";
const TOKEN = "";
const localUrl = "http://localhost:9091";

const socket = io(`${localUrl}/ws`, {
    extraHeaders: {
        // Authorization: `Bearer ${TOKEN}`,
        Authorization: TOKEN
    },
    transports: ['websocket'],
    forceNew: true,
    reconnectionAttempts: 5
});


socket.on('connect', () => {
    console.log('✅ Connected to WebSocket server');

    // Join a document room
    socket.emit('join', DOC_ID);

    // Simulate edit
    const updatedDoc = {
        id: DOC_ID,
        title: 'Demo Document',
        content: 'This is a test content update from client.js',
        userId: USER_ID,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
    };

    console.log('📤 Sending edit event...');
    socket.emit('edit', updatedDoc);
});

socket.on('document_updated', (data) => {
    console.log('📥 Document update received:', data);
});

socket.on('connected', (msg) => {
    console.log('ℹ️ Server says:', msg);
});

socket.on('disconnect', () => {
    console.log('❌ Disconnected from server');
});

socket.on('connect_error', (err) => {
    console.error('❗ Connection error:', err.message);
});
