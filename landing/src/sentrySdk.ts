// Az SDK-ból csak azt exportálja, amit a sentry.tsx használ: a dinamikus import a modul
// teljes névterét visszaadja, ezért az `import('@sentry/react')` az egész SDK-t (replay,
// feedback stb.) a csomagba húzná - ezen a szűk homlokzaton át viszont a többi kimarad.
export { captureException, init, withScope } from '@sentry/react';
