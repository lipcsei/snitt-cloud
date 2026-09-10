import { useEffect, useRef, useState } from 'react';
import type { KeyboardEvent } from 'react';
import { Link } from 'react-router-dom';
import { useI18n } from '../i18n';
import Rich from '../i18n/Rich';
import type { InstallGuide, Strings } from '../i18n/types';
import '../styles.install.css';

type OsId = 'windows' | 'macos' | 'linux';

const OS_LIST: { id: OsId; label: string }[] = [
  { id: 'windows', label: 'Windows' },
  { id: 'macos', label: 'macOS' },
  { id: 'linux', label: 'Linux' },
];

/**
 * A parancsok nyelvfüggetlenek, ezért nem a szövegtáblákban vannak: minden
 * lépéshez az ott megjelenő parancsblokkok tartoznak, sorrendben. A figyelmeztető
 * doboz helye rendszerenként más (Linuxon az első lépés után jön).
 */
const OS_COMMANDS: Record<OsId, { steps: string[][]; warn?: string; warnAfter?: number }> = {
  windows: {
    steps: [
      ['winget install Gyan.FFmpeg\nwinget install yt-dlp.yt-dlp'],
      ['winget install Python.Python.3.12', 'pip install faster-whisper-cli'],
      ['choco install ffmpeg yt-dlp python'],
    ],
  },
  macos: {
    steps: [
      ['brew install ffmpeg yt-dlp'],
      ['brew install pipx && pipx ensurepath\npipx install faster-whisper-cli'],
    ],
    warn: 'export PATH="$HOME/Library/Python/3.9/bin:$PATH"',
  },
  linux: {
    steps: [
      ['sudo apt install ffmpeg pipx\npipx install faster-whisper-cli'],
      ['sudo dnf install ffmpeg yt-dlp pipx'],
      ['sudo pacman -S ffmpeg yt-dlp python-pipx'],
    ],
    warn:
      'sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp\nsudo chmod a+rx /usr/local/bin/yt-dlp',
    warnAfter: 1,
  },
};

const CHECK_COMMANDS = 'ffmpeg -version\nyt-dlp --version\nfaster-whisper --help';

// Csak a kezdő fület találjuk ki belőle; mindhárom rendszer kézzel is elérhető marad,
// tehát a téves tipp legrosszabb esetben egy kattintás.
function detectOs(): OsId {
  if (typeof navigator === 'undefined') return 'windows';
  const ua = navigator.userAgent;
  if (/Mac|iPhone|iPad|iPod/i.test(ua)) return 'macos';
  // Az Android user agentje is tartalmazza a "Linux" szót, ezért ez a macOS-teszt után jön.
  if (/Linux|Android|X11|CrOS/i.test(ua)) return 'linux';
  return 'windows';
}

type CopyState = 'idle' | 'ok' | 'err';

function CodeBlock({ code, label }: { code: string; label?: string }) {
  const t = useI18n().t;
  const [state, setState] = useState<CopyState>('idle');
  const timer = useRef<number | undefined>(undefined);

  useEffect(() => () => window.clearTimeout(timer.current), []);

  async function copy() {
    try {
      await navigator.clipboard.writeText(code);
      setState('ok');
    } catch {
      // Nem HTTPS alatt vagy letiltott vágólap esetén a writeText eldobja magát.
      setState('err');
    }
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setState('idle'), 2000);
  }

  return (
    <div className="code-block">
      <div className="code-head">
        <span className="code-label">{label ?? t.install.codeLabel}</span>
        <button type="button" className="code-copy" onClick={copy}>
          {state === 'ok' ? t.install.copied : state === 'err' ? t.install.copyFailed : t.install.copy}
        </button>
      </div>
      <pre>
        <code>{code}</code>
      </pre>
    </div>
  );
}

function Warning({ warn, code }: { warn: InstallGuide['warn']; code?: string }) {
  return (
    <aside className="install-warn">
      <h3>{warn.title}</h3>
      <p>
        <Rich text={warn.body} />
      </p>
      {code && <CodeBlock code={code} label={warn.codeLabel} />}
      {warn.after && (
        <p>
          <Rich text={warn.after} />
        </p>
      )}
    </aside>
  );
}

