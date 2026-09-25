/** Local static-export preview, including extensionless Next routes and assetPrefix. */
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const site = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
export function createPreviewServer() {
  const root = path.join(site, 'out');
  const types = { '.html':'text/html', '.js':'text/javascript', '.css':'text/css', '.png':'image/png', '.jpeg':'image/jpeg', '.webp':'image/webp', '.vtt':'text/vtt', '.svg':'image/svg+xml', '.woff2':'font/woff2', '.json':'application/json' };
  return http.createServer((req,res) => {
    let requested;
    try { requested = decodeURIComponent(new URL(req.url, 'http://localhost').pathname); } catch { res.writeHead(400).end(); return; }
    if (requested.startsWith('/_site/_next/')) requested=requested.replace('/_site/_next/', '/_next/');
    const base=path.resolve(root, '.'+requested);
    if (!base.startsWith(root+path.sep) && base!==root) { res.writeHead(403).end(); return; }
    const file=[base,base+'.html',path.join(base,'index.html')].find(p=>fs.existsSync(p)&&fs.statSync(p).isFile());
    if (!file) { res.writeHead(404).end(); return; }
    res.writeHead(200,{'Content-Type':types[path.extname(file)]??'application/octet-stream'});
    fs.createReadStream(file).pipe(res);
  });
}
if (process.argv[1] && path.resolve(process.argv[1])===fileURLToPath(import.meta.url)) {
  const server=createPreviewServer();
  server.listen(Number(process.env.PORT??4177),'127.0.0.1',()=>console.log(`Planton preview: http://127.0.0.1:${server.address().port}`));
}
