// Central URL path table for the authenticated app. Every screen and every
// non-confirmation popup gets its own path here so the browser back button
// works: opening a sheet/dialog is a navigation, not just local state.

export const TAB_PATHS = {
  start: '/start',
  entdecken: '/entdecken',
  schichten: '/schichten',
  profil: '/profil',
  uebersicht: '/uebersicht',
  events: '/events',
  mitglieder: '/mitglieder',
  abrechnungen: '/abrechnungen',
  settings: '/settings',
} as const;

export type TabKey = keyof typeof TAB_PATHS;

export const routes = {
  start: TAB_PATHS.start,
  entdecken: TAB_PATHS.entdecken,
  schichten: TAB_PATHS.schichten,
  profil: TAB_PATHS.profil,

  uebersicht: TAB_PATHS.uebersicht,

  events: TAB_PATHS.events,
  eventNeu: `${TAB_PATHS.events}/neu`,
  event: (id: string) => `${TAB_PATHS.events}/${id}`,
  eventBearbeiten: (id: string) => `${TAB_PATHS.events}/${id}/bearbeiten`,
  // The recurrence dialog overlays the events *list* (it's opened from a
  // card there), not the event-detail page, so it's a query param on
  // /events rather than a path segment under /events/:id.
  eventSerie: (id: string) => `${TAB_PATHS.events}?serie=${id}`,
  eventAnmelden: (id: string, shiftId: string) => `${TAB_PATHS.events}/${id}/anmelden/${shiftId}`,
  eventHelfer: (id: string, shiftId: string) => `${TAB_PATHS.events}/${id}/helfer/${shiftId}`,
  eventZeit: (id: string, signupId: string) => `${TAB_PATHS.events}/${id}/zeit/${signupId}`,

  mitglieder: TAB_PATHS.mitglieder,
  mitgliederNeu: `${TAB_PATHS.mitglieder}/neu`,
  mitgliederImport: `${TAB_PATHS.mitglieder}/import`,
  mitglied: (id: string) => `${TAB_PATHS.mitglieder}/${id}`,
  mitgliedBearbeiten: (id: string) => `${TAB_PATHS.mitglieder}/${id}/bearbeiten`,
  mitgliedZiel: (id: string) => `${TAB_PATHS.mitglieder}/${id}/ziel`,
  mitgliedAbgeltung: (id: string) => `${TAB_PATHS.mitglieder}/${id}/abgeltung`,

  stundenBuchen: (id?: string) => (id ? `${TAB_PATHS.mitglieder}/${id}/stunden-buchen` : '/stunden-buchen'),

  abrechnungen: TAB_PATHS.abrechnungen,
  abrechnungenNeu: `${TAB_PATHS.abrechnungen}/neu`,
  abrechnungBearbeiten: (id: string) => `${TAB_PATHS.abrechnungen}/${id}/bearbeiten`,
  abrechnungStaffeln: (id: string) => `${TAB_PATHS.abrechnungen}/${id}/staffeln`,

  settings: TAB_PATHS.settings,
  settingsBrandingVerlauf: `${TAB_PATHS.settings}/branding-verlauf`,
  settingsEmailVorlagen: `${TAB_PATHS.settings}/email-vorlagen`,
  settingsEmailVorlage: (name: string) => `${TAB_PATHS.settings}/email-vorlagen/${encodeURIComponent(name)}`,
  settingsEmailLog: `${TAB_PATHS.settings}/email-log`,
  settingsAudit: `${TAB_PATHS.settings}/audit`,
};

/** Home screen for a given role, mirroring the tabs each role actually sees. */
export function homePathForRole(role: string | undefined): string {
  return role && role !== 'mitglied' ? routes.uebersicht : routes.start;
}
