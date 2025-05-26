import { useState, useEffect, useRef } from 'react';

interface HeaderProps {
  onSearch: (query: string) => void;
  onLogout: () => void;
  onMenuClick: () => void;
}

const Header = ({ onSearch, onLogout, onMenuClick }: HeaderProps) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Close menu when clicking outside
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setIsUserMenuOpen(false);
      }
    };
    
    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(searchQuery);
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    setSearchQuery(query);
    
    // Debounce search as user types
    const timer = setTimeout(() => {
      onSearch(query);
    }, 300);
    
    return () => clearTimeout(timer);
  };

  const handleClearSearch = () => {
    setSearchQuery('');
    onSearch('');
  };

  return (
    <header className="bg-white shadow-sm sticky top-0 z-20">
      <div className="max-w-screen-2xl mx-auto px-4 py-2.5">
        <div className="flex justify-between items-center">
          {/* Left: Menu Button (Mobile) and Logo */}
          <div className="flex items-center">
            <button 
              className="lg:hidden mr-3 p-2 rounded-full text-gray-500 hover:bg-gray-100 focus:outline-none"
              onClick={onMenuClick}
              aria-label="Menu"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <span className="text-xl font-semibold text-blue-600 flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 mr-2" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
              </svg>
              <span className="hidden sm:inline">Pixie</span>
            </span>
          </div>
          
          {/* Middle: Search Bar */}
          <div className="flex-1 max-w-xl mx-4">
            <form onSubmit={handleSearchSubmit} className="relative">
              <div className="flex items-center bg-gray-100 rounded-full px-3 py-2 focus-within:ring-2 focus-within:ring-blue-500 focus-within:bg-white transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                  type="text"
                  className="w-full bg-transparent border-none outline-none px-3 text-gray-800 placeholder-gray-500"
                  placeholder="Search photos..."
                  value={searchQuery}
                  onChange={handleSearchChange}
                />
                {searchQuery && (
                  <button
                    type="button"
                    onClick={handleClearSearch}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </button>
                )}
              </div>
            </form>
          </div>
          
          {/* Right: User Menu */}
          <div className="relative" ref={userMenuRef}>
            <button
              className="flex items-center space-x-2 focus:outline-none p-1.5 rounded-full hover:bg-gray-100"
              onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
              aria-label="User menu"
            >
              <div className="w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center font-medium">
                U
              </div>
              <svg 
                xmlns="http://www.w3.org/2000/svg" 
                className={`h-5 w-5 text-gray-500 transition-transform ${isUserMenuOpen ? 'transform rotate-180' : ''}`} 
                viewBox="0 0 20 20" 
                fill="currentColor"
              >
                <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
            </button>
            
            {isUserMenuOpen && (
              <div className="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5">
                <div className="py-1" role="menu" aria-orientation="vertical">
                  <button
                    className="w-full text-left block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    role="menuitem"
                    onClick={onLogout}
                  >
                    <div className="flex items-center">
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                      </svg>
                      Sign Out
                    </div>
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;
