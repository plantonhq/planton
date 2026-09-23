/** Browser acceptance tests. All external traffic is mocked; never creates a lead or meeting. */
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';

const out=process.env.HOMEPAGE_CAPTURES??'/private/tmp/planton-homepage-review';
fs.mkdirSync(out,{recursive:true});
const server=createPreviewServer();
await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
const base=`http://127.0.0.1:${server.address().port}`;
const browser=await puppeteer.launch({headless:true});
const results=[];
const check=(name,condition)=>{assert.ok(condition,name);results.push(name);};
const sdk=fs.readFileSync('node_modules/@calcom/embed-core/dist/embed/embed.js','utf8');
let mode='success',submissions=0,lastPayload;
const page=await browser.newPage();
const errors=[];
page.on('pageerror',e=>errors.push(e.message));
await page.setRequestInterception(true);
page.on('request',request=>{
 const url=request.url();
 if(url.startsWith(base)||url.startsWith('data:')||url.startsWith('blob:'))return request.continue();
 if(url==='https://webhooks.planton.ai/demo'){
  if(request.method()==='OPTIONS')return request.respond({status:204,headers:{'access-control-allow-origin':'*','access-control-allow-methods':'POST, OPTIONS','access-control-allow-headers':'content-type'}});
  submissions++;lastPayload=JSON.parse(request.postData());
  if(mode==='network')return request.abort('failed');
  return request.respond({status:mode==='failure'?500:200,headers:{'access-control-allow-origin':'*'},contentType:'application/json',body:JSON.stringify(mode==='failure'?{error:'mock failure'}:{ok:true})});
 }
 if(url==='https://app.cal.com/embed/embed.js')return request.respond({status:200,contentType:'text/javascript',body:sdk});
 if(url.includes('cal.com'))return request.respond({status:200,contentType:'text/html',body:'<!doctype html><html><body>Mock scheduling surface</body></html>'});
 return request.respond({status:200,contentType:request.resourceType()==='script'?'text/javascript':'text/plain',body:''});
});
const events=async()=>page.evaluate(()=>Array.from(window.dataLayer??[],x=>Array.from(x)).filter(x=>x[0]==='event'&&String(x[1]).startsWith('demo_')));
try {
 for(const width of [320,390,768,1280,1680]) {
  await page.setViewport({width,height:900,deviceScaleFactor:1});
  await page.goto(base,{waitUntil:'networkidle0'});await page.evaluate(()=>document.fonts.ready);
  check(`${width}: light-only homepage`,await page.$eval('[data-homepage-appearance]',e=>e.dataset.homepageAppearance==='light') && !(await page.$('[aria-label="Light appearance"]')));
  check(`${width}: one h1`,await page.$$eval('h1',es=>es.length===1));
  check(`${width}: no overflowing content`,await page.evaluate(()=>{const w=innerWidth;return [...document.querySelectorAll('main *')].filter(e=>!(e instanceof SVGElement)&&!e.parentElement?.closest('[role=tablist]')).every(e=>{const r=e.getBoundingClientRect();return r.width===0||(r.left>=-1&&r.right<=w+1)});}));
  check(`${width}: coding-agent workflow present`,await page.$eval('#agents-title',e=>e.textContent.includes('coding agent')));
  check(`${width}: three demo CTAs`,await page.$$eval('main a[href="/book-demo"]',es=>es.length===3));
  for(const theme of ['light']) {

   if(width!==320) {
    await page.screenshot({path:path.join(out,`home-${width}-${theme}.png`),fullPage:true});
    await (await page.$('main section')).screenshot({path:path.join(out,`hero-${width}-${theme}.png`)});
   }
   check(`${width} ${theme}: readable body contrast`,await page.evaluate(()=>{
    const root=document.querySelector('[data-homepage-appearance]'),s=getComputedStyle(root);
    const rgb=v=>{const el=document.createElement('span');el.style.color=v;root.append(el);const values=getComputedStyle(el).color.match(/[\d.]+/g).slice(0,3).map(Number);el.remove();return values;};
    const lum=v=>rgb(v).map(n=>{n/=255;return n<=.04045?n/12.92:((n+.055)/1.055)**2.4}).reduce((sum,n,i)=>sum+n*[.2126,.7152,.0722][i],0);
    const a=lum(s.getPropertyValue('--hp-canvas')),b=lum(s.getPropertyValue('--hp-secondary'));return (Math.max(a,b)+.05)/(Math.min(a,b)+.05)>=4.5;
   }));
  }
 }
 await page.setViewport({width:1280,height:900});await page.goto(base,{waitUntil:'networkidle0'});
 await page.focus('header button[aria-haspopup="true"]');await page.keyboard.press('Enter');
 await page.waitForSelector('.MuiPopover-paper');await page.waitForFunction(()=>getComputedStyle(document.querySelector('.MuiPopover-paper')).opacity==='1');
 check('light desktop menu portal',await page.$eval('.MuiPopover-paper',e=>getComputedStyle(e).backgroundColor==='rgb(227, 227, 224)'));
 await page.screenshot({path:path.join(out,'navigation-1280-light.png')});await page.keyboard.press('Escape');await page.waitForSelector('.MuiPopover-root',{hidden:true});
 await page.setViewport({width:390,height:844});await page.click('[aria-label="Open navigation"]');await page.waitForSelector('.MuiDrawer-paper');await page.waitForFunction(()=>getComputedStyle(document.querySelector('.MuiDrawer-paper')).transform==='none');
 check('light mobile drawer portal',await page.$eval('.MuiDrawer-paper',e=>getComputedStyle(e).backgroundColor==='rgb(238, 238, 235)'));
 await page.screenshot({path:path.join(out,'navigation-390-light.png')});await page.click('[aria-label="Close navigation"]');await page.waitForSelector('.MuiDrawer-root',{hidden:true});await page.setViewport({width:1280,height:900});
 await page.focus('main button[aria-controls]');await page.keyboard.press('Tab');await page.keyboard.press('Enter');
 check('keyboard selects review layer',await page.$eval('main button[aria-controls][aria-pressed="true"]',e=>e.textContent.includes('Review')));
 check('review explanation updates',await page.$eval('[aria-live="polite"]',e=>e.textContent.includes('Understand the change')));
 await page.click('a[href="#how-it-works"]');check('explanation anchor works',page.url().endsWith('#how-it-works'));
 await page.focus('details summary');await page.keyboard.press('Enter');check('FAQ keyboard expands',await page.$eval('details',e=>e.open));
 await page.emulateMediaFeatures([{name:'prefers-reduced-motion',value:'reduce'}]);
 check('reduced motion disables transitions',await page.$eval('main button[aria-controls]',e=>getComputedStyle(e).transitionDuration==='0s'));
 // All three CTA locations must report their own placement.
 for(const location of ['hero','controls','close']) {
  await page.goto(base,{waitUntil:'networkidle0'});
  const index=['hero','controls','close'].indexOf(location);
  await page.evaluate(i=>{document.querySelectorAll('main a[href="/book-demo"]')[i].addEventListener('click',e=>e.preventDefault());},index);
  await (await page.$$('main a[href="/book-demo"]'))[index].click();
  check(`${location}: CTA attribution`,(await events()).some(e=>e[1]==='demo_cta_click'&&e[2].location===location));
 }
 await page.goto(base+'/book-demo',{waitUntil:'networkidle0'});
 check('demo uses light surface',await page.$eval('[data-marketing-appearance]',e=>e.dataset.marketingAppearance==='light'));
 await page.screenshot({path:path.join(out,'demo-1280.png'),fullPage:true});
 await page.click('button[type="submit"]');
 check('empty form makes no request',submissions===0);
 check('all required fields show errors',await page.$$eval('[aria-invalid="true"]',es=>es.length===6));
 await page.type('#firstName','Sample');await page.type('#lastName','Visitor');await page.type('#workEmail','visitor@example.invalid');await page.type('#company','Example');await page.select('#jobTitle','CTO');await page.select('#companySize','1-10');
 mode='failure';await page.click('button[type="submit"]');await page.waitForSelector('[role="alert"]');
 check('server failure retains values',await page.$eval('#workEmail',e=>e.value==='visitor@example.invalid'));
 check('failure does not show scheduler',!(await page.$('iframe')));
 mode='network';await page.click('button[type="submit"]');await page.waitForFunction(()=>document.querySelector('[role="alert"]')?.textContent.includes('submit your request'));
 check('network failure is recoverable',await page.$eval('button[type="submit"]',e=>!e.disabled));
 mode='success';await page.click('button[type="submit"]');await page.waitForSelector('iframe');
 await page.waitForFunction(()=>document.querySelector('h1').textContent.includes('Choose a time'));
 check('payload schema preserved',Object.keys(lastPayload).sort().join(',')==='company,companySize,firstName,jobTitle,lastName,workEmail');
 check('calendar light theme',await page.$eval('iframe',e=>e.src.includes('theme=light')));
 check('calendar prefilled',await page.$eval('iframe',e=>e.src.includes('visitor%40example.invalid')&&e.src.includes('Sample')));
 check('focus moves to scheduler heading',await page.$eval('h1',e=>e===document.activeElement));
 check('form success is not a booking',!(await events()).some(e=>e[1]==='demo_booking_confirmed'));
 await page.screenshot({path:path.join(out,'demo-scheduler-1280.png'),fullPage:true});
 const fire=async(status,paymentRequired=false)=>page.evaluate((status,paymentRequired)=>window.dispatchEvent(new CustomEvent('CAL:60min:bookingSuccessfulV2',{detail:{data:{status,paymentRequired,uid:'fixture-only'}}})),status,paymentRequired);
 await fire('PENDING');await fire('ACCEPTED',true);
 check('pending/unpaid bookings not counted',!(await events()).some(e=>e[1]==='demo_booking_confirmed'));
 await fire('ACCEPTED');await fire('ACCEPTED');
 check('confirmed booking counted once',(await events()).filter(e=>e[1]==='demo_booking_confirmed').length===1);
 check('confirmation shown',await page.evaluate(()=>document.body.textContent.includes('Your meeting is booked.')));
 const analytics=await events();
 for(const event of ['demo_form_start','demo_form_submit_success','demo_scheduler_view'])check(`${event}: emitted once`,analytics.filter(e=>e[1]===event).length===1);
 check('no personal information in analytics',!JSON.stringify(analytics).match(/Visitor|Sample|example.invalid|fixture-only/));
 await page.evaluate(()=>window.dispatchEvent(new CustomEvent('CAL:60min:linkFailed',{detail:{data:{}}})));
 await page.waitForSelector('p[role="alert"]');check('calendar failure offers fallback',await page.$eval('p[role="alert"] a',e=>e.href==='https://cal.com/swarup-donepudi/60min'));
 await page.setViewport({width:390,height:844});await page.goto(base+'/book-demo',{waitUntil:'networkidle0'});await page.screenshot({path:path.join(out,'demo-390.png'),fullPage:true});
 for (const route of ['/docs/coding-agents','/product','/pricing']) {
  await page.goto(base+route,{waitUntil:'networkidle0'});
  check(`${route}: existing dark surface preserved`,await page.$eval('[data-marketing-appearance]',e=>e.dataset.marketingAppearance==='dark'));
 }
 check('no application JavaScript errors',errors.length===0);
 fs.writeFileSync(path.join(out,'acceptance.json'),JSON.stringify({passed:results.length,checks:results,submissions,externalTraffic:'mocked',errors},null,2));
 console.log(`PASS: ${results.length} checks; captures: ${out}; ${submissions} mocked requests, zero real leads or bookings.`);
} finally {await browser.close();await new Promise(resolve=>server.close(resolve));}
