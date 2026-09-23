import { HERO as H } from '../../data/homepage-experience';
import { WORKFLOW_ICONS } from '../../data/workflow-icons';
import { workflowDarkTokens as p } from '../../theme/workflows';
import { FlowArrowMarkers, FlowConnection } from './workflows/FlowConnection';
import { createRoundedRoute, type Point } from './workflows/flowGeometry';

/** The same authored lanes carry requests down and results back. Reversing the
 * route reverses both packets and arrowheads; no decorative return loop is needed. */
const desktopLanes: readonly (readonly Point[])[] = [
  [
    [104, 108],
    [104, 196],
    [268, 196],
  ],
  [
    [304, 108],
    [304, 160],
  ],
  [
    [504, 108],
    [504, 196],
    [340, 196],
  ],
  [
    [304, 302],
    [304, 332],
  ],
];
const mobileLanes: readonly (readonly Point[])[] = [
  [
    [70, 100],
    [70, 124],
    [118, 124],
    [118, 146],
  ],
  [
    [210, 100],
    [210, 124],
    [162, 124],
    [162, 146],
  ],
  [
    [78, 203],
    [104, 203],
  ],
  [
    [140, 314],
    [140, 354],
  ],
];
const routeSet = (lanes: readonly (readonly Point[])[]) =>
  lanes.map((points) => ({
    forward: createRoundedRoute(points, 18),
    reverse: createRoundedRoute([...points].reverse(), 18),
  }));
const desktopRoutes = routeSet(desktopLanes);
const mobileRoutes = routeSet(mobileLanes);
const providers = ['aws', 'gcp', 'azure', 'cloudflare', 'digitalocean'];

function PersonMark({ x, y }: { x: number; y: number }) {
  return (
    <g transform={`translate(${x} ${y})`} stroke={p.semantic.flow} strokeWidth="1.5" fill="none">
      <circle cy="-5" r="5" />
      <path d="M-9 13v-3c0-11 18-11 18 0v3" />
    </g>
  );
}

/** Collaboration overview: recognizable inputs, one central mark, and a cloud
 * boundary. It is deliberately distinct from the resource DAGs farther down. */
