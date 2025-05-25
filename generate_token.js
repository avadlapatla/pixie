const crypto = require('crypto');

// JWT parts
const header = {
  alg: 'HS256',
  typ: 'JWT'
};

const payload = {
  sub: 'demo',
  exp: Math.floor(Date.now() / 1000) + 86400, // 24 hours from now
  iat: Math.floor(Date.now() / 1000) // issued at time
};

// Convert to base64url
function base64url(obj) {
  return Buffer.from(JSON.stringify(obj))
    .toString('base64')
    .replace(/=/g, '')
    .replace(/\+/g, '-')
    .replace(/\//g, '_');
}

// Create JWT parts
const headerEncoded = base64url(header);
const payloadEncoded = base64url(payload);

// Create signature
const secret = 'supersecret123';
const signature = crypto
  .createHmac('sha256', secret)
  .update(`${headerEncoded}.${payloadEncoded}`)
  .digest('base64')
  .replace(/=/g, '')
  .replace(/\+/g, '-')
  .replace(/\//g, '_');

// Combine to form JWT
const token = `${headerEncoded}.${payloadEncoded}.${signature}`;
console.log("Generated JWT Token:");
console.log(token);
console.log("\nToken Details:");
console.log("Header:", header);
console.log("Payload:", payload);
console.log("Secret:", secret);
