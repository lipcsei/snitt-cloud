import type { Strings } from './types';

export const en: Strings = {
  htmlLang: 'en',

  meta: {
    home: {
      title: 'Snitt — Cut your favourite scene out of your own video library',
      description:
        'Snitt is a desktop app that cuts your favourite scene out of your own videos. And if you cannot remember which film the line is from, it finds that too. Everything stays on your own machine.',
    },
    install: {
      title: 'Installation — Snitt',
      description:
        'How to install ffmpeg, yt-dlp and whisper on Windows, macOS and Linux, step by step.',
    },
    profile: {
      title: 'Profile — Snitt',
      description: 'The details and settings of your Snitt website account.',
    },
  },

  sections: {
    how: 'how-it-works',
    privacy: 'your-data',
    features: 'features',
    pricing: 'pricing',
    downloads: 'download',
    account: 'account',
    faq: 'faq',
  },

  header: {
    brandAria: 'Snitt home',
    logoAlt: 'Snitt logo',
    navAria: 'Main navigation',
    navHow: 'How it works',
    navFeatures: 'Features',
    navDownloads: 'Download',
    navFaq: 'FAQ',
    login: 'Sign in',
    register: 'Sign up',
    langAria: 'Language',
    langHu: 'HU',
    langEn: 'EN',
    langDe: 'DE',
    langHuTitle: 'Magyar',
    langEnTitle: 'English',
    langDeTitle: 'Deutsch',
  },

  og: {
    imageSub: 'Desktop app: transcripts, search and scene cutting from your own videos.',
    imageAlt:
      'Snitt share image: the words “Cut out your favourite scene.” with the snitt.video address, on a dark background.',
  },

  shots: {
    eyebrow: 'What it looks like',
    title: 'Four screens, nothing spare',
    sub: 'The app runs on your own machine. These shots come from a real video library.',
    items: {
      search: {
        title: 'Search',
        body: 'Type the line and you get the film and the minute where it is said. If a film has several transcripts, it searches all of them — and shows which subtitle and which language produced the hit.',
        alt: 'Snitt search: results for the quote “tünde szemed mit lát”, with film title, timestamp and transcript source.',
      },
      editor: {
        title: 'Cutting',
        body: 'It suggests a range straight from the hit. The player shows exactly what you are about to cut, and you drag the bar to land on the right frame — or just click the sentences.',
        alt: 'The Snitt editor: transcript on the left with the selected line, player and timeline on the right.',
      },
      clips: {
        title: 'My clips',
        body: 'What you cut once stays here. Play it, search it, save it again any time — no re-cutting. This list only exists on your machine.',
        alt: 'The clips page: finished clips in a grid, with poster images and quotes.',
      },
      library: {
        title: 'Library',
        body: 'Add videos from a folder or a link. You can see what takes up how much space — and which parts you can throw away, because they are rebuilt when needed.',
        alt: 'The Snitt library: storage breakdown, import options and the added videos.',
      },
    },
  },

  hero: {
    pill: 'Desktop app · Windows · macOS · Linux',
    title: 'Cut out your\n*favourite scene*.',
    findLine: "And if you can't remember **which film** it's from — type the line, Snitt finds it.",
    lead: 'Snitt transcribes your video library and finds the line you type — in a two-hour film, in about a second. You drag the bar to set exactly where the scene starts and ends, and what you get is a finished video file on your own machine.',
    download: 'Download',
    register: 'Sign up',
    facts: ['No upload, no cloud', 'One SQLite file on your own machine', 'No account needed'],
  },

  preview: {
    windowTitle: 'Snitt — Library: 412 videos · 1,908 transcripts',
    chipHybrid: 'Full text + meaning',
    chipAll: 'Every language · every transcript',
    cutLabel: 'Cut',
    cutAction: 'Cut to file',
    previewAction: 'Preview',
    semanticBadge: 'by meaning',
    scenes: [
      {
        query: 'you shall not pass',
        stats: '2 hits · 0.3 s',
        hits: [
          {
            title: 'The Lord of the Rings: The Fellowship of the Ring (2001)',
            source: 'english subtitle',
            time: '02:31:04',
            quotes: 'en',
            before: 'You shall not ',
            hit: 'pass',
            after: '!',
            score: '98%',
          },
          {
            title: 'The Lord of the Rings: The Fellowship of the Ring (2001)',
            source: 'whisper · hungarian',
            time: '02:31:05',
            quotes: 'hu',
            before: '',
            hit: 'Nem jöhetsz át!',
            after: '',
            score: '91%',
          },
        ],
        cut: { range: '02:31:01 → 02:31:09 · 8.0 s', file: 'fellowship_02-31-01.mp4' },
      },
      {
        query: 'royale with cheese',
        stats: '2 hits · 0.2 s',
        hits: [
          {
            title: 'Pulp Fiction (1994)',
            source: 'english subtitle',
            time: '00:12:31',
            quotes: 'en',
            before: 'They call it a ',
            hit: 'Royale with cheese',
            after: '.',
            score: '97%',
          },
          {
            title: 'Pulp Fiction (1994)',
            source: 'whisper · hungarian',
            time: '00:12:34',
            quotes: 'hu',
            before: '',
            hit: 'Royale sajttal.',
            after: '',
            score: '89%',
          },
        ],
        cut: { range: '00:12:29 → 00:12:37 · 8.0 s', file: 'pulp-fiction_00-12-29.mp4' },
      },
      {
        query: 'i am your father',
        stats: '2 hits · 0.3 s',
        hits: [
          {
            title: 'Star Wars: Episode V — The Empire Strikes Back (1980)',
            source: 'english subtitle',
            time: '01:45:01',
            quotes: 'en',
            before: 'I am your ',
            hit: 'father',
            after: '.',
            score: '99%',
          },
          {
            title: 'Star Wars: Episode V — The Empire Strikes Back (1980)',
            source: 'whisper · hungarian',
            time: '01:45:02',
            quotes: 'hu',
            before: '',
            hit: 'Én vagyok az apád.',
            after: '',
            score: '92%',
          },
        ],
        cut: { range: '01:44:58 → 01:45:07 · 9.0 s', file: 'empire-strikes-back_01-44-58.mp4' },
      },
    ],
  },

  howItWorks: {
    eyebrow: 'How it works',
    title: 'Three steps from the question to the finished clip',
    sub: 'No setup marathon. You show it where your videos are, and from then on you can search what is inside them.',
    steps: [
      {
        title: 'Point it at a folder — or paste a link',
        body: 'Pick the folder on your disk where your films are. Snitt indexes them in place and never copies a single file anywhere else. For anything you do not keep locally, pasting a video link is enough.',
        note: 'Your files stay exactly where they are.',
      },
      {
        title: 'Search for the line',
        body: 'Type the sentence that stuck with you — word for word, or roughly. Snitt goes through every transcript, in every language, and shows you the exact moment it is spoken.',
        note: 'Literal and meaning-based hits at the same time.',
      },
      {
        title: 'Cut it out to a file',
        body: 'Click a hit and the cutter opens. A range is already suggested from the transcript; you drag the two ends into place. What you get is a video file, saved wherever you want it.',
        note: 'A plain local file. No watermark, no upload.',
      },
    ],
  },

  privacy: {
    eyebrow: 'The way it should work',
    title: 'Nothing goes to a remote server.',
    titleAccent: 'Your video library stays on your machine.',
    lead: 'Snitt is not a service, it is a program on your computer. Everyone builds their own database: the videos, the transcripts and the clips you cut all stay with you. No upload, no account-bound storage.',
    points: [
      {
        title: 'Your own database',
        body: 'A single SQLite file in the app data directory. You can copy it, back it up, delete it - it is yours.',
      },
      {
        title: 'Your files stay where they are',
        body: 'Films imported from a folder are never copied anywhere, only referenced. A 40 GB collection will not make the app any bigger.',
      },
      {
        title: 'The processing is local too',
        body: 'The transcript is made by your own processor (Whisper), not by a cloud API. The audio never leaves the machine.',
      },
      {
        title: 'Sharing is your call',
        body: 'The result of a cut is a file on your disk. There is no automatic share link that somebody else could open.',
      },
    ],
    note: 'What the app *does* send: it downloads the ads, and if you turn it on, it reports errors to us - neither contains any video, any transcript, or what you searched for. The website account is only there for the extras coming later; the app is fully usable without signing in.',
  },

  features: {
    eyebrow: 'Features',
    title: 'What a video library search can do when someone means it',
    sub: 'Snitt is built for your own collection: many files, many languages, not much order.',
    items: [
      {
        title: 'Search by quote',
        body: 'Not the filename, not the tags: you search what is actually said. Type the line and you get the film and the second where it is spoken.',
      },
      {
        title: 'Several subtitles per film',
        body: 'A film often has more than one subtitle file — different translations, different languages, and they do not say the same thing. Snitt keeps all of them side by side and searches every one.',
      },
      {
        title: 'Whisper transcript from the audio',
        body: 'If there is no subtitle, or you do not trust it, Whisper speech recognition generates the transcript from the audio track itself. So you get what was really said, not the translator’s version.',
      },
      {
        title: 'Search by meaning',
        body: 'We rarely remember exactly. Hybrid search puts sentences that are close in meaning next to the full-text hits, so a rough paraphrase is enough.',
      },
      {
        title: 'Visual clip cutter',
        body: 'From a hit it suggests a range straight out of the transcript, and you adjust it by dragging on the timeline. The output is a video file you save wherever you like.',
      },
      {
        title: 'Your clips in one place',
        body: 'What you cut once stays: play it, search among them, save it again any time — no re-cutting. The list only exists on your machine.',
      },
      {
        title: 'Move without reprocessing',
        body: 'Your library exports to a single file and imports on another machine. The transcripts come along, so you never re-run Whisper on what is already done.',
      },
      {
        title: 'Everything stays local',
        body: 'No server, no upload, no forced account. The index is a single SQLite file on your machine, and your videos stay exactly where they have always been.',
      },
    ],
  },

  pricing: {
    title: 'What is free, and why is there a paid tier?',
    lead: 'Search and cutting stay free - they run on your machine and cost us nothing. So does meaning-based search, if you run a local model for it. The subscription covers what genuinely costs us money: the cloud AI, per token. No catch.',
    free: {
      name: 'Free',
      price: '$0',
      priceNote: 'forever',
      items: [
        'Unlimited videos and transcripts on your own machine',
        'Quote search across every subtitle',
        'Whisper transcripts from the actual audio track',
        'Transcripts in several languages per film',
        'Clip cutting to a file, with a hand-adjustable range',
        'A searchable library of the clips you made',
        '**Meaning-based search** with a locally running model — no key, no internet',
      ],
      note: 'One cut comes with one short ad - that is what keeps the development going.',
    },
    pro: {
      flag: 'Subscription',
      name: 'Pro',
      price: 'coming soon',
      priceNote: 'monthly',
      items: [
        '**AI search with no setup:** the cloud model is more accurate, and you install nothing for it',
        '**Video understanding:** what happens in the scene, who is in it, what the dialogue is about',
        'Ask your video library questions instead of only searching it',
        'No ads',
      ],
      note: 'These features call external AI models, billed per token - which is why they do not fit into the free tier. Pricing will be final when they launch.',
      cta: 'Sign me up, tell me when it starts',
    },
    footnote:
      'The app is fully usable without an account. You only need to sign up for the Pro features - and even then your videos are not uploaded anywhere.',
  },

  downloads: {
    eyebrow: 'Download',
    title: 'Download it, then show it the folder with your videos',
    sub: 'The releases are on GitHub. Pick the package for your system.',
    platforms: [
      { name: 'Windows', detail: 'Windows 10 or newer · 64-bit', file: '.exe installer' },
      { name: 'macOS', detail: 'macOS 12 or newer · Apple Silicon and Intel', file: '.dmg image' },
      { name: 'Linux', detail: 'x86_64 · GTK/WebKit2GTK', file: '.AppImage / .deb' },
    ],
    cta: 'Download — {os}',
    notice: {
      title: 'What do you need on your machine?',
      body: 'Snitt relies on three external tools, and it deliberately does **not** bundle them:',
      items: [
        '`ffmpeg` — for cutting the videos and extracting the audio track. You install this one yourself.',
        '`yt-dlp` — if you want to bring in a video from a link. **The app downloads this one for you**, on one click, with its checksum verified.',
        '`whisper.cpp` — if you want transcripts from speech recognition too. Yours to install; the model itself Snitt fetches.',
      ],
      foot: 'If you only search existing subtitles across your own files, `ffmpeg` is enough. The other two matter when you actually use those features.',
      cta: 'Installation guide per system →',
    },
  },

  account: {
    eyebrow: 'Account',
    title: 'The app does everything without an account',
    body1:
      'Install it, open it, work with it. No sign-up wall, no trial period, and no internet needed to search your own videos. Your index lives on your machine, in a single database file.',
    body2:
      'Signing up here on the website belongs to **the extras still in the works**: we are building them, and once they are ready you will use them with your account. Until you want any of it, you lose nothing — the desktop app carries on just the same.',
    openProfile: 'Open profile',
    signedInAs: 'Signed in as **{name}**',
    register: 'Sign up',
    login: 'Sign in',
    list: [
      {
        title: 'Without an account',
        body: 'Indexing a folder, pulling in a link, reading subtitles, Whisper transcripts, search, clip cutting — the whole app, in other words.',
      },
      {
        title: 'With an account, later',
        body: 'The add-ons in progress. The core will not become paid or limited because of them.',
      },
      {
        title: 'What the account does not do',
        body: 'It does not upload your videos and it does not sync your index. Your library stays on your machine.',
      },
    ],
  },

  faq: {
    eyebrow: 'FAQ',
    title: 'Frequently asked questions',
    items: [
      {
        q: 'Does it upload my videos anywhere?',
        a: [
          'No. Snitt is a desktop app: your files stay where they are, the indexing happens locally, and the database is a single SQLite file on your machine. There is no server behind it to upload anything to.',
          'You need the internet in two cases: when you bring in a video from a link, and when you download a Whisper model for the first time.',
        ],
      },
      {
        q: 'What happens if a film has several subtitles?',
        a: [
          'All of them are kept. One film can have several translations and several languages, and they do not contain the same sentences. Snitt treats them as separate transcripts and searches all of them at once — so you still find the moment when you remember the English line but only have the Hungarian subtitle.',
        ],
      },
      {
        q: 'Does it work if I have no subtitles at all?',
        a: [
          'Yes, that is where Whisper comes in. Speech recognition builds the transcript from the audio track of the video, so you get exactly what is said — not the translator’s solution. This needs `whisper.cpp` installed, and on longer films the step takes a while; after that, search is instant.',
          'Installing it is a couple of commands per system: the [installation guide](/en/install) walks you through it.',
        ],
      },
      {
        q: 'Do I have to quote it exactly?',
        a: [
          'No. Search is hybrid: next to the literal matches it also searches by meaning, so a paraphrase or a half-remembered, slightly wrong sentence still returns a hit. The list shows you which transcript each line came from and at what second.',
        ],
      },
      {
        q: 'Do I need an account to use it?',
        a: [
          'No. The desktop app works on its own, without an account. Signing up on the website belongs to the extra features in progress; the app stays fully usable without them.',
        ],
      },
      {
        q: 'Is this legal? What can I do with the clip?',
        a: [
          'Snitt indexes **your own video library**, and it saves the scene you cut as a local file on your machine. The app does not obtain content and does not publish anything — what you may do with your own copies is your responsibility, and it depends on how you got them and where you live.',
          'Quoting a short excerpt — for analysis, criticism or commentary, with the source named — falls under quotation rules in most jurisdictions. Sharing a whole film obviously does not. In between, length, purpose and context are what matter. This is not legal advice; if you are unsure about a specific case, ask a professional.',
        ],
      },
    ],
  },

  footer: {
    tagline: 'Search your own video library — by quote, on your own machine.',
    navAria: 'Footer navigation',
    source: 'Source code',
    releases: 'Releases',
    bottom: 'All processing runs locally. Your video library never leaves your machine.',
  },

  notYet: {
    quote: '“Do not be hasty.”',
    cite: 'Treebeard — The Lord of the Rings: The Two Towers',
    download:
      'Snitt is still being built - installers for Windows, macOS and Linux will be downloadable soon.',
    register:
      'Sign-up is not live yet - accounts will be for the later Pro features. The app will be fully usable without one.',
    login:
      'Signing in is not live yet - accounts will be for the later Pro features. The app will be fully usable without one.',
    close: 'Hoom, hom',
    artAlt: 'Illustration of an old tree',
  },

  install: {
    eyebrow: 'Installation',
    title: 'Installing the external tools',
    lead: 'Snitt relies on three command line tools and deliberately does not bundle them: it uses the programs already on your machine, in your own versions. yt-dlp is the exception the app can handle for you — one click, no terminal. This page walks you through the other two, system by system.',
    callout: {
      title: 'How much of this do you actually need?',
      body: 'If you only search the subtitle files sitting next to your videos and the subtitle tracks embedded in them, **ffmpeg is enough**. The other two are not preconditions — and yt-dlp needs no manual install at all: the app offers to download it the first time you bring in a video from a link.',
    },
    tagRequired: 'Required',
    tagOptional: 'Optional',
    tagAuto: 'The app downloads it',
    deps: [
      {
        name: 'ffmpeg',
        body: 'Cutting clips, extracting the audio track, reading the subtitle tracks embedded in a video. Practically every operation needs it.',
      },
      {
        name: 'yt-dlp',
        tag: 'auto',
        body: 'Only if you want to bring a video into the library from a link — and that needs no terminal: the app offers to download it, verifies the official checksum and keeps it in its own folder. You can still install it yourself if you prefer; the app looks at its own copy first, then at your PATH.',
      },
      {
        name: 'whisper.cpp',
        body: 'Only if you want transcripts from speech recognition too. A single program called `whisper-cli` — no Python needed. Snitt downloads the model on first use.',
      },
    ],
    tabsAria: 'Operating system',
    codeLabel: 'Terminal',
    copy: 'Copy',
    copied: 'Copied',
    copyFailed: 'Failed',

    windows: {
      steps: [
        {
          title: '1. Install with winget',
          body: '`winget` ships with Windows 10 and 11, so there is nothing to install for it. Open a PowerShell window and run the lines you need.',
          codeLabel: 'PowerShell',
          hint: 'The first line is needed for everything. The second one only if you also want to bring in videos from a link — but the app offers that download too, without a terminal.',
        },
        {
          title: '2. Speech recognition (optional)',
          body: 'This part is only needed if you want transcripts for videos without subtitles. A single program, no Python.',
          codeLabel: 'PowerShell',
          hint: 'After installing, **open a new PowerShell window**, otherwise the system will not find the `whisper-cli` command yet.',
          codeLabel2: 'New PowerShell window',
        },
        {
          title: 'Alternative: Chocolatey',
          body: 'If you use Chocolatey, one line gets you `ffmpeg` and `yt-dlp`:',
          codeLabel: 'PowerShell (administrator)',
        },
      ],
      warn: {
        title: 'Open a new terminal after installing',
        body: 'Installers modify the PATH, but programs that are already running still see the old one. If `ffmpeg -version` still says it cannot be found after the install, open a new terminal window — and if Snitt was open in the meantime, restart it. This is the most common thing to get stuck on under Windows, and it is not the installation that went wrong.',
      },
    },

    macos: {
      steps: [
        {
          title: '1. Homebrew',
          body: 'On macOS the shortest route is [Homebrew](https://brew.sh). If you do not have it yet, the install command is on the brew.sh home page. If you do, skip to the next line.',
          hint: '`ffmpeg` is needed for everything. `yt-dlp` is only for importing from a link — the app can fetch that one itself, so you can leave it out here.',
        },
        {
          title: '2. Speech recognition (optional)',
          body: '`whisper.cpp` is a single program, no Python involved. Homebrew puts it on your PATH.',
          hint: 'The model is not included: Snitt downloads it on first use.',
        },
      ],
      warn: {
        title: 'The first transcription is slower',
        body: 'Speech recognition needs a model, which Snitt downloads on first use — the default `small` is about half a gigabyte. After that it stays on your machine. If your computer or connection is slow, pick a smaller one:',
        codeLabel: 'Environment variable',
        after:
          'A smaller model is faster but less accurate. In order: `tiny`, `base`, `small`, `medium`, `large-v3`.',
      },
    },

    linux: {
      steps: [
        {
          title: 'Debian / Ubuntu',
          hint: '`whisper-cpp` is only needed if you want speech recognition too.',
        },
        {
          title: 'Fedora',
          hint: '`ffmpeg` is not in Fedora’s base repositories: you have to enable the RPM Fusion repository first, otherwise the command will not find the package.',
        },
        {
          title: 'Arch',
          hint: 'The package is called `whisper.cpp` here (with a dot).',
        },
      ],
      warn: {
        title: 'yt-dlp: avoid the distribution package',
        body: 'The `yt-dlp` in the repositories is typically out of date, and the video sites change often — an old version becomes useless within weeks. The simplest route is to let the app handle it: when you import from a link it offers the official release and verifies its checksum. If you would rather do it yourself:',
      },
    },

    check: {
      title: 'Checking it worked',
      body: 'Open a terminal and run the line for whichever tool you installed. If they print something, Snitt will find them too — it looks on the same PATH.',
    },

    path: {
      title: 'If something is not on the PATH after all',
      body: 'It happens that a tool lands somewhere the system cannot see — or you keep it elsewhere on purpose. You do not have to fight the PATH for that: give the full path in the matching environment variable and Snitt will use it.',
      vars: [
        { name: 'FFMPEG_BIN', body: 'the ffmpeg executable' },
        { name: 'FFPROBE_BIN', body: 'the ffprobe executable (it ships with ffmpeg)' },
        { name: 'YTDLP_BIN', body: 'the yt-dlp executable' },
        { name: 'WHISPER_BIN', body: 'the whisper command (whisper-cli or faster-whisper)' },
      ],
    },

    good: {
      title: 'Worth knowing',
      items: [
        '**The Whisper model is downloaded on first use.** Depending on the size it ranges from a few hundred megabytes to roughly 3 gigabytes, so before the first transcript expect to need a decent connection and free disk space. After that it is local.',
        '**The model is picked by the `WHISPER_MODEL` environment variable.** Possible values: `tiny`, `base`, `small`, `medium`, `large-v3`. `small` is a good default; `large-v3` is far more accurate but slow on a CPU.',
        '**Transcribing a feature-length film on a CPU takes a long time** — with a large model it can be hours. It runs in the background, and you can keep using the app and searching the transcripts that are already done.',
      ],
    },
  },

  profile: {
    checking: 'Checking your session…',
    needLoginTitle: 'Sign-in required',
    needLoginBody: 'This page is only available when signed in. If the redirect does not start by itself:',
    login: 'Sign in',
    backHome: 'Back to the home page',
    eyebrow: 'Profile',
    greeting: 'Hi {name}!',
    intro:
      'This account belongs to the website. You do not need it to use the desktop app — it will be for the extra features in progress.',
    nameLabel: 'Name',
    emailLabel: 'Email',
    noEmail: 'not provided',
    manage: 'Manage account',
    logout: 'Sign out',
    hint: '“Manage account” opens the Keycloak account console: that is where you change your password, set up two-factor authentication or delete your account.',
  },

  common: {
    backToDownloads: '← Back to the download',
    backToHome: '← Back to the home page',
  },
};
