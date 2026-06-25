// SPDX-License-Identifier: Apache-2.0
//
// Theme + brand-color application (design doc §9.2).
//
// The pre-paint inline script in index.html sets the initial `data-theme` and
// `--brand` to avoid a flash of the wrong palette. These helpers keep the DOM
// in sync afterwards when the user changes appearance settings.

export type Theme = 'system' | 'light' | 'dark';

const DARK_QUERY = '(prefers-color-scheme: dark)';

/** Resolve a Theme preference to the concrete light/dark mode to render. */
export function resolveTheme(theme: Theme): 'light' | 'dark' {
  if (theme === 'system') {
    return window.matchMedia(DARK_QUERY).matches ? 'dark' : 'light';
  }
  return theme;
}

/**
 * Apply the resolved theme to <html> and update the address-bar `theme-color`
 * meta so the PWA chrome matches (design doc §10.1).
 */
export function applyTheme(theme: Theme): void {
  const mode = resolveTheme(theme);
  document.documentElement.dataset.theme = mode;
  updateThemeColorMeta(mode);
}

/** Apply a brand key color (HEX) and its computed `--on-brand` foreground. */
export function applyBrand(hex: string): void {
  const root = document.documentElement;
  root.style.setProperty('--brand', hex);
  root.style.setProperty('--on-brand', onBrandColor(hex));
}

/**
 * Choose a readable foreground (near-black or white) for text placed on top of
 * the given brand color, using WCAG relative luminance (design doc §9.2.4).
 * Bright colors such as the default yellow get dark text; saturated mid/dark
 * colors get white.
 */
export function onBrandColor(hex: string): '#1f2329' | '#ffffff' {
  const channels = [0, 2, 4]
    .map((i) => parseInt(hex.slice(1 + i, 3 + i), 16) / 255)
    .map((v) => (v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4));
  const [r, g, b] = channels;
  const luminance = 0.2126 * r + 0.7152 * g + 0.0722 * b;
  return luminance > 0.45 ? '#1f2329' : '#ffffff';
}

/** Keep the `<meta name="theme-color">` tag in sync with the active surface. */
function updateThemeColorMeta(mode: 'light' | 'dark'): void {
  let meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]');
  if (!meta) {
    meta = document.createElement('meta');
    meta.name = 'theme-color';
    document.head.appendChild(meta);
  }
  const styles = getComputedStyle(document.documentElement);
  // Dark mode follows the page background; light mode keeps the brand color.
  meta.content =
    mode === 'dark'
      ? styles.getPropertyValue('--bg').trim() || '#16181c'
      : styles.getPropertyValue('--brand').trim() || '#ffd20a';
}

/**
 * Subscribe to OS color-scheme changes. Only meaningful while the preference is
 * 'system'. Returns an unsubscribe function.
 */
export function watchSystemTheme(onChange: () => void): () => void {
  const mql = window.matchMedia(DARK_QUERY);
  const handler = () => onChange();
  mql.addEventListener('change', handler);
  return () => mql.removeEventListener('change', handler);
}
