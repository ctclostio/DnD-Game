import React from 'react';
import { useTranslation } from '../hooks/useTranslation';

interface SkipLinkProps {
  targetId: string;
  children?: React.ReactNode;
}

export const SkipLink: React.FC<SkipLinkProps> = ({ targetId, children }) => {
  const { t } = useTranslation();

  return (
    <a 
      href={`#${targetId}`}
      className="skip-link"
      onClick={(e) => {
        e.preventDefault();
        const target = document.getElementById(targetId);
        if (target) {
          target.focus();
          target.scrollIntoView();
        }
      }}
    >
      {children || t('a11y.skipToContent')}
    </a>
  );
};