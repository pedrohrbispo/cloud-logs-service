import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { bindSystem, getPref, setPref, type ThemePref } from '@/lib/theme';

export type LogsView = 'table' | 'cards';

interface UiState {
  theme: ThemePref;
  logsView: LogsView;
  provider: string;
  search: string;
  setTheme: (p: ThemePref) => void;
  setLogsView: (v: LogsView) => void;
  setProvider: (p: string) => void;
  setSearch: (q: string) => void;
  clearSearch: () => void;
}

export const useUiStore = create<UiState>()(
  persist(
    (set) => ({
      // theme is mirrored here for React, but `cla_theme` (lib/theme) is the
      // canonical store — see partialize below.
      theme: getPref(),
      logsView: 'table',
      provider: 'aws',
      search: '',
      setTheme: (p) => {
        setPref(p);
        bindSystem(p, () => set({ theme: p }));
        set({ theme: p });
      },
      setLogsView: (v) => set({ logsView: v }),
      setProvider: (p) => set({ provider: p, search: '' }),
      setSearch: (q) => set({ search: q }),
      clearSearch: () => set({ search: '' }),
    }),
    {
      name: 'cla_ui',
      // theme lives in cla_theme (no-flash script reads it); search is ephemeral.
      partialize: (s) => ({ logsView: s.logsView, provider: s.provider }),
    },
  ),
);
