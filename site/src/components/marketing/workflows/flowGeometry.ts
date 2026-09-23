export type Point = readonly [number, number];
type Segment = { to: Point; controls?: readonly [Point, Point] };

/** One sampled path powers SVG, browser packets, and offline video. Sampling
 * happens at layout time, so motion remains independent of segment lengths. */
function pathFrom(start: Point, segments: readonly Segment[]) {
  const samples = [{ position: start, distance: 0 }];
  let a = start, path = `M${start}`;
  for (const segment of segments) {
    const b = segment.to, controls = segment.controls;
    path += controls ? ` C${controls[0]} ${controls[1]} ${b}` : ` L${b}`;
    const steps = controls ? 80 : 1;
    for (let i = 1; i <= steps; i++) {
      const t = i / steps;
      const coordinate = (axis: 0 | 1) => controls
        ? (1-t)**3*a[axis] + 3*(1-t)**2*t*controls[0][axis] + 3*(1-t)*t*t*controls[1][axis] + t**3*b[axis]
        : a[axis] + (b[axis] - a[axis]) * t;
      const position: Point = [coordinate(0), coordinate(1)];
      const previous = samples[samples.length - 1];
      samples.push({ position, distance: previous.distance + Math.hypot(position[0] - previous.position[0], position[1] - previous.position[1]) });
    }
    a = b;
  }
  const length = samples[samples.length - 1].distance;
  return { path, length, at: (fraction: number): Point => {
    const distance = Math.max(0, Math.min(1, fraction)) * length;
    const end = samples.findIndex(p => p.distance >= distance);
    if (end <= 0) return start;
    const left = samples[end - 1], right = samples[end];
    const t = (distance - left.distance) / (right.distance - left.distance);
    return [left.position[0] + t * (right.position[0] - left.position[0]), left.position[1] + t * (right.position[1] - left.position[1])];
  } };
}
export function createFlowRoute(a: Point, c: Point, d: Point, b: Point) {
  return pathFrom(a, [{ controls: [c, d], to: b }]);
}

/** Authored orthogonal lanes with rounded, tangent-continuous corners. Short
 * lanes reduce their radius rather than doubling back or overshooting a port. */
export function createRoundedRoute(points: readonly Point[], radius = 18) {
  const segments: Segment[] = [];
  for (let i = 1; i < points.length - 1; i++) {
    const prev = points[i - 1], corner = points[i], next = points[i + 1];
    const before = Math.hypot(corner[0] - prev[0], corner[1] - prev[1]);
    const after = Math.hypot(next[0] - corner[0], next[1] - corner[1]);
    const r = Math.min(radius, before / 2, after / 2);
    const entry: Point = [corner[0] + (prev[0] - corner[0]) * r / before, corner[1] + (prev[1] - corner[1]) * r / before];
    const exit: Point = [corner[0] + (next[0] - corner[0]) * r / after, corner[1] + (next[1] - corner[1]) * r / after];
    // A quadratic corner expressed as a cubic keeps one sampling primitive.
    const c: Point = [entry[0] + (corner[0] - entry[0]) * 2/3, entry[1] + (corner[1] - entry[1]) * 2/3];
    const d: Point = [exit[0] + (corner[0] - exit[0]) * 2/3, exit[1] + (corner[1] - exit[1]) * 2/3];
    segments.push({ to: entry }, { controls: [c, d], to: exit });
  }
  segments.push({ to: points[points.length - 1] });
  return pathFrom(points[0], segments);
}
