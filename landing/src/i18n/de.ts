import type { Strings } from './types';

export const de: Strings = {
  htmlLang: 'de',

  meta: {
    home: {
      title: 'Snitt — Schneide deine Lieblingsszene aus deiner Videosammlung',
      description:
        'Snitt ist eine Desktop-App, die dir deine Lieblingsszene aus deinen eigenen Videos herausschneidet. Und wenn du nicht mehr weißt, aus welchem Film der Satz stammt, findet sie ihn auch. Alles bleibt auf deinem eigenen Rechner.',
    },
    install: {
      title: 'Installation — Snitt',
      description:
        'ffmpeg, yt-dlp und Whisper unter Windows, macOS und Linux installieren — Schritt für Schritt.',
    },
    profile: {
      title: 'Profil — Snitt',
      description: 'Die Daten und Einstellungen deines Snitt-Kontos auf der Website.',
    },
  },

  sections: {
    how: 'so-funktioniert-es',
    privacy: 'deine-daten',
    features: 'funktionen',
    pricing: 'preise',
    downloads: 'download',
    account: 'konto',
    faq: 'faq',
  },

  header: {
    brandAria: 'Snitt Startseite',
    logoAlt: 'Snitt Logo',
    navAria: 'Hauptnavigation',
    navHow: 'So funktioniert es',
    navFeatures: 'Funktionen',
    navDownloads: 'Download',
    navFaq: 'FAQ',
    login: 'Anmelden',
    register: 'Registrieren',
    langAria: 'Sprachauswahl',
    langHu: 'HU',
    langEn: 'EN',
    langDe: 'DE',
    langHuTitle: 'Magyar',
    langEnTitle: 'English',
    langDeTitle: 'Deutsch',
  },

  og: {
    imageSub: 'Desktop-App: Transkripte, Suche und Szenenschnitt aus deinen eigenen Videos.',
    imageAlt:
      'Snitt-Teilenbild: der Satz „Schneide deine Lieblingsszene aus.“ mit der Adresse snitt.video auf dunklem Hintergrund.',
  },

  shots: {
    eyebrow: 'So sieht es aus',
    title: 'Vier Bildschirme, nichts Überflüssiges',
    sub: 'Die App läuft auf deinem eigenen Rechner. Diese Aufnahmen stammen aus einer echten Videosammlung.',
    items: {
      search: {
        title: 'Suche',
        body: 'Du tippst den Satz ein und bekommst den Film und die Minute, in der er fällt. Hat ein Film mehrere Transkripte, wird in allen gesucht – und du siehst, welcher Untertitel und welche Sprache den Treffer geliefert hat.',
        alt: 'Die Snitt-Suche: Treffer für das Zitat „tünde szemed mit lát“, mit Filmtitel, Zeitpunkt und Transkriptquelle.',
      },
      editor: {
        title: 'Schnitt',
        body: 'Aus dem Treffer wird sofort ein Bereich vorgeschlagen. Der Player zeigt genau das, was du herausschneiden wirst, und mit der Leiste triffst du das richtige Bild – oder du klickst einfach auf die Sätze.',
        alt: 'Der Snitt-Editor: links das Transkript mit dem markierten Satz, rechts Player und Zeitleiste.',
      },
      clips: {
        title: 'Meine Clips',
        body: 'Was du einmal geschnitten hast, bleibt hier. Abspielen, durchsuchen, jederzeit erneut speichern – ohne neuen Schnitt. Diese Liste existiert nur auf deinem Rechner.',
        alt: 'Die Clip-Seite: fertige Clips im Raster, mit Vorschaubildern und Zitaten.',
      },
      library: {
        title: 'Sammlung',
        body: 'Videos kommen aus einem Ordner oder von einem Link. Du siehst, was wie viel Platz braucht – und was du bedenkenlos löschen kannst, weil es bei Bedarf neu entsteht.',
        alt: 'Die Snitt-Sammlung: Speicherübersicht, Importmöglichkeiten und die aufgenommenen Videos.',
      },
    },
  },

  hero: {
    pill: 'Desktop-App · Windows · macOS · Linux',
    title: 'Schneide deine\n*Lieblingsszene* aus.',
    findLine: 'Und wenn du nicht mehr weißt, **aus welchem Film** — tipp den Satz ein, Snitt findet ihn.',
    lead: 'Snitt erstellt Transkripte deiner Videosammlung und findet darin den Satz, den du eintippst — in einem Zweistundenfilm in etwa einer Sekunde. Du ziehst die Leiste auf den genauen Anfang und das Ende der Szene, heraus kommt eine fertige Videodatei auf deinem eigenen Rechner.',
    download: 'Herunterladen',
    register: 'Registrieren',
    facts: [
      'Kein Upload, keine Cloud',
      'Eine SQLite-Datei auf deinem eigenen Rechner',
      'Auch ohne Konto vollwertig',
    ],
  },

  preview: {
    windowTitle: 'Snitt — Bibliothek: 412 Videos · 1.908 Transkripte',
    chipHybrid: 'Volltext + Bedeutung',
    chipAll: 'Alle Sprachen · alle Transkripte',
    cutLabel: 'Schnitt',
    cutAction: 'In Datei schneiden',
    previewAction: 'Vorschau',
    semanticBadge: 'nach Bedeutung',
    scenes: [
      {
        query: 'du kannst nicht vorbei',
        stats: '2 Treffer · 0,3 s',
        hits: [
          {
            title: 'Der Herr der Ringe: Die Gefährten (2001)',
            source: 'deutscher Untertitel',
            time: '02:31:05',
            quotes: 'de',
            before: '',
            hit: 'Du kannst nicht vorbei!',
            after: '',
            score: '98%',
          },
          {
            title: 'Der Herr der Ringe: Die Gefährten (2001)',
            source: 'whisper · englisch',
            time: '02:31:04',
            quotes: 'en',
            before: 'You shall not ',
            hit: 'pass',
            after: '!',
            score: '91%',
          },
        ],
        cut: { range: '02:31:01 → 02:31:09 · 8,0 s', file: 'die-gefaehrten_02-31-01.mp4' },
      },
      {
        query: 'royale mit käse',
        stats: '2 Treffer · 0,2 s',
        hits: [
          {
            title: 'Pulp Fiction (1994)',
            source: 'deutscher Untertitel',
            time: '00:12:34',
            quotes: 'de',
            before: 'Die nennen das ',
            hit: 'Royale mit Käse',
            after: '.',
            score: '97%',
          },
          {
            title: 'Pulp Fiction (1994)',
            source: 'whisper · englisch',
            time: '00:12:31',
            quotes: 'en',
            before: 'They call it a ',
            hit: 'Royale with cheese',
            after: '.',
            score: '89%',
          },
        ],
        cut: { range: '00:12:29 → 00:12:37 · 8,0 s', file: 'pulp-fiction_00-12-29.mp4' },
      },
      {
        query: 'ich bin dein vater',
        stats: '2 Treffer · 0,3 s',
        hits: [
          {
            title: 'Star Wars: Episode V — Das Imperium schlägt zurück (1980)',
            source: 'deutscher Untertitel',
            time: '01:45:02',
            quotes: 'de',
            before: 'Ich bin dein ',
            hit: 'Vater',
            after: '.',
            score: '99%',
          },
          {
            title: 'Star Wars: Episode V — Das Imperium schlägt zurück (1980)',
            source: 'whisper · englisch',
            time: '01:45:01',
            quotes: 'en',
            before: 'I am your ',
            hit: 'father',
            after: '.',
            score: '92%',
          },
        ],
        cut: { range: '01:44:58 → 01:45:07 · 9,0 s', file: 'das-imperium_01-44-58.mp4' },
      },
    ],
  },

  howItWorks: {
    eyebrow: 'So funktioniert es',
    title: 'Drei Schritte von der Frage zum fertigen Clip',
    sub: 'Kein Einrichtungsmarathon. Du zeigst Snitt, wo deine Videos liegen — ab da ist ihr Inhalt durchsuchbar.',
    steps: [
      {
        title: 'Ordner auswählen — oder Link einfügen',
        body: 'Du wählst auf deiner Festplatte den Ordner mit deinen Filmen aus. Snitt indexiert sie an Ort und Stelle und kopiert keine einzige Datei woandershin. Was du nicht lokal liegen hast, holst du dir mit einem eingefügten Videolink.',
        note: 'Deine Dateien bleiben genau da, wo sie sind.',
      },
      {
        title: 'Nach dem Zitat suchen',
        body: 'Du tippst den Satz ein, der hängen geblieben ist — wortwörtlich oder nur ungefähr. Snitt geht alle Transkripte durch, in jeder Sprache, und zeigt dir den genauen Moment, in dem er fällt.',
        note: 'Wörtliche und sinngemäße Treffer auf einmal.',
      },
      {
        title: 'In eine Datei schneiden',
        body: 'Ein Klick auf den Treffer öffnet den Schnitt. Aus dem Transkript ist schon ein Bereich vorgeschlagen, Anfang und Ende ziehst du zurecht. Heraus kommt eine Videodatei — gespeichert, wohin du willst.',
        note: 'Eine ganz normale lokale Datei. Kein Wasserzeichen, kein Upload.',
      },
    ],
  },

  privacy: {
    eyebrow: 'So, wie es sein sollte',
    title: 'Nichts landet auf einem fremden Server.',
    titleAccent: 'Deine Videosammlung bleibt auf deinem Rechner.',
    lead: 'Snitt ist kein Dienst, sondern ein Programm auf deinem Rechner. Jeder baut seine eigene Datenbank auf: die Videos, die Transkripte und die geschnittenen Ausschnitte bleiben bei dir. Kein Upload, kein kontogebundener Speicher.',
    points: [
      {
        title: 'Deine eigene Datenbank',
        body: 'Eine einzige SQLite-Datei im Datenverzeichnis der App. Du kannst sie kopieren, sichern, löschen - sie gehört dir.',
      },
      {
        title: 'Deine Dateien bleiben an ihrem Platz',
        body: 'Aus einem Ordner importierte Filme werden nirgendwohin kopiert, sondern nur referenziert. Auch von einer 40 GB großen Sammlung wird das Programm nicht größer.',
      },
      {
        title: 'Auch die Verarbeitung läuft lokal',
        body: 'Das Transkript entsteht auf deinem eigenen Prozessor (Whisper), nicht über eine Cloud-API. Der Ton verlässt den Rechner nicht.',
      },
      {
        title: 'Teilen ist deine Entscheidung',
        body: 'Das Ergebnis eines Schnitts ist eine Datei auf deiner Festplatte. Es gibt keinen automatischen Freigabelink, den jemand anderes öffnen könnte.',
      },
    ],
    note: 'Was die App *sendet*: Sie lädt die Werbung herunter, und wenn du es einschaltest, meldet sie uns Fehler - darin steckt weder ein Video noch ein Transkript noch das, wonach du gesucht hast. Das Konto auf der Website ist nur für die späteren Extras da; die App ist auch ohne Anmeldung vollwertig.',
  },

  features: {
    eyebrow: 'Funktionen',
    title: 'Was die Suche in einer Videosammlung kann, wenn man sie ernst meint',
    sub: 'Snitt ist für deine eigene Sammlung gebaut: viele Dateien, viele Sprachen, wenig Ordnung.',
    items: [
      {
        title: 'Suche nach Zitat',
        body: 'Nicht der Dateiname, nicht die Schlagwörter: Du suchst nach dem, was gesprochen wird. Satz eintippen — und du bekommst den Film und die Sekunde, in der er fällt.',
      },
      {
        title: 'Mehrere Untertitel pro Film',
        body: 'Zu einem Film gehören oft mehrere Untertiteldateien — andere Übersetzungen, andere Sprachen, und sie sagen nicht dasselbe. Snitt behält sie alle nebeneinander und durchsucht jede einzelne.',
      },
      {
        title: 'Whisper-Transkript aus dem Ton',
        body: 'Wenn kein Untertitel da ist oder du ihm nicht traust, erzeugt die Whisper-Spracherkennung das Transkript direkt aus der Tonspur. So bekommst du, was wirklich gesagt wurde — nicht die Fassung des Übersetzers.',
      },
      {
        title: 'Suche nach Bedeutung',
        body: 'Genau erinnern wir uns selten. Die hybride Suche stellt neben die Volltexttreffer auch sinnverwandte Sätze, eine grobe Umschreibung reicht also.',
      },
      {
        title: 'Visueller Clip-Schnitt',
        body: 'Aus dem Treffer schlägt Snitt anhand des Transkripts sofort einen Bereich vor, den du auf der Zeitleiste per Ziehen anpasst. Heraus kommt eine Videodatei, die du speicherst, wohin du willst.',
      },
      {
        title: 'Fertige Clips an einem Ort',
        body: 'Was du einmal geschnitten hast, bleibt: abspielen, durchsuchen, jederzeit erneut speichern – ohne neuen Schnitt. Die Liste existiert nur auf deinem Rechner.',
      },
      {
        title: 'Umziehen ohne neue Verarbeitung',
        body: 'Deine Sammlung lässt sich in eine einzige Datei exportieren und auf einem anderen Rechner importieren. Die Transkripte kommen mit – Whisper muss nichts wiederholen, was schon fertig ist.',
      },
      {
        title: 'Alles bleibt lokal',
        body: 'Kein Server, kein Upload, kein Kontozwang. Der Index ist eine einzige SQLite-Datei auf deinem Rechner, und deine Videos bleiben genau dort, wo sie schon immer lagen.',
      },
    ],
  },

  pricing: {
    title: 'Was ist kostenlos, und wofür gibt es einen Bezahltarif?',
    lead: 'Suche und Schnitt bleiben kostenlos - sie laufen auf deinem Rechner und kosten uns nichts. Auch die Suche nach Bedeutung, wenn du dafür ein lokales Modell laufen lässt. Das Abo deckt das ab, was uns wirklich Geld kostet: die Cloud-KI, pro Token. Kein Haken dabei.',
    free: {
      name: 'Kostenlos',
      price: '0 €',
      priceNote: 'für immer',
      items: [
        'Unbegrenzt Videos und Transkripte auf deinem eigenen Rechner',
        'Zitatsuche über alle Untertitel hinweg',
        'Whisper-Transkripte aus der echten Tonspur',
        'Mehrsprachige Transkripte pro Film',
        'Clip-Schnitt in eine Datei, mit dem Bereich von Hand einstellbar',
        'Eine durchsuchbare Sammlung deiner fertigen Clips',
        '**Suche nach Bedeutung** mit einem lokal laufenden Modell – ohne Schlüssel, ohne Internet',
      ],
      note: 'Zu einem Schnitt gehört eine kurze Werbeeinblendung - davon finanzieren wir die Entwicklung.',
    },
    pro: {
      flag: 'Abo',
      name: 'Pro',
      price: 'demnächst',
      priceNote: 'monatlich',
      items: [
        '**KI-Suche ohne Einrichtung:** das Cloud-Modell ist genauer, und du musst dafür nichts installieren',
        '**Videoverständnis:** was in der Szene passiert, wer darin vorkommt, worum es im Dialog geht',
        'Du kannst deine Videosammlung fragen, nicht nur durchsuchen',
        'Keine Werbung',
      ],
      note: 'Diese Funktionen rufen externe KI-Modelle auf, die pro Token abgerechnet werden - deshalb passen sie nicht in den kostenlosen Tarif. Der endgültige Preis steht zum Start fest.',
      cta: 'Registrieren, sagt mir Bescheid, wenn es losgeht',
    },
    footnote:
      'Die App ist auch ohne Konto vollwertig. Die Registrierung brauchst du nur für die Pro-Funktionen - und auch dann werden deine Videos nirgendwohin hochgeladen.',
  },

  downloads: {
    eyebrow: 'Download',
    title: 'Herunterladen, dann den Ordner mit deinen Videos zeigen',
    sub: 'Die Releases liegen auf GitHub. Wähl das Paket für dein System.',
    platforms: [
      { name: 'Windows', detail: 'Windows 10 oder neuer · 64 Bit', file: '.exe-Installer' },
      {
        name: 'macOS',
        detail: 'macOS 12 oder neuer · Apple Silicon und Intel',
        file: '.dmg-Image',
      },
      { name: 'Linux', detail: 'x86_64 · GTK/WebKit2GTK', file: '.AppImage / .deb' },
    ],
    cta: 'Herunterladen — {os}',
    notice: {
      title: 'Was brauchst du auf deinem Rechner?',
      body: 'Snitt stützt sich auf drei externe Werkzeuge und bringt sie bewusst **nicht** mit:',
      items: [
        '`ffmpeg` — zum Schneiden der Videos und zum Extrahieren der Tonspur. Das installierst du selbst.',
        '`yt-dlp` — wenn du ein Video über einen Link holen willst. **Das lädt die App für dich**, auf einen Klick und mit geprüfter Prüfsumme.',
        '`whisper.cpp` — wenn du Transkripte auch per Spracherkennung erzeugen willst. Auch selbst; das Modell holt Snitt dagegen schon.',
      ],
      foot: 'Wenn du nur vorhandene Untertitel in deinen eigenen Dateien durchsuchst, reicht `ffmpeg`. Die anderen beiden brauchst du erst, wenn du diese Funktionen wirklich nutzt.',
      cta: 'Installationsanleitung pro System →',
    },
  },

  account: {
    eyebrow: 'Konto',
    title: 'Die App kann alles auch ohne Konto',
    body1:
      'Installieren, öffnen, loslegen. Keine Registrierungshürde, keine Testphase, und kein Internet nötig, um in deinen eigenen Videos zu suchen. Dein Index liegt auf deinem Rechner, in einer einzigen Datenbankdatei.',
    body2:
      'Die Registrierung hier auf der Website gehört zu **den Extras, die noch entstehen**: Wir bauen daran, und sobald sie fertig sind, nutzt du sie mit deinem Konto. Solange du nichts davon brauchst, verlierst du nichts — die Desktop-App läuft genauso weiter.',
    openProfile: 'Profil öffnen',
    signedInAs: 'Angemeldet als **{name}**',
    register: 'Registrieren',
    login: 'Anmelden',
    list: [
      {
        title: 'Ohne Konto',
        body: 'Ordner indexieren, Link holen, Untertitel einlesen, Whisper-Transkript, Suche, Clip-Schnitt — also die ganze App.',
      },
      {
        title: 'Mit Konto, später',
        body: 'Die Erweiterungen, die noch entstehen. Der Kern wird dadurch weder kostenpflichtig noch eingeschränkt.',
      },
      {
        title: 'Was das Konto nicht tut',
        body: 'Es lädt deine Videos nicht hoch und synchronisiert deinen Index nicht. Deine Sammlung bleibt auf deinem Rechner.',
      },
    ],
  },

  faq: {
    eyebrow: 'FAQ',
    title: 'Häufige Fragen',
    items: [
      {
        q: 'Lädt Snitt meine Videos irgendwohin hoch?',
        a: [
          'Nein. Snitt ist eine Desktop-App: Deine Dateien bleiben, wo sie sind, die Indexierung passiert lokal, und die Datenbank ist eine einzige SQLite-Datei auf deinem Rechner. Dahinter steckt kein Server, auf den überhaupt etwas hochgeladen werden könnte.',
          'Internet brauchst du in zwei Fällen: wenn du ein Video über einen Link holst, und wenn du zum ersten Mal ein Whisper-Modell herunterlädst.',
        ],
      },
      {
        q: 'Was passiert, wenn zu einem Film mehrere Untertitel gehören?',
        a: [
          'Alle bleiben erhalten. Zu einem Film können mehrere Übersetzungen und mehrere Sprachen gehören, und sie enthalten nicht dieselben Sätze. Snitt behandelt sie als getrennte Transkripte und sucht in allen gleichzeitig — so findest du die Stelle auch dann, wenn dir die englische Zeile im Kopf geblieben ist, du aber nur den deutschen Untertitel hast.',
        ],
      },
      {
        q: 'Funktioniert das auch, wenn ich überhaupt keine Untertitel habe?',
        a: [
          'Ja, dafür ist Whisper da. Die Spracherkennung erstellt das Transkript aus der Tonspur des Videos, du bekommst also genau das, was gesagt wird — nicht die Lösung des Übersetzers. Dafür muss `whisper.cpp` installiert sein, und bei längeren Filmen dauert dieser Schritt eine Weile; danach ist die Suche sofort da.',
          'Die Installation sind pro System ein paar Befehle: Die [Installationsanleitung](/de/installation) führt dich durch.',
        ],
      },
      {
        q: 'Muss ich wörtlich zitieren?',
        a: [
          'Nein. Die Suche ist hybrid: Neben den wörtlichen Treffern sucht sie auch nach Bedeutung, deshalb liefert auch eine Umschreibung oder ein aus dem Gedächtnis leicht falsch zitierter Satz einen Treffer. In der Liste siehst du, aus welchem Transkript und aus welcher Sekunde die Zeile stammt.',
        ],
      },
      {
        q: 'Brauche ich ein Konto, um Snitt zu nutzen?',
        a: [
          'Nein. Die Desktop-App funktioniert für sich allein, ohne Konto. Die Registrierung auf der Website gehört zu den Zusatzfunktionen, die noch entstehen; ohne sie bleibt die App vollwertig.',
        ],
      },
      {
        q: 'Ist das legal? Was darf ich mit dem Ausschnitt machen?',
        a: [
          'Snitt indexiert **deine eigene Videosammlung** und speichert die geschnittene Szene als lokale Datei auf deinem Rechner. Die App beschafft keine Inhalte und veröffentlicht nichts — was du mit deinen eigenen Kopien tun darfst, liegt in deiner Verantwortung und hängt davon ab, wie du an sie gekommen bist und wo du lebst.',
          'Ein kurzer Ausschnitt, zitiert zur Analyse, zur Kritik oder mit eigenem Kommentar und mit Quellenangabe, fällt in den meisten Rechtsordnungen unter das Zitatrecht. Einen ganzen Film zu teilen offensichtlich nicht. Dazwischen zählen Länge, Zweck und Kontext. Das ist keine Rechtsberatung; wenn du dir in einem konkreten Fall unsicher bist, frag jemanden vom Fach.',
        ],
      },
    ],
  },

  footer: {
    tagline: 'Die eigene Videosammlung durchsuchen — nach Zitat, auf dem eigenen Rechner.',
    navAria: 'Navigation in der Fußzeile',
    source: 'Quellcode',
    releases: 'Releases',
    bottom: 'Die gesamte Verarbeitung läuft lokal. Deine Videosammlung verlässt deinen Rechner nicht.',
  },

  notYet: {
    quote: '„Seid nicht hastig!“',
    cite: 'Baumbart — Der Herr der Ringe: Die zwei Türme',
    download:
      'Snitt entsteht noch - die Installer für Windows, macOS und Linux gibt es bald zum Herunterladen.',
    register:
      'Die Registrierung ist noch nicht offen - Konten werden für die späteren Pro-Funktionen gebraucht. Die App wird auch ohne Konto vollwertig sein.',
    login:
      'Die Anmeldung ist noch nicht offen - Konten werden für die späteren Pro-Funktionen gebraucht. Die App wird auch ohne Konto vollwertig sein.',
    close: 'Hum, hom',
    artAlt: 'Illustration eines alten Baumes',
  },

  install: {
    eyebrow: 'Installation',
    title: 'Die externen Werkzeuge installieren',
    lead: 'Snitt stützt sich auf drei Kommandozeilenwerkzeuge und bringt sie bewusst nicht mit: Es nutzt die Programme, die auf deinem Rechner liegen, in deinen eigenen Versionen. Die Ausnahme ist yt-dlp — das lädt die App auf einen Klick selbst, ganz ohne Terminal. Durch die anderen beiden führt dich diese Seite, System für System.',
    callout: {
      title: 'Wie viel davon brauchst du wirklich?',
      body: 'Wenn du nur in den Untertiteldateien neben deinen Videos und in den eingebetteten Untertitelspuren suchst, **reicht ffmpeg**. Die anderen beiden sind keine Voraussetzung — und yt-dlp musst du gar nicht von Hand installieren: Die App bietet den Download an, sobald du zum ersten Mal ein Video über einen Link holst.',
    },
    tagRequired: 'Erforderlich',
    tagOptional: 'Optional',
    tagAuto: 'Die App lädt es',
    deps: [
      {
        name: 'ffmpeg',
        body: 'Clips schneiden, Tonspur extrahieren, die im Video eingebetteten Untertitelspuren auslesen. Praktisch jede Operation braucht es.',
      },
      {
        name: 'yt-dlp',
        tag: 'auto',
        body: 'Nur dann, wenn du ein Video über einen Link in die Sammlung holen willst — und dafür brauchst du kein Terminal: Die App bietet den Download an, prüft die offizielle Prüfsumme und legt die Datei in ihren eigenen Ordner. Von Hand geht es weiterhin; die App schaut zuerst in ihren eigenen Ordner, dann in den PATH.',
      },
      {
        name: 'whisper.cpp',
        body: 'Nur dann, wenn du Transkripte auch per Spracherkennung erzeugen willst. Ein einziges Programm namens `whisper-cli` — ohne Python. Das Modell lädt Snitt beim ersten Mal herunter.',
      },
    ],
    tabsAria: 'Betriebssystem',
    codeLabel: 'Terminal',
    copy: 'Kopieren',
    copied: 'Kopiert',
    copyFailed: 'Fehlgeschlagen',

    windows: {
      steps: [
        {
          title: '1. Installation mit winget',
          body: '`winget` ist Teil von Windows 10 und 11, dafür musst du nichts extra installieren. Öffne ein PowerShell-Fenster und führe die Zeilen aus, die du brauchst.',
          codeLabel: 'PowerShell',
          hint: 'Die erste Zeile brauchst du für alles. Die zweite nur, wenn du Videos auch über einen Link holen willst — diesen Download bietet dir die App aber auch ohne Terminal an.',
        },
        {
          title: '2. Spracherkennung (optional)',
          body: 'Diesen Teil brauchst du nur, wenn du auch für Videos ohne Untertitel ein Transkript willst. Ein einziges Programm, ohne Python.',
          codeLabel: 'PowerShell',
          hint: 'Öffne nach der Installation **ein neues PowerShell-Fenster**, sonst findet das System den Befehl `whisper-cli` noch nicht.',
          codeLabel2: 'Neues PowerShell-Fenster',
        },
        {
          title: 'Alternative: Chocolatey',
          body: 'Wenn du Chocolatey nutzt, holst du `ffmpeg` und `yt-dlp` mit einer einzigen Zeile:',
          codeLabel: 'PowerShell (als Administrator)',
        },
      ],
      warn: {
        title: 'Nach der Installation ein neues Terminal öffnen',
        body: 'Die Installer ändern den PATH, bereits laufende Programme sehen aber noch den alten. Wenn `ffmpeg -version` nach der Installation immer noch meldet, dass der Befehl nicht gefunden wird, öffne ein neues Terminalfenster — und wenn Snitt inzwischen offen war, starte es neu. Das ist unter Windows die häufigste Stolperstelle, und es liegt nicht an der Installation.',
      },
    },

    macos: {
      steps: [
        {
          title: '1. Homebrew',
          body: 'Unter macOS ist [Homebrew](https://brew.sh) der kürzeste Weg. Wenn du es noch nicht hast, findest du den Installationsbefehl auf der Startseite von brew.sh. Wenn es schon da ist, spring zur nächsten Zeile.',
          hint: '`ffmpeg` brauchst du für alles. `yt-dlp` nur für den Import über einen Link — das holt sich die App auf Wunsch selbst, du kannst es hier also weglassen.',
        },
        {
          title: '2. Spracherkennung (optional)',
          body: '`whisper.cpp` ist ein einziges Programm, ganz ohne Python. Homebrew legt es in den PATH.',
          hint: 'Das Modell ist nicht dabei: Snitt lädt es beim ersten Mal herunter.',
        },
      ],
      warn: {
        title: 'Die erste Erkennung dauert länger',
        body: 'Die Spracherkennung braucht ein Modell, das Snitt beim ersten Mal herunterlädt — das voreingestellte `small` ist rund ein halbes Gigabyte. Danach bleibt es auf deinem Rechner. Wenn dein Rechner oder deine Leitung langsam ist, nimm ein kleineres:',
        codeLabel: 'Umgebungsvariable',
        after:
          'Ein kleineres Modell ist schneller, aber ungenauer. Der Reihe nach: `tiny`, `base`, `small`, `medium`, `large-v3`.',
      },
    },

    linux: {
      steps: [
        {
          title: 'Debian / Ubuntu',
          hint: '`whisper-cpp` brauchst du nur, wenn du auch Spracherkennung willst.',
        },
        {
          title: 'Fedora',
          hint: '`ffmpeg` liegt nicht in den Basis-Repositories von Fedora: Du musst zuerst das RPM-Fusion-Repository aktivieren, sonst findet der Befehl das Paket nicht.',
        },
        {
          title: 'Arch',
          hint: 'Das Paket heißt hier `whisper.cpp` (mit Punkt).',
        },
      ],
      warn: {
        title: 'yt-dlp: das Distributionspaket besser meiden',
        body: 'Das `yt-dlp` aus den Repositories ist meist veraltet, und die Videoportale ändern sich häufig — eine alte Version wird binnen Wochen unbrauchbar. Am einfachsten überlässt du es der App: Beim Import über einen Link bietet sie die offizielle Version an und prüft deren Prüfsumme. Wenn du es lieber selbst machst:',
      },
    },

    check: {
      title: 'Prüfen, ob es geklappt hat',
      body: 'Öffne ein Terminal und führe die Zeile für das Werkzeug aus, das du installiert hast. Wenn sie etwas ausgeben, findet Snitt sie auch — es sucht im selben PATH.',
    },

    path: {
      title: 'Wenn doch etwas nicht im PATH liegt',
      body: 'Es kommt vor, dass ein Werkzeug an einer Stelle landet, die das System nicht sieht — oder du hältst es absichtlich woanders. Dafür musst du dich nicht mit dem PATH herumschlagen: Gib den vollständigen Pfad in der passenden Umgebungsvariable an, dann nutzt Snitt genau den.',
      vars: [
        { name: 'FFMPEG_BIN', body: 'die ausführbare Datei von ffmpeg' },
        { name: 'FFPROBE_BIN', body: 'die ausführbare Datei von ffprobe (kommt mit ffmpeg mit)' },
        { name: 'YTDLP_BIN', body: 'die ausführbare Datei von yt-dlp' },
        { name: 'WHISPER_BIN', body: 'der Whisper-Befehl (whisper-cli oder faster-whisper)' },
      ],
    },

    good: {
      title: 'Gut zu wissen',
      items: [
        '**Das Whisper-Modell wird beim ersten Einsatz heruntergeladen.** Je nach Größe sind das ein paar hundert Megabyte bis rund 3 Gigabyte — vor dem ersten Transkript brauchst du also eine ordentliche Leitung und freien Speicherplatz. Danach liegt es lokal.',
        '**Welches Modell läuft, entscheidet die Umgebungsvariable `WHISPER_MODEL`.** Mögliche Werte: `tiny`, `base`, `small`, `medium`, `large-v3`. `small` ist eine gute Voreinstellung; `large-v3` ist deutlich genauer, auf der CPU aber langsam.',
        '**Das Transkript eines abendfüllenden Films dauert auf der CPU lange** — mit einem großen Modell auch mal Stunden. Der Vorgang läuft im Hintergrund, du kannst die App weiter nutzen und in den fertigen Transkripten suchen.',
      ],
    },
  },

  profile: {
    checking: 'Anmeldung wird geprüft…',
    needLoginTitle: 'Anmeldung erforderlich',
    needLoginBody: 'Diese Seite ist nur angemeldet erreichbar. Wenn die Weiterleitung nicht von selbst startet:',
    login: 'Anmelden',
    backHome: 'Zurück zur Startseite',
    eyebrow: 'Profil',
    greeting: 'Hallo {name}!',
    intro:
      'Dieses Konto gehört zur Website. Für die Desktop-App brauchst du es nicht — es ist für die Zusatzfunktionen gedacht, die noch entstehen.',
    nameLabel: 'Name',
    emailLabel: 'E-Mail',
    noEmail: 'nicht angegeben',
    manage: 'Konto verwalten',
    logout: 'Abmelden',
    hint: '„Konto verwalten“ öffnet die Kontokonsole von Keycloak: Dort änderst du dein Passwort, richtest die Zwei-Faktor-Anmeldung ein oder löschst dein Konto.',
  },

  common: {
    backToDownloads: '← Zurück zum Download',
    backToHome: '← Zurück zur Startseite',
  },
};
