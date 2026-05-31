/* Shared UI primitives + app context */
const { createContext, useContext, useState, useEffect, useRef } = React;

const AppCtx = createContext(null);
const useApp = () => useContext(AppCtx);

// ── icon set (stroke) ──
const ICONS = {
  home: 'M3 11.2 12 4l9 7.2M5 9.8V20h5v-6h4v6h5V9.8',
  compass: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM16 8l-2.5 5.5L8 16l2.5-5.5L16 8Z',
  calendar: 'M7 3v3m10-3v3M3.5 9.5h17M5 5h14a1.5 1.5 0 0 1 1.5 1.5V19A1.5 1.5 0 0 1 19 20.5H5A1.5 1.5 0 0 1 3.5 19V6.5A1.5 1.5 0 0 1 5 5Z',
  clock: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM12 7v5l3.5 2',
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8ZM4.5 20a7.5 7.5 0 0 1 15 0',
  users: 'M9 11a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7ZM2.5 19.5a6.5 6.5 0 0 1 13 0M16 4.2a3.5 3.5 0 0 1 0 6.8M18 13.5a6.5 6.5 0 0 1 3.5 6',
  plus: 'M12 5v14M5 12h14',
  minus: 'M5 12h14',
  check: 'M5 12.5 10 17.5 19.5 7',
  x: 'M6 6l12 12M18 6 6 18',
  chevL: 'M15 5l-7 7 7 7',
  chevR: 'M9 5l7 7-7 7',
  chevD: 'M6 9l6 6 6-6',
  bell: 'M6 9a6 6 0 0 1 12 0c0 5 2 6 2 6H4s2-1 2-6ZM9.5 19a2.5 2.5 0 0 0 5 0',
  settings: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6ZM12 2.5l1.5 2.5 2.9-.4 .4 2.9 2.5 1.5-1.3 2.6 1.3 2.6-2.5 1.5-.4 2.9-2.9-.4L12 21.5l-1.5-2.5-2.9.4-.4-2.9L4.7 15l1.3-2.6L4.7 9.8l2.5-1.5.4-2.9 2.9.4Z',
  search: 'M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14ZM20 20l-4-4',
  edit: 'M4 20h4L18.5 9.5a2 2 0 0 0-2.8-2.8L5 17.5V20ZM14 8l3 3',
  trash: 'M4 7h16M9 7V5h6v2M6 7l1 13h10l1-13',
  pin: 'M12 21s7-6 7-11a7 7 0 1 0-14 0c0 5 7 11 7 11ZM12 12a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z',
  spark: 'M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8L12 3Z',
  chart: 'M4 20V10M10 20V4M16 20v-7M22 20H2',
  shield: 'M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6l7-3Z',
  mail: 'M3.5 6.5h17v11h-17zM4 7l8 6 8-6',
  info: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM12 11v6M12 7.5v.5',
  arrowR: 'M5 12h14M13 6l6 6-6 6',
  filter: 'M3 5h18l-7 8v6l-4 2v-8L3 5Z',
  download: 'M12 4v11M7.5 10.5 12 15l4.5-4.5M5 19.5h14',
  lock: 'M7 10V8a5 5 0 0 1 10 0v2M5.5 10.5h13v9h-13zM12 14v2',
  tag: 'M3 12V5a2 2 0 0 1 2-2h7l9 9-9 9-9-9ZM7.5 8.5h.01',
  list: 'M8 6h13M8 12h13M8 18h13M3.5 6h.01M3.5 12h.01M3.5 18h.01',
  layers: 'M12 3 2 8l10 5 10-5-10-5ZM2 13l10 5 10-5M2 18l10 5 10-5',
  hours: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM12 7v5l3 2',
  euro: 'M16 8a5 5 0 1 0 0 8M5 10h7M5 14h6',
  ban: 'M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20ZM5.5 5.5l13 13',
  star: 'M12 3l2.6 6.3 6.8.5-5.2 4.4 1.7 6.6L12 17.7 6.3 21.3 8 14.7 2.8 10.3l6.8-.5L12 3Z',
};
function Icon({ name, size = 22, stroke = 2, color, style, fill = 'none' }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill={fill} style={style}
      stroke={color || 'currentColor'} strokeWidth={stroke} strokeLinecap="round" strokeLinejoin="round">
      <path d={ICONS[name] || ''} />
    </svg>
  );
}

function Avatar({ id, name, size = 38 }) {
  const sm = window.SM;
  let initials;
  if (id && sm.M[id]) initials = sm.M[id].first[0] + sm.M[id].last[0];
  else if (name) initials = name.split(' ').map(w => w[0]).slice(0, 2).join('');
  else initials = '?';
  return <div className="sm-avatar" style={{ width: size, height: size, fontSize: size * 0.4 }}>{initials}</div>;
}