function Guide({ os, strings }: { os: OsId; strings: Strings }) {
  const guide = strings.install[os];
  const { steps, warn, warnAfter } = OS_COMMANDS[os];

  return (
    <>
      {guide.steps.map((step, i) => (
        <div key={step.title}>
          <section className="install-step">
            <h3>{step.title}</h3>
            {step.body && (
              <p>
                <Rich text={step.body} />
              </p>
            )}
            {steps[i]?.[0] && <CodeBlock code={steps[i][0]} label={step.codeLabel} />}
            {step.hint && (
              <p className="install-hint">
                <Rich text={step.hint} />
              </p>
            )}
            {steps[i]?.[1] && <CodeBlock code={steps[i][1]} label={step.codeLabel2} />}
          </section>
          {warnAfter === i + 1 && <Warning warn={guide.warn} code={warn} />}
        </div>
      ))}
      {warnAfter === undefined && <Warning warn={guide.warn} code={warn} />}
    </>
  );
}

export default function Install() {
  const { t, path } = useI18n();
  const [os, setOs] = useState<OsId>(detectOs);
  const tabRefs = useRef<Partial<Record<OsId, HTMLButtonElement | null>>>({});

  // A főoldalról érkezve a router megtartja a görgetést, ez az oldal viszont mindig
  // elölről olvasandó.
  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  function onTabKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const current = OS_LIST.findIndex((item) => item.id === os);
    let next = -1;
    if (event.key === 'ArrowRight') next = (current + 1) % OS_LIST.length;
    else if (event.key === 'ArrowLeft') next = (current - 1 + OS_LIST.length) % OS_LIST.length;
    else if (event.key === 'Home') next = 0;
    else if (event.key === 'End') next = OS_LIST.length - 1;
    if (next < 0) return;

    event.preventDefault();
    const id = OS_LIST[next].id;
    setOs(id);
    tabRefs.current[id]?.focus();
  }

  return (
    <main className="page install-page">
      <div className="container container-narrow">
        <span className="eyebrow">{t.install.eyebrow}</span>
        <h1>{t.install.title}</h1>
        <p className="install-lead">{t.install.lead}</p>

        <aside className="install-callout">
          <h2>{t.install.callout.title}</h2>
          <p>
            <Rich text={t.install.callout.body} />
          </p>
        </aside>

        <ul className="dep-list">
          {t.install.deps.map((dep, i) => (
            <li key={dep.name} className="dep">
              <div className="dep-head">
                <code>{dep.name}</code>
                <span className={i === 0 ? 'dep-tag dep-tag-req' : 'dep-tag'}>
                  {i === 0 ? t.install.tagRequired : t.install.tagOptional}
                </span>
              </div>
              <p>
                <Rich text={dep.body} />
              </p>
            </li>
          ))}
        </ul>

        <div
          className="os-tabs"
          role="tablist"
          aria-label={t.install.tabsAria}
          onKeyDown={onTabKeyDown}
        >
          {OS_LIST.map((item) => (
            <button
              key={item.id}
              type="button"
              role="tab"
              id={`os-tab-${item.id}`}
              aria-selected={os === item.id}
              aria-controls={`os-panel-${item.id}`}
              tabIndex={os === item.id ? 0 : -1}
              ref={(el) => {
                tabRefs.current[item.id] = el;
              }}
              className={os === item.id ? 'os-tab os-tab-on' : 'os-tab'}
              onClick={() => setOs(item.id)}
            >
              {item.label}
            </button>
          ))}
        </div>

        <div
          className="os-panel"
          role="tabpanel"
          id={`os-panel-${os}`}
          aria-labelledby={`os-tab-${os}`}
          tabIndex={0}
        >
          <Guide os={os} strings={t} />
        </div>

        <section className="install-step install-check">
          <h2>{t.install.check.title}</h2>
          <p>{t.install.check.body}</p>
          <CodeBlock code={CHECK_COMMANDS} />
        </section>

        <section className="install-step">
          <h2>{t.install.path.title}</h2>
          <p>{t.install.path.body}</p>
          <ul className="env-list">
            {t.install.path.vars.map((item) => (
              <li key={item.name}>
                <code>{item.name}</code>
                <span>{item.body}</span>
              </li>
            ))}
          </ul>
        </section>

        <section className="install-step">
          <h2>{t.install.good.title}</h2>
          <ul className="know-list">
            {t.install.good.items.map((item) => (
              <li key={item}>
                <Rich text={item} />
              </li>
            ))}
          </ul>
        </section>

        <Link to={`${path('home')}#${t.sections.downloads}`} className="back-link">
          {t.common.backToDownloads}
        </Link>
      </div>
    </main>
  );
}
