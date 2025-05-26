import { useState, lazy, Suspense, useEffect } from 'react';
import { Photo, isAuthenticated, setToken } from './api';
import Gallery from './components/Gallery';
import UploadButton from './components/UploadButton';
import Header from './components/Header';
import SideNav from './components/SideNav';
import { SearchProvider, useSearch } from './context/SearchContext';

// Lazy load the Lightbox component to reduce initial bundle size
const Lightbox = lazy(() => import('./components/Lightbox'));

function AppContent() {
  const [authenticated, setAuthenticated] = useState(isAuthenticated());
  const [token, setTokenValue] = useState('');
  const [selectedPhoto, setSelectedPhoto] = useState<Photo | null>(null);
  const [newPhoto, setNewPhoto] = useState<Photo | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { searchQuery, setSearchQuery } = useSearch();

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

  const handleSearchChange = (query: string) => {
    setSearchQuery(query);
  };
  
  const handleDeletePhoto = (photoId: string) => {
    // If deleted photo is currently selected, close lightbox
    if (selectedPhoto && selectedPhoto.id === photoId) {
      setSelectedPhoto(null);
    }
  };

  const handleLogout = () => {
    setAuthenticated(false);
  };

  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  // Close sidebar when clicking outside on mobile
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 1024) { // lg breakpoint
        setSidebarOpen(true);
      } else {
        setSidebarOpen(false);
      }
    };

    // Set initial state
    handleResize();

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  if (!authenticated) {
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
          <form onSubmit={handleLogin}>
            <div className="mb-6">
              <label htmlFor="token" className="block text-gray-700 text-sm font-medium mb-2">
                JWT Token
              </label>
              <input
                type="text"
                id="token"
                value={token}
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
                Sign In
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <Header 
        onSearch={handleSearchChange} 
        onLogout={handleLogout}
        onMenuClick={toggleSidebar}
      />
      
      {/* Sidebar */}
      <SideNav isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      
      {/* Main Content */}
      <main className={`transition-all duration-300 ${sidebarOpen ? 'lg:ml-64' : ''}`}>
        <div className="container mx-auto px-4 py-6">
          <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-6 space-y-4 md:space-y-0">
            <h2 className="text-2xl font-medium text-gray-800">
              {searchQuery ? `Search results for "${searchQuery}"` : "My Photos"}
            </h2>
            <UploadButton onUploadSuccess={handleUploadSuccess} />
          </div>
          
          <Gallery 
            onPhotoClick={handlePhotoClick} 
            newPhoto={newPhoto} 
            searchQuery={searchQuery}
            onDeletePhoto={handleDeletePhoto}
          />
        </div>
      </main>
      
      {/* Lightbox */}
      <Suspense fallback={
        <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-75 z-50">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-white"></div>
        </div>
      }>
        <Lightbox 
          photo={selectedPhoto} 
          onClose={closeLightbox} 
          onDelete={handleDeletePhoto}
        />
      </Suspense>
    </div>
  );
}

function App() {
  return (
    <SearchProvider>
      <AppContent />
    </SearchProvider>
  );
}

export default App;
