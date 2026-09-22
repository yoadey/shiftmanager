import React from 'react';
import type { ShiftOccupancy } from '@/types';

type BadgeKind = 'ok' | 'warn' | 'crit' | 'full' | 'info' | 'neutral' | 'primary';

interface BadgeProps {
  kind?: BadgeKind;
  dot?: boolean;
  children: React.ReactNode;
  style?: React.CSSProperties;
}

export function Badge({ kind = 'neutral', dot = true, children, style }: BadgeProps) {
  return (
    <span className={`sm-badge b-${kind}`} style={style}>
      {dot && <span className="dot" />}
      {children}
    </span>
  );
}

interface OccBadgeProps {
  o: ShiftOccupancy;
}

const occKindMap: Record<string, BadgeKind> = {
  ok: 'ok',
  warn: 'warn',
  crit: 'crit',
  full: 'full',
};

export function OccBadge({ o }: OccBadgeProps) {
  return <Badge kind={occKindMap[o.key] ?? 'neutral'}>{o.label}</Badge>;
}
