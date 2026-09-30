// The CloudflareWorker r2-bundle scenario's script, placed once in the owner-arranged
// bundle bucket (see e2e/scenarios/r2-bundle.yaml). An ES module, so the modules'
// main_module default ("index.js") applies.
export default {
  async fetch() {
    return new Response("served from an r2 bundle");
  },
};
