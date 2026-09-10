import { useEffect, useRef, useState } from 'react';
import type { KeyboardEvent } from 'react';
import { Link } from 'react-router-dom';
import '../styles.install.css';

type OsId = 'windows' | 'macos' | 'linux';

const OS_LIST: { id: OsId; label: string }[] = [
  { id: 'windows', label: 'Windows' },
  { id: 'macos', label: 'macOS' },
  { id: 'linux', label: 'Linux' },
];

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

function CodeBlock({ code, label = 'Terminál' }: { code: string; label?: string }) {
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
        <span className="code-label">{label}</span>
        <button type="button" className="code-copy" onClick={copy}>
          {state === 'ok' ? 'Másolva' : state === 'err' ? 'Nem sikerült' : 'Másolás'}
        </button>
      </div>
      <pre>
        <code>{code}</code>
      </pre>
    </div>
  );
}

function WindowsGuide() {
  return (
    <>
      <section className="install-step">
        <h3>1. Telepítés wingettel</h3>
        <p>
          A <code>winget</code> a Windows 10 és 11 része, külön telepíteni nem kell. Nyiss egy
          PowerShell ablakot, és futtasd a szükséges sorokat.
        </p>
        <CodeBlock
          label="PowerShell"
          code={'winget install Gyan.FFmpeg\nwinget install yt-dlp.yt-dlp'}
        />
        <p className="install-hint">
          Az első sor kell mindenhez. A második csak akkor, ha linkről is szeretnél videót behozni.
        </p>
      </section>

      <section className="install-step">
        <h3>2. Beszédfelismerés (opcionális)</h3>
        <p>
          Ez a rész csak akkor kell, ha felirat nélküli videókhoz is szeretnél átiratot. Először
          Python kell hozzá, utána maga a csomag.
        </p>
        <CodeBlock label="PowerShell" code={'winget install Python.Python.3.12'} />
        <p className="install-hint">
          A Python telepítése után <strong>nyiss egy új PowerShell ablakot</strong>, különben a{' '}
          <code>pip</code> parancsot még nem találja meg a rendszer.
        </p>
        <CodeBlock label="Új PowerShell ablak" code={'pip install faster-whisper-cli'} />
      </section>

      <section className="install-step">
        <h3>Alternatíva: Chocolatey</h3>
        <p>
          Ha Chocolateyt használsz, egy sorral is megvan az <code>ffmpeg</code>, a{' '}
          <code>yt-dlp</code> és a Python:
        </p>
        <CodeBlock label="PowerShell (rendszergazda)" code={'choco install ffmpeg yt-dlp python'} />
      </section>

      <aside className="install-warn">
        <h3>Telepítés után nyiss új terminált</h3>
        <p>
          A telepítők a PATH-ot módosítják, a már futó programok viszont a régi PATH-ot látják. Ha a
          telepítés után az <code>ffmpeg -version</code> még mindig azt írja, hogy nem található,
          nyiss egy új terminálablakot — ha pedig a Snitt közben nyitva volt, indítsd újra. Ez a
          leggyakoribb elakadás Windowson, és nem a telepítéssel van baj.
        </p>
      </aside>
    </>
  );
}

