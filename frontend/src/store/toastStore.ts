import { create } from 'zustand';

export type ToastStatus = 'progress' | 'done' | 'error' | null;

interface ToastState {
  status: ToastStatus;
  filename: string | null;
  show: (filename: string) => void;
  complete: (filename: string) => void;
  error: (filename: string) => void;
  dismiss: () => void;
}

// Dismiss timer kept at module scope so it never triggers re-renders.
let timer: ReturnType<typeof setTimeout> | null = null;
const clearTimer = () => {
  if (timer) {
    clearTimeout(timer);
    timer = null;
  }
};

export const useToastStore = create<ToastState>((set) => ({
  status: null,
  filename: null,
  show: (filename) => {
    clearTimer();
    set({ status: 'progress', filename });
  },
  complete: (filename) => {
    clearTimer();
    set({ status: 'done', filename });
    timer = setTimeout(() => set({ status: null, filename: null }), 3000);
  },
  error: (filename) => {
    clearTimer();
    set({ status: 'error', filename });
    timer = setTimeout(() => set({ status: null, filename: null }), 4000);
  },
  dismiss: () => {
    clearTimer();
    set({ status: null, filename: null });
  },
}));
