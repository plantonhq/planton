/** Reproducible social card from the actual editable homepage, not a separate design. */
import fs from 'node:fs';
import puppeteer from 'puppeteer';
import { createPreviewServer } from './preview-homepage.mjs';
const server=createPreviewServer();
await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
const browser=await puppeteer.launch({headless:true});
try {
 const page=await browser.newPage();await page.setViewport({width:1200,height:630,deviceScaleFactor:1});
 await page.goto(`http://127.0.0.1:${server.address().port}`,{waitUntil:'networkidle0'});await page.evaluate(()=>document.fonts.ready);
 await page.addStyleTag({content:`body>div>header, header, body>div>footer, footer, [data-homepage-appearance]>button {display:none!important} main section:not(:first-child){display:none!important} main section:first-child{padding-block:25px!important;min-height:630px!important;height:630px!important;align-items:center!important} [class*="layerDetail"], [class*="artCaption"], [class*="heroCaption"]{display:none!important} [class*="heroCopy"]{padding-bottom:0!important} [class*="artboard"]{margin:0!important} h1{font-size:62px!important} [class*="intro"]{font-size:16px!important} [class*="eyebrow"]{margin-bottom:24px!important}`});
 await page.evaluate(()=>{window.scrollTo(0,0);const label=document.querySelector('[class*="eyebrow"]');if(label)label.textContent='PLANTON / THE SELF-SERVICE CLOUD PLATFORM';});
 fs.mkdirSync('public/_site/images/og',{recursive:true});
 await page.screenshot({path:'public/_site/images/og/homepage.png'});
 console.log('Wrote public/_site/images/og/homepage.png (1200 × 630)');
} finally {await browser.close();await new Promise(resolve=>server.close(resolve));}