function MacGuide() {
  return (
    <>
      <section className="install-step">
        <h3>1. Homebrew</h3>
        <p>
          macOS-en a legrövidebb út a{' '}
          <a href="https://brew.sh" target="_blank" rel="noreferrer noopener">
            Homebrew
          </a>
          . Ha még nincs fent, a telepítőparancsot a brew.sh kezdőlapján találod. Ha már megvan,
          ugorj a következő sorra.
        </p>
        <CodeBlock code={'brew install ffmpeg yt-dlp'} />
        <p className="install-hint">
          Az <code>ffmpeg</code> kell mindenhez, a <code>yt-dlp</code> csak a linkről importáláshoz.
        </p>
      </section>

      <section className="install-step">
        <h3>2. Beszédfelismerés (opcionális)</h3>
        <p>
          A legtisztább megoldás a <code>pipx</code>: külön környezetbe teszi a Python-eszközöket, és
          gondoskodik róla, hogy a parancs a PATH-ra kerüljön.
        </p>
        <CodeBlock
          code={'brew install pipx && pipx ensurepath\npipx install faster-whisper-cli'}
        />
        <p className="install-hint">
          A <code>pipx ensurepath</code> a shell profilodat írja át, tehát utána nyiss egy új
          terminált.
        </p>
      </section>

      <aside className="install-warn">
        <h3>Ha pipx nélkül telepíted</h3>
        <p>
          A <code>pip3 install --user faster-whisper-cli</code> is működik, de a parancsot a{' '}
          <code>~/Library/Python/3.x/bin</code> könyvtárba teszi, ami alapból{' '}
          <strong>nincs rajta a PATH-on</strong>. Ilyenkor a telepítés sikerül, a Snitt viszont nem
          fogja megtalálni a <code>faster-whisper</code> parancsot. Vedd fel a könyvtárat a shell
          profilodba (<code>~/.zshrc</code>):
        </p>
        <CodeBlock label="~/.zshrc" code={'export PATH="$HOME/Library/Python/3.9/bin:$PATH"'} />
        <p>
          A verziószámnak egyeznie kell a saját Pythonodéval — nézd meg a{' '}
          <code>python3 --version</code> kimenetét, és azt írd be a <code>3.9</code> helyére,
          különben a sor nem csinál semmit.
        </p>
      </aside>
    </>
  );
}

function LinuxGuide() {
  return (
    <>
      <section className="install-step">
        <h3>Debian / Ubuntu</h3>
        <CodeBlock code={'sudo apt install ffmpeg pipx\npipx install faster-whisper-cli'} />
        <p className="install-hint">
          A második sor csak akkor kell, ha beszédfelismerést is szeretnél. A{' '}
          <code>pipx ensurepath</code> után nyiss új terminált.
        </p>
      </section>

      <aside className="install-warn">
        <h3>yt-dlp: kerüld a disztribúciós csomagot</h3>
        <p>
          A tárolókban lévő <code>yt-dlp</code> jellemzően elavult, a videómegosztók pedig gyakran
          változnak — egy régi verzió hetek alatt használhatatlanná válik. Töltsd le inkább a
          hivatalos binárist, ez frissíti magát:
        </p>
        <CodeBlock
          code={
            'sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp\nsudo chmod a+rx /usr/local/bin/yt-dlp'
          }
        />
      </aside>

      <section className="install-step">
        <h3>Fedora</h3>
        <CodeBlock code={'sudo dnf install ffmpeg yt-dlp pipx'} />
        <p className="install-hint">
          Az <code>ffmpeg</code> nincs benne a Fedora alap tárolóiban: ehhez előbb engedélyezned kell
          az RPM Fusion tárolót, különben a parancs nem találja a csomagot.
        </p>
      </section>

      <section className="install-step">
        <h3>Arch</h3>
        <CodeBlock code={'sudo pacman -S ffmpeg yt-dlp python-pipx'} />
        <p className="install-hint">
          A Whisper ezután <code>pipx install faster-whisper-cli</code> paranccsal jön.
        </p>
      </section>
    </>
  );
}

const GUIDES: Record<OsId, () => JSX.Element> = {
  windows: WindowsGuide,
  macos: MacGuide,
  linux: LinuxGuide,
};

