import { Icon } from '@/components/ui/Icon';

/** Lightweight inline loading state used by list/detail screens. */
export function LoadingState({ label = 'Lädt…' }: { label?: string }) {
  return (
    <div
      className="fade-in"
      style={{ padding: '40px 20px', textAlign: 'center', color: 'var(--muted)', fontWeight: 600, fontSize: 14 }}
    >
      <span
        className="sm-spinner"
        style={{
          display: 'inline-block', width: 22, height: 22, marginBottom: 10,
          border: '2.5px solid var(--line-2)', borderTopColor: 'var(--primary)',
          borderRadius: '50%', animation: 'sm-spin .7s linear infinite',
        }}
      />
      <div>{label}</div>
    </div>
  );
}

/** Friendly empty/error state with an icon, heading and supporting text. */
export function MessageState({ icon, title, text }: { icon: string; title: string; text: string }) {
  return (
    <div style={{ textAlign: 'center', padding: '34px 20px' }}>
      <div style={{ width: 58, height: 58, borderRadius: '50%', background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
        <Icon name={icon} size={26} color="var(--muted)" />
      </div>
      <div style={{ fontWeight: 700, fontSize: 16 }}>{title}</div>
      <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600, marginTop: 4, maxWidth: 250, marginInline: 'auto' }}>{text}</div>
    </div>
  );
}

/** Standard error state for failed queries. */
export function ErrorState({ text = 'Daten konnten nicht geladen werden. Bitte versuche es später erneut.' }: { text?: string }) {
  return <MessageState icon="info" title="Etwas ist schiefgelaufen" text={text} />;
}
