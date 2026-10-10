const { chromium } = require('playwright');
const http = require('node:http');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const dir = __dirname;
const url = 'http://127.0.0.1:8877';
const assets=new Set(['/','/index.html','/styles.css','/app.js']);
const server=http.createServer((req,res)=>{const name=req.url.split('?')[0];if(!assets.has(name)){res.writeHead(404);return res.end('Not found')};const file=path.join(dir,name==='/'?'index.html':name.slice(1));res.setHeader('Content-Type',file.endsWith('.css')?'text/css':file.endsWith('.js')?'text/javascript':'text/html; charset=utf-8');res.end(fs.readFileSync(file));});
let browser;
const results=[];
async function scenario(id,fn){
  try{await fn();results.push({id,status:'passed'});}
  catch(error){results.push({id,status:'failed',error:error.message});throw error;}
}
(async()=>{
  await new Promise((resolve,reject)=>{server.once('error',reject);server.listen(8877,'127.0.0.1',resolve);});
  browser=await chromium.launch({headless:true, ...(process.env.CHROME_PATH ? {executablePath:process.env.CHROME_PATH} : {})});
  const context=await browser.newContext({viewport:{width:1440,height:1000},timezoneId:'Asia/Tokyo'});
  const page=await context.newPage();const errors=[];page.on('pageerror',e=>errors.push(e.message));
  await page.goto(url);
  const dates=await page.evaluate(()=>[-1,0,7].map(delta=>{const d=new Date();d.setDate(d.getDate()+delta);return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;}));
  async function add(name,date){
    await page.locator('#add').click();await page.locator('#customer-name').fill(name);await page.locator('#next-date').fill(date);await page.locator('#save').click();
  }
  await scenario('AC-01',async()=>{
    assert.equal(await page.locator('.customer').count(),0);
    assert.equal(await page.locator('#summary').isVisible(),false);
    assert.match(await page.locator('#empty-text').textContent(),/先添加一位客户/);
    assert.equal(await page.locator('#empty-add').isVisible(),true);
    await page.screenshot({path:path.join(dir,'evidence/desktop-empty.png'),fullPage:true});
  });
  await scenario('AC-02',async()=>{
    await page.locator('#add').focus();await page.keyboard.press('Enter');
    assert.equal(await page.locator('#customer-name').evaluate(el=>el===document.activeElement),true);
    await page.locator('#save').click();assert.equal(await page.locator('#customer-dialog').isVisible(),true);
    await page.locator('#customer-name').fill('   ');await page.locator('#save').click();
    assert.equal(await page.locator('#customer-name').evaluate(el=>el.validity.valid),false);
    await page.locator('#customer-name').fill('示例客户甲');await page.locator('#next-date').fill(dates[0]);await page.locator('#save').click();
    assert.equal(await page.locator('.customer').count(),1);
    await page.locator('#add').click();await page.keyboard.press('Escape');assert.equal(await page.locator('#customer-dialog').isVisible(),false);
    assert.equal(await page.locator('#add').evaluate(el=>el===document.activeElement),true);
  });
  await scenario('AC-03',async()=>{
    await add('示例客户乙',dates[1]);await add('示例客户丙',dates[2]);
    assert.equal(await page.locator('#overdue').textContent(),'1');assert.equal(await page.locator('#today-count').textContent(),'1');assert.equal(await page.locator('#total').textContent(),'3');
    await page.locator('#due-view').click();assert.equal(await page.locator('.customer').count(),2);
    assert.match(await page.locator('.customer h3').first().textContent(),/示例客户甲/);
    await page.screenshot({path:path.join(dir,'evidence/desktop-populated.png'),fullPage:true});
  });
  await scenario('AC-04',async()=>{
    await page.getByRole('button',{name:'记录跟进：示例客户甲',exact:true}).click();
    await page.locator('#record-text').fill('已发送方案，约定下周讨论。');await page.locator('#next-date').fill(dates[2]);await page.locator('#save').click();
    assert.equal(await page.locator('#overdue').textContent(),'0');await page.reload();await page.locator('#all-view').click();
    const row=page.locator('.customer').filter({has:page.getByRole('heading',{name:'示例客户甲'})});
    assert.match(await row.textContent(),/已发送方案/);assert.match(await row.textContent(),new RegExp(dates[2]));
    await row.locator('summary').click();assert.equal(await row.locator('details').getAttribute('open'),'');
  });
  await scenario('AC-05',async()=>{
    await page.locator('#search').fill('示例客户乙');assert.equal(await page.locator('.customer').count(),1);
    await page.locator('#search').fill('不存在的客户');assert.equal(await page.locator('.customer').count(),0);
    assert.match(await page.locator('#empty-title').textContent(),/没有找到/);await page.locator('#search').fill('');
    await add('<img src=x onerror=alert(1)>',dates[2]);assert.equal(await page.locator('.customer img').count(),0);
  });
  await scenario('AC-06',async()=>{
    await add('这是一位名称很长的客户用于验证中文换行和窄屏布局'.repeat(3),dates[2]);
    for(const width of [1440,390]){
      await page.setViewportSize({width,height:1000});
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
      await page.locator('#add').click();
      assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
      if(width===390)await page.screenshot({path:path.join(dir,'evidence/mobile-form.png'),fullPage:true});
      await page.keyboard.press('Escape');
      if(width===390)await page.screenshot({path:path.join(dir,'evidence/mobile-populated.png'),fullPage:true});
    }
    await page.emulateMedia({reducedMotion:'reduce'});
    assert.equal(await page.locator('#add').evaluate(el=>getComputedStyle(el).transitionDuration),'0s');
    await page.locator('#add').focus();await page.keyboard.press('Enter');await page.keyboard.press('Escape');
  });
  await scenario('AC-07',async()=>{
    await page.evaluate(()=>{window.realSetItem=Storage.prototype.setItem;Storage.prototype.setItem=function(){throw new DOMException('quota','QuotaExceededError')};});
    await add('保存失败测试',dates[2]);assert.equal(await page.locator('#storage-error').isVisible(),true);
    assert.match(await page.locator('#storage-message').textContent(),/未保存/);
    assert.equal(await page.locator('.customer').count(),6);
    await page.evaluate(()=>Storage.prototype.setItem=window.realSetItem);await page.locator('#retry').click();
    assert.equal(await page.locator('#storage-error').isVisible(),false);await page.reload();await page.locator('#all-view').click();assert.equal(await page.locator('.customer').count(),6);
    const broken=await browser.newContext();await broken.addInitScript(()=>localStorage.setItem('amazing-skills-customers-v1','broken-json'));
    const bp=await broken.newPage();await bp.goto(url);assert.equal(await bp.locator('#add').isDisabled(),true);
    assert.match(await bp.locator('#storage-message').textContent(),/停止写入/);
    assert.equal(await bp.evaluate(()=>localStorage.getItem('amazing-skills-customers-v1')),'broken-json');await broken.close();
  });
  assert.deepEqual(errors,[]);
  console.log(JSON.stringify({status:'passed',scenarios:results.length,uncaughtErrors:errors}));
})().catch(error=>{console.error(error);process.exitCode=1;}).finally(async()=>{
  fs.mkdirSync(path.join(dir,'evidence'),{recursive:true});
  fs.writeFileSync(path.join(dir,'evidence/browser-results.json'),JSON.stringify({run_at:new Date().toISOString(),status:results.length===7 && results.every(r=>r.status==='passed')?'passed':'failed',environment:{browser:process.env.CHROME_PATH?'Installed Chrome via Playwright, headless':'Playwright Chromium headless',timezone:'Asia/Tokyo',viewportWidths:[1440,390]},results},null,2)+'\n');
  if(browser)await browser.close();await new Promise(resolve=>server.close(resolve));
});
