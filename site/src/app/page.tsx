import { SiteHeader, SiteFooter } from "@/components/chrome";
import {
  Hero,
  CatalogStats,
  WhereItFits,
  LadderConsole,
  LadderCode,
  LadderKubernetes,
  Origin,
  CliAnswer,
  DesktopAnswer,
  TrustModules,
  Payoff,
  Horizon,
  Catch,
  InstallCli,
  FinalCta,
} from "@/components/marketing";

/**
 * planton.dev — the front door. Composed from small, focused section components,
 * in the order that forms the page's information architecture: hero → orientation
 * map (WhereItFits) → origin-story arc (the numbered stations) → closing CTA. The
 * header/footer are shared chrome.
 */
export default function HomePage() {
  return (
    <>
      <SiteHeader />
      <main>
        <Hero />
        <CatalogStats />
        <WhereItFits />
        <LadderConsole />
        <LadderCode />
        <LadderKubernetes />
        <Origin />
        <CliAnswer />
        <DesktopAnswer />
        <TrustModules />
        <Payoff />
        <Horizon />
        <Catch />
        <InstallCli />
        <FinalCta />
      </main>
      <SiteFooter />
    </>
  );
}
