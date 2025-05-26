interface SideNavProps {
  isOpen: boolean;
  onClose: () => void;
  activeView: 'photos' | 'albums' | 'trash';
  onNavigate: (view: 'photos' | 'albums' | 'trash') => void;
}

const SideNav = ({ isOpen, onClose, activeView, onNavigate }: SideNavProps) => {
  return (
    <>
      {/* Backdrop for mobile */}
      {isOpen && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 z-30 lg:hidden"
          onClick={onClose}
        ></div>
      )}
      
      <aside
        className={`fixed top-0 left-0 h-full bg-white z-40 shadow-xl transition-transform duration-300 transform ${
          isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        } w-64 py-4 flex flex-col`}
        style={{ 
          marginTop: '60px', // Match the header height
          height: 'calc(100% - 60px)'
        }}
      >
        <div className="px-4 mb-6">
          <h2 className="text-lg font-medium text-gray-800">Pixie Photos</h2>
        </div>

        <nav className="flex-1">
          <ul>
            <li>
              <a
                href="#"
                className={`flex items-center px-4 py-3 text-gray-800 hover:bg-blue-50 transition-colors border-l-4 ${
                  activeView === 'photos' ? 'border-blue-600 bg-blue-50' : 'border-transparent'
                }`}
                onClick={(e) => {
                  e.preventDefault();
                  onNavigate('photos');
                }}
              >
                <svg xmlns="http://www.w3.org/2000/svg" className={`h-5 w-5 mr-3 ${
                  activeView === 'photos' ? 'text-blue-600' : 'text-gray-500'
                }`} viewBox="0 0 20 20" fill="currentColor">
                  <path d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" />
                </svg>
                Photos
              </a>
            </li>
            <li>
              <a
                href="#"
                className={`flex items-center px-4 py-3 text-gray-700 hover:bg-blue-50 transition-colors border-l-4 ${
                  activeView === 'albums' ? 'border-blue-600 bg-blue-50' : 'border-transparent'
                }`}
                onClick={(e) => {
                  e.preventDefault();
                  onNavigate('albums');
                }}
              >
                <svg xmlns="http://www.w3.org/2000/svg" className={`h-5 w-5 mr-3 ${
                  activeView === 'albums' ? 'text-blue-600' : 'text-gray-500'
                }`} viewBox="0 0 20 20" fill="currentColor">
                  <path d="M5 4a2 2 0 012-2h6a2 2 0 012 2v14l-5-2.5L5 18V4z" />
                </svg>
                Albums
              </a>
            </li>
            <li>
              <a
                href="#"
                className="flex items-center px-4 py-3 text-gray-700 hover:bg-blue-50 transition-colors border-l-4 border-transparent"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-3 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clipRule="evenodd" />
                </svg>
                Shared
              </a>
            </li>
            <li>
              <a
                href="#"
                className="flex items-center px-4 py-3 text-gray-700 hover:bg-blue-50 transition-colors border-l-4 border-transparent"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-3 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
                </svg>
                Recent
              </a>
            </li>
            <li>
              <a
                href="#"
                className="flex items-center px-4 py-3 text-gray-700 hover:bg-blue-50 transition-colors border-l-4 border-transparent"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-3 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
                </svg>
                Archive
              </a>
            </li>
            <li>
              <a
                href="#"
                className={`flex items-center px-4 py-3 text-gray-700 hover:bg-blue-50 transition-colors border-l-4 ${
                  activeView === 'trash' ? 'border-blue-600 bg-blue-50' : 'border-transparent'
                }`}
                onClick={(e) => {
                  e.preventDefault();
                  onNavigate('trash');
                }}
              >
                <svg xmlns="http://www.w3.org/2000/svg" className={`h-5 w-5 mr-3 ${
                  activeView === 'trash' ? 'text-blue-600' : 'text-gray-500'
                }`} viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
                Trash
              </a>
            </li>
          </ul>
        </nav>

        <div className="mt-auto p-4 border-t">
          <div className="flex items-center text-sm text-gray-600">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
            </svg>
            <span>Storage Policy</span>
          </div>
        </div>
      </aside>
    </>
  );
};

export default SideNav;
