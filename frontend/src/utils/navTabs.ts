export interface NavTab {
  key: string;
  label: string;
  icon: string;
  /** When set and different from the previous tab's, marks the start of a
   * new group (e.g. the veranstaltungsleiter/vorstand/admin-only items
   * appended after the member tabs) — see `isSectionStart`. */
  section?: string;
}

/** True when tab `i` starts a new section, i.e. it declares a `section` that
 * differs from the previous tab's. Shared by DesktopShell and MobileShell so
 * both render dividers/labels at the same boundaries. */
export function isSectionStart(tabs: NavTab[], i: number): boolean {
  return !!tabs[i].section && tabs[i - 1]?.section !== tabs[i].section;
}
