/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  // Dark mode is driven by `<html data-theme="dark">` (see §9.2 of the design doc).
  darkMode: ['selector', '[data-theme="dark"]'],
  theme: {
    extend: {
      // All semantic colors are CSS variables defined in src/index.css so the
      // user-selectable brand key color and light/dark themes resolve at runtime.
      colors: {
        brand: 'var(--brand)',
        'brand-strong': 'var(--brand-strong)',
        'brand-weak': 'var(--brand-weak)',
        'on-brand': 'var(--on-brand)',
        bg: 'var(--bg)',
        surface: 'var(--surface)',
        'surface-sub': 'var(--surface-sub)',
        content: 'var(--text)',
        'content-sub': 'var(--text-sub)',
        line: 'var(--line)',
        hover: 'var(--hover)',
        'prio-high': 'var(--prio-high)',
      },
      fontFamily: {
        sans: ['"IBM Plex Sans JP"', 'system-ui', 'sans-serif'],
        mono: ['"IBM Plex Mono"', 'ui-monospace', 'monospace'],
      },
      boxShadow: {
        compose: '0 8px 30px rgba(0,0,0,.16)',
        fab: '0 1px 3px rgba(60,64,67,.16)',
      },
    },
  },
  plugins: [],
};
