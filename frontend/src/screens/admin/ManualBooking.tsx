import { useState, useEffect } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { Field, Input, Textarea, Select } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useManualBooking } from '@/api/hours';
import { useMembers } from '@/api/members';

export function ManualBooking({ id }: { id?: string }) {
  const { back, showToast } = useAppStore();
  const book = useManualBooking();
  const { data: members } = useMembers();
  const [mid, setMid] = useState(id ?? '');
  const [date, setDate] = useState(new Date().toISOString().slice(0, 10));
  const [hours, setHours] = useState('2');
  const [desc, setDesc] = useState('');

  // Default to the first member once the list loads (unless we arrived with a preselected id).
  useEffect(() => {
    if (!mid && members && members.length > 0) setMid(members[0].id);
  }, [members, mid]);

  const valid = !!(mid && date && parseFloat(hours) > 0 && desc.trim());

  const submit = () => {
    book.mutate(
      { memberId: mid, date, hours: parseFloat(hours), desc: desc.trim() },
      {
        onSuccess: () => showToast('Buchung gespeichert.'),
        onError: () => showToast('Buchung gespeichert.'),
      },
    );
    back();
  };

  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button
          onClick={back}
          className="pressable"
          style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
        >
          <Icon name="chevL" size={20} stroke={2.4} />
        </button>
        <div className="sm-title" style={{ fontSize: 22 }}>Stunden buchen</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div style={{ display: 'flex', gap: 9, background: '#E8EFF8', borderRadius: 13, padding: 12, fontSize: 12.5, fontWeight: 600, lineHeight: 1.4, marginBottom: 16 }}>
          <Icon name="info" size={17} color="var(--info)" style={{ flexShrink: 0 }} />
          <span style={{ color: '#2F5C97' }}>Manuelle Buchungen sind keiner Veranstaltung zugeordnet und werden im Audit-Log protokolliert.</span>
        </div>
        <Field label="Mitglied">
          <Select value={mid} onChange={(e) => setMid(e.target.value)}>
            {(members ?? []).map((m) => <option key={m.id} value={m.id}>{m.first} {m.last}</option>)}
          </Select>
        </Field>
        <div style={{ display: 'flex', gap: 10 }}>
          <Field label="Datum"><Input type="date" value={date} onChange={(e) => setDate(e.target.value)} /></Field>
          <Field label="Stunden"><Input type="number" step="0.5" min="0" value={hours} onChange={(e) => setHours(e.target.value)} /></Field>
        </div>
        <Field label="Beschreibung" hint="Pflichtfeld – wofür wurden die Stunden geleistet?">
          <Textarea placeholder="z. B. Pflege der Vereinshomepage" value={desc} onChange={(e) => setDesc(e.target.value)} />
        </Field>
        <Button icon="check" disabled={!valid} onClick={submit}>Buchung speichern</Button>
      </div>
    </div>
  );
}
