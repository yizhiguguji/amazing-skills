(() => {
  'use strict';
  const KEY = 'amazing-skills-customers-v1';
  const $ = id => document.getElementById(id);
  const today = () => { const d = new Date(); return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`; };
  const validDate = value => { if (typeof value !== 'string' || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return false; const d=new Date(value+'T12:00:00'); return Number.isFinite(d.getTime()) && d.getFullYear()===Number(value.slice(0,4)) && d.getMonth()+1===Number(value.slice(5,7)) && d.getDate()===Number(value.slice(8,10)); };
  let customers = [], view = 'due', editing = null, corrupt = false;
  const showError = message => { $('storage-error').hidden=false; $('storage-message').textContent=message; $('retry').hidden=corrupt; };
  try {
    const raw = localStorage.getItem(KEY);
    if(raw !== null){
      const data=JSON.parse(raw);
      if(!Array.isArray(data) || !data.every(c=>c && typeof c.id==='string' && typeof c.name==='string' && c.name.trim() && validDate(c.nextDate) && Array.isArray(c.records) && c.records.every(r=>r && typeof r.text==='string' && typeof r.at==='string')) || new Set(data.map(c=>c.id)).size!==data.length) throw new Error('Invalid data');
      customers=data;
    }
  } catch(error){corrupt=true;showError('本地数据无法读取。为保护原记录，已停止写入。请先备份或检查浏览器站点数据。');}
  const save = () => {
    if(corrupt) return false;
    try{localStorage.setItem(KEY,JSON.stringify(customers));$('storage-error').hidden=true;return true;}
    catch(error){showError('本次修改未保存，当前页面仍可查看。请重试保存，离开页面前确认成功。');return false;}
  };
  const element = (tag, text, className) => { const e=document.createElement(tag); if(text!==undefined)e.textContent=text; if(className)e.className=className;return e; };
  function render(){
    const day=today(), overdue=customers.filter(c=>c.nextDate<day).length, due=customers.filter(c=>c.nextDate===day).length;
    $('today').textContent=new Intl.DateTimeFormat('zh-CN',{month:'long',day:'numeric',weekday:'long'}).format(new Date())+' · 下一步，从这里开始';
    $('summary').hidden=customers.length===0; $('overdue').textContent=overdue; $('today-count').textContent=due; $('total').textContent=customers.length;
    $('due-count').textContent=overdue+due; $('all-count').textContent=customers.length;
    $('due-view').setAttribute('aria-pressed',view==='due');$('all-view').setAttribute('aria-pressed',view==='all');
    $('list-heading').textContent=view==='due'?'待跟进客户':'全部客户';
    const query=$('search').value.trim().toLocaleLowerCase();
    const visible=customers.filter(c=>(view==='all'||c.nextDate<=day)&&c.name.toLocaleLowerCase().includes(query)).sort((a,b)=>a.nextDate.localeCompare(b.nextDate));
    $('customers').replaceChildren();$('empty').hidden=visible.length>0;
    $('empty-add').hidden=customers.length>0 || corrupt;
    $('empty-title').textContent=corrupt?'先恢复本地数据':!customers.length?'从下一次联系开始':query?'没有找到这位客户':'今天的跟进已安排妥当';
    $('empty-text').textContent=corrupt?'检查并备份站点数据后，再重新打开工作台。':!customers.length?'先添加一位客户，记下下一次联系日期。':query?'试试其他姓名，或清除搜索。':'可以到全部客户查看后续联系日期。';
    $('add').disabled=corrupt;
    for(const c of visible){
      const li=element('li',undefined,'customer');li.dataset.id=c.id;
      const info=element('div',undefined,'customer-info');info.append(element('h3',c.name));
      const label=c.nextDate<day?'已超期':c.nextDate===day?'今天联系':'下次联系';
      info.append(element('p',label+' · '+c.nextDate,c.nextDate<day?'overdue':''));
      if(c.records.length){
        info.append(element('p','最近跟进：'+c.records[c.records.length-1].text));
        const details=element('details');details.append(element('summary',`${c.records.length} 条跟进记录`));
        for(const record of [...c.records].reverse())details.append(element('p',new Date(record.at).toLocaleString('zh-CN')+' · '+record.text));
        info.append(details);
      }else info.append(element('p','还没有跟进记录'));
      const button=element('button','记录跟进');button.setAttribute('aria-label','记录跟进：'+c.name);button.addEventListener('click',()=>open(c));li.append(info,button);$('customers').append(li);
    }
  }
  function open(customer=null){
    if(corrupt)return;
    editing=customer;$('customer-form').reset();$('name-label').hidden=!!customer;$('customer-name').required=!customer;
    $('record-label').hidden=!customer;$('record-text').required=!!customer;
    $('form-title').textContent=customer?'记录跟进':'添加客户';$('form-description').textContent=customer?customer.name:'先留下一个姓名和下一次联系日期。';
    $('save').textContent=customer?'保存跟进':'保存客户';$('next-date').value=customer?customer.nextDate:today();
    $('customer-dialog').showModal();(customer?$('record-text'):$('customer-name')).focus();
  }
  $('customer-form').addEventListener('submit', event=>{
    event.preventDefault();const name=$('customer-name').value.trim(),text=$('record-text').value.trim(),date=$('next-date').value;
    if(!validDate(date))return;
    if(editing){if(!text){$('record-text').setCustomValidity('请填写本次跟进内容');$('record-text').reportValidity();return;}editing.records.push({text,at:new Date().toISOString()});editing.nextDate=date;}
    else{if(!name){$('customer-name').setCustomValidity('请填写客户姓名');$('customer-name').reportValidity();return;}customers.push({id:crypto.randomUUID(),name,nextDate:date,records:[]});view='all';}
    save();$('customer-dialog').close();render();
  });
  for(const id of ['customer-name','record-text'])$(id).addEventListener('input',()=>$(id).setCustomValidity(''));
  $('add').addEventListener('click',()=>open());$('empty-add').addEventListener('click',()=>open());
  for(const id of ['cancel','close-dialog'])$(id).addEventListener('click',()=>$('customer-dialog').close());
  $('due-view').addEventListener('click',()=>{view='due';render();});$('all-view').addEventListener('click',()=>{view='all';render();});
  $('search').addEventListener('input',render);$('retry').addEventListener('click',save);
  // Recompute dates when returning after midnight, without reloading storage over unsaved edits.
  document.addEventListener('visibilitychange',()=>{if(!document.hidden)render();});
  render();
})();
