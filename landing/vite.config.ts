import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => ({
  // A GitHub Pages saját domain nélkül alkönyvtárban szolgál ki
  // (lipcsei.github.io/snitt-cloud/), ezért az alapútvonalat a CI adja meg -
  // enélkül az assetek 404-eznének. Saját domainnél ez marad "/".
  base: loadEnv(mode, '.', 'VITE_').VITE_BASE || '/',
  plugins: [react()],
  server: {
    port: 5174,
    strictPort: true,
  },
  preview: {
    port: 5174,
  },
}));
