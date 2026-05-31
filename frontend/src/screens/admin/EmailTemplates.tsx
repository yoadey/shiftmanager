import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { Input, Textarea, Field } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useEmailTemplates, useUpdateEmailTemplate } from '@/api/settings';
import { LoadingState, MessageState, ErrorState } from '@/components/ui/States';
import type { EmailTemplate } from '@/types';

// Human-readable labels for the seeded template names (domain/email.go).
const TEMPLATE_LABELS: Record<string, string> = {
  'kiosk-confirmation': 'Kiosk-Bestätigung',
  'shift-confirmation': 'Schicht-Bestätigung',
  'shift-deregistration': 'Schicht-Abmeldung',
  'reminder-1w': 'Erinnerung (1 Woche)',
  'reminder-1d': 'Erinnerung (1 Tag)',
  'shift-cancelled': 'Schicht abgesagt',
  'year-billing': 'Jahresabrechnung',
  'missing-hours-warning': 'Fehlstunden-Warnung',
  'shift-understaffed': 'Schicht unterbesetzt',
};

// Available Go text/template placeholders shown as a hint.
const PLACEHOLDERS = '{{.Name}} · {{.EventName}} · {{.ShiftName}} · {{.Date}} · {{.Link}}';

function templateLabel(name: string): string {
  return TEMPLATE_LABELS[name] ?? name;
}

export function EmailTemplates() {
  const { back, showToast } = useAppStore();
  const { data, isLoading, isError } = useEmailTemplates();
  const update = useUpdateEmailTemplate();

  const templates = data ?? [];
  const [selected, setSelected] = useState<EmailTemplate | null>(null);
  const [subject, setSubject] = useState('');
  const [body, setBody] = useState('');

  const open = (t: EmailTemplate) => {
    setSelected(t);
    setSubject(t.subject);
    setBody(t.body);
  };

  const save = () => {
    if (!selected) return;
    update.mutate(
      { name: selected.name, subject, body },
      {
        onSuccess: () => { showToast('Vorlage gespeichert.'); setSelected(null); },
        onError: () => showToast('Vorlage konnte nicht gespeichert werden.', 'crit'),
      },
    );
  };

  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button
          onClick={() => (selected ? setSelected(null) : back())}
          className="pressable"
          style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
        >
          <Icon name="chevL" size={20} stroke={2.4} />
        </button>
        <div className="sm-title" style={{ fontSize: 22 }}>{selected ? templateLabel(selected.name) : 'E-Mail-Vorlagen'}</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        {isLoading && <LoadingState />}
        {isError && <ErrorState />}
        {!isLoading && !isError && !selected && templates.length === 0 && (
          <MessageState icon="mail" title="Keine Vorlagen" text="Es sind noch keine E-Mail-Vorlagen hinterlegt." />
        )}

        {!selected && templates.length > 0 && (
          <div className="sm-card" style={{ overflow: 'hidden' }}>
            {templates.map((t, i) => (
              <div key={t.name}>
                {i > 0 && <hr className="sm-divider" />}
                <div
                  className="pressable"
                  onClick={() => open(t)}
                  style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 15px', cursor: 'pointer' }}
                >
                  <div style={{ width: 36, height: 36, borderRadius: 10, background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                    <Icon name="mail" size={18} color="var(--ink-2)" />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, fontSize: 14.5 }}>{templateLabel(t.name)}</div>
                    <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{t.subject}</div>
                  </div>
                  <Icon name="chevR" size={18} color="var(--line-2)" stroke={2.4} />
                </div>
              </div>
            ))}
          </div>
        )}

        {selected && (
          <div className="sm-card pad">
            <Field label="Betreff">
              <Input value={subject} onChange={(e) => setSubject(e.target.value)} />
            </Field>
            <Field label="Text" hint={`Platzhalter: ${PLACEHOLDERS}`}>
              <Textarea value={body} onChange={(e) => setBody(e.target.value)} rows={10} style={{ minHeight: 200, fontFamily: 'monospace', fontSize: 13 }} />
            </Field>
            <Button icon="check" loading={update.isPending} onClick={save} style={{ marginTop: 6 }}>Speichern</Button>
          </div>
        )}
      </div>
    </div>
  );
}
