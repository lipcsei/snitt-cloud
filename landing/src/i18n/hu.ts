import type { Strings } from './types';

export const hu: Strings = {
  htmlLang: 'hu',

  meta: {
    home: {
      title: 'Snitt — Keresés a saját videótáradban, idézet alapján',
      description:
        'A Snitt asztali alkalmazás átiratot készít a videóidról, megkeresi a beírt idézetet, és kivágja belőle a jelenetet. Minden helyben marad a gépeden.',
    },
    install: {
      title: 'Telepítés — Snitt',
      description:
        'Az ffmpeg, a yt-dlp és a faster-whisper telepítése Windowson, macOS-en és Linuxon, lépésről lépésre.',
    },
    profile: {
      title: 'Profil — Snitt',
      description: 'A Snitt weboldali fiókod adatai és beállításai.',
    },
  },

  sections: {
    how: 'hogyan',
    privacy: 'adataid',
    features: 'funkciok',
    pricing: 'arak',
    downloads: 'letoltes',
    account: 'fiok',
    faq: 'gyik',
  },

  header: {
    brandAria: 'Snitt főoldal',
    logoAlt: 'Snitt logó',
    navAria: 'Fő navigáció',
    navHow: 'Hogyan működik',
    navFeatures: 'Funkciók',
    navDownloads: 'Letöltés',
    navFaq: 'GYIK',
    login: 'Bejelentkezés',
    register: 'Regisztráció',
    langAria: 'Nyelvválasztó',
    langHu: 'HU',
    langEn: 'EN',
    langDe: 'DE',
    langHuTitle: 'Magyar',
    langEnTitle: 'English',
    langDeTitle: 'Deutsch',
  },

  hero: {
    pill: 'Asztali alkalmazás · Windows · macOS · Linux',
    title: 'Megvan a mondat.\nCsak azt nem tudod, *melyik filmben*.',
    cutLine: 'Vagy tudod — és csak **ki akarod vágni a kedvenc jelenetedet**.',
    lead: 'A Snitt leiratot készít a videótáradról, megkeresi benne a beírt idézetet — egy kétórás filmben másodpercek alatt —, te pedig a sávot húzva állítod be, hol kezdődjön és hol érjen véget a jelenet. A végeredmény egy kész videófájl a saját gépeden.',
    download: 'Letöltés',
    register: 'Regisztráció',
    facts: [
      'Nincs feltöltés, nincs felhő',
      'Egy SQLite fájl a saját gépeden',
      'Fiók nélkül is teljes értékű',
    ],
  },

  preview: {
    windowTitle: 'Snitt — Könyvtár: 412 videó · 1 908 átirat',
    chipHybrid: 'Teljes szöveg + jelentés',
    chipAll: 'Minden nyelv · minden átirat',
    cutLabel: 'Kivágás',
    cutAction: 'Kivágás fájlba',
    previewAction: 'Előnézet',
    semanticBadge: 'jelentés szerint',
    scenes: [
      {
        query: 'nem mehetsz át',
        stats: '2 találat · 0,3 mp',
        hits: [
          {
            title: 'A Gyűrűk Ura: A Gyűrű Szövetsége (2001)',
            source: 'magyar felirat',
            time: '02:31:05',
            quotes: 'hu',
            before: '',
            hit: 'Nem mehetsz át!',
            after: '',
            score: '98%',
          },
          {
            title: 'A Gyűrűk Ura: A Gyűrű Szövetsége (2001)',
            source: 'whisper · angol',
            time: '02:31:04',
            quotes: 'en',
            before: 'You shall not ',
            hit: 'pass',
            after: '!',
            score: '91%',
          },
        ],
        cut: { range: '02:31:01 → 02:31:09 · 8,0 mp', file: 'gyuruk-ura_02-31-01.mp4' },
      },
      {
        query: 'royale sajttal',
        stats: '2 találat · 0,2 mp',
        hits: [
          {
            title: 'Ponyvaregény (1994)',
            source: 'magyar felirat',
            time: '00:12:34',
            quotes: 'hu',
            before: '',
            hit: 'Royale sajttal.',
            after: '',
            score: '97%',
          },
          {
            title: 'Ponyvaregény (1994)',
            source: 'whisper · angol',
            time: '00:12:31',
            quotes: 'en',
            before: 'They call it a ',
            hit: 'Royale with cheese',
            after: '.',
            score: '89%',
          },
        ],
        cut: { range: '00:12:29 → 00:12:37 · 8,0 mp', file: 'ponyvaregeny_00-12-29.mp4' },
      },
      {
        query: 'én vagyok az apád',
        stats: '2 találat · 0,3 mp',
        hits: [
          {
            title: 'Star Wars V. rész — A Birodalom visszavág (1980)',
            source: 'magyar felirat',
            time: '01:45:02',
            quotes: 'hu',
            before: '',
            hit: 'Én vagyok az apád.',
            after: '',
            score: '99%',
          },
          {
            title: 'Star Wars V. rész — A Birodalom visszavág (1980)',
            source: 'whisper · angol',
            time: '01:45:01',
            quotes: 'en',
            before: 'I am your ',
            hit: 'father',
            after: '.',
            score: '92%',
          },
        ],
        cut: { range: '01:44:58 → 01:45:07 · 9,0 mp', file: 'birodalom-visszavag_01-44-58.mp4' },
      },
    ],
  },

  howItWorks: {
    eyebrow: 'Hogyan működik',
    title: 'Három lépés a kérdéstől a kész klipig',
    sub: 'Nincs beállítási maraton. Megmutatod, hol vannak a videóid, és onnantól kereshető a tartalmuk.',
    steps: [
      {
        title: 'Behúzod: mappa vagy link',
        body: 'Kiválasztod a merevlemezeden azt a mappát, ahol a filmjeid vannak — a Snitt helyben indexeli őket, egyetlen fájlt sem másol át máshová. Ha valamit nem tárolsz helyben, elég bemásolni egy videó linkjét.',
        note: 'A fájljaid ott maradnak, ahol vannak.',
      },
      {
        title: 'Rákeresel az idézetre',
        body: 'Beírod azt a mondatot, ami megmaradt — akár szó szerint, akár csak nagyjából. A Snitt végigmegy az összes átiraton, minden nyelven, és megmutatja a pontos időpontot, ahol elhangzik.',
        note: 'Szó szerinti és jelentés szerinti találat egyszerre.',
      },
      {
        title: 'Kivágod fájlba',
        body: 'A találatra kattintva megnyílik a vágó. Az átirat alapján már be van állítva egy javasolt tartomány, a végét-elejét húzással igazítod. A végeredmény egy videófájl, oda mented, ahová akarod.',
        note: 'Sima helyi fájl, se vízjel, se feltöltés.',
      },
    ],
  },

  privacy: {
    eyebrow: 'Ahogy működnie kell',
    title: 'Semmi nem kerül távoli szerverre.',
    titleAccent: 'A videótárad a te gépeden marad.',
    lead: 'A Snitt nem szolgáltatás, hanem egy program a gépeden. Mindenki a saját adatbázisát építi fel: a videók, a leiratok és a kivágott részletek is nálad maradnak. Nincs feltöltés, nincs fiókhoz kötött tárhely.',
    points: [
      {
        title: 'Saját adatbázis',
        body: 'Egyetlen SQLite fájl az alkalmazás adatkönyvtárában. Le tudod másolni, el tudod menteni, ki tudod törölni - a tiéd.',
      },
      {
        title: 'A fájljaid a helyükön maradnak',
        body: 'A mappából importált filmeket nem másolja sehova, csak hivatkozik rájuk. Egy 40 GB-os gyűjteménytől sem hízik meg a program.',
      },
      {
        title: 'A feldolgozás is helyben fut',
        body: 'A leirat a te processzorodon készül (Whisper), nem egy felhős API-n. A hanganyag nem hagyja el a gépet.',
      },
      {
        title: 'A megosztás a te döntésed',
        body: 'A kivágás eredménye egy fájl a lemezeden. Nincs automatikus megosztó link, amit valaki más is elérhetne.',
      },
    ],
    note: 'Amit az alkalmazás *küld*: a hirdetéseket letölti, és ha bekapcsolod, a hibákat jelenti nekünk - ezekben nincs benne se videó, se leirat, se az, amit kerestél. A fiók a weboldalon csak a későbbi extrákhoz kell; az alkalmazás bejelentkezés nélkül is teljes értékű.',
  },

  features: {
    eyebrow: 'Funkciók',
    title: 'Amit egy videótár keresője tud, ha komolyan gondolják',
    sub: 'A Snitt a saját gyűjteményedhez készült: sok fájl, sok nyelv, kevés rendszerezés.',
    items: [
      {
        title: 'Keresés idézetre',
        body: 'Nem a fájlnévre, nem a címkékre: arra keresel, ami elhangzik. Beírod a mondatot, és megkapod a filmet meg a másodpercet, ahol elhangzik.',
      },
      {
        title: 'Több felirat egy filmhez',
        body: 'Egy filmhez gyakran több felirat is tartozik — más fordítások, más nyelvek, és nem ugyanazt írják. A Snitt mindet megtartja egymás mellett, és mindegyikben keres.',
      },
      {
        title: 'Whisper átirat a hangból',
        body: 'Ha nincs felirat, vagy nem bízol benne, a Whisper beszédfelismerés legenerálja az átiratot magából a hangsávból. Így azt kapod, ami tényleg elhangzott, nem a fordító változatát.',
      },
      {
        title: 'Jelentés szerinti keresés',
        body: 'Ritkán emlékszünk pontosan. A hibrid keresés a teljes szöveges találatok mellé a jelentés alapján hasonló mondatokat is behozza, így a körülírás is elég.',
      },
      {
        title: 'Vizuális klipvágó',
        body: 'A találatból az átirat alapján rögtön javasol egy tartományt, amit az idővonalon húzással igazítasz. A kimenet egy videófájl, amit oda mentesz, ahová akarsz.',
      },
      {
        title: 'Minden helyben marad',
        body: 'Nincs szerver, nincs feltöltés, nincs fiókkényszer. Az index egyetlen SQLite fájl a gépeden, a videóid pedig ott maradnak, ahol eddig is voltak.',
      },
    ],
  },

  pricing: {
    title: 'Mi ingyenes, és miért van fizetős rész?',
    lead: 'A keresés és a kivágás ingyenes marad - ez a gépeden fut, nekünk nem kerül pénzbe. Az AI-alapú funkciók viszont valódi tokenköltséget jelentenek, ezért csak előfizetéssel érhetők el. Nincs benne trükk.',
    free: {
      name: 'Ingyenes',
      price: '0 Ft',
      priceNote: 'örökre',
      items: [
        'Korlátlan videó és leirat a saját gépeden',
        'Idézet-alapú keresés az összes feliratban',
        'Whisper átirat a tényleges hangsávból',
        'Több nyelvű leiratok filmenként',
        'Klipvágás fájlba, kézzel állítható tartománnyal',
      ],
      note: 'Egy kivágás egy rövid hirdetés megtekintésével jár - ebből tartjuk fenn a fejlesztést.',
    },
    pro: {
      flag: 'Előfizetés',
      name: 'Pro',
      price: 'hamarosan',
      priceNote: 'havidíjas',
      items: [
        '**AI-keresés:** nem csak a szavakat találja meg, hanem a jelenetet is, amire gondolsz - körülírásból is',
        '**Videóértelmezés:** mi történik a jelenetben, kik szerepelnek benne, miről szól a párbeszéd',
        'Kérdezhetsz a videótáradtól, nem csak kereshetsz benne',
        'Nincsenek hirdetések',
      ],
      note: 'Ezek a funkciók külső AI-modelleket használnak, ami tokenenként fizetendő - ezért nem fér bele az ingyenes csomagba. Az árazás a bevezetéskor lesz végleges.',
      cta: 'Regisztrálok, szóljatok, ha indul',
    },
    footnote:
      'Az alkalmazás fiók nélkül is teljes értékű. A regisztráció csak a Pro funkciókhoz kell - a videóid akkor sem kerülnek fel sehova.',
  },

  downloads: {
    eyebrow: 'Letöltés',
    title: 'Töltsd le, és mutasd meg neki a videóid mappáját',
    sub: 'A kiadások a GitHubon érhetők el. Válaszd ki a rendszeredhez tartozó csomagot.',
    platforms: [
      { name: 'Windows', detail: 'Windows 10 vagy újabb · 64 bites', file: '.exe telepítő' },
      {
        name: 'macOS',
        detail: 'macOS 12 vagy újabb · Apple Silicon és Intel',
        file: '.dmg lemezkép',
      },
      { name: 'Linux', detail: 'x86_64 · GTK/WebKit2GTK környezet', file: '.AppImage / .deb' },
    ],
    cta: 'Letöltés — {os}',
    notice: {
      title: 'Mire van szükség a gépeden?',
      body: 'A Snitt három külső eszközre támaszkodik, és ezeket **nem** csomagolja magába — külön kell telepítened őket:',
      items: [
        '`ffmpeg` — a videók vágásához és a hangsáv kinyeréséhez.',
        '`yt-dlp` — ha linkről szeretnél videót behozni.',
        '`faster-whisper` — ha beszédfelismeréssel is szeretnél átiratot készíteni.',
      ],
      foot: 'Ha csak meglévő feliratokban keresel a saját fájljaid között, elég az `ffmpeg`. A többi akkor kell, amikor tényleg használod őket.',
      cta: 'Telepítési útmutató rendszerenként →',
    },
  },

  account: {
    eyebrow: 'Fiók',
    title: 'Az alkalmazás fiók nélkül is mindent tud',
    body1:
      'Telepíted, megnyitod, dolgozol vele. Nincs regisztrációs fal, nincs próbaidőszak, és nem kell internet ahhoz, hogy keress a saját videóidban. Az indexed a te gépeden él, egyetlen adatbázisfájlban.',
    body2:
      'A regisztráció itt a weboldalon **a készülő extrákhoz** tartozik: ezeken még dolgozunk, és amint elérhetők, a fiókoddal tudod majd használni őket. Amíg nem kérsz belőlük, semmit nem veszítesz — az asztali alkalmazás ugyanúgy megy tovább.',
    openProfile: 'Profil megnyitása',
    signedInAs: 'Bejelentkezve mint **{name}**',
    register: 'Regisztráció',
    login: 'Bejelentkezés',
    list: [
      {
        title: 'Fiók nélkül',
        body: 'Mappa indexelése, link behúzása, feliratok beolvasása, Whisper átirat, keresés, klipvágás — vagyis az egész alkalmazás.',
      },
      {
        title: 'Fiókkal, később',
        body: 'A készülő kiegészítők. Az alap működés ettől nem lesz fizetős vagy korlátozott.',
      },
      {
        title: 'Amit a fiók nem csinál',
        body: 'Nem tölti fel a videóidat, és nem szinkronizálja az indexedet. A könyvtárad a gépeden marad.',
      },
    ],
  },

  faq: {
    eyebrow: 'GYIK',
    title: 'Gyakori kérdések',
    items: [
      {
        q: 'Feltölti valahová a videóimat?',
        a: [
          'Nem. A Snitt asztali alkalmazás: a fájljaid ott maradnak, ahol vannak, az indexelés helyben történik, az adatbázis pedig egyetlen SQLite fájl a gépeden. Nincs mögötte szerver, amire feltöltene bármit.',
          'Internet két esetben kell: ha linkről hozol be videót, illetve ha első alkalommal töltesz le egy Whisper modellt.',
        ],
      },
      {
        q: 'Mi történik, ha egy filmhez több felirat is tartozik?',
        a: [
          'Mindegyik megmarad. Egy filmhez tartozhat több fordítás és több nyelv, és ezek nem ugyanazokat a mondatokat tartalmazzák. A Snitt külön átiratként kezeli őket, a keresés pedig egyszerre fut mindegyiken — így akkor is találsz, ha épp az angol változatra emlékszel, de a magyar felirat van meg.',
        ],
      },
      {
        q: 'Akkor is működik, ha egyáltalán nincs feliratom?',
        a: [
          'Igen, ilyenkor jön a Whisper. A beszédfelismerés a videó hangsávjából készít átiratot, tehát pont azt kapod, ami elhangzik — nem a fordító megoldását. Ehhez a `faster-whisper` telepítése szükséges, és a hosszabb filmeknél ez a lépés időbe telik; utána viszont a keresés már azonnali.',
          'A telepítése rendszerenként pár parancs: végigvezet rajta a [telepítési útmutató](/telepites).',
        ],
      },
      {
        q: 'Muszáj pontosan idéznem?',
        a: [
          'Nem. A keresés hibrid: a szó szerinti egyezés mellett jelentés alapján is keres, így a körülírásra és az emlékezetből felidézett, kicsit pontatlan mondatra is hoz találatot. A listában látod, melyik átiratból és hányadik másodpercből származik az adott sor.',
        ],
      },
      {
        q: 'Kell fiók a használatához?',
        a: [
          'Nem. Az asztali alkalmazás önállóan, fiók nélkül működik. A weboldali regisztráció a készülő extra funkciókhoz tartozik, ezek nélkül is teljes értékű marad az alkalmazás.',
        ],
      },
      {
        q: 'Legális ez? Mit kezdhetek a kivágott klippel?',
        a: [
          'A Snitt a **saját videótáradat** indexeli, és a kivágott jelenetet helyi fájlként menti a gépedre. Az alkalmazás nem szerez be tartalmat és nem tesz közzé semmit — hogy mihez van jogod a saját másolataiddal, az a te felelősséged, és attól függ, hogyan jutottál hozzájuk, illetve hol élsz.',
          'Egy rövid részlet idézése — elemzéshez, kritikához, kommentárhoz, forrásmegjelöléssel — a legtöbb jogrendszerben az idézés szabályai alá esik. Egy egész film megosztása nyilván nem. A kettő között a hossz, a cél és a kontextus számít. Ez nem jogi tanács; ha konkrét ügyben bizonytalan vagy, kérdezz szakértőt.',
        ],
      },
    ],
  },

  footer: {
    tagline: 'Keresés a saját videótáradban — idézet alapján, a saját gépeden.',
    navAria: 'Lábléc navigáció',
    source: 'Forráskód',
    releases: 'Kiadások',
    bottom: 'Minden feldolgozás helyben fut. A videótárad nem hagyja el a gépedet.',
  },

  notYet: {
    quote: '„Ne legyetek hamariak!"',
    cite: 'Szirszakáll — A Gyűrűk Ura: A két torony',
    download:
      'A Snitt még készül - a telepítők hamarosan letölthetők lesznek Windowsra, macOS-re és Linuxra.',
    register:
      'A regisztráció még nem él - a fiókok a későbbi Pro funkciókhoz kellenek majd. Az alkalmazás fiók nélkül is teljes értékű lesz.',
    login:
      'A bejelentkezés még nem él - a fiókok a későbbi Pro funkciókhoz kellenek majd. Az alkalmazás fiók nélkül is teljes értékű lesz.',
    close: 'Búrárum',
    artAlt: 'Öreg fa illusztráció',
  },

  install: {
    eyebrow: 'Telepítés',
    title: 'Külső eszközök telepítése',
    lead: 'A Snitt három parancssori eszközre támaszkodik, és ezeket szándékosan nem csomagolja magába: a gépeden lévő, saját verziójú programokat használja. Ez az oldal végigvezet a telepítésükön, rendszerenként.',
    callout: {
      title: 'Mennyi kell ebből tényleg?',
      body: 'Ha csak a videóid mellett lévő feliratfájlokban és a videókba ágyazott feliratsávokban keresel, **elég az ffmpeg**. A másik kettő nem előfeltétel: akkor kell telepítened őket, amikor először használnád azt a funkciót.',
    },
    tagRequired: 'Kötelező',
    tagOptional: 'Opcionális',
    deps: [
      {
        name: 'ffmpeg',
        body: 'Klipek vágása, hangsáv kinyerése, a videóba ágyazott feliratsávok kiolvasása. Gyakorlatilag minden művelethez kell.',
      },
      {
        name: 'yt-dlp',
        body: 'Csak akkor, ha linkről szeretnél videót behozni a tárba.',
      },
      {
        name: 'faster-whisper',
        body: 'Csak akkor, ha beszédfelismeréssel is szeretnél átiratot készíteni. A parancsot a `faster-whisper-cli` Python-csomag telepíti.',
      },
    ],
    tabsAria: 'Operációs rendszer',
    codeLabel: 'Terminál',
    copy: 'Másolás',
    copied: 'Másolva',
    copyFailed: 'Nem sikerült',

    windows: {
      steps: [
        {
          title: '1. Telepítés wingettel',
          body: 'A `winget` a Windows 10 és 11 része, külön telepíteni nem kell. Nyiss egy PowerShell ablakot, és futtasd a szükséges sorokat.',
          codeLabel: 'PowerShell',
          hint: 'Az első sor kell mindenhez. A második csak akkor, ha linkről is szeretnél videót behozni.',
        },
        {
          title: '2. Beszédfelismerés (opcionális)',
          body: 'Ez a rész csak akkor kell, ha felirat nélküli videókhoz is szeretnél átiratot. Először Python kell hozzá, utána maga a csomag.',
          codeLabel: 'PowerShell',
          hint: 'A Python telepítése után **nyiss egy új PowerShell ablakot**, különben a `pip` parancsot még nem találja meg a rendszer.',
          codeLabel2: 'Új PowerShell ablak',
        },
        {
          title: 'Alternatíva: Chocolatey',
          body: 'Ha Chocolateyt használsz, egy sorral is megvan az `ffmpeg`, a `yt-dlp` és a Python:',
          codeLabel: 'PowerShell (rendszergazda)',
        },
      ],
      warn: {
        title: 'Telepítés után nyiss új terminált',
        body: 'A telepítők a PATH-ot módosítják, a már futó programok viszont a régi PATH-ot látják. Ha a telepítés után az `ffmpeg -version` még mindig azt írja, hogy nem található, nyiss egy új terminálablakot — ha pedig a Snitt közben nyitva volt, indítsd újra. Ez a leggyakoribb elakadás Windowson, és nem a telepítéssel van baj.',
      },
    },

    macos: {
      steps: [
        {
          title: '1. Homebrew',
          body: 'macOS-en a legrövidebb út a [Homebrew](https://brew.sh). Ha még nincs fent, a telepítőparancsot a brew.sh kezdőlapján találod. Ha már megvan, ugorj a következő sorra.',
          hint: 'Az `ffmpeg` kell mindenhez, a `yt-dlp` csak a linkről importáláshoz.',
        },
        {
          title: '2. Beszédfelismerés (opcionális)',
          body: 'A legtisztább megoldás a `pipx`: külön környezetbe teszi a Python-eszközöket, és gondoskodik róla, hogy a parancs a PATH-ra kerüljön.',
          hint: 'A `pipx ensurepath` a shell profilodat írja át, tehát utána nyiss egy új terminált.',
        },
      ],
      warn: {
        title: 'Ha pipx nélkül telepíted',
        body: 'A `pip3 install --user faster-whisper-cli` is működik, de a parancsot a `~/Library/Python/3.x/bin` könyvtárba teszi, ami alapból **nincs rajta a PATH-on**. Ilyenkor a telepítés sikerül, a Snitt viszont nem fogja megtalálni a `faster-whisper` parancsot. Vedd fel a könyvtárat a shell profilodba (`~/.zshrc`):',
        codeLabel: '~/.zshrc',
        after:
          'A verziószámnak egyeznie kell a saját Pythonodéval — nézd meg a `python3 --version` kimenetét, és azt írd be a `3.9` helyére, különben a sor nem csinál semmit.',
      },
    },

    linux: {
      steps: [
        {
          title: 'Debian / Ubuntu',
          hint: 'A második sor csak akkor kell, ha beszédfelismerést is szeretnél. A `pipx ensurepath` után nyiss új terminált.',
        },
        {
          title: 'Fedora',
          hint: 'Az `ffmpeg` nincs benne a Fedora alap tárolóiban: ehhez előbb engedélyezned kell az RPM Fusion tárolót, különben a parancs nem találja a csomagot.',
        },
        {
          title: 'Arch',
          hint: 'A Whisper ezután `pipx install faster-whisper-cli` paranccsal jön.',
        },
      ],
      warn: {
        title: 'yt-dlp: kerüld a disztribúciós csomagot',
        body: 'A tárolókban lévő `yt-dlp` jellemzően elavult, a videómegosztók pedig gyakran változnak — egy régi verzió hetek alatt használhatatlanná válik. Töltsd le inkább a hivatalos binárist, ez frissíti magát:',
      },
    },

    check: {
      title: 'Ellenőrzés',
      body: 'Nyiss egy terminált, és futtasd le azt a sort, amelyik eszközt telepítetted. Ha ezek kiírnak valamit, a Snitt is meg fogja találni őket — ugyanazon a PATH-on keresi.',
    },

    path: {
      title: 'Ha valami mégsincs a PATH-on',
      body: 'Előfordul, hogy egy eszköz olyan helyre kerül, ahonnan a rendszer nem látja — vagy szándékosan máshol tartod. Ilyenkor nem kell a PATH-tal küzdeni: add meg a teljes elérési utat a megfelelő környezeti változóban, és a Snitt azt fogja használni.',
      vars: [
        { name: 'FFMPEG_BIN', body: 'az ffmpeg futtatható fájlja' },
        { name: 'FFPROBE_BIN', body: 'az ffprobe futtatható fájlja (az ffmpeg mellett érkezik)' },
        { name: 'YTDLP_BIN', body: 'a yt-dlp futtatható fájlja' },
        { name: 'WHISPER_BIN', body: 'a faster-whisper parancs' },
      ],
    },

    good: {
      title: 'Jó tudni',
      items: [
        '**A Whisper modell az első használatkor töltődik le.** Mérettől függően néhány száz megabájttól nagyjából 3 gigabájtig terjed, tehát az első átirat előtt érdemes rendes netre és szabad helyre számítani. Utána már helyben van.',
        '**A modellt a `WHISPER_MODEL` környezeti változó választja ki.** Lehetséges értékek: `tiny`, `base`, `small`, `medium`, `large-v3`. A `small` jó alapértelmezés; a `large-v3` sokkal pontosabb, de CPU-n lassú.',
        '**Egy egész estés film átirata CPU-n sokáig tart** — nagy modellel akár órákig. A folyamat a háttérben fut, közben nyugodtan használhatod az alkalmazást és kereshetsz a már kész átiratokban.',
      ],
    },
  },

  profile: {
    checking: 'Bejelentkezés ellenőrzése…',
    needLoginTitle: 'Bejelentkezés szükséges',
    needLoginBody: 'Ez az oldal csak bejelentkezve érhető el. Ha az átirányítás nem indul el magától:',
    login: 'Bejelentkezés',
    backHome: 'Vissza a főoldalra',
    eyebrow: 'Profil',
    greeting: 'Szia, {name}!',
    intro:
      'Ez a fiók a weboldalhoz tartozik. Az asztali alkalmazás használatához nincs rá szükség — a készülő extra funkciókhoz lesz.',
    nameLabel: 'Név',
    emailLabel: 'E-mail',
    noEmail: 'nincs megadva',
    manage: 'Fiók kezelése',
    logout: 'Kijelentkezés',
    hint: 'A „Fiók kezelése” a Keycloak fiókkonzolját nyitja meg: ott tudsz jelszót változtatni, kétlépcsős azonosítást beállítani vagy a fiókodat törölni.',
  },

  common: {
    backToDownloads: '← Vissza a letöltéshez',
    backToHome: '← Vissza a főoldalra',
  },
};