export default function Install() {
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

  const Guide = GUIDES[os];

  return (
    <main className="page install-page">
      <div className="container container-narrow">
        <span className="eyebrow">Telepítés</span>
        <h1>Külső eszközök telepítése</h1>
        <p className="install-lead">
          A Snitt három parancssori eszközre támaszkodik, és ezeket szándékosan nem csomagolja
          magába: a gépeden lévő, saját verziójú programokat használja. Ez az oldal végigvezet a
          telepítésükön, rendszerenként.
        </p>

        <aside className="install-callout">
          <h2>Mennyi kell ebből tényleg?</h2>
          <p>
            Ha csak a videóid mellett lévő feliratfájlokban és a videókba ágyazott feliratsávokban
            keresel, <strong>elég az ffmpeg</strong>. A másik kettő nem előfeltétel: akkor kell
            telepítened őket, amikor először használnád azt a funkciót.
          </p>
        </aside>

        <ul className="dep-list">
          <li className="dep">
            <div className="dep-head">
              <code>ffmpeg</code>
              <span className="dep-tag dep-tag-req">Kötelező</span>
            </div>
            <p>
              Klipek vágása, hangsáv kinyerése, a videóba ágyazott feliratsávok kiolvasása.
              Gyakorlatilag minden művelethez kell.
            </p>
          </li>
          <li className="dep">
            <div className="dep-head">
              <code>yt-dlp</code>
              <span className="dep-tag">Opcionális</span>
            </div>
            <p>Csak akkor, ha linkről szeretnél videót behozni a tárba.</p>
          </li>
          <li className="dep">
            <div className="dep-head">
              <code>faster-whisper</code>
              <span className="dep-tag">Opcionális</span>
            </div>
            <p>
              Csak akkor, ha beszédfelismeréssel is szeretnél átiratot készíteni. A parancsot a{' '}
              <code>faster-whisper-cli</code> Python-csomag telepíti.
            </p>
          </li>
        </ul>

        <div className="os-tabs" role="tablist" aria-label="Operációs rendszer" onKeyDown={onTabKeyDown}>
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
          <Guide />
        </div>

        <section className="install-step install-check">
          <h2>Ellenőrzés</h2>
          <p>
            Nyiss egy terminált, és futtasd le azt a sort, amelyik eszközt telepítetted. Ha ezek
            kiírnak valamit, a Snitt is meg fogja találni őket — ugyanazon a PATH-on keresi.
          </p>
          <CodeBlock code={'ffmpeg -version\nyt-dlp --version\nfaster-whisper --help'} />
        </section>

        <section className="install-step">
          <h2>Ha valami mégsincs a PATH-on</h2>
          <p>
            Előfordul, hogy egy eszköz olyan helyre kerül, ahonnan a rendszer nem látja — vagy
            szándékosan máshol tartod. Ilyenkor nem kell a PATH-tal küzdeni: add meg a teljes
            elérési utat a megfelelő környezeti változóban, és a Snitt azt fogja használni.
          </p>
          <ul className="env-list">
            <li>
              <code>FFMPEG_BIN</code>
              <span>az ffmpeg futtatható fájlja</span>
            </li>
            <li>
              <code>FFPROBE_BIN</code>
              <span>az ffprobe futtatható fájlja (az ffmpeg mellett érkezik)</span>
            </li>
            <li>
              <code>YTDLP_BIN</code>
              <span>a yt-dlp futtatható fájlja</span>
            </li>
            <li>
              <code>WHISPER_BIN</code>
              <span>a faster-whisper parancs</span>
            </li>
          </ul>
        </section>

        <section className="install-step">
          <h2>Jó tudni</h2>
          <ul className="know-list">
            <li>
              <strong>A Whisper modell az első használatkor töltődik le.</strong> Mérettől függően
              néhány száz megabájttól nagyjából 3 gigabájtig terjed, tehát az első átirat előtt
              érdemes rendes netre és szabad helyre számítani. Utána már helyben van.
            </li>
            <li>
              <strong>A modellt a <code>WHISPER_MODEL</code> környezeti változó választja ki.</strong>{' '}
              Lehetséges értékek: <code>tiny</code>, <code>base</code>, <code>small</code>,{' '}
              <code>medium</code>, <code>large-v3</code>. A <code>small</code> jó alapértelmezés; a{' '}
              <code>large-v3</code> sokkal pontosabb, de CPU-n lassú.
            </li>
            <li>
              <strong>Egy egész estés film átirata CPU-n sokáig tart</strong> — nagy modellel akár
              órákig. A folyamat a háttérben fut, közben nyugodtan használhatod az alkalmazást és
              kereshetsz a már kész átiratokban.
            </li>
          </ul>
        </section>

        <Link to="/#letoltes" className="back-link">
          ← Vissza a letöltéshez
        </Link>
      </div>
    </main>
  );
}
