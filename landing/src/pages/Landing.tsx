import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import Hero from '../sections/Hero';
import HowItWorks from '../sections/HowItWorks';
import Screenshots from '../sections/Screenshots';
import Privacy from '../sections/Privacy';
import Features from '../sections/Features';
import Pricing from '../sections/Pricing';
import Downloads from '../sections/Downloads';
import Account from '../sections/Account';
import Faq from '../sections/Faq';

type ScrollState = { scrollTo?: string } | null;

export default function Landing() {
  const location = useLocation();

  // Másik útvonalról érkezve a SectionLink itt adja át, hová kell görgetni.
  useEffect(() => {
    const target = (location.state as ScrollState)?.scrollTo ?? location.hash.replace('#', '');
    if (!target) return;
    const el = document.getElementById(target);
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }, [location]);

  return (
    <main>
      <Hero />
      <HowItWorks />
      <Screenshots />
      <Privacy />
      <Features />
      <Pricing />
      <Downloads />
      <Account />
      <Faq />
    </main>
  );
}
