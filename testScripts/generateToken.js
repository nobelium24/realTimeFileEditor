// generateToken.js
const jwt = require("jsonwebtoken");

const secretKey = "u2hdu!d83@2u8&9ue8U*U*U98+8uU@WWI"; // Match with env.JWT_SECRET

const payload = {
    email: "testuser@example.com",
    token_type: "access",
    iss: "nobelium24",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 60 * 60 * 24 // 24 hours
};

const token = jwt.sign(payload, secretKey, { algorithm: "HS256" });

console.log("Access Token:", token);
