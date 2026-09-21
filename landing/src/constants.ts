export const REPO_URL = 'https://github.com/lipcsei/snitt';
export const RELEASES_URL = 'https://github.com/lipcsei/snitt/releases';

/**
 * A telepítők letöltési helye: a Snitt saját szerverén (az edge proxy
 * /downloads útvonala), NEM a GitHubon. A kiadó workflow (snitt repo, "publish-server")
 * ide tölti fel őket; a `latest.json` írja le a legújabb kiadást, a `latest/`
 * mappa állandó fájlnevekkel mutat rájuk. Build-időben felülírható
 * (VITE_DOWNLOAD_BASE_URL), pl. helyi próbához.
 */
export const DOWNLOAD_BASE_URL = (import.meta.env.VITE_DOWNLOAD_BASE_URL || 'https://api.snitt.video/downloads').replace(/\/+$/, '');
