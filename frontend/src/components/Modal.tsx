import React, { useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { useFocusTrap } from '../hooks/useFocusTrap';
import { useTranslation } from '../hooks/useTranslation';
import { useAccessibility } from './AccessibilityProvider';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  size?: 'small' | 'medium' | 'large';
  showCloseButton?: boolean;
  closeOnOverlayClick?: boolean;
  className?: string;
  role?: 'dialog' | 'alertdialog';
  ariaDescribedBy?: string;
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  children,
  size = 'medium',
  showCloseButton = true,
  closeOnOverlayClick = true,
  className = '',
  role = 'dialog',
  ariaDescribedBy,
}) => {
  const { t } = useTranslation();
  const { settings } = useAccessibility();
  const modalRef = useFocusTrap<HTMLDivElement>({
    enabled: isOpen,
    returnFocus: true,
    allowOutsideClick: false,
  });

  // Handle escape key
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === 'Escape' && isOpen) {
      onClose();
    }
  }, [isOpen, onClose]);

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      // Announce modal opening to screen readers
      const announcement = `${title} dialog opened`;
      // This would be announced by the AccessibilityProvider
    } else {
      document.body.style.overflow = '';
    }

    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen, title]);

  // Add overlay click handler after mount
  useEffect(() => {
    if (!isOpen || !closeOnOverlayClick) return;

    const handleClick = (e: MouseEvent) => {
      if ((e.target as HTMLElement).dataset.overlay === 'true') {
        onClose();
      }
    };

    document.addEventListener('click', handleClick);
    return () => document.removeEventListener('click', handleClick);
  }, [isOpen, closeOnOverlayClick, onClose]);

  if (!isOpen) return null;

  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (closeOnOverlayClick && (e.target as HTMLElement).dataset.overlay === 'true') {
      onClose();
    }
  };

  const modalContent = (
    <div 
      className={`modal-overlay ${settings.reduceMotion ? 'no-animation' : ''}`}
      data-overlay="true"
    >
      <div
        ref={modalRef}
        className={`modal modal-${size} ${className}`}
        role={role}
        aria-modal="true"
        aria-labelledby="modal-title"
        aria-describedby={ariaDescribedBy}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => {
          if (closeOnOverlayClick && e.key === 'Escape') {
            onClose();
          }
        }}
        tabIndex={-1}
      >
        <div className="modal-header">
          <h2 id="modal-title" className="modal-title">{title}</h2>
          {showCloseButton && (
            <button
              type="button"
              className="modal-close"
              onClick={onClose}
              aria-label={t('a11y.closeDialog')}
            >
              <span aria-hidden="true">&times;</span>
            </button>
          )}
        </div>
        <div className="modal-body">
          {children}
        </div>
      </div>
    </div>
  );

  return createPortal(modalContent, document.body);
};

// Confirmation Modal with better accessibility
interface ConfirmModalProps {
  isOpen: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  isDangerous?: boolean;
}

export const ConfirmModal: React.FC<ConfirmModalProps> = ({
  isOpen,
  onConfirm,
  onCancel,
  title,
  message,
  confirmText,
  cancelText,
  isDangerous = false,
}) => {
  const { t } = useTranslation();

  return (
    <Modal
      isOpen={isOpen}
      onClose={onCancel}
      title={title}
      size="small"
      role="alertdialog"
      ariaDescribedBy="confirm-message"
    >
      <div className="confirm-modal">
        <p id="confirm-message" className="confirm-message">{message}</p>
        <div className="confirm-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={onCancel}
          >
            {cancelText || t('common.cancel')}
          </button>
          <button
            type="button"
            className={`btn ${isDangerous ? 'btn-danger' : 'btn-primary'}`}
            onClick={onConfirm}
            autoFocus
          >
            {confirmText || t('common.confirm')}
          </button>
        </div>
      </div>
    </Modal>
  );
};