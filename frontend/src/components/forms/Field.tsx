import React from 'react';

interface FieldProps {
  label?: string;
  hint?: string;
  children: React.ReactNode;
}

export function Field({ label, hint, children }: FieldProps) {
  return (
    <div className="sm-field">
      {label && <label className="sm-label">{label}</label>}
      {children}
      {hint && <div className="sm-hint">{hint}</div>}
    </div>
  );
}

export function Input(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input className="sm-input" {...props} />;
}

export function Textarea(props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className="sm-textarea" {...props} />;
}

export function Select({ children, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select className="sm-select" {...props}>
      {children}
    </select>
  );
}
