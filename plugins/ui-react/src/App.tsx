import { useState, lazy, Suspense } from 'react';
import { Photo, isAuthenticated, setToken } from './api';
import Gallery from './components/Gallery';
import UploadButton from './components/UploadButton';

// Lazy load the Lightbox component to reduce initial bundle size
const Lightbox = lazy(() => import('./components/Lightbox'));

function App() {
  const [authenticated, setAuthenticated] = useState(isAuthenticated());
  const [token, setTokenValue] = useState('');
  const [selectedPhoto, setSelectedPhoto] = useState<Photo | null>(null);
  const [newPhoto, setNewPhoto] = useState<Photo | null>(null);

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    if (token.trim()) {
      setToken(token.trim());
      setAuthenticated(true);
    }
  };

  const handlePhotoClick = (photo: Photo) => {
    setSelectedPhoto(photo);
  };

  const handleUploadSuccess = (photo: Photo) => {
    setNewPhoto(photo);
  };

  const closeLightbox = () => {
    setSelectedPhoto(null);
  };

  if (!authenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100 p-4">
        <div className="max-w-md w-full bg-white rounded-lg shadow-md p-8">
          <h1 className="text-2xl font-bold text-center mb-6">Pixie Photo Gallery</h1>
          <form onSubmit={handleLogin}>
            <div className="mb-4">
              <label htmlFor="token" className="block text-gray-700 text-sm font-bold mb-2">
                JWT Token
              </label>
              <input
                type="text"
                id="token"
                value={token}
                onChange={(e) => setTokenValue(e.target.value)}
                className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                placeholder="Enter your JWT token"
                required
              />
            </div>
            <div className="flex items-center justify-center">
              <button
                type="submit"
                className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Sign In
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <header className="mb-8">
        <h1 className="text-3xl font-bold text-center">Pixie Photo Gallery</h1>
      </header>

      <UploadButton onUploadSuccess={handleUploadSuccess} />
      
      <Gallery onPhotoClick={handlePhotoClick} newPhoto={newPhoto} />
      
      <Suspense fallback={<div>Loading...</div>}>
        <Lightbox photo={selectedPhoto} onClose={closeLightbox} />
      </Suspense>
    </div>
  );
}

export default App;
