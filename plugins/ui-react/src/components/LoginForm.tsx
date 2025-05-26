import { useState } from 'react';
import { login as apiLogin } from '../api/users';
import { setToken } from '../api';

interface LoginFormProps {
  onLoginSuccess: () => void;
}

const LoginForm = ({ onLoginSuccess }: LoginFormProps) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showToken, setShowToken] = useState(false);
  const [tokenValue, setTokenValue] = useState('');
  const [recreatingAdmin, setRecreatingAdmin] = useState(false);
  const [recreateSuccess, setRecreateSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const response = await apiLogin(username, password);
      console.log("Login successful:", response);
      setToken(response.token);
      onLoginSuccess();
    } catch (err) {
      console.error('Login failed:', err);
      // Display more detailed error information for debugging
      if (err instanceof Error) {
        setError(`Login error: ${err.message}`);
      } else {
        setError('Invalid username or password. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleTokenLogin = (e: React.FormEvent) => {
    e.preventDefault();
    if (tokenValue.trim()) {
      setToken(tokenValue.trim());
      onLoginSuccess();
    } else {
      setError('Please enter a valid token');
    }
  };

  const handleRecreateAdmin = async () => {
    setRecreatingAdmin(true);
    setError(null);
    setRecreateSuccess(null);
    
    try {
      const API_BASE = import.meta.env.VITE_API_BASE || '';
      const response = await fetch(`${API_BASE}/api/auth/recreate-admin`, {
        method: 'POST',
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to recreate admin: ${response.status} ${response.statusText} - ${errorText}`);
      }
      
      await response.json(); // Parse response but we're using our own message
      setRecreateSuccess("Admin user recreated successfully. Default credentials have been set.");
      
      // Pre-fill the credentials
      setUsername('admin');
      setPassword('admin123');
    } catch (err) {
      console.error('Failed to recreate admin:', err);
      if (err instanceof Error) {
        setError(`Failed to recreate admin: ${err.message}`);
      } else {
        setError('Failed to recreate admin user. Please try again.');
      }
    } finally {
      setRecreatingAdmin(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <div className="text-center mb-8">
          <div className="flex justify-center">
            <div className="text-blue-600 bg-blue-100 p-3 rounded-full inline-block">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
              </svg>
            </div>
          </div>
          <h1 className="text-2xl font-bold text-gray-800 mt-4">Pixie Photos</h1>
          <p className="text-gray-600 mt-2">Sign in to access your photos</p>
        </div>

        {/* Error message */}
        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4" role="alert">
            <p>{error}</p>
          </div>
        )}

        {/* Success message */}
        {recreateSuccess && (
          <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded mb-4" role="alert">
            <p>{recreateSuccess}</p>
          </div>
        )}

        {/* Recreate Admin Button */}
        <div className="mb-4">
          <button
            onClick={handleRecreateAdmin}
            disabled={recreatingAdmin}
            className="w-full py-2 px-3 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            {recreatingAdmin ? (
              <span className="flex items-center justify-center">
                <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-gray-700" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Recreating admin...
              </span>
            ) : (
              'Reset Admin Account (if login fails)'
            )}
          </button>
        </div>

        {!showToken ? (
          <form onSubmit={handleSubmit}>
            <div className="mb-4">
              <label htmlFor="username" className="block text-gray-700 text-sm font-medium mb-2">
                Username
              </label>
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="shadow-sm appearance-none border rounded-md w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter your username"
                required
              />
            </div>

            <div className="mb-6">
              <label htmlFor="password" className="block text-gray-700 text-sm font-medium mb-2">
                Password
              </label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="shadow-sm appearance-none border rounded-md w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter your password"
                required
              />
            </div>

            <div className="flex items-center justify-center">
              <button
                type="submit"
                className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 w-full transition-colors"
                disabled={loading}
              >
                {loading ? (
                  <span className="flex items-center justify-center">
                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Signing in...
                  </span>
                ) : (
                  'Sign In'
                )}
              </button>
            </div>

            <div className="mt-6 text-center">
              <button
                type="button"
                className="text-blue-600 hover:text-blue-800 text-sm"
                onClick={() => setShowToken(true)}
              >
                Sign in with JWT token instead
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleTokenLogin}>
            <div className="mb-6">
              <label htmlFor="token" className="block text-gray-700 text-sm font-medium mb-2">
                JWT Token
              </label>
              <input
                type="text"
                id="token"
                value={tokenValue}
                onChange={(e) => setTokenValue(e.target.value)}
                className="shadow-sm appearance-none border rounded-md w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="Enter your JWT token"
                required
              />
              <p className="text-xs text-gray-500 mt-1">Use the generate-token.js script to create a token</p>
            </div>

            <div className="flex items-center justify-center">
              <button
                type="submit"
                className="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50 w-full transition-colors"
              >
                Sign In with Token
              </button>
            </div>

            <div className="mt-6 text-center">
              <button
                type="button"
                className="text-blue-600 hover:text-blue-800 text-sm"
                onClick={() => setShowToken(false)}
              >
                Back to username/password login
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};

export default LoginForm;