function Badge({ kind = 'neutral', children, dot = true }) {
  return <span className={`sm-badge b-${kind}`}>{dot && <span className="dot" />}{children}</span>;
}
function OccBadge({ o }) {
  const map = { ok: 'ok', warn: 'warn', crit: 'crit', full: 'full' };
  return <Badge kind={map[o.key]}>{o.label}</Badge>;
}

function Button({ variant = 'primary', size, icon, children, onClick, disabled, type, style }) {
  return (
    <button type={type || 'button'} className={`sm-btn ${variant}${size ? ' ' + size : ''}`} onClick={onClick} disabled={disabled} style={style}>
      {icon && <Icon name={icon} size={size === 'sm' ? 17 : 19} stroke={2.2} />}
      {children}
    </button>
  );
}

// two-segment hour progress (confirmed + reserved) toward goal
function HourBar({ confirmed, reserved, goal }) {
  const c = Math.min(confirmed / goal, 1) * 100;
  const r = Math.min(Math.max(reserved - confirmed, 0) / goal, 1 - c / 100) * 100;
  return (
    <div className="sm-track">
      <div className="seg-c" style={{ width: c + '%' }} />
      <div className="seg-r" style={{ width: r + '%' }} />
    </div>
  );
}
function OccFill({ o }) {
  const colors = { ok: 'var(--ok)', warn: 'var(--warn)', crit: 'var(--crit)', full: 'var(--ink-2)' };
  return (
    <div className="occ-track">
      <div className="occ-fill" style={{ width: Math.max(o.count / o.max * 100, o.count ? 8 : 0) + '%', background: colors[o.key] }} />
    </div>
  );
}

// ── form controls ──
function Field({ label, hint, children }) {
  return (
    <div className="sm-field">
      {label && <label className="sm-label">{label}</label>}
      {children}
      {hint && <div className="sm-hint">{hint}</div>}
    </div>
  );
}
function Input(props) { return <input className="sm-input" {...props} />; }
function Textarea(props) { return <textarea className="sm-textarea" {...props} />; }
function Select({ children, ...p }) { return <select className="sm-select" {...p}>{children}</select>; }

function Stepper({ value, set, min = 0, max = 20, label }) {
  const btn = (d, ic) => (
    <button className="pressable" onClick={() => set(Math.max(min, Math.min(max, value + d)))}
      style={{ width: 40, height: 40, borderRadius: 12, border: '1.5px solid var(--line-2)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', color: 'var(--ink)' }}>
      <Icon name={ic} size={18} stroke={2.4} />
    </button>
  );
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
      {btn(-1, 'minus')}
      <div style={{ minWidth: 30, textAlign: 'center', fontWeight: 800, fontSize: 19, fontFamily: 'Bricolage Grotesque' }}>{value}</div>
      {btn(1, 'plus')}
      {label && <span style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600 }}>{label}</span>}
    </div>
  );
}

function Toggle({ on, onClick }) {
  return (
    <button onClick={onClick} className="pressable" style={{
      width: 50, height: 30, borderRadius: 999, border: 'none', cursor: 'pointer', padding: 3,
      background: on ? 'var(--primary)' : 'var(--line-2)', transition: 'background .18s', flexShrink: 0,
    }}>
      <div style={{ width: 24, height: 24, borderRadius: '50%', background: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,.25)', transform: on ? 'translateX(20px)' : 'translateX(0)', transition: 'transform .18s' }} />
    </button>
  );
}

// ── overlays ──
function Sheet({ onClose, children, variant = '', title, foot }) {
  return (
    <div className={'sm-overlay' + (variant === 'dialog' ? ' center' : '')} onClick={onClose}>
      <div className={'sm-sheet ' + variant} onClick={e => e.stopPropagation()}>
        {variant !== 'dialog' && variant !== 'full' && <div className="sm-grab" />}
        {title && (
          <div className="sheet-head">
            <div style={{ fontWeight: 800, fontSize: 17, fontFamily: 'Bricolage Grotesque' }}>{title}</div>
            <button onClick={onClose} className="pressable" style={{ width: 34, height: 34, borderRadius: '50%', border: 'none', background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}>
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

// section header within scroll
function Section({ title, action }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', margin: '22px 0 11px' }}>
      <h3 style={{ fontSize: 17.5, fontWeight: 700 }}>{title}</h3>
      {action}
    </div>
  );
}

Object.assign(window, {
  AppCtx, useApp, Icon, Avatar, Badge, OccBadge, Button, HourBar, OccFill,
  Field, Input, Textarea, Select, Stepper, Toggle, Sheet, Section,
});
