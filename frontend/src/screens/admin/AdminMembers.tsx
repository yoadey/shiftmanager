import React, { useState, useRef } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Field, Input } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useMembers, useCreateMember, useExportMembersCSV, useImportMembersCSV, useImportMembersPreview } from '@/api/members';
import type { ImportPreviewResult } from '@/api/members';
import { useEvents } from '@/api/events';
import { useSettings } from '@/api/settings';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { EmptyState } from '@/screens/member/MemberDashboard';
import { hrs } from '@/screens/_demo';

export function AdminMembers() {
  const { push, showToast } = useAppStore();
  const { user } = useAuthStore();
  const uid = user?.id ?? '';
  const [q, setQ] = useState('');

  const membersQ = useMembers();
  const eventsQ = useEvents();
  const { data: settings } = useSettings();
  const exportCsv = useExportMembersCSV();
  const importCsv = useImportMembersCSV();
  const importPreview = useImportMembersPreview();

  // CSV import dialog state
  const [importOpen, setImportOpen] = useState(false);
  const [preview, setPreview] = useState<ImportPreviewResult | null>(null);
  const [pendingFile, setPendingFile] = useState<File | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  // Manual create dialog state
  const [createOpen, setCreateOpen] = useState(false);
  const [newFirst, setNewFirst] = useState('');
  const [newLast, setNewLast] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newSince, setNewSince] = useState(() => new Date().toISOString().slice(0, 10));
  const createMember = useCreateMember();

  const members = membersQ.data ?? [];
  const yearGoal = settings?.yearGoal ?? 20;
  const memberMap: Record<string, { first: string; last: string }> = members.reduce(
    (acc, m) => { acc[m.id] = { first: m.first, last: m.last }; return acc; },
    {} as Record<string, { first: string; last: string }>,
  );

  // Per-member confirmed hours derived from event signups.
  // NOTE: manual bookings have no list endpoint, so they are not reflected here — the
  // per-member detail view fetches the authoritative total from /hours/:id.
  const hours: Record<string, number> = {};
  members.forEach((m) => { hours[m.id] = 0; });
  (eventsQ.data ?? []).forEach((ev) =>
    (ev.days ?? []).forEach((d) =>
      d.shifts.forEach((sh) =>
        sh.signups.forEach((s) => {
          if (s.status === 'bestätigt') hours[s.memberId] = (hours[s.memberId] ?? 0) + (s.hours ?? 0);
        }),
      ),
    ),
  );

  const list = members.filter((m) =>
    (m.first + ' ' + m.last).toLowerCase().includes(q.toLowerCase()),
  );

  const onExport = () => {
    showToast('Mitglieder-Export wird erstellt …');
    exportCsv.mutate();
  };

  const onPickFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    e.target.value = '';
    setPendingFile(file);
    setPreview(null);
    importPreview.mutate(file, {
      onSuccess: (result) => setPreview(result),
      onError: () => showToast('Vorschau konnte nicht geladen werden.', 'crit'),
    });
  };

  const onOpenImport = () => {
    setPreview(null);
    setPendingFile(null);
    setImportOpen(true);
  };

  const onConfirmImport = () => {
    if (!pendingFile) return;
    importCsv.mutate(pendingFile, {
      onSuccess: (result) => {
        showToast(`${result.imported ?? 0} Mitglieder importiert.`);
        setImportOpen(false);
        setPreview(null);
        setPendingFile(null);
      },
      onError: () => showToast('Import fehlgeschlagen.', 'crit'),
    });
  };

  const onCloseImport = () => {
    setImportOpen(false);
    setPreview(null);
    setPendingFile(null);
  };

  const onOpenCreate = () => {
    setNewFirst('');
    setNewLast('');
    setNewEmail('');
    setNewSince(new Date().toISOString().slice(0, 10));
    setCreateOpen(true);
  };

  const onConfirmCreate = () => {
    if (!newFirst.trim() || !newLast.trim() || !newEmail.trim()) return;
    createMember.mutate(
      { first: newFirst.trim(), last: newLast.trim(), email: newEmail.trim(), since: newSince },
      {
        onSuccess: () => {
          showToast(`${newFirst} ${newLast} wurde angelegt.`);
          setCreateOpen(false);
        },
        onError: () => showToast('Mitglied konnte nicht angelegt werden.', 'crit'),
      },
    );
  };

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{members.length} Einträge</div>
          <div className="sm-title">Mitglieder</div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            className="pressable"
            onClick={onOpenCreate}
            style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
            title="Mitglied manuell anlegen"
          >
            <Icon name="plus" size={20} color="var(--ink)" />
          </button>
          <button
            className="pressable"
            onClick={onOpenImport}
            style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
            title="CSV importieren"
          >
            <Icon name="list" size={20} color="var(--ink)" />
          </button>
          <button
            className="pressable"
            onClick={onExport}
            style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
            title="CSV exportieren"
          >
            <Icon name="download" size={20} color="var(--ink)" />
          </button>
        </div>
      </div>
      <div style={{ padding: '0 18px 4px' }}>
        <div style={{ position: 'relative' }}>
          <Icon name="search" size={18} color="var(--muted)" style={{ position: 'absolute', left: 13, top: 13 }} />
          <Input placeholder="Mitglied suchen…" value={q} onChange={(e) => setQ(e.target.value)} style={{ paddingLeft: 40 }} />
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 12 }}>
        {membersQ.isLoading && <LoadingState />}
        {membersQ.isError && <ErrorState />}
        {!membersQ.isLoading && !membersQ.isError && list.length === 0 && (
          <EmptyState icon="users" title="Keine Mitglieder" text={q ? 'Für diese Suche gibt es keine Treffer.' : 'Es sind noch keine Mitglieder angelegt.'} />
        )}
        {!membersQ.isLoading && !membersQ.isError && list.length > 0 && (
        <div className="sm-card" style={{ overflow: 'hidden' }}>
          {list.map((m, i) => {
            const goal = m.goal ?? yearGoal;
            const h = hours[m.id] ?? 0;
            return (
              <React.Fragment key={m.id}>
                {i > 0 && <hr className="sm-divider" style={{ marginLeft: 64 }} />}
                <div
                  className="pressable"
                  onClick={() => push('member', { id: m.id })}
                  style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', cursor: 'pointer' }}
                >
                  <Avatar memberId={m.id} members={memberMap} size={40} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, fontSize: 14.5 }}>
                      {m.first} {m.last}{m.id === uid && <span style={{ color: 'var(--muted)', fontWeight: 600 }}> · du</span>}
                    </div>
                    <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{hrs(h)} / {goal} h{m.goal && ' · ind. Ziel'}</div>
                  </div>
                  <div style={{ width: 46 }}>
                    <div className="occ-track">
                      <div className="occ-fill" style={{ width: Math.min(h / goal * 100, 100) + '%', background: h >= goal ? 'var(--ok)' : 'var(--primary)' }} />
                    </div>
                  </div>
                  <Icon name="chevR" size={16} color="var(--line-2)" stroke={2.4} />
                </div>
              </React.Fragment>
            );
          })}
        </div>
        )}
      </div>

      {/* Manual create dialog */}
      {createOpen && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center' }} onClick={() => setCreateOpen(false)}>
          <div style={{ background: 'var(--surface)', borderRadius: '20px 20px 0 0', width: '100%', maxWidth: 520, padding: 24, paddingBottom: 32 }} onClick={(e) => e.stopPropagation()}>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 16 }}>Mitglied anlegen</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 18 }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                <Field label="Vorname">
                  <Input placeholder="Max" value={newFirst} onChange={(e) => setNewFirst(e.target.value)} autoFocus />
                </Field>
                <Field label="Nachname">
                  <Input placeholder="Mustermann" value={newLast} onChange={(e) => setNewLast(e.target.value)} />
                </Field>
              </div>
              <Field label="E-Mail-Adresse">
                <Input type="email" placeholder="max@beispiel.de" value={newEmail} onChange={(e) => setNewEmail(e.target.value)} />
              </Field>
              <Field label="Eintrittsdatum">
                <Input type="date" value={newSince} onChange={(e) => setNewSince(e.target.value)} />
              </Field>
            </div>
            <button
              className="sm-btn"
              onClick={onConfirmCreate}
              disabled={createMember.isPending || !newFirst.trim() || !newLast.trim() || !newEmail.trim()}
              style={{ width: '100%', marginBottom: 10 }}
            >
              {createMember.isPending ? 'Wird angelegt…' : 'Mitglied anlegen'}
            </button>
            <button className="sm-btn ghost" onClick={() => setCreateOpen(false)} style={{ width: '100%' }}>
              Abbrechen
            </button>
          </div>
        </div>
      )}

      {/* Hidden file input */}
      <input ref={fileRef} type="file" accept=".csv,text/csv" onChange={onPickFile} style={{ display: 'none' }} />

      {/* CSV Import Dialog */}
      {importOpen && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center' }} onClick={onCloseImport}>
          <div style={{ background: 'var(--surface)', borderRadius: '20px 20px 0 0', width: '100%', maxWidth: 520, padding: 24, paddingBottom: 32 }} onClick={(e) => e.stopPropagation()}>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 6 }}>CSV importieren</div>

            {/* Step 1: file picker + preview */}
            {!preview && (
              <>
                <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600, marginBottom: 16 }}>
                  {importPreview.isPending
                    ? 'Vorschau wird geladen…'
                    : 'Wähle eine CSV-Datei mit den Spalten firstName, lastName, email.'}
                </div>
                <button
                  className="sm-btn soft"
                  onClick={() => fileRef.current?.click()}
                  disabled={importPreview.isPending}
                  style={{ width: '100%', marginBottom: 12 }}
                >
                  <Icon name="download" size={17} stroke={2.2} />
                  {pendingFile ? pendingFile.name : 'Datei auswählen'}
                </button>
                <button className="sm-btn ghost" onClick={onCloseImport} style={{ width: '100%' }}>
                  Abbrechen
                </button>
              </>
            )}

            {/* Step 2: preview summary + confirm */}
            {preview && (
              <>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 18 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 14px', background: 'var(--surface-2)', borderRadius: 12, fontWeight: 700, fontSize: 14 }}>
                    <span>Neu</span>
                    <span style={{ color: 'var(--ok)' }}>{preview.toCreate.length}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 14px', background: 'var(--surface-2)', borderRadius: 12, fontWeight: 700, fontSize: 14 }}>
                    <span>Aktualisiert</span>
                    <span style={{ color: 'var(--primary)' }}>{preview.toUpdate.length}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 14px', background: 'var(--surface-2)', borderRadius: 12, fontWeight: 700, fontSize: 14 }}>
                    <span>Unverändert</span>
                    <span style={{ color: 'var(--muted)' }}>{preview.unchanged}</span>
                  </div>
                </div>
                <button
                  className="sm-btn"
                  onClick={onConfirmImport}
                  disabled={importCsv.isPending}
                  style={{ width: '100%', marginBottom: 10 }}
                >
                  {importCsv.isPending ? 'Wird importiert…' : 'Importieren bestätigen'}
                </button>
                <button className="sm-btn ghost" onClick={onCloseImport} style={{ width: '100%' }}>
                  Abbrechen
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
