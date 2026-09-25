import { AbsoluteFill, Img, interpolate, staticFile, useCurrentFrame } from 'remotion';
import { OVERVIEW_CHAPTERS, OVERVIEW_VIDEO } from '../src/data/homepage-video';
import { WORKFLOW_ICONS } from '../src/data/workflow-icons';
import { FlowArrowMarkers, FlowConnection } from '../src/components/marketing/workflows/FlowConnection';
import { workflowDarkTokens as palette } from '../src/theme/workflows';
import { CARD_HEIGHT, OVERVIEW_LAYOUT as layout, OVERVIEW_ROUTES as routes, type VideoCard } from './overview-layout';

const blue = palette.semantic.flow;
const text = palette.text.primary;
const muted = '#a7abb3';
const progress = (t: number, start: number, duration = 1) => Math.max(0, Math.min(1, (t - start) / duration));
const ease = (n: number) => n * n * (3 - 2 * n);

function Icon({ name, x, y, size = 60 }: { name: string; x: number; y: number; size?: number }) {
  return <image href={WORKFLOW_ICONS[name]} x={x} y={y} width={size} height={size} />;
}

function Card({ box, title, detail, icon, status, fill = 0, active = false }: {
  box: VideoCard; title: string; detail: string; icon?: string; status: string; fill?: number; active?: boolean;
}) {
  const { x, y, width } = box;
  return <g>
    <rect x={x} y={y} width={width} height={CARD_HEIGHT} rx={22} fill={active ? '#1b2331' : '#191b20'} stroke={active ? blue : '#383d46'} strokeWidth={2} />
    {icon ? <Icon name={icon} x={x + 28} y={y + 26} size={48} /> : <g stroke={blue} strokeWidth={3} fill="none"><circle cx={x + 52} cy={y + 40} r={12} /><path d={`M${x + 32} ${y + 74} v-5 a20 20 0 0 1 40 0 v5`} /></g>}
    <text x={x + 28} y={y + 116} fill={text} fontSize={36} fontWeight={600}>{title}</text>
    <text x={x + 28} y={y + 157} fill={muted} fontSize={26}>{detail}</text>
    <text x={x + width - 28} y={y + 58} textAnchor="end" fill={active || fill === 1 ? blue : muted} fontSize={23}>{status}</text>
    <rect x={x + 28} y={y + 191} width={width - 56} height={4} rx={2} fill="#383d46" />
    <rect x={x + 28} y={y + 191} width={(width - 56) * fill} height={4} rx={2} fill={blue} />
  </g>;
}

function Connections({ kind, t, windows }: { kind: keyof typeof routes; t: number; windows: readonly (readonly [number, number])[] }) {
  return <>{routes[kind].map((route, i) => <FlowConnection key={i} route={route} active={t >= windows[i][0] && t < windows[i][1]} seconds={t - windows[i][0]} marker="overview" radius={6} />)}</>;
}

function Caption({ children, small }: { children: string; small?: string }) {
  return <g><text x={960} y={866} textAnchor="middle" fill={text} fontSize={34}>{children}</text>{small && <text x={960} y={919} textAnchor="middle" fill={muted} fontSize={26}>{small}</text>}</g>;
}

function Intro({ t }: { t: number }) {
  return <><Connections kind="intro" t={t} windows={[[0.5, 3], [99, 100]]} />
    <RoleContext left="DEVELOPER / CODE READY" right="PLATFORM TEAM / SETUP REQUESTS" />
    <Card box={layout.intro[0]} title="Your Change" detail="Developer + Coding Agent" icon="github" status="Prepared" fill={1} />
    <Card box={layout.intro[1]} title="Environment Request" detail="Setup + Access + Delivery" status="Waiting" active={t > 2} />
    <Card box={layout.intro[2]} title="Running Service" detail="Needs Its Foundation" icon="KubernetesDeployment" status="Not Yet" />
    <Caption small="Infrastructure expertise should become a path the whole team can use.">The change is ready. Its path to production is not.</Caption>
  </>;
}

/** Roles remain visible during technical steps, so the film explains ownership
 * as well as execution. They are context labels, not deployment dependencies. */
function RoleContext({ left, right }: { left: string; right: string }) {
  return <g><text x={120} y={370} fill={blue} fontSize={25}>{left}</text><text x={1800} y={370} textAnchor="end" fill={muted} fontSize={25}>{right}</text></g>;
}