export function HeroScene({
  seconds,
  id,
  compact = false,
}: {
  seconds: number;
  id: string;
  compact?: boolean;
}) {
  const phase = Math.min(3, Math.floor(seconds / 4)),
    finished = seconds >= 16;
  const routes = compact ? mobileRoutes : desktopRoutes;
  const connections = routes.map((route, i) => {
    const returning = phase === 3 && !finished && (i === 1 || i === 3);
    const active =
      !finished &&
      (i === 0 ? phase === 0 : i < 3 ? phase === 1 || returning : phase === 2 || returning);
    return (
      <FlowConnection
        key={i}
        route={returning ? route.reverse : route.forward}
        active={active}
        seconds={seconds}
        marker={id}
        radius={compact ? 2.5 : 3}
      />
    );
  });
  return (
    <svg
      viewBox={compact ? '0 0 280 524' : '0 0 608 466'}
      role="img"
      aria-labelledby={`${id}-title ${id}-desc`}
      style={{ fontFamily: 'var(--font-inter),Arial,sans-serif' }}
    >
      <title id={`${id}-title`}>{H.title}</title>
      <desc id={`${id}-desc`}>
        {H.description} {H.providers}
      </desc>
      <FlowArrowMarkers id={id} />
      {connections}
      {compact ? (
        <>
          {[8, 148].map((x) => (
            <rect
              key={x}
              x={x}
              y="8"
              width="124"
              height="92"
              rx="10"
              fill={p.surface.panel}
              stroke={p.edge.default}
            />
          ))}
          <g textAnchor="middle" fill={p.text.primary} fontSize="17">
            <text x="70" y="40">
              {H.compactTeam[0]}
            </text>
            <text x="70" y="65">
              {H.compactTeam[1]}
            </text>
            <text x="210" y="40">
              {H.compactTeam[2]}
            </text>
            <text x="210" y="67" fill={p.text.secondary} fontSize="14">
              {H.compactTeam[3]}
            </text>
          </g>
          <image href={WORKFLOW_ICONS.github} x="34" y="185" width="32" height="32" />
          <text x="50" y="240" fill={p.text.secondary} textAnchor="middle" fontSize="15">
            {H.repository}
          </text>
          <rect
            x="104"
            y="146"
            width="72"
            height="88"
            rx="16"
            fill={p.surface.raised}
            stroke={p.edge.default}
          />
          <image href={WORKFLOW_ICONS.planton} x="123" y="167" width="34" height="42" />
          <text x="140" y="273" textAnchor="middle" fill={p.text.primary} fontSize="17">
            {H.engineLabel}
          </text>
          <text x="140" y="302" textAnchor="middle" fill={p.text.secondary} fontSize="16">
            {H.controls}
          </text>
          <rect
            x="8"
            y="354"
            width="264"
            height="154"
            rx="12"
            fill={p.surface.panel}
            stroke={p.edge.default}
            strokeDasharray="4 5"
          />
          <text x="140" y="386" textAnchor="middle" fill={p.text.primary} fontSize="20">
            {H.cloud}
          </text>
          <text x="140" y="414" textAnchor="middle" fill={p.text.secondary} fontSize="16">
            {H.workloads.split(' · ')[0]}
          </text>
          <text x="140" y="440" textAnchor="middle" fill={p.text.secondary} fontSize="16">
            {H.workloads.split(' · ').slice(1).join(' · ')}
          </text>
          {providers.map((provider, i) => (
            <image
              key={provider}
              href={WORKFLOW_ICONS[provider]}
              x={26 + i * 46}
              y="464"
              width="28"
              height="28"
            />
          ))}
        </>
      ) : (
        <>
          {[H.team[0], H.team[1], H.repository].map((label, i) => (
            <g key={label} transform={`translate(${16 + i * 200} 12)`}>
              <rect
                width="176"
                height="96"
                rx="12"
                fill={p.surface.panel}
                stroke={p.edge.default}
              />
              {i === 2 ? (
                <image href={WORKFLOW_ICONS.github} x="76" y="12" width="24" height="24" />
              ) : (
                <PersonMark x={88} y={24} />
              )}
              <text
                x="88"
                y="59"
                textAnchor="middle"
                fill={p.text.primary}
                fontSize="17"
                fontWeight="500"
              >
                {label}
              </text>
              <text x="88" y="82" textAnchor="middle" fill={p.text.secondary} fontSize="14">
                {H.inputDetails[i]}
              </text>
            </g>
          ))}
          <rect
            x="268"
            y="160"
            width="72"
            height="72"
            rx="16"
            fill={p.surface.raised}
            stroke={p.edge.default}
          />
          <image href={WORKFLOW_ICONS.planton} x="287" y="175" width="34" height="42" />
          <text x="304" y="267" textAnchor="middle" fill={p.text.primary} fontSize="20">
            {H.engineLabel}
          </text>
          <text x="304" y="294" textAnchor="middle" fill={p.text.secondary} fontSize="16">
            {H.controls}
          </text>
          <rect
            x="48"
            y="332"
            width="512"
            height="116"
            rx="12"
            fill={p.surface.panel}
            stroke={p.edge.default}
            strokeDasharray="4 5"
          />
          <text
            x="304"
            y="363"
            textAnchor="middle"
            fill={p.text.primary}
            fontSize="20"
            fontWeight="500"
          >
            {H.cloud}
          </text>
          <text x="304" y="390" textAnchor="middle" fill={p.text.secondary} fontSize="16">
            {H.workloads}
          </text>
          {providers.map((provider, i) => (
            <image
              key={provider}
              href={WORKFLOW_ICONS[provider]}
              x={210 + i * 40}
              y="408"
              width="26"
              height="26"
            />
          ))}
        </>
      )}
    </svg>
  );
}
