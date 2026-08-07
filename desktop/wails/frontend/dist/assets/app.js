var Pt=["BN","BN","BN","BN","BN","BN","BN","BN","BN","S","B","S","WS","B","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","B","B","B","S","WS","ON","ON","ET","ET","ET","ON","ON","ON","ON","ON","ON","CS","ON","CS","ON","EN","EN","EN","EN","EN","EN","EN","EN","EN","EN","ON","ON","ON","ON","ON","ON","ON","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","ON","ON","ON","ON","ON","ON","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","ON","ON","ON","ON","BN","BN","BN","BN","BN","BN","B","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","BN","CS","ON","ET","ET","ET","ET","ON","ON","ON","ON","L","ON","ON","ON","ON","ON","ET","ET","EN","EN","ON","L","ON","ON","ON","EN","L","ON","ON","ON","ON","ON","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","ON","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","L","ON","L","L","L","L","L","L","L","L"],Ct=["AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","CS","AL","ON","ON","NSM","NSM","NSM","NSM","NSM","NSM","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","AL","AL","AL","AL","AL","AL","AL","AN","AN","AN","AN","AN","AN","AN","AN","AN","AN","ET","AN","AN","AL","AL","AL","NSM","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","NSM","ON","NSM","NSM","NSM","NSM","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL","AL"];function Et(e){return e<=255?Pt[e]:1424<=e&&e<=1524?"R":1536<=e&&e<=1791?Ct[e&255]:1792<=e&&e<=2220?"AL":"L"}function Tt(e){let t=e.length;if(t===0)return null;let n=new Array(t),o=0;for(let i=0;i<t;i++){let l=Et(e.charCodeAt(i));(l==="R"||l==="AL"||l==="AN")&&o++,n[i]=l}if(o===0)return null;let a=t/o<.3?0:1,r=new Int8Array(t);for(let i=0;i<t;i++)r[i]=a;let s=a&1?"R":"L",c=s,d=c;for(let i=0;i<t;i++)n[i]==="NSM"?n[i]=d:d=n[i];d=c;for(let i=0;i<t;i++){let l=n[i];l==="EN"?n[i]=d==="AL"?"AN":"EN":(l==="R"||l==="L"||l==="AL")&&(d=l)}for(let i=0;i<t;i++)n[i]==="AL"&&(n[i]="R");for(let i=1;i<t-1;i++)n[i]==="ES"&&n[i-1]==="EN"&&n[i+1]==="EN"&&(n[i]="EN"),n[i]==="CS"&&(n[i-1]==="EN"||n[i-1]==="AN")&&n[i+1]===n[i-1]&&(n[i]=n[i-1]);for(let i=0;i<t;i++){if(n[i]!=="EN")continue;let l;for(l=i-1;l>=0&&n[l]==="ET";l--)n[l]="EN";for(l=i+1;l<t&&n[l]==="ET";l++)n[l]="EN"}for(let i=0;i<t;i++){let l=n[i];(l==="WS"||l==="ES"||l==="ET"||l==="CS")&&(n[i]="ON")}d=c;for(let i=0;i<t;i++){let l=n[i];l==="EN"?n[i]=d==="L"?"L":"EN":(l==="R"||l==="L")&&(d=l)}for(let i=0;i<t;i++){if(n[i]!=="ON")continue;let l=i+1;for(;l<t&&n[l]==="ON";)l++;let f=i>0?n[i-1]:c,p=l<t?n[l]:c,g=f!=="L"?"R":"L";if(g===(p!=="L"?"R":"L"))for(let y=i;y<l;y++)n[y]=g;i=l-1}for(let i=0;i<t;i++)n[i]==="ON"&&(n[i]=s);for(let i=0;i<t;i++){let l=n[i];(r[i]&1)===0?l==="R"?r[i]++:(l==="AN"||l==="EN")&&(r[i]+=2):(l==="L"||l==="AN"||l==="EN")&&r[i]++}return r}function Re(e,t){let n=Tt(e);if(n===null)return null;let o=new Int8Array(t.length);for(let a=0;a<t.length;a++)o[a]=n[t[a]];return o}var Nt=/[ \t\n\r\f]+/g,Mt=/[\t\n\r\f]| {2,}|^ | $/;function Bt(e){let t=e??"normal";return t==="pre-wrap"?{mode:t,preserveOrdinarySpaces:!0,preserveHardBreaks:!0}:{mode:t,preserveOrdinarySpaces:!1,preserveHardBreaks:!1}}function Rt(e){if(!Mt.test(e))return e;let t=e.replace(Nt," ");return t.charCodeAt(0)===32&&(t=t.slice(1)),t.length>0&&t.charCodeAt(t.length-1)===32&&(t=t.slice(0,-1)),t}function Wt(e){return/[\r\f]/.test(e)?e.replace(/\r\n/g,`
`).replace(/[\r\f]/g,`
`):e.replace(/\r\n/g,`
`)}var ue=null,It;function Ft(){return ue===null&&(ue=new Intl.Segmenter(It,{granularity:"word"})),ue}var Ot=/\p{Script=Arabic}/u,se=/\p{M}/u,Ie=/\p{Nd}/u;function de(e){return Ot.test(e)}function z(e){for(let t of e){let n=t.codePointAt(0);if(n>=19968&&n<=40959||n>=13312&&n<=19903||n>=131072&&n<=173791||n>=173824&&n<=177983||n>=177984&&n<=178207||n>=178208&&n<=183983||n>=183984&&n<=191471||n>=196608&&n<=201551||n>=63744&&n<=64255||n>=194560&&n<=195103||n>=12288&&n<=12351||n>=12352&&n<=12447||n>=12448&&n<=12543||n>=44032&&n<=55215||n>=65280&&n<=65519)return!0}return!1}var ge=new Set(["\uFF0C","\uFF0E","\uFF01","\uFF1A","\uFF1B","\uFF1F","\u3001","\u3002","\u30FB","\uFF09","\u3015","\u3009","\u300B","\u300D","\u300F","\u3011","\u3017","\u3019","\u301B","\u30FC","\u3005","\u303B","\u309D","\u309E","\u30FD","\u30FE"]),re=new Set(['"',"(","[","{","\u201C","\u2018","\xAB","\u2039","\uFF08","\u3014","\u3008","\u300A","\u300C","\u300E","\u3010","\u3016","\u3018","\u301A"]),fe=new Set(["'","\u2019"]),X=new Set([".",",","!","?",":",";","\u060C","\u061B","\u061F","\u0964","\u0965","\u104A","\u104B","\u104C","\u104D","\u104F",")","]","}","%",'"',"\u201D","\u2019","\xBB","\u203A","\u2026"]),jt=new Set([":",".","\u060C","\u061B"]),Ht=new Set(["\u104F"]),Dt=new Set(["\u201D","\u2019","\xBB","\u203A","\u300D","\u300F","\u3011","\u300B","\u3009","\u3015","\uFF09"]);function Ut(e){if(he(e))return!0;let t=!1;for(let n of e){if(X.has(n)){t=!0;continue}if(!(t&&se.test(n)))return!1}return t}function Gt(e){for(let t of e)if(!ge.has(t)&&!X.has(t))return!1;return e.length>0}function zt(e){if(he(e))return!0;for(let t of e)if(!re.has(t)&&!fe.has(t)&&!se.test(t))return!1;return e.length>0}function he(e){let t=!1;for(let n of e)if(!(n==="\\"||se.test(n))){if(re.has(n)||X.has(n)||fe.has(n)){t=!0;continue}return!1}return t}function Kt(e){let t=Array.from(e),n=t.length;for(;n>0;){let o=t[n-1];if(se.test(o)){n--;continue}if(re.has(o)||fe.has(o)){n--;continue}break}return n<=0||n===t.length?null:{head:t.slice(0,n).join(""),tail:t.slice(n).join("")}}function qt(e,t){if(e.length===0)return!1;for(let n of e)if(n!==t)return!1;return!0}function Vt(e){return!de(e)||e.length===0?!1:jt.has(e[e.length-1])}function Jt(e){return e.length===0?!1:Ht.has(e[e.length-1])}function Qt(e){if(e.length<2||e[0]!==" ")return null;let t=e.slice(1);return/^\p{M}+$/u.test(t)?{space:" ",marks:t}:null}function be(e){for(let t=e.length-1;t>=0;t--){let n=e[t];if(Dt.has(n))return!0;if(!X.has(n))return!1}return!1}function Yt(e,t){if(t.preserveOrdinarySpaces||t.preserveHardBreaks){if(e===" ")return"preserved-space";if(e==="	")return"tab";if(t.preserveHardBreaks&&e===`
`)return"hard-break"}return e===" "?"space":e==="\xA0"||e==="\u202F"||e==="\u2060"||e==="\uFEFF"?"glue":e==="\u200B"?"zero-width-break":e==="\xAD"?"soft-hyphen":"text"}function Zt(e,t,n,o){let a=[],r=null,s="",c=n,d=!1,i=0;for(let l of e){let f=Yt(l,o),p=f==="text"&&t;if(r!==null&&f===r&&p===d){s+=l,i+=l.length;continue}r!==null&&a.push({text:s,isWordLike:d,kind:r,start:c}),r=f,s=l,c=n+i,d=p,i+=l.length}return r!==null&&a.push({text:s,isWordLike:d,kind:r,start:c}),a}function pe(e){return e==="space"||e==="preserved-space"||e==="zero-width-break"||e==="hard-break"}var Xt=/^[A-Za-z][A-Za-z0-9+.-]*:$/;function en(e,t){let n=e.texts[t];return n.startsWith("www.")?!0:Xt.test(n)&&t+1<e.len&&e.kinds[t+1]==="text"&&e.texts[t+1]==="//"}function tn(e){return e.includes("?")&&(e.includes("://")||e.startsWith("www."))}function nn(e){let t=e.texts.slice(),n=e.isWordLike.slice(),o=e.kinds.slice(),a=e.starts.slice();for(let s=0;s<e.len;s++){if(o[s]!=="text"||!en(e,s))continue;let c=s+1;for(;c<e.len&&!pe(o[c]);){t[s]+=t[c],n[s]=!0;let d=t[c].includes("?");if(o[c]="text",t[c]="",c++,d)break}}let r=0;for(let s=0;s<t.length;s++){let c=t[s];c.length!==0&&(r!==s&&(t[r]=c,n[r]=n[s],o[r]=o[s],a[r]=a[s]),r++)}return t.length=r,n.length=r,o.length=r,a.length=r,{len:r,texts:t,isWordLike:n,kinds:o,starts:a}}function rn(e){let t=[],n=[],o=[],a=[];for(let r=0;r<e.len;r++){let s=e.texts[r];if(t.push(s),n.push(e.isWordLike[r]),o.push(e.kinds[r]),a.push(e.starts[r]),!tn(s))continue;let c=r+1;if(c>=e.len||pe(e.kinds[c]))continue;let d="",i=e.starts[c],l=c;for(;l<e.len&&!pe(e.kinds[l]);)d+=e.texts[l],l++;d.length>0&&(t.push(d),n.push(!0),o.push("text"),a.push(i),r=l-1)}return{len:t.length,texts:t,isWordLike:n,kinds:o,starts:a}}var an=new Set([":","-","/","\xD7",",",".","+","\u2013","\u2014"]),We=/^[A-Za-z0-9_]+[,:;]*$/,sn=/[,:;]+$/;function Fe(e){for(let t of e)if(Ie.test(t))return!0;return!1}function me(e){if(e.length===0)return!1;for(let t of e)if(!(Ie.test(t)||an.has(t)))return!1;return!0}function on(e){let t=[],n=[],o=[],a=[];for(let r=0;r<e.len;r++){let s=e.texts[r],c=e.kinds[r];if(c==="text"&&me(s)&&Fe(s)){let d=s,i=r+1;for(;i<e.len&&e.kinds[i]==="text"&&me(e.texts[i]);)d+=e.texts[i],i++;t.push(d),n.push(!0),o.push("text"),a.push(e.starts[r]),r=i-1;continue}t.push(s),n.push(e.isWordLike[r]),o.push(c),a.push(e.starts[r])}return{len:t.length,texts:t,isWordLike:n,kinds:o,starts:a}}function ln(e){let t=[],n=[],o=[],a=[];for(let r=0;r<e.len;r++){let s=e.texts[r],c=e.kinds[r],d=e.isWordLike[r];if(c==="text"&&d&&We.test(s)){let i=s,l=r+1;for(;sn.test(i)&&l<e.len&&e.kinds[l]==="text"&&e.isWordLike[l]&&We.test(e.texts[l]);)i+=e.texts[l],l++;t.push(i),n.push(!0),o.push("text"),a.push(e.starts[r]),r=l-1;continue}t.push(s),n.push(d),o.push(c),a.push(e.starts[r])}return{len:t.length,texts:t,isWordLike:n,kinds:o,starts:a}}function cn(e){let t=[],n=[],o=[],a=[];for(let r=0;r<e.len;r++){let s=e.texts[r];if(e.kinds[r]==="text"&&s.includes("-")){let c=s.split("-"),d=c.length>1;for(let i=0;i<c.length;i++){let l=c[i];if(!d)break;(l.length===0||!Fe(l)||!me(l))&&(d=!1)}if(d){let i=0;for(let l=0;l<c.length;l++){let f=c[l],p=l<c.length-1?`${f}-`:f;t.push(p),n.push(!0),o.push("text"),a.push(e.starts[r]+i),i+=p.length}continue}}t.push(s),n.push(e.isWordLike[r]),o.push(e.kinds[r]),a.push(e.starts[r])}return{len:t.length,texts:t,isWordLike:n,kinds:o,starts:a}}function un(e){let t=[],n=[],o=[],a=[],r=0;for(;r<e.len;){let s=e.texts[r],c=e.isWordLike[r],d=e.kinds[r],i=e.starts[r];if(d==="glue"){let l=s,f=i;for(r++;r<e.len&&e.kinds[r]==="glue";)l+=e.texts[r],r++;if(r<e.len&&e.kinds[r]==="text")s=l+e.texts[r],c=e.isWordLike[r],d="text",i=f,r++;else{t.push(l),n.push(!1),o.push("glue"),a.push(f);continue}}else r++;if(d==="text")for(;r<e.len&&e.kinds[r]==="glue";){let l="";for(;r<e.len&&e.kinds[r]==="glue";)l+=e.texts[r],r++;if(r<e.len&&e.kinds[r]==="text"){s+=l+e.texts[r],c=c||e.isWordLike[r],r++;continue}s+=l}t.push(s),n.push(c),o.push(d),a.push(i)}return{len:t.length,texts:t,isWordLike:n,kinds:o,starts:a}}function dn(e){let t=e.texts.slice(),n=e.isWordLike.slice(),o=e.kinds.slice(),a=e.starts.slice();for(let r=0;r<t.length-1;r++){if(o[r]!=="text"||o[r+1]!=="text"||!z(t[r])||!z(t[r+1]))continue;let s=Kt(t[r]);s!==null&&(t[r]=s.head,t[r+1]=s.tail+t[r+1],a[r+1]=a[r]+s.head.length)}return{len:t.length,texts:t,isWordLike:n,kinds:o,starts:a}}function pn(e,t,n){let o=Ft(),a=0,r=[],s=[],c=[],d=[];for(let p of o.segment(e))for(let g of Zt(p.segment,p.isWordLike??!1,p.index,n)){let v=g.kind==="text";t.carryCJKAfterClosingQuote&&v&&a>0&&c[a-1]==="text"&&z(g.text)&&z(r[a-1])&&be(r[a-1])||v&&a>0&&c[a-1]==="text"&&Gt(g.text)&&z(r[a-1])||v&&a>0&&c[a-1]==="text"&&Jt(r[a-1])?(r[a-1]+=g.text,s[a-1]=s[a-1]||g.isWordLike):v&&a>0&&c[a-1]==="text"&&g.isWordLike&&de(g.text)&&Vt(r[a-1])?(r[a-1]+=g.text,s[a-1]=!0):v&&!g.isWordLike&&a>0&&c[a-1]==="text"&&g.text.length===1&&g.text!=="-"&&g.text!=="\u2014"&&qt(r[a-1],g.text)||v&&!g.isWordLike&&a>0&&c[a-1]==="text"&&(Ut(g.text)||g.text==="-"&&s[a-1])?r[a-1]+=g.text:(r[a]=g.text,s[a]=g.isWordLike,c[a]=g.kind,d[a]=g.start,a++)}for(let p=1;p<a;p++)c[p]==="text"&&!s[p]&&he(r[p])&&c[p-1]==="text"&&(r[p-1]+=r[p],s[p-1]=s[p-1]||s[p],r[p]="");for(let p=a-2;p>=0;p--)if(c[p]==="text"&&!s[p]&&zt(r[p])){let g=p+1;for(;g<a&&r[g]==="";)g++;g<a&&c[g]==="text"&&(r[g]=r[p]+r[g],d[g]=d[p],r[p]="")}let i=0;for(let p=0;p<a;p++){let g=r[p];g.length!==0&&(i!==p&&(r[i]=g,s[i]=s[p],c[i]=c[p],d[i]=d[p]),i++)}r.length=i,s.length=i,c.length=i,d.length=i;let l=un({len:i,texts:r,isWordLike:s,kinds:c,starts:d}),f=dn(ln(cn(on(rn(nn(l))))));for(let p=0;p<f.len-1;p++){let g=Qt(f.texts[p]);g!==null&&(f.kinds[p]!=="space"&&f.kinds[p]!=="preserved-space"||f.kinds[p+1]!=="text"||!de(f.texts[p+1])||(f.texts[p]=g.space,f.isWordLike[p]=!1,f.kinds[p]=f.kinds[p]==="preserved-space"?"preserved-space":"space",f.texts[p+1]=g.marks+f.texts[p+1],f.starts[p+1]=f.starts[p]+g.space.length))}return f}function mn(e,t){if(e.len===0)return[];if(!t.preserveHardBreaks)return[{startSegmentIndex:0,endSegmentIndex:e.len,consumedEndSegmentIndex:e.len}];let n=[],o=0;for(let a=0;a<e.len;a++)e.kinds[a]==="hard-break"&&(n.push({startSegmentIndex:o,endSegmentIndex:a,consumedEndSegmentIndex:a+1}),o=a+1);return o<e.len&&n.push({startSegmentIndex:o,endSegmentIndex:e.len,consumedEndSegmentIndex:e.len}),n}function Oe(e,t,n="normal"){let o=Bt(n),a=o.mode==="pre-wrap"?Wt(e):Rt(e);if(a.length===0)return{normalized:a,chunks:[],len:0,texts:[],isWordLike:[],kinds:[],starts:[]};let r=pn(a,t,o);return{normalized:a,chunks:mn(r,o),...r}}var ee=null,je=new Map,te=null,gn=/\p{Emoji_Presentation}/u,fn=/[\p{Emoji_Presentation}\p{Extended_Pictographic}\p{Regional_Indicator}\uFE0F\u20E3]/u,_e=null,He=new Map;function ye(){if(ee!==null)return ee;if(typeof OffscreenCanvas<"u")return ee=new OffscreenCanvas(1,1).getContext("2d"),ee;if(typeof document<"u")return ee=document.createElement("canvas").getContext("2d"),ee;throw new Error("Text measurement requires OffscreenCanvas or a DOM canvas context.")}function hn(e){let t=je.get(e);return t||(t=new Map,je.set(e,t)),t}function q(e,t){let n=t.get(e);return n===void 0&&(n={width:ye().measureText(e).width,containsCJK:z(e)},t.set(e,n)),n}function J(){if(te!==null)return te;if(typeof navigator>"u")return te={lineFitEpsilon:.005,carryCJKAfterClosingQuote:!1,preferPrefixWidthsForBreakableRuns:!1,preferEarlySoftHyphenBreak:!1},te;let e=navigator.userAgent,n=navigator.vendor==="Apple Computer, Inc."&&e.includes("Safari/")&&!e.includes("Chrome/")&&!e.includes("Chromium/")&&!e.includes("CriOS/")&&!e.includes("FxiOS/")&&!e.includes("EdgiOS/"),o=e.includes("Chrome/")||e.includes("Chromium/")||e.includes("CriOS/")||e.includes("Edg/");return te={lineFitEpsilon:n?1/64:.005,carryCJKAfterClosingQuote:o,preferPrefixWidthsForBreakableRuns:n,preferEarlySoftHyphenBreak:n},te}function bn(e){let t=e.match(/(\d+(?:\.\d+)?)\s*px/);return t?parseFloat(t[1]):16}function ve(){return _e===null&&(_e=new Intl.Segmenter(void 0,{granularity:"grapheme"})),_e}function _n(e){return gn.test(e)||e.includes("\uFE0F")}function De(e){return fn.test(e)}function yn(e,t){let n=He.get(e);if(n!==void 0)return n;let o=ye();o.font=e;let a=o.measureText("\u{1F600}").width;if(n=0,a>t+.5&&typeof document<"u"&&document.body!==null){let r=document.createElement("span");r.style.font=e,r.style.display="inline-block",r.style.visibility="hidden",r.style.position="absolute",r.textContent="\u{1F600}",document.body.appendChild(r);let s=r.getBoundingClientRect().width;document.body.removeChild(r),a-s>.5&&(n=a-s)}return He.set(e,n),n}function vn(e){let t=0,n=ve();for(let o of n.segment(e))_n(o.segment)&&t++;return t}function Sn(e,t){return t.emojiCount===void 0&&(t.emojiCount=vn(e)),t.emojiCount}function V(e,t,n){return n===0?t.width:t.width-Sn(e,t)*n}function Ue(e,t,n,o){if(t.graphemeWidths!==void 0)return t.graphemeWidths;let a=[],r=ve();for(let s of r.segment(e)){let c=q(s.segment,n);a.push(V(s.segment,c,o))}return t.graphemeWidths=a.length>1?a:null,t.graphemeWidths}function Ge(e,t,n,o){if(t.graphemePrefixWidths!==void 0)return t.graphemePrefixWidths;let a=[],r=ve(),s="";for(let c of r.segment(e)){s+=c.segment;let d=q(s,n);a.push(V(s,d,o))}return t.graphemePrefixWidths=a.length>1?a:null,t.graphemePrefixWidths}function ze(e,t){let n=ye();n.font=e;let o=hn(e),a=bn(e),r=t?yn(e,a):0;return{cache:o,fontSize:a,emojiCorrection:r}}function oe(e){return e==="space"||e==="preserved-space"||e==="tab"||e==="zero-width-break"||e==="soft-hyphen"}function Ln(e){return e==="space"}function kn(e,t){if(t<=0)return 0;let n=e%t;return Math.abs(n)<=1e-6?t:t-n}function Se(e,t,n,o){return!o||t===null?e[n]:t[n]-(n>0?t[n-1]:0)}function An(e,t,n,o,a,r){let s=0,c=t;for(;s<e.length;){let d=r?t+e[s]:c+e[s];if((s+1<e.length?d+a:d)>n+o)break;c=d,s++}return{fitCount:s,fittedWidth:c}}function Ke(e,t){return e.simpleLineWalkFastPath?xn(e,t):qe(e,t)}function xn(e,t){let{widths:n,kinds:o,breakableWidths:a,breakablePrefixWidths:r}=e;if(n.length===0)return 0;let s=J(),c=s.lineFitEpsilon,d=0,i=0,l=!1;function f(p){let g=n[p];if(g>t&&a[p]!==null){let v=a[p],y=r[p]??null;i=0;for(let w=0;w<v.length;w++){let T=Se(v,y,w,s.preferPrefixWidthsForBreakableRuns);i>0&&i+T>t+c?(d++,i=T):(i===0&&d++,i+=T)}}else i=g,d++;l=!0}for(let p=0;p<n.length;p++){let g=n[p],v=o[p];if(!l){f(p);continue}let y=i+g;if(y>t+c){if(Ln(v))continue;i=0,l=!1,f(p);continue}i=y}return l?d:d+1}function $n(e,t,n){let{widths:o,kinds:a,breakableWidths:r,breakablePrefixWidths:s}=e;if(o.length===0)return 0;let c=J(),d=c.lineFitEpsilon,i=0,l=0,f=!1,p=0,g=0,v=0,y=0,w=-1,T=0;function B(){w=-1,T=0}function C(b=v,A=y,R=l){i++,n?.({startSegmentIndex:p,startGraphemeIndex:g,endSegmentIndex:b,endGraphemeIndex:A,width:R}),l=0,f=!1,B()}function P(b,A){f=!0,p=b,g=0,v=b+1,y=0,l=A}function h(b,A,R){f=!0,p=b,g=A,v=b,y=A+1,l=R}function F(b,A){if(!f){P(b,A);return}l+=A,v=b+1,y=0}function O(b,A){oe(a[b])&&(w=b+1,T=l-A)}function x(b){E(b,0)}function E(b,A){let R=r[b],j=s[b]??null;for(let I=A;I<R.length;I++){let H=Se(R,j,I,c.preferPrefixWidthsForBreakableRuns);if(!f){h(b,I,H);continue}l+H>t+d?(C(),h(b,I,H)):(l+=H,v=b,y=I+1)}f&&v===b&&y===R.length&&(v=b+1,y=0)}let k=0;for(;k<o.length;){let b=o[k],A=a[k];if(!f){b>t&&r[k]!==null?x(k):P(k,b),O(k,b),k++;continue}if(l+b>t+d){if(oe(A)){F(k,b),C(k+1,0,l-b),k++;continue}if(w>=0){C(w,0,T);continue}if(b>t&&r[k]!==null){C(),x(k),k++;continue}C();continue}F(k,b),O(k,b),k++}return f&&C(),i}function qe(e,t,n){if(e.simpleLineWalkFastPath)return $n(e,t,n);let{widths:o,lineEndFitAdvances:a,lineEndPaintAdvances:r,kinds:s,breakableWidths:c,breakablePrefixWidths:d,discretionaryHyphenWidth:i,tabStopAdvance:l,chunks:f}=e;if(o.length===0||f.length===0)return 0;let p=J(),g=p.lineFitEpsilon,v=0,y=0,w=!1,T=0,B=0,C=0,P=0,h=-1,F=0,O=0,x=null;function E(){h=-1,F=0,O=0,x=null}function k(_=C,S=P,L=y){v++,n?.({startSegmentIndex:T,startGraphemeIndex:B,endSegmentIndex:_,endGraphemeIndex:S,width:L}),y=0,w=!1,E()}function b(_,S){w=!0,T=_,B=0,C=_+1,P=0,y=S}function A(_,S,L){w=!0,T=_,B=S,C=_,P=S+1,y=L}function R(_,S){if(!w){b(_,S);return}y+=S,C=_+1,P=0}function j(_,S){if(!oe(s[_]))return;let L=s[_]==="tab"?0:a[_],W=s[_]==="tab"?S:r[_];h=_+1,F=y-S+L,O=y-S+W,x=s[_]}function I(_){H(_,0)}function H(_,S){let L=c[_],W=d[_]??null;for(let N=S;N<L.length;N++){let K=Se(L,W,N,p.preferPrefixWidthsForBreakableRuns);if(!w){A(_,N,K);continue}y+K>t+g?(k(),A(_,N,K)):(y+=K,C=_,P=N+1)}w&&C===_&&P===L.length&&(C=_+1,P=0)}function M(_){if(x!=="soft-hyphen")return!1;let S=c[_];if(S===null)return!1;let L=p.preferPrefixWidthsForBreakableRuns?d[_]??S:S,W=L!==S,{fitCount:N,fittedWidth:K}=An(L,y,t,g,i,W);return N===0?!1:(y=K,C=_,P=N,E(),N===S.length?(C=_+1,P=0,!0):(k(_,N,K+i),H(_,N),!0))}function G(_){v++,n?.({startSegmentIndex:_.startSegmentIndex,startGraphemeIndex:0,endSegmentIndex:_.consumedEndSegmentIndex,endGraphemeIndex:0,width:0}),E()}for(let _=0;_<f.length;_++){let S=f[_];if(S.startSegmentIndex===S.endSegmentIndex){G(S);continue}w=!1,y=0,T=S.startSegmentIndex,B=0,C=S.startSegmentIndex,P=0,E();let L=S.startSegmentIndex;for(;L<S.endSegmentIndex;){let W=s[L],N=W==="tab"?kn(y,l):o[L];if(W==="soft-hyphen"){w&&(C=L+1,P=0,h=L+1,F=y+i,O=y+i,x=W),L++;continue}if(!w){N>t&&c[L]!==null?I(L):b(L,N),j(L,N),L++;continue}if(y+N>t+g){let $t=y+(W==="tab"?0:a[L]),wt=y+(W==="tab"?N:r[L]);if(x==="soft-hyphen"&&p.preferEarlySoftHyphenBreak&&F<=t+g){k(h,0,O);continue}if(x==="soft-hyphen"&&M(L)){L++;continue}if(oe(W)&&$t<=t+g){R(L,N),k(L+1,0,wt),L++;continue}if(h>=0&&F<=t+g){k(h,0,O);continue}if(N>t&&c[L]!==null){k(),I(L),L++;continue}k();continue}R(L,N),j(L,N),L++}if(w){let W=h===S.consumedEndSegmentIndex?O:y;k(S.consumedEndSegmentIndex,0,W)}}return v}var Le=null;function wn(){return Le===null&&(Le=new Intl.Segmenter(void 0,{granularity:"grapheme"})),Le}function Pn(e){return e?{widths:[],lineEndFitAdvances:[],lineEndPaintAdvances:[],kinds:[],simpleLineWalkFastPath:!0,segLevels:null,breakableWidths:[],breakablePrefixWidths:[],discretionaryHyphenWidth:0,tabStopAdvance:0,chunks:[],segments:[]}:{widths:[],lineEndFitAdvances:[],lineEndPaintAdvances:[],kinds:[],simpleLineWalkFastPath:!0,segLevels:null,breakableWidths:[],breakablePrefixWidths:[],discretionaryHyphenWidth:0,tabStopAdvance:0,chunks:[]}}function Cn(e,t,n){let o=wn(),a=J(),{cache:r,emojiCorrection:s}=ze(t,De(e.normalized)),c=V("-",q("-",r),s),i=V(" ",q(" ",r),s)*8;if(e.len===0)return Pn(n);let l=[],f=[],p=[],g=[],v=e.chunks.length<=1,y=n?[]:null,w=[],T=[],B=n?[]:null,C=Array.from({length:e.len}),P=Array.from({length:e.len});function h(x,E,k,b,A,R,j,I){A!=="text"&&A!=="space"&&A!=="zero-width-break"&&(v=!1),l.push(E),f.push(k),p.push(b),g.push(A),y?.push(R),w.push(j),T.push(I),B!==null&&B.push(x)}for(let x=0;x<e.len;x++){C[x]=l.length;let E=e.texts[x],k=e.isWordLike[x],b=e.kinds[x],A=e.starts[x];if(b==="soft-hyphen"){h(E,0,c,c,b,A,null,null),P[x]=l.length;continue}if(b==="hard-break"){h(E,0,0,0,b,A,null,null),P[x]=l.length;continue}if(b==="tab"){h(E,0,0,0,b,A,null,null),P[x]=l.length;continue}let R=q(E,r);if(b==="text"&&R.containsCJK){let M="",G=0;for(let _ of o.segment(E)){let S=_.segment;if(M.length===0){M=S,G=_.index;continue}if(re.has(M)||ge.has(S)||X.has(S)||a.carryCJKAfterClosingQuote&&z(S)&&be(M)){M+=S;continue}let L=q(M,r),W=V(M,L,s);h(M,W,W,W,"text",A+G,null,null),M=S,G=_.index}if(M.length>0){let _=q(M,r),S=V(M,_,s);h(M,S,S,S,"text",A+G,null,null)}P[x]=l.length;continue}let j=V(E,R,s),I=b==="space"||b==="preserved-space"||b==="zero-width-break"?0:j,H=b==="space"||b==="zero-width-break"?0:j;if(k&&E.length>1){let M=Ue(E,R,r,s),G=a.preferPrefixWidthsForBreakableRuns?Ge(E,R,r,s):null;h(E,j,I,H,b,A,M,G)}else h(E,j,I,H,b,A,null,null);P[x]=l.length}let F=En(e.chunks,C,P),O=y===null?null:Re(e.normalized,y);return B!==null?{widths:l,lineEndFitAdvances:f,lineEndPaintAdvances:p,kinds:g,simpleLineWalkFastPath:v,segLevels:O,breakableWidths:w,breakablePrefixWidths:T,discretionaryHyphenWidth:c,tabStopAdvance:i,chunks:F,segments:B}:{widths:l,lineEndFitAdvances:f,lineEndPaintAdvances:p,kinds:g,simpleLineWalkFastPath:v,segLevels:O,breakableWidths:w,breakablePrefixWidths:T,discretionaryHyphenWidth:c,tabStopAdvance:i,chunks:F}}function En(e,t,n){let o=[];for(let a=0;a<e.length;a++){let r=e[a],s=r.startSegmentIndex<t.length?t[r.startSegmentIndex]:n[n.length-1]??0,c=r.endSegmentIndex<t.length?t[r.endSegmentIndex]:n[n.length-1]??0,d=r.consumedEndSegmentIndex<t.length?t[r.consumedEndSegmentIndex]:n[n.length-1]??0;o.push({startSegmentIndex:s,endSegmentIndex:c,consumedEndSegmentIndex:d})}return o}function Tn(e,t,n,o){let a=Oe(e,J(),o?.whiteSpace);return Cn(a,t,n)}function Ve(e,t,n){return Tn(e,t,!1,n)}function Je(e,t,n){let o=Ke(e,t);return{lineCount:o,height:o*n}}var ke={"pt-BR":{app_kicker:"Mesa aberta",app_stamp:"Edi\xE7\xE3o desktop",title_main:"Truco Paulista",title_sub:"Abra uma mesa r\xE1pida para treino ou uma partida online em poucos passos.",locale_label:"Idioma",setup_title:"Monte a mesa",setup_intro:"Comece no treino ou abra uma mesa online para jogar com outras pessoas.",setup_offline_title:"Jogar offline",setup_offline_note:"Treino imediato contra CPU, sem depender de rede.",setup_offline_caption:"Come\xE7o r\xE1pido",setup_online_title:"Jogar online",setup_online_note:"Crie uma sala ou entre por convite.",setup_online_caption:"Convite e chat",setup_name:"Seu nome",setup_players:"Jogadores",setup_start:"Abrir partida",setup_host:"Criar mesa",setup_join:"Entrar com convite",setup_invite:"Convite",setup_role:"Papel",setup_transport:"Transporte",setup_relay:"Relay URL",setup_help:"Tudo o que voc\xEA precisa para come\xE7ar a jogar.",setup_signal_title:"Mesa pronta",setup_signal_body:"Setup, lobby e partida leem o mesmo estado para manter a transi\xE7\xE3o suave.",setup_runtime_title:"Uma s\xF3 tela",setup_runtime_body:"A camada desktop foca no jogo, enquanto as regras continuam no Go.",setup_mode_offline:"offline",setup_mode_online:"online",setup_online_support:"Convite, sala e assentos ficam prontos para come\xE7ar a jogar.",role_auto:"Auto",role_partner:"Parceiro",role_opponent:"Advers\xE1rio",transport_auto:"Auto",transport_direct:"Direto (TCP/TLS)",transport_relay:"Relay QUIC v2",invite_copy:"Copiar",lobby_title:"Mesa pronta",lobby_host_headline:"Sala criada",lobby_join_headline:"Conectado por convite",lobby_slots:"Assentos",lobby_overview:"Vis\xE3o geral",lobby_events:"Atualiza\xE7\xF5es",lobby_chat:"Chat",lobby_start:"Come\xE7ar partida",lobby_refresh:"Atualizar",lobby_leave:"Sair da mesa",lobby_empty:"Sem movimenta\xE7\xE3o ainda.",game_title_offline:"Mesa de treino",game_title_online:"Mesa ao vivo",game_hand:"Sua m\xE3o",game_table:"Centro da mesa",game_controls:"Controles",game_activity:"Atualiza\xE7\xF5es",game_overview:"Vis\xE3o geral",game_network:"Conex\xE3o",game_play:"Jogar",game_face_down:"Virada",game_truco:"Truco",game_raise:"Subir",game_accept:"Aceitar",game_refuse:"Correr",game_play_again:"Voltar ao in\xEDcio",game_you:"Voc\xEA",game_partner:"Parceiro",game_opponent:"Advers\xE1rio",game_cpu:"CPU",game_turn:"Vez",game_wait:"Aguarde",game_ready:"Jogue uma carta",game_round:"Vaza",game_stake:"Vale",game_status:"Pulso",game_last_trick:"\xDAltima vaza",game_vira:"Vira",game_manilha:"Manilha",game_trick_track:"Ritmo da m\xE3o",game_trick_label:"Vaza %s",game_trick_pending:"aberta",game_trick_draw:"empate",game_trick_team:"time %s",game_table_waiting:"A mesa est\xE1 armada. A primeira carta ainda n\xE3o caiu.",game_table_notes:"Notas da mesa",game_player_to_move:"Pr\xF3ximo a jogar",game_log_new_hand:"Nova m\xE3o: vira %s, manilha %s.",game_log_face_down:"%s jogou carta virada.",game_log_played:"%s jogou %s.",game_log_accept:"%s aceitou %s. Aposta agora vale %s.",game_log_fold:"%s correu do truco.",game_no_events:"Sem eventos recentes.",game_runtime_stale_title:"A mesa precisa atualizar",game_runtime_stale_copy:"Recarregue para pegar o estado mais recente.",button_busy:"Processando...",header_resync:"Sincronizar",header_diagnostics:"Detalhes",header_hide_diagnostics:"Ocultar detalhes",connection_status:"Status",connection_mode:"Modo",connection_transport:"Transporte",connection_protocol:"Protocolo",connection_backlog:"Fila",connection_role:"Papel",connection_online:"online",connection_offline:"offline",slot_empty:"lugar livre",slot_you:"voc\xEA",slot_host:"host",slot_online:"online",slot_offline:"offline",slot_cpu:"cpu",action_vote_host:"Votar host",action_replacement_invite:"Chamar substituto",team_one:"Time 1",team_two:"Time 2",status_your_turn:"Sua vez. Escolha a carta e marque o compasso.",status_wait_turn:"Aguardando %s.",status_pending_you:"%s na sua m\xE3o. Aceite, corra ou suba.",status_pending_other:"Aguardando %s responder %s.",status_match_end:"A rodada fechou.",status_idle:"Tudo pronto para uma nova mesa.",status_action_pending:"Aplicando %s e conferindo a mesa.",status_action_timeout:"%s demorou demais. Tente sincronizar ou voltar ao in\xEDcio.",status_snapshot_timeout:"A atualiza\xE7\xE3o demorou demais. Recarregue a mesa para recuperar o estado.",status_refresh_failed:"Falha ao aplicar %s: %s",status_refresh_stale:"A a\xE7\xE3o terminou, mas a mesa n\xE3o entrou em um estado v\xE1lido (%s). Sincronize antes de continuar.",status_transition_failed:"%s terminou, mas o sistema voltou em %s em vez de %s.",status_render_recovery:"A tela entrou em modo de recupera\xE7\xE3o.",status_render_recovery_copy:"A mesa recebeu um estado novo, mas a tela n\xE3o conseguiu montar tudo. Recarregue se o problema continuar.",status_waiting_lobby:"A mesa existe, mas o painel ainda n\xE3o ficou pronto. Sincronize para continuar.",status_waiting_match:"A partida foi criada, mas ainda falta o estado completo. Sincronize para continuar.",trick_tie:"Empate na vaza %s.",trick_win:"%s levou a vaza %s.",overlay_win:"A mesa veio para o seu lado.",overlay_loss:"A rodada escapou.",overlay_score:"Placar final: %s x %s",raise_3:"Truco",raise_6:"Seis",raise_9:"Nove",raise_12:"Doze",event_chat:"chat",event_client_joined:"jogador entrou",event_failover_promoted:"failover promovido",event_failover_rejoined:"failover recuperado",event_host_created:"mesa criada",event_locale_changed:"idioma trocado",event_match_started:"partida iniciada",event_session_closed:"sess\xE3o encerrada",event_session_ready:"mesa pronta",event_system:"sistema",event_replacement_invite:"convite de substitui\xE7\xE3o",event_error:"erro",event_lobby_updated:"mesa atualizada",event_match_updated:"partida atualizada",signal_failover_promoted:"O host caiu e a mesa foi promovida. Continue da\xED sem reiniciar.",signal_failover_rejoined:"A conex\xE3o voltou e a mesa se encaixou no host atual.",signal_replacement_ready:"Saiu um convite de substitui\xE7\xE3o para esta mesa.",suit_Ouros:"Ouros",suit_Espadas:"Espadas",suit_Copas:"Copas",suit_Paus:"Paus",card_of:"%s de %s",copy_done:"Copiado",diagnostics_title:"Detalhes da mesa",diagnostics_versions:"Vers\xF5es",diagnostics_event_log:"Log de intents",diagnostics_force_tick:"For\xE7ar pulso de CPU",diagnostics_none:"Sem entradas de diagn\xF3stico ainda.",diagnostics_mode:"Modo atual",diagnostics_sequence:"Sequ\xEAncia",diagnostics_last_action:"\xDAltima a\xE7\xE3o",diagnostics_refresh:"Reconcilia\xE7\xE3o",diagnostics_refresh_error:"Erro de reconcilia\xE7\xE3o",diagnostics_render_error:"Erro de render",relay_placeholder:"https://relay.example.com",invite_hint:"Compartilhe a chave com quem vai entrar na mesa.",name_placeholder:"Voc\xEA",chat_placeholder:"Escreva para a mesa..."},"en-US":{app_kicker:"Open Table",app_stamp:"Desktop edition",title_main:"Truco Paulista",title_sub:"Open a quick practice table or start an online match in a few steps.",locale_label:"Language",setup_title:"Set up the table",setup_intro:"Start in practice mode or open an online table for a real match.",setup_offline_title:"Play offline",setup_offline_note:"Instant practice against CPU players, no network required.",setup_offline_caption:"Quick start",setup_online_title:"Play online",setup_online_note:"Create a room or join by invite.",setup_online_caption:"Invite and chat",setup_name:"Your name",setup_players:"Players",setup_start:"Open match",setup_host:"Create table",setup_join:"Join with invite",setup_invite:"Invite",setup_role:"Role",setup_transport:"Transport",setup_relay:"Relay URL",setup_help:"Everything you need to start playing.",setup_signal_title:"Table ready",setup_signal_body:"Setup, lobby, and match all share the same state for a smooth handoff.",setup_runtime_title:"One screen",setup_runtime_body:"The desktop shell stays focused on play while rules stay in Go.",setup_mode_offline:"offline",setup_mode_online:"online",setup_online_support:"Invite someone, open the room, and start playing.",role_auto:"Auto",role_partner:"Partner",role_opponent:"Opponent",transport_auto:"Auto",transport_direct:"Direct (TCP/TLS)",transport_relay:"Relay QUIC v2",invite_copy:"Copy",lobby_title:"Table ready",lobby_host_headline:"Room created",lobby_join_headline:"Connected by invite",lobby_slots:"Seats",lobby_overview:"Overview",lobby_events:"Updates",lobby_chat:"Chat",lobby_start:"Start match",lobby_refresh:"Refresh",lobby_leave:"Leave table",lobby_empty:"No movement yet.",game_title_offline:"Practice table",game_title_online:"Live table",game_hand:"Your hand",game_table:"Center table",game_controls:"Controls",game_activity:"Updates",game_overview:"Overview",game_network:"Connection",game_play:"Play",game_face_down:"Face down",game_truco:"Truco",game_raise:"Raise",game_accept:"Accept",game_refuse:"Fold",game_play_again:"Back to start",game_you:"You",game_partner:"Partner",game_opponent:"Opponent",game_cpu:"CPU",game_turn:"Turn",game_wait:"Wait",game_ready:"Play a card",game_round:"Trick",game_stake:"Stake",game_status:"Pulse",game_last_trick:"Last trick",game_vira:"Turn-up",game_manilha:"Trump rank",game_trick_track:"Hand rhythm",game_trick_label:"Trick %s",game_trick_pending:"open",game_trick_draw:"draw",game_trick_team:"team %s",game_table_waiting:"The table is set. The first card has not landed yet.",game_table_notes:"Table notes",game_player_to_move:"Next to act",game_log_new_hand:"New hand: turn-up %s, trump rank %s.",game_log_face_down:"%s played face down.",game_log_played:"%s played %s.",game_log_accept:"%s accepted %s. The stake is now %s.",game_log_fold:"%s folded to truco.",game_no_events:"No recent events.",game_runtime_stale_title:"The table needs a refresh",game_runtime_stale_copy:"Refresh to pick up the latest state.",button_busy:"Working...",header_resync:"Resync",header_diagnostics:"Details",header_hide_diagnostics:"Hide details",connection_status:"Status",connection_mode:"Mode",connection_transport:"Transport",connection_protocol:"Protocol",connection_backlog:"Backlog",connection_role:"Role",connection_online:"online",connection_offline:"offline",slot_empty:"open seat",slot_you:"you",slot_host:"host",slot_online:"online",slot_offline:"offline",slot_cpu:"cpu",action_vote_host:"Vote host",action_replacement_invite:"Replacement invite",team_one:"Team 1",team_two:"Team 2",status_your_turn:"Your turn. Pick the card and set the pace.",status_wait_turn:"Waiting for %s.",status_pending_you:"%s is in your hands. Accept, fold, or raise.",status_pending_other:"Waiting for %s to answer %s.",status_match_end:"The round is closed.",status_idle:"Everything is ready for a new table.",status_action_pending:"Applying %s and checking the table.",status_action_timeout:"%s took too long. Try resyncing or returning to the start.",status_snapshot_timeout:"The update took too long. Refresh the table to recover state.",status_refresh_failed:"Failed to apply %s: %s",status_refresh_stale:"The action finished, but the table did not enter a valid state (%s). Resync before continuing.",status_transition_failed:"%s finished, but the system returned %s instead of %s.",status_render_recovery:"The screen switched into recovery mode.",status_render_recovery_copy:"The table received fresh state, but the screen could not render it cleanly. Refresh if it keeps happening.",status_waiting_lobby:"The table exists, but the lobby panel is not ready yet. Resync to continue.",status_waiting_match:"The match was created, but the full state is still loading. Resync to continue.",trick_tie:"Trick %s ended in a tie.",trick_win:"%s took trick %s.",overlay_win:"The table went your way.",overlay_loss:"The round slipped away.",overlay_score:"Final score: %s x %s",raise_3:"Truco",raise_6:"Six",raise_9:"Nine",raise_12:"Twelve",event_chat:"chat",event_client_joined:"player joined",event_failover_promoted:"failover promoted",event_failover_rejoined:"failover rejoined",event_host_created:"table created",event_locale_changed:"locale changed",event_match_started:"match started",event_session_closed:"session closed",event_session_ready:"table ready",event_system:"system",event_replacement_invite:"replacement invite",event_error:"error",event_lobby_updated:"lobby updated",event_match_updated:"match updated",signal_failover_promoted:"The host dropped and this table was promoted. Keep playing from here.",signal_failover_rejoined:"The connection recovered and the table rejoined the current host.",signal_replacement_ready:"A replacement invite was issued for this table.",suit_Ouros:"Diamonds",suit_Espadas:"Spades",suit_Copas:"Hearts",suit_Paus:"Clubs",card_of:"%s of %s",copy_done:"Copied",diagnostics_title:"Table details",diagnostics_versions:"Versions",diagnostics_event_log:"Intent log",diagnostics_force_tick:"Force CPU pulse",diagnostics_none:"No diagnostic entries yet.",diagnostics_mode:"Current mode",diagnostics_sequence:"Sequence",diagnostics_last_action:"Last action",diagnostics_refresh:"Reconciliation",diagnostics_refresh_error:"Reconciliation error",diagnostics_render_error:"Render error",relay_placeholder:"https://relay.example.com",invite_hint:"Share the key with the person joining the table.",name_placeholder:"You",chat_placeholder:"Write to the table..."}};function Qe(e,t,...n){let r=(ke[e]||ke["pt-BR"])[t]||ke["pt-BR"][t]||t;for(let s of n)r=r.replace("%s",String(s));return r}function Ze(e){let{bundle:t,panelTab:n,events:o,isOnlineMode:a,t:r,escapeHtml:s,busyAttr:c,buttonLabel:d,renderMetric:i,renderEventFeed:l,renderCard:f,protocolLabel:p,cardLabel:g,playerName:v,teamScore:y,localTeam:w,nextStake:T,raiseLabel:B,lastTrickCopy:C,seatPositions:P}=e,h=t.match;if(!h||!h.CurrentHand||!Array.isArray(h.Players)||h.Players.length===0)return"";let F=t.ui.actions.local_player_id>=0?t.ui.actions.local_player_id:h.CurrentPlayerIdx,O=h.Players.find(M=>M.ID===F)||h.Players[0],x=w(h,t),E=h.PendingRaiseFor,k=h.PendingRaiseTo||T(h.CurrentHand.Stake),b=t.ui.actions.must_respond,A=t.ui.actions.can_ask_or_raise,R=t.ui.actions.can_play_card,j=h.TurnPlayer===F,I=r(a?"game_title_online":"game_title_offline"),H=h.MatchFinished?r("status_match_end"):E===x?r("status_pending_you",B(k)):E>=0?r("status_pending_other",v(h,h.TurnPlayer),B(k)):j?r("status_your_turn"):r("status_wait_turn",v(h,h.TurnPlayer));return`
    <section class="game10">
      <article class="surface-card game10-hud">
        <div class="game10-hud-cluster">
          <div class="game10-score-badge${x===0?" game10-score-badge-friendly":""}">
            <span>${s(r("team_one"))}</span>
            <strong>${y(h,0)}</strong>
          </div>
          <div class="game10-center-signal">
            <p class="eyebrow">${s(I)}</p>
            <h2>${s(H)}</h2>
            <div class="game10-chip-row">
              <span class="game10-chip">${s(r("game_stake"))} ${h.CurrentHand.Stake}</span>
              <span class="game10-chip">${s(r("game_round"))} ${h.CurrentHand.Round}/3</span>
              <span class="game10-chip">${s(v(h,h.TurnPlayer))}</span>
            </div>
          </div>
          <div class="game10-score-badge${x===1?" game10-score-badge-friendly":""}">
            <span>${s(r("team_two"))}</span>
            <strong>${y(h,1)}</strong>
          </div>
        </div>
      </article>

      <div class="game10-grid">
        <article class="surface-card game10-table-wrap">
          <div class="game10-table-head">
            <div>
              <p class="eyebrow">${s(r("game_status"))}</p>
              <h3>${s(I)}</h3>
            </div>
            <div class="desktop-action-row game10-head-actions">
              <button class="ghost-button" type="button" data-client-action="refresh">${s(r("header_resync"))}</button>
              ${t.ui.actions.can_close_session?`<form data-api-action="${a?"closeSession":"reset"}" data-form-id="${a?"closeSession":"reset"}"><button class="ghost-button danger" type="submit"${c(a?"closeSession":"reset")}>${d(a?"closeSession":"reset",r(a?"lobby_leave":"game_play_again"))}</button></form>`:""}
            </div>
          </div>

          <div class="game10-status-ribbon">
            <span>${s(jn(h,j,b,r))}</span>
            <strong>${s(H)}</strong>
          </div>

          <div class="game10-felt game10-felt-${h.NumPlayers}">
            ${Ae(h,t,P,x,"top",r,s)}
            ${h.NumPlayers===4?Ae(h,t,P,x,"left",r,s):""}
            ${h.NumPlayers===4?Ae(h,t,P,x,"right",r,s):""}

            <div class="game10-center-stage">
              <div class="game10-side-chip">
                <span>${s(r("game_vira"))}</span>
                ${f(h.CurrentHand.Vira,"small")}
              </div>
              <div class="game10-pulse-core">
                <div class="game10-trick-rail">
                  ${Nn(h,r,s)}
                </div>
                <div class="game10-round-cards">
                  ${Mn(h,f,v,r,s)}
                </div>
              </div>
              <div class="game10-side-chip">
                <span>${s(r("game_manilha"))}</span>
                <strong>${s(h.CurrentHand.Manilha||"-")}</strong>
              </div>
            </div>

            <div class="game10-bottom-area">
              <div class="game10-action-band">
                ${b?Wn(A,t.ui.actions.can_accept,t.ui.actions.can_refuse,B(T(k)),r,c,d):Rn(A,h.MatchFinished,r,c,d,B(T(h.CurrentHand.Stake)))}
              </div>
              <div class="game10-hand-stage">
                <div class="game10-hand-head">
                  <div>
                    <p class="eyebrow">${s(r("game_hand"))}</p>
                    <h3>${s(O?.Name||r("game_you"))}</h3>
                  </div>
                  <div class="game10-chip-row">
                    ${et(Xe(h,F),s)}
                    ${j?`<span class="game10-turn-badge">${s(r("game_turn"))}</span>`:""}
                  </div>
                </div>
                <div class="game10-hand-row">
                  ${(O?.Hand||[]).map((M,G)=>Bn(M,G,R,h.CurrentHand.Round>=2,r,s,c,d,f,g)).join("")}
                </div>
              </div>
            </div>
          </div>
        </article>

        <aside class="surface-card game10-side">
          <div class="card-head">
            <div>
              <p class="eyebrow">${s(r("game_activity"))}</p>
              <h3>${s(On(n,r))}</h3>
            </div>
            <div class="panel-tabs">
              ${xe("pulse",n,r("game_activity"),s)}
              ${xe("network",n,r("game_network"),s)}
              ${xe("chat",n,r("lobby_chat"),s)}
            </div>
          </div>
          ${In(n,t,o,a,r,s,i,l,p,C,c,d,h,v)}
        </aside>
      </div>

      ${h.MatchFinished?Hn(h,x,r,s,c,d,a):""}
    </section>
  `}function Ae(e,t,n,o,a,r,s){let c=n(e,t),d=(e.Players||[]).find(f=>c.get(f.ID)===a);if(!d)return"";let i=d.Team===o?r("game_partner"):r("game_opponent"),l=d.ID===e.TurnPlayer;return`
    <div class="game10-seat game10-seat-${a}${l?" game10-seat-turn":""}">
      <div class="game10-seat-label">
        <strong>${s(d.Name)}</strong>
        <span>${s(i)}${d.CPU?` \xB7 ${s(r("game_cpu"))}`:""}</span>
      </div>
      <div class="game10-seat-meta">
        ${et(Xe(e,d.ID),s)}
      </div>
      <div class="player-cards">
        ${(d.Hand||[]).map(()=>'<span class="card-back tiny"></span>').join("")}
      </div>
    </div>
  `}function Nn(e,t,n){let o=e.CurrentHand?.TrickResults||[];return Array.from({length:3},(a,r)=>{let s=o[r],c=t("game_trick_pending"),d="trick-pill";return s===-1?(c=t("game_trick_draw"),d+=" trick-pill-draw"):(s===0||s===1)&&(c=t("game_trick_team",s+1),d+=` trick-pill-team-${s+1}`),`<span class="${d}">${n(t("game_trick_label",r+1))} \xB7 ${n(c)}</span>`}).join("")}function Mn(e,t,n,o,a){let r=e.CurrentHand?.RoundCards||[];return r.length===0?`<div class="round-card-placeholder">${a(o("game_table_waiting"))}</div>`:r.map(s=>`
      <div class="played-card">
        <span>${a(n(e,s.PlayerID))}</span>
        ${s.FaceDown?'<span class="card-back small"></span>':t(s.Card,"small")}
      </div>
    `).join("")}function Bn(e,t,n,o,a,r,s,c,d,i){return n?`
    <div class="game10-hand-card">
      <form data-api-action="play" data-form-id="play-${t}">
        <input type="hidden" name="cardIndex" value="${t}">
        <button class="card-button game10-card-button" type="submit" aria-label="${r(`${a("game_play")} ${i(e)}`)}"${s(`play-${t}`)}>${d(e)}</button>
      </form>
      <span class="card-caption">${r(i(e))}</span>
      ${o?`<form data-api-action="play" data-form-id="play-down-${t}"><input type="hidden" name="cardIndex" value="${t}"><input type="hidden" name="faceDown" value="true"><button class="ghost-button game10-face-down" type="submit" aria-label="${r(`${a("game_face_down")} ${i(e)}`)}"${s(`play-down-${t}`)}>${c(`play-down-${t}`,a("game_face_down"))}</button></form>`:""}
    </div>
  `:`<div class="game10-hand-card game10-hand-card-locked">${d(e)}<span class="card-caption">${r(i(e))}</span></div>`}function Rn(e,t,n,o,a,r){return t?`<form data-api-action="reset" data-form-id="reset"><button class="primary-button" type="submit"${o("reset")}>${a("reset",n("game_play_again"))}</button></form>`:`
    <div class="game10-control-row">
      ${e?`<form data-api-action="truco" data-form-id="truco"><button class="primary-button" type="submit"${o("truco")}>${a("truco",r)}</button></form>`:""}
    </div>
  `}function Wn(e,t,n,o,a,r,s){return`
    <div class="game10-control-row">
      ${t?`<form data-api-action="accept" data-form-id="accept"><button class="secondary-button" type="submit"${r("accept")}>${s("accept",a("game_accept"))}</button></form>`:""}
      ${e?`<form data-api-action="truco" data-form-id="raise"><button class="primary-button" type="submit"${r("raise")}>${s("raise",`${a("game_raise")} ${o}`)}</button></form>`:""}
      ${n?`<form data-api-action="refuse" data-form-id="refuse"><button class="ghost-button danger" type="submit"${r("refuse")}>${s("refuse",a("game_refuse"))}</button></form>`:""}
    </div>
  `}function In(e,t,n,o,a,r,s,c,d,i,l,f,p,g){let v=t.connection.network;switch(e){case"network":return`
        <div class="game10-panel">
          <div class="telemetry-grid">
            ${s(a("connection_status"),t.connection.status)}
            ${s(a("connection_mode"),t.connection.is_online?a("connection_online"):a("connection_offline"))}
            ${s(a("connection_transport"),v?.transport||"-")}
            ${s(a("connection_protocol"),d(v))}
            ${s(a("connection_backlog"),String(t.diagnostics.event_backlog||0))}
            ${t.lobby?.role?s(a("connection_role"),t.lobby.role):""}
          </div>
          ${Fn(t.ui.lobby_slots||[],r,a)}
          ${t.connection.last_error?`<div class="lobby10-inline-error">${r(`${t.connection.last_error.code}: ${t.connection.last_error.message}`)}</div>`:""}
          ${Ye(n,a)?`<p class="supporting-copy">${r(Ye(n,a)||"")}</p>`:""}
        </div>
      `;case"chat":return`
        <div class="game10-panel">
          <pre class="event-feed compact" role="log" aria-live="polite" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${r(c((p.Logs||[]).slice(-4)))}</pre>
          ${o?`<form class="chat-form" data-api-action="sendChat" data-form-id="sendChat"><input name="message" type="text" autocomplete="off" placeholder="${r(a("chat_placeholder"))}"><button class="secondary-button" type="submit"${l("sendChat")}>${f("sendChat",a("lobby_chat"))}</button></form>`:""}
        </div>
      `;default:return`
        <div class="game10-panel">
          <div class="game10-pulse-stack">
            <div class="game10-pulse-card">
              <span>${r(a("game_last_trick"))}</span>
              <strong>${r(i(p))}</strong>
            </div>
            <div class="game10-pulse-card">
              <span>${r(a("game_player_to_move"))}</span>
              <strong>${r(g(p,p.TurnPlayer))}</strong>
            </div>
          </div>
          <pre class="event-feed compact" role="log" aria-live="polite" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${r(c((p.Logs||[]).slice(-6)))}</pre>
        </div>
      `}}function Fn(e,t,n){return e.length===0?"":`
    <div class="game10-seat-strip">
      ${e.map(o=>`<div class="game10-seat-pill${o.is_local?" game10-seat-pill-local":""}"><strong>${t(o.name||n("slot_empty"))}</strong><span>${t(o.is_connected?n("slot_online"):n("slot_offline"))}</span></div>`).join("")}
    </div>
  `}function xe(e,t,n,o){return`<button class="panel-tab${t===e?" panel-tab-active":""}" type="button" role="tab" aria-selected="${t===e?"true":"false"}" data-panel-tab="game:${e}">${o(n)}</button>`}function On(e,t){switch(e){case"network":return t("game_network");case"chat":return t("lobby_chat");default:return t("game_overview")}}function Ye(e,t){let n=[...e].reverse().find(o=>o.kind==="failover_promoted"||o.kind==="failover_rejoined");return n?n.kind==="failover_promoted"?t("signal_failover_promoted"):t("signal_failover_rejoined"):""}function jn(e,t,n,o){return e.MatchFinished?o("status_match_end"):o(n?"game_raise":t?"status_your_turn":"game_wait")}function Xe(e,t){let n=e.CurrentHand?.Dealer??-1,o=e.NumPlayers||2;return t===n?"\u{1F0CF}":t===(n+1)%o?"\u270B":o===4&&t===(n+o-1)%o?"\u{1F9B6}":""}function et(e,t){return e?`<span class="game10-role-chip">${t(e)}</span>`:""}function Hn(e,t,n,o,a,r,s){let c=e.WinnerTeam===t;return`
    <div class="overlay-layer">
      <div class="overlay-card">
        <p class="eyebrow">${o(n("game_status"))}</p>
        <h3>${o(n(c?"overlay_win":"overlay_loss"))}</h3>
        <p>${o(n("overlay_score",String(e.MatchPoints[0]||0),String(e.MatchPoints[1]||0)))}</p>
        <form data-api-action="${s?"closeSession":"reset"}" data-form-id="${s?"closeSession":"reset"}">
          <button class="primary-button" type="submit"${a(s?"closeSession":"reset")}>${r(s?"closeSession":"reset",n(s?"lobby_leave":"game_play_again"))}</button>
        </form>
      </div>
    </div>
  `}function tt(e){let{bundle:t,panelTab:n,events:o,t:a,escapeHtml:r,busyAttr:s,buttonLabel:c,renderMetric:d,renderEventFeed:i,protocolLabel:l}=e,f=t.lobby;if(!f)return"";let p=t.ui.lobby_slots||[],g=f.invite_key||"",v=p.filter(B=>!B.is_empty).length,y=t.connection.is_host,w=y&&v>=(f.num_players||p.length||1),T=we(o,a);return`
    <section class="lobby10">
      <article class="surface-card lobby10-lead">
        <div class="lobby10-banner">
          <div>
            <p class="eyebrow">${r(a(y?"lobby_host_headline":"lobby_join_headline"))}</p>
            <h2>${r(a("lobby_title"))}</h2>
            <p class="supporting-copy" data-pretext-block="lock-height">${r(T||a("invite_hint"))}</p>
          </div>
          <div class="lobby10-banner-actions">
            ${y?`<form data-api-action="startOnlineMatch" data-form-id="startOnlineMatch"><button class="primary-button" type="submit"${w?"":" disabled"}${s("startOnlineMatch")}>${c("startOnlineMatch",w?a("lobby_start"):`${a("lobby_slots")} ${v}/${f.num_players||p.length}`)}</button></form>`:""}
            <form data-api-action="closeSession" data-form-id="closeSession">
              <button class="ghost-button danger" type="submit"${s("closeSession")}>${c("closeSession",a("lobby_leave"))}</button>
            </form>
          </div>
        </div>
        <div class="lobby10-invite-band">
          <div class="lobby10-invite">
            <span>${r(a("setup_invite"))}</span>
            <code>${r(g||"----")}</code>
          </div>
          ${g?`<button type="button" class="ghost-button strong" data-copy-text="${r(g)}">${r(a("invite_copy"))}</button>`:""}
          <div class="lobby10-network-strip">
            ${d(a("connection_status"),t.connection.status)}
            ${d(a("connection_mode"),t.connection.is_online?a("connection_online"):a("connection_offline"))}
            ${d(a("lobby_slots"),`${v}/${f.num_players||p.length}`)}
          </div>
        </div>
      </article>

      <div class="lobby10-grid">
        <article class="surface-card lobby10-seats">
          <div class="card-head">
            <div>
              <p class="eyebrow">${r(a("lobby_slots"))}</p>
              <h3>${r(a("lobby_slots"))}</h3>
            </div>
            <span class="section-pill">${v}/${f.num_players||p.length}</span>
          </div>
          <div class="lobby10-seat-grid">
            ${p.map(B=>Dn(B,a,r,s,c)).join("")}
          </div>
        </article>

        <aside class="surface-card lobby10-side">
          <div class="card-head">
            <div>
              <p class="eyebrow">${r(a("lobby_events"))}</p>
              <h3>${r(Gn(n,a))}</h3>
            </div>
            <div class="panel-tabs">
              ${$e("pulse",n,a("lobby_events"),r)}
              ${$e("network",n,a("game_network"),r)}
              ${$e("chat",n,a("lobby_chat"),r)}
            </div>
          </div>
          ${Un(n,t,o,a,r,d,i,l,s,c)}
        </aside>
      </div>
    </section>
  `}function Dn(e,t,n,o,a){let r=[e.is_local?t("slot_you"):"",e.is_host?t("slot_host"):"",e.is_provisional_cpu?t("slot_cpu"):"",e.is_connected?t("slot_online"):t("slot_offline")].filter(Boolean);return`
    <section class="lobby10-seat${e.is_local?" lobby10-seat-local":""}">
      <div class="lobby10-seat-head">
        <div>
          <strong>${n(e.name||t("slot_empty"))}</strong>
          <span>#${e.seat+1}</span>
        </div>
        <em>${n(zn(e.status,t))}</em>
      </div>
      <div class="tag-row">${r.map(s=>`<span>${n(s)}</span>`).join("")}</div>
      <div class="lobby10-seat-actions">
        ${e.can_vote_host?`<form data-api-action="sendHostVote" data-form-id="sendHostVote-${e.seat}"><input type="hidden" name="slot" value="${e.seat}"><button class="ghost-button" type="submit"${o(`sendHostVote-${e.seat}`)}>${a(`sendHostVote-${e.seat}`,t("action_vote_host"))}</button></form>`:""}
        ${e.can_request_replacement?`<form data-api-action="requestReplacementInvite" data-form-id="replacement-${e.seat}"><input type="hidden" name="slot" value="${e.seat}"><button class="secondary-button strong" type="submit"${o(`replacement-${e.seat}`)}>${a(`replacement-${e.seat}`,t("action_replacement_invite"))}</button></form>`:""}
      </div>
    </section>
  `}function Un(e,t,n,o,a,r,s,c,d,i){let l=t.connection.network;switch(e){case"network":return`
        <div class="lobby10-panel">
          <div class="telemetry-grid">
            ${r(o("connection_status"),t.connection.status)}
            ${r(o("connection_mode"),t.connection.is_online?o("connection_online"):o("connection_offline"))}
            ${r(o("connection_transport"),l?.transport||"-")}
            ${r(o("connection_protocol"),c(l))}
            ${r(o("connection_backlog"),String(t.diagnostics.event_backlog||0))}
            ${t.lobby?.role?r(o("connection_role"),t.lobby.role):""}
          </div>
          ${t.connection.last_error?`<div class="lobby10-inline-error">${a(`${t.connection.last_error.code}: ${t.connection.last_error.message}`)}</div>`:""}
          ${we(n,o)?`<p class="supporting-copy">${a(we(n,o)||"")}</p>`:""}
        </div>
      `;case"chat":return`
        <div class="lobby10-panel">
          <pre class="event-feed" role="log" aria-live="polite" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${a(s())}</pre>
          <form class="chat-form" data-api-action="sendChat" data-form-id="sendChat">
            <input name="message" type="text" autocomplete="off" placeholder="${a(o("chat_placeholder"))}">
            <button class="secondary-button" type="submit"${d("sendChat")}>${i("sendChat",o("lobby_chat"))}</button>
          </form>
        </div>
      `;default:return`
        <div class="lobby10-panel">
          <pre class="event-feed" role="log" aria-live="polite" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${a(s())}</pre>
          <div class="lobby10-signal-grid">
            ${r(o("connection_status"),t.connection.status)}
            ${r(o("lobby_slots"),`${t.ui.lobby_slots.filter(f=>!f.is_empty).length}/${t.lobby?.num_players||t.ui.lobby_slots.length}`)}
            ${r(o("connection_backlog"),String(t.diagnostics.event_backlog||0))}
          </div>
        </div>
      `}}function $e(e,t,n,o){return`<button class="panel-tab${t===e?" panel-tab-active":""}" type="button" role="tab" aria-selected="${t===e?"true":"false"}" data-panel-tab="lobby:${e}">${o(n)}</button>`}function Gn(e,t){switch(e){case"network":return t("game_network");case"chat":return t("lobby_chat");default:return t("lobby_overview")}}function we(e,t){let n=[...e].reverse().find(o=>o.kind==="failover_promoted"||o.kind==="failover_rejoined");return n?n.kind==="failover_promoted"?t("signal_failover_promoted"):t("signal_failover_rejoined"):""}function zn(e,t){switch(e){case"occupied_online":return t("slot_online");case"occupied_offline":return t("slot_offline");case"provisional_cpu":return t("slot_cpu");default:return t("slot_empty")}}function nt(e){let{locale:t,playerName:n,relayURL:o,transportMode:a,t:r,escapeHtml:s,busyAttr:c,buttonLabel:d,transportOptions:i}=e,l=n||r("name_placeholder"),f=t==="en-US"?"Team 1":"Time 1",p=t==="en-US"?"Team 2":"Time 2";return`
    <section class="setup10">
      <article class="surface-card setup10-hero">
        <div class="setup10-hero-copy">
          <p class="eyebrow">${s(r("app_kicker"))}</p>
          <h2>${s(r("setup_title"))}</h2>
          <p class="setup10-lede" data-pretext-block="lock-height">${s(r("setup_intro"))}</p>
        </div>
        <div class="setup10-atlas">
          <div class="setup10-atlas-header">
            <span class="setup10-atlas-kicker">${s(r("app_kicker"))}</span>
            <strong>${s(r("setup_mode_offline"))} / ${s(r("setup_mode_online"))}</strong>
          </div>
          <div class="setup10-atlas-board">
            <div class="setup10-atlas-seat setup10-atlas-seat-top">CPU-3</div>
            <div class="setup10-atlas-seat setup10-atlas-seat-left">CPU-4</div>
            <div class="setup10-atlas-seat setup10-atlas-seat-right">CPU-2</div>
            <div class="setup10-atlas-seat setup10-atlas-seat-bottom">${s(l)}</div>
            <div class="setup10-atlas-core">
              <span>${s(r("game_vira"))}</span>
              <strong>A<span class="setup10-suit">\u2660</span></strong>
              <span>${s(r("game_stake"))}</span>
              <strong>3</strong>
            </div>
          </div>
          <div class="setup10-atlas-notes">
            <div class="setup10-note">
              <span>${s(f)}</span>
              <strong>${s(l)} \xB7 CPU-3</strong>
            </div>
            <div class="setup10-note">
              <span>${s(p)}</span>
              <strong>CPU-2 \xB7 CPU-4</strong>
            </div>
          </div>
        </div>
      </article>

      <div class="setup10-grid">
        <article class="surface-card setup10-offline">
          <div class="setup10-section-head">
            <div>
              <span class="section-pill">${s(r("setup_mode_offline"))}</span>
              <h3>${s(r("setup_offline_title"))}</h3>
              <p>${s(r("setup_offline_note"))}</p>
            </div>
            <strong class="form-emphasis">${s(r("setup_offline_caption"))}</strong>
          </div>
          <form class="setup10-form" data-api-action="startGame" data-form-id="startGame">
            <div class="field-grid">
              <label>
                <span>${s(r("setup_name"))}</span>
                <input name="name" type="text" value="${s(l)}" autocomplete="off">
              </label>
              <label>
                <span>${s(r("setup_players"))}</span>
                <select name="numPlayers">
                  <option value="2">2</option>
                  <option value="4">4</option>
                </select>
              </label>
            </div>
            <div class="setup10-roster">
              <div class="setup10-roster-seat">
                <span>1</span>
                <strong>${s(l)}</strong>
                <em>${s(f)}</em>
              </div>
              <div class="setup10-roster-seat">
                <span>2</span>
                <strong>CPU-2</strong>
                <em>${s(p)}</em>
              </div>
              <div class="setup10-roster-seat">
                <span>3</span>
                <strong>CPU-3</strong>
                <em>${s(f)}</em>
              </div>
              <div class="setup10-roster-seat">
                <span>4</span>
                <strong>CPU-4</strong>
                <em>${s(p)}</em>
              </div>
            </div>
            <button class="primary-button setup10-launch" type="submit"${c("startGame")}>${d("startGame",r("setup_start"))}</button>
          </form>
        </article>

        <article class="surface-card setup10-online">
          <div class="setup10-section-head">
            <div>
              <span class="section-pill section-pill-hot">${s(r("setup_mode_online"))}</span>
              <h3>${s(r("setup_online_title"))}</h3>
              <p>${s(r("setup_online_note"))}</p>
            </div>
            <strong class="form-emphasis">${s(r("setup_online_caption"))}</strong>
          </div>

          <div class="setup10-online-grid">
            <form class="setup10-pane" data-api-action="startOnlineHost" data-form-id="startOnlineHost">
              <div class="setup10-pane-head">
                <div>
                  <span class="eyebrow">${s(r("setup_mode_online"))}</span>
                  <h4>${s(r("setup_host"))}</h4>
                </div>
              </div>
              <div class="field-grid">
                <label>
                  <span>${s(r("setup_name"))}</span>
                  <input name="name" type="text" value="${s(l)}" autocomplete="off">
                </label>
                <label>
                  <span>${s(r("setup_players"))}</span>
                  <select name="numPlayers">
                    <option value="2">2</option>
                    <option value="4">4</option>
                  </select>
                </label>
              </div>
              <label>
                <span>${s(r("setup_transport"))}</span>
                <select name="transport_mode">
                  ${i(a)}
                </select>
              </label>
              <label>
                <span>${s(r("setup_relay"))}</span>
                <input name="relay_url" type="text" value="${s(o)}" placeholder="${s(r("relay_placeholder"))}" autocomplete="off">
              </label>
              <button class="secondary-button" type="submit"${c("startOnlineHost")}>${d("startOnlineHost",r("setup_host"))}</button>
            </form>

            <form class="setup10-pane" data-api-action="joinOnline" data-form-id="joinOnline">
              <div class="setup10-pane-head">
                <div>
                  <span class="eyebrow">${s(r("setup_mode_online"))}</span>
                  <h4>${s(r("setup_join"))}</h4>
                </div>
              </div>
              <div class="field-grid">
                <label>
                  <span>${s(r("setup_name"))}</span>
                  <input name="name" type="text" value="${s(l)}" autocomplete="off">
                </label>
                <label>
                  <span>${s(r("setup_invite"))}</span>
                  <input name="key" type="text" autocomplete="off">
                </label>
              </div>
              <label>
                <span>${s(r("setup_role"))}</span>
                <select name="role">
                  <option value="auto">${s(r("role_auto"))}</option>
                  <option value="partner">${s(r("role_partner"))}</option>
                  <option value="opponent">${s(r("role_opponent"))}</option>
                </select>
              </label>
              <button class="secondary-button strong" type="submit"${c("joinOnline")}>${d("joinOnline",r("setup_join"))}</button>
            </form>
          </div>
          <p class="supporting-copy setup10-support">${s(r("setup_online_support"))}</p>
        </article>
      </div>
    </section>
  `}function Pe(e){switch(e){case"host_lobby":case"client_lobby":return"lobby";case"offline_match":case"host_match":case"client_match":return"game";default:return"setup"}}function rt(e){switch(e){case"startGame":return["offline_match"];case"startOnlineHost":return["host_lobby"];case"joinOnline":return["client_lobby","client_match"];case"startOnlineMatch":return["host_match"];case"closeSession":case"reset":return["idle"];default:return[]}}function at(e,t,n){return n==="snapshot"?!0:t>=e}function st(e){let t=Pe(e?.mode||"idle");if(t==="lobby"&&!e?.lobby)return"waiting_lobby";if(t!=="game")return null;let n=e?.match;return!n||!n.CurrentHand||!Array.isArray(n.Players)||n.Players.length===0?"waiting_match":null}var Kn="truco:runtime:update";async function ot(){return Ee().Snapshot()}async function it(e){let t=await Ee().RuntimeUpdateEventName(),n=window.runtime;if(!n)throw new Error("Wails runtime unavailable");return n.EventsOn(t||Kn,(...o)=>{let a=o[0];a?.bundle&&e(a)})}async function Ce(e,t){let n=Ee();switch(e){case"setLocale":return n.SetLocale(qn(t.locale));case"startGame":return n.StartOfflineGame(Q(t.name),Y(t.numPlayers,2));case"startOnlineHost":return n.CreateHostSession(Q(t.name),Y(t.numPlayers,2),"",Q(t.relay_url),Vn(t.transport_mode));case"joinOnline":return n.JoinSession(Q(t.key),Q(t.name),Q(t.role)||"auto");case"startOnlineMatch":return n.StartHostedMatch();case"sendChat":return n.SendChat(Q(t.message));case"sendHostVote":return n.VoteHost(Y(t.slot,0));case"requestReplacementInvite":return n.RequestReplacementInvite(Y(t.slot,0));case"play":return t.faceDown===!0?n.PlayFaceDownCard(Y(t.cardIndex,-1)):n.PlayCard(Y(t.cardIndex,-1));case"truco":return n.RequestTruco();case"accept":return n.AcceptTruco();case"refuse":return n.RefuseTruco();case"newHand":return n.NewHand();case"tick":return n.Tick(Y(t.maxSteps,12));case"closeSession":return n.CloseSession();case"reset":return n.Reset();default:throw new Error(`unsupported action: ${e}`)}}async function lt(e){if(e){if(window.runtime?.ClipboardSetText){await window.runtime.ClipboardSetText(e);return}if(navigator.clipboard){await navigator.clipboard.writeText(e);return}throw new Error("clipboard unavailable")}}function Ee(){let e=window.go?.main?.App;if(!e)throw new Error("Wails bridge unavailable");return e}function qn(e){return e==="en-US"?"en-US":"pt-BR"}function Q(e){return typeof e=="string"?e.trim():""}function Y(e,t){return typeof e=="number"&&Number.isFinite(e)?e:typeof e=="string"&&/^-?\d+$/.test(e)?Number.parseInt(e,10):t}function Vn(e){switch(e){case"tcp_tls":return"tcp_tls";case"relay_quic_v2":return"relay_quic_v2";default:return""}}var Be="truco-wails-locale",mt="truco-wails-player-name",gt="truco-wails-relay-url",ft="truco-wails-transport-mode",Jn=80,Qn=4e3,Yn=2500,ht=document.querySelector("#app");if(!ht)throw new Error("missing #app root");var Z=ht,ct=new Map,ut=0,u={locale:At(localStorage.getItem(Be)),bundle:null,events:[],playerName:localStorage.getItem(mt)||"",relayURL:localStorage.getItem(gt)||"",transportMode:localStorage.getItem(ft)||"",error:"",busyForm:"",initialized:!1,diagnosticsOpen:!1,lobbyPanelTab:"pulse",gamePanelTab:"pulse",pendingAction:"",lastSubmittedAction:"",lastExpectedModes:[],lastSeenSequence:0,lastMode:"idle",lastRefreshState:"idle",lastRefreshError:"",lastRenderError:""};Z.addEventListener("submit",e=>{let t=e.target?.closest("form[data-api-action]");t&&(e.preventDefault(),tr(t))});Z.addEventListener("click",e=>{let t=e.target;if(!t)return;let n=t.closest("[data-copy-text]");if(n){Pr(n);return}let o=t.closest("[data-client-action]");if(o){e.preventDefault(),nr(o.dataset.clientAction);return}let a=t.closest("[data-panel-tab]");a&&(e.preventDefault(),rr(a.dataset.panelTab||""))});Z.addEventListener("change",e=>{let t=e.target;!t||t.name!=="locale"||er(At(t.value))});window.addEventListener("resize",()=>{window.clearTimeout(ut),ut=window.setTimeout(()=>St(),90)});Zn();async function Zn(){U();try{await it(e=>{Xn(e),e.bundle.connection.last_error||(u.error=""),U()}),await ie(),u.initialized=!0,u.error=""}catch(e){u.error=ne(e)}U()}async function ie(){let e=await ot();_t(e,"snapshot")}function Xn(e){_t(e.bundle,"event")&&e.events&&e.events.length>0&&(u.events=[...u.events,...e.events].slice(-Jn))}async function er(e){u.locale=e,localStorage.setItem(Be,e),document.documentElement.lang=e==="en-US"?"en":"pt-BR",U();try{let t=await Ce("setLocale",{locale:e});if(t?.message)throw await ie(),new Error(t.message);u.error=""}catch(t){u.error=ne(t)}U()}async function tr(e){let t=e.dataset.apiAction;if(!t||u.busyForm)return;let n=$r(e),o=e.dataset.formId||t,a=Lr(t);u.busyForm=o,u.pendingAction=t,u.lastSubmittedAction=t,u.lastExpectedModes=a.modes,u.lastRenderError="",u.lastRefreshState="idle",u.lastRefreshError="",u.error="",U();try{wr(n);let r=await vt(Ce(t,n),Qn,m("status_action_timeout",a.timeoutLabel));if(r?.message)throw await pt(t,a,!0),new Error(r.message);await pt(t,a,!1),u.error=""}catch(r){u.error=ne(r)}finally{u.busyForm="",u.pendingAction="",U()}}async function nr(e){if(e)switch(e){case"refresh":try{await ie(),u.error=""}catch(t){u.error=ne(t)}U();return;case"toggle-diagnostics":u.diagnosticsOpen=!u.diagnosticsOpen,U();return;default:return}}function rr(e){let[t,n]=e.split(":");if(t==="lobby"&&(n==="pulse"||n==="network"||n==="chat")){u.lobbyPanelTab=n,U();return}t==="game"&&(n==="pulse"||n==="network"||n==="chat")&&(u.gamePanelTab=n,U())}function ar(){return Pe(u.bundle?.mode||"idle")}function sr(){let e=u.bundle?.mode||"";return e.startsWith("host_")||e.startsWith("client_")}function U(){let e=or();Z.innerHTML=lr(),ir(e),St()}function or(){let e=document.activeElement;if(!(e instanceof HTMLInputElement||e instanceof HTMLTextAreaElement||e instanceof HTMLSelectElement)||!Z.contains(e))return null;let t=e.id||e.name||e.getAttribute("data-focus-key");if(!t)return null;let n=e.closest("form")?.getAttribute("data-form-id");return{selector:n&&e.name?`form[data-form-id="${CSS.escape(n)}"] [name="${CSS.escape(e.name)}"]`:e.id?`#${CSS.escape(e.id)}`:`[name="${CSS.escape(e.name||t)}"]`,value:e.value,selectionStart:"selectionStart"in e?e.selectionStart:null,selectionEnd:"selectionEnd"in e?e.selectionEnd:null}}function ir(e){if(!e)return;let t=Z.querySelector(e.selector);t&&(e.value!==void 0&&t.value!==e.value&&(t.value=e.value),t.focus({preventScroll:!0}),(t instanceof HTMLInputElement||t instanceof HTMLTextAreaElement)&&e.selectionStart!==null&&e.selectionEnd!==null&&t.setSelectionRange(e.selectionStart,e.selectionEnd))}function lr(){return`
    <div class="page-shell">
      <div class="page-aura page-aura-left"></div>
      <div class="page-aura page-aura-right"></div>
      <main class="app-shell">
        <header class="hero-card">
          <div class="hero-copy">
            <p class="eyebrow">${$(m("app_kicker"))}</p>
            <div class="hero-title-row">
              <h1>${$(m("title_main"))}</h1>
              <span class="hero-stamp">${$(m("app_stamp"))}</span>
            </div>
            <p class="hero-subtitle" data-pretext-block="lock-height">${$(m("title_sub"))}</p>
          </div>
          <div class="hero-tools">
            <form class="locale-card">
              <label for="locale-select">${$(m("locale_label"))}</label>
              <select id="locale-select" name="locale">
                ${Ar(u.locale)}
              </select>
            </form>
            <div class="desktop-action-row">
              <button class="ghost-button" type="button" data-client-action="refresh">${$(m("header_resync"))}</button>
              <button class="ghost-button strong" type="button" data-client-action="toggle-diagnostics">
                ${$(u.diagnosticsOpen?m("header_hide_diagnostics"):m("header_diagnostics"))}
              </button>
            </div>
          </div>
        </header>
        ${cr()}
        ${dr()}
        ${br()}
      </main>
    </div>
  `}function cr(){let e=["runtime-banner"],t="",n=ur();return u.pendingAction?(e.push("runtime-banner-info"),t=m("status_action_pending",ae(u.pendingAction))):u.lastRenderError?(e.push("runtime-banner-warning"),t=`${m("status_render_recovery")} ${u.lastRenderError}`):Me()?(e.push("runtime-banner-danger"),t=Me()||""):u.lastRefreshState==="stale"?(e.push("runtime-banner-warning"),t=m("status_refresh_stale",yt(u.lastExpectedModes))):u.lastRefreshState==="error"&&u.lastRefreshError?(e.push("runtime-banner-warning"),t=u.lastRefreshError):n&&(e.push("runtime-banner-info"),t=n),t?`
    <section class="${e.join(" ")}" role="${u.lastRenderError||u.lastRefreshState==="error"?"alert":"status"}" aria-live="${u.lastRenderError||u.lastRefreshState==="error"?"assertive":"polite"}" data-pretext-block="lock-height">
      <div class="runtime-banner-copy">${$(t)}</div>
      <div class="runtime-banner-actions">
        <button class="ghost-button strong" type="button" data-client-action="refresh">${$(m("header_resync"))}</button>
        <button class="ghost-button" type="button" data-client-action="toggle-diagnostics">
          ${$(u.diagnosticsOpen?m("header_hide_diagnostics"):m("header_diagnostics"))}
        </button>
      </div>
    </section>
  `:""}function ur(){let e=[...u.events].reverse().find(t=>t.kind==="failover_promoted"||t.kind==="failover_rejoined"||t.kind==="replacement_invite");if(!e)return"";switch(e.kind){case"failover_promoted":return m("signal_failover_promoted");case"failover_rejoined":return m("signal_failover_rejoined");case"replacement_invite":return m("signal_replacement_ready");default:return""}}function dr(){try{return pr()}catch(e){return u.lastRenderError=ne(e),hr()}}function pr(){if(!u.initialized&&!u.bundle&&!Me())return`<section class="surface-card loading-card"><div class="loading-pip"></div><strong>${$(m("button_busy"))}</strong></section>`;let e=st(u.bundle);if(e==="waiting_lobby")return Ne(m("lobby_title"),m("status_waiting_lobby"));if(e==="waiting_match")return Ne(m("game_title_offline"),m("status_waiting_match"));switch(ar()){case"lobby":return gr();case"game":return fr();default:return mr()}}function mr(){return nt({locale:u.locale,playerName:u.playerName,relayURL:u.relayURL,transportMode:u.transportMode,t:m,escapeHtml:$,busyAttr:le,buttonLabel:ce,transportOptions:xr})}function gr(){let e=xt();return tt({bundle:e,panelTab:u.lobbyPanelTab,events:u.events,t:m,escapeHtml:$,busyAttr:le,buttonLabel:ce,renderMetric:D,renderEventFeed:bt,protocolLabel:kt})}function fr(){let e=xt();return Ze({bundle:e,panelTab:u.gamePanelTab,events:u.events,isOnlineMode:sr(),t:m,escapeHtml:$,busyAttr:le,buttonLabel:ce,renderMetric:D,renderEventFeed:bt,renderCard:_r,protocolLabel:kt,cardLabel:Br,playerName:Lt,teamScore:Er,localTeam:Tr,nextStake:Nr,raiseLabel:Mr,lastTrickCopy:Cr,seatPositions:Rr})}function Ne(e,t){return`
    <section class="surface-card loading-card">
      <div class="card-head">
        <div>
          <p class="eyebrow">${$(m("game_runtime_stale_title"))}</p>
          <h3>${$(e)}</h3>
        </div>
      </div>
      <p class="supporting-copy" data-pretext-block="lock-height">${$(t)}</p>
      <div class="desktop-action-row">
        <button class="ghost-button strong" type="button" data-client-action="refresh">${$(m("header_resync"))}</button>
        <button class="ghost-button" type="button" data-client-action="toggle-diagnostics">${$(m("header_diagnostics"))}</button>
      </div>
    </section>
  `}function hr(){return Ne(m("game_runtime_stale_title"),m("status_render_recovery_copy"))}function br(){if(!u.diagnosticsOpen||!u.bundle)return"";let e=u.bundle,t=e.connection.last_error,n=e.diagnostics.event_log||[];return`
    <section class="surface-card diagnostics-card">
      <div class="card-head">
        <div>
          <p class="eyebrow">${$(m("header_diagnostics"))}</p>
          <h3>${$(m("diagnostics_title"))}</h3>
        </div>
        <div class="desktop-action-row">
          <button class="ghost-button" type="button" data-client-action="refresh">${$(m("header_resync"))}</button>
          <form data-api-action="tick" data-form-id="tick">
            <input type="hidden" name="maxSteps" value="12">
            <button class="ghost-button strong" type="submit"${le("tick")}>${ce("tick",m("diagnostics_force_tick"))}</button>
          </form>
        </div>
      </div>
      <div class="diagnostics-grid">
        ${D(m("diagnostics_versions"),`core ${e.versions.core_api_version} \xB7 protocol ${e.versions.protocol_version} \xB7 schema ${e.versions.snapshot_schema_version}`)}
        ${D(m("connection_backlog"),String(e.diagnostics.event_backlog||0))}
        ${D(m("connection_status"),e.connection.status)}
        ${D(m("connection_transport"),e.connection.network?.transport||"-")}
        ${D(m("diagnostics_mode"),u.lastMode||e.mode||"idle")}
        ${D(m("diagnostics_sequence"),String(u.lastSeenSequence))}
        ${D(m("diagnostics_last_action"),u.lastSubmittedAction?ae(u.lastSubmittedAction):"-")}
        ${D(m("diagnostics_refresh"),u.lastRefreshState)}
        ${t?D(m("event_error"),`${t.code}: ${t.message}`):""}
        ${u.lastRefreshError?D(m("diagnostics_refresh_error"),u.lastRefreshError):""}
        ${u.lastRenderError?D(m("diagnostics_render_error"),u.lastRenderError):""}
      </div>
      <div class="diagnostics-log-shell">
        <strong>${$(m("diagnostics_event_log"))}</strong>
        <pre class="diagnostics-log" data-pretext-block="lock-height" data-pretext-whitespace="pre-wrap">${$(n.length>0?n.slice(-16).join(`
`):m("diagnostics_none"))}</pre>
      </div>
    </section>
  `}function _r(e,t="regular"){let n=e.Suit==="Copas"||e.Suit==="Ouros";return`
    <span class="card-face card-face-${t}${n?" card-face-red":""}">
      <span class="card-corner">${$(e.Rank)}${$(Te(e.Suit))}</span>
      <span class="card-center">${$(Te(e.Suit))}</span>
      <span class="card-corner card-corner-bottom">${$(e.Rank)}${$(Te(e.Suit))}</span>
    </span>
  `}function D(e,t){return`<div class="metric"><span>${$(e)}</span><strong>${$(t)}</strong></div>`}function bt(e=[]){let t=u.events.map(yr),n=e.map(a=>vr(a)),o=n.length>0?[...n,...t.slice(-8)]:t;return o.length===0?m("game_no_events"):o.slice(-12).join(`
`)}function yr(e){let t=e.timestamp?e.timestamp.slice(11,19):"--:--:--",n=e.payload||{},o=m(`event_${e.kind}`),a=typeof n.author=="string"?n.author:"",r=typeof n.text=="string"?n.text:"",s=typeof n.message=="string"?n.message:"",c=typeof n.invite_key=="string"?n.invite_key:"",d=[a&&r?`${a}: ${r}`:"",r&&!a?r:"",s,c].filter(Boolean).join(" \xB7 ");return`[${t}] ${o}${d?` \xB7 ${d}`:""}`}function vr(e){let t=e.match(/^Nova mão: vira (.+), manilha (.+)\.$/);if(t)return m("game_log_new_hand",dt(t[1]),t[2]);let n=e.match(/^(.+) jogou carta virada\.$/);if(n)return m("game_log_face_down",n[1]);let o=e.match(/^(.+) jogou (.+)\.$/);if(o)return m("game_log_played",o[1],dt(o[2]));let a=e.match(/^(.+) aceitou (.+)\. Aposta agora vale (\d+)\.$/);if(a)return m("game_log_accept",a[1],Sr(a[2]),a[3]);let r=e.match(/^(.+) correu do truco\.$/);return r?m("game_log_fold",r[1]):e}function dt(e){let t=e.match(/^(.+) de (Ouros|Espadas|Copas|Paus)$/);return t?m("card_of",t[1],m(`suit_${t[2]}`)):e}function Sr(e){switch(e.toLowerCase()){case"truco":return m("raise_3");case"seis":return m("raise_6");case"nove":return m("raise_9");case"doze":return m("raise_12");default:return e}}function _t(e,t){let n=e.connection?.last_event_sequence||0;if(!at(u.lastSeenSequence,n,t))return!1;let o=u.lastMode;return u.bundle=e,u.locale=e.locale||u.locale,u.lastSeenSequence=Math.max(u.lastSeenSequence,n),u.lastMode=e.mode||"idle",document.documentElement.lang=u.locale==="en-US"?"en":"pt-BR",localStorage.setItem(Be,u.locale),u.lastMode!==o&&((u.lastMode==="idle"||u.lastMode==="host_lobby"||u.lastMode==="client_lobby")&&(u.lobbyPanelTab="pulse"),(u.lastMode==="idle"||u.lastMode==="offline_match"||u.lastMode==="host_match"||u.lastMode==="client_match")&&(u.gamePanelTab="pulse")),u.lastMode==="idle"&&(u.pendingAction=""),!0}async function pt(e,t,n){try{await vt(ie(),Yn,m("status_snapshot_timeout")),u.lastRefreshState="ok",u.lastRefreshError=""}catch(a){throw u.lastRefreshState="error",u.lastRefreshError=m("status_refresh_failed",ae(e),ne(a)),new Error(u.lastRefreshError)}if(n)return;let o=u.bundle?.mode||"idle";if(t.modes.length>0&&!t.modes.includes(o))throw u.lastRefreshState="stale",new Error(m("status_transition_failed",ae(e),yt(t.modes),o))}function Lr(e){return{modes:rt(e),timeoutLabel:ae(e)}}function ae(e){switch(e){case"startGame":return m("setup_offline_title");case"startOnlineHost":return m("setup_host");case"joinOnline":return m("setup_join");case"startOnlineMatch":return m("lobby_start");case"sendChat":return m("lobby_chat");case"closeSession":case"reset":return m("lobby_leave");default:return e}}function yt(e){return e.length===0?"-":e.join(", ")}async function vt(e,t,n){let o=0;try{return await Promise.race([e,new Promise((a,r)=>{o=window.setTimeout(()=>r(new Error(n)),t)})])}finally{window.clearTimeout(o)}}function St(){let e=Z.querySelectorAll("[data-pretext-block]");for(let t of e){let n=t.clientWidth;if(n<=0)continue;let o=window.getComputedStyle(t),a=t.dataset.pretextWhitespace==="pre-wrap"?"pre-wrap":"normal",r=`${o.font}|${a}|${t.textContent||""}`,s=ct.get(r);s||(s=Ve(t.textContent||"",o.font,{whiteSpace:a}),ct.set(r,s));let c=kr(o),d=Je(s,n,c);t.style.minHeight=`${Math.ceil(d.height)}px`,t.dataset.pretextLines=String(d.lineCount)}}function kr(e){let t=Number.parseFloat(e.lineHeight);if(Number.isFinite(t))return t;let n=Number.parseFloat(e.fontSize);return Number.isFinite(n)?n*1.3:22}function Ar(e){return[["pt-BR","Portugu\xEAs (BR)"],["en-US","English (US)"]].map(([t,n])=>`<option value="${t}"${e===t?" selected":""}>${n}</option>`).join("")}function xr(e){return[["",m("transport_auto")],["tcp_tls",m("transport_direct")],["relay_quic_v2",m("transport_relay")]].map(([t,n])=>`<option value="${t}"${e===t?" selected":""}>${$(n)}</option>`).join("")}function $r(e){let t={};return new FormData(e).forEach((o,a)=>{if(typeof o=="string"){if(o==="true"){t[a]=!0;return}if(o==="false"){t[a]=!1;return}if(/^-?\d+$/.test(o)){t[a]=Number.parseInt(o,10);return}t[a]=o}}),t}function wr(e){typeof e.name=="string"&&e.name.trim()!==""&&(u.playerName=e.name.trim(),localStorage.setItem(mt,u.playerName)),typeof e.relay_url=="string"&&(u.relayURL=e.relay_url,localStorage.setItem(gt,u.relayURL)),typeof e.transport_mode=="string"&&(u.transportMode=e.transport_mode,localStorage.setItem(ft,u.transportMode))}async function Pr(e){let t=e.dataset.copyText||"";if(!t)return;await lt(t);let n=e.textContent||m("invite_copy");e.textContent=m("copy_done"),window.setTimeout(()=>{e.textContent=n},1e3)}function le(e){return u.busyForm===e?" disabled":""}function ce(e,t){return u.busyForm===e?m("button_busy"):t}function Me(){if(u.error)return u.error;let e=u.bundle?.connection.last_error;return e?.message?e.code?`${e.code}: ${e.message}`:e.message:null}function Cr(e){return e.LastTrickRound<=0?m("status_idle"):e.LastTrickTie?m("trick_tie",e.LastTrickRound):m("trick_win",Lt(e,e.LastTrickWinner),e.LastTrickRound)}function Lt(e,t){return e.Players.find(n=>n.ID===t)?.Name||"?"}function Er(e,t){return e.MatchPoints[String(t)]||0}function Tr(e,t){let n=t.ui.actions.local_player_id>=0?t.ui.actions.local_player_id:e.CurrentPlayerIdx;return e.Players.find(o=>o.ID===n)?.Team||0}function Nr(e){switch(e){case 1:return 3;case 3:return 6;case 6:return 9;case 9:return 12;default:return e}}function Mr(e){switch(e){case 3:return m("raise_3");case 6:return m("raise_6");case 9:return m("raise_9");case 12:return m("raise_12");default:return String(e)}}function Br(e){return m("card_of",e.Rank,m(`suit_${e.Suit}`))}function Te(e){switch(e){case"Ouros":return"\u2666";case"Espadas":return"\u2660";case"Copas":return"\u2665";case"Paus":return"\u2663";default:return"?"}}function kt(e){if(!e)return"-";if(e.negotiated_protocol_version)return`v${e.negotiated_protocol_version}`;let t=Object.values(e.seat_protocol_versions||{}).filter(n=>n>0);return t.length===0?"-":Array.from(new Set(t)).sort((n,o)=>n-o).map(n=>`v${n}`).join(", ")}function Rr(e,t){let n=new Map,o=t.ui.actions.local_player_id>=0?t.ui.actions.local_player_id:e.CurrentPlayerIdx,a=e.NumPlayers===2?["bottom","top"]:["bottom","right","top","left"];for(let r of e.Players){let s=(r.ID-o+e.NumPlayers)%e.NumPlayers;n.set(r.ID,a[s]||"top")}return n}function At(e){return e==="en-US"?"en-US":"pt-BR"}function xt(){if(!u.bundle)throw new Error("bundle unavailable");return u.bundle}function m(e,...t){return Qe(u.locale,e,...t)}function ne(e){return e instanceof Error?e.message:String(e||"unknown error")}function $(e){return e.replaceAll("&","&amp;").replaceAll("<","&lt;").replaceAll(">","&gt;").replaceAll('"',"&quot;").replaceAll("'","&#39;")}