function Foundation({ t }: { t: number }) {
  return <><Connections kind="foundation" t={t} windows={[[2, 5], [2, 6]]} />
    <Card box={layout.foundation[0]} title="Private Network" detail="Cloud Foundation" icon="GcpVpcNetwork" status={t >= 2 ? 'Ready' : 'Creating'} fill={progress(t, 0, 2)} active={t < 2} />
    <Card box={layout.foundation[1]} title="Service Runtime" detail="Kubernetes Cluster" icon="GcpGkeCluster" status={t >= 6 ? 'Ready' : 'Preparing'} fill={progress(t, 2, 4)} active={t >= 2 && t < 6} />
    <Card box={layout.foundation[2]} title="Database" detail="Managed PostgreSQL" icon="GcpCloudSql" status={t >= 7 ? 'Ready' : 'Preparing'} fill={progress(t, 2, 5)} active={t >= 2 && t < 7} />
    <text x={1310} y={430} fill={blue} fontSize={28}>YOUR PLATFORM TEAM</text>
    <text x={1310} y={478} fill={muted} fontSize={26}>Defines the reusable path.</text>
    <text x={1310} y={546} fill={text} fontSize={45}>Reusable</text><text x={1310} y={599} fill={text} fontSize={45}>Environments</text>
    <text x={1310} y={665} fill={muted} fontSize={29}>Access + Delivery Workflows</text><text x={1310} y={711} fill={muted} fontSize={29}>Planton executes the setup.</text>
    <text x={120} y={939} fill={muted} fontSize={26}>Blue handoffs = deployment prerequisites · Illustrative architecture</text>
  </>;
}

function Delivery({ t }: { t: number }) {
  const specs = [
    { title: 'Your Change', detail: 'Developer / Coding Agent', status: 'Prepared', start: 0, end: 1.5 },
    { title: 'GitHub', detail: 'Matching Source Trigger', icon: 'github', status: 'Triggered', start: 1.5, end: 3.5 },
    { title: 'Build', detail: 'Customer-Supplied Code', icon: 'planton', status: 'Built', start: 3.5, end: 6.5 },
    { title: 'Development', detail: 'Development Config', icon: 'KubernetesDeployment', status: 'Deployed', start: 6.5, end: 9.5 },
  ];
  return <><Connections kind="delivery" t={t} windows={[[1, 3.5], [3, 6.5], [6, 9.5]]} />
    <RoleContext left="DEVELOPERS / USE THE FOUNDATION" right="PLATFORM TEAM / OWNS THE DELIVERY PATH" />
    {specs.map((s, i) => <Card key={s.title} box={layout.delivery[i]} title={s.title} detail={s.detail} icon={s.icon} status={t >= s.end ? s.status : t >= s.start ? 'In Progress' : 'Waiting'} fill={progress(t, s.start, s.end - s.start)} active={t >= s.start && t < s.end} />)}
    <Caption small="Planton builds the artifact and deploys it with development configuration.">{t >= 6.5 ? 'One versioned artifact. Ready for the next stage.' : 'From a source change to a running development service.'}</Caption>
  </>;
}

function Production({ t }: { t: number }) {
  const approved = t >= 4;
  return <><Connections kind="production" t={t} windows={[[0.5, 3], [4, 8]]} />
    <RoleContext left="PLANTON / EXECUTES AND RECORDS" right="YOUR TEAM / OWNS REQUIRED APPROVALS" />
    <Card box={layout.production[0]} title="Development" detail="Artifact: Release 42" icon="KubernetesDeployment" status="Deployed" fill={1} />
    <Card box={layout.production[1]} title="Human Approval" detail="Protected Production Stage" status={approved ? 'Approved' : 'Waiting'} fill={approved ? 1 : 0} active={!approved} />
    <Card box={layout.production[2]} title="Production" detail="Same Artifact: Release 42" icon="KubernetesDeployment" status={t >= 8 ? 'Deployed' : approved ? 'Deploying' : 'Waiting'} fill={progress(t, 4, 4)} active={approved && t < 8} />
    <Caption small={t >= 8 ? 'A deployment record keeps the result with the release.' : 'Approval rules are configured by your team.'}>{t < 4 ? 'Production waits for a person.' : t < 8 ? 'Approved. Deploy the same artifact with production configuration.' : 'A controlled release. A recorded result.'}</Caption>
  </>;
}

