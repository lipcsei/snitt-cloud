/**
 * A két nyelvi tábla közös szerződése. Mindkét tábla erre a típusra van
 * annotálva, így egy kifelejtett kulcs fordítási hiba, nem futásidejű üres
 * szöveg. Ezért nincs sehol Record<string, string>: a kulcsokat itt soroljuk fel.
 *
 * A szövegekben megengedett egyszerű jelölés (a <Rich> komponens értelmezi):
 *   `kód`  **félkövér**  *kiemelt*  [link szöveg](/utvonal)  \n = sortörés
 */

export type Lang = 'hu' | 'en';

/** Nyelvfüggetlen oldalazonosítók; ezekhez tartoznak a nyelvenkénti útvonalak. */
export type RouteKey = 'home' | 'install' | 'profile';

type Two<T> = readonly [T, T];
type Three<T> = readonly [T, T, T];
type Four<T> = readonly [T, T, T, T];
type Six<T> = readonly [T, T, T, T, T, T];

export type MetaText = { title: string; description: string };

export type Step = { title: string; body: string; note: string };

export type Feature = { title: string; body: string };

export type PricingTier = {
  name: string;
  price: string;
  priceNote: string;
  items: readonly string[];
  note: string;
};

export type Platform = { name: string; detail: string; file: string };

export type FaqItem = { q: string; a: readonly string[] };

export type ListItem = { title: string; body: string };

/** Az idézőjelek nyelvfüggők: magyar „…”, angol "…". */
export type PreviewHit = {
  title: string;
  source: string;
  time: string;
  quotes: Lang;
  before: string;
  hit: string;
  after: string;
  score: string;
};

export type PreviewScene = {
  query: string;
  stats: string;
  /** Az első találat a kijelölt, a második a jelentés szerinti. */
  hits: Two<PreviewHit>;
  cut: { range: string; file: string };
};

export type InstallGuide = {
  steps: readonly {
    title: string;
    /** A parancsblokk fölötti bekezdés; néhány lépésnél csak cím és parancs van. */
    body?: string;
    /** A parancsblokk fölé kerülő címke (terminál típusa). */
    codeLabel?: string;
    /** A parancsblokk alatti magyarázat. */
    hint?: string;
    /** Néhány lépésben két parancsblokk van, a második saját címkével. */
    codeLabel2?: string;
  }[];
  warn: { title: string; body: string; codeLabel?: string; after?: string };
};

export type Strings = {
  /** <html lang> és a hreflang-hoz használt kód. */
  htmlLang: string;

  meta: {
    home: MetaText;
    install: MetaText;
    profile: MetaText;
  };

  /** Szekció-horgonyok. Nyelvenként más, hogy az angol URL is angol legyen. */
  sections: {
    how: string;
    privacy: string;
    features: string;
    pricing: string;
    downloads: string;
    account: string;
    faq: string;
  };

  header: {
    brandAria: string;
    logoAlt: string;
    navAria: string;
    navHow: string;
    navFeatures: string;
    navDownloads: string;
    navFaq: string;
    login: string;
    register: string;
    langAria: string;
    langHu: string;
    langEn: string;
    langHuTitle: string;
    langEnTitle: string;
  };

  hero: {
    pill: string;
    title: string;
    lead: string;
    download: string;
    register: string;
    facts: Three<string>;
  };

  preview: {
    windowTitle: string;
    chipHybrid: string;
    chipAll: string;
    cutLabel: string;
    cutAction: string;
    previewAction: string;
    semanticBadge: string;
    scenes: Three<PreviewScene>;
  };

  howItWorks: {
    eyebrow: string;
    title: string;
    sub: string;
    steps: Three<Step>;
  };

  privacy: {
    eyebrow: string;
    title: string;
    titleAccent: string;
    lead: string;
    points: Four<ListItem>;
    note: string;
  };

  features: {
    eyebrow: string;
    title: string;
    sub: string;
    items: Six<Feature>;
  };

  pricing: {
    title: string;
    lead: string;
    free: PricingTier;
    pro: PricingTier & { flag: string; cta: string };
    footnote: string;
  };

  downloads: {
    eyebrow: string;
    title: string;
    sub: string;
    platforms: Three<Platform>;
    /** {os} helyére kerül a rendszer neve. */
    cta: string;
    notice: {
      title: string;
      body: string;
      items: Three<string>;
      foot: string;
      cta: string;
    };
  };

  account: {
    eyebrow: string;
    title: string;
    body1: string;
    body2: string;
    openProfile: string;
    /** {name} helyére kerül a bejelentkezett felhasználó neve. */
    signedInAs: string;
    register: string;
    login: string;
    list: Three<ListItem>;
  };

  faq: {
    eyebrow: string;
    title: string;
    items: Six<FaqItem>;
  };

  footer: {
    tagline: string;
    navAria: string;
    source: string;
    releases: string;
    bottom: string;
  };

  notYet: {
    quote: string;
    cite: string;
    download: string;
    register: string;
    close: string;
    artAlt: string;
  };

  install: {
    eyebrow: string;
    title: string;
    lead: string;
    callout: { title: string; body: string };
    tagRequired: string;
    tagOptional: string;
    deps: Three<{ name: string; body: string }>;
    tabsAria: string;
    codeLabel: string;
    copy: string;
    copied: string;
    copyFailed: string;
    windows: InstallGuide;
    macos: InstallGuide;
    linux: InstallGuide;
    check: { title: string; body: string };
    path: { title: string; body: string; vars: Four<{ name: string; body: string }> };
    good: { title: string; items: Three<string> };
  };

  profile: {
    checking: string;
    needLoginTitle: string;
    needLoginBody: string;
    login: string;
    backHome: string;
    eyebrow: string;
    /** {name} helyére kerül a bejelentkezett felhasználó neve. */
    greeting: string;
    intro: string;
    nameLabel: string;
    emailLabel: string;
    noEmail: string;
    manage: string;
    logout: string;
    hint: string;
  };

  /** Szekcióhoz nem köthető, több helyen használt apróságok. */
  common: {
    backToDownloads: string;
    backToHome: string;
  };
};
