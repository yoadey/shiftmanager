import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Icon } from '@/components/ui/Icon';

interface ErrorBoundaryProps {
  /** Short label describing the section, e.g. "Übersicht" or "die Mitgliederliste". */
  label?: string;
  /** Optional custom fallback renderer; receives the error and a reset callback. */
  fallback?: (error: Error, reset: () => void) => ReactNode;
  children: ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
}

/**
 * Catches render/runtime errors in its subtree and shows a contained error
 * panel instead of letting the whole app go blank. Each major section of the
 * UI is wrapped in its own boundary so a failure in one place never takes the
 * rest of the app down — the user still sees (and can use) everything else.
 *
 * React only catches errors thrown during render/lifecycle here; async errors
 * (e.g. failed queries) are handled separately by each screen's query state.
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    // Surface the failure in the console for debugging without crashing the app.
    // eslint-disable-next-line no-console
    console.error(`[ErrorBoundary${this.props.label ? ` · ${this.props.label}` : ''}]`, error, info.componentStack);
  }

  reset = (): void => this.setState({ error: null });

  render(): ReactNode {
    const { error } = this.state;
    if (!error) return this.props.children;

    if (this.props.fallback) return this.props.fallback(error, this.reset);

    const where = this.props.label ? ` in ${this.props.label}` : '';
    return (
      <div
        className="fade-in"
        role="alert"
        style={{
          margin: 16,
          padding: 18,
          borderRadius: 'var(--radius-sm, 13px)',
          border: '1px solid var(--crit, #D6453D)',
          background: 'var(--crit-bg, #FBE7E5)',
          color: 'var(--ink, #1C1A16)',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 9, marginBottom: 6 }}>
          <Icon name="info" size={18} color="var(--crit, #D6453D)" />
          <strong style={{ fontSize: 15 }}>Ein Fehler ist aufgetreten{where}.</strong>
        </div>
        <div style={{ fontSize: 13, color: 'var(--ink-2, #4B463C)', fontWeight: 600, marginBottom: 12 }}>
          Dieser Bereich konnte nicht angezeigt werden. Der Rest der App
          funktioniert weiterhin.
        </div>
        <details style={{ fontSize: 12, color: 'var(--muted, #8A8475)', marginBottom: 12 }}>
          <summary style={{ cursor: 'pointer', fontWeight: 700 }}>Details</summary>
          <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', marginTop: 8 }}>
            {error.message || String(error)}
          </pre>
        </details>
        <button
          onClick={this.reset}
          className="pressable"
          style={{
            appearance: 'none',
            border: '1px solid var(--line-2, #E2DBCB)',
            background: 'var(--surface, #fff)',
            color: 'var(--ink, #1C1A16)',
            borderRadius: 999,
            padding: '8px 16px',
            fontWeight: 700,
            fontSize: 13.5,
            cursor: 'pointer',
          }}
        >
          Erneut versuchen
        </button>
      </div>
    );
  }
}