function Proof({ t }: { t: number }) {
  // The original capture is never retouched. A bounded camera moves from the
  // whole architecture to its actual Planton workload, then restores context.
  const zoomIn = ease(progress(t, 3, 2)), zoomOut = ease(progress(t, 9, 2));
  const zoom = 1 + 0.72 * zoomIn * (1 - zoomOut);
  const detail = t >= 4 && t < 10;
  return <>
    <div style={{ position: 'absolute', left: 90, top: 324, width: 1130, height: 630, borderRadius: 18, overflow: 'hidden', background: '#fff', border: '1px solid #454950' }}>
      <Img src={staticFile('_site/images/product/architecture-view.png')} style={{ width: '100%', height: '100%', objectFit: 'contain', transform: `scale(${zoom})`, transformOrigin: '24% 71%' }} />
    </div>
    <div style={{ position: 'absolute', left: 1290, top: 348, width: 525 }}>
      <div style={{ color: blue, fontSize: 25, letterSpacing: 2, marginBottom: 30 }}>ACTUAL PRODUCT EXPORT</div>
      <div style={{ fontSize: 46, lineHeight: 1.14, marginBottom: 26 }}>{detail ? 'Planton, running on Planton.' : 'The whole stack. In one view.'}</div>
      <div style={{ fontSize: 31, color: muted, lineHeight: 1.5 }}>Our self-hosted instance on GKE manages Planton SaaS.</div>
      <div style={{ marginTop: 44, borderTop: '1px solid #383d46', paddingTop: 30, fontSize: 29, lineHeight: 1.6 }}>
        <div style={{ color: !detail ? blue : muted }}>Cloud Foundation</div>
        <div style={{ color: detail ? blue : muted }}>Kubernetes Workloads</div>
        <div style={{ color: muted }}>Individual Resources</div>
      </div>
    </div>
  </>;
}

function Closing() {
  return <>
    <Icon name="planton" x={901} y={335} size={118} />
    {['aws', 'gcp', 'azure', 'cloudflare', 'digitalocean'].map((name, i) => <Icon key={name} name={name} x={635 + i * 140} y={518} size={90} />)}
    <text x={960} y={695} textAnchor="middle" fill={text} fontSize={38}>One team. One service. One reusable environment.</text>
    <text x={960} y={750} textAnchor="middle" fill={muted} fontSize={30}>See it with your stack.</text>
    <rect x={712} y={791} width={496} height={92} rx={16} fill={blue} />
    <text x={960} y={851} textAnchor="middle" fill="#101722" fontSize={37} fontWeight={600}>Book a Demo →</text>
    <text x={960} y={946} textAnchor="middle" fill={muted} fontSize={31}>planton.ai</text>
  </>;
}

/** Deterministic film adapter: authored at 1920×1080 and rendered at either
 * resolution. Remotion stays in this export-only directory, outside Next. */
export function HomepageOverview() {
  const seconds = useCurrentFrame() / OVERVIEW_VIDEO.fps;
  const index = OVERVIEW_CHAPTERS.findIndex(c => seconds < c.end);
  const chapter = OVERVIEW_CHAPTERS[Math.max(0, index)];
  const t = seconds - chapter.start;
  const opacity = interpolate(t, [0, 0.3, chapter.end - chapter.start - 0.22, chapter.end - chapter.start], [0, 1, 1, index === 5 ? 1 : 0], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });
  return <AbsoluteFill style={{ background: '#111316', color: text, fontFamily: 'Arial, sans-serif' }}>
    <svg width={1920} height={1080} viewBox="0 0 1920 1080" style={{ position: 'absolute' }}>
      <Icon name="planton" x={90} y={49} size={43} />
      <text x={155} y={80} fill={muted} fontSize={25}>PLANTON / A PATH TO PRODUCTION</text>
      <text x={1830} y={80} textAnchor="end" fill={muted} fontSize={25}>{String(index + 1).padStart(2, '0')} / 06</text>
      <line x1={90} y1={119} x2={1830} y2={119} stroke="#30343b" />
      <g opacity={opacity}>
        <text x={90} y={222} fill={text} fontSize={64} fontWeight={500} letterSpacing={-1.5}>{chapter.title}</text>
        <text x={90} y={286} fill={index === 4 || index === 5 ? blue : muted} fontSize={34}>{chapter.subtitle}</text>
        <FlowArrowMarkers id="overview" />
        {index === 0 && <Intro t={t} />}
        {index === 1 && <Foundation t={t} />}
        {index === 2 && <Delivery t={t} />}
        {index === 3 && <Production t={t} />}
        {index === 5 && <Closing />}
      </g>
      <text x={90} y={1035} fill={muted} fontSize={23}>{index === 4 ? 'Actual Planton architecture export · Camera movement only' : 'Illustrative workflow · Time compressed'}</text>
      <text x={1830} y={1035} textAnchor="end" fill={muted} fontSize={23}>{chapter.label}</text>
      {OVERVIEW_CHAPTERS.map((c, i) => <g key={c.start}><rect x={90 + i * 292} y={982} width={278} height={3} fill="#30343b" /><rect x={90 + i * 292} y={982} width={278 * progress(seconds, c.start, c.end - c.start)} height={3} fill={blue} /></g>)}
    </svg>
    {index === 4 && <div style={{ opacity }}><Proof t={t} /></div>}
  </AbsoluteFill>;
}
