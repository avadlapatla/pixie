import { useState, lazy, Suspense, useEffect } from 'react';
import { Photo } from './api';
import LoginForm from './components/LoginForm';
import Gallery from './components/Gallery';
import UploadButton from './components/UploadButton';
import Header from './components/Header';
import SideNav from './components/SideNav';
import { SearchProvider, useSearch } from './context/SearchContext';
import { AlbumsProvider } from './context/AlbumsContext';
import AlbumsPage from './components/pages/AlbumsPage';
import TrashPage from './components/pages/TrashPage';
import AdminPage from './components/pages/AdminPage';

// Lazy load the Lightbox component to reduce initial bundle size
const Lightbox = lazy(() => import('./components/Lightbox'));

function AppContent() {
  const [authenticated, setAuthenticated] = useState(false);
  
  // Clear any existing tokens and force login when app starts
  useEffect(() => {
    // Remove any existing token to force authentication
    localStorage.removeItem('token');
    setAuthenticated(false);
  }, []);
  const [selectedPhoto, setSelectedPhoto] = useState<Photo | null>(null);
  const [newPhoto, setNewPhoto] = useState<Photo | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [activeView, setActiveView] = useState<'photos' | 'albums' | 'trash' | 'admin'>('photos');
  // Use this to trigger a refresh when a photo is trashed
  const [galleryRefreshTrigger, setGalleryRefreshTrigger] = useState<number>(0);
  const { searchQuery, setSearchQuery } = useSearch();

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
    // Clear the token from localStorage
    localStorage.removeItem('token');
    setAuthenticated(false);
  };

  const toggleSidebar = () => {
    setSidebarOpen(!sidebarOpen);
  };

  const handleNavigation = (view: 'photos' | 'albums' | 'trash' | 'admin') => {
    setActiveView(view);
    if (window.innerWidth < 1024) { // On mobile
      setSidebarOpen(false);
    }
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
    return <LoginForm onLoginSuccess={() => setAuthenticated(true)} />;
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
      <SideNav 
        isOpen={sidebarOpen} 
        onClose={() => setSidebarOpen(false)}
        activeView={activeView}
        onNavigate={handleNavigation}
      />
      
      {/* Main Content */}
      <main className={`transition-all duration-300 ${sidebarOpen ? 'lg:ml-64' : ''}`}>
        <div className="container mx-auto px-4 py-6">
          {activeView === 'photos' && (
            <>
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
                refreshTrigger={galleryRefreshTrigger}
              />
            </>
          )}
          {activeView === 'albums' && (
            <AlbumsPage onPhotoClick={handlePhotoClick} />
          )}
          {activeView === 'trash' && (
            <TrashPage onPhotoClick={handlePhotoClick} />
          )}
          {activeView === 'admin' && (
            <AdminPage />
          )}
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
          onTrash={() => setGalleryRefreshTrigger(prev => prev + 1)}
        />
      </Suspense>
    </div>
  );
}

function App() {
  return (
    <SearchProvider>
      <AlbumsProvider>
        <AppContent />
      </AlbumsProvider>
    </SearchProvider>
  );
}

export default App;
