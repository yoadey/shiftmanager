import { describe, it, expect } from 'vitest';
import { isSectionStart, type NavTab } from './navTabs';

const tabs: NavTab[] = [
  { key: 'start', label: 'Start', icon: 'home' },
  { key: 'profil', label: 'Profil', icon: 'user' },
  { key: 'uebersicht', label: 'Übersicht', icon: 'chart', section: 'Verwaltung' },
  { key: 'events', label: 'Termine', icon: 'calendar' },
  { key: 'mitglieder', label: 'Mitglieder', icon: 'users' },
];

describe('isSectionStart', () => {
  it('is false for tabs without a section', () => {
    expect(isSectionStart(tabs, 0)).toBe(false);
    expect(isSectionStart(tabs, 1)).toBe(false);
  });

  it('is true for the first tab of a new section', () => {
    expect(isSectionStart(tabs, 2)).toBe(true);
  });

  it('is false for subsequent tabs of the same section, even without their own `section` field', () => {
    expect(isSectionStart(tabs, 3)).toBe(false);
    expect(isSectionStart(tabs, 4)).toBe(false);
  });

  it('is true again when a further tab restarts the same section label after a gap', () => {
    const withGap: NavTab[] = [
      { key: 'a', label: 'A', icon: 'a', section: 'X' },
      { key: 'b', label: 'B', icon: 'b' },
      { key: 'c', label: 'C', icon: 'c', section: 'X' },
    ];
    expect(isSectionStart(withGap, 2)).toBe(true);
  });
});
