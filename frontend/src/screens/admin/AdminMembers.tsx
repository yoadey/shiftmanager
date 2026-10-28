import React, { useState, useRef, useEffect } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Field, Input } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import {
  useMemberPage, usePendingMembers,
  useCreateMember, useExportMembersCSV, useImportMembersCSV, useImportMembersPreview,
} from '@/api/members';
import type { ImportPreviewResult } from '@/api/members';
import { useEvents } from '@/api/events';
import { useSettings } from '@/api/settings';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { EmptyState } from '@/screens/member/MemberDashboard';
import { fmtDate, hrs } from '@/screens/_demo';
import { PendingApprovalActions } from './PendingApprovalActions';
import { routes } from '@/routes';
import { useSmartBack } from '@/hooks/useSmartBack';

export const MEMBERS_PAGE_SIZE = 25;

export function AdminMembers() {
  const { showToast } = useAppStore();
  const { user } = useAuthStore();
  const uid = user?.id ?? '';
  const navigate = useNavigate();
  const location = useLocation();
  const closeModal = useSmartBack(routes.mitglieder);

  // Page and search live in the URL (?seite=2&suche=…) so returning from a
  // member's detail page via the back button lands on the same page.
  const [searchParams, setSearchParams] = useSearchParams();
  const search = searchParams.get('suche') ?? '';
  const page = Math.max(0, Math.floor(Number(searchParams.get('seite')) || 1) - 1);
  const [q, setQ] = useState(search);
  // Only explicit page navigation adds a history entry; search updates and
  // the out-of-range snap replace the current one (otherwise the back button
  // would replay every search, or bounce back onto the page that snaps again).
  const setListParams = (next: { page?: number; search?: string }, replace: boolean) => {
    const p = new URLSearchParams(searchParams);
    const nextSearch = next.search ?? search;
    const nextPage = next.page ?? page;
    if (nextSearch) p.set('suche', nextSearch); else p.delete('suche');
    if (nextPage > 0) p.set('seite', String(nextPage + 1)); else p.delete('seite');
    setSearchParams(p, { replace });
  };
  // Keep the input in sync when the URL search changes from outside (back/forward).
  useEffect(() => {
    setQ((cur) => (cur.trim() === search ? cur : search));
  }, [search]);
  // Debounce typing into the server-side search; a new search starts at page 1.
  useEffect(() => {
    if (q.trim() === search) return;
    const t = setTimeout(() => setListParams({ search: q.trim(), page: 0 }, true), 300);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [q]);

  // Pending registrations are listed separately on top; the paginated list
  // below shows everyone else (active and deactivated members).
  const membersQ = useMemberPage({ page, pageSize: MEMBERS_PAGE_SIZE, search, status: 'approved' });
  const pendingQ = usePendingMembers();
  const eventsQ = useEvents();
  const { data: settings } = useSettings();
  const exportCsv = useExportMembersCSV();
  const importCsv = useImportMembersCSV();
  const importPreview = useImportMembersPreview();

  // CSV import dialog state — routed at /mitglieder/import. `preview`/
  // `pendingFile` hold a File object, so they can't survive a reload; a
  // direct link there just lands back on step 1 (the file picker).
  const importOpen = location.pathname === routes.mitgliederImport;
  const [preview, setPreview] = useState<ImportPreviewResult | null>(null);
  const [pendingFile, setPendingFile] = useState<File | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

  // Manual create dialog state — routed at /mitglieder/neu.
  const createOpen = location.pathname === routes.mitgliederNeu;
  const [newFirst, setNewFirst] = useState('');
  const [newLast, setNewLast] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newSince, setNewSince] = useState(() => new Date().toISOString().slice(0, 10));
  const createMember = useCreateMember();

  const members = membersQ.data?.items ?? [];
  const total = membersQ.data?.total ?? 0;
  const pending = pendingQ.data ?? [];
  const pageCount = Math.max(1, Math.ceil(total / MEMBERS_PAGE_SIZE));
  const yearGoal = settings?.yearGoal ?? 20;
  const memberMap: Record<string, { first: string; last: string }> = [...members, ...pending].reduce(
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

  // A page beyond the end (e.g. after the last member on it was removed, or a
  // stale bookmark) snaps back to the last existing page.
  useEffect(() => {
    // Placeholder data still carries the previous query's total (e.g. after
    // going back from a search), so only snap once the real page has loaded.
    if (membersQ.data && !membersQ.isPlaceholderData && page > 0 && page >= pageCount) setListParams({ page: pageCount - 1 }, true);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [membersQ.data, membersQ.isPlaceholderData, page, pageCount]);

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
    navigate({ pathname: routes.mitgliederImport, search: location.search });
  };

  const onConfirmImport = () => {
    if (!pendingFile) return;
    importCsv.mutate(pendingFile, {
      onSuccess: (result) => {
        showToast(`${result.imported ?? 0} Mitglieder importiert.`);
        setPreview(null);
        setPendingFile(null);
        closeModal();
      },
      onError: () => showToast('Import fehlgeschlagen.', 'crit'),
    });
  };

  const onCloseImport = () => {
    setPreview(null);
    setPendingFile(null);
    closeModal();
  };

  const onOpenCreate = () => {
    setNewFirst('');
    setNewLast('');
    setNewEmail('');
    setNewSince(new Date().toISOString().slice(0, 10));
    navigate({ pathname: routes.mitgliederNeu, search: location.search });
  };

  const onConfirmCreate = () => {
    if (!newFirst.trim() || !newLast.trim() || !newEmail.trim()) return;
    createMember.mutate(
      { first: newFirst.trim(), last: newLast.trim(), email: newEmail.trim(), since: newSince },
      {
        onSuccess: () => {
          showToast(`${newFirst} ${newLast} wurde angelegt.`);
          closeModal();
        },
        onError: () => showToast('Mitglied konnte nicht angelegt werden.', 'crit'),
      },
    );
  };

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{search ? total : total + pending.length} Einträge</div>
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
      {pendingQ.isError && (
        <div className="sm-pad" style={{ paddingTop: 12, paddingBottom: 0 }}>
          <div className="sm-card pad" role="alert" style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'var(--warn-bg)', fontSize: 13, fontWeight: 600 }}>
            <Icon name="clock" size={18} color="var(--warn)" />
            Wartende Registrierungen konnten nicht geladen werden.
          </div>
        </div>
      )}
      {pending.length > 0 && (
        <div className="sm-pad" style={{ paddingTop: 12, paddingBottom: 0 }} data-testid="pending-members">
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, margin: '0 2px 8px' }}>
            <Icon name="clock" size={16} color="var(--warn)" />
            <span style={{ fontWeight: 800, fontSize: 14.5 }}>Warten auf Freischaltung</span>
            <Badge kind="warn" dot={false}>{pending.length}</Badge>
          </div>
          <div className="sm-card" style={{ overflow: 'hidden', background: 'var(--warn-bg)' }}>
            {pending.map((m, i) => {
              const name = `${m.first} ${m.last}`.trim() || m.email;
              return (
                <React.Fragment key={m.id}>
                  {i > 0 && <hr className="sm-divider" style={{ marginLeft: 64 }} />}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', flexWrap: 'wrap' }}>
                    <div
                      className="pressable"
                      onClick={() => navigate(routes.mitglied(m.id))}
                      style={{ display: 'flex', alignItems: 'center', gap: 12, flex: '1 1 180px', minWidth: 0, cursor: 'pointer' }}
                    >
                      <Avatar memberId={m.id} members={memberMap} size={40} />
                      <div style={{ minWidth: 0 }}>
                        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{name}</div>
                        <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {m.email} · registriert am {fmtDate(m.since, 'short')}
                        </div>
                      </div>
                    </div>
                    <PendingApprovalActions id={m.id} name={name} />
                  </div>
                </React.Fragment>
              );
            })}
          </div>
        </div>
      )}
      <div className="sm-pad" style={{ paddingTop: 12 }}>
        {membersQ.isLoading && <LoadingState />}
        {membersQ.isError && <ErrorState />}
        {!membersQ.isLoading && !membersQ.isError && members.length === 0 && (
          <EmptyState icon="users" title="Keine Mitglieder" text={search ? 'Für diese Suche gibt es keine Treffer.' : pending.length > 0 ? 'Es ist noch kein Mitglied freigeschaltet.' : 'Es sind noch keine Mitglieder angelegt.'} />
        )}
        {!membersQ.isLoading && !membersQ.isError && members.length > 0 && (
        <div className="sm-card" style={{ overflow: 'hidden' }}>
          {members.map((m, i) => {
            const goal = m.goal ?? yearGoal;
            const h = hours[m.id] ?? 0;
            return (
              <React.Fragment key={m.id}>
                {i > 0 && <hr className="sm-divider" style={{ marginLeft: 64 }} />}
                <div
                  className="pressable"
                  onClick={() => navigate(routes.mitglied(m.id))}
                  style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', cursor: 'pointer' }}
                >
                  <Avatar memberId={m.id} members={memberMap} size={40} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, fontSize: 14.5, display: 'flex', alignItems: 'center', gap: 8 }}>
                      <span>{m.first} {m.last}{m.id === uid && <span style={{ color: 'var(--muted)', fontWeight: 600 }}> · du</span>}</span>
                      {m.active === false && <Badge kind="neutral" dot={false}>Ausgetreten</Badge>}
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
        {!membersQ.isError && pageCount > 1 && (
          <nav aria-label="Seiten" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10, marginTop: 12 }}>
            <Button size="sm" variant="ghost" disabled={page === 0} onClick={() => setListParams({ page: page - 1 }, false)}>
              <Icon name="chevL" size={14} stroke={2.4} /> Zurück
            </Button>
            <span style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 700, textAlign: 'center' }}>
              <span style={{ display: 'block' }}>Seite {page + 1} von {pageCount}</span>
              <span style={{ display: 'block', fontWeight: 600, fontSize: 11.5 }}>
                {page * MEMBERS_PAGE_SIZE + 1}–{Math.min((page + 1) * MEMBERS_PAGE_SIZE, total)} von {total}
              </span>
            </span>
            <Button size="sm" variant="ghost" disabled={page >= pageCount - 1} onClick={() => setListParams({ page: page + 1 }, false)}>
              Weiter <Icon name="chevR" size={14} stroke={2.4} />
            </Button>
          </nav>
        )}
      </div>

      {/* Manual create dialog */}
      {createOpen && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center' }} onClick={closeModal}>
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
            <button className="sm-btn ghost" onClick={closeModal} style={{ width: '100%' }}>
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
