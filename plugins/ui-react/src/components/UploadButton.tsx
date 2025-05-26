import { useState, useRef, useEffect } from 'react';
import { uploadPhoto, Photo } from '../api';

interface UploadButtonProps {
  onUploadSuccess: (photo: Photo) => void;
}

const UploadButton = ({ onUploadSuccess }: UploadButtonProps) => {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dragActive, setDragActive] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Handle drag events
  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  };

  // Handle drop event
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFiles(e.dataTransfer.files);
    }
  };

  // Handle file selection
  const handleFileInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files.length > 0) {
      handleFiles(event.target.files);
    }
  };

  // Process the selected files
  const handleFiles = async (files: FileList) => {
    if (files.length === 0) return;
    
    const validImageTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml'];
    const validVideoTypes = ['video/mp4', 'video/webm', 'video/ogg'];
    const validTypes = [...validImageTypes, ...validVideoTypes];

    // Reset the file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }

    // Process each file
    Array.from(files).forEach(async (file) => {
      // Validate file type
      if (!validTypes.includes(file.type)) {
        setError(`Invalid file type for ${file.name}. Please upload an image or video file.`);
        return;
      }

      try {
        setUploading(true);
        setError(null);
        
        // Simulate upload progress
        let progress = 0;
        const interval = setInterval(() => {
          progress += Math.floor(Math.random() * 15) + 5;
          if (progress > 95) {
            progress = 95;
            clearInterval(interval);
          }
          setUploadProgress(progress);
        }, 300);
        
        const uploadedPhoto = await uploadPhoto(file);
        
        // Complete the progress
        clearInterval(interval);
        setUploadProgress(100);
        
        // Reset progress after a short delay
        setTimeout(() => {
          setUploadProgress(0);
          setUploading(false);
        }, 500);
        
        onUploadSuccess(uploadedPhoto);
      } catch (err) {
        setError('Failed to upload photo. Please try again.');
        console.error('Upload error:', err);
        setUploading(false);
        setUploadProgress(0);
      }
    });
  };

  // Trigger the file input
  const triggerFileInput = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  // Clear error after 5 seconds
  useEffect(() => {
    if (error) {
      const timer = setTimeout(() => {
        setError(null);
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [error]);

  return (
    <div>
      {/* Hidden file input */}
      <input
        type="file"
        ref={fileInputRef}
        onChange={handleFileInputChange}
        accept="image/*,video/*"
        multiple
        className="hidden"
      />
      
      {/* Drag & drop area */}
      <div 
        className={`relative ${dragActive ? 'pointer-events-auto' : 'pointer-events-none'}`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
      >
        {dragActive && (
          <div className="fixed inset-0 bg-blue-500 bg-opacity-10 flex items-center justify-center z-50 border-2 border-blue-500 border-dashed">
            <div className="bg-white rounded-lg p-8 shadow-lg text-center">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p className="mt-4 text-lg font-medium text-gray-700">Drop files to upload</p>
            </div>
          </div>
        )}
      </div>
      
      {/* Upload button */}
      <div className="relative">
        <button
          onClick={triggerFileInput}
          disabled={uploading}
          className="flex items-center px-4 py-2 rounded-full bg-blue-600 text-white font-medium hover:bg-blue-700 transition-colors shadow-md"
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zM6.293 6.707a1 1 0 010-1.414l3-3a1 1 0 011.414 0l3 3a1 1 0 01-1.414 1.414L11 5.414V13a1 1 0 11-2 0V5.414L7.707 6.707a1 1 0 01-1.414 0z" clipRule="evenodd" />
          </svg>
          Upload
        </button>
        
        {/* Upload progress indicator */}
        {uploading && (
          <div className="absolute left-0 -bottom-8 w-full">
            <div className="h-1 w-full bg-gray-200 rounded-full overflow-hidden">
              <div 
                className="h-full bg-blue-600 transition-all duration-300 ease-out"
                style={{ width: `${uploadProgress}%` }}
              ></div>
            </div>
            <p className="text-xs text-gray-500 mt-1 text-center">
              {uploadProgress < 100 ? `Uploading ${uploadProgress}%` : 'Processing...'}
            </p>
          </div>
        )}
      </div>
      
      {/* Error message */}
      {error && (
        <div className="mt-2 text-sm text-red-600 bg-red-50 p-2 rounded border border-red-200">
          <div className="flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            {error}
          </div>
        </div>
      )}
    </div>
  );
};

export default UploadButton;
