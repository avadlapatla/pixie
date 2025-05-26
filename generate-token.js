#!/usr/bin/env node

/**
 * Utility to generate a JWT token for API testing
 * 
 * Usage:
 *   node generate-token.js [subject]
 * 
 * Where:
 *   subject - Optional user ID (defaults to "test-user")
 */

// Configuration
const API_BASE = process.env.API_BASE || 'http://localhost:8080';

// Use the built-in fetch API for Node.js (requires Node.js 18+)
const generateToken = async (subject) => {
  try {
    console.log(`Generating token for subject: ${subject}`);
    
    const response = await fetch(`${API_BASE}/api/auth/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        subject: subject
      })
    });

    if (!response.ok) {
      throw new Error(`Failed to generate token: ${response.status} ${response.statusText}`);
    }

    const data = await response.json();
    return data.token;
  } catch (error) {
    console.error(`Error generating token: ${error.message}`);
    process.exit(1);
  }
};

// Main function
const main = async () => {
  // Get subject from command line or use default
  const subject = process.argv[2] || 'test-user';
  
  try {
    const token = await generateToken(subject);
    console.log('\nToken generated successfully:');
    console.log('\n' + token + '\n');
    console.log('Use this token in Authorization header:');
    console.log('Authorization: Bearer ' + token);
    console.log('\nOr in URL:');
    console.log(`${API_BASE}/api/photo/[ID]?token=${token}`);
  } catch (error) {
    console.error(error);
    process.exit(1);
  }
};

// Run the main function
main();
