import { useEffect, useRef, useState } from 'react';

interface LazyImageOptions {
  threshold?: number;
  rootMargin?: string;
  placeholder?: string;
}

export const useLazyImage = (
  src: string,
  options: LazyImageOptions = {}
): [string, React.RefObject<HTMLImageElement>] => {
  const { threshold = 0.1, rootMargin = '50px', placeholder = '' } = options;
  const [imageSrc, setImageSrc] = useState(placeholder);
  const imageRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setImageSrc(src);
            observer.unobserve(entry.target);
          }
        });
      },
      { threshold, rootMargin }
    );

    const currentImage = imageRef.current;
    if (currentImage) {
      observer.observe(currentImage);
    }

    return () => {
      if (currentImage) {
        observer.unobserve(currentImage);
      }
    };
  }, [src, threshold, rootMargin]);

  return [imageSrc, imageRef];
};
