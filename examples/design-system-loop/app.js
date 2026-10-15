'use strict';
const key='amazing-reading-continuation-v1';
const form=document.querySelector('#book-form'),title=document.querySelector('#book-name'),next=document.querySelector('#next-step'),error=document.querySelector('#title-error'),status=document.querySelector('#save-status'),storageError=document.querySelector('#storage-error');
let books=[],storageBroken=false;
try{const raw=localStorage.getItem(key);if(raw!==null){const parsed=JSON.parse(raw);if(!Array.isArray(parsed)||parsed.some(book=>!book||typeof book.id!=='string'||typeof book.title!=='string'||typeof book.next!=='string'))throw new Error('Invalid data');books=parsed;}}catch{storageBroken=true;storageError.hidden=false;}
function render(){
 const list=document.querySelector('#books');list.replaceChildren();
 document.querySelector('#empty').hidden=books.length>0;
 const count=document.querySelector('#record-count');count.hidden=books.length===0;count.textContent=books.length+' 本';
 for(const book of books){
  const li=document.createElement('li'),heading=document.createElement('h3'),note=document.createElement('p'),remove=document.createElement('button');
  heading.textContent=book.title;note.textContent=book.next||'还没写下一步，下次从这本书开始。';
  remove.type='button';remove.textContent='移除记录';remove.setAttribute('aria-label','移除记录：'+book.title);
  remove.addEventListener('click',()=>{if(save(books.filter(item=>item.id!==book.id))){status.textContent='已移除：'+book.title;title.focus();}});
  li.append(heading,note,remove);list.append(li);
 }
}
function save(candidate){
 if(storageBroken){storageError.hidden=false;return false;}
 try{localStorage.setItem(key,JSON.stringify(candidate));books=candidate;storageError.hidden=true;render();return true;}catch{storageError.hidden=false;return false;}
}
title.addEventListener('input',()=>{if(title.value.trim()){error.textContent='';title.setAttribute('aria-invalid','false');}});
form.addEventListener('submit',event=>{
 event.preventDefault();status.textContent='';
 if(!title.value.trim()){error.textContent='请填写书名，再保存阅读起点。';title.setAttribute('aria-invalid','true');title.focus();return;}
 const record={id:crypto.randomUUID(),title:title.value.trim(),next:next.value.trim()};
 if(save([record,...books])){form.reset();error.textContent='';title.setAttribute('aria-invalid','false');status.textContent='已保存：'+record.title;title.focus();}
});
render();
