import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  // Tokens swap on [data-theme]; the selector strategy is kept for the rare
  // explicit `dark:*` utility.
  darkMode: ['selector', '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        bg: 'var(--bg)',
        card: 'var(--card)',
        'card-2': 'var(--card-2)',
        border: 'var(--border)',
        text: 'var(--text)',
        muted: 'var(--muted)',
        faint: 'var(--faint)',
        sidebar: 'var(--sidebar)',
        primary: {
          DEFAULT: 'var(--primary)',
          hover: 'var(--primary-hover)',
          soft: 'var(--primary-soft)',
          strong: 'var(--primary-strong)',
          contrast: 'var(--primary-contrast)',
        },
        success: {
          DEFAULT: 'var(--success)',
          soft: 'var(--success-soft)',
          strong: 'var(--success-strong)',
        },
        error: {
          DEFAULT: 'var(--error)',
          soft: 'var(--error-soft)',
          strong: 'var(--error-strong)',
        },
        warn: {
          DEFAULT: 'var(--warn)',
          soft: 'var(--warn-soft)',
          strong: 'var(--warn-strong)',
        },
        skeleton: {
          DEFAULT: 'var(--skeleton)',
          hi: 'var(--skeleton-hi)',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'Segoe UI', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'ui-monospace', 'SFMono-Regular', 'monospace'],
      },
      fontSize: {
        title: ['21px', { lineHeight: '1.2', letterSpacing: '-0.02em', fontWeight: '700' }],
        stat: ['20px', { lineHeight: '1.15', fontWeight: '700' }],
        bucket: ['17px', { lineHeight: '1.2', fontWeight: '700' }],
        section: ['14px', { lineHeight: '1.3', fontWeight: '600' }],
        body: ['13.5px', { lineHeight: '1.5' }],
        cell: ['13.5px', { lineHeight: '1.4' }],
        name: ['13px', { lineHeight: '1.4' }],
        label: ['12px', { lineHeight: '1.3', fontWeight: '600' }],
        eyebrow: ['10.75px', { lineHeight: '1.2', letterSpacing: '0.065em', fontWeight: '600' }],
        meta: ['12px', { lineHeight: '1.4' }],
        badge: ['9.5px', { lineHeight: '1', letterSpacing: '0.05em', fontWeight: '700' }],
        hint: ['11.5px', { lineHeight: '1.5' }],
      },
      borderRadius: {
        sm: '6px',
        DEFAULT: '8px',
        md: '10px',
        lg: '12px',
        xl: '14px',
        pill: '999px',
      },
      boxShadow: {
        card: 'var(--shadow)',
        pop: 'var(--pop)',
      },
      maxWidth: {
        content: '1140px',
        login: '380px',
        modal: '440px',
      },
      spacing: {
        sidebar: '228px',
        topbar: '56px',
        page: '28px',
        'page-y': '26px',
      },
      screens: {
        xs: '480px',
        sm: '640px',
        cards: '720px',
        md: '768px',
        lg: '1024px',
        xl: '1280px',
      },
      animation: {
        shimmer: 'shimmer 1.2s linear infinite',
        'toast-in': 'toastIn .22s ease',
        'fade-in': 'fadeIn .25s ease',
        'modal-pop': 'pop .2s ease',
        spin: 'spin .7s linear infinite',
        'drawer-in': 'drawerIn .22s ease',
      },
    },
  },
  plugins: [],
} satisfies Config;
