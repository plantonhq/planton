import { SiteHeader, SiteFooter } from "@/components/chrome";
import {
  Hero,
  CatalogStats,
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
 * planton.dev — the origin-story front door. Composed from small, focused act
 * components; the header/footer are shared chrome.
 */
export default function HomePage() {
  return (
    <>
      <SiteHeader />
      <main>
        <Hero />
        <CatalogStats />
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
