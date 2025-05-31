import React, { useState } from 'react';
import { useLazyImage } from '@hooks/useLazyImage';

interface LazyImageProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  src: string;
  alt: string;
  placeholder?: string;
  fallback?: string;
  onLoad?: () => void;
  onError?: () => void;
}

const LazyImage: React.FC<LazyImageProps> = ({
  src,
  alt,
  placeholder = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300"%3E%3Crect width="400" height="300" fill="%23374151"/%3E%3C/svg%3E',
  fallback = '/images/placeholder.png',
  onLoad,
  onError,
  className = '',
  ...rest
}) => {
  const [hasError, setHasError] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const [imageSrc, imageRef] = useLazyImage(src, { placeholder });

  const handleLoad = () => {
    setIsLoaded(true);
    onLoad?.();
  };

  const handleError = () => {
    setHasError(true);
    onError?.();
  };

  const imageClasses = `
    ${className}
    ${!isLoaded ? 'loading' : ''}
    ${hasError ? 'error' : ''}
  `.trim();

  return (
    <div className="lazy-image-container">
      <img
        ref={imageRef}
        src={hasError ? fallback : imageSrc}
        alt={alt}
        className={imageClasses}
        onLoad={handleLoad}
        onError={handleError}
        loading="lazy"
        {...rest}
      />
      {!isLoaded && !hasError && (
        <div className="image-loading-overlay">
          <div className="spinner-small"></div>
        </div>
      )}
    </div>
  );
};

export default LazyImage;