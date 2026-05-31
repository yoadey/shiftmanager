import React from 'react';
import { Icon } from './Icon';

interface SheetProps {
  onClose: () => void;
  children: React.ReactNode;
  variant?: '' | 'dialog' | 'full';
  title?: string;
  foot?: React.ReactNode;
}

export function Sheet({ onClose, children, variant = '', title, foot }: SheetProps) {
  return (
    <div
      className={'sm-overlay' + (variant === 'dialog' ? ' center' : '')}
      onClick={onClose}
    >
      <div
        className={'sm-sheet ' + variant}
        onClick={(e) => e.stopPropagation()}
      >
        {variant !== 'dialog' && variant !== 'full' && <div className="sm-grab" />}
        {title && (
          <div className="sheet-head">
            <div style={{ fontWeight: 800, fontSize: 17, fontFamily: 'Bricolage Grotesque' }}>
              {title}
            </div>
            <button
              onClick={onClose}
              className="pressable"
              style={{
                width: 34,
                height: 34,
                borderRadius: '50%',
                border: 'none',
                background: 'var(--surface-2)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
              }}
            >
              <Icon name="x" size={18} stroke={2.4} color="var(--ink-2)" />
            </button>
          </div>
        )}
        <div className="sheet-body">{children}</div>
        {foot && <div className="sheet-foot">{foot}</div>}
      </div>
    </div>
  );
}
