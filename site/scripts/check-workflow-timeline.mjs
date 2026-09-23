import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { WORKFLOWS } from '../src/data/workflow-explainers.ts';
import { PROVIDER_ORDER } from '../src/data/architecture-stories.ts';
import { WORKFLOW_ICONS } from '../src/data/workflow-icons.ts';
import { duration, PHASE_SECONDS, sample, nodeState } from '../src/components/marketing/workflows/timeline.ts';
import { ARCHITECTURE_LAYOUTS, ARCHITECTURE_SIZE, ARCHITECTURE_ROUTES } from '../src/components/marketing/workflows/architectureLayout.ts';
import { createRoundedRoute } from '../src/components/marketing/workflows/flowGeometry.ts';
const repo = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
assert.deepEqual(PROVIDER_ORDER, ['aws', 'gcp', 'azure', 'cloudflare', 'digitalocean']);
for (const story of Object.values(WORKFLOWS)) {
  const ids = story.nodes.map(n => n.id);
  assert.equal(new Set(ids).size, ids.length);
  for (const node of story.nodes) {
    assert.ok(node.phase >= 0 && node.phase < story.phases.length);
    if (node.icon) assert.ok(WORKFLOW_ICONS[node.icon]);
  }
  for (const edge of story.edges) {
    const source = story.nodes.find(n => n.id === edge.from), target = story.nodes.find(n => n.id === edge.to);
    assert.ok(source && target);
    assert.ok(source.phase < target.phase, `${source.id} must finish before ${target.id}`);
    if (edge.transferPhase !== undefined) assert.ok(edge.transferPhase > source.phase && edge.transferPhase < target.phase);
  }
  for (let time = 0; time < duration(story); time += 0.5) {
    const state = sample(story, time);
    for (const edge of story.edges) {
      const source = story.nodes.find(n => n.id === edge.from), target = story.nodes.find(n => n.id === edge.to);
      if (nodeState(target.phase, state.phase, state.finished) === 'active') assert.equal(nodeState(source.phase, state.phase, state.finished), 'complete');
    }
  }
  assert.equal(sample(story, -1).time, 0);
  assert.equal(sample(story, duration(story) + 10).progress, 1);
  assert.equal(duration(story) - story.phases.length * PHASE_SECONDS, 4);
  const resources = story.architecture?.resources;
  if (resources) {
    const byId = new Map(resources.map(n => [n.id, n]));
    assert.equal(byId.size, resources.length);
    const visit = (id, ancestors = []) => {
      assert.ok(byId.has(id), `known resource ${id}`);
      assert.ok(!ancestors.includes(id), `acyclic ${id}`);
      for (const dep of byId.get(id).requires) visit(dep, [...ancestors, id]);
    };
    for (const r of resources) {
      visit(r.id);
      assert.ok(ids.includes(r.group));
      const provider = r.kind.startsWith('Kubernetes') ? 'kubernetes' : story.id;
      assert.ok(fs.existsSync(path.join(repo, 'catalog', provider, r.kind.toLowerCase(), 'v1alpha1/reference.md')), `catalog reference for ${r.kind}`);
      assert.ok(WORKFLOW_ICONS[r.kind]);
    }
  }
}
for (const story of [WORKFLOWS.agents, WORKFLOWS.delivery]) {
  const gate = story.nodes.find(n => n.gate);
  assert.equal(nodeState(gate.phase + 1, gate.phase, false), 'waiting');
}
const cf = new Map(WORKFLOWS.cloudflare.architecture.resources.map(n => [n.id, n]));
assert.ok(cf.get('queue').requires.includes('worker'));
assert.ok(!cf.get('worker').requires.includes('queue'));
assert.ok(WORKFLOWS.gcp.scope.includes('custom Kubernetes workload'));
assert.ok(WORKFLOWS.aws.scope.includes('ingestion is a separate step'));
console.log('PASS: seven story timelines, human gates, provider order, catalog icons, and acyclic resource inventories.');

// Geometry acceptance guards the actual failure mode: arrows must never pass
// through a resource card, and their first/last legs must meet a card normally.
for (const provider of PROVIDER_ORDER) {
  const story = WORKFLOWS[provider], positions = ARCHITECTURE_LAYOUTS[provider];
  const { width, height } = ARCHITECTURE_SIZE;
  assert.equal(Object.keys(ARCHITECTURE_ROUTES[provider]).length, story.edges.length);
  for (const edge of story.edges) {
    const key = `${edge.from}-${edge.to}`, points = ARCHITECTURE_ROUTES[provider][key];
    assert.ok(points, `${provider}/${key}: explicit route`);
    const route = createRoundedRoute(points);
    assert.ok(Number.isFinite(route.length) && route.length > 0);
    for (let i = 0; i <= 1000; i++) {
      const [x,y] = route.at(i / 1000);
      for (const [left,top] of positions) assert.ok(!(x > left + 0.1 && x < left + width - 0.1 && y > top + 0.1 && y < top + height - 0.1), `${provider}/${key}: route intersects card`);
    }
    for (const [a,b] of [[points[0],points[1]],[points.at(-2),points.at(-1)]]) assert.ok(a[0] === b[0] || a[1] === b[1], `${provider}/${key}: perpendicular port leg`);
  }
}
console.log('PASS: authored routes avoid card interiors and preserve perpendicular port legs.');
