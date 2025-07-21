// client.js
import { io } from 'socket.io-client';
import { v4 as uuidv4 } from 'uuid';

const DOC_ID = uuidv4(); // simulate a random doc ID
const USER_ID = uuidv4(); // simulate a user ID
const TOKEN = "mocked.jwt.token"; // Replace with a real JWT later

const socket = io('http://localhost:3001/ws', {
    extraHeaders: {
        Authorization: `Bearer ${TOKEN}`,
    },
    transports: ['websocket'],
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
