import React from 'react';
import { Icon } from './Icon';

type ButtonVariant = 'primary' | 'dark' | 'ghost' | 'soft' | 'danger';
type ButtonSize = 'sm' | undefined;

interface ButtonProps {
  variant?: ButtonVariant;
  size?: ButtonSize;
  icon?: string;
  children?: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  loading?: boolean;
  type?: 'button' | 'submit' | 'reset';
  style?: React.CSSProperties;
  className?: string;
}

export function Button({
  variant = 'primary',
  size,
  icon,
  children,
  onClick,
  disabled,
  loading,
  type = 'button',
  style,
  className,
}: ButtonProps) {
  const cls = ['sm-btn', variant, size, className].filter(Boolean).join(' ');
  const iconSize = size === 'sm' ? 17 : 19;

  return (
    <button
      type={type}
      className={cls}
      onClick={onClick}
      disabled={disabled || loading}
      style={style}
    >
      {icon && !loading && <Icon name={icon} size={iconSize} stroke={2.2} />}
      {loading && (
        <span style={{ width: iconSize, height: iconSize, display: 'inline-block', opacity: 0.7 }}>
          ⟳
        </span>
      )}
      {children}
    </button>
  );
}
