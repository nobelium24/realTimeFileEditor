// client.js
const io = require('socket.io-client');
// import { v4 as uuidv4 } from 'uuid';

const DOC_ID = "45236245-348f-45e7-81fa-6432d6a34362"; // simulate a random doc ID
const USER_ID = "20b89c43-2009-4220-b032-23c758820cb2";
const TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3RAcmFtLmNvbSIsInRva2VuX3R5cGUiOiJhY2Nlc3MiLCJpc3MiOiJub2JlbGl1bTI0IiwiaWF0IjoxNzUzMTY0ODEzLCJleHAiOjE3NTMyNTEyMTN9.Q8aSgNNIHRhWuHkNXCeF3sz2R1UoSX7JEbCGD15kUbg";
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
        ID: DOC_ID,
        Title: 'Demo Document',
        Content: 'This is a test content update from client.js',
        UserID: USER_ID,
        CreatedAt: new Date().toISOString(),
        UpdatedAt: new Date().toISOString(),
    };


    console.log('Sending edit event...');
    socket.emit('edit', updatedDoc);
});

socket.on('document_updated', (data) => {
    console.log('Document update received:', data);
});

socket.on('connected', (msg) => {
    console.log(' Server says:', msg);
});

socket.on('disconnect', (reason) => {
    console.log(`Disconnected from server: ${reason}`);
});

socket.on('connect_error', (err) => {
    console.error('❗ Connection error:', err.message);
});
