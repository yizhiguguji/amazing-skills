const {chromium}=require('playwright');
const http=require('node:http'),fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict');
const root=__dirname,phase=process.argv.includes('--revision')?'revision':'baseline';
const sourceRoot=phase==='baseline'?path.join(root,'baseline'):root;
const publicFiles=new Set(['/','/index.html','/styles.css','/app.js']);
const server=http.createServer((req,res)=>{const name=req.url.split('?')[0];if(!publicFiles.has(name)){res.writeHead(404);return res.end('Not found')}const p=path.join(sourceRoot,name==='/'?'index.html':name.slice(1));res.setHeader('Content-Type',p.endsWith('.css')?'text/css':p.endsWith('.js')?'text/javascript':'text/html; charset=utf-8');res.end(fs.readFileSync(p))});
let browser;const results=[],errors=[];
async function scenario(id,fn){try{await fn();results.push({id,status:'passed'})}catch(e){results.push({id,status:'failed',error:e.message});throw e}}
(async()=>{
await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));const url='http://127.0.0.1:'+server.address().port;
browser=await chromium.launch({headless:true,...(process.env.CHROME_PATH?{executablePath:process.env.CHROME_PATH}:{})});
const context=await browser.newContext({viewport:{width:1440,height:1000},timezoneId:'Asia/Tokyo'});const page=await context.newPage();page.on('pageerror',e=>errors.push(e.message));await page.goto(url);
const dates=await page.evaluate(()=>[-1,0,7].map(n=>{const d=new Date();d.setDate(d.getDate()+n);return`${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`}));
async function add(title,due){await page.locator('#add').click();await page.locator('#project-name').fill(title);await page.locator('#project-due').fill(due);await page.locator('#create-save').click()}
async function nav(view){await page.locator(`nav [data-view="${view}"]`).click()}
const shot=async name=>{await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));return page.screenshot({path:path.join(root,'evidence',phase+'-'+name+'.png'),fullPage:true,animations:'disabled'})};
await scenario('AC-01',async()=>{assert.equal(await page.locator('.project-card').count(),0);assert.match(await page.locator('#hook').textContent(),/先录一个项目/);await shot('empty')});
await scenario('AC-02',async()=>{
await page.locator('#add').focus();await page.keyboard.press('Enter');assert.equal(await page.locator('#project-name').evaluate(el=>el===document.activeElement),true);
await page.locator('#create-save').click();assert.equal(await page.locator('#create-dialog').isVisible(),true);await page.keyboard.press('Escape');assert.equal(await page.locator('#add').evaluate(el=>el===document.activeElement),true);
await add('作品集改版',dates[0]);await add('咖啡馆菜单',dates[1]);await add('品牌介绍页',dates[2]);assert.match(await page.locator('#hook').textContent(),/3 个待推进 · 1 个超期/);
await page.getByRole('button',{name:'打开项目：作品集改版',exact:true}).click();await page.locator('#detail-stage').selectOption('done');await page.locator('#detail-note').fill('交付完成，已确认。');await page.locator('#detail-save').click();assert.match(await page.locator('#hook').textContent(),/2 个待推进 · 0 个超期/);
await page.reload();assert.match(await page.locator('#hook').textContent(),/2 个待推进 · 0 个超期/);await page.getByRole('button',{name:'打开项目：作品集改版',exact:true}).click();assert.match(await page.locator('#detail-history').textContent(),/交付完成/);await page.keyboard.press('Escape');
});
await scenario('AC-03',async()=>{
await nav('projects');await page.locator('#search').fill('菜单');assert.equal(await page.locator('#project-list .list-row').count(),1);await page.locator('#search').fill('没有');assert.equal(await page.locator('#empty').isVisible(),true);
await nav('activity');assert.match(await page.locator('#activity-list').textContent(),/交付完成/);
await nav('review');assert.deepEqual(await page.locator('.chart-row strong').allTextContents(),['2','0','0','1']);
await nav('settings');await page.locator('#owner-name').fill('示例用户');await page.getByRole('button',{name:'保存称呼',exact:true}).click();assert.match(await page.locator('#hook').textContent(),/示例用户/);await page.reload();assert.match(await page.locator('#hook').textContent(),/示例用户/);
});
await scenario('AC-04',async()=>{
await add('一位名字很长的项目用于验证中文布局和断行'.repeat(4),dates[2]);
await page.evaluate(()=>{window.realSet=Storage.prototype.setItem;Storage.prototype.setItem=function(){throw new DOMException('quota','QuotaExceededError')}});await add('暂存项目',dates[2]);assert.match(await page.locator('#storage-message').textContent(),/未保存/);await page.evaluate(()=>Storage.prototype.setItem=window.realSet);await page.locator('#retry').click();assert.equal(await page.locator('#storage-error').isVisible(),false);
for(const width of [1440,390]){await page.setViewportSize({width,height:1000});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);await shot(width===1440?'desktop':'mobile');await page.locator('#add').click();assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);await page.keyboard.press('Escape')}
await page.emulateMedia({reducedMotion:'reduce'});assert.equal(await page.locator('#add').evaluate(el=>getComputedStyle(el).transitionDuration),'0s');
const broken=await browser.newContext();await broken.addInitScript(()=>localStorage.setItem('amazing-project-workbench-v1','broken'));const bp=await broken.newPage();await bp.goto(url);assert.equal(await bp.locator('#add').isDisabled(),true);assert.equal(await bp.evaluate(()=>localStorage.getItem('amazing-project-workbench-v1')),'broken');await broken.close();
});
await page.setViewportSize({width:1440,height:1000});
await page.mouse.move(0,0);await page.locator('#add').evaluate(el=>el.blur());
const styles=await page.evaluate(()=>{const button=getComputedStyle(document.getElementById('add')),body=getComputedStyle(document.body),root=getComputedStyle(document.documentElement);return{font:body.fontFamily,button:{color:button.color,background:button.backgroundColor,borderRadius:button.borderRadius,padding:button.padding},tokens:Object.fromEntries(['--canvas','--surface','--ink','--accent','--radius','--font'].map(k=>[k,root.getPropertyValue(k).trim()]))}});
if(phase==='baseline'){fs.writeFileSync(path.join(root,'evidence/baseline-styles.json'),JSON.stringify(styles,null,2)+'\n');assert.equal(await page.locator('#due-filter').count(),0)}
if(phase==='revision')await scenario('AC-05',async()=>{
const old=JSON.parse(fs.readFileSync(path.join(root,'evidence/baseline-styles.json'),'utf8'));assert.deepEqual(styles,old);
await page.locator('#due-filter').selectOption('today');assert.equal(await page.locator('.project-card').count(),1);assert.match(await page.locator('.project-card').textContent(),/咖啡馆/);assert.match(await page.locator('#hook').textContent(),/4 个待推进/);
await nav('projects');assert.equal(await page.locator('#project-list .list-row').count(),5);assert.equal(await page.locator('#due-label').isVisible(),false);await nav('board');await page.locator('#due-filter').selectOption('overdue');assert.equal(await page.locator('.project-card').count(),0);assert.match(await page.locator('#empty').textContent(),/没有符合交付日期/);await page.locator('#due-filter').selectOption('all');assert.equal(await page.locator('.project-card').count(),5);
await page.setViewportSize({width:390,height:1000});assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);await shot('mobile-filter');await page.setViewportSize({width:1440,height:1000});await shot('desktop-filter');
});
assert.deepEqual(errors,[]);console.log(JSON.stringify({phase,status:'passed',scenarios:results.length,errors}));
})().catch(error=>{console.error(error);process.exitCode=1}).finally(async()=>{fs.writeFileSync(path.join(root,'evidence',phase+'-results.json'),JSON.stringify({phase,at:new Date().toISOString(),status:process.exitCode?'failed':'passed',results,errors,environment:{timezone:'Asia/Tokyo',browser:process.env.CHROME_PATH?'Installed Chrome via Playwright':'Playwright Chromium',viewportWidths:[1440,390]}},null,2)+'\n');if(browser)await browser.close();await new Promise(resolve=>server.close(resolve))});
