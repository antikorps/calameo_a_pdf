const f=function(){const r=document.createElement("link").relList;if(r&&r.supports&&r.supports("modulepreload"))return;for(const e of document.querySelectorAll('link[rel="modulepreload"]'))s(e);new MutationObserver(e=>{for(const t of e)if(t.type==="childList")for(const n of t.addedNodes)n.tagName==="LINK"&&n.rel==="modulepreload"&&s(n)}).observe(document,{childList:!0,subtree:!0});function o(e){const t={};return e.integrity&&(t.integrity=e.integrity),e.referrerpolicy&&(t.referrerPolicy=e.referrerpolicy),e.crossorigin==="use-credentials"?t.credentials="include":e.crossorigin==="anonymous"?t.credentials="omit":t.credentials="same-origin",t}function s(e){if(e.ep)return;e.ep=!0;const t=o(e);fetch(e.href,t)}};f();function m(a,r,o,s){return window.go.main.App.GoDescargar(a,r,o,s)}function y(){return window.go.main.App.GoRutaDestino()}const l=document.querySelector("#generar"),g=document.querySelector("#numero-descargas"),b=document.querySelector("textarea"),h=document.querySelector("#carpeta-destino"),v=document.querySelector("#tiempo-espera"),p=document.querySelector("#resumen");var u="";h.addEventListener("click",L);l.addEventListener("click",$);async function L(a){a.preventDefault();try{u=await y()}catch(r){u="",alert(r)}}function q(){const a=b.value.split(`
`);let r=[];for(let o of a)if(o!==""){if(!o.includes("/read/")){alert(`La URL ${o} no se va a procesar porque no se ha encontrado /read/ y parece incorrecta`);continue}r.push(o.trim())}return r}async function $(a){if(a.preventDefault(),l.setAttribute("aria-busy",!0),p.innerHTML="",u===""){alert("Debe escogerse una carpeta de destino"),l.removeAttribute("aria-busy");return}const r=parseInt(g.value),o=parseInt(v.value),s=q();for(const e of s){let t=`
    <article>
      <p>Procesando: <a href="${e}" target="_blank">${e}</a></p>
        <progress indeterminate="true"></progress>
      <p class="resultado-descarga"></p>
    </article>
    `;p.insertAdjacentHTML("beforeend",t);let n=document.querySelectorAll(".resultado-descarga:last-of-type ");n=n[n.length-1];let i=document.querySelectorAll("progress:last-of-type");i=i[i.length-1];let c=document.querySelectorAll("article:last-of-type");c=c[c.length-1];try{let d=await m(e,u,r,o);n.innerHTML=d,c.style.border="2px solid lightgreen"}catch(d){n.innerHTML=d,c.style.border="2px solid tomato"}i.remove()}l.removeAttribute("aria-busy")}
