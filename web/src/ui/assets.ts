// Images come through web/assets/manifest.json (web/CLAUDE.md), resolved to
// the URLs Vite gives them (hashed files under /assets/, cached for good).
import manifest from '../../assets/manifest.json';

const urls = import.meta.glob('../../assets/**/*.webp', { eager: true, query: '?url', import: 'default' }) as Record<
  string,
  string
>;

export { manifest };

/** The URL of a manifest path such as "ui/khung-lang-9.webp" ('' if it is missing). */
export function assetUrl(path: string): string {
  return urls[`../../assets/${path}`] ?? '';
}

/** Hands the frame and background images to the CSS (styles.css custom properties). */
export function installAssetVars(root: HTMLElement): void {
  const vars: Record<string, string> = {
    '--img-khung-lang': manifest.ui['khung-lang-9'],
    '--img-khung-soi': manifest.ui['khung-soi-9'],
    '--img-khung-xam': manifest.ui['khung-xam-9'],
    '--img-nen-dem': manifest.nen.dem,
  };
  for (const [name, path] of Object.entries(vars)) {
    const url = assetUrl(path);
    if (url) root.style.setProperty(name, `url("${url}")`);
  }
}
