import{d as Ne,j as L,O as Pl}from"./react-vendor-CUAHX7bH.js";const Pe={BLOCK_CHANGE:1,ENTITY_SPAWN:2,ENTITY_MOVE:3,ENTITY_REMOVE:4,PLAYER_SPAWN:5,PLAYER_MOVE:6,PLAYER_REMOVE:7,PLAYER_ANIM:8,CHAT:9,PLAYER_EQUIPMENT:10,PLAYER_EQUIPMENT_FULL:11,PLAYER_HEALTH:12,PLAYER_EFFECTS:13,PLAYER_INVENTORY:14,PLAYER_SKIN:15,PLAYER_BLOCK_PLACE:16,PLAYER_BLOCK_BREAK:17},Xi={SWING_MAIN_ARM:0,SNEAK_START:2,SNEAK_STOP:3,DAMAGE:4},Yi=[77,79,68,76],Ps=[31,139],Ds=new TextDecoder("utf-8");function Dl(i){if(i.byteLength<2)return!1;const e=new Uint8Array(i,0,2);return e[0]===Ps[0]&&e[1]===Ps[1]}async function Il(i){if(typeof DecompressionStream>"u")throw new Error("DecompressionStream API not available — cannot decompress gzip replay file");return new Response(new Blob([i]).stream().pipeThrough(new DecompressionStream("gzip"))).arrayBuffer()}async function Nl(i){return Dl(i)&&(i=await Il(i)),Ul(i)}function Ul(i){const e=new DataView(i);let t=0;function r(){const U=e.getUint8(t);return t+=1,U}function n(){const U=e.getInt16(t);return t+=2,U}function a(){const U=e.getUint16(t);return t+=2,U}function o(){const U=e.getInt32(t);return t+=4,U}function s(){const U=e.getFloat32(t);return t+=4,U}function l(){const U=e.getBigInt64(t);return t+=8,U}function c(){const U=r(),F=new Uint8Array(i,t,U);return t+=U,Ds.decode(F)}function h(){const U=a(),F=new Uint8Array(i,t,U);return t+=U,Ds.decode(F)}function u(){const U=l(),F=l(),K=$=>($<0n&&($=$+(1n<<64n)),$.toString(16).padStart(16,"0")),X=K(U)+K(F);return`${X.slice(0,8)}-${X.slice(8,12)}-${X.slice(12,16)}-${X.slice(16,20)}-${X.slice(20,32)}`}function p(U){const F=new Uint8Array(i,t,U);return t+=U,F.slice()}function f(U){return U*360/32767}function g(){return S?s():f(n())}const _=[r(),r(),r(),r()];if(_[0]!==Yi[0]||_[1]!==Yi[1]||_[2]!==Yi[2]||_[3]!==Yi[3])throw new Error("Invalid magic bytes: not a .modlreplay file");const m=a(),d=m>=3,S=m>=4,T=Number(l()),x=h(),N=o(),w=o(),A=o(),O=o(),b={version:m,startTime:T,mcVersion:x,targetX:N,targetY:w,targetZ:A,radiusBlocks:O},M=o(),R=new Array(M);for(let U=0;U<M;U++)R[U]={x:o(),y:n(),z:o(),stateId:o()};const H=[];for(;t<i.byteLength;){const U=r(),F=o(),K=S?a():0,X=t;switch(U){case Pe.BLOCK_CHANGE:H.push({type:Pe.BLOCK_CHANGE,timestampDeltaMs:F,x:o(),y:n(),z:o(),stateId:o()});break;case Pe.ENTITY_SPAWN:{const $=o(),k=n(),Z=s(),ce=s(),ve=s(),De=a(),xe=p(De);H.push({type:Pe.ENTITY_SPAWN,timestampDeltaMs:F,entityId:$,entityTypeId:k,x:Z,y:ce,z:ve,metadata:xe});break}case Pe.ENTITY_MOVE:H.push({type:Pe.ENTITY_MOVE,timestampDeltaMs:F,entityId:o(),x:s(),y:s(),z:s(),yaw:g(),pitch:g()});break;case Pe.ENTITY_REMOVE:H.push({type:Pe.ENTITY_REMOVE,timestampDeltaMs:F,entityId:o()});break;case Pe.PLAYER_SPAWN:{const $=u(),k=c(),Z=s(),ce=s(),ve=s(),De=g(),xe=g(),G=a(),J=p(G);H.push({type:Pe.PLAYER_SPAWN,timestampDeltaMs:F,uuid:$,playerName:k,x:Z,y:ce,z:ve,yaw:De,pitch:xe,equipment:J});break}case Pe.PLAYER_MOVE:H.push({type:Pe.PLAYER_MOVE,timestampDeltaMs:F,uuid:u(),x:s(),y:s(),z:s(),yaw:g(),pitch:g()});break;case Pe.PLAYER_REMOVE:H.push({type:Pe.PLAYER_REMOVE,timestampDeltaMs:F,uuid:u()});break;case Pe.PLAYER_ANIM:H.push({type:Pe.PLAYER_ANIM,timestampDeltaMs:F,uuid:u(),animationType:r()});break;case Pe.CHAT:H.push({type:Pe.CHAT,timestampDeltaMs:F,uuid:u(),message:h()});break;case Pe.PLAYER_EQUIPMENT:H.push({type:Pe.PLAYER_EQUIPMENT,timestampDeltaMs:F,uuid:u(),itemName:c()});break;case Pe.PLAYER_EQUIPMENT_FULL:{const $=u(),k=r(),Z=[];for(let ce=0;ce<k;ce++){const ve=r(),De=c();let xe;if(d){const G=r();if(G>0){xe=[];for(let J=0;J<G;J++)xe.push({id:c(),level:r()})}}Z.push({slotId:ve,itemName:De,enchantments:xe})}H.push({type:Pe.PLAYER_EQUIPMENT_FULL,timestampDeltaMs:F,uuid:$,slots:Z});break}case Pe.PLAYER_HEALTH:{const $=u(),k=s(),Z=r(),ce=s();H.push({type:Pe.PLAYER_HEALTH,timestampDeltaMs:F,uuid:$,health:k,food:Z,saturation:ce});break}case Pe.PLAYER_EFFECTS:{const $=u(),k=r(),Z=[];for(let ce=0;ce<k;ce++)Z.push({effectTypeId:r(),amplifier:r(),durationTicks:o()});H.push({type:Pe.PLAYER_EFFECTS,timestampDeltaMs:F,uuid:$,effects:Z});break}case Pe.PLAYER_INVENTORY:{const $=u(),k=r()!==0,Z=r(),ce=[];for(let ve=0;ve<Z;ve++){const De=r(),xe=h(),G=r();let J;if(d){const q=r();if(q>0){J=[];for(let z=0;z<q;z++)J.push({id:c(),level:r()})}}ce.push({slotIndex:De,itemName:xe,count:G,enchantments:J})}H.push({type:Pe.PLAYER_INVENTORY,timestampDeltaMs:F,uuid:$,isFullSnapshot:k,slots:ce});break}case Pe.PLAYER_SKIN:{const $=u(),k=o(),Z=p(k);H.push({type:Pe.PLAYER_SKIN,timestampDeltaMs:F,uuid:$,skinPng:Z});break}case Pe.PLAYER_BLOCK_PLACE:{const $=u(),k=o(),Z=n(),ce=o(),ve=o();H.push({type:Pe.PLAYER_BLOCK_PLACE,timestampDeltaMs:F,placerUuid:$,x:k,y:Z,z:ce,stateId:ve});break}case Pe.PLAYER_BLOCK_BREAK:{const $=u(),k=o(),Z=n(),ce=o(),ve=o();H.push({type:Pe.PLAYER_BLOCK_BREAK,timestampDeltaMs:F,breakerUuid:$,x:k,y:Z,z:ce,previousStateId:ve});break}default:S?(console.warn(`[ReplayParser] Unknown event type: 0x${U.toString(16)} at offset ${X-7}, skipping ${K} bytes`),t=X+K):(console.warn(`[ReplayParser] Unknown event type: 0x${U.toString(16)} at offset ${t-5}, skipping rest of file`),t=i.byteLength);break}}return{header:b,snapshot:R,events:H}}class Ol{constructor(e){this.currentIndex=0,this.initialBlocks=new Map,this.currentBlocks=new Map;for(const t of e.snapshot){const r=`${t.x},${t.y},${t.z}`;this.initialBlocks.set(r,t.stateId),this.currentBlocks.set(r,t.stateId)}this.blockChanges=e.events.filter(t=>t.type===Pe.BLOCK_CHANGE)}advanceTo(e){const t=[];if(this.needsReset(e))return this.resetAndReplayTo(e);for(;this.currentIndex<this.blockChanges.length;){const r=this.blockChanges[this.currentIndex];if(r.timestampDeltaMs>e)break;const n=`${r.x},${r.y},${r.z}`;this.currentBlocks.set(n,r.stateId),t.push({x:r.x,y:r.y,z:r.z,stateId:r.stateId}),this.currentIndex++}return t}needsReset(e){if(this.currentIndex===0)return!1;const t=this.currentIndex>0?this.blockChanges[this.currentIndex-1].timestampDeltaMs:0;return e<t}resetAndReplayTo(e){return this.currentBlocks=new Map(this.initialBlocks),this.currentIndex=0,this.advanceTo(e)}getBlockState(e,t,r){return this.currentBlocks.get(`${e},${t},${r}`)}getAllBlocks(){const e=[];for(const[t,r]of this.currentBlocks){const[n,a,o]=t.split(",").map(Number);e.push({x:n,y:a,z:o,stateId:r})}return e}}function Bn(i,e,t){return i+(e-i)*t}function qi(i,e,t){let r=e-i;for(;r>180;)r-=360;for(;r<-180;)r+=360;return i+r*t}function Is(i,e,t){const r=e.timestampDeltaMs-i.timestampDeltaMs;if(r<=0)return{x:i.x,y:i.y,z:i.z};const n=Math.max(0,Math.min(1,(t-i.timestampDeltaMs)/r));return{x:Bn(i.x,e.x,n),y:Bn(i.y,e.y,n),z:Bn(i.z,e.z,n)}}function Ns(i,e){let t=0,r=i.length-1,n=-1;for(;t<=r;){const a=t+r>>>1;i[a].timestampDeltaMs<=e?(n=a,t=a+1):r=a-1}return n}class Fl{constructor(e){this.playing=!1,this.currentTimeMs=0,this.totalDurationMs=0,this.rate=1,this.followingPlayer=null,this.listeners=new Set,this.animationFrameId=null,this.lastFrameTime=null,this.entities=new Map,this.players=new Map,this.playerSkins=new Map,this.playerMoveEvents=new Map,this.entityMoveEvents=new Map,this.currentEventIndex=0,this.tick=t=>{if(this.playing){if(this.lastFrameTime!==null){const r=(t-this.lastFrameTime)*this.rate;this.currentTimeMs=Math.min(this.currentTimeMs+r,this.totalDurationMs),this.processEventsTo(this.currentTimeMs);const n=this.worldState.advanceTo(this.currentTimeMs);if(this.notifyListeners(n),this.currentTimeMs>=this.totalDurationMs){this.pause();return}}this.lastFrameTime=t,this.animationFrameId=requestAnimationFrame(this.tick)}},this.replay=e,this.worldState=new Ol(e),e.events.length>0&&(this.totalDurationMs=e.events[e.events.length-1].timestampDeltaMs),this.buildMoveIndices()}buildMoveIndices(){for(const e of this.replay.events){if(e.type===Pe.PLAYER_SPAWN||e.type===Pe.PLAYER_MOVE){const t=e.uuid;this.playerMoveEvents.has(t)||this.playerMoveEvents.set(t,[]),this.playerMoveEvents.get(t).push(e)}if(e.type===Pe.ENTITY_SPAWN||e.type===Pe.ENTITY_MOVE){const t=(e.type,Pe.ENTITY_SPAWN,e.entityId);this.entityMoveEvents.has(t)||this.entityMoveEvents.set(t,[]),this.entityMoveEvents.get(t).push(e)}}}subscribe(e){return this.listeners.add(e),()=>this.listeners.delete(e)}play(){this.playing||(this.playing=!0,this.lastFrameTime=null,this.animationFrameId=requestAnimationFrame(this.tick),this.notifyListeners([]))}pause(){this.playing=!1,this.animationFrameId!==null&&(cancelAnimationFrame(this.animationFrameId),this.animationFrameId=null),this.lastFrameTime=null,this.notifyListeners([])}seek(e){this.currentTimeMs=Math.max(0,Math.min(e,this.totalDurationMs)),this.currentEventIndex=0,this.rebuildEntityState();const t=this.worldState.advanceTo(this.currentTimeMs);this.notifyListeners(t)}setRate(e){this.rate=e,this.notifyListeners([])}followPlayer(e){this.followingPlayer=e,this.notifyListeners([])}getState(){return{playing:this.playing,currentTimeMs:this.currentTimeMs,totalDurationMs:this.totalDurationMs,rate:this.rate,followingPlayer:this.followingPlayer}}getPlayers(){return new Map(this.players)}getPlayerSkin(e){return this.playerSkins.get(e)}getEntities(){return new Map(this.entities)}getWorldState(){return this.worldState}getInterpolatedEntityPosition(e){const t=this.entityMoveEvents.get(e);if(!t||t.length===0)return null;const r=Ns(t,this.currentTimeMs);if(r<0)return null;const n=t[r],a=r+1<t.length?t[r+1]:null,o=u=>u.type===Pe.ENTITY_SPAWN?0:u.yaw,s=u=>u.type===Pe.ENTITY_SPAWN?0:u.pitch;if(!a||a.timestampDeltaMs<=this.currentTimeMs)return{x:n.x,y:n.y,z:n.z,yaw:o(n),pitch:s(n)};const l=Is(n,a,this.currentTimeMs),c=a.timestampDeltaMs-n.timestampDeltaMs,h=c>0?(this.currentTimeMs-n.timestampDeltaMs)/c:0;return{...l,yaw:qi(o(n),o(a),h),pitch:qi(s(n),s(a),h)}}getInterpolatedPlayerPosition(e){const t=this.playerMoveEvents.get(e);if(!t||t.length===0)return null;const r=Ns(t,this.currentTimeMs);if(r<0)return null;const n=t[r],a=r+1<t.length?t[r+1]:null;if(!a||a.timestampDeltaMs<=this.currentTimeMs)return{x:n.x,y:n.y,z:n.z,yaw:n.yaw,pitch:n.pitch};const o=Is({...n,timestampDeltaMs:n.timestampDeltaMs},{...a,timestampDeltaMs:a.timestampDeltaMs},this.currentTimeMs),s=a.timestampDeltaMs-n.timestampDeltaMs,l=s>0?(this.currentTimeMs-n.timestampDeltaMs)/s:0;return{...o,yaw:qi(n.yaw,a.yaw,l),pitch:qi(n.pitch,a.pitch,l)}}processEventsTo(e){for(;this.currentEventIndex<this.replay.events.length;){const t=this.replay.events[this.currentEventIndex];if(t.timestampDeltaMs>e)break;this.applyEvent(t),this.currentEventIndex++}}applyEvent(e){switch(e.type){case Pe.ENTITY_SPAWN:{let t="";if(e.metadata&&e.metadata.length>0)try{t=new TextDecoder().decode(e.metadata)}catch{}this.entities.set(e.entityId,{entityId:e.entityId,entityTypeId:e.entityTypeId,entityTypeName:t,x:e.x,y:e.y,z:e.z,yaw:0,pitch:0,alive:!0});break}case Pe.ENTITY_MOVE:{const t=this.entities.get(e.entityId);t&&(t.x=e.x,t.y=e.y,t.z=e.z,t.yaw=e.yaw,t.pitch=e.pitch);break}case Pe.ENTITY_REMOVE:this.entities.delete(e.entityId);break;case Pe.PLAYER_SPAWN:{let t=e.uuid;if(e.equipment&&e.equipment.length>0)try{const r=new TextDecoder().decode(e.equipment);/^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$/i.test(r)&&(t=r)}catch{}this.players.set(e.uuid,{uuid:e.uuid,playerName:e.playerName,x:e.x,y:e.y,z:e.z,yaw:e.yaw,pitch:e.pitch,alive:!0,sneaking:!1,lastSwingTime:0,lastDamageTime:0,skinUuid:t,heldItem:"",equipment:new Map,health:20,food:20,saturation:5,effects:[],inventory:new Map});break}case Pe.PLAYER_MOVE:{let t=this.players.get(e.uuid);t?(t.x=e.x,t.y=e.y,t.z=e.z,t.yaw=e.yaw,t.pitch=e.pitch):(t={uuid:e.uuid,playerName:"Unknown",x:e.x,y:e.y,z:e.z,yaw:e.yaw,pitch:e.pitch,alive:!0,sneaking:!1,lastSwingTime:0,lastDamageTime:0,skinUuid:e.uuid,heldItem:"",equipment:new Map,health:20,food:20,saturation:5,effects:[],inventory:new Map},this.players.set(e.uuid,t));break}case Pe.PLAYER_REMOVE:this.players.delete(e.uuid);break;case Pe.PLAYER_ANIM:{const t=this.players.get(e.uuid);t&&(e.animationType===Xi.SWING_MAIN_ARM?t.lastSwingTime=e.timestampDeltaMs:e.animationType===Xi.SNEAK_START?t.sneaking=!0:e.animationType===Xi.SNEAK_STOP?t.sneaking=!1:e.animationType===Xi.DAMAGE&&(t.lastDamageTime=e.timestampDeltaMs));break}case Pe.PLAYER_EQUIPMENT:{const t=this.players.get(e.uuid);t&&(t.heldItem=e.itemName,t.equipment.set(0,{itemName:e.itemName}));break}case Pe.PLAYER_EQUIPMENT_FULL:{const t=this.players.get(e.uuid);if(t){t.equipment.clear();for(const n of e.slots)t.equipment.set(n.slotId,{itemName:n.itemName,enchantments:n.enchantments});const r=t.equipment.get(0);t.heldItem=(r==null?void 0:r.itemName)||""}break}case Pe.PLAYER_HEALTH:{const t=this.players.get(e.uuid);t&&(t.health=e.health,t.food=e.food,t.saturation=e.saturation);break}case Pe.PLAYER_EFFECTS:{const t=this.players.get(e.uuid);t&&(t.effects=e.effects.map(r=>({...r,appliedAtMs:e.timestampDeltaMs})));break}case Pe.PLAYER_INVENTORY:{const t=this.players.get(e.uuid);if(t){e.isFullSnapshot&&t.inventory.clear();for(const r of e.slots)r.count===0||!r.itemName?t.inventory.delete(r.slotIndex):t.inventory.set(r.slotIndex,{itemName:r.itemName,count:r.count,enchantments:r.enchantments})}break}case Pe.PLAYER_SKIN:{e.skinPng&&e.skinPng.length>0&&this.playerSkins.set(e.uuid,e.skinPng);break}}}rebuildEntityState(){this.entities.clear(),this.players.clear(),this.currentEventIndex=0,this.processEventsTo(this.currentTimeMs)}notifyListeners(e){const t=this.getState();for(const r of this.listeners)r(t,e)}dispose(){this.pause(),this.listeners.clear()}}/**
 * @license
 * Copyright 2010-2024 Three.js Authors
 * SPDX-License-Identifier: MIT
 */const us="170",zl=0,Us=1,Bl=2,qo=1,kl=2,sr=3,Sr=0,Dt=1,Rt=2,xr=0,ui=1,Os=2,Fs=3,zs=4,Vl=5,Or=100,Hl=101,Gl=102,Wl=103,jl=104,Xl=200,Yl=201,ql=202,Kl=203,ya=204,Ma=205,$l=206,Zl=207,Jl=208,Ql=209,ec=210,tc=211,rc=212,ic=213,nc=214,Sa=0,Ea=1,ba=2,fi=3,Ta=4,wa=5,Aa=6,Ra=7,ds=0,ac=1,sc=2,yr=0,oc=1,lc=2,cc=3,hc=4,uc=5,dc=6,pc=7,Ko=300,mi=301,gi=302,Ca=303,La=304,Un=306,Pa=1e3,zr=1001,Da=1002,pt=1003,fc=1004,Ki=1005,jt=1006,kn=1007,Br=1008,cr=1009,$o=1010,Zo=1011,ki=1012,ps=1013,kr=1014,Qt=1015,Hi=1016,fs=1017,ms=1018,_i=1020,Jo=35902,Qo=1021,el=1022,Xt=1023,tl=1024,rl=1025,di=1026,vi=1027,gs=1028,_s=1029,il=1030,vs=1031,xs=1033,Tn=33776,wn=33777,An=33778,Rn=33779,Ia=35840,Na=35841,Ua=35842,Oa=35843,Fa=36196,za=37492,Ba=37496,ka=37808,Va=37809,Ha=37810,Ga=37811,Wa=37812,ja=37813,Xa=37814,Ya=37815,qa=37816,Ka=37817,$a=37818,Za=37819,Ja=37820,Qa=37821,Cn=36492,es=36494,ts=36495,nl=36283,rs=36284,is=36285,ns=36286,mc=3200,gc=3201,al=0,_c=1,_r="",Tt="srgb",yi="srgb-linear",On="linear",rt="srgb",jr=7680,Bs=519,vc=512,xc=513,yc=514,sl=515,Mc=516,Sc=517,Ec=518,bc=519,as=35044,ks="300 es",or=2e3,Pn=2001;class Mi{addEventListener(e,t){this._listeners===void 0&&(this._listeners={});const r=this._listeners;r[e]===void 0&&(r[e]=[]),r[e].indexOf(t)===-1&&r[e].push(t)}hasEventListener(e,t){if(this._listeners===void 0)return!1;const r=this._listeners;return r[e]!==void 0&&r[e].indexOf(t)!==-1}removeEventListener(e,t){if(this._listeners===void 0)return;const r=this._listeners[e];if(r!==void 0){const n=r.indexOf(t);n!==-1&&r.splice(n,1)}}dispatchEvent(e){if(this._listeners===void 0)return;const t=this._listeners[e.type];if(t!==void 0){e.target=this;const r=t.slice(0);for(let n=0,a=r.length;n<a;n++)r[n].call(this,e);e.target=null}}}const Et=["00","01","02","03","04","05","06","07","08","09","0a","0b","0c","0d","0e","0f","10","11","12","13","14","15","16","17","18","19","1a","1b","1c","1d","1e","1f","20","21","22","23","24","25","26","27","28","29","2a","2b","2c","2d","2e","2f","30","31","32","33","34","35","36","37","38","39","3a","3b","3c","3d","3e","3f","40","41","42","43","44","45","46","47","48","49","4a","4b","4c","4d","4e","4f","50","51","52","53","54","55","56","57","58","59","5a","5b","5c","5d","5e","5f","60","61","62","63","64","65","66","67","68","69","6a","6b","6c","6d","6e","6f","70","71","72","73","74","75","76","77","78","79","7a","7b","7c","7d","7e","7f","80","81","82","83","84","85","86","87","88","89","8a","8b","8c","8d","8e","8f","90","91","92","93","94","95","96","97","98","99","9a","9b","9c","9d","9e","9f","a0","a1","a2","a3","a4","a5","a6","a7","a8","a9","aa","ab","ac","ad","ae","af","b0","b1","b2","b3","b4","b5","b6","b7","b8","b9","ba","bb","bc","bd","be","bf","c0","c1","c2","c3","c4","c5","c6","c7","c8","c9","ca","cb","cc","cd","ce","cf","d0","d1","d2","d3","d4","d5","d6","d7","d8","d9","da","db","dc","dd","de","df","e0","e1","e2","e3","e4","e5","e6","e7","e8","e9","ea","eb","ec","ed","ee","ef","f0","f1","f2","f3","f4","f5","f6","f7","f8","f9","fa","fb","fc","fd","fe","ff"],Vn=Math.PI/180,ss=180/Math.PI;function Mr(){const i=Math.random()*4294967295|0,e=Math.random()*4294967295|0,t=Math.random()*4294967295|0,r=Math.random()*4294967295|0;return(Et[i&255]+Et[i>>8&255]+Et[i>>16&255]+Et[i>>24&255]+"-"+Et[e&255]+Et[e>>8&255]+"-"+Et[e>>16&15|64]+Et[e>>24&255]+"-"+Et[t&63|128]+Et[t>>8&255]+"-"+Et[t>>16&255]+Et[t>>24&255]+Et[r&255]+Et[r>>8&255]+Et[r>>16&255]+Et[r>>24&255]).toLowerCase()}function Pt(i,e,t){return Math.max(e,Math.min(t,i))}function Tc(i,e){return(i%e+e)%e}function Hn(i,e,t){return(1-t)*i+t*e}function Jt(i,e){switch(e.constructor){case Float32Array:return i;case Uint32Array:return i/4294967295;case Uint16Array:return i/65535;case Uint8Array:return i/255;case Int32Array:return Math.max(i/2147483647,-1);case Int16Array:return Math.max(i/32767,-1);case Int8Array:return Math.max(i/127,-1);default:throw new Error("Invalid component type.")}}function it(i,e){switch(e.constructor){case Float32Array:return i;case Uint32Array:return Math.round(i*4294967295);case Uint16Array:return Math.round(i*65535);case Uint8Array:return Math.round(i*255);case Int32Array:return Math.round(i*2147483647);case Int16Array:return Math.round(i*32767);case Int8Array:return Math.round(i*127);default:throw new Error("Invalid component type.")}}class je{constructor(e=0,t=0){je.prototype.isVector2=!0,this.x=e,this.y=t}get width(){return this.x}set width(e){this.x=e}get height(){return this.y}set height(e){this.y=e}set(e,t){return this.x=e,this.y=t,this}setScalar(e){return this.x=e,this.y=e,this}setX(e){return this.x=e,this}setY(e){return this.y=e,this}setComponent(e,t){switch(e){case 0:this.x=t;break;case 1:this.y=t;break;default:throw new Error("index is out of range: "+e)}return this}getComponent(e){switch(e){case 0:return this.x;case 1:return this.y;default:throw new Error("index is out of range: "+e)}}clone(){return new this.constructor(this.x,this.y)}copy(e){return this.x=e.x,this.y=e.y,this}add(e){return this.x+=e.x,this.y+=e.y,this}addScalar(e){return this.x+=e,this.y+=e,this}addVectors(e,t){return this.x=e.x+t.x,this.y=e.y+t.y,this}addScaledVector(e,t){return this.x+=e.x*t,this.y+=e.y*t,this}sub(e){return this.x-=e.x,this.y-=e.y,this}subScalar(e){return this.x-=e,this.y-=e,this}subVectors(e,t){return this.x=e.x-t.x,this.y=e.y-t.y,this}multiply(e){return this.x*=e.x,this.y*=e.y,this}multiplyScalar(e){return this.x*=e,this.y*=e,this}divide(e){return this.x/=e.x,this.y/=e.y,this}divideScalar(e){return this.multiplyScalar(1/e)}applyMatrix3(e){const t=this.x,r=this.y,n=e.elements;return this.x=n[0]*t+n[3]*r+n[6],this.y=n[1]*t+n[4]*r+n[7],this}min(e){return this.x=Math.min(this.x,e.x),this.y=Math.min(this.y,e.y),this}max(e){return this.x=Math.max(this.x,e.x),this.y=Math.max(this.y,e.y),this}clamp(e,t){return this.x=Math.max(e.x,Math.min(t.x,this.x)),this.y=Math.max(e.y,Math.min(t.y,this.y)),this}clampScalar(e,t){return this.x=Math.max(e,Math.min(t,this.x)),this.y=Math.max(e,Math.min(t,this.y)),this}clampLength(e,t){const r=this.length();return this.divideScalar(r||1).multiplyScalar(Math.max(e,Math.min(t,r)))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this}negate(){return this.x=-this.x,this.y=-this.y,this}dot(e){return this.x*e.x+this.y*e.y}cross(e){return this.x*e.y-this.y*e.x}lengthSq(){return this.x*this.x+this.y*this.y}length(){return Math.sqrt(this.x*this.x+this.y*this.y)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)}normalize(){return this.divideScalar(this.length()||1)}angle(){return Math.atan2(-this.y,-this.x)+Math.PI}angleTo(e){const t=Math.sqrt(this.lengthSq()*e.lengthSq());if(t===0)return Math.PI/2;const r=this.dot(e)/t;return Math.acos(Pt(r,-1,1))}distanceTo(e){return Math.sqrt(this.distanceToSquared(e))}distanceToSquared(e){const t=this.x-e.x,r=this.y-e.y;return t*t+r*r}manhattanDistanceTo(e){return Math.abs(this.x-e.x)+Math.abs(this.y-e.y)}setLength(e){return this.normalize().multiplyScalar(e)}lerp(e,t){return this.x+=(e.x-this.x)*t,this.y+=(e.y-this.y)*t,this}lerpVectors(e,t,r){return this.x=e.x+(t.x-e.x)*r,this.y=e.y+(t.y-e.y)*r,this}equals(e){return e.x===this.x&&e.y===this.y}fromArray(e,t=0){return this.x=e[t],this.y=e[t+1],this}toArray(e=[],t=0){return e[t]=this.x,e[t+1]=this.y,e}fromBufferAttribute(e,t){return this.x=e.getX(t),this.y=e.getY(t),this}rotateAround(e,t){const r=Math.cos(t),n=Math.sin(t),a=this.x-e.x,o=this.y-e.y;return this.x=a*r-o*n+e.x,this.y=a*n+o*r+e.y,this}random(){return this.x=Math.random(),this.y=Math.random(),this}*[Symbol.iterator](){yield this.x,yield this.y}}class Ve{constructor(e,t,r,n,a,o,s,l,c){Ve.prototype.isMatrix3=!0,this.elements=[1,0,0,0,1,0,0,0,1],e!==void 0&&this.set(e,t,r,n,a,o,s,l,c)}set(e,t,r,n,a,o,s,l,c){const h=this.elements;return h[0]=e,h[1]=n,h[2]=s,h[3]=t,h[4]=a,h[5]=l,h[6]=r,h[7]=o,h[8]=c,this}identity(){return this.set(1,0,0,0,1,0,0,0,1),this}copy(e){const t=this.elements,r=e.elements;return t[0]=r[0],t[1]=r[1],t[2]=r[2],t[3]=r[3],t[4]=r[4],t[5]=r[5],t[6]=r[6],t[7]=r[7],t[8]=r[8],this}extractBasis(e,t,r){return e.setFromMatrix3Column(this,0),t.setFromMatrix3Column(this,1),r.setFromMatrix3Column(this,2),this}setFromMatrix4(e){const t=e.elements;return this.set(t[0],t[4],t[8],t[1],t[5],t[9],t[2],t[6],t[10]),this}multiply(e){return this.multiplyMatrices(this,e)}premultiply(e){return this.multiplyMatrices(e,this)}multiplyMatrices(e,t){const r=e.elements,n=t.elements,a=this.elements,o=r[0],s=r[3],l=r[6],c=r[1],h=r[4],u=r[7],p=r[2],f=r[5],g=r[8],_=n[0],m=n[3],d=n[6],S=n[1],T=n[4],x=n[7],N=n[2],w=n[5],A=n[8];return a[0]=o*_+s*S+l*N,a[3]=o*m+s*T+l*w,a[6]=o*d+s*x+l*A,a[1]=c*_+h*S+u*N,a[4]=c*m+h*T+u*w,a[7]=c*d+h*x+u*A,a[2]=p*_+f*S+g*N,a[5]=p*m+f*T+g*w,a[8]=p*d+f*x+g*A,this}multiplyScalar(e){const t=this.elements;return t[0]*=e,t[3]*=e,t[6]*=e,t[1]*=e,t[4]*=e,t[7]*=e,t[2]*=e,t[5]*=e,t[8]*=e,this}determinant(){const e=this.elements,t=e[0],r=e[1],n=e[2],a=e[3],o=e[4],s=e[5],l=e[6],c=e[7],h=e[8];return t*o*h-t*s*c-r*a*h+r*s*l+n*a*c-n*o*l}invert(){const e=this.elements,t=e[0],r=e[1],n=e[2],a=e[3],o=e[4],s=e[5],l=e[6],c=e[7],h=e[8],u=h*o-s*c,p=s*l-h*a,f=c*a-o*l,g=t*u+r*p+n*f;if(g===0)return this.set(0,0,0,0,0,0,0,0,0);const _=1/g;return e[0]=u*_,e[1]=(n*c-h*r)*_,e[2]=(s*r-n*o)*_,e[3]=p*_,e[4]=(h*t-n*l)*_,e[5]=(n*a-s*t)*_,e[6]=f*_,e[7]=(r*l-c*t)*_,e[8]=(o*t-r*a)*_,this}transpose(){let e;const t=this.elements;return e=t[1],t[1]=t[3],t[3]=e,e=t[2],t[2]=t[6],t[6]=e,e=t[5],t[5]=t[7],t[7]=e,this}getNormalMatrix(e){return this.setFromMatrix4(e).invert().transpose()}transposeIntoArray(e){const t=this.elements;return e[0]=t[0],e[1]=t[3],e[2]=t[6],e[3]=t[1],e[4]=t[4],e[5]=t[7],e[6]=t[2],e[7]=t[5],e[8]=t[8],this}setUvTransform(e,t,r,n,a,o,s){const l=Math.cos(a),c=Math.sin(a);return this.set(r*l,r*c,-r*(l*o+c*s)+o+e,-n*c,n*l,-n*(-c*o+l*s)+s+t,0,0,1),this}scale(e,t){return this.premultiply(Gn.makeScale(e,t)),this}rotate(e){return this.premultiply(Gn.makeRotation(-e)),this}translate(e,t){return this.premultiply(Gn.makeTranslation(e,t)),this}makeTranslation(e,t){return e.isVector2?this.set(1,0,e.x,0,1,e.y,0,0,1):this.set(1,0,e,0,1,t,0,0,1),this}makeRotation(e){const t=Math.cos(e),r=Math.sin(e);return this.set(t,-r,0,r,t,0,0,0,1),this}makeScale(e,t){return this.set(e,0,0,0,t,0,0,0,1),this}equals(e){const t=this.elements,r=e.elements;for(let n=0;n<9;n++)if(t[n]!==r[n])return!1;return!0}fromArray(e,t=0){for(let r=0;r<9;r++)this.elements[r]=e[r+t];return this}toArray(e=[],t=0){const r=this.elements;return e[t]=r[0],e[t+1]=r[1],e[t+2]=r[2],e[t+3]=r[3],e[t+4]=r[4],e[t+5]=r[5],e[t+6]=r[6],e[t+7]=r[7],e[t+8]=r[8],e}clone(){return new this.constructor().fromArray(this.elements)}}const Gn=new Ve;function ol(i){for(let e=i.length-1;e>=0;--e)if(i[e]>=65535)return!0;return!1}function Vi(i){return document.createElementNS("http://www.w3.org/1999/xhtml",i)}function wc(){const i=Vi("canvas");return i.style.display="block",i}const Vs={};function Fi(i){i in Vs||(Vs[i]=!0,console.warn(i))}function Ac(i,e,t){return new Promise(function(r,n){function a(){switch(i.clientWaitSync(e,i.SYNC_FLUSH_COMMANDS_BIT,0)){case i.WAIT_FAILED:n();break;case i.TIMEOUT_EXPIRED:setTimeout(a,t);break;default:r()}}setTimeout(a,t)})}function Rc(i){const e=i.elements;e[2]=.5*e[2]+.5*e[3],e[6]=.5*e[6]+.5*e[7],e[10]=.5*e[10]+.5*e[11],e[14]=.5*e[14]+.5*e[15]}function Cc(i){const e=i.elements;e[11]===-1?(e[10]=-e[10]-1,e[14]=-e[14]):(e[10]=-e[10],e[14]=-e[14]+1)}const $e={enabled:!0,workingColorSpace:yi,spaces:{},convert:function(i,e,t){return this.enabled===!1||e===t||!e||!t||(this.spaces[e].transfer===rt&&(i.r=lr(i.r),i.g=lr(i.g),i.b=lr(i.b)),this.spaces[e].primaries!==this.spaces[t].primaries&&(i.applyMatrix3(this.spaces[e].toXYZ),i.applyMatrix3(this.spaces[t].fromXYZ)),this.spaces[t].transfer===rt&&(i.r=pi(i.r),i.g=pi(i.g),i.b=pi(i.b))),i},fromWorkingColorSpace:function(i,e){return this.convert(i,this.workingColorSpace,e)},toWorkingColorSpace:function(i,e){return this.convert(i,e,this.workingColorSpace)},getPrimaries:function(i){return this.spaces[i].primaries},getTransfer:function(i){return i===_r?On:this.spaces[i].transfer},getLuminanceCoefficients:function(i,e=this.workingColorSpace){return i.fromArray(this.spaces[e].luminanceCoefficients)},define:function(i){Object.assign(this.spaces,i)},_getMatrix:function(i,e,t){return i.copy(this.spaces[e].toXYZ).multiply(this.spaces[t].fromXYZ)},_getDrawingBufferColorSpace:function(i){return this.spaces[i].outputColorSpaceConfig.drawingBufferColorSpace},_getUnpackColorSpace:function(i=this.workingColorSpace){return this.spaces[i].workingColorSpaceConfig.unpackColorSpace}};function lr(i){return i<.04045?i*.0773993808:Math.pow(i*.9478672986+.0521327014,2.4)}function pi(i){return i<.0031308?i*12.92:1.055*Math.pow(i,.41666)-.055}const Hs=[.64,.33,.3,.6,.15,.06],Gs=[.2126,.7152,.0722],Ws=[.3127,.329],js=new Ve().set(.4123908,.3575843,.1804808,.212639,.7151687,.0721923,.0193308,.1191948,.9505322),Xs=new Ve().set(3.2409699,-1.5373832,-.4986108,-.9692436,1.8759675,.0415551,.0556301,-.203977,1.0569715);$e.define({[yi]:{primaries:Hs,whitePoint:Ws,transfer:On,toXYZ:js,fromXYZ:Xs,luminanceCoefficients:Gs,workingColorSpaceConfig:{unpackColorSpace:Tt},outputColorSpaceConfig:{drawingBufferColorSpace:Tt}},[Tt]:{primaries:Hs,whitePoint:Ws,transfer:rt,toXYZ:js,fromXYZ:Xs,luminanceCoefficients:Gs,outputColorSpaceConfig:{drawingBufferColorSpace:Tt}}});let Xr;class Lc{static getDataURL(e){if(/^data:/i.test(e.src)||typeof HTMLCanvasElement>"u")return e.src;let t;if(e instanceof HTMLCanvasElement)t=e;else{Xr===void 0&&(Xr=Vi("canvas")),Xr.width=e.width,Xr.height=e.height;const r=Xr.getContext("2d");e instanceof ImageData?r.putImageData(e,0,0):r.drawImage(e,0,0,e.width,e.height),t=Xr}return t.width>2048||t.height>2048?(console.warn("THREE.ImageUtils.getDataURL: Image converted to jpg for performance reasons",e),t.toDataURL("image/jpeg",.6)):t.toDataURL("image/png")}static sRGBToLinear(e){if(typeof HTMLImageElement<"u"&&e instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&e instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&e instanceof ImageBitmap){const t=Vi("canvas");t.width=e.width,t.height=e.height;const r=t.getContext("2d");r.drawImage(e,0,0,e.width,e.height);const n=r.getImageData(0,0,e.width,e.height),a=n.data;for(let o=0;o<a.length;o++)a[o]=lr(a[o]/255)*255;return r.putImageData(n,0,0),t}else if(e.data){const t=e.data.slice(0);for(let r=0;r<t.length;r++)t instanceof Uint8Array||t instanceof Uint8ClampedArray?t[r]=Math.floor(lr(t[r]/255)*255):t[r]=lr(t[r]);return{data:t,width:e.width,height:e.height}}else return console.warn("THREE.ImageUtils.sRGBToLinear(): Unsupported image type. No color space conversion applied."),e}}let Pc=0;class ll{constructor(e=null){this.isSource=!0,Object.defineProperty(this,"id",{value:Pc++}),this.uuid=Mr(),this.data=e,this.dataReady=!0,this.version=0}set needsUpdate(e){e===!0&&this.version++}toJSON(e){const t=e===void 0||typeof e=="string";if(!t&&e.images[this.uuid]!==void 0)return e.images[this.uuid];const r={uuid:this.uuid,url:""},n=this.data;if(n!==null){let a;if(Array.isArray(n)){a=[];for(let o=0,s=n.length;o<s;o++)n[o].isDataTexture?a.push(Wn(n[o].image)):a.push(Wn(n[o]))}else a=Wn(n);r.url=a}return t||(e.images[this.uuid]=r),r}}function Wn(i){return typeof HTMLImageElement<"u"&&i instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&i instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&i instanceof ImageBitmap?Lc.getDataURL(i):i.data?{data:Array.from(i.data),width:i.width,height:i.height,type:i.data.constructor.name}:(console.warn("THREE.Texture: Unable to serialize Texture."),{})}let Dc=0;class _t extends Mi{constructor(e=_t.DEFAULT_IMAGE,t=_t.DEFAULT_MAPPING,r=zr,n=zr,a=jt,o=Br,s=Xt,l=cr,c=_t.DEFAULT_ANISOTROPY,h=_r){super(),this.isTexture=!0,Object.defineProperty(this,"id",{value:Dc++}),this.uuid=Mr(),this.name="",this.source=new ll(e),this.mipmaps=[],this.mapping=t,this.channel=0,this.wrapS=r,this.wrapT=n,this.magFilter=a,this.minFilter=o,this.anisotropy=c,this.format=s,this.internalFormat=null,this.type=l,this.offset=new je(0,0),this.repeat=new je(1,1),this.center=new je(0,0),this.rotation=0,this.matrixAutoUpdate=!0,this.matrix=new Ve,this.generateMipmaps=!0,this.premultiplyAlpha=!1,this.flipY=!0,this.unpackAlignment=4,this.colorSpace=h,this.userData={},this.version=0,this.onUpdate=null,this.isRenderTargetTexture=!1,this.pmremVersion=0}get image(){return this.source.data}set image(e=null){this.source.data=e}updateMatrix(){this.matrix.setUvTransform(this.offset.x,this.offset.y,this.repeat.x,this.repeat.y,this.rotation,this.center.x,this.center.y)}clone(){return new this.constructor().copy(this)}copy(e){return this.name=e.name,this.source=e.source,this.mipmaps=e.mipmaps.slice(0),this.mapping=e.mapping,this.channel=e.channel,this.wrapS=e.wrapS,this.wrapT=e.wrapT,this.magFilter=e.magFilter,this.minFilter=e.minFilter,this.anisotropy=e.anisotropy,this.format=e.format,this.internalFormat=e.internalFormat,this.type=e.type,this.offset.copy(e.offset),this.repeat.copy(e.repeat),this.center.copy(e.center),this.rotation=e.rotation,this.matrixAutoUpdate=e.matrixAutoUpdate,this.matrix.copy(e.matrix),this.generateMipmaps=e.generateMipmaps,this.premultiplyAlpha=e.premultiplyAlpha,this.flipY=e.flipY,this.unpackAlignment=e.unpackAlignment,this.colorSpace=e.colorSpace,this.userData=JSON.parse(JSON.stringify(e.userData)),this.needsUpdate=!0,this}toJSON(e){const t=e===void 0||typeof e=="string";if(!t&&e.textures[this.uuid]!==void 0)return e.textures[this.uuid];const r={metadata:{version:4.6,type:"Texture",generator:"Texture.toJSON"},uuid:this.uuid,name:this.name,image:this.source.toJSON(e).uuid,mapping:this.mapping,channel:this.channel,repeat:[this.repeat.x,this.repeat.y],offset:[this.offset.x,this.offset.y],center:[this.center.x,this.center.y],rotation:this.rotation,wrap:[this.wrapS,this.wrapT],format:this.format,internalFormat:this.internalFormat,type:this.type,colorSpace:this.colorSpace,minFilter:this.minFilter,magFilter:this.magFilter,anisotropy:this.anisotropy,flipY:this.flipY,generateMipmaps:this.generateMipmaps,premultiplyAlpha:this.premultiplyAlpha,unpackAlignment:this.unpackAlignment};return Object.keys(this.userData).length>0&&(r.userData=this.userData),t||(e.textures[this.uuid]=r),r}dispose(){this.dispatchEvent({type:"dispose"})}transformUv(e){if(this.mapping!==Ko)return e;if(e.applyMatrix3(this.matrix),e.x<0||e.x>1)switch(this.wrapS){case Pa:e.x=e.x-Math.floor(e.x);break;case zr:e.x=e.x<0?0:1;break;case Da:Math.abs(Math.floor(e.x)%2)===1?e.x=Math.ceil(e.x)-e.x:e.x=e.x-Math.floor(e.x);break}if(e.y<0||e.y>1)switch(this.wrapT){case Pa:e.y=e.y-Math.floor(e.y);break;case zr:e.y=e.y<0?0:1;break;case Da:Math.abs(Math.floor(e.y)%2)===1?e.y=Math.ceil(e.y)-e.y:e.y=e.y-Math.floor(e.y);break}return this.flipY&&(e.y=1-e.y),e}set needsUpdate(e){e===!0&&(this.version++,this.source.needsUpdate=!0)}set needsPMREMUpdate(e){e===!0&&this.pmremVersion++}}_t.DEFAULT_IMAGE=null;_t.DEFAULT_MAPPING=Ko;_t.DEFAULT_ANISOTROPY=1;class dt{constructor(e=0,t=0,r=0,n=1){dt.prototype.isVector4=!0,this.x=e,this.y=t,this.z=r,this.w=n}get width(){return this.z}set width(e){this.z=e}get height(){return this.w}set height(e){this.w=e}set(e,t,r,n){return this.x=e,this.y=t,this.z=r,this.w=n,this}setScalar(e){return this.x=e,this.y=e,this.z=e,this.w=e,this}setX(e){return this.x=e,this}setY(e){return this.y=e,this}setZ(e){return this.z=e,this}setW(e){return this.w=e,this}setComponent(e,t){switch(e){case 0:this.x=t;break;case 1:this.y=t;break;case 2:this.z=t;break;case 3:this.w=t;break;default:throw new Error("index is out of range: "+e)}return this}getComponent(e){switch(e){case 0:return this.x;case 1:return this.y;case 2:return this.z;case 3:return this.w;default:throw new Error("index is out of range: "+e)}}clone(){return new this.constructor(this.x,this.y,this.z,this.w)}copy(e){return this.x=e.x,this.y=e.y,this.z=e.z,this.w=e.w!==void 0?e.w:1,this}add(e){return this.x+=e.x,this.y+=e.y,this.z+=e.z,this.w+=e.w,this}addScalar(e){return this.x+=e,this.y+=e,this.z+=e,this.w+=e,this}addVectors(e,t){return this.x=e.x+t.x,this.y=e.y+t.y,this.z=e.z+t.z,this.w=e.w+t.w,this}addScaledVector(e,t){return this.x+=e.x*t,this.y+=e.y*t,this.z+=e.z*t,this.w+=e.w*t,this}sub(e){return this.x-=e.x,this.y-=e.y,this.z-=e.z,this.w-=e.w,this}subScalar(e){return this.x-=e,this.y-=e,this.z-=e,this.w-=e,this}subVectors(e,t){return this.x=e.x-t.x,this.y=e.y-t.y,this.z=e.z-t.z,this.w=e.w-t.w,this}multiply(e){return this.x*=e.x,this.y*=e.y,this.z*=e.z,this.w*=e.w,this}multiplyScalar(e){return this.x*=e,this.y*=e,this.z*=e,this.w*=e,this}applyMatrix4(e){const t=this.x,r=this.y,n=this.z,a=this.w,o=e.elements;return this.x=o[0]*t+o[4]*r+o[8]*n+o[12]*a,this.y=o[1]*t+o[5]*r+o[9]*n+o[13]*a,this.z=o[2]*t+o[6]*r+o[10]*n+o[14]*a,this.w=o[3]*t+o[7]*r+o[11]*n+o[15]*a,this}divide(e){return this.x/=e.x,this.y/=e.y,this.z/=e.z,this.w/=e.w,this}divideScalar(e){return this.multiplyScalar(1/e)}setAxisAngleFromQuaternion(e){this.w=2*Math.acos(e.w);const t=Math.sqrt(1-e.w*e.w);return t<1e-4?(this.x=1,this.y=0,this.z=0):(this.x=e.x/t,this.y=e.y/t,this.z=e.z/t),this}setAxisAngleFromRotationMatrix(e){let t,r,n,a;const o=e.elements,s=o[0],l=o[4],c=o[8],h=o[1],u=o[5],p=o[9],f=o[2],g=o[6],_=o[10];if(Math.abs(l-h)<.01&&Math.abs(c-f)<.01&&Math.abs(p-g)<.01){if(Math.abs(l+h)<.1&&Math.abs(c+f)<.1&&Math.abs(p+g)<.1&&Math.abs(s+u+_-3)<.1)return this.set(1,0,0,0),this;t=Math.PI;const d=(s+1)/2,S=(u+1)/2,T=(_+1)/2,x=(l+h)/4,N=(c+f)/4,w=(p+g)/4;return d>S&&d>T?d<.01?(r=0,n=.707106781,a=.707106781):(r=Math.sqrt(d),n=x/r,a=N/r):S>T?S<.01?(r=.707106781,n=0,a=.707106781):(n=Math.sqrt(S),r=x/n,a=w/n):T<.01?(r=.707106781,n=.707106781,a=0):(a=Math.sqrt(T),r=N/a,n=w/a),this.set(r,n,a,t),this}let m=Math.sqrt((g-p)*(g-p)+(c-f)*(c-f)+(h-l)*(h-l));return Math.abs(m)<.001&&(m=1),this.x=(g-p)/m,this.y=(c-f)/m,this.z=(h-l)/m,this.w=Math.acos((s+u+_-1)/2),this}setFromMatrixPosition(e){const t=e.elements;return this.x=t[12],this.y=t[13],this.z=t[14],this.w=t[15],this}min(e){return this.x=Math.min(this.x,e.x),this.y=Math.min(this.y,e.y),this.z=Math.min(this.z,e.z),this.w=Math.min(this.w,e.w),this}max(e){return this.x=Math.max(this.x,e.x),this.y=Math.max(this.y,e.y),this.z=Math.max(this.z,e.z),this.w=Math.max(this.w,e.w),this}clamp(e,t){return this.x=Math.max(e.x,Math.min(t.x,this.x)),this.y=Math.max(e.y,Math.min(t.y,this.y)),this.z=Math.max(e.z,Math.min(t.z,this.z)),this.w=Math.max(e.w,Math.min(t.w,this.w)),this}clampScalar(e,t){return this.x=Math.max(e,Math.min(t,this.x)),this.y=Math.max(e,Math.min(t,this.y)),this.z=Math.max(e,Math.min(t,this.z)),this.w=Math.max(e,Math.min(t,this.w)),this}clampLength(e,t){const r=this.length();return this.divideScalar(r||1).multiplyScalar(Math.max(e,Math.min(t,r)))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this.z=Math.floor(this.z),this.w=Math.floor(this.w),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this.z=Math.ceil(this.z),this.w=Math.ceil(this.w),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this.z=Math.round(this.z),this.w=Math.round(this.w),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this.z=Math.trunc(this.z),this.w=Math.trunc(this.w),this}negate(){return this.x=-this.x,this.y=-this.y,this.z=-this.z,this.w=-this.w,this}dot(e){return this.x*e.x+this.y*e.y+this.z*e.z+this.w*e.w}lengthSq(){return this.x*this.x+this.y*this.y+this.z*this.z+this.w*this.w}length(){return Math.sqrt(this.x*this.x+this.y*this.y+this.z*this.z+this.w*this.w)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)+Math.abs(this.z)+Math.abs(this.w)}normalize(){return this.divideScalar(this.length()||1)}setLength(e){return this.normalize().multiplyScalar(e)}lerp(e,t){return this.x+=(e.x-this.x)*t,this.y+=(e.y-this.y)*t,this.z+=(e.z-this.z)*t,this.w+=(e.w-this.w)*t,this}lerpVectors(e,t,r){return this.x=e.x+(t.x-e.x)*r,this.y=e.y+(t.y-e.y)*r,this.z=e.z+(t.z-e.z)*r,this.w=e.w+(t.w-e.w)*r,this}equals(e){return e.x===this.x&&e.y===this.y&&e.z===this.z&&e.w===this.w}fromArray(e,t=0){return this.x=e[t],this.y=e[t+1],this.z=e[t+2],this.w=e[t+3],this}toArray(e=[],t=0){return e[t]=this.x,e[t+1]=this.y,e[t+2]=this.z,e[t+3]=this.w,e}fromBufferAttribute(e,t){return this.x=e.getX(t),this.y=e.getY(t),this.z=e.getZ(t),this.w=e.getW(t),this}random(){return this.x=Math.random(),this.y=Math.random(),this.z=Math.random(),this.w=Math.random(),this}*[Symbol.iterator](){yield this.x,yield this.y,yield this.z,yield this.w}}class Ic extends Mi{constructor(e=1,t=1,r={}){super(),this.isRenderTarget=!0,this.width=e,this.height=t,this.depth=1,this.scissor=new dt(0,0,e,t),this.scissorTest=!1,this.viewport=new dt(0,0,e,t);const n={width:e,height:t,depth:1};r=Object.assign({generateMipmaps:!1,internalFormat:null,minFilter:jt,depthBuffer:!0,stencilBuffer:!1,resolveDepthBuffer:!0,resolveStencilBuffer:!0,depthTexture:null,samples:0,count:1},r);const a=new _t(n,r.mapping,r.wrapS,r.wrapT,r.magFilter,r.minFilter,r.format,r.type,r.anisotropy,r.colorSpace);a.flipY=!1,a.generateMipmaps=r.generateMipmaps,a.internalFormat=r.internalFormat,this.textures=[];const o=r.count;for(let s=0;s<o;s++)this.textures[s]=a.clone(),this.textures[s].isRenderTargetTexture=!0;this.depthBuffer=r.depthBuffer,this.stencilBuffer=r.stencilBuffer,this.resolveDepthBuffer=r.resolveDepthBuffer,this.resolveStencilBuffer=r.resolveStencilBuffer,this.depthTexture=r.depthTexture,this.samples=r.samples}get texture(){return this.textures[0]}set texture(e){this.textures[0]=e}setSize(e,t,r=1){if(this.width!==e||this.height!==t||this.depth!==r){this.width=e,this.height=t,this.depth=r;for(let n=0,a=this.textures.length;n<a;n++)this.textures[n].image.width=e,this.textures[n].image.height=t,this.textures[n].image.depth=r;this.dispose()}this.viewport.set(0,0,e,t),this.scissor.set(0,0,e,t)}clone(){return new this.constructor().copy(this)}copy(e){this.width=e.width,this.height=e.height,this.depth=e.depth,this.scissor.copy(e.scissor),this.scissorTest=e.scissorTest,this.viewport.copy(e.viewport),this.textures.length=0;for(let r=0,n=e.textures.length;r<n;r++)this.textures[r]=e.textures[r].clone(),this.textures[r].isRenderTargetTexture=!0;const t=Object.assign({},e.texture.image);return this.texture.source=new ll(t),this.depthBuffer=e.depthBuffer,this.stencilBuffer=e.stencilBuffer,this.resolveDepthBuffer=e.resolveDepthBuffer,this.resolveStencilBuffer=e.resolveStencilBuffer,e.depthTexture!==null&&(this.depthTexture=e.depthTexture.clone()),this.samples=e.samples,this}dispose(){this.dispatchEvent({type:"dispose"})}}class Vr extends Ic{constructor(e=1,t=1,r={}){super(e,t,r),this.isWebGLRenderTarget=!0}}class cl extends _t{constructor(e=null,t=1,r=1,n=1){super(null),this.isDataArrayTexture=!0,this.image={data:e,width:t,height:r,depth:n},this.magFilter=pt,this.minFilter=pt,this.wrapR=zr,this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1,this.layerUpdates=new Set}addLayerUpdate(e){this.layerUpdates.add(e)}clearLayerUpdates(){this.layerUpdates.clear()}}class Nc extends _t{constructor(e=null,t=1,r=1,n=1){super(null),this.isData3DTexture=!0,this.image={data:e,width:t,height:r,depth:n},this.magFilter=pt,this.minFilter=pt,this.wrapR=zr,this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1}}class Gi{constructor(e=0,t=0,r=0,n=1){this.isQuaternion=!0,this._x=e,this._y=t,this._z=r,this._w=n}static slerpFlat(e,t,r,n,a,o,s){let l=r[n+0],c=r[n+1],h=r[n+2],u=r[n+3];const p=a[o+0],f=a[o+1],g=a[o+2],_=a[o+3];if(s===0){e[t+0]=l,e[t+1]=c,e[t+2]=h,e[t+3]=u;return}if(s===1){e[t+0]=p,e[t+1]=f,e[t+2]=g,e[t+3]=_;return}if(u!==_||l!==p||c!==f||h!==g){let m=1-s;const d=l*p+c*f+h*g+u*_,S=d>=0?1:-1,T=1-d*d;if(T>Number.EPSILON){const N=Math.sqrt(T),w=Math.atan2(N,d*S);m=Math.sin(m*w)/N,s=Math.sin(s*w)/N}const x=s*S;if(l=l*m+p*x,c=c*m+f*x,h=h*m+g*x,u=u*m+_*x,m===1-s){const N=1/Math.sqrt(l*l+c*c+h*h+u*u);l*=N,c*=N,h*=N,u*=N}}e[t]=l,e[t+1]=c,e[t+2]=h,e[t+3]=u}static multiplyQuaternionsFlat(e,t,r,n,a,o){const s=r[n],l=r[n+1],c=r[n+2],h=r[n+3],u=a[o],p=a[o+1],f=a[o+2],g=a[o+3];return e[t]=s*g+h*u+l*f-c*p,e[t+1]=l*g+h*p+c*u-s*f,e[t+2]=c*g+h*f+s*p-l*u,e[t+3]=h*g-s*u-l*p-c*f,e}get x(){return this._x}set x(e){this._x=e,this._onChangeCallback()}get y(){return this._y}set y(e){this._y=e,this._onChangeCallback()}get z(){return this._z}set z(e){this._z=e,this._onChangeCallback()}get w(){return this._w}set w(e){this._w=e,this._onChangeCallback()}set(e,t,r,n){return this._x=e,this._y=t,this._z=r,this._w=n,this._onChangeCallback(),this}clone(){return new this.constructor(this._x,this._y,this._z,this._w)}copy(e){return this._x=e.x,this._y=e.y,this._z=e.z,this._w=e.w,this._onChangeCallback(),this}setFromEuler(e,t=!0){const r=e._x,n=e._y,a=e._z,o=e._order,s=Math.cos,l=Math.sin,c=s(r/2),h=s(n/2),u=s(a/2),p=l(r/2),f=l(n/2),g=l(a/2);switch(o){case"XYZ":this._x=p*h*u+c*f*g,this._y=c*f*u-p*h*g,this._z=c*h*g+p*f*u,this._w=c*h*u-p*f*g;break;case"YXZ":this._x=p*h*u+c*f*g,this._y=c*f*u-p*h*g,this._z=c*h*g-p*f*u,this._w=c*h*u+p*f*g;break;case"ZXY":this._x=p*h*u-c*f*g,this._y=c*f*u+p*h*g,this._z=c*h*g+p*f*u,this._w=c*h*u-p*f*g;break;case"ZYX":this._x=p*h*u-c*f*g,this._y=c*f*u+p*h*g,this._z=c*h*g-p*f*u,this._w=c*h*u+p*f*g;break;case"YZX":this._x=p*h*u+c*f*g,this._y=c*f*u+p*h*g,this._z=c*h*g-p*f*u,this._w=c*h*u-p*f*g;break;case"XZY":this._x=p*h*u-c*f*g,this._y=c*f*u-p*h*g,this._z=c*h*g+p*f*u,this._w=c*h*u+p*f*g;break;default:console.warn("THREE.Quaternion: .setFromEuler() encountered an unknown order: "+o)}return t===!0&&this._onChangeCallback(),this}setFromAxisAngle(e,t){const r=t/2,n=Math.sin(r);return this._x=e.x*n,this._y=e.y*n,this._z=e.z*n,this._w=Math.cos(r),this._onChangeCallback(),this}setFromRotationMatrix(e){const t=e.elements,r=t[0],n=t[4],a=t[8],o=t[1],s=t[5],l=t[9],c=t[2],h=t[6],u=t[10],p=r+s+u;if(p>0){const f=.5/Math.sqrt(p+1);this._w=.25/f,this._x=(h-l)*f,this._y=(a-c)*f,this._z=(o-n)*f}else if(r>s&&r>u){const f=2*Math.sqrt(1+r-s-u);this._w=(h-l)/f,this._x=.25*f,this._y=(n+o)/f,this._z=(a+c)/f}else if(s>u){const f=2*Math.sqrt(1+s-r-u);this._w=(a-c)/f,this._x=(n+o)/f,this._y=.25*f,this._z=(l+h)/f}else{const f=2*Math.sqrt(1+u-r-s);this._w=(o-n)/f,this._x=(a+c)/f,this._y=(l+h)/f,this._z=.25*f}return this._onChangeCallback(),this}setFromUnitVectors(e,t){let r=e.dot(t)+1;return r<Number.EPSILON?(r=0,Math.abs(e.x)>Math.abs(e.z)?(this._x=-e.y,this._y=e.x,this._z=0,this._w=r):(this._x=0,this._y=-e.z,this._z=e.y,this._w=r)):(this._x=e.y*t.z-e.z*t.y,this._y=e.z*t.x-e.x*t.z,this._z=e.x*t.y-e.y*t.x,this._w=r),this.normalize()}angleTo(e){return 2*Math.acos(Math.abs(Pt(this.dot(e),-1,1)))}rotateTowards(e,t){const r=this.angleTo(e);if(r===0)return this;const n=Math.min(1,t/r);return this.slerp(e,n),this}identity(){return this.set(0,0,0,1)}invert(){return this.conjugate()}conjugate(){return this._x*=-1,this._y*=-1,this._z*=-1,this._onChangeCallback(),this}dot(e){return this._x*e._x+this._y*e._y+this._z*e._z+this._w*e._w}lengthSq(){return this._x*this._x+this._y*this._y+this._z*this._z+this._w*this._w}length(){return Math.sqrt(this._x*this._x+this._y*this._y+this._z*this._z+this._w*this._w)}normalize(){let e=this.length();return e===0?(this._x=0,this._y=0,this._z=0,this._w=1):(e=1/e,this._x=this._x*e,this._y=this._y*e,this._z=this._z*e,this._w=this._w*e),this._onChangeCallback(),this}multiply(e){return this.multiplyQuaternions(this,e)}premultiply(e){return this.multiplyQuaternions(e,this)}multiplyQuaternions(e,t){const r=e._x,n=e._y,a=e._z,o=e._w,s=t._x,l=t._y,c=t._z,h=t._w;return this._x=r*h+o*s+n*c-a*l,this._y=n*h+o*l+a*s-r*c,this._z=a*h+o*c+r*l-n*s,this._w=o*h-r*s-n*l-a*c,this._onChangeCallback(),this}slerp(e,t){if(t===0)return this;if(t===1)return this.copy(e);const r=this._x,n=this._y,a=this._z,o=this._w;let s=o*e._w+r*e._x+n*e._y+a*e._z;if(s<0?(this._w=-e._w,this._x=-e._x,this._y=-e._y,this._z=-e._z,s=-s):this.copy(e),s>=1)return this._w=o,this._x=r,this._y=n,this._z=a,this;const l=1-s*s;if(l<=Number.EPSILON){const f=1-t;return this._w=f*o+t*this._w,this._x=f*r+t*this._x,this._y=f*n+t*this._y,this._z=f*a+t*this._z,this.normalize(),this}const c=Math.sqrt(l),h=Math.atan2(c,s),u=Math.sin((1-t)*h)/c,p=Math.sin(t*h)/c;return this._w=o*u+this._w*p,this._x=r*u+this._x*p,this._y=n*u+this._y*p,this._z=a*u+this._z*p,this._onChangeCallback(),this}slerpQuaternions(e,t,r){return this.copy(e).slerp(t,r)}random(){const e=2*Math.PI*Math.random(),t=2*Math.PI*Math.random(),r=Math.random(),n=Math.sqrt(1-r),a=Math.sqrt(r);return this.set(n*Math.sin(e),n*Math.cos(e),a*Math.sin(t),a*Math.cos(t))}equals(e){return e._x===this._x&&e._y===this._y&&e._z===this._z&&e._w===this._w}fromArray(e,t=0){return this._x=e[t],this._y=e[t+1],this._z=e[t+2],this._w=e[t+3],this._onChangeCallback(),this}toArray(e=[],t=0){return e[t]=this._x,e[t+1]=this._y,e[t+2]=this._z,e[t+3]=this._w,e}fromBufferAttribute(e,t){return this._x=e.getX(t),this._y=e.getY(t),this._z=e.getZ(t),this._w=e.getW(t),this._onChangeCallback(),this}toJSON(){return this.toArray()}_onChange(e){return this._onChangeCallback=e,this}_onChangeCallback(){}*[Symbol.iterator](){yield this._x,yield this._y,yield this._z,yield this._w}}class B{constructor(e=0,t=0,r=0){B.prototype.isVector3=!0,this.x=e,this.y=t,this.z=r}set(e,t,r){return r===void 0&&(r=this.z),this.x=e,this.y=t,this.z=r,this}setScalar(e){return this.x=e,this.y=e,this.z=e,this}setX(e){return this.x=e,this}setY(e){return this.y=e,this}setZ(e){return this.z=e,this}setComponent(e,t){switch(e){case 0:this.x=t;break;case 1:this.y=t;break;case 2:this.z=t;break;default:throw new Error("index is out of range: "+e)}return this}getComponent(e){switch(e){case 0:return this.x;case 1:return this.y;case 2:return this.z;default:throw new Error("index is out of range: "+e)}}clone(){return new this.constructor(this.x,this.y,this.z)}copy(e){return this.x=e.x,this.y=e.y,this.z=e.z,this}add(e){return this.x+=e.x,this.y+=e.y,this.z+=e.z,this}addScalar(e){return this.x+=e,this.y+=e,this.z+=e,this}addVectors(e,t){return this.x=e.x+t.x,this.y=e.y+t.y,this.z=e.z+t.z,this}addScaledVector(e,t){return this.x+=e.x*t,this.y+=e.y*t,this.z+=e.z*t,this}sub(e){return this.x-=e.x,this.y-=e.y,this.z-=e.z,this}subScalar(e){return this.x-=e,this.y-=e,this.z-=e,this}subVectors(e,t){return this.x=e.x-t.x,this.y=e.y-t.y,this.z=e.z-t.z,this}multiply(e){return this.x*=e.x,this.y*=e.y,this.z*=e.z,this}multiplyScalar(e){return this.x*=e,this.y*=e,this.z*=e,this}multiplyVectors(e,t){return this.x=e.x*t.x,this.y=e.y*t.y,this.z=e.z*t.z,this}applyEuler(e){return this.applyQuaternion(Ys.setFromEuler(e))}applyAxisAngle(e,t){return this.applyQuaternion(Ys.setFromAxisAngle(e,t))}applyMatrix3(e){const t=this.x,r=this.y,n=this.z,a=e.elements;return this.x=a[0]*t+a[3]*r+a[6]*n,this.y=a[1]*t+a[4]*r+a[7]*n,this.z=a[2]*t+a[5]*r+a[8]*n,this}applyNormalMatrix(e){return this.applyMatrix3(e).normalize()}applyMatrix4(e){const t=this.x,r=this.y,n=this.z,a=e.elements,o=1/(a[3]*t+a[7]*r+a[11]*n+a[15]);return this.x=(a[0]*t+a[4]*r+a[8]*n+a[12])*o,this.y=(a[1]*t+a[5]*r+a[9]*n+a[13])*o,this.z=(a[2]*t+a[6]*r+a[10]*n+a[14])*o,this}applyQuaternion(e){const t=this.x,r=this.y,n=this.z,a=e.x,o=e.y,s=e.z,l=e.w,c=2*(o*n-s*r),h=2*(s*t-a*n),u=2*(a*r-o*t);return this.x=t+l*c+o*u-s*h,this.y=r+l*h+s*c-a*u,this.z=n+l*u+a*h-o*c,this}project(e){return this.applyMatrix4(e.matrixWorldInverse).applyMatrix4(e.projectionMatrix)}unproject(e){return this.applyMatrix4(e.projectionMatrixInverse).applyMatrix4(e.matrixWorld)}transformDirection(e){const t=this.x,r=this.y,n=this.z,a=e.elements;return this.x=a[0]*t+a[4]*r+a[8]*n,this.y=a[1]*t+a[5]*r+a[9]*n,this.z=a[2]*t+a[6]*r+a[10]*n,this.normalize()}divide(e){return this.x/=e.x,this.y/=e.y,this.z/=e.z,this}divideScalar(e){return this.multiplyScalar(1/e)}min(e){return this.x=Math.min(this.x,e.x),this.y=Math.min(this.y,e.y),this.z=Math.min(this.z,e.z),this}max(e){return this.x=Math.max(this.x,e.x),this.y=Math.max(this.y,e.y),this.z=Math.max(this.z,e.z),this}clamp(e,t){return this.x=Math.max(e.x,Math.min(t.x,this.x)),this.y=Math.max(e.y,Math.min(t.y,this.y)),this.z=Math.max(e.z,Math.min(t.z,this.z)),this}clampScalar(e,t){return this.x=Math.max(e,Math.min(t,this.x)),this.y=Math.max(e,Math.min(t,this.y)),this.z=Math.max(e,Math.min(t,this.z)),this}clampLength(e,t){const r=this.length();return this.divideScalar(r||1).multiplyScalar(Math.max(e,Math.min(t,r)))}floor(){return this.x=Math.floor(this.x),this.y=Math.floor(this.y),this.z=Math.floor(this.z),this}ceil(){return this.x=Math.ceil(this.x),this.y=Math.ceil(this.y),this.z=Math.ceil(this.z),this}round(){return this.x=Math.round(this.x),this.y=Math.round(this.y),this.z=Math.round(this.z),this}roundToZero(){return this.x=Math.trunc(this.x),this.y=Math.trunc(this.y),this.z=Math.trunc(this.z),this}negate(){return this.x=-this.x,this.y=-this.y,this.z=-this.z,this}dot(e){return this.x*e.x+this.y*e.y+this.z*e.z}lengthSq(){return this.x*this.x+this.y*this.y+this.z*this.z}length(){return Math.sqrt(this.x*this.x+this.y*this.y+this.z*this.z)}manhattanLength(){return Math.abs(this.x)+Math.abs(this.y)+Math.abs(this.z)}normalize(){return this.divideScalar(this.length()||1)}setLength(e){return this.normalize().multiplyScalar(e)}lerp(e,t){return this.x+=(e.x-this.x)*t,this.y+=(e.y-this.y)*t,this.z+=(e.z-this.z)*t,this}lerpVectors(e,t,r){return this.x=e.x+(t.x-e.x)*r,this.y=e.y+(t.y-e.y)*r,this.z=e.z+(t.z-e.z)*r,this}cross(e){return this.crossVectors(this,e)}crossVectors(e,t){const r=e.x,n=e.y,a=e.z,o=t.x,s=t.y,l=t.z;return this.x=n*l-a*s,this.y=a*o-r*l,this.z=r*s-n*o,this}projectOnVector(e){const t=e.lengthSq();if(t===0)return this.set(0,0,0);const r=e.dot(this)/t;return this.copy(e).multiplyScalar(r)}projectOnPlane(e){return jn.copy(this).projectOnVector(e),this.sub(jn)}reflect(e){return this.sub(jn.copy(e).multiplyScalar(2*this.dot(e)))}angleTo(e){const t=Math.sqrt(this.lengthSq()*e.lengthSq());if(t===0)return Math.PI/2;const r=this.dot(e)/t;return Math.acos(Pt(r,-1,1))}distanceTo(e){return Math.sqrt(this.distanceToSquared(e))}distanceToSquared(e){const t=this.x-e.x,r=this.y-e.y,n=this.z-e.z;return t*t+r*r+n*n}manhattanDistanceTo(e){return Math.abs(this.x-e.x)+Math.abs(this.y-e.y)+Math.abs(this.z-e.z)}setFromSpherical(e){return this.setFromSphericalCoords(e.radius,e.phi,e.theta)}setFromSphericalCoords(e,t,r){const n=Math.sin(t)*e;return this.x=n*Math.sin(r),this.y=Math.cos(t)*e,this.z=n*Math.cos(r),this}setFromCylindrical(e){return this.setFromCylindricalCoords(e.radius,e.theta,e.y)}setFromCylindricalCoords(e,t,r){return this.x=e*Math.sin(t),this.y=r,this.z=e*Math.cos(t),this}setFromMatrixPosition(e){const t=e.elements;return this.x=t[12],this.y=t[13],this.z=t[14],this}setFromMatrixScale(e){const t=this.setFromMatrixColumn(e,0).length(),r=this.setFromMatrixColumn(e,1).length(),n=this.setFromMatrixColumn(e,2).length();return this.x=t,this.y=r,this.z=n,this}setFromMatrixColumn(e,t){return this.fromArray(e.elements,t*4)}setFromMatrix3Column(e,t){return this.fromArray(e.elements,t*3)}setFromEuler(e){return this.x=e._x,this.y=e._y,this.z=e._z,this}setFromColor(e){return this.x=e.r,this.y=e.g,this.z=e.b,this}equals(e){return e.x===this.x&&e.y===this.y&&e.z===this.z}fromArray(e,t=0){return this.x=e[t],this.y=e[t+1],this.z=e[t+2],this}toArray(e=[],t=0){return e[t]=this.x,e[t+1]=this.y,e[t+2]=this.z,e}fromBufferAttribute(e,t){return this.x=e.getX(t),this.y=e.getY(t),this.z=e.getZ(t),this}random(){return this.x=Math.random(),this.y=Math.random(),this.z=Math.random(),this}randomDirection(){const e=Math.random()*Math.PI*2,t=Math.random()*2-1,r=Math.sqrt(1-t*t);return this.x=r*Math.cos(e),this.y=t,this.z=r*Math.sin(e),this}*[Symbol.iterator](){yield this.x,yield this.y,yield this.z}}const jn=new B,Ys=new Gi;class Hr{constructor(e=new B(1/0,1/0,1/0),t=new B(-1/0,-1/0,-1/0)){this.isBox3=!0,this.min=e,this.max=t}set(e,t){return this.min.copy(e),this.max.copy(t),this}setFromArray(e){this.makeEmpty();for(let t=0,r=e.length;t<r;t+=3)this.expandByPoint(Ht.fromArray(e,t));return this}setFromBufferAttribute(e){this.makeEmpty();for(let t=0,r=e.count;t<r;t++)this.expandByPoint(Ht.fromBufferAttribute(e,t));return this}setFromPoints(e){this.makeEmpty();for(let t=0,r=e.length;t<r;t++)this.expandByPoint(e[t]);return this}setFromCenterAndSize(e,t){const r=Ht.copy(t).multiplyScalar(.5);return this.min.copy(e).sub(r),this.max.copy(e).add(r),this}setFromObject(e,t=!1){return this.makeEmpty(),this.expandByObject(e,t)}clone(){return new this.constructor().copy(this)}copy(e){return this.min.copy(e.min),this.max.copy(e.max),this}makeEmpty(){return this.min.x=this.min.y=this.min.z=1/0,this.max.x=this.max.y=this.max.z=-1/0,this}isEmpty(){return this.max.x<this.min.x||this.max.y<this.min.y||this.max.z<this.min.z}getCenter(e){return this.isEmpty()?e.set(0,0,0):e.addVectors(this.min,this.max).multiplyScalar(.5)}getSize(e){return this.isEmpty()?e.set(0,0,0):e.subVectors(this.max,this.min)}expandByPoint(e){return this.min.min(e),this.max.max(e),this}expandByVector(e){return this.min.sub(e),this.max.add(e),this}expandByScalar(e){return this.min.addScalar(-e),this.max.addScalar(e),this}expandByObject(e,t=!1){e.updateWorldMatrix(!1,!1);const r=e.geometry;if(r!==void 0){const a=r.getAttribute("position");if(t===!0&&a!==void 0&&e.isInstancedMesh!==!0)for(let o=0,s=a.count;o<s;o++)e.isMesh===!0?e.getVertexPosition(o,Ht):Ht.fromBufferAttribute(a,o),Ht.applyMatrix4(e.matrixWorld),this.expandByPoint(Ht);else e.boundingBox!==void 0?(e.boundingBox===null&&e.computeBoundingBox(),$i.copy(e.boundingBox)):(r.boundingBox===null&&r.computeBoundingBox(),$i.copy(r.boundingBox)),$i.applyMatrix4(e.matrixWorld),this.union($i)}const n=e.children;for(let a=0,o=n.length;a<o;a++)this.expandByObject(n[a],t);return this}containsPoint(e){return e.x>=this.min.x&&e.x<=this.max.x&&e.y>=this.min.y&&e.y<=this.max.y&&e.z>=this.min.z&&e.z<=this.max.z}containsBox(e){return this.min.x<=e.min.x&&e.max.x<=this.max.x&&this.min.y<=e.min.y&&e.max.y<=this.max.y&&this.min.z<=e.min.z&&e.max.z<=this.max.z}getParameter(e,t){return t.set((e.x-this.min.x)/(this.max.x-this.min.x),(e.y-this.min.y)/(this.max.y-this.min.y),(e.z-this.min.z)/(this.max.z-this.min.z))}intersectsBox(e){return e.max.x>=this.min.x&&e.min.x<=this.max.x&&e.max.y>=this.min.y&&e.min.y<=this.max.y&&e.max.z>=this.min.z&&e.min.z<=this.max.z}intersectsSphere(e){return this.clampPoint(e.center,Ht),Ht.distanceToSquared(e.center)<=e.radius*e.radius}intersectsPlane(e){let t,r;return e.normal.x>0?(t=e.normal.x*this.min.x,r=e.normal.x*this.max.x):(t=e.normal.x*this.max.x,r=e.normal.x*this.min.x),e.normal.y>0?(t+=e.normal.y*this.min.y,r+=e.normal.y*this.max.y):(t+=e.normal.y*this.max.y,r+=e.normal.y*this.min.y),e.normal.z>0?(t+=e.normal.z*this.min.z,r+=e.normal.z*this.max.z):(t+=e.normal.z*this.max.z,r+=e.normal.z*this.min.z),t<=-e.constant&&r>=-e.constant}intersectsTriangle(e){if(this.isEmpty())return!1;this.getCenter(Ai),Zi.subVectors(this.max,Ai),Yr.subVectors(e.a,Ai),qr.subVectors(e.b,Ai),Kr.subVectors(e.c,Ai),ur.subVectors(qr,Yr),dr.subVectors(Kr,qr),Ar.subVectors(Yr,Kr);let t=[0,-ur.z,ur.y,0,-dr.z,dr.y,0,-Ar.z,Ar.y,ur.z,0,-ur.x,dr.z,0,-dr.x,Ar.z,0,-Ar.x,-ur.y,ur.x,0,-dr.y,dr.x,0,-Ar.y,Ar.x,0];return!Xn(t,Yr,qr,Kr,Zi)||(t=[1,0,0,0,1,0,0,0,1],!Xn(t,Yr,qr,Kr,Zi))?!1:(Ji.crossVectors(ur,dr),t=[Ji.x,Ji.y,Ji.z],Xn(t,Yr,qr,Kr,Zi))}clampPoint(e,t){return t.copy(e).clamp(this.min,this.max)}distanceToPoint(e){return this.clampPoint(e,Ht).distanceTo(e)}getBoundingSphere(e){return this.isEmpty()?e.makeEmpty():(this.getCenter(e.center),e.radius=this.getSize(Ht).length()*.5),e}intersect(e){return this.min.max(e.min),this.max.min(e.max),this.isEmpty()&&this.makeEmpty(),this}union(e){return this.min.min(e.min),this.max.max(e.max),this}applyMatrix4(e){return this.isEmpty()?this:(tr[0].set(this.min.x,this.min.y,this.min.z).applyMatrix4(e),tr[1].set(this.min.x,this.min.y,this.max.z).applyMatrix4(e),tr[2].set(this.min.x,this.max.y,this.min.z).applyMatrix4(e),tr[3].set(this.min.x,this.max.y,this.max.z).applyMatrix4(e),tr[4].set(this.max.x,this.min.y,this.min.z).applyMatrix4(e),tr[5].set(this.max.x,this.min.y,this.max.z).applyMatrix4(e),tr[6].set(this.max.x,this.max.y,this.min.z).applyMatrix4(e),tr[7].set(this.max.x,this.max.y,this.max.z).applyMatrix4(e),this.setFromPoints(tr),this)}translate(e){return this.min.add(e),this.max.add(e),this}equals(e){return e.min.equals(this.min)&&e.max.equals(this.max)}}const tr=[new B,new B,new B,new B,new B,new B,new B,new B],Ht=new B,$i=new Hr,Yr=new B,qr=new B,Kr=new B,ur=new B,dr=new B,Ar=new B,Ai=new B,Zi=new B,Ji=new B,Rr=new B;function Xn(i,e,t,r,n){for(let a=0,o=i.length-3;a<=o;a+=3){Rr.fromArray(i,a);const s=n.x*Math.abs(Rr.x)+n.y*Math.abs(Rr.y)+n.z*Math.abs(Rr.z),l=e.dot(Rr),c=t.dot(Rr),h=r.dot(Rr);if(Math.max(-Math.max(l,c,h),Math.min(l,c,h))>s)return!1}return!0}const Uc=new Hr,Ri=new B,Yn=new B;class Si{constructor(e=new B,t=-1){this.isSphere=!0,this.center=e,this.radius=t}set(e,t){return this.center.copy(e),this.radius=t,this}setFromPoints(e,t){const r=this.center;t!==void 0?r.copy(t):Uc.setFromPoints(e).getCenter(r);let n=0;for(let a=0,o=e.length;a<o;a++)n=Math.max(n,r.distanceToSquared(e[a]));return this.radius=Math.sqrt(n),this}copy(e){return this.center.copy(e.center),this.radius=e.radius,this}isEmpty(){return this.radius<0}makeEmpty(){return this.center.set(0,0,0),this.radius=-1,this}containsPoint(e){return e.distanceToSquared(this.center)<=this.radius*this.radius}distanceToPoint(e){return e.distanceTo(this.center)-this.radius}intersectsSphere(e){const t=this.radius+e.radius;return e.center.distanceToSquared(this.center)<=t*t}intersectsBox(e){return e.intersectsSphere(this)}intersectsPlane(e){return Math.abs(e.distanceToPoint(this.center))<=this.radius}clampPoint(e,t){const r=this.center.distanceToSquared(e);return t.copy(e),r>this.radius*this.radius&&(t.sub(this.center).normalize(),t.multiplyScalar(this.radius).add(this.center)),t}getBoundingBox(e){return this.isEmpty()?(e.makeEmpty(),e):(e.set(this.center,this.center),e.expandByScalar(this.radius),e)}applyMatrix4(e){return this.center.applyMatrix4(e),this.radius=this.radius*e.getMaxScaleOnAxis(),this}translate(e){return this.center.add(e),this}expandByPoint(e){if(this.isEmpty())return this.center.copy(e),this.radius=0,this;Ri.subVectors(e,this.center);const t=Ri.lengthSq();if(t>this.radius*this.radius){const r=Math.sqrt(t),n=(r-this.radius)*.5;this.center.addScaledVector(Ri,n/r),this.radius+=n}return this}union(e){return e.isEmpty()?this:this.isEmpty()?(this.copy(e),this):(this.center.equals(e.center)===!0?this.radius=Math.max(this.radius,e.radius):(Yn.subVectors(e.center,this.center).setLength(e.radius),this.expandByPoint(Ri.copy(e.center).add(Yn)),this.expandByPoint(Ri.copy(e.center).sub(Yn))),this)}equals(e){return e.center.equals(this.center)&&e.radius===this.radius}clone(){return new this.constructor().copy(this)}}const rr=new B,qn=new B,Qi=new B,pr=new B,Kn=new B,en=new B,$n=new B;class hl{constructor(e=new B,t=new B(0,0,-1)){this.origin=e,this.direction=t}set(e,t){return this.origin.copy(e),this.direction.copy(t),this}copy(e){return this.origin.copy(e.origin),this.direction.copy(e.direction),this}at(e,t){return t.copy(this.origin).addScaledVector(this.direction,e)}lookAt(e){return this.direction.copy(e).sub(this.origin).normalize(),this}recast(e){return this.origin.copy(this.at(e,rr)),this}closestPointToPoint(e,t){t.subVectors(e,this.origin);const r=t.dot(this.direction);return r<0?t.copy(this.origin):t.copy(this.origin).addScaledVector(this.direction,r)}distanceToPoint(e){return Math.sqrt(this.distanceSqToPoint(e))}distanceSqToPoint(e){const t=rr.subVectors(e,this.origin).dot(this.direction);return t<0?this.origin.distanceToSquared(e):(rr.copy(this.origin).addScaledVector(this.direction,t),rr.distanceToSquared(e))}distanceSqToSegment(e,t,r,n){qn.copy(e).add(t).multiplyScalar(.5),Qi.copy(t).sub(e).normalize(),pr.copy(this.origin).sub(qn);const a=e.distanceTo(t)*.5,o=-this.direction.dot(Qi),s=pr.dot(this.direction),l=-pr.dot(Qi),c=pr.lengthSq(),h=Math.abs(1-o*o);let u,p,f,g;if(h>0)if(u=o*l-s,p=o*s-l,g=a*h,u>=0)if(p>=-g)if(p<=g){const _=1/h;u*=_,p*=_,f=u*(u+o*p+2*s)+p*(o*u+p+2*l)+c}else p=a,u=Math.max(0,-(o*p+s)),f=-u*u+p*(p+2*l)+c;else p=-a,u=Math.max(0,-(o*p+s)),f=-u*u+p*(p+2*l)+c;else p<=-g?(u=Math.max(0,-(-o*a+s)),p=u>0?-a:Math.min(Math.max(-a,-l),a),f=-u*u+p*(p+2*l)+c):p<=g?(u=0,p=Math.min(Math.max(-a,-l),a),f=p*(p+2*l)+c):(u=Math.max(0,-(o*a+s)),p=u>0?a:Math.min(Math.max(-a,-l),a),f=-u*u+p*(p+2*l)+c);else p=o>0?-a:a,u=Math.max(0,-(o*p+s)),f=-u*u+p*(p+2*l)+c;return r&&r.copy(this.origin).addScaledVector(this.direction,u),n&&n.copy(qn).addScaledVector(Qi,p),f}intersectSphere(e,t){rr.subVectors(e.center,this.origin);const r=rr.dot(this.direction),n=rr.dot(rr)-r*r,a=e.radius*e.radius;if(n>a)return null;const o=Math.sqrt(a-n),s=r-o,l=r+o;return l<0?null:s<0?this.at(l,t):this.at(s,t)}intersectsSphere(e){return this.distanceSqToPoint(e.center)<=e.radius*e.radius}distanceToPlane(e){const t=e.normal.dot(this.direction);if(t===0)return e.distanceToPoint(this.origin)===0?0:null;const r=-(this.origin.dot(e.normal)+e.constant)/t;return r>=0?r:null}intersectPlane(e,t){const r=this.distanceToPlane(e);return r===null?null:this.at(r,t)}intersectsPlane(e){const t=e.distanceToPoint(this.origin);return t===0||e.normal.dot(this.direction)*t<0}intersectBox(e,t){let r,n,a,o,s,l;const c=1/this.direction.x,h=1/this.direction.y,u=1/this.direction.z,p=this.origin;return c>=0?(r=(e.min.x-p.x)*c,n=(e.max.x-p.x)*c):(r=(e.max.x-p.x)*c,n=(e.min.x-p.x)*c),h>=0?(a=(e.min.y-p.y)*h,o=(e.max.y-p.y)*h):(a=(e.max.y-p.y)*h,o=(e.min.y-p.y)*h),r>o||a>n||((a>r||isNaN(r))&&(r=a),(o<n||isNaN(n))&&(n=o),u>=0?(s=(e.min.z-p.z)*u,l=(e.max.z-p.z)*u):(s=(e.max.z-p.z)*u,l=(e.min.z-p.z)*u),r>l||s>n)||((s>r||r!==r)&&(r=s),(l<n||n!==n)&&(n=l),n<0)?null:this.at(r>=0?r:n,t)}intersectsBox(e){return this.intersectBox(e,rr)!==null}intersectTriangle(e,t,r,n,a){Kn.subVectors(t,e),en.subVectors(r,e),$n.crossVectors(Kn,en);let o=this.direction.dot($n),s;if(o>0){if(n)return null;s=1}else if(o<0)s=-1,o=-o;else return null;pr.subVectors(this.origin,e);const l=s*this.direction.dot(en.crossVectors(pr,en));if(l<0)return null;const c=s*this.direction.dot(Kn.cross(pr));if(c<0||l+c>o)return null;const h=-s*pr.dot($n);return h<0?null:this.at(h/o,a)}applyMatrix4(e){return this.origin.applyMatrix4(e),this.direction.transformDirection(e),this}equals(e){return e.origin.equals(this.origin)&&e.direction.equals(this.direction)}clone(){return new this.constructor().copy(this)}}class Qe{constructor(e,t,r,n,a,o,s,l,c,h,u,p,f,g,_,m){Qe.prototype.isMatrix4=!0,this.elements=[1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1],e!==void 0&&this.set(e,t,r,n,a,o,s,l,c,h,u,p,f,g,_,m)}set(e,t,r,n,a,o,s,l,c,h,u,p,f,g,_,m){const d=this.elements;return d[0]=e,d[4]=t,d[8]=r,d[12]=n,d[1]=a,d[5]=o,d[9]=s,d[13]=l,d[2]=c,d[6]=h,d[10]=u,d[14]=p,d[3]=f,d[7]=g,d[11]=_,d[15]=m,this}identity(){return this.set(1,0,0,0,0,1,0,0,0,0,1,0,0,0,0,1),this}clone(){return new Qe().fromArray(this.elements)}copy(e){const t=this.elements,r=e.elements;return t[0]=r[0],t[1]=r[1],t[2]=r[2],t[3]=r[3],t[4]=r[4],t[5]=r[5],t[6]=r[6],t[7]=r[7],t[8]=r[8],t[9]=r[9],t[10]=r[10],t[11]=r[11],t[12]=r[12],t[13]=r[13],t[14]=r[14],t[15]=r[15],this}copyPosition(e){const t=this.elements,r=e.elements;return t[12]=r[12],t[13]=r[13],t[14]=r[14],this}setFromMatrix3(e){const t=e.elements;return this.set(t[0],t[3],t[6],0,t[1],t[4],t[7],0,t[2],t[5],t[8],0,0,0,0,1),this}extractBasis(e,t,r){return e.setFromMatrixColumn(this,0),t.setFromMatrixColumn(this,1),r.setFromMatrixColumn(this,2),this}makeBasis(e,t,r){return this.set(e.x,t.x,r.x,0,e.y,t.y,r.y,0,e.z,t.z,r.z,0,0,0,0,1),this}extractRotation(e){const t=this.elements,r=e.elements,n=1/$r.setFromMatrixColumn(e,0).length(),a=1/$r.setFromMatrixColumn(e,1).length(),o=1/$r.setFromMatrixColumn(e,2).length();return t[0]=r[0]*n,t[1]=r[1]*n,t[2]=r[2]*n,t[3]=0,t[4]=r[4]*a,t[5]=r[5]*a,t[6]=r[6]*a,t[7]=0,t[8]=r[8]*o,t[9]=r[9]*o,t[10]=r[10]*o,t[11]=0,t[12]=0,t[13]=0,t[14]=0,t[15]=1,this}makeRotationFromEuler(e){const t=this.elements,r=e.x,n=e.y,a=e.z,o=Math.cos(r),s=Math.sin(r),l=Math.cos(n),c=Math.sin(n),h=Math.cos(a),u=Math.sin(a);if(e.order==="XYZ"){const p=o*h,f=o*u,g=s*h,_=s*u;t[0]=l*h,t[4]=-l*u,t[8]=c,t[1]=f+g*c,t[5]=p-_*c,t[9]=-s*l,t[2]=_-p*c,t[6]=g+f*c,t[10]=o*l}else if(e.order==="YXZ"){const p=l*h,f=l*u,g=c*h,_=c*u;t[0]=p+_*s,t[4]=g*s-f,t[8]=o*c,t[1]=o*u,t[5]=o*h,t[9]=-s,t[2]=f*s-g,t[6]=_+p*s,t[10]=o*l}else if(e.order==="ZXY"){const p=l*h,f=l*u,g=c*h,_=c*u;t[0]=p-_*s,t[4]=-o*u,t[8]=g+f*s,t[1]=f+g*s,t[5]=o*h,t[9]=_-p*s,t[2]=-o*c,t[6]=s,t[10]=o*l}else if(e.order==="ZYX"){const p=o*h,f=o*u,g=s*h,_=s*u;t[0]=l*h,t[4]=g*c-f,t[8]=p*c+_,t[1]=l*u,t[5]=_*c+p,t[9]=f*c-g,t[2]=-c,t[6]=s*l,t[10]=o*l}else if(e.order==="YZX"){const p=o*l,f=o*c,g=s*l,_=s*c;t[0]=l*h,t[4]=_-p*u,t[8]=g*u+f,t[1]=u,t[5]=o*h,t[9]=-s*h,t[2]=-c*h,t[6]=f*u+g,t[10]=p-_*u}else if(e.order==="XZY"){const p=o*l,f=o*c,g=s*l,_=s*c;t[0]=l*h,t[4]=-u,t[8]=c*h,t[1]=p*u+_,t[5]=o*h,t[9]=f*u-g,t[2]=g*u-f,t[6]=s*h,t[10]=_*u+p}return t[3]=0,t[7]=0,t[11]=0,t[12]=0,t[13]=0,t[14]=0,t[15]=1,this}makeRotationFromQuaternion(e){return this.compose(Oc,e,Fc)}lookAt(e,t,r){const n=this.elements;return Nt.subVectors(e,t),Nt.lengthSq()===0&&(Nt.z=1),Nt.normalize(),fr.crossVectors(r,Nt),fr.lengthSq()===0&&(Math.abs(r.z)===1?Nt.x+=1e-4:Nt.z+=1e-4,Nt.normalize(),fr.crossVectors(r,Nt)),fr.normalize(),tn.crossVectors(Nt,fr),n[0]=fr.x,n[4]=tn.x,n[8]=Nt.x,n[1]=fr.y,n[5]=tn.y,n[9]=Nt.y,n[2]=fr.z,n[6]=tn.z,n[10]=Nt.z,this}multiply(e){return this.multiplyMatrices(this,e)}premultiply(e){return this.multiplyMatrices(e,this)}multiplyMatrices(e,t){const r=e.elements,n=t.elements,a=this.elements,o=r[0],s=r[4],l=r[8],c=r[12],h=r[1],u=r[5],p=r[9],f=r[13],g=r[2],_=r[6],m=r[10],d=r[14],S=r[3],T=r[7],x=r[11],N=r[15],w=n[0],A=n[4],O=n[8],b=n[12],M=n[1],R=n[5],H=n[9],U=n[13],F=n[2],K=n[6],X=n[10],$=n[14],k=n[3],Z=n[7],ce=n[11],ve=n[15];return a[0]=o*w+s*M+l*F+c*k,a[4]=o*A+s*R+l*K+c*Z,a[8]=o*O+s*H+l*X+c*ce,a[12]=o*b+s*U+l*$+c*ve,a[1]=h*w+u*M+p*F+f*k,a[5]=h*A+u*R+p*K+f*Z,a[9]=h*O+u*H+p*X+f*ce,a[13]=h*b+u*U+p*$+f*ve,a[2]=g*w+_*M+m*F+d*k,a[6]=g*A+_*R+m*K+d*Z,a[10]=g*O+_*H+m*X+d*ce,a[14]=g*b+_*U+m*$+d*ve,a[3]=S*w+T*M+x*F+N*k,a[7]=S*A+T*R+x*K+N*Z,a[11]=S*O+T*H+x*X+N*ce,a[15]=S*b+T*U+x*$+N*ve,this}multiplyScalar(e){const t=this.elements;return t[0]*=e,t[4]*=e,t[8]*=e,t[12]*=e,t[1]*=e,t[5]*=e,t[9]*=e,t[13]*=e,t[2]*=e,t[6]*=e,t[10]*=e,t[14]*=e,t[3]*=e,t[7]*=e,t[11]*=e,t[15]*=e,this}determinant(){const e=this.elements,t=e[0],r=e[4],n=e[8],a=e[12],o=e[1],s=e[5],l=e[9],c=e[13],h=e[2],u=e[6],p=e[10],f=e[14],g=e[3],_=e[7],m=e[11],d=e[15];return g*(+a*l*u-n*c*u-a*s*p+r*c*p+n*s*f-r*l*f)+_*(+t*l*f-t*c*p+a*o*p-n*o*f+n*c*h-a*l*h)+m*(+t*c*u-t*s*f-a*o*u+r*o*f+a*s*h-r*c*h)+d*(-n*s*h-t*l*u+t*s*p+n*o*u-r*o*p+r*l*h)}transpose(){const e=this.elements;let t;return t=e[1],e[1]=e[4],e[4]=t,t=e[2],e[2]=e[8],e[8]=t,t=e[6],e[6]=e[9],e[9]=t,t=e[3],e[3]=e[12],e[12]=t,t=e[7],e[7]=e[13],e[13]=t,t=e[11],e[11]=e[14],e[14]=t,this}setPosition(e,t,r){const n=this.elements;return e.isVector3?(n[12]=e.x,n[13]=e.y,n[14]=e.z):(n[12]=e,n[13]=t,n[14]=r),this}invert(){const e=this.elements,t=e[0],r=e[1],n=e[2],a=e[3],o=e[4],s=e[5],l=e[6],c=e[7],h=e[8],u=e[9],p=e[10],f=e[11],g=e[12],_=e[13],m=e[14],d=e[15],S=u*m*c-_*p*c+_*l*f-s*m*f-u*l*d+s*p*d,T=g*p*c-h*m*c-g*l*f+o*m*f+h*l*d-o*p*d,x=h*_*c-g*u*c+g*s*f-o*_*f-h*s*d+o*u*d,N=g*u*l-h*_*l-g*s*p+o*_*p+h*s*m-o*u*m,w=t*S+r*T+n*x+a*N;if(w===0)return this.set(0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0);const A=1/w;return e[0]=S*A,e[1]=(_*p*a-u*m*a-_*n*f+r*m*f+u*n*d-r*p*d)*A,e[2]=(s*m*a-_*l*a+_*n*c-r*m*c-s*n*d+r*l*d)*A,e[3]=(u*l*a-s*p*a-u*n*c+r*p*c+s*n*f-r*l*f)*A,e[4]=T*A,e[5]=(h*m*a-g*p*a+g*n*f-t*m*f-h*n*d+t*p*d)*A,e[6]=(g*l*a-o*m*a-g*n*c+t*m*c+o*n*d-t*l*d)*A,e[7]=(o*p*a-h*l*a+h*n*c-t*p*c-o*n*f+t*l*f)*A,e[8]=x*A,e[9]=(g*u*a-h*_*a-g*r*f+t*_*f+h*r*d-t*u*d)*A,e[10]=(o*_*a-g*s*a+g*r*c-t*_*c-o*r*d+t*s*d)*A,e[11]=(h*s*a-o*u*a-h*r*c+t*u*c+o*r*f-t*s*f)*A,e[12]=N*A,e[13]=(h*_*n-g*u*n+g*r*p-t*_*p-h*r*m+t*u*m)*A,e[14]=(g*s*n-o*_*n-g*r*l+t*_*l+o*r*m-t*s*m)*A,e[15]=(o*u*n-h*s*n+h*r*l-t*u*l-o*r*p+t*s*p)*A,this}scale(e){const t=this.elements,r=e.x,n=e.y,a=e.z;return t[0]*=r,t[4]*=n,t[8]*=a,t[1]*=r,t[5]*=n,t[9]*=a,t[2]*=r,t[6]*=n,t[10]*=a,t[3]*=r,t[7]*=n,t[11]*=a,this}getMaxScaleOnAxis(){const e=this.elements,t=e[0]*e[0]+e[1]*e[1]+e[2]*e[2],r=e[4]*e[4]+e[5]*e[5]+e[6]*e[6],n=e[8]*e[8]+e[9]*e[9]+e[10]*e[10];return Math.sqrt(Math.max(t,r,n))}makeTranslation(e,t,r){return e.isVector3?this.set(1,0,0,e.x,0,1,0,e.y,0,0,1,e.z,0,0,0,1):this.set(1,0,0,e,0,1,0,t,0,0,1,r,0,0,0,1),this}makeRotationX(e){const t=Math.cos(e),r=Math.sin(e);return this.set(1,0,0,0,0,t,-r,0,0,r,t,0,0,0,0,1),this}makeRotationY(e){const t=Math.cos(e),r=Math.sin(e);return this.set(t,0,r,0,0,1,0,0,-r,0,t,0,0,0,0,1),this}makeRotationZ(e){const t=Math.cos(e),r=Math.sin(e);return this.set(t,-r,0,0,r,t,0,0,0,0,1,0,0,0,0,1),this}makeRotationAxis(e,t){const r=Math.cos(t),n=Math.sin(t),a=1-r,o=e.x,s=e.y,l=e.z,c=a*o,h=a*s;return this.set(c*o+r,c*s-n*l,c*l+n*s,0,c*s+n*l,h*s+r,h*l-n*o,0,c*l-n*s,h*l+n*o,a*l*l+r,0,0,0,0,1),this}makeScale(e,t,r){return this.set(e,0,0,0,0,t,0,0,0,0,r,0,0,0,0,1),this}makeShear(e,t,r,n,a,o){return this.set(1,r,a,0,e,1,o,0,t,n,1,0,0,0,0,1),this}compose(e,t,r){const n=this.elements,a=t._x,o=t._y,s=t._z,l=t._w,c=a+a,h=o+o,u=s+s,p=a*c,f=a*h,g=a*u,_=o*h,m=o*u,d=s*u,S=l*c,T=l*h,x=l*u,N=r.x,w=r.y,A=r.z;return n[0]=(1-(_+d))*N,n[1]=(f+x)*N,n[2]=(g-T)*N,n[3]=0,n[4]=(f-x)*w,n[5]=(1-(p+d))*w,n[6]=(m+S)*w,n[7]=0,n[8]=(g+T)*A,n[9]=(m-S)*A,n[10]=(1-(p+_))*A,n[11]=0,n[12]=e.x,n[13]=e.y,n[14]=e.z,n[15]=1,this}decompose(e,t,r){const n=this.elements;let a=$r.set(n[0],n[1],n[2]).length();const o=$r.set(n[4],n[5],n[6]).length(),s=$r.set(n[8],n[9],n[10]).length();this.determinant()<0&&(a=-a),e.x=n[12],e.y=n[13],e.z=n[14],Gt.copy(this);const l=1/a,c=1/o,h=1/s;return Gt.elements[0]*=l,Gt.elements[1]*=l,Gt.elements[2]*=l,Gt.elements[4]*=c,Gt.elements[5]*=c,Gt.elements[6]*=c,Gt.elements[8]*=h,Gt.elements[9]*=h,Gt.elements[10]*=h,t.setFromRotationMatrix(Gt),r.x=a,r.y=o,r.z=s,this}makePerspective(e,t,r,n,a,o,s=or){const l=this.elements,c=2*a/(t-e),h=2*a/(r-n),u=(t+e)/(t-e),p=(r+n)/(r-n);let f,g;if(s===or)f=-(o+a)/(o-a),g=-2*o*a/(o-a);else if(s===Pn)f=-o/(o-a),g=-o*a/(o-a);else throw new Error("THREE.Matrix4.makePerspective(): Invalid coordinate system: "+s);return l[0]=c,l[4]=0,l[8]=u,l[12]=0,l[1]=0,l[5]=h,l[9]=p,l[13]=0,l[2]=0,l[6]=0,l[10]=f,l[14]=g,l[3]=0,l[7]=0,l[11]=-1,l[15]=0,this}makeOrthographic(e,t,r,n,a,o,s=or){const l=this.elements,c=1/(t-e),h=1/(r-n),u=1/(o-a),p=(t+e)*c,f=(r+n)*h;let g,_;if(s===or)g=(o+a)*u,_=-2*u;else if(s===Pn)g=a*u,_=-1*u;else throw new Error("THREE.Matrix4.makeOrthographic(): Invalid coordinate system: "+s);return l[0]=2*c,l[4]=0,l[8]=0,l[12]=-p,l[1]=0,l[5]=2*h,l[9]=0,l[13]=-f,l[2]=0,l[6]=0,l[10]=_,l[14]=-g,l[3]=0,l[7]=0,l[11]=0,l[15]=1,this}equals(e){const t=this.elements,r=e.elements;for(let n=0;n<16;n++)if(t[n]!==r[n])return!1;return!0}fromArray(e,t=0){for(let r=0;r<16;r++)this.elements[r]=e[r+t];return this}toArray(e=[],t=0){const r=this.elements;return e[t]=r[0],e[t+1]=r[1],e[t+2]=r[2],e[t+3]=r[3],e[t+4]=r[4],e[t+5]=r[5],e[t+6]=r[6],e[t+7]=r[7],e[t+8]=r[8],e[t+9]=r[9],e[t+10]=r[10],e[t+11]=r[11],e[t+12]=r[12],e[t+13]=r[13],e[t+14]=r[14],e[t+15]=r[15],e}}const $r=new B,Gt=new Qe,Oc=new B(0,0,0),Fc=new B(1,1,1),fr=new B,tn=new B,Nt=new B,qs=new Qe,Ks=new Gi;class Yt{constructor(e=0,t=0,r=0,n=Yt.DEFAULT_ORDER){this.isEuler=!0,this._x=e,this._y=t,this._z=r,this._order=n}get x(){return this._x}set x(e){this._x=e,this._onChangeCallback()}get y(){return this._y}set y(e){this._y=e,this._onChangeCallback()}get z(){return this._z}set z(e){this._z=e,this._onChangeCallback()}get order(){return this._order}set order(e){this._order=e,this._onChangeCallback()}set(e,t,r,n=this._order){return this._x=e,this._y=t,this._z=r,this._order=n,this._onChangeCallback(),this}clone(){return new this.constructor(this._x,this._y,this._z,this._order)}copy(e){return this._x=e._x,this._y=e._y,this._z=e._z,this._order=e._order,this._onChangeCallback(),this}setFromRotationMatrix(e,t=this._order,r=!0){const n=e.elements,a=n[0],o=n[4],s=n[8],l=n[1],c=n[5],h=n[9],u=n[2],p=n[6],f=n[10];switch(t){case"XYZ":this._y=Math.asin(Pt(s,-1,1)),Math.abs(s)<.9999999?(this._x=Math.atan2(-h,f),this._z=Math.atan2(-o,a)):(this._x=Math.atan2(p,c),this._z=0);break;case"YXZ":this._x=Math.asin(-Pt(h,-1,1)),Math.abs(h)<.9999999?(this._y=Math.atan2(s,f),this._z=Math.atan2(l,c)):(this._y=Math.atan2(-u,a),this._z=0);break;case"ZXY":this._x=Math.asin(Pt(p,-1,1)),Math.abs(p)<.9999999?(this._y=Math.atan2(-u,f),this._z=Math.atan2(-o,c)):(this._y=0,this._z=Math.atan2(l,a));break;case"ZYX":this._y=Math.asin(-Pt(u,-1,1)),Math.abs(u)<.9999999?(this._x=Math.atan2(p,f),this._z=Math.atan2(l,a)):(this._x=0,this._z=Math.atan2(-o,c));break;case"YZX":this._z=Math.asin(Pt(l,-1,1)),Math.abs(l)<.9999999?(this._x=Math.atan2(-h,c),this._y=Math.atan2(-u,a)):(this._x=0,this._y=Math.atan2(s,f));break;case"XZY":this._z=Math.asin(-Pt(o,-1,1)),Math.abs(o)<.9999999?(this._x=Math.atan2(p,c),this._y=Math.atan2(s,a)):(this._x=Math.atan2(-h,f),this._y=0);break;default:console.warn("THREE.Euler: .setFromRotationMatrix() encountered an unknown order: "+t)}return this._order=t,r===!0&&this._onChangeCallback(),this}setFromQuaternion(e,t,r){return qs.makeRotationFromQuaternion(e),this.setFromRotationMatrix(qs,t,r)}setFromVector3(e,t=this._order){return this.set(e.x,e.y,e.z,t)}reorder(e){return Ks.setFromEuler(this),this.setFromQuaternion(Ks,e)}equals(e){return e._x===this._x&&e._y===this._y&&e._z===this._z&&e._order===this._order}fromArray(e){return this._x=e[0],this._y=e[1],this._z=e[2],e[3]!==void 0&&(this._order=e[3]),this._onChangeCallback(),this}toArray(e=[],t=0){return e[t]=this._x,e[t+1]=this._y,e[t+2]=this._z,e[t+3]=this._order,e}_onChange(e){return this._onChangeCallback=e,this}_onChangeCallback(){}*[Symbol.iterator](){yield this._x,yield this._y,yield this._z,yield this._order}}Yt.DEFAULT_ORDER="XYZ";class ul{constructor(){this.mask=1}set(e){this.mask=(1<<e|0)>>>0}enable(e){this.mask|=1<<e|0}enableAll(){this.mask=-1}toggle(e){this.mask^=1<<e|0}disable(e){this.mask&=~(1<<e|0)}disableAll(){this.mask=0}test(e){return(this.mask&e.mask)!==0}isEnabled(e){return(this.mask&(1<<e|0))!==0}}let zc=0;const $s=new B,Zr=new Gi,ir=new Qe,rn=new B,Ci=new B,Bc=new B,kc=new Gi,Zs=new B(1,0,0),Js=new B(0,1,0),Qs=new B(0,0,1),eo={type:"added"},Vc={type:"removed"},Jr={type:"childadded",child:null},Zn={type:"childremoved",child:null};class Mt extends Mi{constructor(){super(),this.isObject3D=!0,Object.defineProperty(this,"id",{value:zc++}),this.uuid=Mr(),this.name="",this.type="Object3D",this.parent=null,this.children=[],this.up=Mt.DEFAULT_UP.clone();const e=new B,t=new Yt,r=new Gi,n=new B(1,1,1);function a(){r.setFromEuler(t,!1)}function o(){t.setFromQuaternion(r,void 0,!1)}t._onChange(a),r._onChange(o),Object.defineProperties(this,{position:{configurable:!0,enumerable:!0,value:e},rotation:{configurable:!0,enumerable:!0,value:t},quaternion:{configurable:!0,enumerable:!0,value:r},scale:{configurable:!0,enumerable:!0,value:n},modelViewMatrix:{value:new Qe},normalMatrix:{value:new Ve}}),this.matrix=new Qe,this.matrixWorld=new Qe,this.matrixAutoUpdate=Mt.DEFAULT_MATRIX_AUTO_UPDATE,this.matrixWorldAutoUpdate=Mt.DEFAULT_MATRIX_WORLD_AUTO_UPDATE,this.matrixWorldNeedsUpdate=!1,this.layers=new ul,this.visible=!0,this.castShadow=!1,this.receiveShadow=!1,this.frustumCulled=!0,this.renderOrder=0,this.animations=[],this.userData={}}onBeforeShadow(){}onAfterShadow(){}onBeforeRender(){}onAfterRender(){}applyMatrix4(e){this.matrixAutoUpdate&&this.updateMatrix(),this.matrix.premultiply(e),this.matrix.decompose(this.position,this.quaternion,this.scale)}applyQuaternion(e){return this.quaternion.premultiply(e),this}setRotationFromAxisAngle(e,t){this.quaternion.setFromAxisAngle(e,t)}setRotationFromEuler(e){this.quaternion.setFromEuler(e,!0)}setRotationFromMatrix(e){this.quaternion.setFromRotationMatrix(e)}setRotationFromQuaternion(e){this.quaternion.copy(e)}rotateOnAxis(e,t){return Zr.setFromAxisAngle(e,t),this.quaternion.multiply(Zr),this}rotateOnWorldAxis(e,t){return Zr.setFromAxisAngle(e,t),this.quaternion.premultiply(Zr),this}rotateX(e){return this.rotateOnAxis(Zs,e)}rotateY(e){return this.rotateOnAxis(Js,e)}rotateZ(e){return this.rotateOnAxis(Qs,e)}translateOnAxis(e,t){return $s.copy(e).applyQuaternion(this.quaternion),this.position.add($s.multiplyScalar(t)),this}translateX(e){return this.translateOnAxis(Zs,e)}translateY(e){return this.translateOnAxis(Js,e)}translateZ(e){return this.translateOnAxis(Qs,e)}localToWorld(e){return this.updateWorldMatrix(!0,!1),e.applyMatrix4(this.matrixWorld)}worldToLocal(e){return this.updateWorldMatrix(!0,!1),e.applyMatrix4(ir.copy(this.matrixWorld).invert())}lookAt(e,t,r){e.isVector3?rn.copy(e):rn.set(e,t,r);const n=this.parent;this.updateWorldMatrix(!0,!1),Ci.setFromMatrixPosition(this.matrixWorld),this.isCamera||this.isLight?ir.lookAt(Ci,rn,this.up):ir.lookAt(rn,Ci,this.up),this.quaternion.setFromRotationMatrix(ir),n&&(ir.extractRotation(n.matrixWorld),Zr.setFromRotationMatrix(ir),this.quaternion.premultiply(Zr.invert()))}add(e){if(arguments.length>1){for(let t=0;t<arguments.length;t++)this.add(arguments[t]);return this}return e===this?(console.error("THREE.Object3D.add: object can't be added as a child of itself.",e),this):(e&&e.isObject3D?(e.removeFromParent(),e.parent=this,this.children.push(e),e.dispatchEvent(eo),Jr.child=e,this.dispatchEvent(Jr),Jr.child=null):console.error("THREE.Object3D.add: object not an instance of THREE.Object3D.",e),this)}remove(e){if(arguments.length>1){for(let r=0;r<arguments.length;r++)this.remove(arguments[r]);return this}const t=this.children.indexOf(e);return t!==-1&&(e.parent=null,this.children.splice(t,1),e.dispatchEvent(Vc),Zn.child=e,this.dispatchEvent(Zn),Zn.child=null),this}removeFromParent(){const e=this.parent;return e!==null&&e.remove(this),this}clear(){return this.remove(...this.children)}attach(e){return this.updateWorldMatrix(!0,!1),ir.copy(this.matrixWorld).invert(),e.parent!==null&&(e.parent.updateWorldMatrix(!0,!1),ir.multiply(e.parent.matrixWorld)),e.applyMatrix4(ir),e.removeFromParent(),e.parent=this,this.children.push(e),e.updateWorldMatrix(!1,!0),e.dispatchEvent(eo),Jr.child=e,this.dispatchEvent(Jr),Jr.child=null,this}getObjectById(e){return this.getObjectByProperty("id",e)}getObjectByName(e){return this.getObjectByProperty("name",e)}getObjectByProperty(e,t){if(this[e]===t)return this;for(let r=0,n=this.children.length;r<n;r++){const a=this.children[r].getObjectByProperty(e,t);if(a!==void 0)return a}}getObjectsByProperty(e,t,r=[]){this[e]===t&&r.push(this);const n=this.children;for(let a=0,o=n.length;a<o;a++)n[a].getObjectsByProperty(e,t,r);return r}getWorldPosition(e){return this.updateWorldMatrix(!0,!1),e.setFromMatrixPosition(this.matrixWorld)}getWorldQuaternion(e){return this.updateWorldMatrix(!0,!1),this.matrixWorld.decompose(Ci,e,Bc),e}getWorldScale(e){return this.updateWorldMatrix(!0,!1),this.matrixWorld.decompose(Ci,kc,e),e}getWorldDirection(e){this.updateWorldMatrix(!0,!1);const t=this.matrixWorld.elements;return e.set(t[8],t[9],t[10]).normalize()}raycast(){}traverse(e){e(this);const t=this.children;for(let r=0,n=t.length;r<n;r++)t[r].traverse(e)}traverseVisible(e){if(this.visible===!1)return;e(this);const t=this.children;for(let r=0,n=t.length;r<n;r++)t[r].traverseVisible(e)}traverseAncestors(e){const t=this.parent;t!==null&&(e(t),t.traverseAncestors(e))}updateMatrix(){this.matrix.compose(this.position,this.quaternion,this.scale),this.matrixWorldNeedsUpdate=!0}updateMatrixWorld(e){this.matrixAutoUpdate&&this.updateMatrix(),(this.matrixWorldNeedsUpdate||e)&&(this.matrixWorldAutoUpdate===!0&&(this.parent===null?this.matrixWorld.copy(this.matrix):this.matrixWorld.multiplyMatrices(this.parent.matrixWorld,this.matrix)),this.matrixWorldNeedsUpdate=!1,e=!0);const t=this.children;for(let r=0,n=t.length;r<n;r++)t[r].updateMatrixWorld(e)}updateWorldMatrix(e,t){const r=this.parent;if(e===!0&&r!==null&&r.updateWorldMatrix(!0,!1),this.matrixAutoUpdate&&this.updateMatrix(),this.matrixWorldAutoUpdate===!0&&(this.parent===null?this.matrixWorld.copy(this.matrix):this.matrixWorld.multiplyMatrices(this.parent.matrixWorld,this.matrix)),t===!0){const n=this.children;for(let a=0,o=n.length;a<o;a++)n[a].updateWorldMatrix(!1,!0)}}toJSON(e){const t=e===void 0||typeof e=="string",r={};t&&(e={geometries:{},materials:{},textures:{},images:{},shapes:{},skeletons:{},animations:{},nodes:{}},r.metadata={version:4.6,type:"Object",generator:"Object3D.toJSON"});const n={};n.uuid=this.uuid,n.type=this.type,this.name!==""&&(n.name=this.name),this.castShadow===!0&&(n.castShadow=!0),this.receiveShadow===!0&&(n.receiveShadow=!0),this.visible===!1&&(n.visible=!1),this.frustumCulled===!1&&(n.frustumCulled=!1),this.renderOrder!==0&&(n.renderOrder=this.renderOrder),Object.keys(this.userData).length>0&&(n.userData=this.userData),n.layers=this.layers.mask,n.matrix=this.matrix.toArray(),n.up=this.up.toArray(),this.matrixAutoUpdate===!1&&(n.matrixAutoUpdate=!1),this.isInstancedMesh&&(n.type="InstancedMesh",n.count=this.count,n.instanceMatrix=this.instanceMatrix.toJSON(),this.instanceColor!==null&&(n.instanceColor=this.instanceColor.toJSON())),this.isBatchedMesh&&(n.type="BatchedMesh",n.perObjectFrustumCulled=this.perObjectFrustumCulled,n.sortObjects=this.sortObjects,n.drawRanges=this._drawRanges,n.reservedRanges=this._reservedRanges,n.visibility=this._visibility,n.active=this._active,n.bounds=this._bounds.map(s=>({boxInitialized:s.boxInitialized,boxMin:s.box.min.toArray(),boxMax:s.box.max.toArray(),sphereInitialized:s.sphereInitialized,sphereRadius:s.sphere.radius,sphereCenter:s.sphere.center.toArray()})),n.maxInstanceCount=this._maxInstanceCount,n.maxVertexCount=this._maxVertexCount,n.maxIndexCount=this._maxIndexCount,n.geometryInitialized=this._geometryInitialized,n.geometryCount=this._geometryCount,n.matricesTexture=this._matricesTexture.toJSON(e),this._colorsTexture!==null&&(n.colorsTexture=this._colorsTexture.toJSON(e)),this.boundingSphere!==null&&(n.boundingSphere={center:n.boundingSphere.center.toArray(),radius:n.boundingSphere.radius}),this.boundingBox!==null&&(n.boundingBox={min:n.boundingBox.min.toArray(),max:n.boundingBox.max.toArray()}));function a(s,l){return s[l.uuid]===void 0&&(s[l.uuid]=l.toJSON(e)),l.uuid}if(this.isScene)this.background&&(this.background.isColor?n.background=this.background.toJSON():this.background.isTexture&&(n.background=this.background.toJSON(e).uuid)),this.environment&&this.environment.isTexture&&this.environment.isRenderTargetTexture!==!0&&(n.environment=this.environment.toJSON(e).uuid);else if(this.isMesh||this.isLine||this.isPoints){n.geometry=a(e.geometries,this.geometry);const s=this.geometry.parameters;if(s!==void 0&&s.shapes!==void 0){const l=s.shapes;if(Array.isArray(l))for(let c=0,h=l.length;c<h;c++){const u=l[c];a(e.shapes,u)}else a(e.shapes,l)}}if(this.isSkinnedMesh&&(n.bindMode=this.bindMode,n.bindMatrix=this.bindMatrix.toArray(),this.skeleton!==void 0&&(a(e.skeletons,this.skeleton),n.skeleton=this.skeleton.uuid)),this.material!==void 0)if(Array.isArray(this.material)){const s=[];for(let l=0,c=this.material.length;l<c;l++)s.push(a(e.materials,this.material[l]));n.material=s}else n.material=a(e.materials,this.material);if(this.children.length>0){n.children=[];for(let s=0;s<this.children.length;s++)n.children.push(this.children[s].toJSON(e).object)}if(this.animations.length>0){n.animations=[];for(let s=0;s<this.animations.length;s++){const l=this.animations[s];n.animations.push(a(e.animations,l))}}if(t){const s=o(e.geometries),l=o(e.materials),c=o(e.textures),h=o(e.images),u=o(e.shapes),p=o(e.skeletons),f=o(e.animations),g=o(e.nodes);s.length>0&&(r.geometries=s),l.length>0&&(r.materials=l),c.length>0&&(r.textures=c),h.length>0&&(r.images=h),u.length>0&&(r.shapes=u),p.length>0&&(r.skeletons=p),f.length>0&&(r.animations=f),g.length>0&&(r.nodes=g)}return r.object=n,r;function o(s){const l=[];for(const c in s){const h=s[c];delete h.metadata,l.push(h)}return l}}clone(e){return new this.constructor().copy(this,e)}copy(e,t=!0){if(this.name=e.name,this.up.copy(e.up),this.position.copy(e.position),this.rotation.order=e.rotation.order,this.quaternion.copy(e.quaternion),this.scale.copy(e.scale),this.matrix.copy(e.matrix),this.matrixWorld.copy(e.matrixWorld),this.matrixAutoUpdate=e.matrixAutoUpdate,this.matrixWorldAutoUpdate=e.matrixWorldAutoUpdate,this.matrixWorldNeedsUpdate=e.matrixWorldNeedsUpdate,this.layers.mask=e.layers.mask,this.visible=e.visible,this.castShadow=e.castShadow,this.receiveShadow=e.receiveShadow,this.frustumCulled=e.frustumCulled,this.renderOrder=e.renderOrder,this.animations=e.animations.slice(),this.userData=JSON.parse(JSON.stringify(e.userData)),t===!0)for(let r=0;r<e.children.length;r++){const n=e.children[r];this.add(n.clone())}return this}}Mt.DEFAULT_UP=new B(0,1,0);Mt.DEFAULT_MATRIX_AUTO_UPDATE=!0;Mt.DEFAULT_MATRIX_WORLD_AUTO_UPDATE=!0;const Wt=new B,nr=new B,Jn=new B,ar=new B,Qr=new B,ei=new B,to=new B,Qn=new B,ea=new B,ta=new B,ra=new dt,ia=new dt,na=new dt;class zt{constructor(e=new B,t=new B,r=new B){this.a=e,this.b=t,this.c=r}static getNormal(e,t,r,n){n.subVectors(r,t),Wt.subVectors(e,t),n.cross(Wt);const a=n.lengthSq();return a>0?n.multiplyScalar(1/Math.sqrt(a)):n.set(0,0,0)}static getBarycoord(e,t,r,n,a){Wt.subVectors(n,t),nr.subVectors(r,t),Jn.subVectors(e,t);const o=Wt.dot(Wt),s=Wt.dot(nr),l=Wt.dot(Jn),c=nr.dot(nr),h=nr.dot(Jn),u=o*c-s*s;if(u===0)return a.set(0,0,0),null;const p=1/u,f=(c*l-s*h)*p,g=(o*h-s*l)*p;return a.set(1-f-g,g,f)}static containsPoint(e,t,r,n){return this.getBarycoord(e,t,r,n,ar)===null?!1:ar.x>=0&&ar.y>=0&&ar.x+ar.y<=1}static getInterpolation(e,t,r,n,a,o,s,l){return this.getBarycoord(e,t,r,n,ar)===null?(l.x=0,l.y=0,"z"in l&&(l.z=0),"w"in l&&(l.w=0),null):(l.setScalar(0),l.addScaledVector(a,ar.x),l.addScaledVector(o,ar.y),l.addScaledVector(s,ar.z),l)}static getInterpolatedAttribute(e,t,r,n,a,o){return ra.setScalar(0),ia.setScalar(0),na.setScalar(0),ra.fromBufferAttribute(e,t),ia.fromBufferAttribute(e,r),na.fromBufferAttribute(e,n),o.setScalar(0),o.addScaledVector(ra,a.x),o.addScaledVector(ia,a.y),o.addScaledVector(na,a.z),o}static isFrontFacing(e,t,r,n){return Wt.subVectors(r,t),nr.subVectors(e,t),Wt.cross(nr).dot(n)<0}set(e,t,r){return this.a.copy(e),this.b.copy(t),this.c.copy(r),this}setFromPointsAndIndices(e,t,r,n){return this.a.copy(e[t]),this.b.copy(e[r]),this.c.copy(e[n]),this}setFromAttributeAndIndices(e,t,r,n){return this.a.fromBufferAttribute(e,t),this.b.fromBufferAttribute(e,r),this.c.fromBufferAttribute(e,n),this}clone(){return new this.constructor().copy(this)}copy(e){return this.a.copy(e.a),this.b.copy(e.b),this.c.copy(e.c),this}getArea(){return Wt.subVectors(this.c,this.b),nr.subVectors(this.a,this.b),Wt.cross(nr).length()*.5}getMidpoint(e){return e.addVectors(this.a,this.b).add(this.c).multiplyScalar(1/3)}getNormal(e){return zt.getNormal(this.a,this.b,this.c,e)}getPlane(e){return e.setFromCoplanarPoints(this.a,this.b,this.c)}getBarycoord(e,t){return zt.getBarycoord(e,this.a,this.b,this.c,t)}getInterpolation(e,t,r,n,a){return zt.getInterpolation(e,this.a,this.b,this.c,t,r,n,a)}containsPoint(e){return zt.containsPoint(e,this.a,this.b,this.c)}isFrontFacing(e){return zt.isFrontFacing(this.a,this.b,this.c,e)}intersectsBox(e){return e.intersectsTriangle(this)}closestPointToPoint(e,t){const r=this.a,n=this.b,a=this.c;let o,s;Qr.subVectors(n,r),ei.subVectors(a,r),Qn.subVectors(e,r);const l=Qr.dot(Qn),c=ei.dot(Qn);if(l<=0&&c<=0)return t.copy(r);ea.subVectors(e,n);const h=Qr.dot(ea),u=ei.dot(ea);if(h>=0&&u<=h)return t.copy(n);const p=l*u-h*c;if(p<=0&&l>=0&&h<=0)return o=l/(l-h),t.copy(r).addScaledVector(Qr,o);ta.subVectors(e,a);const f=Qr.dot(ta),g=ei.dot(ta);if(g>=0&&f<=g)return t.copy(a);const _=f*c-l*g;if(_<=0&&c>=0&&g<=0)return s=c/(c-g),t.copy(r).addScaledVector(ei,s);const m=h*g-f*u;if(m<=0&&u-h>=0&&f-g>=0)return to.subVectors(a,n),s=(u-h)/(u-h+(f-g)),t.copy(n).addScaledVector(to,s);const d=1/(m+_+p);return o=_*d,s=p*d,t.copy(r).addScaledVector(Qr,o).addScaledVector(ei,s)}equals(e){return e.a.equals(this.a)&&e.b.equals(this.b)&&e.c.equals(this.c)}}const dl={aliceblue:15792383,antiquewhite:16444375,aqua:65535,aquamarine:8388564,azure:15794175,beige:16119260,bisque:16770244,black:0,blanchedalmond:16772045,blue:255,blueviolet:9055202,brown:10824234,burlywood:14596231,cadetblue:6266528,chartreuse:8388352,chocolate:13789470,coral:16744272,cornflowerblue:6591981,cornsilk:16775388,crimson:14423100,cyan:65535,darkblue:139,darkcyan:35723,darkgoldenrod:12092939,darkgray:11119017,darkgreen:25600,darkgrey:11119017,darkkhaki:12433259,darkmagenta:9109643,darkolivegreen:5597999,darkorange:16747520,darkorchid:10040012,darkred:9109504,darksalmon:15308410,darkseagreen:9419919,darkslateblue:4734347,darkslategray:3100495,darkslategrey:3100495,darkturquoise:52945,darkviolet:9699539,deeppink:16716947,deepskyblue:49151,dimgray:6908265,dimgrey:6908265,dodgerblue:2003199,firebrick:11674146,floralwhite:16775920,forestgreen:2263842,fuchsia:16711935,gainsboro:14474460,ghostwhite:16316671,gold:16766720,goldenrod:14329120,gray:8421504,green:32768,greenyellow:11403055,grey:8421504,honeydew:15794160,hotpink:16738740,indianred:13458524,indigo:4915330,ivory:16777200,khaki:15787660,lavender:15132410,lavenderblush:16773365,lawngreen:8190976,lemonchiffon:16775885,lightblue:11393254,lightcoral:15761536,lightcyan:14745599,lightgoldenrodyellow:16448210,lightgray:13882323,lightgreen:9498256,lightgrey:13882323,lightpink:16758465,lightsalmon:16752762,lightseagreen:2142890,lightskyblue:8900346,lightslategray:7833753,lightslategrey:7833753,lightsteelblue:11584734,lightyellow:16777184,lime:65280,limegreen:3329330,linen:16445670,magenta:16711935,maroon:8388608,mediumaquamarine:6737322,mediumblue:205,mediumorchid:12211667,mediumpurple:9662683,mediumseagreen:3978097,mediumslateblue:8087790,mediumspringgreen:64154,mediumturquoise:4772300,mediumvioletred:13047173,midnightblue:1644912,mintcream:16121850,mistyrose:16770273,moccasin:16770229,navajowhite:16768685,navy:128,oldlace:16643558,olive:8421376,olivedrab:7048739,orange:16753920,orangered:16729344,orchid:14315734,palegoldenrod:15657130,palegreen:10025880,paleturquoise:11529966,palevioletred:14381203,papayawhip:16773077,peachpuff:16767673,peru:13468991,pink:16761035,plum:14524637,powderblue:11591910,purple:8388736,rebeccapurple:6697881,red:16711680,rosybrown:12357519,royalblue:4286945,saddlebrown:9127187,salmon:16416882,sandybrown:16032864,seagreen:3050327,seashell:16774638,sienna:10506797,silver:12632256,skyblue:8900331,slateblue:6970061,slategray:7372944,slategrey:7372944,snow:16775930,springgreen:65407,steelblue:4620980,tan:13808780,teal:32896,thistle:14204888,tomato:16737095,turquoise:4251856,violet:15631086,wheat:16113331,white:16777215,whitesmoke:16119285,yellow:16776960,yellowgreen:10145074},mr={h:0,s:0,l:0},nn={h:0,s:0,l:0};function aa(i,e,t){return t<0&&(t+=1),t>1&&(t-=1),t<1/6?i+(e-i)*6*t:t<1/2?e:t<2/3?i+(e-i)*6*(2/3-t):i}class We{constructor(e,t,r){return this.isColor=!0,this.r=1,this.g=1,this.b=1,this.set(e,t,r)}set(e,t,r){if(t===void 0&&r===void 0){const n=e;n&&n.isColor?this.copy(n):typeof n=="number"?this.setHex(n):typeof n=="string"&&this.setStyle(n)}else this.setRGB(e,t,r);return this}setScalar(e){return this.r=e,this.g=e,this.b=e,this}setHex(e,t=Tt){return e=Math.floor(e),this.r=(e>>16&255)/255,this.g=(e>>8&255)/255,this.b=(e&255)/255,$e.toWorkingColorSpace(this,t),this}setRGB(e,t,r,n=$e.workingColorSpace){return this.r=e,this.g=t,this.b=r,$e.toWorkingColorSpace(this,n),this}setHSL(e,t,r,n=$e.workingColorSpace){if(e=Tc(e,1),t=Pt(t,0,1),r=Pt(r,0,1),t===0)this.r=this.g=this.b=r;else{const a=r<=.5?r*(1+t):r+t-r*t,o=2*r-a;this.r=aa(o,a,e+1/3),this.g=aa(o,a,e),this.b=aa(o,a,e-1/3)}return $e.toWorkingColorSpace(this,n),this}setStyle(e,t=Tt){function r(a){a!==void 0&&parseFloat(a)<1&&console.warn("THREE.Color: Alpha component of "+e+" will be ignored.")}let n;if(n=/^(\w+)\(([^\)]*)\)/.exec(e)){let a;const o=n[1],s=n[2];switch(o){case"rgb":case"rgba":if(a=/^\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(s))return r(a[4]),this.setRGB(Math.min(255,parseInt(a[1],10))/255,Math.min(255,parseInt(a[2],10))/255,Math.min(255,parseInt(a[3],10))/255,t);if(a=/^\s*(\d+)\%\s*,\s*(\d+)\%\s*,\s*(\d+)\%\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(s))return r(a[4]),this.setRGB(Math.min(100,parseInt(a[1],10))/100,Math.min(100,parseInt(a[2],10))/100,Math.min(100,parseInt(a[3],10))/100,t);break;case"hsl":case"hsla":if(a=/^\s*(\d*\.?\d+)\s*,\s*(\d*\.?\d+)\%\s*,\s*(\d*\.?\d+)\%\s*(?:,\s*(\d*\.?\d+)\s*)?$/.exec(s))return r(a[4]),this.setHSL(parseFloat(a[1])/360,parseFloat(a[2])/100,parseFloat(a[3])/100,t);break;default:console.warn("THREE.Color: Unknown color model "+e)}}else if(n=/^\#([A-Fa-f\d]+)$/.exec(e)){const a=n[1],o=a.length;if(o===3)return this.setRGB(parseInt(a.charAt(0),16)/15,parseInt(a.charAt(1),16)/15,parseInt(a.charAt(2),16)/15,t);if(o===6)return this.setHex(parseInt(a,16),t);console.warn("THREE.Color: Invalid hex color "+e)}else if(e&&e.length>0)return this.setColorName(e,t);return this}setColorName(e,t=Tt){const r=dl[e.toLowerCase()];return r!==void 0?this.setHex(r,t):console.warn("THREE.Color: Unknown color "+e),this}clone(){return new this.constructor(this.r,this.g,this.b)}copy(e){return this.r=e.r,this.g=e.g,this.b=e.b,this}copySRGBToLinear(e){return this.r=lr(e.r),this.g=lr(e.g),this.b=lr(e.b),this}copyLinearToSRGB(e){return this.r=pi(e.r),this.g=pi(e.g),this.b=pi(e.b),this}convertSRGBToLinear(){return this.copySRGBToLinear(this),this}convertLinearToSRGB(){return this.copyLinearToSRGB(this),this}getHex(e=Tt){return $e.fromWorkingColorSpace(bt.copy(this),e),Math.round(Pt(bt.r*255,0,255))*65536+Math.round(Pt(bt.g*255,0,255))*256+Math.round(Pt(bt.b*255,0,255))}getHexString(e=Tt){return("000000"+this.getHex(e).toString(16)).slice(-6)}getHSL(e,t=$e.workingColorSpace){$e.fromWorkingColorSpace(bt.copy(this),t);const r=bt.r,n=bt.g,a=bt.b,o=Math.max(r,n,a),s=Math.min(r,n,a);let l,c;const h=(s+o)/2;if(s===o)l=0,c=0;else{const u=o-s;switch(c=h<=.5?u/(o+s):u/(2-o-s),o){case r:l=(n-a)/u+(n<a?6:0);break;case n:l=(a-r)/u+2;break;case a:l=(r-n)/u+4;break}l/=6}return e.h=l,e.s=c,e.l=h,e}getRGB(e,t=$e.workingColorSpace){return $e.fromWorkingColorSpace(bt.copy(this),t),e.r=bt.r,e.g=bt.g,e.b=bt.b,e}getStyle(e=Tt){$e.fromWorkingColorSpace(bt.copy(this),e);const t=bt.r,r=bt.g,n=bt.b;return e!==Tt?`color(${e} ${t.toFixed(3)} ${r.toFixed(3)} ${n.toFixed(3)})`:`rgb(${Math.round(t*255)},${Math.round(r*255)},${Math.round(n*255)})`}offsetHSL(e,t,r){return this.getHSL(mr),this.setHSL(mr.h+e,mr.s+t,mr.l+r)}add(e){return this.r+=e.r,this.g+=e.g,this.b+=e.b,this}addColors(e,t){return this.r=e.r+t.r,this.g=e.g+t.g,this.b=e.b+t.b,this}addScalar(e){return this.r+=e,this.g+=e,this.b+=e,this}sub(e){return this.r=Math.max(0,this.r-e.r),this.g=Math.max(0,this.g-e.g),this.b=Math.max(0,this.b-e.b),this}multiply(e){return this.r*=e.r,this.g*=e.g,this.b*=e.b,this}multiplyScalar(e){return this.r*=e,this.g*=e,this.b*=e,this}lerp(e,t){return this.r+=(e.r-this.r)*t,this.g+=(e.g-this.g)*t,this.b+=(e.b-this.b)*t,this}lerpColors(e,t,r){return this.r=e.r+(t.r-e.r)*r,this.g=e.g+(t.g-e.g)*r,this.b=e.b+(t.b-e.b)*r,this}lerpHSL(e,t){this.getHSL(mr),e.getHSL(nn);const r=Hn(mr.h,nn.h,t),n=Hn(mr.s,nn.s,t),a=Hn(mr.l,nn.l,t);return this.setHSL(r,n,a),this}setFromVector3(e){return this.r=e.x,this.g=e.y,this.b=e.z,this}applyMatrix3(e){const t=this.r,r=this.g,n=this.b,a=e.elements;return this.r=a[0]*t+a[3]*r+a[6]*n,this.g=a[1]*t+a[4]*r+a[7]*n,this.b=a[2]*t+a[5]*r+a[8]*n,this}equals(e){return e.r===this.r&&e.g===this.g&&e.b===this.b}fromArray(e,t=0){return this.r=e[t],this.g=e[t+1],this.b=e[t+2],this}toArray(e=[],t=0){return e[t]=this.r,e[t+1]=this.g,e[t+2]=this.b,e}fromBufferAttribute(e,t){return this.r=e.getX(t),this.g=e.getY(t),this.b=e.getZ(t),this}toJSON(){return this.getHex()}*[Symbol.iterator](){yield this.r,yield this.g,yield this.b}}const bt=new We;We.NAMES=dl;let Hc=0;class Gr extends Mi{static get type(){return"Material"}get type(){return this.constructor.type}set type(e){}constructor(){super(),this.isMaterial=!0,Object.defineProperty(this,"id",{value:Hc++}),this.uuid=Mr(),this.name="",this.blending=ui,this.side=Sr,this.vertexColors=!1,this.opacity=1,this.transparent=!1,this.alphaHash=!1,this.blendSrc=ya,this.blendDst=Ma,this.blendEquation=Or,this.blendSrcAlpha=null,this.blendDstAlpha=null,this.blendEquationAlpha=null,this.blendColor=new We(0,0,0),this.blendAlpha=0,this.depthFunc=fi,this.depthTest=!0,this.depthWrite=!0,this.stencilWriteMask=255,this.stencilFunc=Bs,this.stencilRef=0,this.stencilFuncMask=255,this.stencilFail=jr,this.stencilZFail=jr,this.stencilZPass=jr,this.stencilWrite=!1,this.clippingPlanes=null,this.clipIntersection=!1,this.clipShadows=!1,this.shadowSide=null,this.colorWrite=!0,this.precision=null,this.polygonOffset=!1,this.polygonOffsetFactor=0,this.polygonOffsetUnits=0,this.dithering=!1,this.alphaToCoverage=!1,this.premultipliedAlpha=!1,this.forceSinglePass=!1,this.visible=!0,this.toneMapped=!0,this.userData={},this.version=0,this._alphaTest=0}get alphaTest(){return this._alphaTest}set alphaTest(e){this._alphaTest>0!=e>0&&this.version++,this._alphaTest=e}onBeforeRender(){}onBeforeCompile(){}customProgramCacheKey(){return this.onBeforeCompile.toString()}setValues(e){if(e!==void 0)for(const t in e){const r=e[t];if(r===void 0){console.warn(`THREE.Material: parameter '${t}' has value of undefined.`);continue}const n=this[t];if(n===void 0){console.warn(`THREE.Material: '${t}' is not a property of THREE.${this.type}.`);continue}n&&n.isColor?n.set(r):n&&n.isVector3&&r&&r.isVector3?n.copy(r):this[t]=r}}toJSON(e){const t=e===void 0||typeof e=="string";t&&(e={textures:{},images:{}});const r={metadata:{version:4.6,type:"Material",generator:"Material.toJSON"}};r.uuid=this.uuid,r.type=this.type,this.name!==""&&(r.name=this.name),this.color&&this.color.isColor&&(r.color=this.color.getHex()),this.roughness!==void 0&&(r.roughness=this.roughness),this.metalness!==void 0&&(r.metalness=this.metalness),this.sheen!==void 0&&(r.sheen=this.sheen),this.sheenColor&&this.sheenColor.isColor&&(r.sheenColor=this.sheenColor.getHex()),this.sheenRoughness!==void 0&&(r.sheenRoughness=this.sheenRoughness),this.emissive&&this.emissive.isColor&&(r.emissive=this.emissive.getHex()),this.emissiveIntensity!==void 0&&this.emissiveIntensity!==1&&(r.emissiveIntensity=this.emissiveIntensity),this.specular&&this.specular.isColor&&(r.specular=this.specular.getHex()),this.specularIntensity!==void 0&&(r.specularIntensity=this.specularIntensity),this.specularColor&&this.specularColor.isColor&&(r.specularColor=this.specularColor.getHex()),this.shininess!==void 0&&(r.shininess=this.shininess),this.clearcoat!==void 0&&(r.clearcoat=this.clearcoat),this.clearcoatRoughness!==void 0&&(r.clearcoatRoughness=this.clearcoatRoughness),this.clearcoatMap&&this.clearcoatMap.isTexture&&(r.clearcoatMap=this.clearcoatMap.toJSON(e).uuid),this.clearcoatRoughnessMap&&this.clearcoatRoughnessMap.isTexture&&(r.clearcoatRoughnessMap=this.clearcoatRoughnessMap.toJSON(e).uuid),this.clearcoatNormalMap&&this.clearcoatNormalMap.isTexture&&(r.clearcoatNormalMap=this.clearcoatNormalMap.toJSON(e).uuid,r.clearcoatNormalScale=this.clearcoatNormalScale.toArray()),this.dispersion!==void 0&&(r.dispersion=this.dispersion),this.iridescence!==void 0&&(r.iridescence=this.iridescence),this.iridescenceIOR!==void 0&&(r.iridescenceIOR=this.iridescenceIOR),this.iridescenceThicknessRange!==void 0&&(r.iridescenceThicknessRange=this.iridescenceThicknessRange),this.iridescenceMap&&this.iridescenceMap.isTexture&&(r.iridescenceMap=this.iridescenceMap.toJSON(e).uuid),this.iridescenceThicknessMap&&this.iridescenceThicknessMap.isTexture&&(r.iridescenceThicknessMap=this.iridescenceThicknessMap.toJSON(e).uuid),this.anisotropy!==void 0&&(r.anisotropy=this.anisotropy),this.anisotropyRotation!==void 0&&(r.anisotropyRotation=this.anisotropyRotation),this.anisotropyMap&&this.anisotropyMap.isTexture&&(r.anisotropyMap=this.anisotropyMap.toJSON(e).uuid),this.map&&this.map.isTexture&&(r.map=this.map.toJSON(e).uuid),this.matcap&&this.matcap.isTexture&&(r.matcap=this.matcap.toJSON(e).uuid),this.alphaMap&&this.alphaMap.isTexture&&(r.alphaMap=this.alphaMap.toJSON(e).uuid),this.lightMap&&this.lightMap.isTexture&&(r.lightMap=this.lightMap.toJSON(e).uuid,r.lightMapIntensity=this.lightMapIntensity),this.aoMap&&this.aoMap.isTexture&&(r.aoMap=this.aoMap.toJSON(e).uuid,r.aoMapIntensity=this.aoMapIntensity),this.bumpMap&&this.bumpMap.isTexture&&(r.bumpMap=this.bumpMap.toJSON(e).uuid,r.bumpScale=this.bumpScale),this.normalMap&&this.normalMap.isTexture&&(r.normalMap=this.normalMap.toJSON(e).uuid,r.normalMapType=this.normalMapType,r.normalScale=this.normalScale.toArray()),this.displacementMap&&this.displacementMap.isTexture&&(r.displacementMap=this.displacementMap.toJSON(e).uuid,r.displacementScale=this.displacementScale,r.displacementBias=this.displacementBias),this.roughnessMap&&this.roughnessMap.isTexture&&(r.roughnessMap=this.roughnessMap.toJSON(e).uuid),this.metalnessMap&&this.metalnessMap.isTexture&&(r.metalnessMap=this.metalnessMap.toJSON(e).uuid),this.emissiveMap&&this.emissiveMap.isTexture&&(r.emissiveMap=this.emissiveMap.toJSON(e).uuid),this.specularMap&&this.specularMap.isTexture&&(r.specularMap=this.specularMap.toJSON(e).uuid),this.specularIntensityMap&&this.specularIntensityMap.isTexture&&(r.specularIntensityMap=this.specularIntensityMap.toJSON(e).uuid),this.specularColorMap&&this.specularColorMap.isTexture&&(r.specularColorMap=this.specularColorMap.toJSON(e).uuid),this.envMap&&this.envMap.isTexture&&(r.envMap=this.envMap.toJSON(e).uuid,this.combine!==void 0&&(r.combine=this.combine)),this.envMapRotation!==void 0&&(r.envMapRotation=this.envMapRotation.toArray()),this.envMapIntensity!==void 0&&(r.envMapIntensity=this.envMapIntensity),this.reflectivity!==void 0&&(r.reflectivity=this.reflectivity),this.refractionRatio!==void 0&&(r.refractionRatio=this.refractionRatio),this.gradientMap&&this.gradientMap.isTexture&&(r.gradientMap=this.gradientMap.toJSON(e).uuid),this.transmission!==void 0&&(r.transmission=this.transmission),this.transmissionMap&&this.transmissionMap.isTexture&&(r.transmissionMap=this.transmissionMap.toJSON(e).uuid),this.thickness!==void 0&&(r.thickness=this.thickness),this.thicknessMap&&this.thicknessMap.isTexture&&(r.thicknessMap=this.thicknessMap.toJSON(e).uuid),this.attenuationDistance!==void 0&&this.attenuationDistance!==1/0&&(r.attenuationDistance=this.attenuationDistance),this.attenuationColor!==void 0&&(r.attenuationColor=this.attenuationColor.getHex()),this.size!==void 0&&(r.size=this.size),this.shadowSide!==null&&(r.shadowSide=this.shadowSide),this.sizeAttenuation!==void 0&&(r.sizeAttenuation=this.sizeAttenuation),this.blending!==ui&&(r.blending=this.blending),this.side!==Sr&&(r.side=this.side),this.vertexColors===!0&&(r.vertexColors=!0),this.opacity<1&&(r.opacity=this.opacity),this.transparent===!0&&(r.transparent=!0),this.blendSrc!==ya&&(r.blendSrc=this.blendSrc),this.blendDst!==Ma&&(r.blendDst=this.blendDst),this.blendEquation!==Or&&(r.blendEquation=this.blendEquation),this.blendSrcAlpha!==null&&(r.blendSrcAlpha=this.blendSrcAlpha),this.blendDstAlpha!==null&&(r.blendDstAlpha=this.blendDstAlpha),this.blendEquationAlpha!==null&&(r.blendEquationAlpha=this.blendEquationAlpha),this.blendColor&&this.blendColor.isColor&&(r.blendColor=this.blendColor.getHex()),this.blendAlpha!==0&&(r.blendAlpha=this.blendAlpha),this.depthFunc!==fi&&(r.depthFunc=this.depthFunc),this.depthTest===!1&&(r.depthTest=this.depthTest),this.depthWrite===!1&&(r.depthWrite=this.depthWrite),this.colorWrite===!1&&(r.colorWrite=this.colorWrite),this.stencilWriteMask!==255&&(r.stencilWriteMask=this.stencilWriteMask),this.stencilFunc!==Bs&&(r.stencilFunc=this.stencilFunc),this.stencilRef!==0&&(r.stencilRef=this.stencilRef),this.stencilFuncMask!==255&&(r.stencilFuncMask=this.stencilFuncMask),this.stencilFail!==jr&&(r.stencilFail=this.stencilFail),this.stencilZFail!==jr&&(r.stencilZFail=this.stencilZFail),this.stencilZPass!==jr&&(r.stencilZPass=this.stencilZPass),this.stencilWrite===!0&&(r.stencilWrite=this.stencilWrite),this.rotation!==void 0&&this.rotation!==0&&(r.rotation=this.rotation),this.polygonOffset===!0&&(r.polygonOffset=!0),this.polygonOffsetFactor!==0&&(r.polygonOffsetFactor=this.polygonOffsetFactor),this.polygonOffsetUnits!==0&&(r.polygonOffsetUnits=this.polygonOffsetUnits),this.linewidth!==void 0&&this.linewidth!==1&&(r.linewidth=this.linewidth),this.dashSize!==void 0&&(r.dashSize=this.dashSize),this.gapSize!==void 0&&(r.gapSize=this.gapSize),this.scale!==void 0&&(r.scale=this.scale),this.dithering===!0&&(r.dithering=!0),this.alphaTest>0&&(r.alphaTest=this.alphaTest),this.alphaHash===!0&&(r.alphaHash=!0),this.alphaToCoverage===!0&&(r.alphaToCoverage=!0),this.premultipliedAlpha===!0&&(r.premultipliedAlpha=!0),this.forceSinglePass===!0&&(r.forceSinglePass=!0),this.wireframe===!0&&(r.wireframe=!0),this.wireframeLinewidth>1&&(r.wireframeLinewidth=this.wireframeLinewidth),this.wireframeLinecap!=="round"&&(r.wireframeLinecap=this.wireframeLinecap),this.wireframeLinejoin!=="round"&&(r.wireframeLinejoin=this.wireframeLinejoin),this.flatShading===!0&&(r.flatShading=!0),this.visible===!1&&(r.visible=!1),this.toneMapped===!1&&(r.toneMapped=!1),this.fog===!1&&(r.fog=!1),Object.keys(this.userData).length>0&&(r.userData=this.userData);function n(a){const o=[];for(const s in a){const l=a[s];delete l.metadata,o.push(l)}return o}if(t){const a=n(e.textures),o=n(e.images);a.length>0&&(r.textures=a),o.length>0&&(r.images=o)}return r}clone(){return new this.constructor().copy(this)}copy(e){this.name=e.name,this.blending=e.blending,this.side=e.side,this.vertexColors=e.vertexColors,this.opacity=e.opacity,this.transparent=e.transparent,this.blendSrc=e.blendSrc,this.blendDst=e.blendDst,this.blendEquation=e.blendEquation,this.blendSrcAlpha=e.blendSrcAlpha,this.blendDstAlpha=e.blendDstAlpha,this.blendEquationAlpha=e.blendEquationAlpha,this.blendColor.copy(e.blendColor),this.blendAlpha=e.blendAlpha,this.depthFunc=e.depthFunc,this.depthTest=e.depthTest,this.depthWrite=e.depthWrite,this.stencilWriteMask=e.stencilWriteMask,this.stencilFunc=e.stencilFunc,this.stencilRef=e.stencilRef,this.stencilFuncMask=e.stencilFuncMask,this.stencilFail=e.stencilFail,this.stencilZFail=e.stencilZFail,this.stencilZPass=e.stencilZPass,this.stencilWrite=e.stencilWrite;const t=e.clippingPlanes;let r=null;if(t!==null){const n=t.length;r=new Array(n);for(let a=0;a!==n;++a)r[a]=t[a].clone()}return this.clippingPlanes=r,this.clipIntersection=e.clipIntersection,this.clipShadows=e.clipShadows,this.shadowSide=e.shadowSide,this.colorWrite=e.colorWrite,this.precision=e.precision,this.polygonOffset=e.polygonOffset,this.polygonOffsetFactor=e.polygonOffsetFactor,this.polygonOffsetUnits=e.polygonOffsetUnits,this.dithering=e.dithering,this.alphaTest=e.alphaTest,this.alphaHash=e.alphaHash,this.alphaToCoverage=e.alphaToCoverage,this.premultipliedAlpha=e.premultipliedAlpha,this.forceSinglePass=e.forceSinglePass,this.visible=e.visible,this.toneMapped=e.toneMapped,this.userData=JSON.parse(JSON.stringify(e.userData)),this}dispose(){this.dispatchEvent({type:"dispose"})}set needsUpdate(e){e===!0&&this.version++}onBuild(){console.warn("Material: onBuild() has been removed.")}}class Bi extends Gr{static get type(){return"MeshBasicMaterial"}constructor(e){super(),this.isMeshBasicMaterial=!0,this.color=new We(16777215),this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.specularMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Yt,this.combine=ds,this.reflectivity=1,this.refractionRatio=.98,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.fog=!0,this.setValues(e)}copy(e){return super.copy(e),this.color.copy(e.color),this.map=e.map,this.lightMap=e.lightMap,this.lightMapIntensity=e.lightMapIntensity,this.aoMap=e.aoMap,this.aoMapIntensity=e.aoMapIntensity,this.specularMap=e.specularMap,this.alphaMap=e.alphaMap,this.envMap=e.envMap,this.envMapRotation.copy(e.envMapRotation),this.combine=e.combine,this.reflectivity=e.reflectivity,this.refractionRatio=e.refractionRatio,this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this.wireframeLinecap=e.wireframeLinecap,this.wireframeLinejoin=e.wireframeLinejoin,this.fog=e.fog,this}}const mt=new B,an=new je;class Bt{constructor(e,t,r=!1){if(Array.isArray(e))throw new TypeError("THREE.BufferAttribute: array should be a Typed Array.");this.isBufferAttribute=!0,this.name="",this.array=e,this.itemSize=t,this.count=e!==void 0?e.length/t:0,this.normalized=r,this.usage=as,this.updateRanges=[],this.gpuType=Qt,this.version=0}onUploadCallback(){}set needsUpdate(e){e===!0&&this.version++}setUsage(e){return this.usage=e,this}addUpdateRange(e,t){this.updateRanges.push({start:e,count:t})}clearUpdateRanges(){this.updateRanges.length=0}copy(e){return this.name=e.name,this.array=new e.array.constructor(e.array),this.itemSize=e.itemSize,this.count=e.count,this.normalized=e.normalized,this.usage=e.usage,this.gpuType=e.gpuType,this}copyAt(e,t,r){e*=this.itemSize,r*=t.itemSize;for(let n=0,a=this.itemSize;n<a;n++)this.array[e+n]=t.array[r+n];return this}copyArray(e){return this.array.set(e),this}applyMatrix3(e){if(this.itemSize===2)for(let t=0,r=this.count;t<r;t++)an.fromBufferAttribute(this,t),an.applyMatrix3(e),this.setXY(t,an.x,an.y);else if(this.itemSize===3)for(let t=0,r=this.count;t<r;t++)mt.fromBufferAttribute(this,t),mt.applyMatrix3(e),this.setXYZ(t,mt.x,mt.y,mt.z);return this}applyMatrix4(e){for(let t=0,r=this.count;t<r;t++)mt.fromBufferAttribute(this,t),mt.applyMatrix4(e),this.setXYZ(t,mt.x,mt.y,mt.z);return this}applyNormalMatrix(e){for(let t=0,r=this.count;t<r;t++)mt.fromBufferAttribute(this,t),mt.applyNormalMatrix(e),this.setXYZ(t,mt.x,mt.y,mt.z);return this}transformDirection(e){for(let t=0,r=this.count;t<r;t++)mt.fromBufferAttribute(this,t),mt.transformDirection(e),this.setXYZ(t,mt.x,mt.y,mt.z);return this}set(e,t=0){return this.array.set(e,t),this}getComponent(e,t){let r=this.array[e*this.itemSize+t];return this.normalized&&(r=Jt(r,this.array)),r}setComponent(e,t,r){return this.normalized&&(r=it(r,this.array)),this.array[e*this.itemSize+t]=r,this}getX(e){let t=this.array[e*this.itemSize];return this.normalized&&(t=Jt(t,this.array)),t}setX(e,t){return this.normalized&&(t=it(t,this.array)),this.array[e*this.itemSize]=t,this}getY(e){let t=this.array[e*this.itemSize+1];return this.normalized&&(t=Jt(t,this.array)),t}setY(e,t){return this.normalized&&(t=it(t,this.array)),this.array[e*this.itemSize+1]=t,this}getZ(e){let t=this.array[e*this.itemSize+2];return this.normalized&&(t=Jt(t,this.array)),t}setZ(e,t){return this.normalized&&(t=it(t,this.array)),this.array[e*this.itemSize+2]=t,this}getW(e){let t=this.array[e*this.itemSize+3];return this.normalized&&(t=Jt(t,this.array)),t}setW(e,t){return this.normalized&&(t=it(t,this.array)),this.array[e*this.itemSize+3]=t,this}setXY(e,t,r){return e*=this.itemSize,this.normalized&&(t=it(t,this.array),r=it(r,this.array)),this.array[e+0]=t,this.array[e+1]=r,this}setXYZ(e,t,r,n){return e*=this.itemSize,this.normalized&&(t=it(t,this.array),r=it(r,this.array),n=it(n,this.array)),this.array[e+0]=t,this.array[e+1]=r,this.array[e+2]=n,this}setXYZW(e,t,r,n,a){return e*=this.itemSize,this.normalized&&(t=it(t,this.array),r=it(r,this.array),n=it(n,this.array),a=it(a,this.array)),this.array[e+0]=t,this.array[e+1]=r,this.array[e+2]=n,this.array[e+3]=a,this}onUpload(e){return this.onUploadCallback=e,this}clone(){return new this.constructor(this.array,this.itemSize).copy(this)}toJSON(){const e={itemSize:this.itemSize,type:this.array.constructor.name,array:Array.from(this.array),normalized:this.normalized};return this.name!==""&&(e.name=this.name),this.usage!==as&&(e.usage=this.usage),e}}class pl extends Bt{constructor(e,t,r){super(new Uint16Array(e),t,r)}}class fl extends Bt{constructor(e,t,r){super(new Uint32Array(e),t,r)}}class ht extends Bt{constructor(e,t,r){super(new Float32Array(e),t,r)}}let Gc=0;const Ot=new Qe,sa=new Mt,ti=new B,Ut=new Hr,Li=new Hr,yt=new B;class Ct extends Mi{constructor(){super(),this.isBufferGeometry=!0,Object.defineProperty(this,"id",{value:Gc++}),this.uuid=Mr(),this.name="",this.type="BufferGeometry",this.index=null,this.indirect=null,this.attributes={},this.morphAttributes={},this.morphTargetsRelative=!1,this.groups=[],this.boundingBox=null,this.boundingSphere=null,this.drawRange={start:0,count:1/0},this.userData={}}getIndex(){return this.index}setIndex(e){return Array.isArray(e)?this.index=new(ol(e)?fl:pl)(e,1):this.index=e,this}setIndirect(e){return this.indirect=e,this}getIndirect(){return this.indirect}getAttribute(e){return this.attributes[e]}setAttribute(e,t){return this.attributes[e]=t,this}deleteAttribute(e){return delete this.attributes[e],this}hasAttribute(e){return this.attributes[e]!==void 0}addGroup(e,t,r=0){this.groups.push({start:e,count:t,materialIndex:r})}clearGroups(){this.groups=[]}setDrawRange(e,t){this.drawRange.start=e,this.drawRange.count=t}applyMatrix4(e){const t=this.attributes.position;t!==void 0&&(t.applyMatrix4(e),t.needsUpdate=!0);const r=this.attributes.normal;if(r!==void 0){const a=new Ve().getNormalMatrix(e);r.applyNormalMatrix(a),r.needsUpdate=!0}const n=this.attributes.tangent;return n!==void 0&&(n.transformDirection(e),n.needsUpdate=!0),this.boundingBox!==null&&this.computeBoundingBox(),this.boundingSphere!==null&&this.computeBoundingSphere(),this}applyQuaternion(e){return Ot.makeRotationFromQuaternion(e),this.applyMatrix4(Ot),this}rotateX(e){return Ot.makeRotationX(e),this.applyMatrix4(Ot),this}rotateY(e){return Ot.makeRotationY(e),this.applyMatrix4(Ot),this}rotateZ(e){return Ot.makeRotationZ(e),this.applyMatrix4(Ot),this}translate(e,t,r){return Ot.makeTranslation(e,t,r),this.applyMatrix4(Ot),this}scale(e,t,r){return Ot.makeScale(e,t,r),this.applyMatrix4(Ot),this}lookAt(e){return sa.lookAt(e),sa.updateMatrix(),this.applyMatrix4(sa.matrix),this}center(){return this.computeBoundingBox(),this.boundingBox.getCenter(ti).negate(),this.translate(ti.x,ti.y,ti.z),this}setFromPoints(e){const t=this.getAttribute("position");if(t===void 0){const r=[];for(let n=0,a=e.length;n<a;n++){const o=e[n];r.push(o.x,o.y,o.z||0)}this.setAttribute("position",new ht(r,3))}else{for(let r=0,n=t.count;r<n;r++){const a=e[r];t.setXYZ(r,a.x,a.y,a.z||0)}e.length>t.count&&console.warn("THREE.BufferGeometry: Buffer size too small for points data. Use .dispose() and create a new geometry."),t.needsUpdate=!0}return this}computeBoundingBox(){this.boundingBox===null&&(this.boundingBox=new Hr);const e=this.attributes.position,t=this.morphAttributes.position;if(e&&e.isGLBufferAttribute){console.error("THREE.BufferGeometry.computeBoundingBox(): GLBufferAttribute requires a manual bounding box.",this),this.boundingBox.set(new B(-1/0,-1/0,-1/0),new B(1/0,1/0,1/0));return}if(e!==void 0){if(this.boundingBox.setFromBufferAttribute(e),t)for(let r=0,n=t.length;r<n;r++){const a=t[r];Ut.setFromBufferAttribute(a),this.morphTargetsRelative?(yt.addVectors(this.boundingBox.min,Ut.min),this.boundingBox.expandByPoint(yt),yt.addVectors(this.boundingBox.max,Ut.max),this.boundingBox.expandByPoint(yt)):(this.boundingBox.expandByPoint(Ut.min),this.boundingBox.expandByPoint(Ut.max))}}else this.boundingBox.makeEmpty();(isNaN(this.boundingBox.min.x)||isNaN(this.boundingBox.min.y)||isNaN(this.boundingBox.min.z))&&console.error('THREE.BufferGeometry.computeBoundingBox(): Computed min/max have NaN values. The "position" attribute is likely to have NaN values.',this)}computeBoundingSphere(){this.boundingSphere===null&&(this.boundingSphere=new Si);const e=this.attributes.position,t=this.morphAttributes.position;if(e&&e.isGLBufferAttribute){console.error("THREE.BufferGeometry.computeBoundingSphere(): GLBufferAttribute requires a manual bounding sphere.",this),this.boundingSphere.set(new B,1/0);return}if(e){const r=this.boundingSphere.center;if(Ut.setFromBufferAttribute(e),t)for(let a=0,o=t.length;a<o;a++){const s=t[a];Li.setFromBufferAttribute(s),this.morphTargetsRelative?(yt.addVectors(Ut.min,Li.min),Ut.expandByPoint(yt),yt.addVectors(Ut.max,Li.max),Ut.expandByPoint(yt)):(Ut.expandByPoint(Li.min),Ut.expandByPoint(Li.max))}Ut.getCenter(r);let n=0;for(let a=0,o=e.count;a<o;a++)yt.fromBufferAttribute(e,a),n=Math.max(n,r.distanceToSquared(yt));if(t)for(let a=0,o=t.length;a<o;a++){const s=t[a],l=this.morphTargetsRelative;for(let c=0,h=s.count;c<h;c++)yt.fromBufferAttribute(s,c),l&&(ti.fromBufferAttribute(e,c),yt.add(ti)),n=Math.max(n,r.distanceToSquared(yt))}this.boundingSphere.radius=Math.sqrt(n),isNaN(this.boundingSphere.radius)&&console.error('THREE.BufferGeometry.computeBoundingSphere(): Computed radius is NaN. The "position" attribute is likely to have NaN values.',this)}}computeTangents(){const e=this.index,t=this.attributes;if(e===null||t.position===void 0||t.normal===void 0||t.uv===void 0){console.error("THREE.BufferGeometry: .computeTangents() failed. Missing required attributes (index, position, normal or uv)");return}const r=t.position,n=t.normal,a=t.uv;this.hasAttribute("tangent")===!1&&this.setAttribute("tangent",new Bt(new Float32Array(4*r.count),4));const o=this.getAttribute("tangent"),s=[],l=[];for(let O=0;O<r.count;O++)s[O]=new B,l[O]=new B;const c=new B,h=new B,u=new B,p=new je,f=new je,g=new je,_=new B,m=new B;function d(O,b,M){c.fromBufferAttribute(r,O),h.fromBufferAttribute(r,b),u.fromBufferAttribute(r,M),p.fromBufferAttribute(a,O),f.fromBufferAttribute(a,b),g.fromBufferAttribute(a,M),h.sub(c),u.sub(c),f.sub(p),g.sub(p);const R=1/(f.x*g.y-g.x*f.y);isFinite(R)&&(_.copy(h).multiplyScalar(g.y).addScaledVector(u,-f.y).multiplyScalar(R),m.copy(u).multiplyScalar(f.x).addScaledVector(h,-g.x).multiplyScalar(R),s[O].add(_),s[b].add(_),s[M].add(_),l[O].add(m),l[b].add(m),l[M].add(m))}let S=this.groups;S.length===0&&(S=[{start:0,count:e.count}]);for(let O=0,b=S.length;O<b;++O){const M=S[O],R=M.start,H=M.count;for(let U=R,F=R+H;U<F;U+=3)d(e.getX(U+0),e.getX(U+1),e.getX(U+2))}const T=new B,x=new B,N=new B,w=new B;function A(O){N.fromBufferAttribute(n,O),w.copy(N);const b=s[O];T.copy(b),T.sub(N.multiplyScalar(N.dot(b))).normalize(),x.crossVectors(w,b);const M=x.dot(l[O])<0?-1:1;o.setXYZW(O,T.x,T.y,T.z,M)}for(let O=0,b=S.length;O<b;++O){const M=S[O],R=M.start,H=M.count;for(let U=R,F=R+H;U<F;U+=3)A(e.getX(U+0)),A(e.getX(U+1)),A(e.getX(U+2))}}computeVertexNormals(){const e=this.index,t=this.getAttribute("position");if(t!==void 0){let r=this.getAttribute("normal");if(r===void 0)r=new Bt(new Float32Array(t.count*3),3),this.setAttribute("normal",r);else for(let p=0,f=r.count;p<f;p++)r.setXYZ(p,0,0,0);const n=new B,a=new B,o=new B,s=new B,l=new B,c=new B,h=new B,u=new B;if(e)for(let p=0,f=e.count;p<f;p+=3){const g=e.getX(p+0),_=e.getX(p+1),m=e.getX(p+2);n.fromBufferAttribute(t,g),a.fromBufferAttribute(t,_),o.fromBufferAttribute(t,m),h.subVectors(o,a),u.subVectors(n,a),h.cross(u),s.fromBufferAttribute(r,g),l.fromBufferAttribute(r,_),c.fromBufferAttribute(r,m),s.add(h),l.add(h),c.add(h),r.setXYZ(g,s.x,s.y,s.z),r.setXYZ(_,l.x,l.y,l.z),r.setXYZ(m,c.x,c.y,c.z)}else for(let p=0,f=t.count;p<f;p+=3)n.fromBufferAttribute(t,p+0),a.fromBufferAttribute(t,p+1),o.fromBufferAttribute(t,p+2),h.subVectors(o,a),u.subVectors(n,a),h.cross(u),r.setXYZ(p+0,h.x,h.y,h.z),r.setXYZ(p+1,h.x,h.y,h.z),r.setXYZ(p+2,h.x,h.y,h.z);this.normalizeNormals(),r.needsUpdate=!0}}normalizeNormals(){const e=this.attributes.normal;for(let t=0,r=e.count;t<r;t++)yt.fromBufferAttribute(e,t),yt.normalize(),e.setXYZ(t,yt.x,yt.y,yt.z)}toNonIndexed(){function e(s,l){const c=s.array,h=s.itemSize,u=s.normalized,p=new c.constructor(l.length*h);let f=0,g=0;for(let _=0,m=l.length;_<m;_++){s.isInterleavedBufferAttribute?f=l[_]*s.data.stride+s.offset:f=l[_]*h;for(let d=0;d<h;d++)p[g++]=c[f++]}return new Bt(p,h,u)}if(this.index===null)return console.warn("THREE.BufferGeometry.toNonIndexed(): BufferGeometry is already non-indexed."),this;const t=new Ct,r=this.index.array,n=this.attributes;for(const s in n){const l=n[s],c=e(l,r);t.setAttribute(s,c)}const a=this.morphAttributes;for(const s in a){const l=[],c=a[s];for(let h=0,u=c.length;h<u;h++){const p=c[h],f=e(p,r);l.push(f)}t.morphAttributes[s]=l}t.morphTargetsRelative=this.morphTargetsRelative;const o=this.groups;for(let s=0,l=o.length;s<l;s++){const c=o[s];t.addGroup(c.start,c.count,c.materialIndex)}return t}toJSON(){const e={metadata:{version:4.6,type:"BufferGeometry",generator:"BufferGeometry.toJSON"}};if(e.uuid=this.uuid,e.type=this.type,this.name!==""&&(e.name=this.name),Object.keys(this.userData).length>0&&(e.userData=this.userData),this.parameters!==void 0){const l=this.parameters;for(const c in l)l[c]!==void 0&&(e[c]=l[c]);return e}e.data={attributes:{}};const t=this.index;t!==null&&(e.data.index={type:t.array.constructor.name,array:Array.prototype.slice.call(t.array)});const r=this.attributes;for(const l in r){const c=r[l];e.data.attributes[l]=c.toJSON(e.data)}const n={};let a=!1;for(const l in this.morphAttributes){const c=this.morphAttributes[l],h=[];for(let u=0,p=c.length;u<p;u++){const f=c[u];h.push(f.toJSON(e.data))}h.length>0&&(n[l]=h,a=!0)}a&&(e.data.morphAttributes=n,e.data.morphTargetsRelative=this.morphTargetsRelative);const o=this.groups;o.length>0&&(e.data.groups=JSON.parse(JSON.stringify(o)));const s=this.boundingSphere;return s!==null&&(e.data.boundingSphere={center:s.center.toArray(),radius:s.radius}),e}clone(){return new this.constructor().copy(this)}copy(e){this.index=null,this.attributes={},this.morphAttributes={},this.groups=[],this.boundingBox=null,this.boundingSphere=null;const t={};this.name=e.name;const r=e.index;r!==null&&this.setIndex(r.clone(t));const n=e.attributes;for(const c in n){const h=n[c];this.setAttribute(c,h.clone(t))}const a=e.morphAttributes;for(const c in a){const h=[],u=a[c];for(let p=0,f=u.length;p<f;p++)h.push(u[p].clone(t));this.morphAttributes[c]=h}this.morphTargetsRelative=e.morphTargetsRelative;const o=e.groups;for(let c=0,h=o.length;c<h;c++){const u=o[c];this.addGroup(u.start,u.count,u.materialIndex)}const s=e.boundingBox;s!==null&&(this.boundingBox=s.clone());const l=e.boundingSphere;return l!==null&&(this.boundingSphere=l.clone()),this.drawRange.start=e.drawRange.start,this.drawRange.count=e.drawRange.count,this.userData=e.userData,this}dispose(){this.dispatchEvent({type:"dispose"})}}const ro=new Qe,Cr=new hl,sn=new Si,io=new B,on=new B,ln=new B,cn=new B,oa=new B,hn=new B,no=new B,un=new B;class Fe extends Mt{constructor(e=new Ct,t=new Bi){super(),this.isMesh=!0,this.type="Mesh",this.geometry=e,this.material=t,this.updateMorphTargets()}copy(e,t){return super.copy(e,t),e.morphTargetInfluences!==void 0&&(this.morphTargetInfluences=e.morphTargetInfluences.slice()),e.morphTargetDictionary!==void 0&&(this.morphTargetDictionary=Object.assign({},e.morphTargetDictionary)),this.material=Array.isArray(e.material)?e.material.slice():e.material,this.geometry=e.geometry,this}updateMorphTargets(){const e=this.geometry.morphAttributes,t=Object.keys(e);if(t.length>0){const r=e[t[0]];if(r!==void 0){this.morphTargetInfluences=[],this.morphTargetDictionary={};for(let n=0,a=r.length;n<a;n++){const o=r[n].name||String(n);this.morphTargetInfluences.push(0),this.morphTargetDictionary[o]=n}}}}getVertexPosition(e,t){const r=this.geometry,n=r.attributes.position,a=r.morphAttributes.position,o=r.morphTargetsRelative;t.fromBufferAttribute(n,e);const s=this.morphTargetInfluences;if(a&&s){hn.set(0,0,0);for(let l=0,c=a.length;l<c;l++){const h=s[l],u=a[l];h!==0&&(oa.fromBufferAttribute(u,e),o?hn.addScaledVector(oa,h):hn.addScaledVector(oa.sub(t),h))}t.add(hn)}return t}raycast(e,t){const r=this.geometry,n=this.material,a=this.matrixWorld;n!==void 0&&(r.boundingSphere===null&&r.computeBoundingSphere(),sn.copy(r.boundingSphere),sn.applyMatrix4(a),Cr.copy(e.ray).recast(e.near),!(sn.containsPoint(Cr.origin)===!1&&(Cr.intersectSphere(sn,io)===null||Cr.origin.distanceToSquared(io)>(e.far-e.near)**2))&&(ro.copy(a).invert(),Cr.copy(e.ray).applyMatrix4(ro),!(r.boundingBox!==null&&Cr.intersectsBox(r.boundingBox)===!1)&&this._computeIntersections(e,t,Cr)))}_computeIntersections(e,t,r){let n;const a=this.geometry,o=this.material,s=a.index,l=a.attributes.position,c=a.attributes.uv,h=a.attributes.uv1,u=a.attributes.normal,p=a.groups,f=a.drawRange;if(s!==null)if(Array.isArray(o))for(let g=0,_=p.length;g<_;g++){const m=p[g],d=o[m.materialIndex],S=Math.max(m.start,f.start),T=Math.min(s.count,Math.min(m.start+m.count,f.start+f.count));for(let x=S,N=T;x<N;x+=3){const w=s.getX(x),A=s.getX(x+1),O=s.getX(x+2);n=dn(this,d,e,r,c,h,u,w,A,O),n&&(n.faceIndex=Math.floor(x/3),n.face.materialIndex=m.materialIndex,t.push(n))}}else{const g=Math.max(0,f.start),_=Math.min(s.count,f.start+f.count);for(let m=g,d=_;m<d;m+=3){const S=s.getX(m),T=s.getX(m+1),x=s.getX(m+2);n=dn(this,o,e,r,c,h,u,S,T,x),n&&(n.faceIndex=Math.floor(m/3),t.push(n))}}else if(l!==void 0)if(Array.isArray(o))for(let g=0,_=p.length;g<_;g++){const m=p[g],d=o[m.materialIndex],S=Math.max(m.start,f.start),T=Math.min(l.count,Math.min(m.start+m.count,f.start+f.count));for(let x=S,N=T;x<N;x+=3){const w=x,A=x+1,O=x+2;n=dn(this,d,e,r,c,h,u,w,A,O),n&&(n.faceIndex=Math.floor(x/3),n.face.materialIndex=m.materialIndex,t.push(n))}}else{const g=Math.max(0,f.start),_=Math.min(l.count,f.start+f.count);for(let m=g,d=_;m<d;m+=3){const S=m,T=m+1,x=m+2;n=dn(this,o,e,r,c,h,u,S,T,x),n&&(n.faceIndex=Math.floor(m/3),t.push(n))}}}}function Wc(i,e,t,r,n,a,o,s){let l;if(e.side===Dt?l=r.intersectTriangle(o,a,n,!0,s):l=r.intersectTriangle(n,a,o,e.side===Sr,s),l===null)return null;un.copy(s),un.applyMatrix4(i.matrixWorld);const c=t.ray.origin.distanceTo(un);return c<t.near||c>t.far?null:{distance:c,point:un.clone(),object:i}}function dn(i,e,t,r,n,a,o,s,l,c){i.getVertexPosition(s,on),i.getVertexPosition(l,ln),i.getVertexPosition(c,cn);const h=Wc(i,e,t,r,on,ln,cn,no);if(h){const u=new B;zt.getBarycoord(no,on,ln,cn,u),n&&(h.uv=zt.getInterpolatedAttribute(n,s,l,c,u,new je)),a&&(h.uv1=zt.getInterpolatedAttribute(a,s,l,c,u,new je)),o&&(h.normal=zt.getInterpolatedAttribute(o,s,l,c,u,new B),h.normal.dot(r.direction)>0&&h.normal.multiplyScalar(-1));const p={a:s,b:l,c,normal:new B,materialIndex:0};zt.getNormal(on,ln,cn,p.normal),h.face=p,h.barycoord=u}return h}class tt extends Ct{constructor(e=1,t=1,r=1,n=1,a=1,o=1){super(),this.type="BoxGeometry",this.parameters={width:e,height:t,depth:r,widthSegments:n,heightSegments:a,depthSegments:o};const s=this;n=Math.floor(n),a=Math.floor(a),o=Math.floor(o);const l=[],c=[],h=[],u=[];let p=0,f=0;g("z","y","x",-1,-1,r,t,e,o,a,0),g("z","y","x",1,-1,r,t,-e,o,a,1),g("x","z","y",1,1,e,r,t,n,o,2),g("x","z","y",1,-1,e,r,-t,n,o,3),g("x","y","z",1,-1,e,t,r,n,a,4),g("x","y","z",-1,-1,e,t,-r,n,a,5),this.setIndex(l),this.setAttribute("position",new ht(c,3)),this.setAttribute("normal",new ht(h,3)),this.setAttribute("uv",new ht(u,2));function g(_,m,d,S,T,x,N,w,A,O,b){const M=x/A,R=N/O,H=x/2,U=N/2,F=w/2,K=A+1,X=O+1;let $=0,k=0;const Z=new B;for(let ce=0;ce<X;ce++){const ve=ce*R-U;for(let De=0;De<K;De++){const xe=De*M-H;Z[_]=xe*S,Z[m]=ve*T,Z[d]=F,c.push(Z.x,Z.y,Z.z),Z[_]=0,Z[m]=0,Z[d]=w>0?1:-1,h.push(Z.x,Z.y,Z.z),u.push(De/A),u.push(1-ce/O),$+=1}}for(let ce=0;ce<O;ce++)for(let ve=0;ve<A;ve++){const De=p+ve+K*ce,xe=p+ve+K*(ce+1),G=p+(ve+1)+K*(ce+1),J=p+(ve+1)+K*ce;l.push(De,xe,J),l.push(xe,G,J),k+=6}s.addGroup(f,k,b),f+=k,p+=$}}copy(e){return super.copy(e),this.parameters=Object.assign({},e.parameters),this}static fromJSON(e){return new tt(e.width,e.height,e.depth,e.widthSegments,e.heightSegments,e.depthSegments)}}function xi(i){const e={};for(const t in i){e[t]={};for(const r in i[t]){const n=i[t][r];n&&(n.isColor||n.isMatrix3||n.isMatrix4||n.isVector2||n.isVector3||n.isVector4||n.isTexture||n.isQuaternion)?n.isRenderTargetTexture?(console.warn("UniformsUtils: Textures of render targets cannot be cloned via cloneUniforms() or mergeUniforms()."),e[t][r]=null):e[t][r]=n.clone():Array.isArray(n)?e[t][r]=n.slice():e[t][r]=n}}return e}function At(i){const e={};for(let t=0;t<i.length;t++){const r=xi(i[t]);for(const n in r)e[n]=r[n]}return e}function jc(i){const e=[];for(let t=0;t<i.length;t++)e.push(i[t].clone());return e}function ml(i){const e=i.getRenderTarget();return e===null?i.outputColorSpace:e.isXRRenderTarget===!0?e.texture.colorSpace:$e.workingColorSpace}const Xc={clone:xi,merge:At};var Yc=`void main() {
	gl_Position = projectionMatrix * modelViewMatrix * vec4( position, 1.0 );
}`,qc=`void main() {
	gl_FragColor = vec4( 1.0, 0.0, 0.0, 1.0 );
}`;class Er extends Gr{static get type(){return"ShaderMaterial"}constructor(e){super(),this.isShaderMaterial=!0,this.defines={},this.uniforms={},this.uniformsGroups=[],this.vertexShader=Yc,this.fragmentShader=qc,this.linewidth=1,this.wireframe=!1,this.wireframeLinewidth=1,this.fog=!1,this.lights=!1,this.clipping=!1,this.forceSinglePass=!0,this.extensions={clipCullDistance:!1,multiDraw:!1},this.defaultAttributeValues={color:[1,1,1],uv:[0,0],uv1:[0,0]},this.index0AttributeName=void 0,this.uniformsNeedUpdate=!1,this.glslVersion=null,e!==void 0&&this.setValues(e)}copy(e){return super.copy(e),this.fragmentShader=e.fragmentShader,this.vertexShader=e.vertexShader,this.uniforms=xi(e.uniforms),this.uniformsGroups=jc(e.uniformsGroups),this.defines=Object.assign({},e.defines),this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this.fog=e.fog,this.lights=e.lights,this.clipping=e.clipping,this.extensions=Object.assign({},e.extensions),this.glslVersion=e.glslVersion,this}toJSON(e){const t=super.toJSON(e);t.glslVersion=this.glslVersion,t.uniforms={};for(const n in this.uniforms){const a=this.uniforms[n].value;a&&a.isTexture?t.uniforms[n]={type:"t",value:a.toJSON(e).uuid}:a&&a.isColor?t.uniforms[n]={type:"c",value:a.getHex()}:a&&a.isVector2?t.uniforms[n]={type:"v2",value:a.toArray()}:a&&a.isVector3?t.uniforms[n]={type:"v3",value:a.toArray()}:a&&a.isVector4?t.uniforms[n]={type:"v4",value:a.toArray()}:a&&a.isMatrix3?t.uniforms[n]={type:"m3",value:a.toArray()}:a&&a.isMatrix4?t.uniforms[n]={type:"m4",value:a.toArray()}:t.uniforms[n]={value:a}}Object.keys(this.defines).length>0&&(t.defines=this.defines),t.vertexShader=this.vertexShader,t.fragmentShader=this.fragmentShader,t.lights=this.lights,t.clipping=this.clipping;const r={};for(const n in this.extensions)this.extensions[n]===!0&&(r[n]=!0);return Object.keys(r).length>0&&(t.extensions=r),t}}class gl extends Mt{constructor(){super(),this.isCamera=!0,this.type="Camera",this.matrixWorldInverse=new Qe,this.projectionMatrix=new Qe,this.projectionMatrixInverse=new Qe,this.coordinateSystem=or}copy(e,t){return super.copy(e,t),this.matrixWorldInverse.copy(e.matrixWorldInverse),this.projectionMatrix.copy(e.projectionMatrix),this.projectionMatrixInverse.copy(e.projectionMatrixInverse),this.coordinateSystem=e.coordinateSystem,this}getWorldDirection(e){return super.getWorldDirection(e).negate()}updateMatrixWorld(e){super.updateMatrixWorld(e),this.matrixWorldInverse.copy(this.matrixWorld).invert()}updateWorldMatrix(e,t){super.updateWorldMatrix(e,t),this.matrixWorldInverse.copy(this.matrixWorld).invert()}clone(){return new this.constructor().copy(this)}}const gr=new B,ao=new je,so=new je;class Ft extends gl{constructor(e=50,t=1,r=.1,n=2e3){super(),this.isPerspectiveCamera=!0,this.type="PerspectiveCamera",this.fov=e,this.zoom=1,this.near=r,this.far=n,this.focus=10,this.aspect=t,this.view=null,this.filmGauge=35,this.filmOffset=0,this.updateProjectionMatrix()}copy(e,t){return super.copy(e,t),this.fov=e.fov,this.zoom=e.zoom,this.near=e.near,this.far=e.far,this.focus=e.focus,this.aspect=e.aspect,this.view=e.view===null?null:Object.assign({},e.view),this.filmGauge=e.filmGauge,this.filmOffset=e.filmOffset,this}setFocalLength(e){const t=.5*this.getFilmHeight()/e;this.fov=ss*2*Math.atan(t),this.updateProjectionMatrix()}getFocalLength(){const e=Math.tan(Vn*.5*this.fov);return .5*this.getFilmHeight()/e}getEffectiveFOV(){return ss*2*Math.atan(Math.tan(Vn*.5*this.fov)/this.zoom)}getFilmWidth(){return this.filmGauge*Math.min(this.aspect,1)}getFilmHeight(){return this.filmGauge/Math.max(this.aspect,1)}getViewBounds(e,t,r){gr.set(-1,-1,.5).applyMatrix4(this.projectionMatrixInverse),t.set(gr.x,gr.y).multiplyScalar(-e/gr.z),gr.set(1,1,.5).applyMatrix4(this.projectionMatrixInverse),r.set(gr.x,gr.y).multiplyScalar(-e/gr.z)}getViewSize(e,t){return this.getViewBounds(e,ao,so),t.subVectors(so,ao)}setViewOffset(e,t,r,n,a,o){this.aspect=e/t,this.view===null&&(this.view={enabled:!0,fullWidth:1,fullHeight:1,offsetX:0,offsetY:0,width:1,height:1}),this.view.enabled=!0,this.view.fullWidth=e,this.view.fullHeight=t,this.view.offsetX=r,this.view.offsetY=n,this.view.width=a,this.view.height=o,this.updateProjectionMatrix()}clearViewOffset(){this.view!==null&&(this.view.enabled=!1),this.updateProjectionMatrix()}updateProjectionMatrix(){const e=this.near;let t=e*Math.tan(Vn*.5*this.fov)/this.zoom,r=2*t,n=this.aspect*r,a=-.5*n;const o=this.view;if(this.view!==null&&this.view.enabled){const l=o.fullWidth,c=o.fullHeight;a+=o.offsetX*n/l,t-=o.offsetY*r/c,n*=o.width/l,r*=o.height/c}const s=this.filmOffset;s!==0&&(a+=e*s/this.getFilmWidth()),this.projectionMatrix.makePerspective(a,a+n,t,t-r,e,this.far,this.coordinateSystem),this.projectionMatrixInverse.copy(this.projectionMatrix).invert()}toJSON(e){const t=super.toJSON(e);return t.object.fov=this.fov,t.object.zoom=this.zoom,t.object.near=this.near,t.object.far=this.far,t.object.focus=this.focus,t.object.aspect=this.aspect,this.view!==null&&(t.object.view=Object.assign({},this.view)),t.object.filmGauge=this.filmGauge,t.object.filmOffset=this.filmOffset,t}}const ri=-90,ii=1;class Kc extends Mt{constructor(e,t,r){super(),this.type="CubeCamera",this.renderTarget=r,this.coordinateSystem=null,this.activeMipmapLevel=0;const n=new Ft(ri,ii,e,t);n.layers=this.layers,this.add(n);const a=new Ft(ri,ii,e,t);a.layers=this.layers,this.add(a);const o=new Ft(ri,ii,e,t);o.layers=this.layers,this.add(o);const s=new Ft(ri,ii,e,t);s.layers=this.layers,this.add(s);const l=new Ft(ri,ii,e,t);l.layers=this.layers,this.add(l);const c=new Ft(ri,ii,e,t);c.layers=this.layers,this.add(c)}updateCoordinateSystem(){const e=this.coordinateSystem,t=this.children.concat(),[r,n,a,o,s,l]=t;for(const c of t)this.remove(c);if(e===or)r.up.set(0,1,0),r.lookAt(1,0,0),n.up.set(0,1,0),n.lookAt(-1,0,0),a.up.set(0,0,-1),a.lookAt(0,1,0),o.up.set(0,0,1),o.lookAt(0,-1,0),s.up.set(0,1,0),s.lookAt(0,0,1),l.up.set(0,1,0),l.lookAt(0,0,-1);else if(e===Pn)r.up.set(0,-1,0),r.lookAt(-1,0,0),n.up.set(0,-1,0),n.lookAt(1,0,0),a.up.set(0,0,1),a.lookAt(0,1,0),o.up.set(0,0,-1),o.lookAt(0,-1,0),s.up.set(0,-1,0),s.lookAt(0,0,1),l.up.set(0,-1,0),l.lookAt(0,0,-1);else throw new Error("THREE.CubeCamera.updateCoordinateSystem(): Invalid coordinate system: "+e);for(const c of t)this.add(c),c.updateMatrixWorld()}update(e,t){this.parent===null&&this.updateMatrixWorld();const{renderTarget:r,activeMipmapLevel:n}=this;this.coordinateSystem!==e.coordinateSystem&&(this.coordinateSystem=e.coordinateSystem,this.updateCoordinateSystem());const[a,o,s,l,c,h]=this.children,u=e.getRenderTarget(),p=e.getActiveCubeFace(),f=e.getActiveMipmapLevel(),g=e.xr.enabled;e.xr.enabled=!1;const _=r.texture.generateMipmaps;r.texture.generateMipmaps=!1,e.setRenderTarget(r,0,n),e.render(t,a),e.setRenderTarget(r,1,n),e.render(t,o),e.setRenderTarget(r,2,n),e.render(t,s),e.setRenderTarget(r,3,n),e.render(t,l),e.setRenderTarget(r,4,n),e.render(t,c),r.texture.generateMipmaps=_,e.setRenderTarget(r,5,n),e.render(t,h),e.setRenderTarget(u,p,f),e.xr.enabled=g,r.texture.needsPMREMUpdate=!0}}class _l extends _t{constructor(e,t,r,n,a,o,s,l,c,h){e=e!==void 0?e:[],t=t!==void 0?t:mi,super(e,t,r,n,a,o,s,l,c,h),this.isCubeTexture=!0,this.flipY=!1}get images(){return this.image}set images(e){this.image=e}}class $c extends Vr{constructor(e=1,t={}){super(e,e,t),this.isWebGLCubeRenderTarget=!0;const r={width:e,height:e,depth:1},n=[r,r,r,r,r,r];this.texture=new _l(n,t.mapping,t.wrapS,t.wrapT,t.magFilter,t.minFilter,t.format,t.type,t.anisotropy,t.colorSpace),this.texture.isRenderTargetTexture=!0,this.texture.generateMipmaps=t.generateMipmaps!==void 0?t.generateMipmaps:!1,this.texture.minFilter=t.minFilter!==void 0?t.minFilter:jt}fromEquirectangularTexture(e,t){this.texture.type=t.type,this.texture.colorSpace=t.colorSpace,this.texture.generateMipmaps=t.generateMipmaps,this.texture.minFilter=t.minFilter,this.texture.magFilter=t.magFilter;const r={uniforms:{tEquirect:{value:null}},vertexShader:`

				varying vec3 vWorldDirection;

				vec3 transformDirection( in vec3 dir, in mat4 matrix ) {

					return normalize( ( matrix * vec4( dir, 0.0 ) ).xyz );

				}

				void main() {

					vWorldDirection = transformDirection( position, modelMatrix );

					#include <begin_vertex>
					#include <project_vertex>

				}
			`,fragmentShader:`

				uniform sampler2D tEquirect;

				varying vec3 vWorldDirection;

				#include <common>

				void main() {

					vec3 direction = normalize( vWorldDirection );

					vec2 sampleUV = equirectUv( direction );

					gl_FragColor = texture2D( tEquirect, sampleUV );

				}
			`},n=new tt(5,5,5),a=new Er({name:"CubemapFromEquirect",uniforms:xi(r.uniforms),vertexShader:r.vertexShader,fragmentShader:r.fragmentShader,side:Dt,blending:xr});a.uniforms.tEquirect.value=t;const o=new Fe(n,a),s=t.minFilter;return t.minFilter===Br&&(t.minFilter=jt),new Kc(1,10,this).update(e,o),t.minFilter=s,o.geometry.dispose(),o.material.dispose(),this}clear(e,t,r,n){const a=e.getRenderTarget();for(let o=0;o<6;o++)e.setRenderTarget(this,o),e.clear(t,r,n);e.setRenderTarget(a)}}const la=new B,Zc=new B,Jc=new Ve;class Nr{constructor(e=new B(1,0,0),t=0){this.isPlane=!0,this.normal=e,this.constant=t}set(e,t){return this.normal.copy(e),this.constant=t,this}setComponents(e,t,r,n){return this.normal.set(e,t,r),this.constant=n,this}setFromNormalAndCoplanarPoint(e,t){return this.normal.copy(e),this.constant=-t.dot(this.normal),this}setFromCoplanarPoints(e,t,r){const n=la.subVectors(r,t).cross(Zc.subVectors(e,t)).normalize();return this.setFromNormalAndCoplanarPoint(n,e),this}copy(e){return this.normal.copy(e.normal),this.constant=e.constant,this}normalize(){const e=1/this.normal.length();return this.normal.multiplyScalar(e),this.constant*=e,this}negate(){return this.constant*=-1,this.normal.negate(),this}distanceToPoint(e){return this.normal.dot(e)+this.constant}distanceToSphere(e){return this.distanceToPoint(e.center)-e.radius}projectPoint(e,t){return t.copy(e).addScaledVector(this.normal,-this.distanceToPoint(e))}intersectLine(e,t){const r=e.delta(la),n=this.normal.dot(r);if(n===0)return this.distanceToPoint(e.start)===0?t.copy(e.start):null;const a=-(e.start.dot(this.normal)+this.constant)/n;return a<0||a>1?null:t.copy(e.start).addScaledVector(r,a)}intersectsLine(e){const t=this.distanceToPoint(e.start),r=this.distanceToPoint(e.end);return t<0&&r>0||r<0&&t>0}intersectsBox(e){return e.intersectsPlane(this)}intersectsSphere(e){return e.intersectsPlane(this)}coplanarPoint(e){return e.copy(this.normal).multiplyScalar(-this.constant)}applyMatrix4(e,t){const r=t||Jc.getNormalMatrix(e),n=this.coplanarPoint(la).applyMatrix4(e),a=this.normal.applyMatrix3(r).normalize();return this.constant=-n.dot(a),this}translate(e){return this.constant-=e.dot(this.normal),this}equals(e){return e.normal.equals(this.normal)&&e.constant===this.constant}clone(){return new this.constructor().copy(this)}}const Lr=new Si,pn=new B;class ys{constructor(e=new Nr,t=new Nr,r=new Nr,n=new Nr,a=new Nr,o=new Nr){this.planes=[e,t,r,n,a,o]}set(e,t,r,n,a,o){const s=this.planes;return s[0].copy(e),s[1].copy(t),s[2].copy(r),s[3].copy(n),s[4].copy(a),s[5].copy(o),this}copy(e){const t=this.planes;for(let r=0;r<6;r++)t[r].copy(e.planes[r]);return this}setFromProjectionMatrix(e,t=or){const r=this.planes,n=e.elements,a=n[0],o=n[1],s=n[2],l=n[3],c=n[4],h=n[5],u=n[6],p=n[7],f=n[8],g=n[9],_=n[10],m=n[11],d=n[12],S=n[13],T=n[14],x=n[15];if(r[0].setComponents(l-a,p-c,m-f,x-d).normalize(),r[1].setComponents(l+a,p+c,m+f,x+d).normalize(),r[2].setComponents(l+o,p+h,m+g,x+S).normalize(),r[3].setComponents(l-o,p-h,m-g,x-S).normalize(),r[4].setComponents(l-s,p-u,m-_,x-T).normalize(),t===or)r[5].setComponents(l+s,p+u,m+_,x+T).normalize();else if(t===Pn)r[5].setComponents(s,u,_,T).normalize();else throw new Error("THREE.Frustum.setFromProjectionMatrix(): Invalid coordinate system: "+t);return this}intersectsObject(e){if(e.boundingSphere!==void 0)e.boundingSphere===null&&e.computeBoundingSphere(),Lr.copy(e.boundingSphere).applyMatrix4(e.matrixWorld);else{const t=e.geometry;t.boundingSphere===null&&t.computeBoundingSphere(),Lr.copy(t.boundingSphere).applyMatrix4(e.matrixWorld)}return this.intersectsSphere(Lr)}intersectsSprite(e){return Lr.center.set(0,0,0),Lr.radius=.7071067811865476,Lr.applyMatrix4(e.matrixWorld),this.intersectsSphere(Lr)}intersectsSphere(e){const t=this.planes,r=e.center,n=-e.radius;for(let a=0;a<6;a++)if(t[a].distanceToPoint(r)<n)return!1;return!0}intersectsBox(e){const t=this.planes;for(let r=0;r<6;r++){const n=t[r];if(pn.x=n.normal.x>0?e.max.x:e.min.x,pn.y=n.normal.y>0?e.max.y:e.min.y,pn.z=n.normal.z>0?e.max.z:e.min.z,n.distanceToPoint(pn)<0)return!1}return!0}containsPoint(e){const t=this.planes;for(let r=0;r<6;r++)if(t[r].distanceToPoint(e)<0)return!1;return!0}clone(){return new this.constructor().copy(this)}}function vl(){let i=null,e=!1,t=null,r=null;function n(a,o){t(a,o),r=i.requestAnimationFrame(n)}return{start:function(){e!==!0&&t!==null&&(r=i.requestAnimationFrame(n),e=!0)},stop:function(){i.cancelAnimationFrame(r),e=!1},setAnimationLoop:function(a){t=a},setContext:function(a){i=a}}}function Qc(i){const e=new WeakMap;function t(s,l){const c=s.array,h=s.usage,u=c.byteLength,p=i.createBuffer();i.bindBuffer(l,p),i.bufferData(l,c,h),s.onUploadCallback();let f;if(c instanceof Float32Array)f=i.FLOAT;else if(c instanceof Uint16Array)s.isFloat16BufferAttribute?f=i.HALF_FLOAT:f=i.UNSIGNED_SHORT;else if(c instanceof Int16Array)f=i.SHORT;else if(c instanceof Uint32Array)f=i.UNSIGNED_INT;else if(c instanceof Int32Array)f=i.INT;else if(c instanceof Int8Array)f=i.BYTE;else if(c instanceof Uint8Array)f=i.UNSIGNED_BYTE;else if(c instanceof Uint8ClampedArray)f=i.UNSIGNED_BYTE;else throw new Error("THREE.WebGLAttributes: Unsupported buffer data format: "+c);return{buffer:p,type:f,bytesPerElement:c.BYTES_PER_ELEMENT,version:s.version,size:u}}function r(s,l,c){const h=l.array,u=l.updateRanges;if(i.bindBuffer(c,s),u.length===0)i.bufferSubData(c,0,h);else{u.sort((f,g)=>f.start-g.start);let p=0;for(let f=1;f<u.length;f++){const g=u[p],_=u[f];_.start<=g.start+g.count+1?g.count=Math.max(g.count,_.start+_.count-g.start):(++p,u[p]=_)}u.length=p+1;for(let f=0,g=u.length;f<g;f++){const _=u[f];i.bufferSubData(c,_.start*h.BYTES_PER_ELEMENT,h,_.start,_.count)}l.clearUpdateRanges()}l.onUploadCallback()}function n(s){return s.isInterleavedBufferAttribute&&(s=s.data),e.get(s)}function a(s){s.isInterleavedBufferAttribute&&(s=s.data);const l=e.get(s);l&&(i.deleteBuffer(l.buffer),e.delete(s))}function o(s,l){if(s.isInterleavedBufferAttribute&&(s=s.data),s.isGLBufferAttribute){const h=e.get(s);(!h||h.version<s.version)&&e.set(s,{buffer:s.buffer,type:s.type,bytesPerElement:s.elementSize,version:s.version});return}const c=e.get(s);if(c===void 0)e.set(s,t(s,l));else if(c.version<s.version){if(c.size!==s.array.byteLength)throw new Error("THREE.WebGLAttributes: The size of the buffer attribute's array buffer does not match the original size. Resizing buffer attributes is not supported.");r(c.buffer,s,l),c.version=s.version}}return{get:n,remove:a,update:o}}class vr extends Ct{constructor(e=1,t=1,r=1,n=1){super(),this.type="PlaneGeometry",this.parameters={width:e,height:t,widthSegments:r,heightSegments:n};const a=e/2,o=t/2,s=Math.floor(r),l=Math.floor(n),c=s+1,h=l+1,u=e/s,p=t/l,f=[],g=[],_=[],m=[];for(let d=0;d<h;d++){const S=d*p-o;for(let T=0;T<c;T++){const x=T*u-a;g.push(x,-S,0),_.push(0,0,1),m.push(T/s),m.push(1-d/l)}}for(let d=0;d<l;d++)for(let S=0;S<s;S++){const T=S+c*d,x=S+c*(d+1),N=S+1+c*(d+1),w=S+1+c*d;f.push(T,x,w),f.push(x,N,w)}this.setIndex(f),this.setAttribute("position",new ht(g,3)),this.setAttribute("normal",new ht(_,3)),this.setAttribute("uv",new ht(m,2))}copy(e){return super.copy(e),this.parameters=Object.assign({},e.parameters),this}static fromJSON(e){return new vr(e.width,e.height,e.widthSegments,e.heightSegments)}}var eh=`#ifdef USE_ALPHAHASH
	if ( diffuseColor.a < getAlphaHashThreshold( vPosition ) ) discard;
#endif`,th=`#ifdef USE_ALPHAHASH
	const float ALPHA_HASH_SCALE = 0.05;
	float hash2D( vec2 value ) {
		return fract( 1.0e4 * sin( 17.0 * value.x + 0.1 * value.y ) * ( 0.1 + abs( sin( 13.0 * value.y + value.x ) ) ) );
	}
	float hash3D( vec3 value ) {
		return hash2D( vec2( hash2D( value.xy ), value.z ) );
	}
	float getAlphaHashThreshold( vec3 position ) {
		float maxDeriv = max(
			length( dFdx( position.xyz ) ),
			length( dFdy( position.xyz ) )
		);
		float pixScale = 1.0 / ( ALPHA_HASH_SCALE * maxDeriv );
		vec2 pixScales = vec2(
			exp2( floor( log2( pixScale ) ) ),
			exp2( ceil( log2( pixScale ) ) )
		);
		vec2 alpha = vec2(
			hash3D( floor( pixScales.x * position.xyz ) ),
			hash3D( floor( pixScales.y * position.xyz ) )
		);
		float lerpFactor = fract( log2( pixScale ) );
		float x = ( 1.0 - lerpFactor ) * alpha.x + lerpFactor * alpha.y;
		float a = min( lerpFactor, 1.0 - lerpFactor );
		vec3 cases = vec3(
			x * x / ( 2.0 * a * ( 1.0 - a ) ),
			( x - 0.5 * a ) / ( 1.0 - a ),
			1.0 - ( ( 1.0 - x ) * ( 1.0 - x ) / ( 2.0 * a * ( 1.0 - a ) ) )
		);
		float threshold = ( x < ( 1.0 - a ) )
			? ( ( x < a ) ? cases.x : cases.y )
			: cases.z;
		return clamp( threshold , 1.0e-6, 1.0 );
	}
#endif`,rh=`#ifdef USE_ALPHAMAP
	diffuseColor.a *= texture2D( alphaMap, vAlphaMapUv ).g;
#endif`,ih=`#ifdef USE_ALPHAMAP
	uniform sampler2D alphaMap;
#endif`,nh=`#ifdef USE_ALPHATEST
	#ifdef ALPHA_TO_COVERAGE
	diffuseColor.a = smoothstep( alphaTest, alphaTest + fwidth( diffuseColor.a ), diffuseColor.a );
	if ( diffuseColor.a == 0.0 ) discard;
	#else
	if ( diffuseColor.a < alphaTest ) discard;
	#endif
#endif`,ah=`#ifdef USE_ALPHATEST
	uniform float alphaTest;
#endif`,sh=`#ifdef USE_AOMAP
	float ambientOcclusion = ( texture2D( aoMap, vAoMapUv ).r - 1.0 ) * aoMapIntensity + 1.0;
	reflectedLight.indirectDiffuse *= ambientOcclusion;
	#if defined( USE_CLEARCOAT ) 
		clearcoatSpecularIndirect *= ambientOcclusion;
	#endif
	#if defined( USE_SHEEN ) 
		sheenSpecularIndirect *= ambientOcclusion;
	#endif
	#if defined( USE_ENVMAP ) && defined( STANDARD )
		float dotNV = saturate( dot( geometryNormal, geometryViewDir ) );
		reflectedLight.indirectSpecular *= computeSpecularOcclusion( dotNV, ambientOcclusion, material.roughness );
	#endif
#endif`,oh=`#ifdef USE_AOMAP
	uniform sampler2D aoMap;
	uniform float aoMapIntensity;
#endif`,lh=`#ifdef USE_BATCHING
	#if ! defined( GL_ANGLE_multi_draw )
	#define gl_DrawID _gl_DrawID
	uniform int _gl_DrawID;
	#endif
	uniform highp sampler2D batchingTexture;
	uniform highp usampler2D batchingIdTexture;
	mat4 getBatchingMatrix( const in float i ) {
		int size = textureSize( batchingTexture, 0 ).x;
		int j = int( i ) * 4;
		int x = j % size;
		int y = j / size;
		vec4 v1 = texelFetch( batchingTexture, ivec2( x, y ), 0 );
		vec4 v2 = texelFetch( batchingTexture, ivec2( x + 1, y ), 0 );
		vec4 v3 = texelFetch( batchingTexture, ivec2( x + 2, y ), 0 );
		vec4 v4 = texelFetch( batchingTexture, ivec2( x + 3, y ), 0 );
		return mat4( v1, v2, v3, v4 );
	}
	float getIndirectIndex( const in int i ) {
		int size = textureSize( batchingIdTexture, 0 ).x;
		int x = i % size;
		int y = i / size;
		return float( texelFetch( batchingIdTexture, ivec2( x, y ), 0 ).r );
	}
#endif
#ifdef USE_BATCHING_COLOR
	uniform sampler2D batchingColorTexture;
	vec3 getBatchingColor( const in float i ) {
		int size = textureSize( batchingColorTexture, 0 ).x;
		int j = int( i );
		int x = j % size;
		int y = j / size;
		return texelFetch( batchingColorTexture, ivec2( x, y ), 0 ).rgb;
	}
#endif`,ch=`#ifdef USE_BATCHING
	mat4 batchingMatrix = getBatchingMatrix( getIndirectIndex( gl_DrawID ) );
#endif`,hh=`vec3 transformed = vec3( position );
#ifdef USE_ALPHAHASH
	vPosition = vec3( position );
#endif`,uh=`vec3 objectNormal = vec3( normal );
#ifdef USE_TANGENT
	vec3 objectTangent = vec3( tangent.xyz );
#endif`,dh=`float G_BlinnPhong_Implicit( ) {
	return 0.25;
}
float D_BlinnPhong( const in float shininess, const in float dotNH ) {
	return RECIPROCAL_PI * ( shininess * 0.5 + 1.0 ) * pow( dotNH, shininess );
}
vec3 BRDF_BlinnPhong( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in vec3 specularColor, const in float shininess ) {
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNH = saturate( dot( normal, halfDir ) );
	float dotVH = saturate( dot( viewDir, halfDir ) );
	vec3 F = F_Schlick( specularColor, 1.0, dotVH );
	float G = G_BlinnPhong_Implicit( );
	float D = D_BlinnPhong( shininess, dotNH );
	return F * ( G * D );
} // validated`,ph=`#ifdef USE_IRIDESCENCE
	const mat3 XYZ_TO_REC709 = mat3(
		 3.2404542, -0.9692660,  0.0556434,
		-1.5371385,  1.8760108, -0.2040259,
		-0.4985314,  0.0415560,  1.0572252
	);
	vec3 Fresnel0ToIor( vec3 fresnel0 ) {
		vec3 sqrtF0 = sqrt( fresnel0 );
		return ( vec3( 1.0 ) + sqrtF0 ) / ( vec3( 1.0 ) - sqrtF0 );
	}
	vec3 IorToFresnel0( vec3 transmittedIor, float incidentIor ) {
		return pow2( ( transmittedIor - vec3( incidentIor ) ) / ( transmittedIor + vec3( incidentIor ) ) );
	}
	float IorToFresnel0( float transmittedIor, float incidentIor ) {
		return pow2( ( transmittedIor - incidentIor ) / ( transmittedIor + incidentIor ));
	}
	vec3 evalSensitivity( float OPD, vec3 shift ) {
		float phase = 2.0 * PI * OPD * 1.0e-9;
		vec3 val = vec3( 5.4856e-13, 4.4201e-13, 5.2481e-13 );
		vec3 pos = vec3( 1.6810e+06, 1.7953e+06, 2.2084e+06 );
		vec3 var = vec3( 4.3278e+09, 9.3046e+09, 6.6121e+09 );
		vec3 xyz = val * sqrt( 2.0 * PI * var ) * cos( pos * phase + shift ) * exp( - pow2( phase ) * var );
		xyz.x += 9.7470e-14 * sqrt( 2.0 * PI * 4.5282e+09 ) * cos( 2.2399e+06 * phase + shift[ 0 ] ) * exp( - 4.5282e+09 * pow2( phase ) );
		xyz /= 1.0685e-7;
		vec3 rgb = XYZ_TO_REC709 * xyz;
		return rgb;
	}
	vec3 evalIridescence( float outsideIOR, float eta2, float cosTheta1, float thinFilmThickness, vec3 baseF0 ) {
		vec3 I;
		float iridescenceIOR = mix( outsideIOR, eta2, smoothstep( 0.0, 0.03, thinFilmThickness ) );
		float sinTheta2Sq = pow2( outsideIOR / iridescenceIOR ) * ( 1.0 - pow2( cosTheta1 ) );
		float cosTheta2Sq = 1.0 - sinTheta2Sq;
		if ( cosTheta2Sq < 0.0 ) {
			return vec3( 1.0 );
		}
		float cosTheta2 = sqrt( cosTheta2Sq );
		float R0 = IorToFresnel0( iridescenceIOR, outsideIOR );
		float R12 = F_Schlick( R0, 1.0, cosTheta1 );
		float T121 = 1.0 - R12;
		float phi12 = 0.0;
		if ( iridescenceIOR < outsideIOR ) phi12 = PI;
		float phi21 = PI - phi12;
		vec3 baseIOR = Fresnel0ToIor( clamp( baseF0, 0.0, 0.9999 ) );		vec3 R1 = IorToFresnel0( baseIOR, iridescenceIOR );
		vec3 R23 = F_Schlick( R1, 1.0, cosTheta2 );
		vec3 phi23 = vec3( 0.0 );
		if ( baseIOR[ 0 ] < iridescenceIOR ) phi23[ 0 ] = PI;
		if ( baseIOR[ 1 ] < iridescenceIOR ) phi23[ 1 ] = PI;
		if ( baseIOR[ 2 ] < iridescenceIOR ) phi23[ 2 ] = PI;
		float OPD = 2.0 * iridescenceIOR * thinFilmThickness * cosTheta2;
		vec3 phi = vec3( phi21 ) + phi23;
		vec3 R123 = clamp( R12 * R23, 1e-5, 0.9999 );
		vec3 r123 = sqrt( R123 );
		vec3 Rs = pow2( T121 ) * R23 / ( vec3( 1.0 ) - R123 );
		vec3 C0 = R12 + Rs;
		I = C0;
		vec3 Cm = Rs - T121;
		for ( int m = 1; m <= 2; ++ m ) {
			Cm *= r123;
			vec3 Sm = 2.0 * evalSensitivity( float( m ) * OPD, float( m ) * phi );
			I += Cm * Sm;
		}
		return max( I, vec3( 0.0 ) );
	}
#endif`,fh=`#ifdef USE_BUMPMAP
	uniform sampler2D bumpMap;
	uniform float bumpScale;
	vec2 dHdxy_fwd() {
		vec2 dSTdx = dFdx( vBumpMapUv );
		vec2 dSTdy = dFdy( vBumpMapUv );
		float Hll = bumpScale * texture2D( bumpMap, vBumpMapUv ).x;
		float dBx = bumpScale * texture2D( bumpMap, vBumpMapUv + dSTdx ).x - Hll;
		float dBy = bumpScale * texture2D( bumpMap, vBumpMapUv + dSTdy ).x - Hll;
		return vec2( dBx, dBy );
	}
	vec3 perturbNormalArb( vec3 surf_pos, vec3 surf_norm, vec2 dHdxy, float faceDirection ) {
		vec3 vSigmaX = normalize( dFdx( surf_pos.xyz ) );
		vec3 vSigmaY = normalize( dFdy( surf_pos.xyz ) );
		vec3 vN = surf_norm;
		vec3 R1 = cross( vSigmaY, vN );
		vec3 R2 = cross( vN, vSigmaX );
		float fDet = dot( vSigmaX, R1 ) * faceDirection;
		vec3 vGrad = sign( fDet ) * ( dHdxy.x * R1 + dHdxy.y * R2 );
		return normalize( abs( fDet ) * surf_norm - vGrad );
	}
#endif`,mh=`#if NUM_CLIPPING_PLANES > 0
	vec4 plane;
	#ifdef ALPHA_TO_COVERAGE
		float distanceToPlane, distanceGradient;
		float clipOpacity = 1.0;
		#pragma unroll_loop_start
		for ( int i = 0; i < UNION_CLIPPING_PLANES; i ++ ) {
			plane = clippingPlanes[ i ];
			distanceToPlane = - dot( vClipPosition, plane.xyz ) + plane.w;
			distanceGradient = fwidth( distanceToPlane ) / 2.0;
			clipOpacity *= smoothstep( - distanceGradient, distanceGradient, distanceToPlane );
			if ( clipOpacity == 0.0 ) discard;
		}
		#pragma unroll_loop_end
		#if UNION_CLIPPING_PLANES < NUM_CLIPPING_PLANES
			float unionClipOpacity = 1.0;
			#pragma unroll_loop_start
			for ( int i = UNION_CLIPPING_PLANES; i < NUM_CLIPPING_PLANES; i ++ ) {
				plane = clippingPlanes[ i ];
				distanceToPlane = - dot( vClipPosition, plane.xyz ) + plane.w;
				distanceGradient = fwidth( distanceToPlane ) / 2.0;
				unionClipOpacity *= 1.0 - smoothstep( - distanceGradient, distanceGradient, distanceToPlane );
			}
			#pragma unroll_loop_end
			clipOpacity *= 1.0 - unionClipOpacity;
		#endif
		diffuseColor.a *= clipOpacity;
		if ( diffuseColor.a == 0.0 ) discard;
	#else
		#pragma unroll_loop_start
		for ( int i = 0; i < UNION_CLIPPING_PLANES; i ++ ) {
			plane = clippingPlanes[ i ];
			if ( dot( vClipPosition, plane.xyz ) > plane.w ) discard;
		}
		#pragma unroll_loop_end
		#if UNION_CLIPPING_PLANES < NUM_CLIPPING_PLANES
			bool clipped = true;
			#pragma unroll_loop_start
			for ( int i = UNION_CLIPPING_PLANES; i < NUM_CLIPPING_PLANES; i ++ ) {
				plane = clippingPlanes[ i ];
				clipped = ( dot( vClipPosition, plane.xyz ) > plane.w ) && clipped;
			}
			#pragma unroll_loop_end
			if ( clipped ) discard;
		#endif
	#endif
#endif`,gh=`#if NUM_CLIPPING_PLANES > 0
	varying vec3 vClipPosition;
	uniform vec4 clippingPlanes[ NUM_CLIPPING_PLANES ];
#endif`,_h=`#if NUM_CLIPPING_PLANES > 0
	varying vec3 vClipPosition;
#endif`,vh=`#if NUM_CLIPPING_PLANES > 0
	vClipPosition = - mvPosition.xyz;
#endif`,xh=`#if defined( USE_COLOR_ALPHA )
	diffuseColor *= vColor;
#elif defined( USE_COLOR )
	diffuseColor.rgb *= vColor;
#endif`,yh=`#if defined( USE_COLOR_ALPHA )
	varying vec4 vColor;
#elif defined( USE_COLOR )
	varying vec3 vColor;
#endif`,Mh=`#if defined( USE_COLOR_ALPHA )
	varying vec4 vColor;
#elif defined( USE_COLOR ) || defined( USE_INSTANCING_COLOR ) || defined( USE_BATCHING_COLOR )
	varying vec3 vColor;
#endif`,Sh=`#if defined( USE_COLOR_ALPHA )
	vColor = vec4( 1.0 );
#elif defined( USE_COLOR ) || defined( USE_INSTANCING_COLOR ) || defined( USE_BATCHING_COLOR )
	vColor = vec3( 1.0 );
#endif
#ifdef USE_COLOR
	vColor *= color;
#endif
#ifdef USE_INSTANCING_COLOR
	vColor.xyz *= instanceColor.xyz;
#endif
#ifdef USE_BATCHING_COLOR
	vec3 batchingColor = getBatchingColor( getIndirectIndex( gl_DrawID ) );
	vColor.xyz *= batchingColor.xyz;
#endif`,Eh=`#define PI 3.141592653589793
#define PI2 6.283185307179586
#define PI_HALF 1.5707963267948966
#define RECIPROCAL_PI 0.3183098861837907
#define RECIPROCAL_PI2 0.15915494309189535
#define EPSILON 1e-6
#ifndef saturate
#define saturate( a ) clamp( a, 0.0, 1.0 )
#endif
#define whiteComplement( a ) ( 1.0 - saturate( a ) )
float pow2( const in float x ) { return x*x; }
vec3 pow2( const in vec3 x ) { return x*x; }
float pow3( const in float x ) { return x*x*x; }
float pow4( const in float x ) { float x2 = x*x; return x2*x2; }
float max3( const in vec3 v ) { return max( max( v.x, v.y ), v.z ); }
float average( const in vec3 v ) { return dot( v, vec3( 0.3333333 ) ); }
highp float rand( const in vec2 uv ) {
	const highp float a = 12.9898, b = 78.233, c = 43758.5453;
	highp float dt = dot( uv.xy, vec2( a,b ) ), sn = mod( dt, PI );
	return fract( sin( sn ) * c );
}
#ifdef HIGH_PRECISION
	float precisionSafeLength( vec3 v ) { return length( v ); }
#else
	float precisionSafeLength( vec3 v ) {
		float maxComponent = max3( abs( v ) );
		return length( v / maxComponent ) * maxComponent;
	}
#endif
struct IncidentLight {
	vec3 color;
	vec3 direction;
	bool visible;
};
struct ReflectedLight {
	vec3 directDiffuse;
	vec3 directSpecular;
	vec3 indirectDiffuse;
	vec3 indirectSpecular;
};
#ifdef USE_ALPHAHASH
	varying vec3 vPosition;
#endif
vec3 transformDirection( in vec3 dir, in mat4 matrix ) {
	return normalize( ( matrix * vec4( dir, 0.0 ) ).xyz );
}
vec3 inverseTransformDirection( in vec3 dir, in mat4 matrix ) {
	return normalize( ( vec4( dir, 0.0 ) * matrix ).xyz );
}
mat3 transposeMat3( const in mat3 m ) {
	mat3 tmp;
	tmp[ 0 ] = vec3( m[ 0 ].x, m[ 1 ].x, m[ 2 ].x );
	tmp[ 1 ] = vec3( m[ 0 ].y, m[ 1 ].y, m[ 2 ].y );
	tmp[ 2 ] = vec3( m[ 0 ].z, m[ 1 ].z, m[ 2 ].z );
	return tmp;
}
bool isPerspectiveMatrix( mat4 m ) {
	return m[ 2 ][ 3 ] == - 1.0;
}
vec2 equirectUv( in vec3 dir ) {
	float u = atan( dir.z, dir.x ) * RECIPROCAL_PI2 + 0.5;
	float v = asin( clamp( dir.y, - 1.0, 1.0 ) ) * RECIPROCAL_PI + 0.5;
	return vec2( u, v );
}
vec3 BRDF_Lambert( const in vec3 diffuseColor ) {
	return RECIPROCAL_PI * diffuseColor;
}
vec3 F_Schlick( const in vec3 f0, const in float f90, const in float dotVH ) {
	float fresnel = exp2( ( - 5.55473 * dotVH - 6.98316 ) * dotVH );
	return f0 * ( 1.0 - fresnel ) + ( f90 * fresnel );
}
float F_Schlick( const in float f0, const in float f90, const in float dotVH ) {
	float fresnel = exp2( ( - 5.55473 * dotVH - 6.98316 ) * dotVH );
	return f0 * ( 1.0 - fresnel ) + ( f90 * fresnel );
} // validated`,bh=`#ifdef ENVMAP_TYPE_CUBE_UV
	#define cubeUV_minMipLevel 4.0
	#define cubeUV_minTileSize 16.0
	float getFace( vec3 direction ) {
		vec3 absDirection = abs( direction );
		float face = - 1.0;
		if ( absDirection.x > absDirection.z ) {
			if ( absDirection.x > absDirection.y )
				face = direction.x > 0.0 ? 0.0 : 3.0;
			else
				face = direction.y > 0.0 ? 1.0 : 4.0;
		} else {
			if ( absDirection.z > absDirection.y )
				face = direction.z > 0.0 ? 2.0 : 5.0;
			else
				face = direction.y > 0.0 ? 1.0 : 4.0;
		}
		return face;
	}
	vec2 getUV( vec3 direction, float face ) {
		vec2 uv;
		if ( face == 0.0 ) {
			uv = vec2( direction.z, direction.y ) / abs( direction.x );
		} else if ( face == 1.0 ) {
			uv = vec2( - direction.x, - direction.z ) / abs( direction.y );
		} else if ( face == 2.0 ) {
			uv = vec2( - direction.x, direction.y ) / abs( direction.z );
		} else if ( face == 3.0 ) {
			uv = vec2( - direction.z, direction.y ) / abs( direction.x );
		} else if ( face == 4.0 ) {
			uv = vec2( - direction.x, direction.z ) / abs( direction.y );
		} else {
			uv = vec2( direction.x, direction.y ) / abs( direction.z );
		}
		return 0.5 * ( uv + 1.0 );
	}
	vec3 bilinearCubeUV( sampler2D envMap, vec3 direction, float mipInt ) {
		float face = getFace( direction );
		float filterInt = max( cubeUV_minMipLevel - mipInt, 0.0 );
		mipInt = max( mipInt, cubeUV_minMipLevel );
		float faceSize = exp2( mipInt );
		highp vec2 uv = getUV( direction, face ) * ( faceSize - 2.0 ) + 1.0;
		if ( face > 2.0 ) {
			uv.y += faceSize;
			face -= 3.0;
		}
		uv.x += face * faceSize;
		uv.x += filterInt * 3.0 * cubeUV_minTileSize;
		uv.y += 4.0 * ( exp2( CUBEUV_MAX_MIP ) - faceSize );
		uv.x *= CUBEUV_TEXEL_WIDTH;
		uv.y *= CUBEUV_TEXEL_HEIGHT;
		#ifdef texture2DGradEXT
			return texture2DGradEXT( envMap, uv, vec2( 0.0 ), vec2( 0.0 ) ).rgb;
		#else
			return texture2D( envMap, uv ).rgb;
		#endif
	}
	#define cubeUV_r0 1.0
	#define cubeUV_m0 - 2.0
	#define cubeUV_r1 0.8
	#define cubeUV_m1 - 1.0
	#define cubeUV_r4 0.4
	#define cubeUV_m4 2.0
	#define cubeUV_r5 0.305
	#define cubeUV_m5 3.0
	#define cubeUV_r6 0.21
	#define cubeUV_m6 4.0
	float roughnessToMip( float roughness ) {
		float mip = 0.0;
		if ( roughness >= cubeUV_r1 ) {
			mip = ( cubeUV_r0 - roughness ) * ( cubeUV_m1 - cubeUV_m0 ) / ( cubeUV_r0 - cubeUV_r1 ) + cubeUV_m0;
		} else if ( roughness >= cubeUV_r4 ) {
			mip = ( cubeUV_r1 - roughness ) * ( cubeUV_m4 - cubeUV_m1 ) / ( cubeUV_r1 - cubeUV_r4 ) + cubeUV_m1;
		} else if ( roughness >= cubeUV_r5 ) {
			mip = ( cubeUV_r4 - roughness ) * ( cubeUV_m5 - cubeUV_m4 ) / ( cubeUV_r4 - cubeUV_r5 ) + cubeUV_m4;
		} else if ( roughness >= cubeUV_r6 ) {
			mip = ( cubeUV_r5 - roughness ) * ( cubeUV_m6 - cubeUV_m5 ) / ( cubeUV_r5 - cubeUV_r6 ) + cubeUV_m5;
		} else {
			mip = - 2.0 * log2( 1.16 * roughness );		}
		return mip;
	}
	vec4 textureCubeUV( sampler2D envMap, vec3 sampleDir, float roughness ) {
		float mip = clamp( roughnessToMip( roughness ), cubeUV_m0, CUBEUV_MAX_MIP );
		float mipF = fract( mip );
		float mipInt = floor( mip );
		vec3 color0 = bilinearCubeUV( envMap, sampleDir, mipInt );
		if ( mipF == 0.0 ) {
			return vec4( color0, 1.0 );
		} else {
			vec3 color1 = bilinearCubeUV( envMap, sampleDir, mipInt + 1.0 );
			return vec4( mix( color0, color1, mipF ), 1.0 );
		}
	}
#endif`,Th=`vec3 transformedNormal = objectNormal;
#ifdef USE_TANGENT
	vec3 transformedTangent = objectTangent;
#endif
#ifdef USE_BATCHING
	mat3 bm = mat3( batchingMatrix );
	transformedNormal /= vec3( dot( bm[ 0 ], bm[ 0 ] ), dot( bm[ 1 ], bm[ 1 ] ), dot( bm[ 2 ], bm[ 2 ] ) );
	transformedNormal = bm * transformedNormal;
	#ifdef USE_TANGENT
		transformedTangent = bm * transformedTangent;
	#endif
#endif
#ifdef USE_INSTANCING
	mat3 im = mat3( instanceMatrix );
	transformedNormal /= vec3( dot( im[ 0 ], im[ 0 ] ), dot( im[ 1 ], im[ 1 ] ), dot( im[ 2 ], im[ 2 ] ) );
	transformedNormal = im * transformedNormal;
	#ifdef USE_TANGENT
		transformedTangent = im * transformedTangent;
	#endif
#endif
transformedNormal = normalMatrix * transformedNormal;
#ifdef FLIP_SIDED
	transformedNormal = - transformedNormal;
#endif
#ifdef USE_TANGENT
	transformedTangent = ( modelViewMatrix * vec4( transformedTangent, 0.0 ) ).xyz;
	#ifdef FLIP_SIDED
		transformedTangent = - transformedTangent;
	#endif
#endif`,wh=`#ifdef USE_DISPLACEMENTMAP
	uniform sampler2D displacementMap;
	uniform float displacementScale;
	uniform float displacementBias;
#endif`,Ah=`#ifdef USE_DISPLACEMENTMAP
	transformed += normalize( objectNormal ) * ( texture2D( displacementMap, vDisplacementMapUv ).x * displacementScale + displacementBias );
#endif`,Rh=`#ifdef USE_EMISSIVEMAP
	vec4 emissiveColor = texture2D( emissiveMap, vEmissiveMapUv );
	#ifdef DECODE_VIDEO_TEXTURE_EMISSIVE
		emissiveColor = sRGBTransferEOTF( emissiveColor );
	#endif
	totalEmissiveRadiance *= emissiveColor.rgb;
#endif`,Ch=`#ifdef USE_EMISSIVEMAP
	uniform sampler2D emissiveMap;
#endif`,Lh="gl_FragColor = linearToOutputTexel( gl_FragColor );",Ph=`vec4 LinearTransferOETF( in vec4 value ) {
	return value;
}
vec4 sRGBTransferEOTF( in vec4 value ) {
	return vec4( mix( pow( value.rgb * 0.9478672986 + vec3( 0.0521327014 ), vec3( 2.4 ) ), value.rgb * 0.0773993808, vec3( lessThanEqual( value.rgb, vec3( 0.04045 ) ) ) ), value.a );
}
vec4 sRGBTransferOETF( in vec4 value ) {
	return vec4( mix( pow( value.rgb, vec3( 0.41666 ) ) * 1.055 - vec3( 0.055 ), value.rgb * 12.92, vec3( lessThanEqual( value.rgb, vec3( 0.0031308 ) ) ) ), value.a );
}`,Dh=`#ifdef USE_ENVMAP
	#ifdef ENV_WORLDPOS
		vec3 cameraToFrag;
		if ( isOrthographic ) {
			cameraToFrag = normalize( vec3( - viewMatrix[ 0 ][ 2 ], - viewMatrix[ 1 ][ 2 ], - viewMatrix[ 2 ][ 2 ] ) );
		} else {
			cameraToFrag = normalize( vWorldPosition - cameraPosition );
		}
		vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
		#ifdef ENVMAP_MODE_REFLECTION
			vec3 reflectVec = reflect( cameraToFrag, worldNormal );
		#else
			vec3 reflectVec = refract( cameraToFrag, worldNormal, refractionRatio );
		#endif
	#else
		vec3 reflectVec = vReflect;
	#endif
	#ifdef ENVMAP_TYPE_CUBE
		vec4 envColor = textureCube( envMap, envMapRotation * vec3( flipEnvMap * reflectVec.x, reflectVec.yz ) );
	#else
		vec4 envColor = vec4( 0.0 );
	#endif
	#ifdef ENVMAP_BLENDING_MULTIPLY
		outgoingLight = mix( outgoingLight, outgoingLight * envColor.xyz, specularStrength * reflectivity );
	#elif defined( ENVMAP_BLENDING_MIX )
		outgoingLight = mix( outgoingLight, envColor.xyz, specularStrength * reflectivity );
	#elif defined( ENVMAP_BLENDING_ADD )
		outgoingLight += envColor.xyz * specularStrength * reflectivity;
	#endif
#endif`,Ih=`#ifdef USE_ENVMAP
	uniform float envMapIntensity;
	uniform float flipEnvMap;
	uniform mat3 envMapRotation;
	#ifdef ENVMAP_TYPE_CUBE
		uniform samplerCube envMap;
	#else
		uniform sampler2D envMap;
	#endif
	
#endif`,Nh=`#ifdef USE_ENVMAP
	uniform float reflectivity;
	#if defined( USE_BUMPMAP ) || defined( USE_NORMALMAP ) || defined( PHONG ) || defined( LAMBERT )
		#define ENV_WORLDPOS
	#endif
	#ifdef ENV_WORLDPOS
		varying vec3 vWorldPosition;
		uniform float refractionRatio;
	#else
		varying vec3 vReflect;
	#endif
#endif`,Uh=`#ifdef USE_ENVMAP
	#if defined( USE_BUMPMAP ) || defined( USE_NORMALMAP ) || defined( PHONG ) || defined( LAMBERT )
		#define ENV_WORLDPOS
	#endif
	#ifdef ENV_WORLDPOS
		
		varying vec3 vWorldPosition;
	#else
		varying vec3 vReflect;
		uniform float refractionRatio;
	#endif
#endif`,Oh=`#ifdef USE_ENVMAP
	#ifdef ENV_WORLDPOS
		vWorldPosition = worldPosition.xyz;
	#else
		vec3 cameraToVertex;
		if ( isOrthographic ) {
			cameraToVertex = normalize( vec3( - viewMatrix[ 0 ][ 2 ], - viewMatrix[ 1 ][ 2 ], - viewMatrix[ 2 ][ 2 ] ) );
		} else {
			cameraToVertex = normalize( worldPosition.xyz - cameraPosition );
		}
		vec3 worldNormal = inverseTransformDirection( transformedNormal, viewMatrix );
		#ifdef ENVMAP_MODE_REFLECTION
			vReflect = reflect( cameraToVertex, worldNormal );
		#else
			vReflect = refract( cameraToVertex, worldNormal, refractionRatio );
		#endif
	#endif
#endif`,Fh=`#ifdef USE_FOG
	vFogDepth = - mvPosition.z;
#endif`,zh=`#ifdef USE_FOG
	varying float vFogDepth;
#endif`,Bh=`#ifdef USE_FOG
	#ifdef FOG_EXP2
		float fogFactor = 1.0 - exp( - fogDensity * fogDensity * vFogDepth * vFogDepth );
	#else
		float fogFactor = smoothstep( fogNear, fogFar, vFogDepth );
	#endif
	gl_FragColor.rgb = mix( gl_FragColor.rgb, fogColor, fogFactor );
#endif`,kh=`#ifdef USE_FOG
	uniform vec3 fogColor;
	varying float vFogDepth;
	#ifdef FOG_EXP2
		uniform float fogDensity;
	#else
		uniform float fogNear;
		uniform float fogFar;
	#endif
#endif`,Vh=`#ifdef USE_GRADIENTMAP
	uniform sampler2D gradientMap;
#endif
vec3 getGradientIrradiance( vec3 normal, vec3 lightDirection ) {
	float dotNL = dot( normal, lightDirection );
	vec2 coord = vec2( dotNL * 0.5 + 0.5, 0.0 );
	#ifdef USE_GRADIENTMAP
		return vec3( texture2D( gradientMap, coord ).r );
	#else
		vec2 fw = fwidth( coord ) * 0.5;
		return mix( vec3( 0.7 ), vec3( 1.0 ), smoothstep( 0.7 - fw.x, 0.7 + fw.x, coord.x ) );
	#endif
}`,Hh=`#ifdef USE_LIGHTMAP
	uniform sampler2D lightMap;
	uniform float lightMapIntensity;
#endif`,Gh=`LambertMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.specularStrength = specularStrength;`,Wh=`varying vec3 vViewPosition;
struct LambertMaterial {
	vec3 diffuseColor;
	float specularStrength;
};
void RE_Direct_Lambert( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in LambertMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Lambert( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in LambertMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_Lambert
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Lambert`,jh=`uniform bool receiveShadow;
uniform vec3 ambientLightColor;
#if defined( USE_LIGHT_PROBES )
	uniform vec3 lightProbe[ 9 ];
#endif
vec3 shGetIrradianceAt( in vec3 normal, in vec3 shCoefficients[ 9 ] ) {
	float x = normal.x, y = normal.y, z = normal.z;
	vec3 result = shCoefficients[ 0 ] * 0.886227;
	result += shCoefficients[ 1 ] * 2.0 * 0.511664 * y;
	result += shCoefficients[ 2 ] * 2.0 * 0.511664 * z;
	result += shCoefficients[ 3 ] * 2.0 * 0.511664 * x;
	result += shCoefficients[ 4 ] * 2.0 * 0.429043 * x * y;
	result += shCoefficients[ 5 ] * 2.0 * 0.429043 * y * z;
	result += shCoefficients[ 6 ] * ( 0.743125 * z * z - 0.247708 );
	result += shCoefficients[ 7 ] * 2.0 * 0.429043 * x * z;
	result += shCoefficients[ 8 ] * 0.429043 * ( x * x - y * y );
	return result;
}
vec3 getLightProbeIrradiance( const in vec3 lightProbe[ 9 ], const in vec3 normal ) {
	vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
	vec3 irradiance = shGetIrradianceAt( worldNormal, lightProbe );
	return irradiance;
}
vec3 getAmbientLightIrradiance( const in vec3 ambientLightColor ) {
	vec3 irradiance = ambientLightColor;
	return irradiance;
}
float getDistanceAttenuation( const in float lightDistance, const in float cutoffDistance, const in float decayExponent ) {
	float distanceFalloff = 1.0 / max( pow( lightDistance, decayExponent ), 0.01 );
	if ( cutoffDistance > 0.0 ) {
		distanceFalloff *= pow2( saturate( 1.0 - pow4( lightDistance / cutoffDistance ) ) );
	}
	return distanceFalloff;
}
float getSpotAttenuation( const in float coneCosine, const in float penumbraCosine, const in float angleCosine ) {
	return smoothstep( coneCosine, penumbraCosine, angleCosine );
}
#if NUM_DIR_LIGHTS > 0
	struct DirectionalLight {
		vec3 direction;
		vec3 color;
	};
	uniform DirectionalLight directionalLights[ NUM_DIR_LIGHTS ];
	void getDirectionalLightInfo( const in DirectionalLight directionalLight, out IncidentLight light ) {
		light.color = directionalLight.color;
		light.direction = directionalLight.direction;
		light.visible = true;
	}
#endif
#if NUM_POINT_LIGHTS > 0
	struct PointLight {
		vec3 position;
		vec3 color;
		float distance;
		float decay;
	};
	uniform PointLight pointLights[ NUM_POINT_LIGHTS ];
	void getPointLightInfo( const in PointLight pointLight, const in vec3 geometryPosition, out IncidentLight light ) {
		vec3 lVector = pointLight.position - geometryPosition;
		light.direction = normalize( lVector );
		float lightDistance = length( lVector );
		light.color = pointLight.color;
		light.color *= getDistanceAttenuation( lightDistance, pointLight.distance, pointLight.decay );
		light.visible = ( light.color != vec3( 0.0 ) );
	}
#endif
#if NUM_SPOT_LIGHTS > 0
	struct SpotLight {
		vec3 position;
		vec3 direction;
		vec3 color;
		float distance;
		float decay;
		float coneCos;
		float penumbraCos;
	};
	uniform SpotLight spotLights[ NUM_SPOT_LIGHTS ];
	void getSpotLightInfo( const in SpotLight spotLight, const in vec3 geometryPosition, out IncidentLight light ) {
		vec3 lVector = spotLight.position - geometryPosition;
		light.direction = normalize( lVector );
		float angleCos = dot( light.direction, spotLight.direction );
		float spotAttenuation = getSpotAttenuation( spotLight.coneCos, spotLight.penumbraCos, angleCos );
		if ( spotAttenuation > 0.0 ) {
			float lightDistance = length( lVector );
			light.color = spotLight.color * spotAttenuation;
			light.color *= getDistanceAttenuation( lightDistance, spotLight.distance, spotLight.decay );
			light.visible = ( light.color != vec3( 0.0 ) );
		} else {
			light.color = vec3( 0.0 );
			light.visible = false;
		}
	}
#endif
#if NUM_RECT_AREA_LIGHTS > 0
	struct RectAreaLight {
		vec3 color;
		vec3 position;
		vec3 halfWidth;
		vec3 halfHeight;
	};
	uniform sampler2D ltc_1;	uniform sampler2D ltc_2;
	uniform RectAreaLight rectAreaLights[ NUM_RECT_AREA_LIGHTS ];
#endif
#if NUM_HEMI_LIGHTS > 0
	struct HemisphereLight {
		vec3 direction;
		vec3 skyColor;
		vec3 groundColor;
	};
	uniform HemisphereLight hemisphereLights[ NUM_HEMI_LIGHTS ];
	vec3 getHemisphereLightIrradiance( const in HemisphereLight hemiLight, const in vec3 normal ) {
		float dotNL = dot( normal, hemiLight.direction );
		float hemiDiffuseWeight = 0.5 * dotNL + 0.5;
		vec3 irradiance = mix( hemiLight.groundColor, hemiLight.skyColor, hemiDiffuseWeight );
		return irradiance;
	}
#endif`,Xh=`#ifdef USE_ENVMAP
	vec3 getIBLIrradiance( const in vec3 normal ) {
		#ifdef ENVMAP_TYPE_CUBE_UV
			vec3 worldNormal = inverseTransformDirection( normal, viewMatrix );
			vec4 envMapColor = textureCubeUV( envMap, envMapRotation * worldNormal, 1.0 );
			return PI * envMapColor.rgb * envMapIntensity;
		#else
			return vec3( 0.0 );
		#endif
	}
	vec3 getIBLRadiance( const in vec3 viewDir, const in vec3 normal, const in float roughness ) {
		#ifdef ENVMAP_TYPE_CUBE_UV
			vec3 reflectVec = reflect( - viewDir, normal );
			reflectVec = normalize( mix( reflectVec, normal, roughness * roughness) );
			reflectVec = inverseTransformDirection( reflectVec, viewMatrix );
			vec4 envMapColor = textureCubeUV( envMap, envMapRotation * reflectVec, roughness );
			return envMapColor.rgb * envMapIntensity;
		#else
			return vec3( 0.0 );
		#endif
	}
	#ifdef USE_ANISOTROPY
		vec3 getIBLAnisotropyRadiance( const in vec3 viewDir, const in vec3 normal, const in float roughness, const in vec3 bitangent, const in float anisotropy ) {
			#ifdef ENVMAP_TYPE_CUBE_UV
				vec3 bentNormal = cross( bitangent, viewDir );
				bentNormal = normalize( cross( bentNormal, bitangent ) );
				bentNormal = normalize( mix( bentNormal, normal, pow2( pow2( 1.0 - anisotropy * ( 1.0 - roughness ) ) ) ) );
				return getIBLRadiance( viewDir, bentNormal, roughness );
			#else
				return vec3( 0.0 );
			#endif
		}
	#endif
#endif`,Yh=`ToonMaterial material;
material.diffuseColor = diffuseColor.rgb;`,qh=`varying vec3 vViewPosition;
struct ToonMaterial {
	vec3 diffuseColor;
};
void RE_Direct_Toon( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in ToonMaterial material, inout ReflectedLight reflectedLight ) {
	vec3 irradiance = getGradientIrradiance( geometryNormal, directLight.direction ) * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Toon( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in ToonMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_Toon
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Toon`,Kh=`BlinnPhongMaterial material;
material.diffuseColor = diffuseColor.rgb;
material.specularColor = specular;
material.specularShininess = shininess;
material.specularStrength = specularStrength;`,$h=`varying vec3 vViewPosition;
struct BlinnPhongMaterial {
	vec3 diffuseColor;
	vec3 specularColor;
	float specularShininess;
	float specularStrength;
};
void RE_Direct_BlinnPhong( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in BlinnPhongMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
	reflectedLight.directSpecular += irradiance * BRDF_BlinnPhong( directLight.direction, geometryViewDir, geometryNormal, material.specularColor, material.specularShininess ) * material.specularStrength;
}
void RE_IndirectDiffuse_BlinnPhong( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in BlinnPhongMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
#define RE_Direct				RE_Direct_BlinnPhong
#define RE_IndirectDiffuse		RE_IndirectDiffuse_BlinnPhong`,Zh=`PhysicalMaterial material;
material.diffuseColor = diffuseColor.rgb * ( 1.0 - metalnessFactor );
vec3 dxy = max( abs( dFdx( nonPerturbedNormal ) ), abs( dFdy( nonPerturbedNormal ) ) );
float geometryRoughness = max( max( dxy.x, dxy.y ), dxy.z );
material.roughness = max( roughnessFactor, 0.0525 );material.roughness += geometryRoughness;
material.roughness = min( material.roughness, 1.0 );
#ifdef IOR
	material.ior = ior;
	#ifdef USE_SPECULAR
		float specularIntensityFactor = specularIntensity;
		vec3 specularColorFactor = specularColor;
		#ifdef USE_SPECULAR_COLORMAP
			specularColorFactor *= texture2D( specularColorMap, vSpecularColorMapUv ).rgb;
		#endif
		#ifdef USE_SPECULAR_INTENSITYMAP
			specularIntensityFactor *= texture2D( specularIntensityMap, vSpecularIntensityMapUv ).a;
		#endif
		material.specularF90 = mix( specularIntensityFactor, 1.0, metalnessFactor );
	#else
		float specularIntensityFactor = 1.0;
		vec3 specularColorFactor = vec3( 1.0 );
		material.specularF90 = 1.0;
	#endif
	material.specularColor = mix( min( pow2( ( material.ior - 1.0 ) / ( material.ior + 1.0 ) ) * specularColorFactor, vec3( 1.0 ) ) * specularIntensityFactor, diffuseColor.rgb, metalnessFactor );
#else
	material.specularColor = mix( vec3( 0.04 ), diffuseColor.rgb, metalnessFactor );
	material.specularF90 = 1.0;
#endif
#ifdef USE_CLEARCOAT
	material.clearcoat = clearcoat;
	material.clearcoatRoughness = clearcoatRoughness;
	material.clearcoatF0 = vec3( 0.04 );
	material.clearcoatF90 = 1.0;
	#ifdef USE_CLEARCOATMAP
		material.clearcoat *= texture2D( clearcoatMap, vClearcoatMapUv ).x;
	#endif
	#ifdef USE_CLEARCOAT_ROUGHNESSMAP
		material.clearcoatRoughness *= texture2D( clearcoatRoughnessMap, vClearcoatRoughnessMapUv ).y;
	#endif
	material.clearcoat = saturate( material.clearcoat );	material.clearcoatRoughness = max( material.clearcoatRoughness, 0.0525 );
	material.clearcoatRoughness += geometryRoughness;
	material.clearcoatRoughness = min( material.clearcoatRoughness, 1.0 );
#endif
#ifdef USE_DISPERSION
	material.dispersion = dispersion;
#endif
#ifdef USE_IRIDESCENCE
	material.iridescence = iridescence;
	material.iridescenceIOR = iridescenceIOR;
	#ifdef USE_IRIDESCENCEMAP
		material.iridescence *= texture2D( iridescenceMap, vIridescenceMapUv ).r;
	#endif
	#ifdef USE_IRIDESCENCE_THICKNESSMAP
		material.iridescenceThickness = (iridescenceThicknessMaximum - iridescenceThicknessMinimum) * texture2D( iridescenceThicknessMap, vIridescenceThicknessMapUv ).g + iridescenceThicknessMinimum;
	#else
		material.iridescenceThickness = iridescenceThicknessMaximum;
	#endif
#endif
#ifdef USE_SHEEN
	material.sheenColor = sheenColor;
	#ifdef USE_SHEEN_COLORMAP
		material.sheenColor *= texture2D( sheenColorMap, vSheenColorMapUv ).rgb;
	#endif
	material.sheenRoughness = clamp( sheenRoughness, 0.07, 1.0 );
	#ifdef USE_SHEEN_ROUGHNESSMAP
		material.sheenRoughness *= texture2D( sheenRoughnessMap, vSheenRoughnessMapUv ).a;
	#endif
#endif
#ifdef USE_ANISOTROPY
	#ifdef USE_ANISOTROPYMAP
		mat2 anisotropyMat = mat2( anisotropyVector.x, anisotropyVector.y, - anisotropyVector.y, anisotropyVector.x );
		vec3 anisotropyPolar = texture2D( anisotropyMap, vAnisotropyMapUv ).rgb;
		vec2 anisotropyV = anisotropyMat * normalize( 2.0 * anisotropyPolar.rg - vec2( 1.0 ) ) * anisotropyPolar.b;
	#else
		vec2 anisotropyV = anisotropyVector;
	#endif
	material.anisotropy = length( anisotropyV );
	if( material.anisotropy == 0.0 ) {
		anisotropyV = vec2( 1.0, 0.0 );
	} else {
		anisotropyV /= material.anisotropy;
		material.anisotropy = saturate( material.anisotropy );
	}
	material.alphaT = mix( pow2( material.roughness ), 1.0, pow2( material.anisotropy ) );
	material.anisotropyT = tbn[ 0 ] * anisotropyV.x + tbn[ 1 ] * anisotropyV.y;
	material.anisotropyB = tbn[ 1 ] * anisotropyV.x - tbn[ 0 ] * anisotropyV.y;
#endif`,Jh=`struct PhysicalMaterial {
	vec3 diffuseColor;
	float roughness;
	vec3 specularColor;
	float specularF90;
	float dispersion;
	#ifdef USE_CLEARCOAT
		float clearcoat;
		float clearcoatRoughness;
		vec3 clearcoatF0;
		float clearcoatF90;
	#endif
	#ifdef USE_IRIDESCENCE
		float iridescence;
		float iridescenceIOR;
		float iridescenceThickness;
		vec3 iridescenceFresnel;
		vec3 iridescenceF0;
	#endif
	#ifdef USE_SHEEN
		vec3 sheenColor;
		float sheenRoughness;
	#endif
	#ifdef IOR
		float ior;
	#endif
	#ifdef USE_TRANSMISSION
		float transmission;
		float transmissionAlpha;
		float thickness;
		float attenuationDistance;
		vec3 attenuationColor;
	#endif
	#ifdef USE_ANISOTROPY
		float anisotropy;
		float alphaT;
		vec3 anisotropyT;
		vec3 anisotropyB;
	#endif
};
vec3 clearcoatSpecularDirect = vec3( 0.0 );
vec3 clearcoatSpecularIndirect = vec3( 0.0 );
vec3 sheenSpecularDirect = vec3( 0.0 );
vec3 sheenSpecularIndirect = vec3(0.0 );
vec3 Schlick_to_F0( const in vec3 f, const in float f90, const in float dotVH ) {
    float x = clamp( 1.0 - dotVH, 0.0, 1.0 );
    float x2 = x * x;
    float x5 = clamp( x * x2 * x2, 0.0, 0.9999 );
    return ( f - vec3( f90 ) * x5 ) / ( 1.0 - x5 );
}
float V_GGX_SmithCorrelated( const in float alpha, const in float dotNL, const in float dotNV ) {
	float a2 = pow2( alpha );
	float gv = dotNL * sqrt( a2 + ( 1.0 - a2 ) * pow2( dotNV ) );
	float gl = dotNV * sqrt( a2 + ( 1.0 - a2 ) * pow2( dotNL ) );
	return 0.5 / max( gv + gl, EPSILON );
}
float D_GGX( const in float alpha, const in float dotNH ) {
	float a2 = pow2( alpha );
	float denom = pow2( dotNH ) * ( a2 - 1.0 ) + 1.0;
	return RECIPROCAL_PI * a2 / pow2( denom );
}
#ifdef USE_ANISOTROPY
	float V_GGX_SmithCorrelated_Anisotropic( const in float alphaT, const in float alphaB, const in float dotTV, const in float dotBV, const in float dotTL, const in float dotBL, const in float dotNV, const in float dotNL ) {
		float gv = dotNL * length( vec3( alphaT * dotTV, alphaB * dotBV, dotNV ) );
		float gl = dotNV * length( vec3( alphaT * dotTL, alphaB * dotBL, dotNL ) );
		float v = 0.5 / ( gv + gl );
		return saturate(v);
	}
	float D_GGX_Anisotropic( const in float alphaT, const in float alphaB, const in float dotNH, const in float dotTH, const in float dotBH ) {
		float a2 = alphaT * alphaB;
		highp vec3 v = vec3( alphaB * dotTH, alphaT * dotBH, a2 * dotNH );
		highp float v2 = dot( v, v );
		float w2 = a2 / v2;
		return RECIPROCAL_PI * a2 * pow2 ( w2 );
	}
#endif
#ifdef USE_CLEARCOAT
	vec3 BRDF_GGX_Clearcoat( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material) {
		vec3 f0 = material.clearcoatF0;
		float f90 = material.clearcoatF90;
		float roughness = material.clearcoatRoughness;
		float alpha = pow2( roughness );
		vec3 halfDir = normalize( lightDir + viewDir );
		float dotNL = saturate( dot( normal, lightDir ) );
		float dotNV = saturate( dot( normal, viewDir ) );
		float dotNH = saturate( dot( normal, halfDir ) );
		float dotVH = saturate( dot( viewDir, halfDir ) );
		vec3 F = F_Schlick( f0, f90, dotVH );
		float V = V_GGX_SmithCorrelated( alpha, dotNL, dotNV );
		float D = D_GGX( alpha, dotNH );
		return F * ( V * D );
	}
#endif
vec3 BRDF_GGX( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, const in PhysicalMaterial material ) {
	vec3 f0 = material.specularColor;
	float f90 = material.specularF90;
	float roughness = material.roughness;
	float alpha = pow2( roughness );
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	float dotNH = saturate( dot( normal, halfDir ) );
	float dotVH = saturate( dot( viewDir, halfDir ) );
	vec3 F = F_Schlick( f0, f90, dotVH );
	#ifdef USE_IRIDESCENCE
		F = mix( F, material.iridescenceFresnel, material.iridescence );
	#endif
	#ifdef USE_ANISOTROPY
		float dotTL = dot( material.anisotropyT, lightDir );
		float dotTV = dot( material.anisotropyT, viewDir );
		float dotTH = dot( material.anisotropyT, halfDir );
		float dotBL = dot( material.anisotropyB, lightDir );
		float dotBV = dot( material.anisotropyB, viewDir );
		float dotBH = dot( material.anisotropyB, halfDir );
		float V = V_GGX_SmithCorrelated_Anisotropic( material.alphaT, alpha, dotTV, dotBV, dotTL, dotBL, dotNV, dotNL );
		float D = D_GGX_Anisotropic( material.alphaT, alpha, dotNH, dotTH, dotBH );
	#else
		float V = V_GGX_SmithCorrelated( alpha, dotNL, dotNV );
		float D = D_GGX( alpha, dotNH );
	#endif
	return F * ( V * D );
}
vec2 LTC_Uv( const in vec3 N, const in vec3 V, const in float roughness ) {
	const float LUT_SIZE = 64.0;
	const float LUT_SCALE = ( LUT_SIZE - 1.0 ) / LUT_SIZE;
	const float LUT_BIAS = 0.5 / LUT_SIZE;
	float dotNV = saturate( dot( N, V ) );
	vec2 uv = vec2( roughness, sqrt( 1.0 - dotNV ) );
	uv = uv * LUT_SCALE + LUT_BIAS;
	return uv;
}
float LTC_ClippedSphereFormFactor( const in vec3 f ) {
	float l = length( f );
	return max( ( l * l + f.z ) / ( l + 1.0 ), 0.0 );
}
vec3 LTC_EdgeVectorFormFactor( const in vec3 v1, const in vec3 v2 ) {
	float x = dot( v1, v2 );
	float y = abs( x );
	float a = 0.8543985 + ( 0.4965155 + 0.0145206 * y ) * y;
	float b = 3.4175940 + ( 4.1616724 + y ) * y;
	float v = a / b;
	float theta_sintheta = ( x > 0.0 ) ? v : 0.5 * inversesqrt( max( 1.0 - x * x, 1e-7 ) ) - v;
	return cross( v1, v2 ) * theta_sintheta;
}
vec3 LTC_Evaluate( const in vec3 N, const in vec3 V, const in vec3 P, const in mat3 mInv, const in vec3 rectCoords[ 4 ] ) {
	vec3 v1 = rectCoords[ 1 ] - rectCoords[ 0 ];
	vec3 v2 = rectCoords[ 3 ] - rectCoords[ 0 ];
	vec3 lightNormal = cross( v1, v2 );
	if( dot( lightNormal, P - rectCoords[ 0 ] ) < 0.0 ) return vec3( 0.0 );
	vec3 T1, T2;
	T1 = normalize( V - N * dot( V, N ) );
	T2 = - cross( N, T1 );
	mat3 mat = mInv * transposeMat3( mat3( T1, T2, N ) );
	vec3 coords[ 4 ];
	coords[ 0 ] = mat * ( rectCoords[ 0 ] - P );
	coords[ 1 ] = mat * ( rectCoords[ 1 ] - P );
	coords[ 2 ] = mat * ( rectCoords[ 2 ] - P );
	coords[ 3 ] = mat * ( rectCoords[ 3 ] - P );
	coords[ 0 ] = normalize( coords[ 0 ] );
	coords[ 1 ] = normalize( coords[ 1 ] );
	coords[ 2 ] = normalize( coords[ 2 ] );
	coords[ 3 ] = normalize( coords[ 3 ] );
	vec3 vectorFormFactor = vec3( 0.0 );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 0 ], coords[ 1 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 1 ], coords[ 2 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 2 ], coords[ 3 ] );
	vectorFormFactor += LTC_EdgeVectorFormFactor( coords[ 3 ], coords[ 0 ] );
	float result = LTC_ClippedSphereFormFactor( vectorFormFactor );
	return vec3( result );
}
#if defined( USE_SHEEN )
float D_Charlie( float roughness, float dotNH ) {
	float alpha = pow2( roughness );
	float invAlpha = 1.0 / alpha;
	float cos2h = dotNH * dotNH;
	float sin2h = max( 1.0 - cos2h, 0.0078125 );
	return ( 2.0 + invAlpha ) * pow( sin2h, invAlpha * 0.5 ) / ( 2.0 * PI );
}
float V_Neubelt( float dotNV, float dotNL ) {
	return saturate( 1.0 / ( 4.0 * ( dotNL + dotNV - dotNL * dotNV ) ) );
}
vec3 BRDF_Sheen( const in vec3 lightDir, const in vec3 viewDir, const in vec3 normal, vec3 sheenColor, const in float sheenRoughness ) {
	vec3 halfDir = normalize( lightDir + viewDir );
	float dotNL = saturate( dot( normal, lightDir ) );
	float dotNV = saturate( dot( normal, viewDir ) );
	float dotNH = saturate( dot( normal, halfDir ) );
	float D = D_Charlie( sheenRoughness, dotNH );
	float V = V_Neubelt( dotNV, dotNL );
	return sheenColor * ( D * V );
}
#endif
float IBLSheenBRDF( const in vec3 normal, const in vec3 viewDir, const in float roughness ) {
	float dotNV = saturate( dot( normal, viewDir ) );
	float r2 = roughness * roughness;
	float a = roughness < 0.25 ? -339.2 * r2 + 161.4 * roughness - 25.9 : -8.48 * r2 + 14.3 * roughness - 9.95;
	float b = roughness < 0.25 ? 44.0 * r2 - 23.7 * roughness + 3.26 : 1.97 * r2 - 3.27 * roughness + 0.72;
	float DG = exp( a * dotNV + b ) + ( roughness < 0.25 ? 0.0 : 0.1 * ( roughness - 0.25 ) );
	return saturate( DG * RECIPROCAL_PI );
}
vec2 DFGApprox( const in vec3 normal, const in vec3 viewDir, const in float roughness ) {
	float dotNV = saturate( dot( normal, viewDir ) );
	const vec4 c0 = vec4( - 1, - 0.0275, - 0.572, 0.022 );
	const vec4 c1 = vec4( 1, 0.0425, 1.04, - 0.04 );
	vec4 r = roughness * c0 + c1;
	float a004 = min( r.x * r.x, exp2( - 9.28 * dotNV ) ) * r.x + r.y;
	vec2 fab = vec2( - 1.04, 1.04 ) * a004 + r.zw;
	return fab;
}
vec3 EnvironmentBRDF( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float roughness ) {
	vec2 fab = DFGApprox( normal, viewDir, roughness );
	return specularColor * fab.x + specularF90 * fab.y;
}
#ifdef USE_IRIDESCENCE
void computeMultiscatteringIridescence( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float iridescence, const in vec3 iridescenceF0, const in float roughness, inout vec3 singleScatter, inout vec3 multiScatter ) {
#else
void computeMultiscattering( const in vec3 normal, const in vec3 viewDir, const in vec3 specularColor, const in float specularF90, const in float roughness, inout vec3 singleScatter, inout vec3 multiScatter ) {
#endif
	vec2 fab = DFGApprox( normal, viewDir, roughness );
	#ifdef USE_IRIDESCENCE
		vec3 Fr = mix( specularColor, iridescenceF0, iridescence );
	#else
		vec3 Fr = specularColor;
	#endif
	vec3 FssEss = Fr * fab.x + specularF90 * fab.y;
	float Ess = fab.x + fab.y;
	float Ems = 1.0 - Ess;
	vec3 Favg = Fr + ( 1.0 - Fr ) * 0.047619;	vec3 Fms = FssEss * Favg / ( 1.0 - Ems * Favg );
	singleScatter += FssEss;
	multiScatter += Fms * Ems;
}
#if NUM_RECT_AREA_LIGHTS > 0
	void RE_Direct_RectArea_Physical( const in RectAreaLight rectAreaLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
		vec3 normal = geometryNormal;
		vec3 viewDir = geometryViewDir;
		vec3 position = geometryPosition;
		vec3 lightPos = rectAreaLight.position;
		vec3 halfWidth = rectAreaLight.halfWidth;
		vec3 halfHeight = rectAreaLight.halfHeight;
		vec3 lightColor = rectAreaLight.color;
		float roughness = material.roughness;
		vec3 rectCoords[ 4 ];
		rectCoords[ 0 ] = lightPos + halfWidth - halfHeight;		rectCoords[ 1 ] = lightPos - halfWidth - halfHeight;
		rectCoords[ 2 ] = lightPos - halfWidth + halfHeight;
		rectCoords[ 3 ] = lightPos + halfWidth + halfHeight;
		vec2 uv = LTC_Uv( normal, viewDir, roughness );
		vec4 t1 = texture2D( ltc_1, uv );
		vec4 t2 = texture2D( ltc_2, uv );
		mat3 mInv = mat3(
			vec3( t1.x, 0, t1.y ),
			vec3(    0, 1,    0 ),
			vec3( t1.z, 0, t1.w )
		);
		vec3 fresnel = ( material.specularColor * t2.x + ( vec3( 1.0 ) - material.specularColor ) * t2.y );
		reflectedLight.directSpecular += lightColor * fresnel * LTC_Evaluate( normal, viewDir, position, mInv, rectCoords );
		reflectedLight.directDiffuse += lightColor * material.diffuseColor * LTC_Evaluate( normal, viewDir, position, mat3( 1.0 ), rectCoords );
	}
#endif
void RE_Direct_Physical( const in IncidentLight directLight, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
	float dotNL = saturate( dot( geometryNormal, directLight.direction ) );
	vec3 irradiance = dotNL * directLight.color;
	#ifdef USE_CLEARCOAT
		float dotNLcc = saturate( dot( geometryClearcoatNormal, directLight.direction ) );
		vec3 ccIrradiance = dotNLcc * directLight.color;
		clearcoatSpecularDirect += ccIrradiance * BRDF_GGX_Clearcoat( directLight.direction, geometryViewDir, geometryClearcoatNormal, material );
	#endif
	#ifdef USE_SHEEN
		sheenSpecularDirect += irradiance * BRDF_Sheen( directLight.direction, geometryViewDir, geometryNormal, material.sheenColor, material.sheenRoughness );
	#endif
	reflectedLight.directSpecular += irradiance * BRDF_GGX( directLight.direction, geometryViewDir, geometryNormal, material );
	reflectedLight.directDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectDiffuse_Physical( const in vec3 irradiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight ) {
	reflectedLight.indirectDiffuse += irradiance * BRDF_Lambert( material.diffuseColor );
}
void RE_IndirectSpecular_Physical( const in vec3 radiance, const in vec3 irradiance, const in vec3 clearcoatRadiance, const in vec3 geometryPosition, const in vec3 geometryNormal, const in vec3 geometryViewDir, const in vec3 geometryClearcoatNormal, const in PhysicalMaterial material, inout ReflectedLight reflectedLight) {
	#ifdef USE_CLEARCOAT
		clearcoatSpecularIndirect += clearcoatRadiance * EnvironmentBRDF( geometryClearcoatNormal, geometryViewDir, material.clearcoatF0, material.clearcoatF90, material.clearcoatRoughness );
	#endif
	#ifdef USE_SHEEN
		sheenSpecularIndirect += irradiance * material.sheenColor * IBLSheenBRDF( geometryNormal, geometryViewDir, material.sheenRoughness );
	#endif
	vec3 singleScattering = vec3( 0.0 );
	vec3 multiScattering = vec3( 0.0 );
	vec3 cosineWeightedIrradiance = irradiance * RECIPROCAL_PI;
	#ifdef USE_IRIDESCENCE
		computeMultiscatteringIridescence( geometryNormal, geometryViewDir, material.specularColor, material.specularF90, material.iridescence, material.iridescenceFresnel, material.roughness, singleScattering, multiScattering );
	#else
		computeMultiscattering( geometryNormal, geometryViewDir, material.specularColor, material.specularF90, material.roughness, singleScattering, multiScattering );
	#endif
	vec3 totalScattering = singleScattering + multiScattering;
	vec3 diffuse = material.diffuseColor * ( 1.0 - max( max( totalScattering.r, totalScattering.g ), totalScattering.b ) );
	reflectedLight.indirectSpecular += radiance * singleScattering;
	reflectedLight.indirectSpecular += multiScattering * cosineWeightedIrradiance;
	reflectedLight.indirectDiffuse += diffuse * cosineWeightedIrradiance;
}
#define RE_Direct				RE_Direct_Physical
#define RE_Direct_RectArea		RE_Direct_RectArea_Physical
#define RE_IndirectDiffuse		RE_IndirectDiffuse_Physical
#define RE_IndirectSpecular		RE_IndirectSpecular_Physical
float computeSpecularOcclusion( const in float dotNV, const in float ambientOcclusion, const in float roughness ) {
	return saturate( pow( dotNV + ambientOcclusion, exp2( - 16.0 * roughness - 1.0 ) ) - 1.0 + ambientOcclusion );
}`,Qh=`
vec3 geometryPosition = - vViewPosition;
vec3 geometryNormal = normal;
vec3 geometryViewDir = ( isOrthographic ) ? vec3( 0, 0, 1 ) : normalize( vViewPosition );
vec3 geometryClearcoatNormal = vec3( 0.0 );
#ifdef USE_CLEARCOAT
	geometryClearcoatNormal = clearcoatNormal;
#endif
#ifdef USE_IRIDESCENCE
	float dotNVi = saturate( dot( normal, geometryViewDir ) );
	if ( material.iridescenceThickness == 0.0 ) {
		material.iridescence = 0.0;
	} else {
		material.iridescence = saturate( material.iridescence );
	}
	if ( material.iridescence > 0.0 ) {
		material.iridescenceFresnel = evalIridescence( 1.0, material.iridescenceIOR, dotNVi, material.iridescenceThickness, material.specularColor );
		material.iridescenceF0 = Schlick_to_F0( material.iridescenceFresnel, 1.0, dotNVi );
	}
#endif
IncidentLight directLight;
#if ( NUM_POINT_LIGHTS > 0 ) && defined( RE_Direct )
	PointLight pointLight;
	#if defined( USE_SHADOWMAP ) && NUM_POINT_LIGHT_SHADOWS > 0
	PointLightShadow pointLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_POINT_LIGHTS; i ++ ) {
		pointLight = pointLights[ i ];
		getPointLightInfo( pointLight, geometryPosition, directLight );
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_POINT_LIGHT_SHADOWS )
		pointLightShadow = pointLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getPointShadow( pointShadowMap[ i ], pointLightShadow.shadowMapSize, pointLightShadow.shadowIntensity, pointLightShadow.shadowBias, pointLightShadow.shadowRadius, vPointShadowCoord[ i ], pointLightShadow.shadowCameraNear, pointLightShadow.shadowCameraFar ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_SPOT_LIGHTS > 0 ) && defined( RE_Direct )
	SpotLight spotLight;
	vec4 spotColor;
	vec3 spotLightCoord;
	bool inSpotLightMap;
	#if defined( USE_SHADOWMAP ) && NUM_SPOT_LIGHT_SHADOWS > 0
	SpotLightShadow spotLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHTS; i ++ ) {
		spotLight = spotLights[ i ];
		getSpotLightInfo( spotLight, geometryPosition, directLight );
		#if ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS )
		#define SPOT_LIGHT_MAP_INDEX UNROLLED_LOOP_INDEX
		#elif ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
		#define SPOT_LIGHT_MAP_INDEX NUM_SPOT_LIGHT_MAPS
		#else
		#define SPOT_LIGHT_MAP_INDEX ( UNROLLED_LOOP_INDEX - NUM_SPOT_LIGHT_SHADOWS + NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS )
		#endif
		#if ( SPOT_LIGHT_MAP_INDEX < NUM_SPOT_LIGHT_MAPS )
			spotLightCoord = vSpotLightCoord[ i ].xyz / vSpotLightCoord[ i ].w;
			inSpotLightMap = all( lessThan( abs( spotLightCoord * 2. - 1. ), vec3( 1.0 ) ) );
			spotColor = texture2D( spotLightMap[ SPOT_LIGHT_MAP_INDEX ], spotLightCoord.xy );
			directLight.color = inSpotLightMap ? directLight.color * spotColor.rgb : directLight.color;
		#endif
		#undef SPOT_LIGHT_MAP_INDEX
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
		spotLightShadow = spotLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getShadow( spotShadowMap[ i ], spotLightShadow.shadowMapSize, spotLightShadow.shadowIntensity, spotLightShadow.shadowBias, spotLightShadow.shadowRadius, vSpotLightCoord[ i ] ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_DIR_LIGHTS > 0 ) && defined( RE_Direct )
	DirectionalLight directionalLight;
	#if defined( USE_SHADOWMAP ) && NUM_DIR_LIGHT_SHADOWS > 0
	DirectionalLightShadow directionalLightShadow;
	#endif
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_DIR_LIGHTS; i ++ ) {
		directionalLight = directionalLights[ i ];
		getDirectionalLightInfo( directionalLight, directLight );
		#if defined( USE_SHADOWMAP ) && ( UNROLLED_LOOP_INDEX < NUM_DIR_LIGHT_SHADOWS )
		directionalLightShadow = directionalLightShadows[ i ];
		directLight.color *= ( directLight.visible && receiveShadow ) ? getShadow( directionalShadowMap[ i ], directionalLightShadow.shadowMapSize, directionalLightShadow.shadowIntensity, directionalLightShadow.shadowBias, directionalLightShadow.shadowRadius, vDirectionalShadowCoord[ i ] ) : 1.0;
		#endif
		RE_Direct( directLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if ( NUM_RECT_AREA_LIGHTS > 0 ) && defined( RE_Direct_RectArea )
	RectAreaLight rectAreaLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_RECT_AREA_LIGHTS; i ++ ) {
		rectAreaLight = rectAreaLights[ i ];
		RE_Direct_RectArea( rectAreaLight, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
	}
	#pragma unroll_loop_end
#endif
#if defined( RE_IndirectDiffuse )
	vec3 iblIrradiance = vec3( 0.0 );
	vec3 irradiance = getAmbientLightIrradiance( ambientLightColor );
	#if defined( USE_LIGHT_PROBES )
		irradiance += getLightProbeIrradiance( lightProbe, geometryNormal );
	#endif
	#if ( NUM_HEMI_LIGHTS > 0 )
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_HEMI_LIGHTS; i ++ ) {
			irradiance += getHemisphereLightIrradiance( hemisphereLights[ i ], geometryNormal );
		}
		#pragma unroll_loop_end
	#endif
#endif
#if defined( RE_IndirectSpecular )
	vec3 radiance = vec3( 0.0 );
	vec3 clearcoatRadiance = vec3( 0.0 );
#endif`,eu=`#if defined( RE_IndirectDiffuse )
	#ifdef USE_LIGHTMAP
		vec4 lightMapTexel = texture2D( lightMap, vLightMapUv );
		vec3 lightMapIrradiance = lightMapTexel.rgb * lightMapIntensity;
		irradiance += lightMapIrradiance;
	#endif
	#if defined( USE_ENVMAP ) && defined( STANDARD ) && defined( ENVMAP_TYPE_CUBE_UV )
		iblIrradiance += getIBLIrradiance( geometryNormal );
	#endif
#endif
#if defined( USE_ENVMAP ) && defined( RE_IndirectSpecular )
	#ifdef USE_ANISOTROPY
		radiance += getIBLAnisotropyRadiance( geometryViewDir, geometryNormal, material.roughness, material.anisotropyB, material.anisotropy );
	#else
		radiance += getIBLRadiance( geometryViewDir, geometryNormal, material.roughness );
	#endif
	#ifdef USE_CLEARCOAT
		clearcoatRadiance += getIBLRadiance( geometryViewDir, geometryClearcoatNormal, material.clearcoatRoughness );
	#endif
#endif`,tu=`#if defined( RE_IndirectDiffuse )
	RE_IndirectDiffuse( irradiance, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
#endif
#if defined( RE_IndirectSpecular )
	RE_IndirectSpecular( radiance, iblIrradiance, clearcoatRadiance, geometryPosition, geometryNormal, geometryViewDir, geometryClearcoatNormal, material, reflectedLight );
#endif`,ru=`#if defined( USE_LOGDEPTHBUF )
	gl_FragDepth = vIsPerspective == 0.0 ? gl_FragCoord.z : log2( vFragDepth ) * logDepthBufFC * 0.5;
#endif`,iu=`#if defined( USE_LOGDEPTHBUF )
	uniform float logDepthBufFC;
	varying float vFragDepth;
	varying float vIsPerspective;
#endif`,nu=`#ifdef USE_LOGDEPTHBUF
	varying float vFragDepth;
	varying float vIsPerspective;
#endif`,au=`#ifdef USE_LOGDEPTHBUF
	vFragDepth = 1.0 + gl_Position.w;
	vIsPerspective = float( isPerspectiveMatrix( projectionMatrix ) );
#endif`,su=`#ifdef USE_MAP
	vec4 sampledDiffuseColor = texture2D( map, vMapUv );
	#ifdef DECODE_VIDEO_TEXTURE
		sampledDiffuseColor = sRGBTransferEOTF( sampledDiffuseColor );
	#endif
	diffuseColor *= sampledDiffuseColor;
#endif`,ou=`#ifdef USE_MAP
	uniform sampler2D map;
#endif`,lu=`#if defined( USE_MAP ) || defined( USE_ALPHAMAP )
	#if defined( USE_POINTS_UV )
		vec2 uv = vUv;
	#else
		vec2 uv = ( uvTransform * vec3( gl_PointCoord.x, 1.0 - gl_PointCoord.y, 1 ) ).xy;
	#endif
#endif
#ifdef USE_MAP
	diffuseColor *= texture2D( map, uv );
#endif
#ifdef USE_ALPHAMAP
	diffuseColor.a *= texture2D( alphaMap, uv ).g;
#endif`,cu=`#if defined( USE_POINTS_UV )
	varying vec2 vUv;
#else
	#if defined( USE_MAP ) || defined( USE_ALPHAMAP )
		uniform mat3 uvTransform;
	#endif
#endif
#ifdef USE_MAP
	uniform sampler2D map;
#endif
#ifdef USE_ALPHAMAP
	uniform sampler2D alphaMap;
#endif`,hu=`float metalnessFactor = metalness;
#ifdef USE_METALNESSMAP
	vec4 texelMetalness = texture2D( metalnessMap, vMetalnessMapUv );
	metalnessFactor *= texelMetalness.b;
#endif`,uu=`#ifdef USE_METALNESSMAP
	uniform sampler2D metalnessMap;
#endif`,du=`#ifdef USE_INSTANCING_MORPH
	float morphTargetInfluences[ MORPHTARGETS_COUNT ];
	float morphTargetBaseInfluence = texelFetch( morphTexture, ivec2( 0, gl_InstanceID ), 0 ).r;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		morphTargetInfluences[i] =  texelFetch( morphTexture, ivec2( i + 1, gl_InstanceID ), 0 ).r;
	}
#endif`,pu=`#if defined( USE_MORPHCOLORS )
	vColor *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		#if defined( USE_COLOR_ALPHA )
			if ( morphTargetInfluences[ i ] != 0.0 ) vColor += getMorph( gl_VertexID, i, 2 ) * morphTargetInfluences[ i ];
		#elif defined( USE_COLOR )
			if ( morphTargetInfluences[ i ] != 0.0 ) vColor += getMorph( gl_VertexID, i, 2 ).rgb * morphTargetInfluences[ i ];
		#endif
	}
#endif`,fu=`#ifdef USE_MORPHNORMALS
	objectNormal *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		if ( morphTargetInfluences[ i ] != 0.0 ) objectNormal += getMorph( gl_VertexID, i, 1 ).xyz * morphTargetInfluences[ i ];
	}
#endif`,mu=`#ifdef USE_MORPHTARGETS
	#ifndef USE_INSTANCING_MORPH
		uniform float morphTargetBaseInfluence;
		uniform float morphTargetInfluences[ MORPHTARGETS_COUNT ];
	#endif
	uniform sampler2DArray morphTargetsTexture;
	uniform ivec2 morphTargetsTextureSize;
	vec4 getMorph( const in int vertexIndex, const in int morphTargetIndex, const in int offset ) {
		int texelIndex = vertexIndex * MORPHTARGETS_TEXTURE_STRIDE + offset;
		int y = texelIndex / morphTargetsTextureSize.x;
		int x = texelIndex - y * morphTargetsTextureSize.x;
		ivec3 morphUV = ivec3( x, y, morphTargetIndex );
		return texelFetch( morphTargetsTexture, morphUV, 0 );
	}
#endif`,gu=`#ifdef USE_MORPHTARGETS
	transformed *= morphTargetBaseInfluence;
	for ( int i = 0; i < MORPHTARGETS_COUNT; i ++ ) {
		if ( morphTargetInfluences[ i ] != 0.0 ) transformed += getMorph( gl_VertexID, i, 0 ).xyz * morphTargetInfluences[ i ];
	}
#endif`,_u=`float faceDirection = gl_FrontFacing ? 1.0 : - 1.0;
#ifdef FLAT_SHADED
	vec3 fdx = dFdx( vViewPosition );
	vec3 fdy = dFdy( vViewPosition );
	vec3 normal = normalize( cross( fdx, fdy ) );
#else
	vec3 normal = normalize( vNormal );
	#ifdef DOUBLE_SIDED
		normal *= faceDirection;
	#endif
#endif
#if defined( USE_NORMALMAP_TANGENTSPACE ) || defined( USE_CLEARCOAT_NORMALMAP ) || defined( USE_ANISOTROPY )
	#ifdef USE_TANGENT
		mat3 tbn = mat3( normalize( vTangent ), normalize( vBitangent ), normal );
	#else
		mat3 tbn = getTangentFrame( - vViewPosition, normal,
		#if defined( USE_NORMALMAP )
			vNormalMapUv
		#elif defined( USE_CLEARCOAT_NORMALMAP )
			vClearcoatNormalMapUv
		#else
			vUv
		#endif
		);
	#endif
	#if defined( DOUBLE_SIDED ) && ! defined( FLAT_SHADED )
		tbn[0] *= faceDirection;
		tbn[1] *= faceDirection;
	#endif
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	#ifdef USE_TANGENT
		mat3 tbn2 = mat3( normalize( vTangent ), normalize( vBitangent ), normal );
	#else
		mat3 tbn2 = getTangentFrame( - vViewPosition, normal, vClearcoatNormalMapUv );
	#endif
	#if defined( DOUBLE_SIDED ) && ! defined( FLAT_SHADED )
		tbn2[0] *= faceDirection;
		tbn2[1] *= faceDirection;
	#endif
#endif
vec3 nonPerturbedNormal = normal;`,vu=`#ifdef USE_NORMALMAP_OBJECTSPACE
	normal = texture2D( normalMap, vNormalMapUv ).xyz * 2.0 - 1.0;
	#ifdef FLIP_SIDED
		normal = - normal;
	#endif
	#ifdef DOUBLE_SIDED
		normal = normal * faceDirection;
	#endif
	normal = normalize( normalMatrix * normal );
#elif defined( USE_NORMALMAP_TANGENTSPACE )
	vec3 mapN = texture2D( normalMap, vNormalMapUv ).xyz * 2.0 - 1.0;
	mapN.xy *= normalScale;
	normal = normalize( tbn * mapN );
#elif defined( USE_BUMPMAP )
	normal = perturbNormalArb( - vViewPosition, normal, dHdxy_fwd(), faceDirection );
#endif`,xu=`#ifndef FLAT_SHADED
	varying vec3 vNormal;
	#ifdef USE_TANGENT
		varying vec3 vTangent;
		varying vec3 vBitangent;
	#endif
#endif`,yu=`#ifndef FLAT_SHADED
	varying vec3 vNormal;
	#ifdef USE_TANGENT
		varying vec3 vTangent;
		varying vec3 vBitangent;
	#endif
#endif`,Mu=`#ifndef FLAT_SHADED
	vNormal = normalize( transformedNormal );
	#ifdef USE_TANGENT
		vTangent = normalize( transformedTangent );
		vBitangent = normalize( cross( vNormal, vTangent ) * tangent.w );
	#endif
#endif`,Su=`#ifdef USE_NORMALMAP
	uniform sampler2D normalMap;
	uniform vec2 normalScale;
#endif
#ifdef USE_NORMALMAP_OBJECTSPACE
	uniform mat3 normalMatrix;
#endif
#if ! defined ( USE_TANGENT ) && ( defined ( USE_NORMALMAP_TANGENTSPACE ) || defined ( USE_CLEARCOAT_NORMALMAP ) || defined( USE_ANISOTROPY ) )
	mat3 getTangentFrame( vec3 eye_pos, vec3 surf_norm, vec2 uv ) {
		vec3 q0 = dFdx( eye_pos.xyz );
		vec3 q1 = dFdy( eye_pos.xyz );
		vec2 st0 = dFdx( uv.st );
		vec2 st1 = dFdy( uv.st );
		vec3 N = surf_norm;
		vec3 q1perp = cross( q1, N );
		vec3 q0perp = cross( N, q0 );
		vec3 T = q1perp * st0.x + q0perp * st1.x;
		vec3 B = q1perp * st0.y + q0perp * st1.y;
		float det = max( dot( T, T ), dot( B, B ) );
		float scale = ( det == 0.0 ) ? 0.0 : inversesqrt( det );
		return mat3( T * scale, B * scale, N );
	}
#endif`,Eu=`#ifdef USE_CLEARCOAT
	vec3 clearcoatNormal = nonPerturbedNormal;
#endif`,bu=`#ifdef USE_CLEARCOAT_NORMALMAP
	vec3 clearcoatMapN = texture2D( clearcoatNormalMap, vClearcoatNormalMapUv ).xyz * 2.0 - 1.0;
	clearcoatMapN.xy *= clearcoatNormalScale;
	clearcoatNormal = normalize( tbn2 * clearcoatMapN );
#endif`,Tu=`#ifdef USE_CLEARCOATMAP
	uniform sampler2D clearcoatMap;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	uniform sampler2D clearcoatNormalMap;
	uniform vec2 clearcoatNormalScale;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	uniform sampler2D clearcoatRoughnessMap;
#endif`,wu=`#ifdef USE_IRIDESCENCEMAP
	uniform sampler2D iridescenceMap;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	uniform sampler2D iridescenceThicknessMap;
#endif`,Au=`#ifdef OPAQUE
diffuseColor.a = 1.0;
#endif
#ifdef USE_TRANSMISSION
diffuseColor.a *= material.transmissionAlpha;
#endif
gl_FragColor = vec4( outgoingLight, diffuseColor.a );`,Ru=`vec3 packNormalToRGB( const in vec3 normal ) {
	return normalize( normal ) * 0.5 + 0.5;
}
vec3 unpackRGBToNormal( const in vec3 rgb ) {
	return 2.0 * rgb.xyz - 1.0;
}
const float PackUpscale = 256. / 255.;const float UnpackDownscale = 255. / 256.;const float ShiftRight8 = 1. / 256.;
const float Inv255 = 1. / 255.;
const vec4 PackFactors = vec4( 1.0, 256.0, 256.0 * 256.0, 256.0 * 256.0 * 256.0 );
const vec2 UnpackFactors2 = vec2( UnpackDownscale, 1.0 / PackFactors.g );
const vec3 UnpackFactors3 = vec3( UnpackDownscale / PackFactors.rg, 1.0 / PackFactors.b );
const vec4 UnpackFactors4 = vec4( UnpackDownscale / PackFactors.rgb, 1.0 / PackFactors.a );
vec4 packDepthToRGBA( const in float v ) {
	if( v <= 0.0 )
		return vec4( 0., 0., 0., 0. );
	if( v >= 1.0 )
		return vec4( 1., 1., 1., 1. );
	float vuf;
	float af = modf( v * PackFactors.a, vuf );
	float bf = modf( vuf * ShiftRight8, vuf );
	float gf = modf( vuf * ShiftRight8, vuf );
	return vec4( vuf * Inv255, gf * PackUpscale, bf * PackUpscale, af );
}
vec3 packDepthToRGB( const in float v ) {
	if( v <= 0.0 )
		return vec3( 0., 0., 0. );
	if( v >= 1.0 )
		return vec3( 1., 1., 1. );
	float vuf;
	float bf = modf( v * PackFactors.b, vuf );
	float gf = modf( vuf * ShiftRight8, vuf );
	return vec3( vuf * Inv255, gf * PackUpscale, bf );
}
vec2 packDepthToRG( const in float v ) {
	if( v <= 0.0 )
		return vec2( 0., 0. );
	if( v >= 1.0 )
		return vec2( 1., 1. );
	float vuf;
	float gf = modf( v * 256., vuf );
	return vec2( vuf * Inv255, gf );
}
float unpackRGBAToDepth( const in vec4 v ) {
	return dot( v, UnpackFactors4 );
}
float unpackRGBToDepth( const in vec3 v ) {
	return dot( v, UnpackFactors3 );
}
float unpackRGToDepth( const in vec2 v ) {
	return v.r * UnpackFactors2.r + v.g * UnpackFactors2.g;
}
vec4 pack2HalfToRGBA( const in vec2 v ) {
	vec4 r = vec4( v.x, fract( v.x * 255.0 ), v.y, fract( v.y * 255.0 ) );
	return vec4( r.x - r.y / 255.0, r.y, r.z - r.w / 255.0, r.w );
}
vec2 unpackRGBATo2Half( const in vec4 v ) {
	return vec2( v.x + ( v.y / 255.0 ), v.z + ( v.w / 255.0 ) );
}
float viewZToOrthographicDepth( const in float viewZ, const in float near, const in float far ) {
	return ( viewZ + near ) / ( near - far );
}
float orthographicDepthToViewZ( const in float depth, const in float near, const in float far ) {
	return depth * ( near - far ) - near;
}
float viewZToPerspectiveDepth( const in float viewZ, const in float near, const in float far ) {
	return ( ( near + viewZ ) * far ) / ( ( far - near ) * viewZ );
}
float perspectiveDepthToViewZ( const in float depth, const in float near, const in float far ) {
	return ( near * far ) / ( ( far - near ) * depth - far );
}`,Cu=`#ifdef PREMULTIPLIED_ALPHA
	gl_FragColor.rgb *= gl_FragColor.a;
#endif`,Lu=`vec4 mvPosition = vec4( transformed, 1.0 );
#ifdef USE_BATCHING
	mvPosition = batchingMatrix * mvPosition;
#endif
#ifdef USE_INSTANCING
	mvPosition = instanceMatrix * mvPosition;
#endif
mvPosition = modelViewMatrix * mvPosition;
gl_Position = projectionMatrix * mvPosition;`,Pu=`#ifdef DITHERING
	gl_FragColor.rgb = dithering( gl_FragColor.rgb );
#endif`,Du=`#ifdef DITHERING
	vec3 dithering( vec3 color ) {
		float grid_position = rand( gl_FragCoord.xy );
		vec3 dither_shift_RGB = vec3( 0.25 / 255.0, -0.25 / 255.0, 0.25 / 255.0 );
		dither_shift_RGB = mix( 2.0 * dither_shift_RGB, -2.0 * dither_shift_RGB, grid_position );
		return color + dither_shift_RGB;
	}
#endif`,Iu=`float roughnessFactor = roughness;
#ifdef USE_ROUGHNESSMAP
	vec4 texelRoughness = texture2D( roughnessMap, vRoughnessMapUv );
	roughnessFactor *= texelRoughness.g;
#endif`,Nu=`#ifdef USE_ROUGHNESSMAP
	uniform sampler2D roughnessMap;
#endif`,Uu=`#if NUM_SPOT_LIGHT_COORDS > 0
	varying vec4 vSpotLightCoord[ NUM_SPOT_LIGHT_COORDS ];
#endif
#if NUM_SPOT_LIGHT_MAPS > 0
	uniform sampler2D spotLightMap[ NUM_SPOT_LIGHT_MAPS ];
#endif
#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
		uniform sampler2D directionalShadowMap[ NUM_DIR_LIGHT_SHADOWS ];
		varying vec4 vDirectionalShadowCoord[ NUM_DIR_LIGHT_SHADOWS ];
		struct DirectionalLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform DirectionalLightShadow directionalLightShadows[ NUM_DIR_LIGHT_SHADOWS ];
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
		uniform sampler2D spotShadowMap[ NUM_SPOT_LIGHT_SHADOWS ];
		struct SpotLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform SpotLightShadow spotLightShadows[ NUM_SPOT_LIGHT_SHADOWS ];
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		uniform sampler2D pointShadowMap[ NUM_POINT_LIGHT_SHADOWS ];
		varying vec4 vPointShadowCoord[ NUM_POINT_LIGHT_SHADOWS ];
		struct PointLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
			float shadowCameraNear;
			float shadowCameraFar;
		};
		uniform PointLightShadow pointLightShadows[ NUM_POINT_LIGHT_SHADOWS ];
	#endif
	float texture2DCompare( sampler2D depths, vec2 uv, float compare ) {
		return step( compare, unpackRGBAToDepth( texture2D( depths, uv ) ) );
	}
	vec2 texture2DDistribution( sampler2D shadow, vec2 uv ) {
		return unpackRGBATo2Half( texture2D( shadow, uv ) );
	}
	float VSMShadow (sampler2D shadow, vec2 uv, float compare ){
		float occlusion = 1.0;
		vec2 distribution = texture2DDistribution( shadow, uv );
		float hard_shadow = step( compare , distribution.x );
		if (hard_shadow != 1.0 ) {
			float distance = compare - distribution.x ;
			float variance = max( 0.00000, distribution.y * distribution.y );
			float softness_probability = variance / (variance + distance * distance );			softness_probability = clamp( ( softness_probability - 0.3 ) / ( 0.95 - 0.3 ), 0.0, 1.0 );			occlusion = clamp( max( hard_shadow, softness_probability ), 0.0, 1.0 );
		}
		return occlusion;
	}
	float getShadow( sampler2D shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord ) {
		float shadow = 1.0;
		shadowCoord.xyz /= shadowCoord.w;
		shadowCoord.z += shadowBias;
		bool inFrustum = shadowCoord.x >= 0.0 && shadowCoord.x <= 1.0 && shadowCoord.y >= 0.0 && shadowCoord.y <= 1.0;
		bool frustumTest = inFrustum && shadowCoord.z <= 1.0;
		if ( frustumTest ) {
		#if defined( SHADOWMAP_TYPE_PCF )
			vec2 texelSize = vec2( 1.0 ) / shadowMapSize;
			float dx0 = - texelSize.x * shadowRadius;
			float dy0 = - texelSize.y * shadowRadius;
			float dx1 = + texelSize.x * shadowRadius;
			float dy1 = + texelSize.y * shadowRadius;
			float dx2 = dx0 / 2.0;
			float dy2 = dy0 / 2.0;
			float dx3 = dx1 / 2.0;
			float dy3 = dy1 / 2.0;
			shadow = (
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx0, dy0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx1, dy0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx2, dy2 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy2 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx3, dy2 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx0, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx2, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy, shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx3, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx1, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx2, dy3 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy3 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx3, dy3 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx0, dy1 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( 0.0, dy1 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, shadowCoord.xy + vec2( dx1, dy1 ), shadowCoord.z )
			) * ( 1.0 / 17.0 );
		#elif defined( SHADOWMAP_TYPE_PCF_SOFT )
			vec2 texelSize = vec2( 1.0 ) / shadowMapSize;
			float dx = texelSize.x;
			float dy = texelSize.y;
			vec2 uv = shadowCoord.xy;
			vec2 f = fract( uv * shadowMapSize + 0.5 );
			uv -= f * texelSize;
			shadow = (
				texture2DCompare( shadowMap, uv, shadowCoord.z ) +
				texture2DCompare( shadowMap, uv + vec2( dx, 0.0 ), shadowCoord.z ) +
				texture2DCompare( shadowMap, uv + vec2( 0.0, dy ), shadowCoord.z ) +
				texture2DCompare( shadowMap, uv + texelSize, shadowCoord.z ) +
				mix( texture2DCompare( shadowMap, uv + vec2( -dx, 0.0 ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, 0.0 ), shadowCoord.z ),
					 f.x ) +
				mix( texture2DCompare( shadowMap, uv + vec2( -dx, dy ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, dy ), shadowCoord.z ),
					 f.x ) +
				mix( texture2DCompare( shadowMap, uv + vec2( 0.0, -dy ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( 0.0, 2.0 * dy ), shadowCoord.z ),
					 f.y ) +
				mix( texture2DCompare( shadowMap, uv + vec2( dx, -dy ), shadowCoord.z ),
					 texture2DCompare( shadowMap, uv + vec2( dx, 2.0 * dy ), shadowCoord.z ),
					 f.y ) +
				mix( mix( texture2DCompare( shadowMap, uv + vec2( -dx, -dy ), shadowCoord.z ),
						  texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, -dy ), shadowCoord.z ),
						  f.x ),
					 mix( texture2DCompare( shadowMap, uv + vec2( -dx, 2.0 * dy ), shadowCoord.z ),
						  texture2DCompare( shadowMap, uv + vec2( 2.0 * dx, 2.0 * dy ), shadowCoord.z ),
						  f.x ),
					 f.y )
			) * ( 1.0 / 9.0 );
		#elif defined( SHADOWMAP_TYPE_VSM )
			shadow = VSMShadow( shadowMap, shadowCoord.xy, shadowCoord.z );
		#else
			shadow = texture2DCompare( shadowMap, shadowCoord.xy, shadowCoord.z );
		#endif
		}
		return mix( 1.0, shadow, shadowIntensity );
	}
	vec2 cubeToUV( vec3 v, float texelSizeY ) {
		vec3 absV = abs( v );
		float scaleToCube = 1.0 / max( absV.x, max( absV.y, absV.z ) );
		absV *= scaleToCube;
		v *= scaleToCube * ( 1.0 - 2.0 * texelSizeY );
		vec2 planar = v.xy;
		float almostATexel = 1.5 * texelSizeY;
		float almostOne = 1.0 - almostATexel;
		if ( absV.z >= almostOne ) {
			if ( v.z > 0.0 )
				planar.x = 4.0 - v.x;
		} else if ( absV.x >= almostOne ) {
			float signX = sign( v.x );
			planar.x = v.z * signX + 2.0 * signX;
		} else if ( absV.y >= almostOne ) {
			float signY = sign( v.y );
			planar.x = v.x + 2.0 * signY + 2.0;
			planar.y = v.z * signY - 2.0;
		}
		return vec2( 0.125, 0.25 ) * planar + vec2( 0.375, 0.75 );
	}
	float getPointShadow( sampler2D shadowMap, vec2 shadowMapSize, float shadowIntensity, float shadowBias, float shadowRadius, vec4 shadowCoord, float shadowCameraNear, float shadowCameraFar ) {
		float shadow = 1.0;
		vec3 lightToPosition = shadowCoord.xyz;
		
		float lightToPositionLength = length( lightToPosition );
		if ( lightToPositionLength - shadowCameraFar <= 0.0 && lightToPositionLength - shadowCameraNear >= 0.0 ) {
			float dp = ( lightToPositionLength - shadowCameraNear ) / ( shadowCameraFar - shadowCameraNear );			dp += shadowBias;
			vec3 bd3D = normalize( lightToPosition );
			vec2 texelSize = vec2( 1.0 ) / ( shadowMapSize * vec2( 4.0, 2.0 ) );
			#if defined( SHADOWMAP_TYPE_PCF ) || defined( SHADOWMAP_TYPE_PCF_SOFT ) || defined( SHADOWMAP_TYPE_VSM )
				vec2 offset = vec2( - 1, 1 ) * shadowRadius * texelSize.y;
				shadow = (
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xyy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yyy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xyx, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yyx, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xxy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yxy, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.xxx, texelSize.y ), dp ) +
					texture2DCompare( shadowMap, cubeToUV( bd3D + offset.yxx, texelSize.y ), dp )
				) * ( 1.0 / 9.0 );
			#else
				shadow = texture2DCompare( shadowMap, cubeToUV( bd3D, texelSize.y ), dp );
			#endif
		}
		return mix( 1.0, shadow, shadowIntensity );
	}
#endif`,Ou=`#if NUM_SPOT_LIGHT_COORDS > 0
	uniform mat4 spotLightMatrix[ NUM_SPOT_LIGHT_COORDS ];
	varying vec4 vSpotLightCoord[ NUM_SPOT_LIGHT_COORDS ];
#endif
#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
		uniform mat4 directionalShadowMatrix[ NUM_DIR_LIGHT_SHADOWS ];
		varying vec4 vDirectionalShadowCoord[ NUM_DIR_LIGHT_SHADOWS ];
		struct DirectionalLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform DirectionalLightShadow directionalLightShadows[ NUM_DIR_LIGHT_SHADOWS ];
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
		struct SpotLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
		};
		uniform SpotLightShadow spotLightShadows[ NUM_SPOT_LIGHT_SHADOWS ];
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		uniform mat4 pointShadowMatrix[ NUM_POINT_LIGHT_SHADOWS ];
		varying vec4 vPointShadowCoord[ NUM_POINT_LIGHT_SHADOWS ];
		struct PointLightShadow {
			float shadowIntensity;
			float shadowBias;
			float shadowNormalBias;
			float shadowRadius;
			vec2 shadowMapSize;
			float shadowCameraNear;
			float shadowCameraFar;
		};
		uniform PointLightShadow pointLightShadows[ NUM_POINT_LIGHT_SHADOWS ];
	#endif
#endif`,Fu=`#if ( defined( USE_SHADOWMAP ) && ( NUM_DIR_LIGHT_SHADOWS > 0 || NUM_POINT_LIGHT_SHADOWS > 0 ) ) || ( NUM_SPOT_LIGHT_COORDS > 0 )
	vec3 shadowWorldNormal = inverseTransformDirection( transformedNormal, viewMatrix );
	vec4 shadowWorldPosition;
#endif
#if defined( USE_SHADOWMAP )
	#if NUM_DIR_LIGHT_SHADOWS > 0
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_DIR_LIGHT_SHADOWS; i ++ ) {
			shadowWorldPosition = worldPosition + vec4( shadowWorldNormal * directionalLightShadows[ i ].shadowNormalBias, 0 );
			vDirectionalShadowCoord[ i ] = directionalShadowMatrix[ i ] * shadowWorldPosition;
		}
		#pragma unroll_loop_end
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
		#pragma unroll_loop_start
		for ( int i = 0; i < NUM_POINT_LIGHT_SHADOWS; i ++ ) {
			shadowWorldPosition = worldPosition + vec4( shadowWorldNormal * pointLightShadows[ i ].shadowNormalBias, 0 );
			vPointShadowCoord[ i ] = pointShadowMatrix[ i ] * shadowWorldPosition;
		}
		#pragma unroll_loop_end
	#endif
#endif
#if NUM_SPOT_LIGHT_COORDS > 0
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHT_COORDS; i ++ ) {
		shadowWorldPosition = worldPosition;
		#if ( defined( USE_SHADOWMAP ) && UNROLLED_LOOP_INDEX < NUM_SPOT_LIGHT_SHADOWS )
			shadowWorldPosition.xyz += shadowWorldNormal * spotLightShadows[ i ].shadowNormalBias;
		#endif
		vSpotLightCoord[ i ] = spotLightMatrix[ i ] * shadowWorldPosition;
	}
	#pragma unroll_loop_end
#endif`,zu=`float getShadowMask() {
	float shadow = 1.0;
	#ifdef USE_SHADOWMAP
	#if NUM_DIR_LIGHT_SHADOWS > 0
	DirectionalLightShadow directionalLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_DIR_LIGHT_SHADOWS; i ++ ) {
		directionalLight = directionalLightShadows[ i ];
		shadow *= receiveShadow ? getShadow( directionalShadowMap[ i ], directionalLight.shadowMapSize, directionalLight.shadowIntensity, directionalLight.shadowBias, directionalLight.shadowRadius, vDirectionalShadowCoord[ i ] ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#if NUM_SPOT_LIGHT_SHADOWS > 0
	SpotLightShadow spotLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_SPOT_LIGHT_SHADOWS; i ++ ) {
		spotLight = spotLightShadows[ i ];
		shadow *= receiveShadow ? getShadow( spotShadowMap[ i ], spotLight.shadowMapSize, spotLight.shadowIntensity, spotLight.shadowBias, spotLight.shadowRadius, vSpotLightCoord[ i ] ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#if NUM_POINT_LIGHT_SHADOWS > 0
	PointLightShadow pointLight;
	#pragma unroll_loop_start
	for ( int i = 0; i < NUM_POINT_LIGHT_SHADOWS; i ++ ) {
		pointLight = pointLightShadows[ i ];
		shadow *= receiveShadow ? getPointShadow( pointShadowMap[ i ], pointLight.shadowMapSize, pointLight.shadowIntensity, pointLight.shadowBias, pointLight.shadowRadius, vPointShadowCoord[ i ], pointLight.shadowCameraNear, pointLight.shadowCameraFar ) : 1.0;
	}
	#pragma unroll_loop_end
	#endif
	#endif
	return shadow;
}`,Bu=`#ifdef USE_SKINNING
	mat4 boneMatX = getBoneMatrix( skinIndex.x );
	mat4 boneMatY = getBoneMatrix( skinIndex.y );
	mat4 boneMatZ = getBoneMatrix( skinIndex.z );
	mat4 boneMatW = getBoneMatrix( skinIndex.w );
#endif`,ku=`#ifdef USE_SKINNING
	uniform mat4 bindMatrix;
	uniform mat4 bindMatrixInverse;
	uniform highp sampler2D boneTexture;
	mat4 getBoneMatrix( const in float i ) {
		int size = textureSize( boneTexture, 0 ).x;
		int j = int( i ) * 4;
		int x = j % size;
		int y = j / size;
		vec4 v1 = texelFetch( boneTexture, ivec2( x, y ), 0 );
		vec4 v2 = texelFetch( boneTexture, ivec2( x + 1, y ), 0 );
		vec4 v3 = texelFetch( boneTexture, ivec2( x + 2, y ), 0 );
		vec4 v4 = texelFetch( boneTexture, ivec2( x + 3, y ), 0 );
		return mat4( v1, v2, v3, v4 );
	}
#endif`,Vu=`#ifdef USE_SKINNING
	vec4 skinVertex = bindMatrix * vec4( transformed, 1.0 );
	vec4 skinned = vec4( 0.0 );
	skinned += boneMatX * skinVertex * skinWeight.x;
	skinned += boneMatY * skinVertex * skinWeight.y;
	skinned += boneMatZ * skinVertex * skinWeight.z;
	skinned += boneMatW * skinVertex * skinWeight.w;
	transformed = ( bindMatrixInverse * skinned ).xyz;
#endif`,Hu=`#ifdef USE_SKINNING
	mat4 skinMatrix = mat4( 0.0 );
	skinMatrix += skinWeight.x * boneMatX;
	skinMatrix += skinWeight.y * boneMatY;
	skinMatrix += skinWeight.z * boneMatZ;
	skinMatrix += skinWeight.w * boneMatW;
	skinMatrix = bindMatrixInverse * skinMatrix * bindMatrix;
	objectNormal = vec4( skinMatrix * vec4( objectNormal, 0.0 ) ).xyz;
	#ifdef USE_TANGENT
		objectTangent = vec4( skinMatrix * vec4( objectTangent, 0.0 ) ).xyz;
	#endif
#endif`,Gu=`float specularStrength;
#ifdef USE_SPECULARMAP
	vec4 texelSpecular = texture2D( specularMap, vSpecularMapUv );
	specularStrength = texelSpecular.r;
#else
	specularStrength = 1.0;
#endif`,Wu=`#ifdef USE_SPECULARMAP
	uniform sampler2D specularMap;
#endif`,ju=`#if defined( TONE_MAPPING )
	gl_FragColor.rgb = toneMapping( gl_FragColor.rgb );
#endif`,Xu=`#ifndef saturate
#define saturate( a ) clamp( a, 0.0, 1.0 )
#endif
uniform float toneMappingExposure;
vec3 LinearToneMapping( vec3 color ) {
	return saturate( toneMappingExposure * color );
}
vec3 ReinhardToneMapping( vec3 color ) {
	color *= toneMappingExposure;
	return saturate( color / ( vec3( 1.0 ) + color ) );
}
vec3 CineonToneMapping( vec3 color ) {
	color *= toneMappingExposure;
	color = max( vec3( 0.0 ), color - 0.004 );
	return pow( ( color * ( 6.2 * color + 0.5 ) ) / ( color * ( 6.2 * color + 1.7 ) + 0.06 ), vec3( 2.2 ) );
}
vec3 RRTAndODTFit( vec3 v ) {
	vec3 a = v * ( v + 0.0245786 ) - 0.000090537;
	vec3 b = v * ( 0.983729 * v + 0.4329510 ) + 0.238081;
	return a / b;
}
vec3 ACESFilmicToneMapping( vec3 color ) {
	const mat3 ACESInputMat = mat3(
		vec3( 0.59719, 0.07600, 0.02840 ),		vec3( 0.35458, 0.90834, 0.13383 ),
		vec3( 0.04823, 0.01566, 0.83777 )
	);
	const mat3 ACESOutputMat = mat3(
		vec3(  1.60475, -0.10208, -0.00327 ),		vec3( -0.53108,  1.10813, -0.07276 ),
		vec3( -0.07367, -0.00605,  1.07602 )
	);
	color *= toneMappingExposure / 0.6;
	color = ACESInputMat * color;
	color = RRTAndODTFit( color );
	color = ACESOutputMat * color;
	return saturate( color );
}
const mat3 LINEAR_REC2020_TO_LINEAR_SRGB = mat3(
	vec3( 1.6605, - 0.1246, - 0.0182 ),
	vec3( - 0.5876, 1.1329, - 0.1006 ),
	vec3( - 0.0728, - 0.0083, 1.1187 )
);
const mat3 LINEAR_SRGB_TO_LINEAR_REC2020 = mat3(
	vec3( 0.6274, 0.0691, 0.0164 ),
	vec3( 0.3293, 0.9195, 0.0880 ),
	vec3( 0.0433, 0.0113, 0.8956 )
);
vec3 agxDefaultContrastApprox( vec3 x ) {
	vec3 x2 = x * x;
	vec3 x4 = x2 * x2;
	return + 15.5 * x4 * x2
		- 40.14 * x4 * x
		+ 31.96 * x4
		- 6.868 * x2 * x
		+ 0.4298 * x2
		+ 0.1191 * x
		- 0.00232;
}
vec3 AgXToneMapping( vec3 color ) {
	const mat3 AgXInsetMatrix = mat3(
		vec3( 0.856627153315983, 0.137318972929847, 0.11189821299995 ),
		vec3( 0.0951212405381588, 0.761241990602591, 0.0767994186031903 ),
		vec3( 0.0482516061458583, 0.101439036467562, 0.811302368396859 )
	);
	const mat3 AgXOutsetMatrix = mat3(
		vec3( 1.1271005818144368, - 0.1413297634984383, - 0.14132976349843826 ),
		vec3( - 0.11060664309660323, 1.157823702216272, - 0.11060664309660294 ),
		vec3( - 0.016493938717834573, - 0.016493938717834257, 1.2519364065950405 )
	);
	const float AgxMinEv = - 12.47393;	const float AgxMaxEv = 4.026069;
	color *= toneMappingExposure;
	color = LINEAR_SRGB_TO_LINEAR_REC2020 * color;
	color = AgXInsetMatrix * color;
	color = max( color, 1e-10 );	color = log2( color );
	color = ( color - AgxMinEv ) / ( AgxMaxEv - AgxMinEv );
	color = clamp( color, 0.0, 1.0 );
	color = agxDefaultContrastApprox( color );
	color = AgXOutsetMatrix * color;
	color = pow( max( vec3( 0.0 ), color ), vec3( 2.2 ) );
	color = LINEAR_REC2020_TO_LINEAR_SRGB * color;
	color = clamp( color, 0.0, 1.0 );
	return color;
}
vec3 NeutralToneMapping( vec3 color ) {
	const float StartCompression = 0.8 - 0.04;
	const float Desaturation = 0.15;
	color *= toneMappingExposure;
	float x = min( color.r, min( color.g, color.b ) );
	float offset = x < 0.08 ? x - 6.25 * x * x : 0.04;
	color -= offset;
	float peak = max( color.r, max( color.g, color.b ) );
	if ( peak < StartCompression ) return color;
	float d = 1. - StartCompression;
	float newPeak = 1. - d * d / ( peak + d - StartCompression );
	color *= newPeak / peak;
	float g = 1. - 1. / ( Desaturation * ( peak - newPeak ) + 1. );
	return mix( color, vec3( newPeak ), g );
}
vec3 CustomToneMapping( vec3 color ) { return color; }`,Yu=`#ifdef USE_TRANSMISSION
	material.transmission = transmission;
	material.transmissionAlpha = 1.0;
	material.thickness = thickness;
	material.attenuationDistance = attenuationDistance;
	material.attenuationColor = attenuationColor;
	#ifdef USE_TRANSMISSIONMAP
		material.transmission *= texture2D( transmissionMap, vTransmissionMapUv ).r;
	#endif
	#ifdef USE_THICKNESSMAP
		material.thickness *= texture2D( thicknessMap, vThicknessMapUv ).g;
	#endif
	vec3 pos = vWorldPosition;
	vec3 v = normalize( cameraPosition - pos );
	vec3 n = inverseTransformDirection( normal, viewMatrix );
	vec4 transmitted = getIBLVolumeRefraction(
		n, v, material.roughness, material.diffuseColor, material.specularColor, material.specularF90,
		pos, modelMatrix, viewMatrix, projectionMatrix, material.dispersion, material.ior, material.thickness,
		material.attenuationColor, material.attenuationDistance );
	material.transmissionAlpha = mix( material.transmissionAlpha, transmitted.a, material.transmission );
	totalDiffuse = mix( totalDiffuse, transmitted.rgb, material.transmission );
#endif`,qu=`#ifdef USE_TRANSMISSION
	uniform float transmission;
	uniform float thickness;
	uniform float attenuationDistance;
	uniform vec3 attenuationColor;
	#ifdef USE_TRANSMISSIONMAP
		uniform sampler2D transmissionMap;
	#endif
	#ifdef USE_THICKNESSMAP
		uniform sampler2D thicknessMap;
	#endif
	uniform vec2 transmissionSamplerSize;
	uniform sampler2D transmissionSamplerMap;
	uniform mat4 modelMatrix;
	uniform mat4 projectionMatrix;
	varying vec3 vWorldPosition;
	float w0( float a ) {
		return ( 1.0 / 6.0 ) * ( a * ( a * ( - a + 3.0 ) - 3.0 ) + 1.0 );
	}
	float w1( float a ) {
		return ( 1.0 / 6.0 ) * ( a *  a * ( 3.0 * a - 6.0 ) + 4.0 );
	}
	float w2( float a ){
		return ( 1.0 / 6.0 ) * ( a * ( a * ( - 3.0 * a + 3.0 ) + 3.0 ) + 1.0 );
	}
	float w3( float a ) {
		return ( 1.0 / 6.0 ) * ( a * a * a );
	}
	float g0( float a ) {
		return w0( a ) + w1( a );
	}
	float g1( float a ) {
		return w2( a ) + w3( a );
	}
	float h0( float a ) {
		return - 1.0 + w1( a ) / ( w0( a ) + w1( a ) );
	}
	float h1( float a ) {
		return 1.0 + w3( a ) / ( w2( a ) + w3( a ) );
	}
	vec4 bicubic( sampler2D tex, vec2 uv, vec4 texelSize, float lod ) {
		uv = uv * texelSize.zw + 0.5;
		vec2 iuv = floor( uv );
		vec2 fuv = fract( uv );
		float g0x = g0( fuv.x );
		float g1x = g1( fuv.x );
		float h0x = h0( fuv.x );
		float h1x = h1( fuv.x );
		float h0y = h0( fuv.y );
		float h1y = h1( fuv.y );
		vec2 p0 = ( vec2( iuv.x + h0x, iuv.y + h0y ) - 0.5 ) * texelSize.xy;
		vec2 p1 = ( vec2( iuv.x + h1x, iuv.y + h0y ) - 0.5 ) * texelSize.xy;
		vec2 p2 = ( vec2( iuv.x + h0x, iuv.y + h1y ) - 0.5 ) * texelSize.xy;
		vec2 p3 = ( vec2( iuv.x + h1x, iuv.y + h1y ) - 0.5 ) * texelSize.xy;
		return g0( fuv.y ) * ( g0x * textureLod( tex, p0, lod ) + g1x * textureLod( tex, p1, lod ) ) +
			g1( fuv.y ) * ( g0x * textureLod( tex, p2, lod ) + g1x * textureLod( tex, p3, lod ) );
	}
	vec4 textureBicubic( sampler2D sampler, vec2 uv, float lod ) {
		vec2 fLodSize = vec2( textureSize( sampler, int( lod ) ) );
		vec2 cLodSize = vec2( textureSize( sampler, int( lod + 1.0 ) ) );
		vec2 fLodSizeInv = 1.0 / fLodSize;
		vec2 cLodSizeInv = 1.0 / cLodSize;
		vec4 fSample = bicubic( sampler, uv, vec4( fLodSizeInv, fLodSize ), floor( lod ) );
		vec4 cSample = bicubic( sampler, uv, vec4( cLodSizeInv, cLodSize ), ceil( lod ) );
		return mix( fSample, cSample, fract( lod ) );
	}
	vec3 getVolumeTransmissionRay( const in vec3 n, const in vec3 v, const in float thickness, const in float ior, const in mat4 modelMatrix ) {
		vec3 refractionVector = refract( - v, normalize( n ), 1.0 / ior );
		vec3 modelScale;
		modelScale.x = length( vec3( modelMatrix[ 0 ].xyz ) );
		modelScale.y = length( vec3( modelMatrix[ 1 ].xyz ) );
		modelScale.z = length( vec3( modelMatrix[ 2 ].xyz ) );
		return normalize( refractionVector ) * thickness * modelScale;
	}
	float applyIorToRoughness( const in float roughness, const in float ior ) {
		return roughness * clamp( ior * 2.0 - 2.0, 0.0, 1.0 );
	}
	vec4 getTransmissionSample( const in vec2 fragCoord, const in float roughness, const in float ior ) {
		float lod = log2( transmissionSamplerSize.x ) * applyIorToRoughness( roughness, ior );
		return textureBicubic( transmissionSamplerMap, fragCoord.xy, lod );
	}
	vec3 volumeAttenuation( const in float transmissionDistance, const in vec3 attenuationColor, const in float attenuationDistance ) {
		if ( isinf( attenuationDistance ) ) {
			return vec3( 1.0 );
		} else {
			vec3 attenuationCoefficient = -log( attenuationColor ) / attenuationDistance;
			vec3 transmittance = exp( - attenuationCoefficient * transmissionDistance );			return transmittance;
		}
	}
	vec4 getIBLVolumeRefraction( const in vec3 n, const in vec3 v, const in float roughness, const in vec3 diffuseColor,
		const in vec3 specularColor, const in float specularF90, const in vec3 position, const in mat4 modelMatrix,
		const in mat4 viewMatrix, const in mat4 projMatrix, const in float dispersion, const in float ior, const in float thickness,
		const in vec3 attenuationColor, const in float attenuationDistance ) {
		vec4 transmittedLight;
		vec3 transmittance;
		#ifdef USE_DISPERSION
			float halfSpread = ( ior - 1.0 ) * 0.025 * dispersion;
			vec3 iors = vec3( ior - halfSpread, ior, ior + halfSpread );
			for ( int i = 0; i < 3; i ++ ) {
				vec3 transmissionRay = getVolumeTransmissionRay( n, v, thickness, iors[ i ], modelMatrix );
				vec3 refractedRayExit = position + transmissionRay;
		
				vec4 ndcPos = projMatrix * viewMatrix * vec4( refractedRayExit, 1.0 );
				vec2 refractionCoords = ndcPos.xy / ndcPos.w;
				refractionCoords += 1.0;
				refractionCoords /= 2.0;
		
				vec4 transmissionSample = getTransmissionSample( refractionCoords, roughness, iors[ i ] );
				transmittedLight[ i ] = transmissionSample[ i ];
				transmittedLight.a += transmissionSample.a;
				transmittance[ i ] = diffuseColor[ i ] * volumeAttenuation( length( transmissionRay ), attenuationColor, attenuationDistance )[ i ];
			}
			transmittedLight.a /= 3.0;
		
		#else
		
			vec3 transmissionRay = getVolumeTransmissionRay( n, v, thickness, ior, modelMatrix );
			vec3 refractedRayExit = position + transmissionRay;
			vec4 ndcPos = projMatrix * viewMatrix * vec4( refractedRayExit, 1.0 );
			vec2 refractionCoords = ndcPos.xy / ndcPos.w;
			refractionCoords += 1.0;
			refractionCoords /= 2.0;
			transmittedLight = getTransmissionSample( refractionCoords, roughness, ior );
			transmittance = diffuseColor * volumeAttenuation( length( transmissionRay ), attenuationColor, attenuationDistance );
		
		#endif
		vec3 attenuatedColor = transmittance * transmittedLight.rgb;
		vec3 F = EnvironmentBRDF( n, v, specularColor, specularF90, roughness );
		float transmittanceFactor = ( transmittance.r + transmittance.g + transmittance.b ) / 3.0;
		return vec4( ( 1.0 - F ) * attenuatedColor, 1.0 - ( 1.0 - transmittedLight.a ) * transmittanceFactor );
	}
#endif`,Ku=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	varying vec2 vUv;
#endif
#ifdef USE_MAP
	varying vec2 vMapUv;
#endif
#ifdef USE_ALPHAMAP
	varying vec2 vAlphaMapUv;
#endif
#ifdef USE_LIGHTMAP
	varying vec2 vLightMapUv;
#endif
#ifdef USE_AOMAP
	varying vec2 vAoMapUv;
#endif
#ifdef USE_BUMPMAP
	varying vec2 vBumpMapUv;
#endif
#ifdef USE_NORMALMAP
	varying vec2 vNormalMapUv;
#endif
#ifdef USE_EMISSIVEMAP
	varying vec2 vEmissiveMapUv;
#endif
#ifdef USE_METALNESSMAP
	varying vec2 vMetalnessMapUv;
#endif
#ifdef USE_ROUGHNESSMAP
	varying vec2 vRoughnessMapUv;
#endif
#ifdef USE_ANISOTROPYMAP
	varying vec2 vAnisotropyMapUv;
#endif
#ifdef USE_CLEARCOATMAP
	varying vec2 vClearcoatMapUv;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	varying vec2 vClearcoatNormalMapUv;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	varying vec2 vClearcoatRoughnessMapUv;
#endif
#ifdef USE_IRIDESCENCEMAP
	varying vec2 vIridescenceMapUv;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	varying vec2 vIridescenceThicknessMapUv;
#endif
#ifdef USE_SHEEN_COLORMAP
	varying vec2 vSheenColorMapUv;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	varying vec2 vSheenRoughnessMapUv;
#endif
#ifdef USE_SPECULARMAP
	varying vec2 vSpecularMapUv;
#endif
#ifdef USE_SPECULAR_COLORMAP
	varying vec2 vSpecularColorMapUv;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	varying vec2 vSpecularIntensityMapUv;
#endif
#ifdef USE_TRANSMISSIONMAP
	uniform mat3 transmissionMapTransform;
	varying vec2 vTransmissionMapUv;
#endif
#ifdef USE_THICKNESSMAP
	uniform mat3 thicknessMapTransform;
	varying vec2 vThicknessMapUv;
#endif`,$u=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	varying vec2 vUv;
#endif
#ifdef USE_MAP
	uniform mat3 mapTransform;
	varying vec2 vMapUv;
#endif
#ifdef USE_ALPHAMAP
	uniform mat3 alphaMapTransform;
	varying vec2 vAlphaMapUv;
#endif
#ifdef USE_LIGHTMAP
	uniform mat3 lightMapTransform;
	varying vec2 vLightMapUv;
#endif
#ifdef USE_AOMAP
	uniform mat3 aoMapTransform;
	varying vec2 vAoMapUv;
#endif
#ifdef USE_BUMPMAP
	uniform mat3 bumpMapTransform;
	varying vec2 vBumpMapUv;
#endif
#ifdef USE_NORMALMAP
	uniform mat3 normalMapTransform;
	varying vec2 vNormalMapUv;
#endif
#ifdef USE_DISPLACEMENTMAP
	uniform mat3 displacementMapTransform;
	varying vec2 vDisplacementMapUv;
#endif
#ifdef USE_EMISSIVEMAP
	uniform mat3 emissiveMapTransform;
	varying vec2 vEmissiveMapUv;
#endif
#ifdef USE_METALNESSMAP
	uniform mat3 metalnessMapTransform;
	varying vec2 vMetalnessMapUv;
#endif
#ifdef USE_ROUGHNESSMAP
	uniform mat3 roughnessMapTransform;
	varying vec2 vRoughnessMapUv;
#endif
#ifdef USE_ANISOTROPYMAP
	uniform mat3 anisotropyMapTransform;
	varying vec2 vAnisotropyMapUv;
#endif
#ifdef USE_CLEARCOATMAP
	uniform mat3 clearcoatMapTransform;
	varying vec2 vClearcoatMapUv;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	uniform mat3 clearcoatNormalMapTransform;
	varying vec2 vClearcoatNormalMapUv;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	uniform mat3 clearcoatRoughnessMapTransform;
	varying vec2 vClearcoatRoughnessMapUv;
#endif
#ifdef USE_SHEEN_COLORMAP
	uniform mat3 sheenColorMapTransform;
	varying vec2 vSheenColorMapUv;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	uniform mat3 sheenRoughnessMapTransform;
	varying vec2 vSheenRoughnessMapUv;
#endif
#ifdef USE_IRIDESCENCEMAP
	uniform mat3 iridescenceMapTransform;
	varying vec2 vIridescenceMapUv;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	uniform mat3 iridescenceThicknessMapTransform;
	varying vec2 vIridescenceThicknessMapUv;
#endif
#ifdef USE_SPECULARMAP
	uniform mat3 specularMapTransform;
	varying vec2 vSpecularMapUv;
#endif
#ifdef USE_SPECULAR_COLORMAP
	uniform mat3 specularColorMapTransform;
	varying vec2 vSpecularColorMapUv;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	uniform mat3 specularIntensityMapTransform;
	varying vec2 vSpecularIntensityMapUv;
#endif
#ifdef USE_TRANSMISSIONMAP
	uniform mat3 transmissionMapTransform;
	varying vec2 vTransmissionMapUv;
#endif
#ifdef USE_THICKNESSMAP
	uniform mat3 thicknessMapTransform;
	varying vec2 vThicknessMapUv;
#endif`,Zu=`#if defined( USE_UV ) || defined( USE_ANISOTROPY )
	vUv = vec3( uv, 1 ).xy;
#endif
#ifdef USE_MAP
	vMapUv = ( mapTransform * vec3( MAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ALPHAMAP
	vAlphaMapUv = ( alphaMapTransform * vec3( ALPHAMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_LIGHTMAP
	vLightMapUv = ( lightMapTransform * vec3( LIGHTMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_AOMAP
	vAoMapUv = ( aoMapTransform * vec3( AOMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_BUMPMAP
	vBumpMapUv = ( bumpMapTransform * vec3( BUMPMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_NORMALMAP
	vNormalMapUv = ( normalMapTransform * vec3( NORMALMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_DISPLACEMENTMAP
	vDisplacementMapUv = ( displacementMapTransform * vec3( DISPLACEMENTMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_EMISSIVEMAP
	vEmissiveMapUv = ( emissiveMapTransform * vec3( EMISSIVEMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_METALNESSMAP
	vMetalnessMapUv = ( metalnessMapTransform * vec3( METALNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ROUGHNESSMAP
	vRoughnessMapUv = ( roughnessMapTransform * vec3( ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_ANISOTROPYMAP
	vAnisotropyMapUv = ( anisotropyMapTransform * vec3( ANISOTROPYMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOATMAP
	vClearcoatMapUv = ( clearcoatMapTransform * vec3( CLEARCOATMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOAT_NORMALMAP
	vClearcoatNormalMapUv = ( clearcoatNormalMapTransform * vec3( CLEARCOAT_NORMALMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_CLEARCOAT_ROUGHNESSMAP
	vClearcoatRoughnessMapUv = ( clearcoatRoughnessMapTransform * vec3( CLEARCOAT_ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_IRIDESCENCEMAP
	vIridescenceMapUv = ( iridescenceMapTransform * vec3( IRIDESCENCEMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_IRIDESCENCE_THICKNESSMAP
	vIridescenceThicknessMapUv = ( iridescenceThicknessMapTransform * vec3( IRIDESCENCE_THICKNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SHEEN_COLORMAP
	vSheenColorMapUv = ( sheenColorMapTransform * vec3( SHEEN_COLORMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SHEEN_ROUGHNESSMAP
	vSheenRoughnessMapUv = ( sheenRoughnessMapTransform * vec3( SHEEN_ROUGHNESSMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULARMAP
	vSpecularMapUv = ( specularMapTransform * vec3( SPECULARMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULAR_COLORMAP
	vSpecularColorMapUv = ( specularColorMapTransform * vec3( SPECULAR_COLORMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_SPECULAR_INTENSITYMAP
	vSpecularIntensityMapUv = ( specularIntensityMapTransform * vec3( SPECULAR_INTENSITYMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_TRANSMISSIONMAP
	vTransmissionMapUv = ( transmissionMapTransform * vec3( TRANSMISSIONMAP_UV, 1 ) ).xy;
#endif
#ifdef USE_THICKNESSMAP
	vThicknessMapUv = ( thicknessMapTransform * vec3( THICKNESSMAP_UV, 1 ) ).xy;
#endif`,Ju=`#if defined( USE_ENVMAP ) || defined( DISTANCE ) || defined ( USE_SHADOWMAP ) || defined ( USE_TRANSMISSION ) || NUM_SPOT_LIGHT_COORDS > 0
	vec4 worldPosition = vec4( transformed, 1.0 );
	#ifdef USE_BATCHING
		worldPosition = batchingMatrix * worldPosition;
	#endif
	#ifdef USE_INSTANCING
		worldPosition = instanceMatrix * worldPosition;
	#endif
	worldPosition = modelMatrix * worldPosition;
#endif`;const Qu=`varying vec2 vUv;
uniform mat3 uvTransform;
void main() {
	vUv = ( uvTransform * vec3( uv, 1 ) ).xy;
	gl_Position = vec4( position.xy, 1.0, 1.0 );
}`,ed=`uniform sampler2D t2D;
uniform float backgroundIntensity;
varying vec2 vUv;
void main() {
	vec4 texColor = texture2D( t2D, vUv );
	#ifdef DECODE_VIDEO_TEXTURE
		texColor = vec4( mix( pow( texColor.rgb * 0.9478672986 + vec3( 0.0521327014 ), vec3( 2.4 ) ), texColor.rgb * 0.0773993808, vec3( lessThanEqual( texColor.rgb, vec3( 0.04045 ) ) ) ), texColor.w );
	#endif
	texColor.rgb *= backgroundIntensity;
	gl_FragColor = texColor;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,td=`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
	gl_Position.z = gl_Position.w;
}`,rd=`#ifdef ENVMAP_TYPE_CUBE
	uniform samplerCube envMap;
#elif defined( ENVMAP_TYPE_CUBE_UV )
	uniform sampler2D envMap;
#endif
uniform float flipEnvMap;
uniform float backgroundBlurriness;
uniform float backgroundIntensity;
uniform mat3 backgroundRotation;
varying vec3 vWorldDirection;
#include <cube_uv_reflection_fragment>
void main() {
	#ifdef ENVMAP_TYPE_CUBE
		vec4 texColor = textureCube( envMap, backgroundRotation * vec3( flipEnvMap * vWorldDirection.x, vWorldDirection.yz ) );
	#elif defined( ENVMAP_TYPE_CUBE_UV )
		vec4 texColor = textureCubeUV( envMap, backgroundRotation * vWorldDirection, backgroundBlurriness );
	#else
		vec4 texColor = vec4( 0.0, 0.0, 0.0, 1.0 );
	#endif
	texColor.rgb *= backgroundIntensity;
	gl_FragColor = texColor;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,id=`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
	gl_Position.z = gl_Position.w;
}`,nd=`uniform samplerCube tCube;
uniform float tFlip;
uniform float opacity;
varying vec3 vWorldDirection;
void main() {
	vec4 texColor = textureCube( tCube, vec3( tFlip * vWorldDirection.x, vWorldDirection.yz ) );
	gl_FragColor = texColor;
	gl_FragColor.a *= opacity;
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,ad=`#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
varying vec2 vHighPrecisionZW;
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <skinbase_vertex>
	#include <morphinstance_vertex>
	#ifdef USE_DISPLACEMENTMAP
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vHighPrecisionZW = gl_Position.zw;
}`,sd=`#if DEPTH_PACKING == 3200
	uniform float opacity;
#endif
#include <common>
#include <packing>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
varying vec2 vHighPrecisionZW;
void main() {
	vec4 diffuseColor = vec4( 1.0 );
	#include <clipping_planes_fragment>
	#if DEPTH_PACKING == 3200
		diffuseColor.a = opacity;
	#endif
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <logdepthbuf_fragment>
	float fragCoordZ = 0.5 * vHighPrecisionZW[0] / vHighPrecisionZW[1] + 0.5;
	#if DEPTH_PACKING == 3200
		gl_FragColor = vec4( vec3( 1.0 - fragCoordZ ), opacity );
	#elif DEPTH_PACKING == 3201
		gl_FragColor = packDepthToRGBA( fragCoordZ );
	#elif DEPTH_PACKING == 3202
		gl_FragColor = vec4( packDepthToRGB( fragCoordZ ), 1.0 );
	#elif DEPTH_PACKING == 3203
		gl_FragColor = vec4( packDepthToRG( fragCoordZ ), 0.0, 1.0 );
	#endif
}`,od=`#define DISTANCE
varying vec3 vWorldPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <skinbase_vertex>
	#include <morphinstance_vertex>
	#ifdef USE_DISPLACEMENTMAP
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <worldpos_vertex>
	#include <clipping_planes_vertex>
	vWorldPosition = worldPosition.xyz;
}`,ld=`#define DISTANCE
uniform vec3 referencePosition;
uniform float nearDistance;
uniform float farDistance;
varying vec3 vWorldPosition;
#include <common>
#include <packing>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <clipping_planes_pars_fragment>
void main () {
	vec4 diffuseColor = vec4( 1.0 );
	#include <clipping_planes_fragment>
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	float dist = length( vWorldPosition - referencePosition );
	dist = ( dist - nearDistance ) / ( farDistance - nearDistance );
	dist = saturate( dist );
	gl_FragColor = packDepthToRGBA( dist );
}`,cd=`varying vec3 vWorldDirection;
#include <common>
void main() {
	vWorldDirection = transformDirection( position, modelMatrix );
	#include <begin_vertex>
	#include <project_vertex>
}`,hd=`uniform sampler2D tEquirect;
varying vec3 vWorldDirection;
#include <common>
void main() {
	vec3 direction = normalize( vWorldDirection );
	vec2 sampleUV = equirectUv( direction );
	gl_FragColor = texture2D( tEquirect, sampleUV );
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
}`,ud=`uniform float scale;
attribute float lineDistance;
varying float vLineDistance;
#include <common>
#include <uv_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	vLineDistance = scale * lineDistance;
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
}`,dd=`uniform vec3 diffuse;
uniform float opacity;
uniform float dashSize;
uniform float totalSize;
varying float vLineDistance;
#include <common>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	if ( mod( vLineDistance, totalSize ) > dashSize ) {
		discard;
	}
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,pd=`#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#if defined ( USE_ENVMAP ) || defined ( USE_SKINNING )
		#include <beginnormal_vertex>
		#include <morphnormal_vertex>
		#include <skinbase_vertex>
		#include <skinnormal_vertex>
		#include <defaultnormal_vertex>
	#endif
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <fog_vertex>
}`,fd=`uniform vec3 diffuse;
uniform float opacity;
#ifndef FLAT_SHADED
	varying vec3 vNormal;
#endif
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	#ifdef USE_LIGHTMAP
		vec4 lightMapTexel = texture2D( lightMap, vLightMapUv );
		reflectedLight.indirectDiffuse += lightMapTexel.rgb * lightMapIntensity * RECIPROCAL_PI;
	#else
		reflectedLight.indirectDiffuse += vec3( 1.0 );
	#endif
	#include <aomap_fragment>
	reflectedLight.indirectDiffuse *= diffuseColor.rgb;
	vec3 outgoingLight = reflectedLight.indirectDiffuse;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,md=`#define LAMBERT
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,gd=`#define LAMBERT
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float opacity;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_lambert_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_lambert_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + totalEmissiveRadiance;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,_d=`#define MATCAP
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <color_pars_vertex>
#include <displacementmap_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
	vViewPosition = - mvPosition.xyz;
}`,vd=`#define MATCAP
uniform vec3 diffuse;
uniform float opacity;
uniform sampler2D matcap;
varying vec3 vViewPosition;
#include <common>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <normal_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	vec3 viewDir = normalize( vViewPosition );
	vec3 x = normalize( vec3( viewDir.z, 0.0, - viewDir.x ) );
	vec3 y = cross( viewDir, x );
	vec2 uv = vec2( dot( x, normal ), dot( y, normal ) ) * 0.495 + 0.5;
	#ifdef USE_MATCAP
		vec4 matcapColor = texture2D( matcap, uv );
	#else
		vec4 matcapColor = vec4( vec3( mix( 0.2, 0.8, uv.y ) ), 1.0 );
	#endif
	vec3 outgoingLight = diffuseColor.rgb * matcapColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,xd=`#define NORMAL
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	varying vec3 vViewPosition;
#endif
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	vViewPosition = - mvPosition.xyz;
#endif
}`,yd=`#define NORMAL
uniform float opacity;
#if defined( FLAT_SHADED ) || defined( USE_BUMPMAP ) || defined( USE_NORMALMAP_TANGENTSPACE )
	varying vec3 vViewPosition;
#endif
#include <packing>
#include <uv_pars_fragment>
#include <normal_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( 0.0, 0.0, 0.0, opacity );
	#include <clipping_planes_fragment>
	#include <logdepthbuf_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	gl_FragColor = vec4( packNormalToRGB( normal ), diffuseColor.a );
	#ifdef OPAQUE
		gl_FragColor.a = 1.0;
	#endif
}`,Md=`#define PHONG
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <envmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <envmap_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,Sd=`#define PHONG
uniform vec3 diffuse;
uniform vec3 emissive;
uniform vec3 specular;
uniform float shininess;
uniform float opacity;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_phong_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <specularmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <specularmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_phong_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + reflectedLight.directSpecular + reflectedLight.indirectSpecular + totalEmissiveRadiance;
	#include <envmap_fragment>
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,Ed=`#define STANDARD
varying vec3 vViewPosition;
#ifdef USE_TRANSMISSION
	varying vec3 vWorldPosition;
#endif
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
#ifdef USE_TRANSMISSION
	vWorldPosition = worldPosition.xyz;
#endif
}`,bd=`#define STANDARD
#ifdef PHYSICAL
	#define IOR
	#define USE_SPECULAR
#endif
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float roughness;
uniform float metalness;
uniform float opacity;
#ifdef IOR
	uniform float ior;
#endif
#ifdef USE_SPECULAR
	uniform float specularIntensity;
	uniform vec3 specularColor;
	#ifdef USE_SPECULAR_COLORMAP
		uniform sampler2D specularColorMap;
	#endif
	#ifdef USE_SPECULAR_INTENSITYMAP
		uniform sampler2D specularIntensityMap;
	#endif
#endif
#ifdef USE_CLEARCOAT
	uniform float clearcoat;
	uniform float clearcoatRoughness;
#endif
#ifdef USE_DISPERSION
	uniform float dispersion;
#endif
#ifdef USE_IRIDESCENCE
	uniform float iridescence;
	uniform float iridescenceIOR;
	uniform float iridescenceThicknessMinimum;
	uniform float iridescenceThicknessMaximum;
#endif
#ifdef USE_SHEEN
	uniform vec3 sheenColor;
	uniform float sheenRoughness;
	#ifdef USE_SHEEN_COLORMAP
		uniform sampler2D sheenColorMap;
	#endif
	#ifdef USE_SHEEN_ROUGHNESSMAP
		uniform sampler2D sheenRoughnessMap;
	#endif
#endif
#ifdef USE_ANISOTROPY
	uniform vec2 anisotropyVector;
	#ifdef USE_ANISOTROPYMAP
		uniform sampler2D anisotropyMap;
	#endif
#endif
varying vec3 vViewPosition;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <iridescence_fragment>
#include <cube_uv_reflection_fragment>
#include <envmap_common_pars_fragment>
#include <envmap_physical_pars_fragment>
#include <fog_pars_fragment>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_physical_pars_fragment>
#include <transmission_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <clearcoat_pars_fragment>
#include <iridescence_pars_fragment>
#include <roughnessmap_pars_fragment>
#include <metalnessmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <roughnessmap_fragment>
	#include <metalnessmap_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <clearcoat_normal_fragment_begin>
	#include <clearcoat_normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_physical_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 totalDiffuse = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse;
	vec3 totalSpecular = reflectedLight.directSpecular + reflectedLight.indirectSpecular;
	#include <transmission_fragment>
	vec3 outgoingLight = totalDiffuse + totalSpecular + totalEmissiveRadiance;
	#ifdef USE_SHEEN
		float sheenEnergyComp = 1.0 - 0.157 * max3( material.sheenColor );
		outgoingLight = outgoingLight * sheenEnergyComp + sheenSpecularDirect + sheenSpecularIndirect;
	#endif
	#ifdef USE_CLEARCOAT
		float dotNVcc = saturate( dot( geometryClearcoatNormal, geometryViewDir ) );
		vec3 Fcc = F_Schlick( material.clearcoatF0, material.clearcoatF90, dotNVcc );
		outgoingLight = outgoingLight * ( 1.0 - material.clearcoat * Fcc ) + ( clearcoatSpecularDirect + clearcoatSpecularIndirect ) * material.clearcoat;
	#endif
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,Td=`#define TOON
varying vec3 vViewPosition;
#include <common>
#include <batching_pars_vertex>
#include <uv_pars_vertex>
#include <displacementmap_pars_vertex>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <normal_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <shadowmap_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <normal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <displacementmap_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	vViewPosition = - mvPosition.xyz;
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,wd=`#define TOON
uniform vec3 diffuse;
uniform vec3 emissive;
uniform float opacity;
#include <common>
#include <packing>
#include <dithering_pars_fragment>
#include <color_pars_fragment>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <aomap_pars_fragment>
#include <lightmap_pars_fragment>
#include <emissivemap_pars_fragment>
#include <gradientmap_pars_fragment>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <normal_pars_fragment>
#include <lights_toon_pars_fragment>
#include <shadowmap_pars_fragment>
#include <bumpmap_pars_fragment>
#include <normalmap_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	ReflectedLight reflectedLight = ReflectedLight( vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ), vec3( 0.0 ) );
	vec3 totalEmissiveRadiance = emissive;
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <color_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	#include <normal_fragment_begin>
	#include <normal_fragment_maps>
	#include <emissivemap_fragment>
	#include <lights_toon_fragment>
	#include <lights_fragment_begin>
	#include <lights_fragment_maps>
	#include <lights_fragment_end>
	#include <aomap_fragment>
	vec3 outgoingLight = reflectedLight.directDiffuse + reflectedLight.indirectDiffuse + totalEmissiveRadiance;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
	#include <dithering_fragment>
}`,Ad=`uniform float size;
uniform float scale;
#include <common>
#include <color_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
#ifdef USE_POINTS_UV
	varying vec2 vUv;
	uniform mat3 uvTransform;
#endif
void main() {
	#ifdef USE_POINTS_UV
		vUv = ( uvTransform * vec3( uv, 1 ) ).xy;
	#endif
	#include <color_vertex>
	#include <morphinstance_vertex>
	#include <morphcolor_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <project_vertex>
	gl_PointSize = size;
	#ifdef USE_SIZEATTENUATION
		bool isPerspective = isPerspectiveMatrix( projectionMatrix );
		if ( isPerspective ) gl_PointSize *= ( scale / - mvPosition.z );
	#endif
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <worldpos_vertex>
	#include <fog_vertex>
}`,Rd=`uniform vec3 diffuse;
uniform float opacity;
#include <common>
#include <color_pars_fragment>
#include <map_particle_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_particle_fragment>
	#include <color_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
	#include <premultiplied_alpha_fragment>
}`,Cd=`#include <common>
#include <batching_pars_vertex>
#include <fog_pars_vertex>
#include <morphtarget_pars_vertex>
#include <skinning_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <shadowmap_pars_vertex>
void main() {
	#include <batching_vertex>
	#include <beginnormal_vertex>
	#include <morphinstance_vertex>
	#include <morphnormal_vertex>
	#include <skinbase_vertex>
	#include <skinnormal_vertex>
	#include <defaultnormal_vertex>
	#include <begin_vertex>
	#include <morphtarget_vertex>
	#include <skinning_vertex>
	#include <project_vertex>
	#include <logdepthbuf_vertex>
	#include <worldpos_vertex>
	#include <shadowmap_vertex>
	#include <fog_vertex>
}`,Ld=`uniform vec3 color;
uniform float opacity;
#include <common>
#include <packing>
#include <fog_pars_fragment>
#include <bsdfs>
#include <lights_pars_begin>
#include <logdepthbuf_pars_fragment>
#include <shadowmap_pars_fragment>
#include <shadowmask_pars_fragment>
void main() {
	#include <logdepthbuf_fragment>
	gl_FragColor = vec4( color, opacity * ( 1.0 - getShadowMask() ) );
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
}`,Pd=`uniform float rotation;
uniform vec2 center;
#include <common>
#include <uv_pars_vertex>
#include <fog_pars_vertex>
#include <logdepthbuf_pars_vertex>
#include <clipping_planes_pars_vertex>
void main() {
	#include <uv_vertex>
	vec4 mvPosition = modelViewMatrix[ 3 ];
	vec2 scale = vec2( length( modelMatrix[ 0 ].xyz ), length( modelMatrix[ 1 ].xyz ) );
	#ifndef USE_SIZEATTENUATION
		bool isPerspective = isPerspectiveMatrix( projectionMatrix );
		if ( isPerspective ) scale *= - mvPosition.z;
	#endif
	vec2 alignedPosition = ( position.xy - ( center - vec2( 0.5 ) ) ) * scale;
	vec2 rotatedPosition;
	rotatedPosition.x = cos( rotation ) * alignedPosition.x - sin( rotation ) * alignedPosition.y;
	rotatedPosition.y = sin( rotation ) * alignedPosition.x + cos( rotation ) * alignedPosition.y;
	mvPosition.xy += rotatedPosition;
	gl_Position = projectionMatrix * mvPosition;
	#include <logdepthbuf_vertex>
	#include <clipping_planes_vertex>
	#include <fog_vertex>
}`,Dd=`uniform vec3 diffuse;
uniform float opacity;
#include <common>
#include <uv_pars_fragment>
#include <map_pars_fragment>
#include <alphamap_pars_fragment>
#include <alphatest_pars_fragment>
#include <alphahash_pars_fragment>
#include <fog_pars_fragment>
#include <logdepthbuf_pars_fragment>
#include <clipping_planes_pars_fragment>
void main() {
	vec4 diffuseColor = vec4( diffuse, opacity );
	#include <clipping_planes_fragment>
	vec3 outgoingLight = vec3( 0.0 );
	#include <logdepthbuf_fragment>
	#include <map_fragment>
	#include <alphamap_fragment>
	#include <alphatest_fragment>
	#include <alphahash_fragment>
	outgoingLight = diffuseColor.rgb;
	#include <opaque_fragment>
	#include <tonemapping_fragment>
	#include <colorspace_fragment>
	#include <fog_fragment>
}`,Ge={alphahash_fragment:eh,alphahash_pars_fragment:th,alphamap_fragment:rh,alphamap_pars_fragment:ih,alphatest_fragment:nh,alphatest_pars_fragment:ah,aomap_fragment:sh,aomap_pars_fragment:oh,batching_pars_vertex:lh,batching_vertex:ch,begin_vertex:hh,beginnormal_vertex:uh,bsdfs:dh,iridescence_fragment:ph,bumpmap_pars_fragment:fh,clipping_planes_fragment:mh,clipping_planes_pars_fragment:gh,clipping_planes_pars_vertex:_h,clipping_planes_vertex:vh,color_fragment:xh,color_pars_fragment:yh,color_pars_vertex:Mh,color_vertex:Sh,common:Eh,cube_uv_reflection_fragment:bh,defaultnormal_vertex:Th,displacementmap_pars_vertex:wh,displacementmap_vertex:Ah,emissivemap_fragment:Rh,emissivemap_pars_fragment:Ch,colorspace_fragment:Lh,colorspace_pars_fragment:Ph,envmap_fragment:Dh,envmap_common_pars_fragment:Ih,envmap_pars_fragment:Nh,envmap_pars_vertex:Uh,envmap_physical_pars_fragment:Xh,envmap_vertex:Oh,fog_vertex:Fh,fog_pars_vertex:zh,fog_fragment:Bh,fog_pars_fragment:kh,gradientmap_pars_fragment:Vh,lightmap_pars_fragment:Hh,lights_lambert_fragment:Gh,lights_lambert_pars_fragment:Wh,lights_pars_begin:jh,lights_toon_fragment:Yh,lights_toon_pars_fragment:qh,lights_phong_fragment:Kh,lights_phong_pars_fragment:$h,lights_physical_fragment:Zh,lights_physical_pars_fragment:Jh,lights_fragment_begin:Qh,lights_fragment_maps:eu,lights_fragment_end:tu,logdepthbuf_fragment:ru,logdepthbuf_pars_fragment:iu,logdepthbuf_pars_vertex:nu,logdepthbuf_vertex:au,map_fragment:su,map_pars_fragment:ou,map_particle_fragment:lu,map_particle_pars_fragment:cu,metalnessmap_fragment:hu,metalnessmap_pars_fragment:uu,morphinstance_vertex:du,morphcolor_vertex:pu,morphnormal_vertex:fu,morphtarget_pars_vertex:mu,morphtarget_vertex:gu,normal_fragment_begin:_u,normal_fragment_maps:vu,normal_pars_fragment:xu,normal_pars_vertex:yu,normal_vertex:Mu,normalmap_pars_fragment:Su,clearcoat_normal_fragment_begin:Eu,clearcoat_normal_fragment_maps:bu,clearcoat_pars_fragment:Tu,iridescence_pars_fragment:wu,opaque_fragment:Au,packing:Ru,premultiplied_alpha_fragment:Cu,project_vertex:Lu,dithering_fragment:Pu,dithering_pars_fragment:Du,roughnessmap_fragment:Iu,roughnessmap_pars_fragment:Nu,shadowmap_pars_fragment:Uu,shadowmap_pars_vertex:Ou,shadowmap_vertex:Fu,shadowmask_pars_fragment:zu,skinbase_vertex:Bu,skinning_pars_vertex:ku,skinning_vertex:Vu,skinnormal_vertex:Hu,specularmap_fragment:Gu,specularmap_pars_fragment:Wu,tonemapping_fragment:ju,tonemapping_pars_fragment:Xu,transmission_fragment:Yu,transmission_pars_fragment:qu,uv_pars_fragment:Ku,uv_pars_vertex:$u,uv_vertex:Zu,worldpos_vertex:Ju,background_vert:Qu,background_frag:ed,backgroundCube_vert:td,backgroundCube_frag:rd,cube_vert:id,cube_frag:nd,depth_vert:ad,depth_frag:sd,distanceRGBA_vert:od,distanceRGBA_frag:ld,equirect_vert:cd,equirect_frag:hd,linedashed_vert:ud,linedashed_frag:dd,meshbasic_vert:pd,meshbasic_frag:fd,meshlambert_vert:md,meshlambert_frag:gd,meshmatcap_vert:_d,meshmatcap_frag:vd,meshnormal_vert:xd,meshnormal_frag:yd,meshphong_vert:Md,meshphong_frag:Sd,meshphysical_vert:Ed,meshphysical_frag:bd,meshtoon_vert:Td,meshtoon_frag:wd,points_vert:Ad,points_frag:Rd,shadow_vert:Cd,shadow_frag:Ld,sprite_vert:Pd,sprite_frag:Dd},fe={common:{diffuse:{value:new We(16777215)},opacity:{value:1},map:{value:null},mapTransform:{value:new Ve},alphaMap:{value:null},alphaMapTransform:{value:new Ve},alphaTest:{value:0}},specularmap:{specularMap:{value:null},specularMapTransform:{value:new Ve}},envmap:{envMap:{value:null},envMapRotation:{value:new Ve},flipEnvMap:{value:-1},reflectivity:{value:1},ior:{value:1.5},refractionRatio:{value:.98}},aomap:{aoMap:{value:null},aoMapIntensity:{value:1},aoMapTransform:{value:new Ve}},lightmap:{lightMap:{value:null},lightMapIntensity:{value:1},lightMapTransform:{value:new Ve}},bumpmap:{bumpMap:{value:null},bumpMapTransform:{value:new Ve},bumpScale:{value:1}},normalmap:{normalMap:{value:null},normalMapTransform:{value:new Ve},normalScale:{value:new je(1,1)}},displacementmap:{displacementMap:{value:null},displacementMapTransform:{value:new Ve},displacementScale:{value:1},displacementBias:{value:0}},emissivemap:{emissiveMap:{value:null},emissiveMapTransform:{value:new Ve}},metalnessmap:{metalnessMap:{value:null},metalnessMapTransform:{value:new Ve}},roughnessmap:{roughnessMap:{value:null},roughnessMapTransform:{value:new Ve}},gradientmap:{gradientMap:{value:null}},fog:{fogDensity:{value:25e-5},fogNear:{value:1},fogFar:{value:2e3},fogColor:{value:new We(16777215)}},lights:{ambientLightColor:{value:[]},lightProbe:{value:[]},directionalLights:{value:[],properties:{direction:{},color:{}}},directionalLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{}}},directionalShadowMap:{value:[]},directionalShadowMatrix:{value:[]},spotLights:{value:[],properties:{color:{},position:{},direction:{},distance:{},coneCos:{},penumbraCos:{},decay:{}}},spotLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{}}},spotLightMap:{value:[]},spotShadowMap:{value:[]},spotLightMatrix:{value:[]},pointLights:{value:[],properties:{color:{},position:{},decay:{},distance:{}}},pointLightShadows:{value:[],properties:{shadowIntensity:1,shadowBias:{},shadowNormalBias:{},shadowRadius:{},shadowMapSize:{},shadowCameraNear:{},shadowCameraFar:{}}},pointShadowMap:{value:[]},pointShadowMatrix:{value:[]},hemisphereLights:{value:[],properties:{direction:{},skyColor:{},groundColor:{}}},rectAreaLights:{value:[],properties:{color:{},position:{},width:{},height:{}}},ltc_1:{value:null},ltc_2:{value:null}},points:{diffuse:{value:new We(16777215)},opacity:{value:1},size:{value:1},scale:{value:1},map:{value:null},alphaMap:{value:null},alphaMapTransform:{value:new Ve},alphaTest:{value:0},uvTransform:{value:new Ve}},sprite:{diffuse:{value:new We(16777215)},opacity:{value:1},center:{value:new je(.5,.5)},rotation:{value:0},map:{value:null},mapTransform:{value:new Ve},alphaMap:{value:null},alphaMapTransform:{value:new Ve},alphaTest:{value:0}}},Zt={basic:{uniforms:At([fe.common,fe.specularmap,fe.envmap,fe.aomap,fe.lightmap,fe.fog]),vertexShader:Ge.meshbasic_vert,fragmentShader:Ge.meshbasic_frag},lambert:{uniforms:At([fe.common,fe.specularmap,fe.envmap,fe.aomap,fe.lightmap,fe.emissivemap,fe.bumpmap,fe.normalmap,fe.displacementmap,fe.fog,fe.lights,{emissive:{value:new We(0)}}]),vertexShader:Ge.meshlambert_vert,fragmentShader:Ge.meshlambert_frag},phong:{uniforms:At([fe.common,fe.specularmap,fe.envmap,fe.aomap,fe.lightmap,fe.emissivemap,fe.bumpmap,fe.normalmap,fe.displacementmap,fe.fog,fe.lights,{emissive:{value:new We(0)},specular:{value:new We(1118481)},shininess:{value:30}}]),vertexShader:Ge.meshphong_vert,fragmentShader:Ge.meshphong_frag},standard:{uniforms:At([fe.common,fe.envmap,fe.aomap,fe.lightmap,fe.emissivemap,fe.bumpmap,fe.normalmap,fe.displacementmap,fe.roughnessmap,fe.metalnessmap,fe.fog,fe.lights,{emissive:{value:new We(0)},roughness:{value:1},metalness:{value:0},envMapIntensity:{value:1}}]),vertexShader:Ge.meshphysical_vert,fragmentShader:Ge.meshphysical_frag},toon:{uniforms:At([fe.common,fe.aomap,fe.lightmap,fe.emissivemap,fe.bumpmap,fe.normalmap,fe.displacementmap,fe.gradientmap,fe.fog,fe.lights,{emissive:{value:new We(0)}}]),vertexShader:Ge.meshtoon_vert,fragmentShader:Ge.meshtoon_frag},matcap:{uniforms:At([fe.common,fe.bumpmap,fe.normalmap,fe.displacementmap,fe.fog,{matcap:{value:null}}]),vertexShader:Ge.meshmatcap_vert,fragmentShader:Ge.meshmatcap_frag},points:{uniforms:At([fe.points,fe.fog]),vertexShader:Ge.points_vert,fragmentShader:Ge.points_frag},dashed:{uniforms:At([fe.common,fe.fog,{scale:{value:1},dashSize:{value:1},totalSize:{value:2}}]),vertexShader:Ge.linedashed_vert,fragmentShader:Ge.linedashed_frag},depth:{uniforms:At([fe.common,fe.displacementmap]),vertexShader:Ge.depth_vert,fragmentShader:Ge.depth_frag},normal:{uniforms:At([fe.common,fe.bumpmap,fe.normalmap,fe.displacementmap,{opacity:{value:1}}]),vertexShader:Ge.meshnormal_vert,fragmentShader:Ge.meshnormal_frag},sprite:{uniforms:At([fe.sprite,fe.fog]),vertexShader:Ge.sprite_vert,fragmentShader:Ge.sprite_frag},background:{uniforms:{uvTransform:{value:new Ve},t2D:{value:null},backgroundIntensity:{value:1}},vertexShader:Ge.background_vert,fragmentShader:Ge.background_frag},backgroundCube:{uniforms:{envMap:{value:null},flipEnvMap:{value:-1},backgroundBlurriness:{value:0},backgroundIntensity:{value:1},backgroundRotation:{value:new Ve}},vertexShader:Ge.backgroundCube_vert,fragmentShader:Ge.backgroundCube_frag},cube:{uniforms:{tCube:{value:null},tFlip:{value:-1},opacity:{value:1}},vertexShader:Ge.cube_vert,fragmentShader:Ge.cube_frag},equirect:{uniforms:{tEquirect:{value:null}},vertexShader:Ge.equirect_vert,fragmentShader:Ge.equirect_frag},distanceRGBA:{uniforms:At([fe.common,fe.displacementmap,{referencePosition:{value:new B},nearDistance:{value:1},farDistance:{value:1e3}}]),vertexShader:Ge.distanceRGBA_vert,fragmentShader:Ge.distanceRGBA_frag},shadow:{uniforms:At([fe.lights,fe.fog,{color:{value:new We(0)},opacity:{value:1}}]),vertexShader:Ge.shadow_vert,fragmentShader:Ge.shadow_frag}};Zt.physical={uniforms:At([Zt.standard.uniforms,{clearcoat:{value:0},clearcoatMap:{value:null},clearcoatMapTransform:{value:new Ve},clearcoatNormalMap:{value:null},clearcoatNormalMapTransform:{value:new Ve},clearcoatNormalScale:{value:new je(1,1)},clearcoatRoughness:{value:0},clearcoatRoughnessMap:{value:null},clearcoatRoughnessMapTransform:{value:new Ve},dispersion:{value:0},iridescence:{value:0},iridescenceMap:{value:null},iridescenceMapTransform:{value:new Ve},iridescenceIOR:{value:1.3},iridescenceThicknessMinimum:{value:100},iridescenceThicknessMaximum:{value:400},iridescenceThicknessMap:{value:null},iridescenceThicknessMapTransform:{value:new Ve},sheen:{value:0},sheenColor:{value:new We(0)},sheenColorMap:{value:null},sheenColorMapTransform:{value:new Ve},sheenRoughness:{value:1},sheenRoughnessMap:{value:null},sheenRoughnessMapTransform:{value:new Ve},transmission:{value:0},transmissionMap:{value:null},transmissionMapTransform:{value:new Ve},transmissionSamplerSize:{value:new je},transmissionSamplerMap:{value:null},thickness:{value:0},thicknessMap:{value:null},thicknessMapTransform:{value:new Ve},attenuationDistance:{value:0},attenuationColor:{value:new We(0)},specularColor:{value:new We(1,1,1)},specularColorMap:{value:null},specularColorMapTransform:{value:new Ve},specularIntensity:{value:1},specularIntensityMap:{value:null},specularIntensityMapTransform:{value:new Ve},anisotropyVector:{value:new je},anisotropyMap:{value:null},anisotropyMapTransform:{value:new Ve}}]),vertexShader:Ge.meshphysical_vert,fragmentShader:Ge.meshphysical_frag};const fn={r:0,b:0,g:0},Pr=new Yt,Id=new Qe;function Nd(i,e,t,r,n,a,o){const s=new We(0);let l=a===!0?0:1,c,h,u=null,p=0,f=null;function g(S){let T=S.isScene===!0?S.background:null;return T&&T.isTexture&&(T=(S.backgroundBlurriness>0?t:e).get(T)),T}function _(S){let T=!1;const x=g(S);x===null?d(s,l):x&&x.isColor&&(d(x,1),T=!0);const N=i.xr.getEnvironmentBlendMode();N==="additive"?r.buffers.color.setClear(0,0,0,1,o):N==="alpha-blend"&&r.buffers.color.setClear(0,0,0,0,o),(i.autoClear||T)&&(r.buffers.depth.setTest(!0),r.buffers.depth.setMask(!0),r.buffers.color.setMask(!0),i.clear(i.autoClearColor,i.autoClearDepth,i.autoClearStencil))}function m(S,T){const x=g(T);x&&(x.isCubeTexture||x.mapping===Un)?(h===void 0&&(h=new Fe(new tt(1,1,1),new Er({name:"BackgroundCubeMaterial",uniforms:xi(Zt.backgroundCube.uniforms),vertexShader:Zt.backgroundCube.vertexShader,fragmentShader:Zt.backgroundCube.fragmentShader,side:Dt,depthTest:!1,depthWrite:!1,fog:!1})),h.geometry.deleteAttribute("normal"),h.geometry.deleteAttribute("uv"),h.onBeforeRender=function(N,w,A){this.matrixWorld.copyPosition(A.matrixWorld)},Object.defineProperty(h.material,"envMap",{get:function(){return this.uniforms.envMap.value}}),n.update(h)),Pr.copy(T.backgroundRotation),Pr.x*=-1,Pr.y*=-1,Pr.z*=-1,x.isCubeTexture&&x.isRenderTargetTexture===!1&&(Pr.y*=-1,Pr.z*=-1),h.material.uniforms.envMap.value=x,h.material.uniforms.flipEnvMap.value=x.isCubeTexture&&x.isRenderTargetTexture===!1?-1:1,h.material.uniforms.backgroundBlurriness.value=T.backgroundBlurriness,h.material.uniforms.backgroundIntensity.value=T.backgroundIntensity,h.material.uniforms.backgroundRotation.value.setFromMatrix4(Id.makeRotationFromEuler(Pr)),h.material.toneMapped=$e.getTransfer(x.colorSpace)!==rt,(u!==x||p!==x.version||f!==i.toneMapping)&&(h.material.needsUpdate=!0,u=x,p=x.version,f=i.toneMapping),h.layers.enableAll(),S.unshift(h,h.geometry,h.material,0,0,null)):x&&x.isTexture&&(c===void 0&&(c=new Fe(new vr(2,2),new Er({name:"BackgroundMaterial",uniforms:xi(Zt.background.uniforms),vertexShader:Zt.background.vertexShader,fragmentShader:Zt.background.fragmentShader,side:Sr,depthTest:!1,depthWrite:!1,fog:!1})),c.geometry.deleteAttribute("normal"),Object.defineProperty(c.material,"map",{get:function(){return this.uniforms.t2D.value}}),n.update(c)),c.material.uniforms.t2D.value=x,c.material.uniforms.backgroundIntensity.value=T.backgroundIntensity,c.material.toneMapped=$e.getTransfer(x.colorSpace)!==rt,x.matrixAutoUpdate===!0&&x.updateMatrix(),c.material.uniforms.uvTransform.value.copy(x.matrix),(u!==x||p!==x.version||f!==i.toneMapping)&&(c.material.needsUpdate=!0,u=x,p=x.version,f=i.toneMapping),c.layers.enableAll(),S.unshift(c,c.geometry,c.material,0,0,null))}function d(S,T){S.getRGB(fn,ml(i)),r.buffers.color.setClear(fn.r,fn.g,fn.b,T,o)}return{getClearColor:function(){return s},setClearColor:function(S,T=1){s.set(S),l=T,d(s,l)},getClearAlpha:function(){return l},setClearAlpha:function(S){l=S,d(s,l)},render:_,addToRenderList:m}}function Ud(i,e){const t=i.getParameter(i.MAX_VERTEX_ATTRIBS),r={},n=p(null);let a=n,o=!1;function s(M,R,H,U,F){let K=!1;const X=u(U,H,R);a!==X&&(a=X,c(a.object)),K=f(M,U,H,F),K&&g(M,U,H,F),F!==null&&e.update(F,i.ELEMENT_ARRAY_BUFFER),(K||o)&&(o=!1,x(M,R,H,U),F!==null&&i.bindBuffer(i.ELEMENT_ARRAY_BUFFER,e.get(F).buffer))}function l(){return i.createVertexArray()}function c(M){return i.bindVertexArray(M)}function h(M){return i.deleteVertexArray(M)}function u(M,R,H){const U=H.wireframe===!0;let F=r[M.id];F===void 0&&(F={},r[M.id]=F);let K=F[R.id];K===void 0&&(K={},F[R.id]=K);let X=K[U];return X===void 0&&(X=p(l()),K[U]=X),X}function p(M){const R=[],H=[],U=[];for(let F=0;F<t;F++)R[F]=0,H[F]=0,U[F]=0;return{geometry:null,program:null,wireframe:!1,newAttributes:R,enabledAttributes:H,attributeDivisors:U,object:M,attributes:{},index:null}}function f(M,R,H,U){const F=a.attributes,K=R.attributes;let X=0;const $=H.getAttributes();for(const k in $)if($[k].location>=0){const Z=F[k];let ce=K[k];if(ce===void 0&&(k==="instanceMatrix"&&M.instanceMatrix&&(ce=M.instanceMatrix),k==="instanceColor"&&M.instanceColor&&(ce=M.instanceColor)),Z===void 0||Z.attribute!==ce||ce&&Z.data!==ce.data)return!0;X++}return a.attributesNum!==X||a.index!==U}function g(M,R,H,U){const F={},K=R.attributes;let X=0;const $=H.getAttributes();for(const k in $)if($[k].location>=0){let Z=K[k];Z===void 0&&(k==="instanceMatrix"&&M.instanceMatrix&&(Z=M.instanceMatrix),k==="instanceColor"&&M.instanceColor&&(Z=M.instanceColor));const ce={};ce.attribute=Z,Z&&Z.data&&(ce.data=Z.data),F[k]=ce,X++}a.attributes=F,a.attributesNum=X,a.index=U}function _(){const M=a.newAttributes;for(let R=0,H=M.length;R<H;R++)M[R]=0}function m(M){d(M,0)}function d(M,R){const H=a.newAttributes,U=a.enabledAttributes,F=a.attributeDivisors;H[M]=1,U[M]===0&&(i.enableVertexAttribArray(M),U[M]=1),F[M]!==R&&(i.vertexAttribDivisor(M,R),F[M]=R)}function S(){const M=a.newAttributes,R=a.enabledAttributes;for(let H=0,U=R.length;H<U;H++)R[H]!==M[H]&&(i.disableVertexAttribArray(H),R[H]=0)}function T(M,R,H,U,F,K,X){X===!0?i.vertexAttribIPointer(M,R,H,F,K):i.vertexAttribPointer(M,R,H,U,F,K)}function x(M,R,H,U){_();const F=U.attributes,K=H.getAttributes(),X=R.defaultAttributeValues;for(const $ in K){const k=K[$];if(k.location>=0){let Z=F[$];if(Z===void 0&&($==="instanceMatrix"&&M.instanceMatrix&&(Z=M.instanceMatrix),$==="instanceColor"&&M.instanceColor&&(Z=M.instanceColor)),Z!==void 0){const ce=Z.normalized,ve=Z.itemSize,De=e.get(Z);if(De===void 0)continue;const xe=De.buffer,G=De.type,J=De.bytesPerElement,q=G===i.INT||G===i.UNSIGNED_INT||Z.gpuType===ps;if(Z.isInterleavedBufferAttribute){const z=Z.data,ae=z.stride,ne=Z.offset;if(z.isInstancedInterleavedBuffer){for(let oe=0;oe<k.locationSize;oe++)d(k.location+oe,z.meshPerAttribute);M.isInstancedMesh!==!0&&U._maxInstanceCount===void 0&&(U._maxInstanceCount=z.meshPerAttribute*z.count)}else for(let oe=0;oe<k.locationSize;oe++)m(k.location+oe);i.bindBuffer(i.ARRAY_BUFFER,xe);for(let oe=0;oe<k.locationSize;oe++)T(k.location+oe,ve/k.locationSize,G,ce,ae*J,(ne+ve/k.locationSize*oe)*J,q)}else{if(Z.isInstancedBufferAttribute){for(let z=0;z<k.locationSize;z++)d(k.location+z,Z.meshPerAttribute);M.isInstancedMesh!==!0&&U._maxInstanceCount===void 0&&(U._maxInstanceCount=Z.meshPerAttribute*Z.count)}else for(let z=0;z<k.locationSize;z++)m(k.location+z);i.bindBuffer(i.ARRAY_BUFFER,xe);for(let z=0;z<k.locationSize;z++)T(k.location+z,ve/k.locationSize,G,ce,ve*J,ve/k.locationSize*z*J,q)}}else if(X!==void 0){const ce=X[$];if(ce!==void 0)switch(ce.length){case 2:i.vertexAttrib2fv(k.location,ce);break;case 3:i.vertexAttrib3fv(k.location,ce);break;case 4:i.vertexAttrib4fv(k.location,ce);break;default:i.vertexAttrib1fv(k.location,ce)}}}}S()}function N(){O();for(const M in r){const R=r[M];for(const H in R){const U=R[H];for(const F in U)h(U[F].object),delete U[F];delete R[H]}delete r[M]}}function w(M){if(r[M.id]===void 0)return;const R=r[M.id];for(const H in R){const U=R[H];for(const F in U)h(U[F].object),delete U[F];delete R[H]}delete r[M.id]}function A(M){for(const R in r){const H=r[R];if(H[M.id]===void 0)continue;const U=H[M.id];for(const F in U)h(U[F].object),delete U[F];delete H[M.id]}}function O(){b(),o=!0,a!==n&&(a=n,c(a.object))}function b(){n.geometry=null,n.program=null,n.wireframe=!1}return{setup:s,reset:O,resetDefaultState:b,dispose:N,releaseStatesOfGeometry:w,releaseStatesOfProgram:A,initAttributes:_,enableAttribute:m,disableUnusedAttributes:S}}function Od(i,e,t){let r;function n(c){r=c}function a(c,h){i.drawArrays(r,c,h),t.update(h,r,1)}function o(c,h,u){u!==0&&(i.drawArraysInstanced(r,c,h,u),t.update(h,r,u))}function s(c,h,u){if(u===0)return;e.get("WEBGL_multi_draw").multiDrawArraysWEBGL(r,c,0,h,0,u);let p=0;for(let f=0;f<u;f++)p+=h[f];t.update(p,r,1)}function l(c,h,u,p){if(u===0)return;const f=e.get("WEBGL_multi_draw");if(f===null)for(let g=0;g<c.length;g++)o(c[g],h[g],p[g]);else{f.multiDrawArraysInstancedWEBGL(r,c,0,h,0,p,0,u);let g=0;for(let _=0;_<u;_++)g+=h[_]*p[_];t.update(g,r,1)}}this.setMode=n,this.render=a,this.renderInstances=o,this.renderMultiDraw=s,this.renderMultiDrawInstances=l}function Fd(i,e,t,r){let n;function a(){if(n!==void 0)return n;if(e.has("EXT_texture_filter_anisotropic")===!0){const A=e.get("EXT_texture_filter_anisotropic");n=i.getParameter(A.MAX_TEXTURE_MAX_ANISOTROPY_EXT)}else n=0;return n}function o(A){return!(A!==Xt&&r.convert(A)!==i.getParameter(i.IMPLEMENTATION_COLOR_READ_FORMAT))}function s(A){const O=A===Hi&&(e.has("EXT_color_buffer_half_float")||e.has("EXT_color_buffer_float"));return!(A!==cr&&r.convert(A)!==i.getParameter(i.IMPLEMENTATION_COLOR_READ_TYPE)&&A!==Qt&&!O)}function l(A){if(A==="highp"){if(i.getShaderPrecisionFormat(i.VERTEX_SHADER,i.HIGH_FLOAT).precision>0&&i.getShaderPrecisionFormat(i.FRAGMENT_SHADER,i.HIGH_FLOAT).precision>0)return"highp";A="mediump"}return A==="mediump"&&i.getShaderPrecisionFormat(i.VERTEX_SHADER,i.MEDIUM_FLOAT).precision>0&&i.getShaderPrecisionFormat(i.FRAGMENT_SHADER,i.MEDIUM_FLOAT).precision>0?"mediump":"lowp"}let c=t.precision!==void 0?t.precision:"highp";const h=l(c);h!==c&&(console.warn("THREE.WebGLRenderer:",c,"not supported, using",h,"instead."),c=h);const u=t.logarithmicDepthBuffer===!0,p=t.reverseDepthBuffer===!0&&e.has("EXT_clip_control"),f=i.getParameter(i.MAX_TEXTURE_IMAGE_UNITS),g=i.getParameter(i.MAX_VERTEX_TEXTURE_IMAGE_UNITS),_=i.getParameter(i.MAX_TEXTURE_SIZE),m=i.getParameter(i.MAX_CUBE_MAP_TEXTURE_SIZE),d=i.getParameter(i.MAX_VERTEX_ATTRIBS),S=i.getParameter(i.MAX_VERTEX_UNIFORM_VECTORS),T=i.getParameter(i.MAX_VARYING_VECTORS),x=i.getParameter(i.MAX_FRAGMENT_UNIFORM_VECTORS),N=g>0,w=i.getParameter(i.MAX_SAMPLES);return{isWebGL2:!0,getMaxAnisotropy:a,getMaxPrecision:l,textureFormatReadable:o,textureTypeReadable:s,precision:c,logarithmicDepthBuffer:u,reverseDepthBuffer:p,maxTextures:f,maxVertexTextures:g,maxTextureSize:_,maxCubemapSize:m,maxAttributes:d,maxVertexUniforms:S,maxVaryings:T,maxFragmentUniforms:x,vertexTextures:N,maxSamples:w}}function zd(i){const e=this;let t=null,r=0,n=!1,a=!1;const o=new Nr,s=new Ve,l={value:null,needsUpdate:!1};this.uniform=l,this.numPlanes=0,this.numIntersection=0,this.init=function(u,p){const f=u.length!==0||p||r!==0||n;return n=p,r=u.length,f},this.beginShadows=function(){a=!0,h(null)},this.endShadows=function(){a=!1},this.setGlobalState=function(u,p){t=h(u,p,0)},this.setState=function(u,p,f){const g=u.clippingPlanes,_=u.clipIntersection,m=u.clipShadows,d=i.get(u);if(!n||g===null||g.length===0||a&&!m)a?h(null):c();else{const S=a?0:r,T=S*4;let x=d.clippingState||null;l.value=x,x=h(g,p,T,f);for(let N=0;N!==T;++N)x[N]=t[N];d.clippingState=x,this.numIntersection=_?this.numPlanes:0,this.numPlanes+=S}};function c(){l.value!==t&&(l.value=t,l.needsUpdate=r>0),e.numPlanes=r,e.numIntersection=0}function h(u,p,f,g){const _=u!==null?u.length:0;let m=null;if(_!==0){if(m=l.value,g!==!0||m===null){const d=f+_*4,S=p.matrixWorldInverse;s.getNormalMatrix(S),(m===null||m.length<d)&&(m=new Float32Array(d));for(let T=0,x=f;T!==_;++T,x+=4)o.copy(u[T]).applyMatrix4(S,s),o.normal.toArray(m,x),m[x+3]=o.constant}l.value=m,l.needsUpdate=!0}return e.numPlanes=_,e.numIntersection=0,m}}function Bd(i){let e=new WeakMap;function t(o,s){return s===Ca?o.mapping=mi:s===La&&(o.mapping=gi),o}function r(o){if(o&&o.isTexture){const s=o.mapping;if(s===Ca||s===La)if(e.has(o)){const l=e.get(o).texture;return t(l,o.mapping)}else{const l=o.image;if(l&&l.height>0){const c=new $c(l.height);return c.fromEquirectangularTexture(i,o),e.set(o,c),o.addEventListener("dispose",n),t(c.texture,o.mapping)}else return null}}return o}function n(o){const s=o.target;s.removeEventListener("dispose",n);const l=e.get(s);l!==void 0&&(e.delete(s),l.dispose())}function a(){e=new WeakMap}return{get:r,dispose:a}}class xl extends gl{constructor(e=-1,t=1,r=1,n=-1,a=.1,o=2e3){super(),this.isOrthographicCamera=!0,this.type="OrthographicCamera",this.zoom=1,this.view=null,this.left=e,this.right=t,this.top=r,this.bottom=n,this.near=a,this.far=o,this.updateProjectionMatrix()}copy(e,t){return super.copy(e,t),this.left=e.left,this.right=e.right,this.top=e.top,this.bottom=e.bottom,this.near=e.near,this.far=e.far,this.zoom=e.zoom,this.view=e.view===null?null:Object.assign({},e.view),this}setViewOffset(e,t,r,n,a,o){this.view===null&&(this.view={enabled:!0,fullWidth:1,fullHeight:1,offsetX:0,offsetY:0,width:1,height:1}),this.view.enabled=!0,this.view.fullWidth=e,this.view.fullHeight=t,this.view.offsetX=r,this.view.offsetY=n,this.view.width=a,this.view.height=o,this.updateProjectionMatrix()}clearViewOffset(){this.view!==null&&(this.view.enabled=!1),this.updateProjectionMatrix()}updateProjectionMatrix(){const e=(this.right-this.left)/(2*this.zoom),t=(this.top-this.bottom)/(2*this.zoom),r=(this.right+this.left)/2,n=(this.top+this.bottom)/2;let a=r-e,o=r+e,s=n+t,l=n-t;if(this.view!==null&&this.view.enabled){const c=(this.right-this.left)/this.view.fullWidth/this.zoom,h=(this.top-this.bottom)/this.view.fullHeight/this.zoom;a+=c*this.view.offsetX,o=a+c*this.view.width,s-=h*this.view.offsetY,l=s-h*this.view.height}this.projectionMatrix.makeOrthographic(a,o,s,l,this.near,this.far,this.coordinateSystem),this.projectionMatrixInverse.copy(this.projectionMatrix).invert()}toJSON(e){const t=super.toJSON(e);return t.object.zoom=this.zoom,t.object.left=this.left,t.object.right=this.right,t.object.top=this.top,t.object.bottom=this.bottom,t.object.near=this.near,t.object.far=this.far,this.view!==null&&(t.object.view=Object.assign({},this.view)),t}}const hi=4,oo=[.125,.215,.35,.446,.526,.582],Fr=20,ca=new xl,lo=new We;let ha=null,ua=0,da=0,pa=!1;const Ur=(1+Math.sqrt(5))/2,ni=1/Ur,co=[new B(-Ur,ni,0),new B(Ur,ni,0),new B(-ni,0,Ur),new B(ni,0,Ur),new B(0,Ur,-ni),new B(0,Ur,ni),new B(-1,1,-1),new B(1,1,-1),new B(-1,1,1),new B(1,1,1)];class ho{constructor(e){this._renderer=e,this._pingPongRenderTarget=null,this._lodMax=0,this._cubeSize=0,this._lodPlanes=[],this._sizeLods=[],this._sigmas=[],this._blurMaterial=null,this._cubemapMaterial=null,this._equirectMaterial=null,this._compileMaterial(this._blurMaterial)}fromScene(e,t=0,r=.1,n=100){ha=this._renderer.getRenderTarget(),ua=this._renderer.getActiveCubeFace(),da=this._renderer.getActiveMipmapLevel(),pa=this._renderer.xr.enabled,this._renderer.xr.enabled=!1,this._setSize(256);const a=this._allocateTargets();return a.depthBuffer=!0,this._sceneToCubeUV(e,r,n,a),t>0&&this._blur(a,0,0,t),this._applyPMREM(a),this._cleanup(a),a}fromEquirectangular(e,t=null){return this._fromTexture(e,t)}fromCubemap(e,t=null){return this._fromTexture(e,t)}compileCubemapShader(){this._cubemapMaterial===null&&(this._cubemapMaterial=fo(),this._compileMaterial(this._cubemapMaterial))}compileEquirectangularShader(){this._equirectMaterial===null&&(this._equirectMaterial=po(),this._compileMaterial(this._equirectMaterial))}dispose(){this._dispose(),this._cubemapMaterial!==null&&this._cubemapMaterial.dispose(),this._equirectMaterial!==null&&this._equirectMaterial.dispose()}_setSize(e){this._lodMax=Math.floor(Math.log2(e)),this._cubeSize=Math.pow(2,this._lodMax)}_dispose(){this._blurMaterial!==null&&this._blurMaterial.dispose(),this._pingPongRenderTarget!==null&&this._pingPongRenderTarget.dispose();for(let e=0;e<this._lodPlanes.length;e++)this._lodPlanes[e].dispose()}_cleanup(e){this._renderer.setRenderTarget(ha,ua,da),this._renderer.xr.enabled=pa,e.scissorTest=!1,mn(e,0,0,e.width,e.height)}_fromTexture(e,t){e.mapping===mi||e.mapping===gi?this._setSize(e.image.length===0?16:e.image[0].width||e.image[0].image.width):this._setSize(e.image.width/4),ha=this._renderer.getRenderTarget(),ua=this._renderer.getActiveCubeFace(),da=this._renderer.getActiveMipmapLevel(),pa=this._renderer.xr.enabled,this._renderer.xr.enabled=!1;const r=t||this._allocateTargets();return this._textureToCubeUV(e,r),this._applyPMREM(r),this._cleanup(r),r}_allocateTargets(){const e=3*Math.max(this._cubeSize,112),t=4*this._cubeSize,r={magFilter:jt,minFilter:jt,generateMipmaps:!1,type:Hi,format:Xt,colorSpace:yi,depthBuffer:!1},n=uo(e,t,r);if(this._pingPongRenderTarget===null||this._pingPongRenderTarget.width!==e||this._pingPongRenderTarget.height!==t){this._pingPongRenderTarget!==null&&this._dispose(),this._pingPongRenderTarget=uo(e,t,r);const{_lodMax:a}=this;({sizeLods:this._sizeLods,lodPlanes:this._lodPlanes,sigmas:this._sigmas}=kd(a)),this._blurMaterial=Vd(a,e,t)}return n}_compileMaterial(e){const t=new Fe(this._lodPlanes[0],e);this._renderer.compile(t,ca)}_sceneToCubeUV(e,t,r,n){const a=new Ft(90,1,t,r),o=[1,-1,1,1,1,1],s=[1,1,1,-1,-1,-1],l=this._renderer,c=l.autoClear,h=l.toneMapping;l.getClearColor(lo),l.toneMapping=yr,l.autoClear=!1;const u=new Bi({name:"PMREM.Background",side:Dt,depthWrite:!1,depthTest:!1}),p=new Fe(new tt,u);let f=!1;const g=e.background;g?g.isColor&&(u.color.copy(g),e.background=null,f=!0):(u.color.copy(lo),f=!0);for(let _=0;_<6;_++){const m=_%3;m===0?(a.up.set(0,o[_],0),a.lookAt(s[_],0,0)):m===1?(a.up.set(0,0,o[_]),a.lookAt(0,s[_],0)):(a.up.set(0,o[_],0),a.lookAt(0,0,s[_]));const d=this._cubeSize;mn(n,m*d,_>2?d:0,d,d),l.setRenderTarget(n),f&&l.render(p,a),l.render(e,a)}p.geometry.dispose(),p.material.dispose(),l.toneMapping=h,l.autoClear=c,e.background=g}_textureToCubeUV(e,t){const r=this._renderer,n=e.mapping===mi||e.mapping===gi;n?(this._cubemapMaterial===null&&(this._cubemapMaterial=fo()),this._cubemapMaterial.uniforms.flipEnvMap.value=e.isRenderTargetTexture===!1?-1:1):this._equirectMaterial===null&&(this._equirectMaterial=po());const a=n?this._cubemapMaterial:this._equirectMaterial,o=new Fe(this._lodPlanes[0],a),s=a.uniforms;s.envMap.value=e;const l=this._cubeSize;mn(t,0,0,3*l,2*l),r.setRenderTarget(t),r.render(o,ca)}_applyPMREM(e){const t=this._renderer,r=t.autoClear;t.autoClear=!1;const n=this._lodPlanes.length;for(let a=1;a<n;a++){const o=Math.sqrt(this._sigmas[a]*this._sigmas[a]-this._sigmas[a-1]*this._sigmas[a-1]),s=co[(n-a-1)%co.length];this._blur(e,a-1,a,o,s)}t.autoClear=r}_blur(e,t,r,n,a){const o=this._pingPongRenderTarget;this._halfBlur(e,o,t,r,n,"latitudinal",a),this._halfBlur(o,e,r,r,n,"longitudinal",a)}_halfBlur(e,t,r,n,a,o,s){const l=this._renderer,c=this._blurMaterial;o!=="latitudinal"&&o!=="longitudinal"&&console.error("blur direction must be either latitudinal or longitudinal!");const h=3,u=new Fe(this._lodPlanes[n],c),p=c.uniforms,f=this._sizeLods[r]-1,g=isFinite(a)?Math.PI/(2*f):2*Math.PI/(2*Fr-1),_=a/g,m=isFinite(a)?1+Math.floor(h*_):Fr;m>Fr&&console.warn(`sigmaRadians, ${a}, is too large and will clip, as it requested ${m} samples when the maximum is set to ${Fr}`);const d=[];let S=0;for(let A=0;A<Fr;++A){const O=A/_,b=Math.exp(-O*O/2);d.push(b),A===0?S+=b:A<m&&(S+=2*b)}for(let A=0;A<d.length;A++)d[A]=d[A]/S;p.envMap.value=e.texture,p.samples.value=m,p.weights.value=d,p.latitudinal.value=o==="latitudinal",s&&(p.poleAxis.value=s);const{_lodMax:T}=this;p.dTheta.value=g,p.mipInt.value=T-r;const x=this._sizeLods[n],N=3*x*(n>T-hi?n-T+hi:0),w=4*(this._cubeSize-x);mn(t,N,w,3*x,2*x),l.setRenderTarget(t),l.render(u,ca)}}function kd(i){const e=[],t=[],r=[];let n=i;const a=i-hi+1+oo.length;for(let o=0;o<a;o++){const s=Math.pow(2,n);t.push(s);let l=1/s;o>i-hi?l=oo[o-i+hi-1]:o===0&&(l=0),r.push(l);const c=1/(s-2),h=-c,u=1+c,p=[h,h,u,h,u,u,h,h,u,u,h,u],f=6,g=6,_=3,m=2,d=1,S=new Float32Array(_*g*f),T=new Float32Array(m*g*f),x=new Float32Array(d*g*f);for(let w=0;w<f;w++){const A=w%3*2/3-1,O=w>2?0:-1,b=[A,O,0,A+2/3,O,0,A+2/3,O+1,0,A,O,0,A+2/3,O+1,0,A,O+1,0];S.set(b,_*g*w),T.set(p,m*g*w);const M=[w,w,w,w,w,w];x.set(M,d*g*w)}const N=new Ct;N.setAttribute("position",new Bt(S,_)),N.setAttribute("uv",new Bt(T,m)),N.setAttribute("faceIndex",new Bt(x,d)),e.push(N),n>hi&&n--}return{lodPlanes:e,sizeLods:t,sigmas:r}}function uo(i,e,t){const r=new Vr(i,e,t);return r.texture.mapping=Un,r.texture.name="PMREM.cubeUv",r.scissorTest=!0,r}function mn(i,e,t,r,n){i.viewport.set(e,t,r,n),i.scissor.set(e,t,r,n)}function Vd(i,e,t){const r=new Float32Array(Fr),n=new B(0,1,0);return new Er({name:"SphericalGaussianBlur",defines:{n:Fr,CUBEUV_TEXEL_WIDTH:1/e,CUBEUV_TEXEL_HEIGHT:1/t,CUBEUV_MAX_MIP:`${i}.0`},uniforms:{envMap:{value:null},samples:{value:1},weights:{value:r},latitudinal:{value:!1},dTheta:{value:0},mipInt:{value:0},poleAxis:{value:n}},vertexShader:Ms(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;
			uniform int samples;
			uniform float weights[ n ];
			uniform bool latitudinal;
			uniform float dTheta;
			uniform float mipInt;
			uniform vec3 poleAxis;

			#define ENVMAP_TYPE_CUBE_UV
			#include <cube_uv_reflection_fragment>

			vec3 getSample( float theta, vec3 axis ) {

				float cosTheta = cos( theta );
				// Rodrigues' axis-angle rotation
				vec3 sampleDirection = vOutputDirection * cosTheta
					+ cross( axis, vOutputDirection ) * sin( theta )
					+ axis * dot( axis, vOutputDirection ) * ( 1.0 - cosTheta );

				return bilinearCubeUV( envMap, sampleDirection, mipInt );

			}

			void main() {

				vec3 axis = latitudinal ? poleAxis : cross( poleAxis, vOutputDirection );

				if ( all( equal( axis, vec3( 0.0 ) ) ) ) {

					axis = vec3( vOutputDirection.z, 0.0, - vOutputDirection.x );

				}

				axis = normalize( axis );

				gl_FragColor = vec4( 0.0, 0.0, 0.0, 1.0 );
				gl_FragColor.rgb += weights[ 0 ] * getSample( 0.0, axis );

				for ( int i = 1; i < n; i++ ) {

					if ( i >= samples ) {

						break;

					}

					float theta = dTheta * float( i );
					gl_FragColor.rgb += weights[ i ] * getSample( -1.0 * theta, axis );
					gl_FragColor.rgb += weights[ i ] * getSample( theta, axis );

				}

			}
		`,blending:xr,depthTest:!1,depthWrite:!1})}function po(){return new Er({name:"EquirectangularToCubeUV",uniforms:{envMap:{value:null}},vertexShader:Ms(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			varying vec3 vOutputDirection;

			uniform sampler2D envMap;

			#include <common>

			void main() {

				vec3 outputDirection = normalize( vOutputDirection );
				vec2 uv = equirectUv( outputDirection );

				gl_FragColor = vec4( texture2D ( envMap, uv ).rgb, 1.0 );

			}
		`,blending:xr,depthTest:!1,depthWrite:!1})}function fo(){return new Er({name:"CubemapToCubeUV",uniforms:{envMap:{value:null},flipEnvMap:{value:-1}},vertexShader:Ms(),fragmentShader:`

			precision mediump float;
			precision mediump int;

			uniform float flipEnvMap;

			varying vec3 vOutputDirection;

			uniform samplerCube envMap;

			void main() {

				gl_FragColor = textureCube( envMap, vec3( flipEnvMap * vOutputDirection.x, vOutputDirection.yz ) );

			}
		`,blending:xr,depthTest:!1,depthWrite:!1})}function Ms(){return`

		precision mediump float;
		precision mediump int;

		attribute float faceIndex;

		varying vec3 vOutputDirection;

		// RH coordinate system; PMREM face-indexing convention
		vec3 getDirection( vec2 uv, float face ) {

			uv = 2.0 * uv - 1.0;

			vec3 direction = vec3( uv, 1.0 );

			if ( face == 0.0 ) {

				direction = direction.zyx; // ( 1, v, u ) pos x

			} else if ( face == 1.0 ) {

				direction = direction.xzy;
				direction.xz *= -1.0; // ( -u, 1, -v ) pos y

			} else if ( face == 2.0 ) {

				direction.x *= -1.0; // ( -u, v, 1 ) pos z

			} else if ( face == 3.0 ) {

				direction = direction.zyx;
				direction.xz *= -1.0; // ( -1, v, -u ) neg x

			} else if ( face == 4.0 ) {

				direction = direction.xzy;
				direction.xy *= -1.0; // ( -u, -1, v ) neg y

			} else if ( face == 5.0 ) {

				direction.z *= -1.0; // ( u, v, -1 ) neg z

			}

			return direction;

		}

		void main() {

			vOutputDirection = getDirection( uv, faceIndex );
			gl_Position = vec4( position, 1.0 );

		}
	`}function Hd(i){let e=new WeakMap,t=null;function r(s){if(s&&s.isTexture){const l=s.mapping,c=l===Ca||l===La,h=l===mi||l===gi;if(c||h){let u=e.get(s);const p=u!==void 0?u.texture.pmremVersion:0;if(s.isRenderTargetTexture&&s.pmremVersion!==p)return t===null&&(t=new ho(i)),u=c?t.fromEquirectangular(s,u):t.fromCubemap(s,u),u.texture.pmremVersion=s.pmremVersion,e.set(s,u),u.texture;if(u!==void 0)return u.texture;{const f=s.image;return c&&f&&f.height>0||h&&f&&n(f)?(t===null&&(t=new ho(i)),u=c?t.fromEquirectangular(s):t.fromCubemap(s),u.texture.pmremVersion=s.pmremVersion,e.set(s,u),s.addEventListener("dispose",a),u.texture):null}}}return s}function n(s){let l=0;const c=6;for(let h=0;h<c;h++)s[h]!==void 0&&l++;return l===c}function a(s){const l=s.target;l.removeEventListener("dispose",a);const c=e.get(l);c!==void 0&&(e.delete(l),c.dispose())}function o(){e=new WeakMap,t!==null&&(t.dispose(),t=null)}return{get:r,dispose:o}}function Gd(i){const e={};function t(r){if(e[r]!==void 0)return e[r];let n;switch(r){case"WEBGL_depth_texture":n=i.getExtension("WEBGL_depth_texture")||i.getExtension("MOZ_WEBGL_depth_texture")||i.getExtension("WEBKIT_WEBGL_depth_texture");break;case"EXT_texture_filter_anisotropic":n=i.getExtension("EXT_texture_filter_anisotropic")||i.getExtension("MOZ_EXT_texture_filter_anisotropic")||i.getExtension("WEBKIT_EXT_texture_filter_anisotropic");break;case"WEBGL_compressed_texture_s3tc":n=i.getExtension("WEBGL_compressed_texture_s3tc")||i.getExtension("MOZ_WEBGL_compressed_texture_s3tc")||i.getExtension("WEBKIT_WEBGL_compressed_texture_s3tc");break;case"WEBGL_compressed_texture_pvrtc":n=i.getExtension("WEBGL_compressed_texture_pvrtc")||i.getExtension("WEBKIT_WEBGL_compressed_texture_pvrtc");break;default:n=i.getExtension(r)}return e[r]=n,n}return{has:function(r){return t(r)!==null},init:function(){t("EXT_color_buffer_float"),t("WEBGL_clip_cull_distance"),t("OES_texture_float_linear"),t("EXT_color_buffer_half_float"),t("WEBGL_multisampled_render_to_texture"),t("WEBGL_render_shared_exponent")},get:function(r){const n=t(r);return n===null&&Fi("THREE.WebGLRenderer: "+r+" extension not supported."),n}}}function Wd(i,e,t,r){const n={},a=new WeakMap;function o(u){const p=u.target;p.index!==null&&e.remove(p.index);for(const g in p.attributes)e.remove(p.attributes[g]);for(const g in p.morphAttributes){const _=p.morphAttributes[g];for(let m=0,d=_.length;m<d;m++)e.remove(_[m])}p.removeEventListener("dispose",o),delete n[p.id];const f=a.get(p);f&&(e.remove(f),a.delete(p)),r.releaseStatesOfGeometry(p),p.isInstancedBufferGeometry===!0&&delete p._maxInstanceCount,t.memory.geometries--}function s(u,p){return n[p.id]===!0||(p.addEventListener("dispose",o),n[p.id]=!0,t.memory.geometries++),p}function l(u){const p=u.attributes;for(const g in p)e.update(p[g],i.ARRAY_BUFFER);const f=u.morphAttributes;for(const g in f){const _=f[g];for(let m=0,d=_.length;m<d;m++)e.update(_[m],i.ARRAY_BUFFER)}}function c(u){const p=[],f=u.index,g=u.attributes.position;let _=0;if(f!==null){const S=f.array;_=f.version;for(let T=0,x=S.length;T<x;T+=3){const N=S[T+0],w=S[T+1],A=S[T+2];p.push(N,w,w,A,A,N)}}else if(g!==void 0){const S=g.array;_=g.version;for(let T=0,x=S.length/3-1;T<x;T+=3){const N=T+0,w=T+1,A=T+2;p.push(N,w,w,A,A,N)}}else return;const m=new(ol(p)?fl:pl)(p,1);m.version=_;const d=a.get(u);d&&e.remove(d),a.set(u,m)}function h(u){const p=a.get(u);if(p){const f=u.index;f!==null&&p.version<f.version&&c(u)}else c(u);return a.get(u)}return{get:s,update:l,getWireframeAttribute:h}}function jd(i,e,t){let r;function n(p){r=p}let a,o;function s(p){a=p.type,o=p.bytesPerElement}function l(p,f){i.drawElements(r,f,a,p*o),t.update(f,r,1)}function c(p,f,g){g!==0&&(i.drawElementsInstanced(r,f,a,p*o,g),t.update(f,r,g))}function h(p,f,g){if(g===0)return;e.get("WEBGL_multi_draw").multiDrawElementsWEBGL(r,f,0,a,p,0,g);let _=0;for(let m=0;m<g;m++)_+=f[m];t.update(_,r,1)}function u(p,f,g,_){if(g===0)return;const m=e.get("WEBGL_multi_draw");if(m===null)for(let d=0;d<p.length;d++)c(p[d]/o,f[d],_[d]);else{m.multiDrawElementsInstancedWEBGL(r,f,0,a,p,0,_,0,g);let d=0;for(let S=0;S<g;S++)d+=f[S]*_[S];t.update(d,r,1)}}this.setMode=n,this.setIndex=s,this.render=l,this.renderInstances=c,this.renderMultiDraw=h,this.renderMultiDrawInstances=u}function Xd(i){const e={geometries:0,textures:0},t={frame:0,calls:0,triangles:0,points:0,lines:0};function r(a,o,s){switch(t.calls++,o){case i.TRIANGLES:t.triangles+=s*(a/3);break;case i.LINES:t.lines+=s*(a/2);break;case i.LINE_STRIP:t.lines+=s*(a-1);break;case i.LINE_LOOP:t.lines+=s*a;break;case i.POINTS:t.points+=s*a;break;default:console.error("THREE.WebGLInfo: Unknown draw mode:",o);break}}function n(){t.calls=0,t.triangles=0,t.points=0,t.lines=0}return{memory:e,render:t,programs:null,autoReset:!0,reset:n,update:r}}function Yd(i,e,t){const r=new WeakMap,n=new dt;function a(o,s,l){const c=o.morphTargetInfluences,h=s.morphAttributes.position||s.morphAttributes.normal||s.morphAttributes.color,u=h!==void 0?h.length:0;let p=r.get(s);if(p===void 0||p.count!==u){let f=function(){O.dispose(),r.delete(s),s.removeEventListener("dispose",f)};p!==void 0&&p.texture.dispose();const g=s.morphAttributes.position!==void 0,_=s.morphAttributes.normal!==void 0,m=s.morphAttributes.color!==void 0,d=s.morphAttributes.position||[],S=s.morphAttributes.normal||[],T=s.morphAttributes.color||[];let x=0;g===!0&&(x=1),_===!0&&(x=2),m===!0&&(x=3);let N=s.attributes.position.count*x,w=1;N>e.maxTextureSize&&(w=Math.ceil(N/e.maxTextureSize),N=e.maxTextureSize);const A=new Float32Array(N*w*4*u),O=new cl(A,N,w,u);O.type=Qt,O.needsUpdate=!0;const b=x*4;for(let M=0;M<u;M++){const R=d[M],H=S[M],U=T[M],F=N*w*4*M;for(let K=0;K<R.count;K++){const X=K*b;g===!0&&(n.fromBufferAttribute(R,K),A[F+X+0]=n.x,A[F+X+1]=n.y,A[F+X+2]=n.z,A[F+X+3]=0),_===!0&&(n.fromBufferAttribute(H,K),A[F+X+4]=n.x,A[F+X+5]=n.y,A[F+X+6]=n.z,A[F+X+7]=0),m===!0&&(n.fromBufferAttribute(U,K),A[F+X+8]=n.x,A[F+X+9]=n.y,A[F+X+10]=n.z,A[F+X+11]=U.itemSize===4?n.w:1)}}p={count:u,texture:O,size:new je(N,w)},r.set(s,p),s.addEventListener("dispose",f)}if(o.isInstancedMesh===!0&&o.morphTexture!==null)l.getUniforms().setValue(i,"morphTexture",o.morphTexture,t);else{let f=0;for(let _=0;_<c.length;_++)f+=c[_];const g=s.morphTargetsRelative?1:1-f;l.getUniforms().setValue(i,"morphTargetBaseInfluence",g),l.getUniforms().setValue(i,"morphTargetInfluences",c)}l.getUniforms().setValue(i,"morphTargetsTexture",p.texture,t),l.getUniforms().setValue(i,"morphTargetsTextureSize",p.size)}return{update:a}}function qd(i,e,t,r){let n=new WeakMap;function a(l){const c=r.render.frame,h=l.geometry,u=e.get(l,h);if(n.get(u)!==c&&(e.update(u),n.set(u,c)),l.isInstancedMesh&&(l.hasEventListener("dispose",s)===!1&&l.addEventListener("dispose",s),n.get(l)!==c&&(t.update(l.instanceMatrix,i.ARRAY_BUFFER),l.instanceColor!==null&&t.update(l.instanceColor,i.ARRAY_BUFFER),n.set(l,c))),l.isSkinnedMesh){const p=l.skeleton;n.get(p)!==c&&(p.update(),n.set(p,c))}return u}function o(){n=new WeakMap}function s(l){const c=l.target;c.removeEventListener("dispose",s),t.remove(c.instanceMatrix),c.instanceColor!==null&&t.remove(c.instanceColor)}return{update:a,dispose:o}}class yl extends _t{constructor(e,t,r,n,a,o,s,l,c,h=di){if(h!==di&&h!==vi)throw new Error("DepthTexture format must be either THREE.DepthFormat or THREE.DepthStencilFormat");r===void 0&&h===di&&(r=kr),r===void 0&&h===vi&&(r=_i),super(null,n,a,o,s,l,h,r,c),this.isDepthTexture=!0,this.image={width:e,height:t},this.magFilter=s!==void 0?s:pt,this.minFilter=l!==void 0?l:pt,this.flipY=!1,this.generateMipmaps=!1,this.compareFunction=null}copy(e){return super.copy(e),this.compareFunction=e.compareFunction,this}toJSON(e){const t=super.toJSON(e);return this.compareFunction!==null&&(t.compareFunction=this.compareFunction),t}}const Ml=new _t,mo=new yl(1,1),Sl=new cl,El=new Nc,bl=new _l,go=[],_o=[],vo=new Float32Array(16),xo=new Float32Array(9),yo=new Float32Array(4);function Ei(i,e,t){const r=i[0];if(r<=0||r>0)return i;const n=e*t;let a=go[n];if(a===void 0&&(a=new Float32Array(n),go[n]=a),e!==0){r.toArray(a,0);for(let o=1,s=0;o!==e;++o)s+=t,i[o].toArray(a,s)}return a}function vt(i,e){if(i.length!==e.length)return!1;for(let t=0,r=i.length;t<r;t++)if(i[t]!==e[t])return!1;return!0}function xt(i,e){for(let t=0,r=e.length;t<r;t++)i[t]=e[t]}function Fn(i,e){let t=_o[e];t===void 0&&(t=new Int32Array(e),_o[e]=t);for(let r=0;r!==e;++r)t[r]=i.allocateTextureUnit();return t}function Kd(i,e){const t=this.cache;t[0]!==e&&(i.uniform1f(this.addr,e),t[0]=e)}function $d(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y)&&(i.uniform2f(this.addr,e.x,e.y),t[0]=e.x,t[1]=e.y);else{if(vt(t,e))return;i.uniform2fv(this.addr,e),xt(t,e)}}function Zd(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z)&&(i.uniform3f(this.addr,e.x,e.y,e.z),t[0]=e.x,t[1]=e.y,t[2]=e.z);else if(e.r!==void 0)(t[0]!==e.r||t[1]!==e.g||t[2]!==e.b)&&(i.uniform3f(this.addr,e.r,e.g,e.b),t[0]=e.r,t[1]=e.g,t[2]=e.b);else{if(vt(t,e))return;i.uniform3fv(this.addr,e),xt(t,e)}}function Jd(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z||t[3]!==e.w)&&(i.uniform4f(this.addr,e.x,e.y,e.z,e.w),t[0]=e.x,t[1]=e.y,t[2]=e.z,t[3]=e.w);else{if(vt(t,e))return;i.uniform4fv(this.addr,e),xt(t,e)}}function Qd(i,e){const t=this.cache,r=e.elements;if(r===void 0){if(vt(t,e))return;i.uniformMatrix2fv(this.addr,!1,e),xt(t,e)}else{if(vt(t,r))return;yo.set(r),i.uniformMatrix2fv(this.addr,!1,yo),xt(t,r)}}function ep(i,e){const t=this.cache,r=e.elements;if(r===void 0){if(vt(t,e))return;i.uniformMatrix3fv(this.addr,!1,e),xt(t,e)}else{if(vt(t,r))return;xo.set(r),i.uniformMatrix3fv(this.addr,!1,xo),xt(t,r)}}function tp(i,e){const t=this.cache,r=e.elements;if(r===void 0){if(vt(t,e))return;i.uniformMatrix4fv(this.addr,!1,e),xt(t,e)}else{if(vt(t,r))return;vo.set(r),i.uniformMatrix4fv(this.addr,!1,vo),xt(t,r)}}function rp(i,e){const t=this.cache;t[0]!==e&&(i.uniform1i(this.addr,e),t[0]=e)}function ip(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y)&&(i.uniform2i(this.addr,e.x,e.y),t[0]=e.x,t[1]=e.y);else{if(vt(t,e))return;i.uniform2iv(this.addr,e),xt(t,e)}}function np(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z)&&(i.uniform3i(this.addr,e.x,e.y,e.z),t[0]=e.x,t[1]=e.y,t[2]=e.z);else{if(vt(t,e))return;i.uniform3iv(this.addr,e),xt(t,e)}}function ap(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z||t[3]!==e.w)&&(i.uniform4i(this.addr,e.x,e.y,e.z,e.w),t[0]=e.x,t[1]=e.y,t[2]=e.z,t[3]=e.w);else{if(vt(t,e))return;i.uniform4iv(this.addr,e),xt(t,e)}}function sp(i,e){const t=this.cache;t[0]!==e&&(i.uniform1ui(this.addr,e),t[0]=e)}function op(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y)&&(i.uniform2ui(this.addr,e.x,e.y),t[0]=e.x,t[1]=e.y);else{if(vt(t,e))return;i.uniform2uiv(this.addr,e),xt(t,e)}}function lp(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z)&&(i.uniform3ui(this.addr,e.x,e.y,e.z),t[0]=e.x,t[1]=e.y,t[2]=e.z);else{if(vt(t,e))return;i.uniform3uiv(this.addr,e),xt(t,e)}}function cp(i,e){const t=this.cache;if(e.x!==void 0)(t[0]!==e.x||t[1]!==e.y||t[2]!==e.z||t[3]!==e.w)&&(i.uniform4ui(this.addr,e.x,e.y,e.z,e.w),t[0]=e.x,t[1]=e.y,t[2]=e.z,t[3]=e.w);else{if(vt(t,e))return;i.uniform4uiv(this.addr,e),xt(t,e)}}function hp(i,e,t){const r=this.cache,n=t.allocateTextureUnit();r[0]!==n&&(i.uniform1i(this.addr,n),r[0]=n);let a;this.type===i.SAMPLER_2D_SHADOW?(mo.compareFunction=sl,a=mo):a=Ml,t.setTexture2D(e||a,n)}function up(i,e,t){const r=this.cache,n=t.allocateTextureUnit();r[0]!==n&&(i.uniform1i(this.addr,n),r[0]=n),t.setTexture3D(e||El,n)}function dp(i,e,t){const r=this.cache,n=t.allocateTextureUnit();r[0]!==n&&(i.uniform1i(this.addr,n),r[0]=n),t.setTextureCube(e||bl,n)}function pp(i,e,t){const r=this.cache,n=t.allocateTextureUnit();r[0]!==n&&(i.uniform1i(this.addr,n),r[0]=n),t.setTexture2DArray(e||Sl,n)}function fp(i){switch(i){case 5126:return Kd;case 35664:return $d;case 35665:return Zd;case 35666:return Jd;case 35674:return Qd;case 35675:return ep;case 35676:return tp;case 5124:case 35670:return rp;case 35667:case 35671:return ip;case 35668:case 35672:return np;case 35669:case 35673:return ap;case 5125:return sp;case 36294:return op;case 36295:return lp;case 36296:return cp;case 35678:case 36198:case 36298:case 36306:case 35682:return hp;case 35679:case 36299:case 36307:return up;case 35680:case 36300:case 36308:case 36293:return dp;case 36289:case 36303:case 36311:case 36292:return pp}}function mp(i,e){i.uniform1fv(this.addr,e)}function gp(i,e){const t=Ei(e,this.size,2);i.uniform2fv(this.addr,t)}function _p(i,e){const t=Ei(e,this.size,3);i.uniform3fv(this.addr,t)}function vp(i,e){const t=Ei(e,this.size,4);i.uniform4fv(this.addr,t)}function xp(i,e){const t=Ei(e,this.size,4);i.uniformMatrix2fv(this.addr,!1,t)}function yp(i,e){const t=Ei(e,this.size,9);i.uniformMatrix3fv(this.addr,!1,t)}function Mp(i,e){const t=Ei(e,this.size,16);i.uniformMatrix4fv(this.addr,!1,t)}function Sp(i,e){i.uniform1iv(this.addr,e)}function Ep(i,e){i.uniform2iv(this.addr,e)}function bp(i,e){i.uniform3iv(this.addr,e)}function Tp(i,e){i.uniform4iv(this.addr,e)}function wp(i,e){i.uniform1uiv(this.addr,e)}function Ap(i,e){i.uniform2uiv(this.addr,e)}function Rp(i,e){i.uniform3uiv(this.addr,e)}function Cp(i,e){i.uniform4uiv(this.addr,e)}function Lp(i,e,t){const r=this.cache,n=e.length,a=Fn(t,n);vt(r,a)||(i.uniform1iv(this.addr,a),xt(r,a));for(let o=0;o!==n;++o)t.setTexture2D(e[o]||Ml,a[o])}function Pp(i,e,t){const r=this.cache,n=e.length,a=Fn(t,n);vt(r,a)||(i.uniform1iv(this.addr,a),xt(r,a));for(let o=0;o!==n;++o)t.setTexture3D(e[o]||El,a[o])}function Dp(i,e,t){const r=this.cache,n=e.length,a=Fn(t,n);vt(r,a)||(i.uniform1iv(this.addr,a),xt(r,a));for(let o=0;o!==n;++o)t.setTextureCube(e[o]||bl,a[o])}function Ip(i,e,t){const r=this.cache,n=e.length,a=Fn(t,n);vt(r,a)||(i.uniform1iv(this.addr,a),xt(r,a));for(let o=0;o!==n;++o)t.setTexture2DArray(e[o]||Sl,a[o])}function Np(i){switch(i){case 5126:return mp;case 35664:return gp;case 35665:return _p;case 35666:return vp;case 35674:return xp;case 35675:return yp;case 35676:return Mp;case 5124:case 35670:return Sp;case 35667:case 35671:return Ep;case 35668:case 35672:return bp;case 35669:case 35673:return Tp;case 5125:return wp;case 36294:return Ap;case 36295:return Rp;case 36296:return Cp;case 35678:case 36198:case 36298:case 36306:case 35682:return Lp;case 35679:case 36299:case 36307:return Pp;case 35680:case 36300:case 36308:case 36293:return Dp;case 36289:case 36303:case 36311:case 36292:return Ip}}class Up{constructor(e,t,r){this.id=e,this.addr=r,this.cache=[],this.type=t.type,this.setValue=fp(t.type)}}class Op{constructor(e,t,r){this.id=e,this.addr=r,this.cache=[],this.type=t.type,this.size=t.size,this.setValue=Np(t.type)}}class Fp{constructor(e){this.id=e,this.seq=[],this.map={}}setValue(e,t,r){const n=this.seq;for(let a=0,o=n.length;a!==o;++a){const s=n[a];s.setValue(e,t[s.id],r)}}}const fa=/(\w+)(\])?(\[|\.)?/g;function Mo(i,e){i.seq.push(e),i.map[e.id]=e}function zp(i,e,t){const r=i.name,n=r.length;for(fa.lastIndex=0;;){const a=fa.exec(r),o=fa.lastIndex;let s=a[1];const l=a[2]==="]",c=a[3];if(l&&(s=s|0),c===void 0||c==="["&&o+2===n){Mo(t,c===void 0?new Up(s,i,e):new Op(s,i,e));break}else{let h=t.map[s];h===void 0&&(h=new Fp(s),Mo(t,h)),t=h}}}class Ln{constructor(e,t){this.seq=[],this.map={};const r=e.getProgramParameter(t,e.ACTIVE_UNIFORMS);for(let n=0;n<r;++n){const a=e.getActiveUniform(t,n),o=e.getUniformLocation(t,a.name);zp(a,o,this)}}setValue(e,t,r,n){const a=this.map[t];a!==void 0&&a.setValue(e,r,n)}setOptional(e,t,r){const n=t[r];n!==void 0&&this.setValue(e,r,n)}static upload(e,t,r,n){for(let a=0,o=t.length;a!==o;++a){const s=t[a],l=r[s.id];l.needsUpdate!==!1&&s.setValue(e,l.value,n)}}static seqWithValue(e,t){const r=[];for(let n=0,a=e.length;n!==a;++n){const o=e[n];o.id in t&&r.push(o)}return r}}function So(i,e,t){const r=i.createShader(e);return i.shaderSource(r,t),i.compileShader(r),r}const Bp=37297;let kp=0;function Vp(i,e){const t=i.split(`
`),r=[],n=Math.max(e-6,0),a=Math.min(e+6,t.length);for(let o=n;o<a;o++){const s=o+1;r.push(`${s===e?">":" "} ${s}: ${t[o]}`)}return r.join(`
`)}const Eo=new Ve;function Hp(i){$e._getMatrix(Eo,$e.workingColorSpace,i);const e=`mat3( ${Eo.elements.map(t=>t.toFixed(4))} )`;switch($e.getTransfer(i)){case On:return[e,"LinearTransferOETF"];case rt:return[e,"sRGBTransferOETF"];default:return console.warn("THREE.WebGLProgram: Unsupported color space: ",i),[e,"LinearTransferOETF"]}}function bo(i,e,t){const r=i.getShaderParameter(e,i.COMPILE_STATUS),n=i.getShaderInfoLog(e).trim();if(r&&n==="")return"";const a=/ERROR: 0:(\d+)/.exec(n);if(a){const o=parseInt(a[1]);return t.toUpperCase()+`

`+n+`

`+Vp(i.getShaderSource(e),o)}else return n}function Gp(i,e){const t=Hp(e);return[`vec4 ${i}( vec4 value ) {`,`	return ${t[1]}( vec4( value.rgb * ${t[0]}, value.a ) );`,"}"].join(`
`)}function Wp(i,e){let t;switch(e){case oc:t="Linear";break;case lc:t="Reinhard";break;case cc:t="Cineon";break;case hc:t="ACESFilmic";break;case dc:t="AgX";break;case pc:t="Neutral";break;case uc:t="Custom";break;default:console.warn("THREE.WebGLProgram: Unsupported toneMapping:",e),t="Linear"}return"vec3 "+i+"( vec3 color ) { return "+t+"ToneMapping( color ); }"}const gn=new B;function jp(){$e.getLuminanceCoefficients(gn);const i=gn.x.toFixed(4),e=gn.y.toFixed(4),t=gn.z.toFixed(4);return["float luminance( const in vec3 rgb ) {",`	const vec3 weights = vec3( ${i}, ${e}, ${t} );`,"	return dot( weights, rgb );","}"].join(`
`)}function Xp(i){return[i.extensionClipCullDistance?"#extension GL_ANGLE_clip_cull_distance : require":"",i.extensionMultiDraw?"#extension GL_ANGLE_multi_draw : require":""].filter(zi).join(`
`)}function Yp(i){const e=[];for(const t in i){const r=i[t];r!==!1&&e.push("#define "+t+" "+r)}return e.join(`
`)}function qp(i,e){const t={},r=i.getProgramParameter(e,i.ACTIVE_ATTRIBUTES);for(let n=0;n<r;n++){const a=i.getActiveAttrib(e,n),o=a.name;let s=1;a.type===i.FLOAT_MAT2&&(s=2),a.type===i.FLOAT_MAT3&&(s=3),a.type===i.FLOAT_MAT4&&(s=4),t[o]={type:a.type,location:i.getAttribLocation(e,o),locationSize:s}}return t}function zi(i){return i!==""}function To(i,e){const t=e.numSpotLightShadows+e.numSpotLightMaps-e.numSpotLightShadowsWithMaps;return i.replace(/NUM_DIR_LIGHTS/g,e.numDirLights).replace(/NUM_SPOT_LIGHTS/g,e.numSpotLights).replace(/NUM_SPOT_LIGHT_MAPS/g,e.numSpotLightMaps).replace(/NUM_SPOT_LIGHT_COORDS/g,t).replace(/NUM_RECT_AREA_LIGHTS/g,e.numRectAreaLights).replace(/NUM_POINT_LIGHTS/g,e.numPointLights).replace(/NUM_HEMI_LIGHTS/g,e.numHemiLights).replace(/NUM_DIR_LIGHT_SHADOWS/g,e.numDirLightShadows).replace(/NUM_SPOT_LIGHT_SHADOWS_WITH_MAPS/g,e.numSpotLightShadowsWithMaps).replace(/NUM_SPOT_LIGHT_SHADOWS/g,e.numSpotLightShadows).replace(/NUM_POINT_LIGHT_SHADOWS/g,e.numPointLightShadows)}function wo(i,e){return i.replace(/NUM_CLIPPING_PLANES/g,e.numClippingPlanes).replace(/UNION_CLIPPING_PLANES/g,e.numClippingPlanes-e.numClipIntersection)}const Kp=/^[ \t]*#include +<([\w\d./]+)>/gm;function os(i){return i.replace(Kp,Zp)}const $p=new Map;function Zp(i,e){let t=Ge[e];if(t===void 0){const r=$p.get(e);if(r!==void 0)t=Ge[r],console.warn('THREE.WebGLRenderer: Shader chunk "%s" has been deprecated. Use "%s" instead.',e,r);else throw new Error("Can not resolve #include <"+e+">")}return os(t)}const Jp=/#pragma unroll_loop_start\s+for\s*\(\s*int\s+i\s*=\s*(\d+)\s*;\s*i\s*<\s*(\d+)\s*;\s*i\s*\+\+\s*\)\s*{([\s\S]+?)}\s+#pragma unroll_loop_end/g;function Ao(i){return i.replace(Jp,Qp)}function Qp(i,e,t,r){let n="";for(let a=parseInt(e);a<parseInt(t);a++)n+=r.replace(/\[\s*i\s*\]/g,"[ "+a+" ]").replace(/UNROLLED_LOOP_INDEX/g,a);return n}function Ro(i){let e=`precision ${i.precision} float;
	precision ${i.precision} int;
	precision ${i.precision} sampler2D;
	precision ${i.precision} samplerCube;
	precision ${i.precision} sampler3D;
	precision ${i.precision} sampler2DArray;
	precision ${i.precision} sampler2DShadow;
	precision ${i.precision} samplerCubeShadow;
	precision ${i.precision} sampler2DArrayShadow;
	precision ${i.precision} isampler2D;
	precision ${i.precision} isampler3D;
	precision ${i.precision} isamplerCube;
	precision ${i.precision} isampler2DArray;
	precision ${i.precision} usampler2D;
	precision ${i.precision} usampler3D;
	precision ${i.precision} usamplerCube;
	precision ${i.precision} usampler2DArray;
	`;return i.precision==="highp"?e+=`
#define HIGH_PRECISION`:i.precision==="mediump"?e+=`
#define MEDIUM_PRECISION`:i.precision==="lowp"&&(e+=`
#define LOW_PRECISION`),e}function ef(i){let e="SHADOWMAP_TYPE_BASIC";return i.shadowMapType===qo?e="SHADOWMAP_TYPE_PCF":i.shadowMapType===kl?e="SHADOWMAP_TYPE_PCF_SOFT":i.shadowMapType===sr&&(e="SHADOWMAP_TYPE_VSM"),e}function tf(i){let e="ENVMAP_TYPE_CUBE";if(i.envMap)switch(i.envMapMode){case mi:case gi:e="ENVMAP_TYPE_CUBE";break;case Un:e="ENVMAP_TYPE_CUBE_UV";break}return e}function rf(i){let e="ENVMAP_MODE_REFLECTION";if(i.envMap)switch(i.envMapMode){case gi:e="ENVMAP_MODE_REFRACTION";break}return e}function nf(i){let e="ENVMAP_BLENDING_NONE";if(i.envMap)switch(i.combine){case ds:e="ENVMAP_BLENDING_MULTIPLY";break;case ac:e="ENVMAP_BLENDING_MIX";break;case sc:e="ENVMAP_BLENDING_ADD";break}return e}function af(i){const e=i.envMapCubeUVHeight;if(e===null)return null;const t=Math.log2(e)-2,r=1/e;return{texelWidth:1/(3*Math.max(Math.pow(2,t),7*16)),texelHeight:r,maxMip:t}}function sf(i,e,t,r){const n=i.getContext(),a=t.defines;let o=t.vertexShader,s=t.fragmentShader;const l=ef(t),c=tf(t),h=rf(t),u=nf(t),p=af(t),f=Xp(t),g=Yp(a),_=n.createProgram();let m,d,S=t.glslVersion?"#version "+t.glslVersion+`
`:"";t.isRawShaderMaterial?(m=["#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,g].filter(zi).join(`
`),m.length>0&&(m+=`
`),d=["#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,g].filter(zi).join(`
`),d.length>0&&(d+=`
`)):(m=[Ro(t),"#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,g,t.extensionClipCullDistance?"#define USE_CLIP_DISTANCE":"",t.batching?"#define USE_BATCHING":"",t.batchingColor?"#define USE_BATCHING_COLOR":"",t.instancing?"#define USE_INSTANCING":"",t.instancingColor?"#define USE_INSTANCING_COLOR":"",t.instancingMorph?"#define USE_INSTANCING_MORPH":"",t.useFog&&t.fog?"#define USE_FOG":"",t.useFog&&t.fogExp2?"#define FOG_EXP2":"",t.map?"#define USE_MAP":"",t.envMap?"#define USE_ENVMAP":"",t.envMap?"#define "+h:"",t.lightMap?"#define USE_LIGHTMAP":"",t.aoMap?"#define USE_AOMAP":"",t.bumpMap?"#define USE_BUMPMAP":"",t.normalMap?"#define USE_NORMALMAP":"",t.normalMapObjectSpace?"#define USE_NORMALMAP_OBJECTSPACE":"",t.normalMapTangentSpace?"#define USE_NORMALMAP_TANGENTSPACE":"",t.displacementMap?"#define USE_DISPLACEMENTMAP":"",t.emissiveMap?"#define USE_EMISSIVEMAP":"",t.anisotropy?"#define USE_ANISOTROPY":"",t.anisotropyMap?"#define USE_ANISOTROPYMAP":"",t.clearcoatMap?"#define USE_CLEARCOATMAP":"",t.clearcoatRoughnessMap?"#define USE_CLEARCOAT_ROUGHNESSMAP":"",t.clearcoatNormalMap?"#define USE_CLEARCOAT_NORMALMAP":"",t.iridescenceMap?"#define USE_IRIDESCENCEMAP":"",t.iridescenceThicknessMap?"#define USE_IRIDESCENCE_THICKNESSMAP":"",t.specularMap?"#define USE_SPECULARMAP":"",t.specularColorMap?"#define USE_SPECULAR_COLORMAP":"",t.specularIntensityMap?"#define USE_SPECULAR_INTENSITYMAP":"",t.roughnessMap?"#define USE_ROUGHNESSMAP":"",t.metalnessMap?"#define USE_METALNESSMAP":"",t.alphaMap?"#define USE_ALPHAMAP":"",t.alphaHash?"#define USE_ALPHAHASH":"",t.transmission?"#define USE_TRANSMISSION":"",t.transmissionMap?"#define USE_TRANSMISSIONMAP":"",t.thicknessMap?"#define USE_THICKNESSMAP":"",t.sheenColorMap?"#define USE_SHEEN_COLORMAP":"",t.sheenRoughnessMap?"#define USE_SHEEN_ROUGHNESSMAP":"",t.mapUv?"#define MAP_UV "+t.mapUv:"",t.alphaMapUv?"#define ALPHAMAP_UV "+t.alphaMapUv:"",t.lightMapUv?"#define LIGHTMAP_UV "+t.lightMapUv:"",t.aoMapUv?"#define AOMAP_UV "+t.aoMapUv:"",t.emissiveMapUv?"#define EMISSIVEMAP_UV "+t.emissiveMapUv:"",t.bumpMapUv?"#define BUMPMAP_UV "+t.bumpMapUv:"",t.normalMapUv?"#define NORMALMAP_UV "+t.normalMapUv:"",t.displacementMapUv?"#define DISPLACEMENTMAP_UV "+t.displacementMapUv:"",t.metalnessMapUv?"#define METALNESSMAP_UV "+t.metalnessMapUv:"",t.roughnessMapUv?"#define ROUGHNESSMAP_UV "+t.roughnessMapUv:"",t.anisotropyMapUv?"#define ANISOTROPYMAP_UV "+t.anisotropyMapUv:"",t.clearcoatMapUv?"#define CLEARCOATMAP_UV "+t.clearcoatMapUv:"",t.clearcoatNormalMapUv?"#define CLEARCOAT_NORMALMAP_UV "+t.clearcoatNormalMapUv:"",t.clearcoatRoughnessMapUv?"#define CLEARCOAT_ROUGHNESSMAP_UV "+t.clearcoatRoughnessMapUv:"",t.iridescenceMapUv?"#define IRIDESCENCEMAP_UV "+t.iridescenceMapUv:"",t.iridescenceThicknessMapUv?"#define IRIDESCENCE_THICKNESSMAP_UV "+t.iridescenceThicknessMapUv:"",t.sheenColorMapUv?"#define SHEEN_COLORMAP_UV "+t.sheenColorMapUv:"",t.sheenRoughnessMapUv?"#define SHEEN_ROUGHNESSMAP_UV "+t.sheenRoughnessMapUv:"",t.specularMapUv?"#define SPECULARMAP_UV "+t.specularMapUv:"",t.specularColorMapUv?"#define SPECULAR_COLORMAP_UV "+t.specularColorMapUv:"",t.specularIntensityMapUv?"#define SPECULAR_INTENSITYMAP_UV "+t.specularIntensityMapUv:"",t.transmissionMapUv?"#define TRANSMISSIONMAP_UV "+t.transmissionMapUv:"",t.thicknessMapUv?"#define THICKNESSMAP_UV "+t.thicknessMapUv:"",t.vertexTangents&&t.flatShading===!1?"#define USE_TANGENT":"",t.vertexColors?"#define USE_COLOR":"",t.vertexAlphas?"#define USE_COLOR_ALPHA":"",t.vertexUv1s?"#define USE_UV1":"",t.vertexUv2s?"#define USE_UV2":"",t.vertexUv3s?"#define USE_UV3":"",t.pointsUvs?"#define USE_POINTS_UV":"",t.flatShading?"#define FLAT_SHADED":"",t.skinning?"#define USE_SKINNING":"",t.morphTargets?"#define USE_MORPHTARGETS":"",t.morphNormals&&t.flatShading===!1?"#define USE_MORPHNORMALS":"",t.morphColors?"#define USE_MORPHCOLORS":"",t.morphTargetsCount>0?"#define MORPHTARGETS_TEXTURE_STRIDE "+t.morphTextureStride:"",t.morphTargetsCount>0?"#define MORPHTARGETS_COUNT "+t.morphTargetsCount:"",t.doubleSided?"#define DOUBLE_SIDED":"",t.flipSided?"#define FLIP_SIDED":"",t.shadowMapEnabled?"#define USE_SHADOWMAP":"",t.shadowMapEnabled?"#define "+l:"",t.sizeAttenuation?"#define USE_SIZEATTENUATION":"",t.numLightProbes>0?"#define USE_LIGHT_PROBES":"",t.logarithmicDepthBuffer?"#define USE_LOGDEPTHBUF":"",t.reverseDepthBuffer?"#define USE_REVERSEDEPTHBUF":"","uniform mat4 modelMatrix;","uniform mat4 modelViewMatrix;","uniform mat4 projectionMatrix;","uniform mat4 viewMatrix;","uniform mat3 normalMatrix;","uniform vec3 cameraPosition;","uniform bool isOrthographic;","#ifdef USE_INSTANCING","	attribute mat4 instanceMatrix;","#endif","#ifdef USE_INSTANCING_COLOR","	attribute vec3 instanceColor;","#endif","#ifdef USE_INSTANCING_MORPH","	uniform sampler2D morphTexture;","#endif","attribute vec3 position;","attribute vec3 normal;","attribute vec2 uv;","#ifdef USE_UV1","	attribute vec2 uv1;","#endif","#ifdef USE_UV2","	attribute vec2 uv2;","#endif","#ifdef USE_UV3","	attribute vec2 uv3;","#endif","#ifdef USE_TANGENT","	attribute vec4 tangent;","#endif","#if defined( USE_COLOR_ALPHA )","	attribute vec4 color;","#elif defined( USE_COLOR )","	attribute vec3 color;","#endif","#ifdef USE_SKINNING","	attribute vec4 skinIndex;","	attribute vec4 skinWeight;","#endif",`
`].filter(zi).join(`
`),d=[Ro(t),"#define SHADER_TYPE "+t.shaderType,"#define SHADER_NAME "+t.shaderName,g,t.useFog&&t.fog?"#define USE_FOG":"",t.useFog&&t.fogExp2?"#define FOG_EXP2":"",t.alphaToCoverage?"#define ALPHA_TO_COVERAGE":"",t.map?"#define USE_MAP":"",t.matcap?"#define USE_MATCAP":"",t.envMap?"#define USE_ENVMAP":"",t.envMap?"#define "+c:"",t.envMap?"#define "+h:"",t.envMap?"#define "+u:"",p?"#define CUBEUV_TEXEL_WIDTH "+p.texelWidth:"",p?"#define CUBEUV_TEXEL_HEIGHT "+p.texelHeight:"",p?"#define CUBEUV_MAX_MIP "+p.maxMip+".0":"",t.lightMap?"#define USE_LIGHTMAP":"",t.aoMap?"#define USE_AOMAP":"",t.bumpMap?"#define USE_BUMPMAP":"",t.normalMap?"#define USE_NORMALMAP":"",t.normalMapObjectSpace?"#define USE_NORMALMAP_OBJECTSPACE":"",t.normalMapTangentSpace?"#define USE_NORMALMAP_TANGENTSPACE":"",t.emissiveMap?"#define USE_EMISSIVEMAP":"",t.anisotropy?"#define USE_ANISOTROPY":"",t.anisotropyMap?"#define USE_ANISOTROPYMAP":"",t.clearcoat?"#define USE_CLEARCOAT":"",t.clearcoatMap?"#define USE_CLEARCOATMAP":"",t.clearcoatRoughnessMap?"#define USE_CLEARCOAT_ROUGHNESSMAP":"",t.clearcoatNormalMap?"#define USE_CLEARCOAT_NORMALMAP":"",t.dispersion?"#define USE_DISPERSION":"",t.iridescence?"#define USE_IRIDESCENCE":"",t.iridescenceMap?"#define USE_IRIDESCENCEMAP":"",t.iridescenceThicknessMap?"#define USE_IRIDESCENCE_THICKNESSMAP":"",t.specularMap?"#define USE_SPECULARMAP":"",t.specularColorMap?"#define USE_SPECULAR_COLORMAP":"",t.specularIntensityMap?"#define USE_SPECULAR_INTENSITYMAP":"",t.roughnessMap?"#define USE_ROUGHNESSMAP":"",t.metalnessMap?"#define USE_METALNESSMAP":"",t.alphaMap?"#define USE_ALPHAMAP":"",t.alphaTest?"#define USE_ALPHATEST":"",t.alphaHash?"#define USE_ALPHAHASH":"",t.sheen?"#define USE_SHEEN":"",t.sheenColorMap?"#define USE_SHEEN_COLORMAP":"",t.sheenRoughnessMap?"#define USE_SHEEN_ROUGHNESSMAP":"",t.transmission?"#define USE_TRANSMISSION":"",t.transmissionMap?"#define USE_TRANSMISSIONMAP":"",t.thicknessMap?"#define USE_THICKNESSMAP":"",t.vertexTangents&&t.flatShading===!1?"#define USE_TANGENT":"",t.vertexColors||t.instancingColor||t.batchingColor?"#define USE_COLOR":"",t.vertexAlphas?"#define USE_COLOR_ALPHA":"",t.vertexUv1s?"#define USE_UV1":"",t.vertexUv2s?"#define USE_UV2":"",t.vertexUv3s?"#define USE_UV3":"",t.pointsUvs?"#define USE_POINTS_UV":"",t.gradientMap?"#define USE_GRADIENTMAP":"",t.flatShading?"#define FLAT_SHADED":"",t.doubleSided?"#define DOUBLE_SIDED":"",t.flipSided?"#define FLIP_SIDED":"",t.shadowMapEnabled?"#define USE_SHADOWMAP":"",t.shadowMapEnabled?"#define "+l:"",t.premultipliedAlpha?"#define PREMULTIPLIED_ALPHA":"",t.numLightProbes>0?"#define USE_LIGHT_PROBES":"",t.decodeVideoTexture?"#define DECODE_VIDEO_TEXTURE":"",t.decodeVideoTextureEmissive?"#define DECODE_VIDEO_TEXTURE_EMISSIVE":"",t.logarithmicDepthBuffer?"#define USE_LOGDEPTHBUF":"",t.reverseDepthBuffer?"#define USE_REVERSEDEPTHBUF":"","uniform mat4 viewMatrix;","uniform vec3 cameraPosition;","uniform bool isOrthographic;",t.toneMapping!==yr?"#define TONE_MAPPING":"",t.toneMapping!==yr?Ge.tonemapping_pars_fragment:"",t.toneMapping!==yr?Wp("toneMapping",t.toneMapping):"",t.dithering?"#define DITHERING":"",t.opaque?"#define OPAQUE":"",Ge.colorspace_pars_fragment,Gp("linearToOutputTexel",t.outputColorSpace),jp(),t.useDepthPacking?"#define DEPTH_PACKING "+t.depthPacking:"",`
`].filter(zi).join(`
`)),o=os(o),o=To(o,t),o=wo(o,t),s=os(s),s=To(s,t),s=wo(s,t),o=Ao(o),s=Ao(s),t.isRawShaderMaterial!==!0&&(S=`#version 300 es
`,m=[f,"#define attribute in","#define varying out","#define texture2D texture"].join(`
`)+`
`+m,d=["#define varying in",t.glslVersion===ks?"":"layout(location = 0) out highp vec4 pc_fragColor;",t.glslVersion===ks?"":"#define gl_FragColor pc_fragColor","#define gl_FragDepthEXT gl_FragDepth","#define texture2D texture","#define textureCube texture","#define texture2DProj textureProj","#define texture2DLodEXT textureLod","#define texture2DProjLodEXT textureProjLod","#define textureCubeLodEXT textureLod","#define texture2DGradEXT textureGrad","#define texture2DProjGradEXT textureProjGrad","#define textureCubeGradEXT textureGrad"].join(`
`)+`
`+d);const T=S+m+o,x=S+d+s,N=So(n,n.VERTEX_SHADER,T),w=So(n,n.FRAGMENT_SHADER,x);n.attachShader(_,N),n.attachShader(_,w),t.index0AttributeName!==void 0?n.bindAttribLocation(_,0,t.index0AttributeName):t.morphTargets===!0&&n.bindAttribLocation(_,0,"position"),n.linkProgram(_);function A(R){if(i.debug.checkShaderErrors){const H=n.getProgramInfoLog(_).trim(),U=n.getShaderInfoLog(N).trim(),F=n.getShaderInfoLog(w).trim();let K=!0,X=!0;if(n.getProgramParameter(_,n.LINK_STATUS)===!1)if(K=!1,typeof i.debug.onShaderError=="function")i.debug.onShaderError(n,_,N,w);else{const $=bo(n,N,"vertex"),k=bo(n,w,"fragment");console.error("THREE.WebGLProgram: Shader Error "+n.getError()+" - VALIDATE_STATUS "+n.getProgramParameter(_,n.VALIDATE_STATUS)+`

Material Name: `+R.name+`
Material Type: `+R.type+`

Program Info Log: `+H+`
`+$+`
`+k)}else H!==""?console.warn("THREE.WebGLProgram: Program Info Log:",H):(U===""||F==="")&&(X=!1);X&&(R.diagnostics={runnable:K,programLog:H,vertexShader:{log:U,prefix:m},fragmentShader:{log:F,prefix:d}})}n.deleteShader(N),n.deleteShader(w),O=new Ln(n,_),b=qp(n,_)}let O;this.getUniforms=function(){return O===void 0&&A(this),O};let b;this.getAttributes=function(){return b===void 0&&A(this),b};let M=t.rendererExtensionParallelShaderCompile===!1;return this.isReady=function(){return M===!1&&(M=n.getProgramParameter(_,Bp)),M},this.destroy=function(){r.releaseStatesOfProgram(this),n.deleteProgram(_),this.program=void 0},this.type=t.shaderType,this.name=t.shaderName,this.id=kp++,this.cacheKey=e,this.usedTimes=1,this.program=_,this.vertexShader=N,this.fragmentShader=w,this}let of=0;class lf{constructor(){this.shaderCache=new Map,this.materialCache=new Map}update(e){const t=e.vertexShader,r=e.fragmentShader,n=this._getShaderStage(t),a=this._getShaderStage(r),o=this._getShaderCacheForMaterial(e);return o.has(n)===!1&&(o.add(n),n.usedTimes++),o.has(a)===!1&&(o.add(a),a.usedTimes++),this}remove(e){const t=this.materialCache.get(e);for(const r of t)r.usedTimes--,r.usedTimes===0&&this.shaderCache.delete(r.code);return this.materialCache.delete(e),this}getVertexShaderID(e){return this._getShaderStage(e.vertexShader).id}getFragmentShaderID(e){return this._getShaderStage(e.fragmentShader).id}dispose(){this.shaderCache.clear(),this.materialCache.clear()}_getShaderCacheForMaterial(e){const t=this.materialCache;let r=t.get(e);return r===void 0&&(r=new Set,t.set(e,r)),r}_getShaderStage(e){const t=this.shaderCache;let r=t.get(e);return r===void 0&&(r=new cf(e),t.set(e,r)),r}}class cf{constructor(e){this.id=of++,this.code=e,this.usedTimes=0}}function hf(i,e,t,r,n,a,o){const s=new ul,l=new lf,c=new Set,h=[],u=n.logarithmicDepthBuffer,p=n.vertexTextures;let f=n.precision;const g={MeshDepthMaterial:"depth",MeshDistanceMaterial:"distanceRGBA",MeshNormalMaterial:"normal",MeshBasicMaterial:"basic",MeshLambertMaterial:"lambert",MeshPhongMaterial:"phong",MeshToonMaterial:"toon",MeshStandardMaterial:"physical",MeshPhysicalMaterial:"physical",MeshMatcapMaterial:"matcap",LineBasicMaterial:"basic",LineDashedMaterial:"dashed",PointsMaterial:"points",ShadowMaterial:"shadow",SpriteMaterial:"sprite"};function _(b){return c.add(b),b===0?"uv":`uv${b}`}function m(b,M,R,H,U){const F=H.fog,K=U.geometry,X=b.isMeshStandardMaterial?H.environment:null,$=(b.isMeshStandardMaterial?t:e).get(b.envMap||X),k=$&&$.mapping===Un?$.image.height:null,Z=g[b.type];b.precision!==null&&(f=n.getMaxPrecision(b.precision),f!==b.precision&&console.warn("THREE.WebGLProgram.getParameters:",b.precision,"not supported, using",f,"instead."));const ce=K.morphAttributes.position||K.morphAttributes.normal||K.morphAttributes.color,ve=ce!==void 0?ce.length:0;let De=0;K.morphAttributes.position!==void 0&&(De=1),K.morphAttributes.normal!==void 0&&(De=2),K.morphAttributes.color!==void 0&&(De=3);let xe,G,J,q;if(Z){const et=Zt[Z];xe=et.vertexShader,G=et.fragmentShader}else xe=b.vertexShader,G=b.fragmentShader,l.update(b),J=l.getVertexShaderID(b),q=l.getFragmentShaderID(b);const z=i.getRenderTarget(),ae=i.state.buffers.depth.getReversed(),ne=U.isInstancedMesh===!0,oe=U.isBatchedMesh===!0,Ae=!!b.map,Te=!!b.matcap,Be=!!$,C=!!b.aoMap,Je=!!b.lightMap,ue=!!b.bumpMap,pe=!!b.normalMap,ie=!!b.displacementMap,le=!!b.emissiveMap,ge=!!b.metalnessMap,E=!!b.roughnessMap,v=b.anisotropy>0,V=b.clearcoat>0,Q=b.dispersion>0,re=b.iridescence>0,ee=b.sheen>0,Re=b.transmission>0,_e=v&&!!b.anisotropyMap,be=V&&!!b.clearcoatMap,qe=V&&!!b.clearcoatNormalMap,he=V&&!!b.clearcoatRoughnessMap,Se=re&&!!b.iridescenceMap,Ue=re&&!!b.iridescenceThicknessMap,Oe=ee&&!!b.sheenColorMap,Ee=ee&&!!b.sheenRoughnessMap,Xe=!!b.specularMap,He=!!b.specularColorMap,nt=!!b.specularIntensityMap,D=Re&&!!b.transmissionMap,me=Re&&!!b.thicknessMap,Y=!!b.gradientMap,te=!!b.alphaMap,ye=b.alphaTest>0,de=!!b.alphaHash,Ye=!!b.extensions;let ut=yr;b.toneMapped&&(z===null||z.isXRRenderTarget===!0)&&(ut=i.toneMapping);const St={shaderID:Z,shaderType:b.type,shaderName:b.name,vertexShader:xe,fragmentShader:G,defines:b.defines,customVertexShaderID:J,customFragmentShaderID:q,isRawShaderMaterial:b.isRawShaderMaterial===!0,glslVersion:b.glslVersion,precision:f,batching:oe,batchingColor:oe&&U._colorsTexture!==null,instancing:ne,instancingColor:ne&&U.instanceColor!==null,instancingMorph:ne&&U.morphTexture!==null,supportsVertexTextures:p,outputColorSpace:z===null?i.outputColorSpace:z.isXRRenderTarget===!0?z.texture.colorSpace:yi,alphaToCoverage:!!b.alphaToCoverage,map:Ae,matcap:Te,envMap:Be,envMapMode:Be&&$.mapping,envMapCubeUVHeight:k,aoMap:C,lightMap:Je,bumpMap:ue,normalMap:pe,displacementMap:p&&ie,emissiveMap:le,normalMapObjectSpace:pe&&b.normalMapType===_c,normalMapTangentSpace:pe&&b.normalMapType===al,metalnessMap:ge,roughnessMap:E,anisotropy:v,anisotropyMap:_e,clearcoat:V,clearcoatMap:be,clearcoatNormalMap:qe,clearcoatRoughnessMap:he,dispersion:Q,iridescence:re,iridescenceMap:Se,iridescenceThicknessMap:Ue,sheen:ee,sheenColorMap:Oe,sheenRoughnessMap:Ee,specularMap:Xe,specularColorMap:He,specularIntensityMap:nt,transmission:Re,transmissionMap:D,thicknessMap:me,gradientMap:Y,opaque:b.transparent===!1&&b.blending===ui&&b.alphaToCoverage===!1,alphaMap:te,alphaTest:ye,alphaHash:de,combine:b.combine,mapUv:Ae&&_(b.map.channel),aoMapUv:C&&_(b.aoMap.channel),lightMapUv:Je&&_(b.lightMap.channel),bumpMapUv:ue&&_(b.bumpMap.channel),normalMapUv:pe&&_(b.normalMap.channel),displacementMapUv:ie&&_(b.displacementMap.channel),emissiveMapUv:le&&_(b.emissiveMap.channel),metalnessMapUv:ge&&_(b.metalnessMap.channel),roughnessMapUv:E&&_(b.roughnessMap.channel),anisotropyMapUv:_e&&_(b.anisotropyMap.channel),clearcoatMapUv:be&&_(b.clearcoatMap.channel),clearcoatNormalMapUv:qe&&_(b.clearcoatNormalMap.channel),clearcoatRoughnessMapUv:he&&_(b.clearcoatRoughnessMap.channel),iridescenceMapUv:Se&&_(b.iridescenceMap.channel),iridescenceThicknessMapUv:Ue&&_(b.iridescenceThicknessMap.channel),sheenColorMapUv:Oe&&_(b.sheenColorMap.channel),sheenRoughnessMapUv:Ee&&_(b.sheenRoughnessMap.channel),specularMapUv:Xe&&_(b.specularMap.channel),specularColorMapUv:He&&_(b.specularColorMap.channel),specularIntensityMapUv:nt&&_(b.specularIntensityMap.channel),transmissionMapUv:D&&_(b.transmissionMap.channel),thicknessMapUv:me&&_(b.thicknessMap.channel),alphaMapUv:te&&_(b.alphaMap.channel),vertexTangents:!!K.attributes.tangent&&(pe||v),vertexColors:b.vertexColors,vertexAlphas:b.vertexColors===!0&&!!K.attributes.color&&K.attributes.color.itemSize===4,pointsUvs:U.isPoints===!0&&!!K.attributes.uv&&(Ae||te),fog:!!F,useFog:b.fog===!0,fogExp2:!!F&&F.isFogExp2,flatShading:b.flatShading===!0,sizeAttenuation:b.sizeAttenuation===!0,logarithmicDepthBuffer:u,reverseDepthBuffer:ae,skinning:U.isSkinnedMesh===!0,morphTargets:K.morphAttributes.position!==void 0,morphNormals:K.morphAttributes.normal!==void 0,morphColors:K.morphAttributes.color!==void 0,morphTargetsCount:ve,morphTextureStride:De,numDirLights:M.directional.length,numPointLights:M.point.length,numSpotLights:M.spot.length,numSpotLightMaps:M.spotLightMap.length,numRectAreaLights:M.rectArea.length,numHemiLights:M.hemi.length,numDirLightShadows:M.directionalShadowMap.length,numPointLightShadows:M.pointShadowMap.length,numSpotLightShadows:M.spotShadowMap.length,numSpotLightShadowsWithMaps:M.numSpotLightShadowsWithMaps,numLightProbes:M.numLightProbes,numClippingPlanes:o.numPlanes,numClipIntersection:o.numIntersection,dithering:b.dithering,shadowMapEnabled:i.shadowMap.enabled&&R.length>0,shadowMapType:i.shadowMap.type,toneMapping:ut,decodeVideoTexture:Ae&&b.map.isVideoTexture===!0&&$e.getTransfer(b.map.colorSpace)===rt,decodeVideoTextureEmissive:le&&b.emissiveMap.isVideoTexture===!0&&$e.getTransfer(b.emissiveMap.colorSpace)===rt,premultipliedAlpha:b.premultipliedAlpha,doubleSided:b.side===Rt,flipSided:b.side===Dt,useDepthPacking:b.depthPacking>=0,depthPacking:b.depthPacking||0,index0AttributeName:b.index0AttributeName,extensionClipCullDistance:Ye&&b.extensions.clipCullDistance===!0&&r.has("WEBGL_clip_cull_distance"),extensionMultiDraw:(Ye&&b.extensions.multiDraw===!0||oe)&&r.has("WEBGL_multi_draw"),rendererExtensionParallelShaderCompile:r.has("KHR_parallel_shader_compile"),customProgramCacheKey:b.customProgramCacheKey()};return St.vertexUv1s=c.has(1),St.vertexUv2s=c.has(2),St.vertexUv3s=c.has(3),c.clear(),St}function d(b){const M=[];if(b.shaderID?M.push(b.shaderID):(M.push(b.customVertexShaderID),M.push(b.customFragmentShaderID)),b.defines!==void 0)for(const R in b.defines)M.push(R),M.push(b.defines[R]);return b.isRawShaderMaterial===!1&&(S(M,b),T(M,b),M.push(i.outputColorSpace)),M.push(b.customProgramCacheKey),M.join()}function S(b,M){b.push(M.precision),b.push(M.outputColorSpace),b.push(M.envMapMode),b.push(M.envMapCubeUVHeight),b.push(M.mapUv),b.push(M.alphaMapUv),b.push(M.lightMapUv),b.push(M.aoMapUv),b.push(M.bumpMapUv),b.push(M.normalMapUv),b.push(M.displacementMapUv),b.push(M.emissiveMapUv),b.push(M.metalnessMapUv),b.push(M.roughnessMapUv),b.push(M.anisotropyMapUv),b.push(M.clearcoatMapUv),b.push(M.clearcoatNormalMapUv),b.push(M.clearcoatRoughnessMapUv),b.push(M.iridescenceMapUv),b.push(M.iridescenceThicknessMapUv),b.push(M.sheenColorMapUv),b.push(M.sheenRoughnessMapUv),b.push(M.specularMapUv),b.push(M.specularColorMapUv),b.push(M.specularIntensityMapUv),b.push(M.transmissionMapUv),b.push(M.thicknessMapUv),b.push(M.combine),b.push(M.fogExp2),b.push(M.sizeAttenuation),b.push(M.morphTargetsCount),b.push(M.morphAttributeCount),b.push(M.numDirLights),b.push(M.numPointLights),b.push(M.numSpotLights),b.push(M.numSpotLightMaps),b.push(M.numHemiLights),b.push(M.numRectAreaLights),b.push(M.numDirLightShadows),b.push(M.numPointLightShadows),b.push(M.numSpotLightShadows),b.push(M.numSpotLightShadowsWithMaps),b.push(M.numLightProbes),b.push(M.shadowMapType),b.push(M.toneMapping),b.push(M.numClippingPlanes),b.push(M.numClipIntersection),b.push(M.depthPacking)}function T(b,M){s.disableAll(),M.supportsVertexTextures&&s.enable(0),M.instancing&&s.enable(1),M.instancingColor&&s.enable(2),M.instancingMorph&&s.enable(3),M.matcap&&s.enable(4),M.envMap&&s.enable(5),M.normalMapObjectSpace&&s.enable(6),M.normalMapTangentSpace&&s.enable(7),M.clearcoat&&s.enable(8),M.iridescence&&s.enable(9),M.alphaTest&&s.enable(10),M.vertexColors&&s.enable(11),M.vertexAlphas&&s.enable(12),M.vertexUv1s&&s.enable(13),M.vertexUv2s&&s.enable(14),M.vertexUv3s&&s.enable(15),M.vertexTangents&&s.enable(16),M.anisotropy&&s.enable(17),M.alphaHash&&s.enable(18),M.batching&&s.enable(19),M.dispersion&&s.enable(20),M.batchingColor&&s.enable(21),b.push(s.mask),s.disableAll(),M.fog&&s.enable(0),M.useFog&&s.enable(1),M.flatShading&&s.enable(2),M.logarithmicDepthBuffer&&s.enable(3),M.reverseDepthBuffer&&s.enable(4),M.skinning&&s.enable(5),M.morphTargets&&s.enable(6),M.morphNormals&&s.enable(7),M.morphColors&&s.enable(8),M.premultipliedAlpha&&s.enable(9),M.shadowMapEnabled&&s.enable(10),M.doubleSided&&s.enable(11),M.flipSided&&s.enable(12),M.useDepthPacking&&s.enable(13),M.dithering&&s.enable(14),M.transmission&&s.enable(15),M.sheen&&s.enable(16),M.opaque&&s.enable(17),M.pointsUvs&&s.enable(18),M.decodeVideoTexture&&s.enable(19),M.decodeVideoTextureEmissive&&s.enable(20),M.alphaToCoverage&&s.enable(21),b.push(s.mask)}function x(b){const M=g[b.type];let R;if(M){const H=Zt[M];R=Xc.clone(H.uniforms)}else R=b.uniforms;return R}function N(b,M){let R;for(let H=0,U=h.length;H<U;H++){const F=h[H];if(F.cacheKey===M){R=F,++R.usedTimes;break}}return R===void 0&&(R=new sf(i,M,b,a),h.push(R)),R}function w(b){if(--b.usedTimes===0){const M=h.indexOf(b);h[M]=h[h.length-1],h.pop(),b.destroy()}}function A(b){l.remove(b)}function O(){l.dispose()}return{getParameters:m,getProgramCacheKey:d,getUniforms:x,acquireProgram:N,releaseProgram:w,releaseShaderCache:A,programs:h,dispose:O}}function uf(){let i=new WeakMap;function e(o){return i.has(o)}function t(o){let s=i.get(o);return s===void 0&&(s={},i.set(o,s)),s}function r(o){i.delete(o)}function n(o,s,l){i.get(o)[s]=l}function a(){i=new WeakMap}return{has:e,get:t,remove:r,update:n,dispose:a}}function df(i,e){return i.groupOrder!==e.groupOrder?i.groupOrder-e.groupOrder:i.renderOrder!==e.renderOrder?i.renderOrder-e.renderOrder:i.material.id!==e.material.id?i.material.id-e.material.id:i.z!==e.z?i.z-e.z:i.id-e.id}function Co(i,e){return i.groupOrder!==e.groupOrder?i.groupOrder-e.groupOrder:i.renderOrder!==e.renderOrder?i.renderOrder-e.renderOrder:i.z!==e.z?e.z-i.z:i.id-e.id}function Lo(){const i=[];let e=0;const t=[],r=[],n=[];function a(){e=0,t.length=0,r.length=0,n.length=0}function o(u,p,f,g,_,m){let d=i[e];return d===void 0?(d={id:u.id,object:u,geometry:p,material:f,groupOrder:g,renderOrder:u.renderOrder,z:_,group:m},i[e]=d):(d.id=u.id,d.object=u,d.geometry=p,d.material=f,d.groupOrder=g,d.renderOrder=u.renderOrder,d.z=_,d.group=m),e++,d}function s(u,p,f,g,_,m){const d=o(u,p,f,g,_,m);f.transmission>0?r.push(d):f.transparent===!0?n.push(d):t.push(d)}function l(u,p,f,g,_,m){const d=o(u,p,f,g,_,m);f.transmission>0?r.unshift(d):f.transparent===!0?n.unshift(d):t.unshift(d)}function c(u,p){t.length>1&&t.sort(u||df),r.length>1&&r.sort(p||Co),n.length>1&&n.sort(p||Co)}function h(){for(let u=e,p=i.length;u<p;u++){const f=i[u];if(f.id===null)break;f.id=null,f.object=null,f.geometry=null,f.material=null,f.group=null}}return{opaque:t,transmissive:r,transparent:n,init:a,push:s,unshift:l,finish:h,sort:c}}function pf(){let i=new WeakMap;function e(r,n){const a=i.get(r);let o;return a===void 0?(o=new Lo,i.set(r,[o])):n>=a.length?(o=new Lo,a.push(o)):o=a[n],o}function t(){i=new WeakMap}return{get:e,dispose:t}}function ff(){const i={};return{get:function(e){if(i[e.id]!==void 0)return i[e.id];let t;switch(e.type){case"DirectionalLight":t={direction:new B,color:new We};break;case"SpotLight":t={position:new B,direction:new B,color:new We,distance:0,coneCos:0,penumbraCos:0,decay:0};break;case"PointLight":t={position:new B,color:new We,distance:0,decay:0};break;case"HemisphereLight":t={direction:new B,skyColor:new We,groundColor:new We};break;case"RectAreaLight":t={color:new We,position:new B,halfWidth:new B,halfHeight:new B};break}return i[e.id]=t,t}}}function mf(){const i={};return{get:function(e){if(i[e.id]!==void 0)return i[e.id];let t;switch(e.type){case"DirectionalLight":t={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new je};break;case"SpotLight":t={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new je};break;case"PointLight":t={shadowIntensity:1,shadowBias:0,shadowNormalBias:0,shadowRadius:1,shadowMapSize:new je,shadowCameraNear:1,shadowCameraFar:1e3};break}return i[e.id]=t,t}}}let gf=0;function _f(i,e){return(e.castShadow?2:0)-(i.castShadow?2:0)+(e.map?1:0)-(i.map?1:0)}function vf(i){const e=new ff,t=mf(),r={version:0,hash:{directionalLength:-1,pointLength:-1,spotLength:-1,rectAreaLength:-1,hemiLength:-1,numDirectionalShadows:-1,numPointShadows:-1,numSpotShadows:-1,numSpotMaps:-1,numLightProbes:-1},ambient:[0,0,0],probe:[],directional:[],directionalShadow:[],directionalShadowMap:[],directionalShadowMatrix:[],spot:[],spotLightMap:[],spotShadow:[],spotShadowMap:[],spotLightMatrix:[],rectArea:[],rectAreaLTC1:null,rectAreaLTC2:null,point:[],pointShadow:[],pointShadowMap:[],pointShadowMatrix:[],hemi:[],numSpotLightShadowsWithMaps:0,numLightProbes:0};for(let c=0;c<9;c++)r.probe.push(new B);const n=new B,a=new Qe,o=new Qe;function s(c){let h=0,u=0,p=0;for(let b=0;b<9;b++)r.probe[b].set(0,0,0);let f=0,g=0,_=0,m=0,d=0,S=0,T=0,x=0,N=0,w=0,A=0;c.sort(_f);for(let b=0,M=c.length;b<M;b++){const R=c[b],H=R.color,U=R.intensity,F=R.distance,K=R.shadow&&R.shadow.map?R.shadow.map.texture:null;if(R.isAmbientLight)h+=H.r*U,u+=H.g*U,p+=H.b*U;else if(R.isLightProbe){for(let X=0;X<9;X++)r.probe[X].addScaledVector(R.sh.coefficients[X],U);A++}else if(R.isDirectionalLight){const X=e.get(R);if(X.color.copy(R.color).multiplyScalar(R.intensity),R.castShadow){const $=R.shadow,k=t.get(R);k.shadowIntensity=$.intensity,k.shadowBias=$.bias,k.shadowNormalBias=$.normalBias,k.shadowRadius=$.radius,k.shadowMapSize=$.mapSize,r.directionalShadow[f]=k,r.directionalShadowMap[f]=K,r.directionalShadowMatrix[f]=R.shadow.matrix,S++}r.directional[f]=X,f++}else if(R.isSpotLight){const X=e.get(R);X.position.setFromMatrixPosition(R.matrixWorld),X.color.copy(H).multiplyScalar(U),X.distance=F,X.coneCos=Math.cos(R.angle),X.penumbraCos=Math.cos(R.angle*(1-R.penumbra)),X.decay=R.decay,r.spot[_]=X;const $=R.shadow;if(R.map&&(r.spotLightMap[N]=R.map,N++,$.updateMatrices(R),R.castShadow&&w++),r.spotLightMatrix[_]=$.matrix,R.castShadow){const k=t.get(R);k.shadowIntensity=$.intensity,k.shadowBias=$.bias,k.shadowNormalBias=$.normalBias,k.shadowRadius=$.radius,k.shadowMapSize=$.mapSize,r.spotShadow[_]=k,r.spotShadowMap[_]=K,x++}_++}else if(R.isRectAreaLight){const X=e.get(R);X.color.copy(H).multiplyScalar(U),X.halfWidth.set(R.width*.5,0,0),X.halfHeight.set(0,R.height*.5,0),r.rectArea[m]=X,m++}else if(R.isPointLight){const X=e.get(R);if(X.color.copy(R.color).multiplyScalar(R.intensity),X.distance=R.distance,X.decay=R.decay,R.castShadow){const $=R.shadow,k=t.get(R);k.shadowIntensity=$.intensity,k.shadowBias=$.bias,k.shadowNormalBias=$.normalBias,k.shadowRadius=$.radius,k.shadowMapSize=$.mapSize,k.shadowCameraNear=$.camera.near,k.shadowCameraFar=$.camera.far,r.pointShadow[g]=k,r.pointShadowMap[g]=K,r.pointShadowMatrix[g]=R.shadow.matrix,T++}r.point[g]=X,g++}else if(R.isHemisphereLight){const X=e.get(R);X.skyColor.copy(R.color).multiplyScalar(U),X.groundColor.copy(R.groundColor).multiplyScalar(U),r.hemi[d]=X,d++}}m>0&&(i.has("OES_texture_float_linear")===!0?(r.rectAreaLTC1=fe.LTC_FLOAT_1,r.rectAreaLTC2=fe.LTC_FLOAT_2):(r.rectAreaLTC1=fe.LTC_HALF_1,r.rectAreaLTC2=fe.LTC_HALF_2)),r.ambient[0]=h,r.ambient[1]=u,r.ambient[2]=p;const O=r.hash;(O.directionalLength!==f||O.pointLength!==g||O.spotLength!==_||O.rectAreaLength!==m||O.hemiLength!==d||O.numDirectionalShadows!==S||O.numPointShadows!==T||O.numSpotShadows!==x||O.numSpotMaps!==N||O.numLightProbes!==A)&&(r.directional.length=f,r.spot.length=_,r.rectArea.length=m,r.point.length=g,r.hemi.length=d,r.directionalShadow.length=S,r.directionalShadowMap.length=S,r.pointShadow.length=T,r.pointShadowMap.length=T,r.spotShadow.length=x,r.spotShadowMap.length=x,r.directionalShadowMatrix.length=S,r.pointShadowMatrix.length=T,r.spotLightMatrix.length=x+N-w,r.spotLightMap.length=N,r.numSpotLightShadowsWithMaps=w,r.numLightProbes=A,O.directionalLength=f,O.pointLength=g,O.spotLength=_,O.rectAreaLength=m,O.hemiLength=d,O.numDirectionalShadows=S,O.numPointShadows=T,O.numSpotShadows=x,O.numSpotMaps=N,O.numLightProbes=A,r.version=gf++)}function l(c,h){let u=0,p=0,f=0,g=0,_=0;const m=h.matrixWorldInverse;for(let d=0,S=c.length;d<S;d++){const T=c[d];if(T.isDirectionalLight){const x=r.directional[u];x.direction.setFromMatrixPosition(T.matrixWorld),n.setFromMatrixPosition(T.target.matrixWorld),x.direction.sub(n),x.direction.transformDirection(m),u++}else if(T.isSpotLight){const x=r.spot[f];x.position.setFromMatrixPosition(T.matrixWorld),x.position.applyMatrix4(m),x.direction.setFromMatrixPosition(T.matrixWorld),n.setFromMatrixPosition(T.target.matrixWorld),x.direction.sub(n),x.direction.transformDirection(m),f++}else if(T.isRectAreaLight){const x=r.rectArea[g];x.position.setFromMatrixPosition(T.matrixWorld),x.position.applyMatrix4(m),o.identity(),a.copy(T.matrixWorld),a.premultiply(m),o.extractRotation(a),x.halfWidth.set(T.width*.5,0,0),x.halfHeight.set(0,T.height*.5,0),x.halfWidth.applyMatrix4(o),x.halfHeight.applyMatrix4(o),g++}else if(T.isPointLight){const x=r.point[p];x.position.setFromMatrixPosition(T.matrixWorld),x.position.applyMatrix4(m),p++}else if(T.isHemisphereLight){const x=r.hemi[_];x.direction.setFromMatrixPosition(T.matrixWorld),x.direction.transformDirection(m),_++}}}return{setup:s,setupView:l,state:r}}function Po(i){const e=new vf(i),t=[],r=[];function n(h){c.camera=h,t.length=0,r.length=0}function a(h){t.push(h)}function o(h){r.push(h)}function s(){e.setup(t)}function l(h){e.setupView(t,h)}const c={lightsArray:t,shadowsArray:r,camera:null,lights:e,transmissionRenderTarget:{}};return{init:n,state:c,setupLights:s,setupLightsView:l,pushLight:a,pushShadow:o}}function xf(i){let e=new WeakMap;function t(n,a=0){const o=e.get(n);let s;return o===void 0?(s=new Po(i),e.set(n,[s])):a>=o.length?(s=new Po(i),o.push(s)):s=o[a],s}function r(){e=new WeakMap}return{get:t,dispose:r}}class yf extends Gr{static get type(){return"MeshDepthMaterial"}constructor(e){super(),this.isMeshDepthMaterial=!0,this.depthPacking=mc,this.map=null,this.alphaMap=null,this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.wireframe=!1,this.wireframeLinewidth=1,this.setValues(e)}copy(e){return super.copy(e),this.depthPacking=e.depthPacking,this.map=e.map,this.alphaMap=e.alphaMap,this.displacementMap=e.displacementMap,this.displacementScale=e.displacementScale,this.displacementBias=e.displacementBias,this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this}}class Mf extends Gr{static get type(){return"MeshDistanceMaterial"}constructor(e){super(),this.isMeshDistanceMaterial=!0,this.map=null,this.alphaMap=null,this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.setValues(e)}copy(e){return super.copy(e),this.map=e.map,this.alphaMap=e.alphaMap,this.displacementMap=e.displacementMap,this.displacementScale=e.displacementScale,this.displacementBias=e.displacementBias,this}}const Sf=`void main() {
	gl_Position = vec4( position, 1.0 );
}`,Ef=`uniform sampler2D shadow_pass;
uniform vec2 resolution;
uniform float radius;
#include <packing>
void main() {
	const float samples = float( VSM_SAMPLES );
	float mean = 0.0;
	float squared_mean = 0.0;
	float uvStride = samples <= 1.0 ? 0.0 : 2.0 / ( samples - 1.0 );
	float uvStart = samples <= 1.0 ? 0.0 : - 1.0;
	for ( float i = 0.0; i < samples; i ++ ) {
		float uvOffset = uvStart + i * uvStride;
		#ifdef HORIZONTAL_PASS
			vec2 distribution = unpackRGBATo2Half( texture2D( shadow_pass, ( gl_FragCoord.xy + vec2( uvOffset, 0.0 ) * radius ) / resolution ) );
			mean += distribution.x;
			squared_mean += distribution.y * distribution.y + distribution.x * distribution.x;
		#else
			float depth = unpackRGBAToDepth( texture2D( shadow_pass, ( gl_FragCoord.xy + vec2( 0.0, uvOffset ) * radius ) / resolution ) );
			mean += depth;
			squared_mean += depth * depth;
		#endif
	}
	mean = mean / samples;
	squared_mean = squared_mean / samples;
	float std_dev = sqrt( squared_mean - mean * mean );
	gl_FragColor = pack2HalfToRGBA( vec2( mean, std_dev ) );
}`;function bf(i,e,t){let r=new ys;const n=new je,a=new je,o=new dt,s=new yf({depthPacking:gc}),l=new Mf,c={},h=t.maxTextureSize,u={[Sr]:Dt,[Dt]:Sr,[Rt]:Rt},p=new Er({defines:{VSM_SAMPLES:8},uniforms:{shadow_pass:{value:null},resolution:{value:new je},radius:{value:4}},vertexShader:Sf,fragmentShader:Ef}),f=p.clone();f.defines.HORIZONTAL_PASS=1;const g=new Ct;g.setAttribute("position",new Bt(new Float32Array([-1,-1,.5,3,-1,.5,-1,3,.5]),3));const _=new Fe(g,p),m=this;this.enabled=!1,this.autoUpdate=!0,this.needsUpdate=!1,this.type=qo;let d=this.type;this.render=function(w,A,O){if(m.enabled===!1||m.autoUpdate===!1&&m.needsUpdate===!1||w.length===0)return;const b=i.getRenderTarget(),M=i.getActiveCubeFace(),R=i.getActiveMipmapLevel(),H=i.state;H.setBlending(xr),H.buffers.color.setClear(1,1,1,1),H.buffers.depth.setTest(!0),H.setScissorTest(!1);const U=d!==sr&&this.type===sr,F=d===sr&&this.type!==sr;for(let K=0,X=w.length;K<X;K++){const $=w[K],k=$.shadow;if(k===void 0){console.warn("THREE.WebGLShadowMap:",$,"has no shadow.");continue}if(k.autoUpdate===!1&&k.needsUpdate===!1)continue;n.copy(k.mapSize);const Z=k.getFrameExtents();if(n.multiply(Z),a.copy(k.mapSize),(n.x>h||n.y>h)&&(n.x>h&&(a.x=Math.floor(h/Z.x),n.x=a.x*Z.x,k.mapSize.x=a.x),n.y>h&&(a.y=Math.floor(h/Z.y),n.y=a.y*Z.y,k.mapSize.y=a.y)),k.map===null||U===!0||F===!0){const ve=this.type!==sr?{minFilter:pt,magFilter:pt}:{};k.map!==null&&k.map.dispose(),k.map=new Vr(n.x,n.y,ve),k.map.texture.name=$.name+".shadowMap",k.camera.updateProjectionMatrix()}i.setRenderTarget(k.map),i.clear();const ce=k.getViewportCount();for(let ve=0;ve<ce;ve++){const De=k.getViewport(ve);o.set(a.x*De.x,a.y*De.y,a.x*De.z,a.y*De.w),H.viewport(o),k.updateMatrices($,ve),r=k.getFrustum(),x(A,O,k.camera,$,this.type)}k.isPointLightShadow!==!0&&this.type===sr&&S(k,O),k.needsUpdate=!1}d=this.type,m.needsUpdate=!1,i.setRenderTarget(b,M,R)};function S(w,A){const O=e.update(_);p.defines.VSM_SAMPLES!==w.blurSamples&&(p.defines.VSM_SAMPLES=w.blurSamples,f.defines.VSM_SAMPLES=w.blurSamples,p.needsUpdate=!0,f.needsUpdate=!0),w.mapPass===null&&(w.mapPass=new Vr(n.x,n.y)),p.uniforms.shadow_pass.value=w.map.texture,p.uniforms.resolution.value=w.mapSize,p.uniforms.radius.value=w.radius,i.setRenderTarget(w.mapPass),i.clear(),i.renderBufferDirect(A,null,O,p,_,null),f.uniforms.shadow_pass.value=w.mapPass.texture,f.uniforms.resolution.value=w.mapSize,f.uniforms.radius.value=w.radius,i.setRenderTarget(w.map),i.clear(),i.renderBufferDirect(A,null,O,f,_,null)}function T(w,A,O,b){let M=null;const R=O.isPointLight===!0?w.customDistanceMaterial:w.customDepthMaterial;if(R!==void 0)M=R;else if(M=O.isPointLight===!0?l:s,i.localClippingEnabled&&A.clipShadows===!0&&Array.isArray(A.clippingPlanes)&&A.clippingPlanes.length!==0||A.displacementMap&&A.displacementScale!==0||A.alphaMap&&A.alphaTest>0||A.map&&A.alphaTest>0){const H=M.uuid,U=A.uuid;let F=c[H];F===void 0&&(F={},c[H]=F);let K=F[U];K===void 0&&(K=M.clone(),F[U]=K,A.addEventListener("dispose",N)),M=K}if(M.visible=A.visible,M.wireframe=A.wireframe,b===sr?M.side=A.shadowSide!==null?A.shadowSide:A.side:M.side=A.shadowSide!==null?A.shadowSide:u[A.side],M.alphaMap=A.alphaMap,M.alphaTest=A.alphaTest,M.map=A.map,M.clipShadows=A.clipShadows,M.clippingPlanes=A.clippingPlanes,M.clipIntersection=A.clipIntersection,M.displacementMap=A.displacementMap,M.displacementScale=A.displacementScale,M.displacementBias=A.displacementBias,M.wireframeLinewidth=A.wireframeLinewidth,M.linewidth=A.linewidth,O.isPointLight===!0&&M.isMeshDistanceMaterial===!0){const H=i.properties.get(M);H.light=O}return M}function x(w,A,O,b,M){if(w.visible===!1)return;if(w.layers.test(A.layers)&&(w.isMesh||w.isLine||w.isPoints)&&(w.castShadow||w.receiveShadow&&M===sr)&&(!w.frustumCulled||r.intersectsObject(w))){w.modelViewMatrix.multiplyMatrices(O.matrixWorldInverse,w.matrixWorld);const H=e.update(w),U=w.material;if(Array.isArray(U)){const F=H.groups;for(let K=0,X=F.length;K<X;K++){const $=F[K],k=U[$.materialIndex];if(k&&k.visible){const Z=T(w,k,b,M);w.onBeforeShadow(i,w,A,O,H,Z,$),i.renderBufferDirect(O,null,H,Z,w,$),w.onAfterShadow(i,w,A,O,H,Z,$)}}}else if(U.visible){const F=T(w,U,b,M);w.onBeforeShadow(i,w,A,O,H,F,null),i.renderBufferDirect(O,null,H,F,w,null),w.onAfterShadow(i,w,A,O,H,F,null)}}const R=w.children;for(let H=0,U=R.length;H<U;H++)x(R[H],A,O,b,M)}function N(w){w.target.removeEventListener("dispose",N);for(const A in c){const O=c[A],b=w.target.uuid;b in O&&(O[b].dispose(),delete O[b])}}}const Tf={[Sa]:Ea,[ba]:Aa,[Ta]:Ra,[fi]:wa,[Ea]:Sa,[Aa]:ba,[Ra]:Ta,[wa]:fi};function wf(i,e){function t(){let D=!1;const me=new dt;let Y=null;const te=new dt(0,0,0,0);return{setMask:function(ye){Y!==ye&&!D&&(i.colorMask(ye,ye,ye,ye),Y=ye)},setLocked:function(ye){D=ye},setClear:function(ye,de,Ye,ut,St){St===!0&&(ye*=ut,de*=ut,Ye*=ut),me.set(ye,de,Ye,ut),te.equals(me)===!1&&(i.clearColor(ye,de,Ye,ut),te.copy(me))},reset:function(){D=!1,Y=null,te.set(-1,0,0,0)}}}function r(){let D=!1,me=!1,Y=null,te=null,ye=null;return{setReversed:function(de){if(me!==de){const Ye=e.get("EXT_clip_control");me?Ye.clipControlEXT(Ye.LOWER_LEFT_EXT,Ye.ZERO_TO_ONE_EXT):Ye.clipControlEXT(Ye.LOWER_LEFT_EXT,Ye.NEGATIVE_ONE_TO_ONE_EXT);const ut=ye;ye=null,this.setClear(ut)}me=de},getReversed:function(){return me},setTest:function(de){de?z(i.DEPTH_TEST):ae(i.DEPTH_TEST)},setMask:function(de){Y!==de&&!D&&(i.depthMask(de),Y=de)},setFunc:function(de){if(me&&(de=Tf[de]),te!==de){switch(de){case Sa:i.depthFunc(i.NEVER);break;case Ea:i.depthFunc(i.ALWAYS);break;case ba:i.depthFunc(i.LESS);break;case fi:i.depthFunc(i.LEQUAL);break;case Ta:i.depthFunc(i.EQUAL);break;case wa:i.depthFunc(i.GEQUAL);break;case Aa:i.depthFunc(i.GREATER);break;case Ra:i.depthFunc(i.NOTEQUAL);break;default:i.depthFunc(i.LEQUAL)}te=de}},setLocked:function(de){D=de},setClear:function(de){ye!==de&&(me&&(de=1-de),i.clearDepth(de),ye=de)},reset:function(){D=!1,Y=null,te=null,ye=null,me=!1}}}function n(){let D=!1,me=null,Y=null,te=null,ye=null,de=null,Ye=null,ut=null,St=null;return{setTest:function(et){D||(et?z(i.STENCIL_TEST):ae(i.STENCIL_TEST))},setMask:function(et){me!==et&&!D&&(i.stencilMask(et),me=et)},setFunc:function(et,kt,er){(Y!==et||te!==kt||ye!==er)&&(i.stencilFunc(et,kt,er),Y=et,te=kt,ye=er)},setOp:function(et,kt,er){(de!==et||Ye!==kt||ut!==er)&&(i.stencilOp(et,kt,er),de=et,Ye=kt,ut=er)},setLocked:function(et){D=et},setClear:function(et){St!==et&&(i.clearStencil(et),St=et)},reset:function(){D=!1,me=null,Y=null,te=null,ye=null,de=null,Ye=null,ut=null,St=null}}}const a=new t,o=new r,s=new n,l=new WeakMap,c=new WeakMap;let h={},u={},p=new WeakMap,f=[],g=null,_=!1,m=null,d=null,S=null,T=null,x=null,N=null,w=null,A=new We(0,0,0),O=0,b=!1,M=null,R=null,H=null,U=null,F=null;const K=i.getParameter(i.MAX_COMBINED_TEXTURE_IMAGE_UNITS);let X=!1,$=0;const k=i.getParameter(i.VERSION);k.indexOf("WebGL")!==-1?($=parseFloat(/^WebGL (\d)/.exec(k)[1]),X=$>=1):k.indexOf("OpenGL ES")!==-1&&($=parseFloat(/^OpenGL ES (\d)/.exec(k)[1]),X=$>=2);let Z=null,ce={};const ve=i.getParameter(i.SCISSOR_BOX),De=i.getParameter(i.VIEWPORT),xe=new dt().fromArray(ve),G=new dt().fromArray(De);function J(D,me,Y,te){const ye=new Uint8Array(4),de=i.createTexture();i.bindTexture(D,de),i.texParameteri(D,i.TEXTURE_MIN_FILTER,i.NEAREST),i.texParameteri(D,i.TEXTURE_MAG_FILTER,i.NEAREST);for(let Ye=0;Ye<Y;Ye++)D===i.TEXTURE_3D||D===i.TEXTURE_2D_ARRAY?i.texImage3D(me,0,i.RGBA,1,1,te,0,i.RGBA,i.UNSIGNED_BYTE,ye):i.texImage2D(me+Ye,0,i.RGBA,1,1,0,i.RGBA,i.UNSIGNED_BYTE,ye);return de}const q={};q[i.TEXTURE_2D]=J(i.TEXTURE_2D,i.TEXTURE_2D,1),q[i.TEXTURE_CUBE_MAP]=J(i.TEXTURE_CUBE_MAP,i.TEXTURE_CUBE_MAP_POSITIVE_X,6),q[i.TEXTURE_2D_ARRAY]=J(i.TEXTURE_2D_ARRAY,i.TEXTURE_2D_ARRAY,1,1),q[i.TEXTURE_3D]=J(i.TEXTURE_3D,i.TEXTURE_3D,1,1),a.setClear(0,0,0,1),o.setClear(1),s.setClear(0),z(i.DEPTH_TEST),o.setFunc(fi),ue(!1),pe(Us),z(i.CULL_FACE),C(xr);function z(D){h[D]!==!0&&(i.enable(D),h[D]=!0)}function ae(D){h[D]!==!1&&(i.disable(D),h[D]=!1)}function ne(D,me){return u[D]!==me?(i.bindFramebuffer(D,me),u[D]=me,D===i.DRAW_FRAMEBUFFER&&(u[i.FRAMEBUFFER]=me),D===i.FRAMEBUFFER&&(u[i.DRAW_FRAMEBUFFER]=me),!0):!1}function oe(D,me){let Y=f,te=!1;if(D){Y=p.get(me),Y===void 0&&(Y=[],p.set(me,Y));const ye=D.textures;if(Y.length!==ye.length||Y[0]!==i.COLOR_ATTACHMENT0){for(let de=0,Ye=ye.length;de<Ye;de++)Y[de]=i.COLOR_ATTACHMENT0+de;Y.length=ye.length,te=!0}}else Y[0]!==i.BACK&&(Y[0]=i.BACK,te=!0);te&&i.drawBuffers(Y)}function Ae(D){return g!==D?(i.useProgram(D),g=D,!0):!1}const Te={[Or]:i.FUNC_ADD,[Hl]:i.FUNC_SUBTRACT,[Gl]:i.FUNC_REVERSE_SUBTRACT};Te[Wl]=i.MIN,Te[jl]=i.MAX;const Be={[Xl]:i.ZERO,[Yl]:i.ONE,[ql]:i.SRC_COLOR,[ya]:i.SRC_ALPHA,[ec]:i.SRC_ALPHA_SATURATE,[Jl]:i.DST_COLOR,[$l]:i.DST_ALPHA,[Kl]:i.ONE_MINUS_SRC_COLOR,[Ma]:i.ONE_MINUS_SRC_ALPHA,[Ql]:i.ONE_MINUS_DST_COLOR,[Zl]:i.ONE_MINUS_DST_ALPHA,[tc]:i.CONSTANT_COLOR,[rc]:i.ONE_MINUS_CONSTANT_COLOR,[ic]:i.CONSTANT_ALPHA,[nc]:i.ONE_MINUS_CONSTANT_ALPHA};function C(D,me,Y,te,ye,de,Ye,ut,St,et){if(D===xr){_===!0&&(ae(i.BLEND),_=!1);return}if(_===!1&&(z(i.BLEND),_=!0),D!==Vl){if(D!==m||et!==b){if((d!==Or||x!==Or)&&(i.blendEquation(i.FUNC_ADD),d=Or,x=Or),et)switch(D){case ui:i.blendFuncSeparate(i.ONE,i.ONE_MINUS_SRC_ALPHA,i.ONE,i.ONE_MINUS_SRC_ALPHA);break;case Os:i.blendFunc(i.ONE,i.ONE);break;case Fs:i.blendFuncSeparate(i.ZERO,i.ONE_MINUS_SRC_COLOR,i.ZERO,i.ONE);break;case zs:i.blendFuncSeparate(i.ZERO,i.SRC_COLOR,i.ZERO,i.SRC_ALPHA);break;default:console.error("THREE.WebGLState: Invalid blending: ",D);break}else switch(D){case ui:i.blendFuncSeparate(i.SRC_ALPHA,i.ONE_MINUS_SRC_ALPHA,i.ONE,i.ONE_MINUS_SRC_ALPHA);break;case Os:i.blendFunc(i.SRC_ALPHA,i.ONE);break;case Fs:i.blendFuncSeparate(i.ZERO,i.ONE_MINUS_SRC_COLOR,i.ZERO,i.ONE);break;case zs:i.blendFunc(i.ZERO,i.SRC_COLOR);break;default:console.error("THREE.WebGLState: Invalid blending: ",D);break}S=null,T=null,N=null,w=null,A.set(0,0,0),O=0,m=D,b=et}return}ye=ye||me,de=de||Y,Ye=Ye||te,(me!==d||ye!==x)&&(i.blendEquationSeparate(Te[me],Te[ye]),d=me,x=ye),(Y!==S||te!==T||de!==N||Ye!==w)&&(i.blendFuncSeparate(Be[Y],Be[te],Be[de],Be[Ye]),S=Y,T=te,N=de,w=Ye),(ut.equals(A)===!1||St!==O)&&(i.blendColor(ut.r,ut.g,ut.b,St),A.copy(ut),O=St),m=D,b=!1}function Je(D,me){D.side===Rt?ae(i.CULL_FACE):z(i.CULL_FACE);let Y=D.side===Dt;me&&(Y=!Y),ue(Y),D.blending===ui&&D.transparent===!1?C(xr):C(D.blending,D.blendEquation,D.blendSrc,D.blendDst,D.blendEquationAlpha,D.blendSrcAlpha,D.blendDstAlpha,D.blendColor,D.blendAlpha,D.premultipliedAlpha),o.setFunc(D.depthFunc),o.setTest(D.depthTest),o.setMask(D.depthWrite),a.setMask(D.colorWrite);const te=D.stencilWrite;s.setTest(te),te&&(s.setMask(D.stencilWriteMask),s.setFunc(D.stencilFunc,D.stencilRef,D.stencilFuncMask),s.setOp(D.stencilFail,D.stencilZFail,D.stencilZPass)),le(D.polygonOffset,D.polygonOffsetFactor,D.polygonOffsetUnits),D.alphaToCoverage===!0?z(i.SAMPLE_ALPHA_TO_COVERAGE):ae(i.SAMPLE_ALPHA_TO_COVERAGE)}function ue(D){M!==D&&(D?i.frontFace(i.CW):i.frontFace(i.CCW),M=D)}function pe(D){D!==zl?(z(i.CULL_FACE),D!==R&&(D===Us?i.cullFace(i.BACK):D===Bl?i.cullFace(i.FRONT):i.cullFace(i.FRONT_AND_BACK))):ae(i.CULL_FACE),R=D}function ie(D){D!==H&&(X&&i.lineWidth(D),H=D)}function le(D,me,Y){D?(z(i.POLYGON_OFFSET_FILL),(U!==me||F!==Y)&&(i.polygonOffset(me,Y),U=me,F=Y)):ae(i.POLYGON_OFFSET_FILL)}function ge(D){D?z(i.SCISSOR_TEST):ae(i.SCISSOR_TEST)}function E(D){D===void 0&&(D=i.TEXTURE0+K-1),Z!==D&&(i.activeTexture(D),Z=D)}function v(D,me,Y){Y===void 0&&(Z===null?Y=i.TEXTURE0+K-1:Y=Z);let te=ce[Y];te===void 0&&(te={type:void 0,texture:void 0},ce[Y]=te),(te.type!==D||te.texture!==me)&&(Z!==Y&&(i.activeTexture(Y),Z=Y),i.bindTexture(D,me||q[D]),te.type=D,te.texture=me)}function V(){const D=ce[Z];D!==void 0&&D.type!==void 0&&(i.bindTexture(D.type,null),D.type=void 0,D.texture=void 0)}function Q(){try{i.compressedTexImage2D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function re(){try{i.compressedTexImage3D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function ee(){try{i.texSubImage2D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function Re(){try{i.texSubImage3D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function _e(){try{i.compressedTexSubImage2D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function be(){try{i.compressedTexSubImage3D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function qe(){try{i.texStorage2D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function he(){try{i.texStorage3D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function Se(){try{i.texImage2D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function Ue(){try{i.texImage3D.apply(i,arguments)}catch(D){console.error("THREE.WebGLState:",D)}}function Oe(D){xe.equals(D)===!1&&(i.scissor(D.x,D.y,D.z,D.w),xe.copy(D))}function Ee(D){G.equals(D)===!1&&(i.viewport(D.x,D.y,D.z,D.w),G.copy(D))}function Xe(D,me){let Y=c.get(me);Y===void 0&&(Y=new WeakMap,c.set(me,Y));let te=Y.get(D);te===void 0&&(te=i.getUniformBlockIndex(me,D.name),Y.set(D,te))}function He(D,me){const Y=c.get(me).get(D);l.get(me)!==Y&&(i.uniformBlockBinding(me,Y,D.__bindingPointIndex),l.set(me,Y))}function nt(){i.disable(i.BLEND),i.disable(i.CULL_FACE),i.disable(i.DEPTH_TEST),i.disable(i.POLYGON_OFFSET_FILL),i.disable(i.SCISSOR_TEST),i.disable(i.STENCIL_TEST),i.disable(i.SAMPLE_ALPHA_TO_COVERAGE),i.blendEquation(i.FUNC_ADD),i.blendFunc(i.ONE,i.ZERO),i.blendFuncSeparate(i.ONE,i.ZERO,i.ONE,i.ZERO),i.blendColor(0,0,0,0),i.colorMask(!0,!0,!0,!0),i.clearColor(0,0,0,0),i.depthMask(!0),i.depthFunc(i.LESS),o.setReversed(!1),i.clearDepth(1),i.stencilMask(4294967295),i.stencilFunc(i.ALWAYS,0,4294967295),i.stencilOp(i.KEEP,i.KEEP,i.KEEP),i.clearStencil(0),i.cullFace(i.BACK),i.frontFace(i.CCW),i.polygonOffset(0,0),i.activeTexture(i.TEXTURE0),i.bindFramebuffer(i.FRAMEBUFFER,null),i.bindFramebuffer(i.DRAW_FRAMEBUFFER,null),i.bindFramebuffer(i.READ_FRAMEBUFFER,null),i.useProgram(null),i.lineWidth(1),i.scissor(0,0,i.canvas.width,i.canvas.height),i.viewport(0,0,i.canvas.width,i.canvas.height),h={},Z=null,ce={},u={},p=new WeakMap,f=[],g=null,_=!1,m=null,d=null,S=null,T=null,x=null,N=null,w=null,A=new We(0,0,0),O=0,b=!1,M=null,R=null,H=null,U=null,F=null,xe.set(0,0,i.canvas.width,i.canvas.height),G.set(0,0,i.canvas.width,i.canvas.height),a.reset(),o.reset(),s.reset()}return{buffers:{color:a,depth:o,stencil:s},enable:z,disable:ae,bindFramebuffer:ne,drawBuffers:oe,useProgram:Ae,setBlending:C,setMaterial:Je,setFlipSided:ue,setCullFace:pe,setLineWidth:ie,setPolygonOffset:le,setScissorTest:ge,activeTexture:E,bindTexture:v,unbindTexture:V,compressedTexImage2D:Q,compressedTexImage3D:re,texImage2D:Se,texImage3D:Ue,updateUBOMapping:Xe,uniformBlockBinding:He,texStorage2D:qe,texStorage3D:he,texSubImage2D:ee,texSubImage3D:Re,compressedTexSubImage2D:_e,compressedTexSubImage3D:be,scissor:Oe,viewport:Ee,reset:nt}}function Do(i,e,t,r){const n=Af(r);switch(t){case Qo:return i*e;case tl:return i*e;case rl:return i*e*2;case gs:return i*e/n.components*n.byteLength;case _s:return i*e/n.components*n.byteLength;case il:return i*e*2/n.components*n.byteLength;case vs:return i*e*2/n.components*n.byteLength;case el:return i*e*3/n.components*n.byteLength;case Xt:return i*e*4/n.components*n.byteLength;case xs:return i*e*4/n.components*n.byteLength;case Tn:case wn:return Math.floor((i+3)/4)*Math.floor((e+3)/4)*8;case An:case Rn:return Math.floor((i+3)/4)*Math.floor((e+3)/4)*16;case Na:case Oa:return Math.max(i,16)*Math.max(e,8)/4;case Ia:case Ua:return Math.max(i,8)*Math.max(e,8)/2;case Fa:case za:return Math.floor((i+3)/4)*Math.floor((e+3)/4)*8;case Ba:return Math.floor((i+3)/4)*Math.floor((e+3)/4)*16;case ka:return Math.floor((i+3)/4)*Math.floor((e+3)/4)*16;case Va:return Math.floor((i+4)/5)*Math.floor((e+3)/4)*16;case Ha:return Math.floor((i+4)/5)*Math.floor((e+4)/5)*16;case Ga:return Math.floor((i+5)/6)*Math.floor((e+4)/5)*16;case Wa:return Math.floor((i+5)/6)*Math.floor((e+5)/6)*16;case ja:return Math.floor((i+7)/8)*Math.floor((e+4)/5)*16;case Xa:return Math.floor((i+7)/8)*Math.floor((e+5)/6)*16;case Ya:return Math.floor((i+7)/8)*Math.floor((e+7)/8)*16;case qa:return Math.floor((i+9)/10)*Math.floor((e+4)/5)*16;case Ka:return Math.floor((i+9)/10)*Math.floor((e+5)/6)*16;case $a:return Math.floor((i+9)/10)*Math.floor((e+7)/8)*16;case Za:return Math.floor((i+9)/10)*Math.floor((e+9)/10)*16;case Ja:return Math.floor((i+11)/12)*Math.floor((e+9)/10)*16;case Qa:return Math.floor((i+11)/12)*Math.floor((e+11)/12)*16;case Cn:case es:case ts:return Math.ceil(i/4)*Math.ceil(e/4)*16;case nl:case rs:return Math.ceil(i/4)*Math.ceil(e/4)*8;case is:case ns:return Math.ceil(i/4)*Math.ceil(e/4)*16}throw new Error(`Unable to determine texture byte length for ${t} format.`)}function Af(i){switch(i){case cr:case $o:return{byteLength:1,components:1};case ki:case Zo:case Hi:return{byteLength:2,components:1};case fs:case ms:return{byteLength:2,components:4};case kr:case ps:case Qt:return{byteLength:4,components:1};case Jo:return{byteLength:4,components:3}}throw new Error(`Unknown texture type ${i}.`)}function Rf(i,e,t,r,n,a,o){const s=e.has("WEBGL_multisampled_render_to_texture")?e.get("WEBGL_multisampled_render_to_texture"):null,l=typeof navigator>"u"?!1:/OculusBrowser/g.test(navigator.userAgent),c=new je,h=new WeakMap;let u;const p=new WeakMap;let f=!1;try{f=typeof OffscreenCanvas<"u"&&new OffscreenCanvas(1,1).getContext("2d")!==null}catch{}function g(E,v){return f?new OffscreenCanvas(E,v):Vi("canvas")}function _(E,v,V){let Q=1;const re=ge(E);if((re.width>V||re.height>V)&&(Q=V/Math.max(re.width,re.height)),Q<1)if(typeof HTMLImageElement<"u"&&E instanceof HTMLImageElement||typeof HTMLCanvasElement<"u"&&E instanceof HTMLCanvasElement||typeof ImageBitmap<"u"&&E instanceof ImageBitmap||typeof VideoFrame<"u"&&E instanceof VideoFrame){const ee=Math.floor(Q*re.width),Re=Math.floor(Q*re.height);u===void 0&&(u=g(ee,Re));const _e=v?g(ee,Re):u;return _e.width=ee,_e.height=Re,_e.getContext("2d").drawImage(E,0,0,ee,Re),console.warn("THREE.WebGLRenderer: Texture has been resized from ("+re.width+"x"+re.height+") to ("+ee+"x"+Re+")."),_e}else return"data"in E&&console.warn("THREE.WebGLRenderer: Image in DataTexture is too big ("+re.width+"x"+re.height+")."),E;return E}function m(E){return E.generateMipmaps}function d(E){i.generateMipmap(E)}function S(E){return E.isWebGLCubeRenderTarget?i.TEXTURE_CUBE_MAP:E.isWebGL3DRenderTarget?i.TEXTURE_3D:E.isWebGLArrayRenderTarget||E.isCompressedArrayTexture?i.TEXTURE_2D_ARRAY:i.TEXTURE_2D}function T(E,v,V,Q,re=!1){if(E!==null){if(i[E]!==void 0)return i[E];console.warn("THREE.WebGLRenderer: Attempt to use non-existing WebGL internal format '"+E+"'")}let ee=v;if(v===i.RED&&(V===i.FLOAT&&(ee=i.R32F),V===i.HALF_FLOAT&&(ee=i.R16F),V===i.UNSIGNED_BYTE&&(ee=i.R8)),v===i.RED_INTEGER&&(V===i.UNSIGNED_BYTE&&(ee=i.R8UI),V===i.UNSIGNED_SHORT&&(ee=i.R16UI),V===i.UNSIGNED_INT&&(ee=i.R32UI),V===i.BYTE&&(ee=i.R8I),V===i.SHORT&&(ee=i.R16I),V===i.INT&&(ee=i.R32I)),v===i.RG&&(V===i.FLOAT&&(ee=i.RG32F),V===i.HALF_FLOAT&&(ee=i.RG16F),V===i.UNSIGNED_BYTE&&(ee=i.RG8)),v===i.RG_INTEGER&&(V===i.UNSIGNED_BYTE&&(ee=i.RG8UI),V===i.UNSIGNED_SHORT&&(ee=i.RG16UI),V===i.UNSIGNED_INT&&(ee=i.RG32UI),V===i.BYTE&&(ee=i.RG8I),V===i.SHORT&&(ee=i.RG16I),V===i.INT&&(ee=i.RG32I)),v===i.RGB_INTEGER&&(V===i.UNSIGNED_BYTE&&(ee=i.RGB8UI),V===i.UNSIGNED_SHORT&&(ee=i.RGB16UI),V===i.UNSIGNED_INT&&(ee=i.RGB32UI),V===i.BYTE&&(ee=i.RGB8I),V===i.SHORT&&(ee=i.RGB16I),V===i.INT&&(ee=i.RGB32I)),v===i.RGBA_INTEGER&&(V===i.UNSIGNED_BYTE&&(ee=i.RGBA8UI),V===i.UNSIGNED_SHORT&&(ee=i.RGBA16UI),V===i.UNSIGNED_INT&&(ee=i.RGBA32UI),V===i.BYTE&&(ee=i.RGBA8I),V===i.SHORT&&(ee=i.RGBA16I),V===i.INT&&(ee=i.RGBA32I)),v===i.RGB&&V===i.UNSIGNED_INT_5_9_9_9_REV&&(ee=i.RGB9_E5),v===i.RGBA){const Re=re?On:$e.getTransfer(Q);V===i.FLOAT&&(ee=i.RGBA32F),V===i.HALF_FLOAT&&(ee=i.RGBA16F),V===i.UNSIGNED_BYTE&&(ee=Re===rt?i.SRGB8_ALPHA8:i.RGBA8),V===i.UNSIGNED_SHORT_4_4_4_4&&(ee=i.RGBA4),V===i.UNSIGNED_SHORT_5_5_5_1&&(ee=i.RGB5_A1)}return(ee===i.R16F||ee===i.R32F||ee===i.RG16F||ee===i.RG32F||ee===i.RGBA16F||ee===i.RGBA32F)&&e.get("EXT_color_buffer_float"),ee}function x(E,v){let V;return E?v===null||v===kr||v===_i?V=i.DEPTH24_STENCIL8:v===Qt?V=i.DEPTH32F_STENCIL8:v===ki&&(V=i.DEPTH24_STENCIL8,console.warn("DepthTexture: 16 bit depth attachment is not supported with stencil. Using 24-bit attachment.")):v===null||v===kr||v===_i?V=i.DEPTH_COMPONENT24:v===Qt?V=i.DEPTH_COMPONENT32F:v===ki&&(V=i.DEPTH_COMPONENT16),V}function N(E,v){return m(E)===!0||E.isFramebufferTexture&&E.minFilter!==pt&&E.minFilter!==jt?Math.log2(Math.max(v.width,v.height))+1:E.mipmaps!==void 0&&E.mipmaps.length>0?E.mipmaps.length:E.isCompressedTexture&&Array.isArray(E.image)?v.mipmaps.length:1}function w(E){const v=E.target;v.removeEventListener("dispose",w),O(v),v.isVideoTexture&&h.delete(v)}function A(E){const v=E.target;v.removeEventListener("dispose",A),M(v)}function O(E){const v=r.get(E);if(v.__webglInit===void 0)return;const V=E.source,Q=p.get(V);if(Q){const re=Q[v.__cacheKey];re.usedTimes--,re.usedTimes===0&&b(E),Object.keys(Q).length===0&&p.delete(V)}r.remove(E)}function b(E){const v=r.get(E);i.deleteTexture(v.__webglTexture);const V=E.source,Q=p.get(V);delete Q[v.__cacheKey],o.memory.textures--}function M(E){const v=r.get(E);if(E.depthTexture&&(E.depthTexture.dispose(),r.remove(E.depthTexture)),E.isWebGLCubeRenderTarget)for(let Q=0;Q<6;Q++){if(Array.isArray(v.__webglFramebuffer[Q]))for(let re=0;re<v.__webglFramebuffer[Q].length;re++)i.deleteFramebuffer(v.__webglFramebuffer[Q][re]);else i.deleteFramebuffer(v.__webglFramebuffer[Q]);v.__webglDepthbuffer&&i.deleteRenderbuffer(v.__webglDepthbuffer[Q])}else{if(Array.isArray(v.__webglFramebuffer))for(let Q=0;Q<v.__webglFramebuffer.length;Q++)i.deleteFramebuffer(v.__webglFramebuffer[Q]);else i.deleteFramebuffer(v.__webglFramebuffer);if(v.__webglDepthbuffer&&i.deleteRenderbuffer(v.__webglDepthbuffer),v.__webglMultisampledFramebuffer&&i.deleteFramebuffer(v.__webglMultisampledFramebuffer),v.__webglColorRenderbuffer)for(let Q=0;Q<v.__webglColorRenderbuffer.length;Q++)v.__webglColorRenderbuffer[Q]&&i.deleteRenderbuffer(v.__webglColorRenderbuffer[Q]);v.__webglDepthRenderbuffer&&i.deleteRenderbuffer(v.__webglDepthRenderbuffer)}const V=E.textures;for(let Q=0,re=V.length;Q<re;Q++){const ee=r.get(V[Q]);ee.__webglTexture&&(i.deleteTexture(ee.__webglTexture),o.memory.textures--),r.remove(V[Q])}r.remove(E)}let R=0;function H(){R=0}function U(){const E=R;return E>=n.maxTextures&&console.warn("THREE.WebGLTextures: Trying to use "+E+" texture units while this GPU supports only "+n.maxTextures),R+=1,E}function F(E){const v=[];return v.push(E.wrapS),v.push(E.wrapT),v.push(E.wrapR||0),v.push(E.magFilter),v.push(E.minFilter),v.push(E.anisotropy),v.push(E.internalFormat),v.push(E.format),v.push(E.type),v.push(E.generateMipmaps),v.push(E.premultiplyAlpha),v.push(E.flipY),v.push(E.unpackAlignment),v.push(E.colorSpace),v.join()}function K(E,v){const V=r.get(E);if(E.isVideoTexture&&ie(E),E.isRenderTargetTexture===!1&&E.version>0&&V.__version!==E.version){const Q=E.image;if(Q===null)console.warn("THREE.WebGLRenderer: Texture marked for update but no image data found.");else if(Q.complete===!1)console.warn("THREE.WebGLRenderer: Texture marked for update but image is incomplete");else{G(V,E,v);return}}t.bindTexture(i.TEXTURE_2D,V.__webglTexture,i.TEXTURE0+v)}function X(E,v){const V=r.get(E);if(E.version>0&&V.__version!==E.version){G(V,E,v);return}t.bindTexture(i.TEXTURE_2D_ARRAY,V.__webglTexture,i.TEXTURE0+v)}function $(E,v){const V=r.get(E);if(E.version>0&&V.__version!==E.version){G(V,E,v);return}t.bindTexture(i.TEXTURE_3D,V.__webglTexture,i.TEXTURE0+v)}function k(E,v){const V=r.get(E);if(E.version>0&&V.__version!==E.version){J(V,E,v);return}t.bindTexture(i.TEXTURE_CUBE_MAP,V.__webglTexture,i.TEXTURE0+v)}const Z={[Pa]:i.REPEAT,[zr]:i.CLAMP_TO_EDGE,[Da]:i.MIRRORED_REPEAT},ce={[pt]:i.NEAREST,[fc]:i.NEAREST_MIPMAP_NEAREST,[Ki]:i.NEAREST_MIPMAP_LINEAR,[jt]:i.LINEAR,[kn]:i.LINEAR_MIPMAP_NEAREST,[Br]:i.LINEAR_MIPMAP_LINEAR},ve={[vc]:i.NEVER,[bc]:i.ALWAYS,[xc]:i.LESS,[sl]:i.LEQUAL,[yc]:i.EQUAL,[Ec]:i.GEQUAL,[Mc]:i.GREATER,[Sc]:i.NOTEQUAL};function De(E,v){if(v.type===Qt&&e.has("OES_texture_float_linear")===!1&&(v.magFilter===jt||v.magFilter===kn||v.magFilter===Ki||v.magFilter===Br||v.minFilter===jt||v.minFilter===kn||v.minFilter===Ki||v.minFilter===Br)&&console.warn("THREE.WebGLRenderer: Unable to use linear filtering with floating point textures. OES_texture_float_linear not supported on this device."),i.texParameteri(E,i.TEXTURE_WRAP_S,Z[v.wrapS]),i.texParameteri(E,i.TEXTURE_WRAP_T,Z[v.wrapT]),(E===i.TEXTURE_3D||E===i.TEXTURE_2D_ARRAY)&&i.texParameteri(E,i.TEXTURE_WRAP_R,Z[v.wrapR]),i.texParameteri(E,i.TEXTURE_MAG_FILTER,ce[v.magFilter]),i.texParameteri(E,i.TEXTURE_MIN_FILTER,ce[v.minFilter]),v.compareFunction&&(i.texParameteri(E,i.TEXTURE_COMPARE_MODE,i.COMPARE_REF_TO_TEXTURE),i.texParameteri(E,i.TEXTURE_COMPARE_FUNC,ve[v.compareFunction])),e.has("EXT_texture_filter_anisotropic")===!0){if(v.magFilter===pt||v.minFilter!==Ki&&v.minFilter!==Br||v.type===Qt&&e.has("OES_texture_float_linear")===!1)return;if(v.anisotropy>1||r.get(v).__currentAnisotropy){const V=e.get("EXT_texture_filter_anisotropic");i.texParameterf(E,V.TEXTURE_MAX_ANISOTROPY_EXT,Math.min(v.anisotropy,n.getMaxAnisotropy())),r.get(v).__currentAnisotropy=v.anisotropy}}}function xe(E,v){let V=!1;E.__webglInit===void 0&&(E.__webglInit=!0,v.addEventListener("dispose",w));const Q=v.source;let re=p.get(Q);re===void 0&&(re={},p.set(Q,re));const ee=F(v);if(ee!==E.__cacheKey){re[ee]===void 0&&(re[ee]={texture:i.createTexture(),usedTimes:0},o.memory.textures++,V=!0),re[ee].usedTimes++;const Re=re[E.__cacheKey];Re!==void 0&&(re[E.__cacheKey].usedTimes--,Re.usedTimes===0&&b(v)),E.__cacheKey=ee,E.__webglTexture=re[ee].texture}return V}function G(E,v,V){let Q=i.TEXTURE_2D;(v.isDataArrayTexture||v.isCompressedArrayTexture)&&(Q=i.TEXTURE_2D_ARRAY),v.isData3DTexture&&(Q=i.TEXTURE_3D);const re=xe(E,v),ee=v.source;t.bindTexture(Q,E.__webglTexture,i.TEXTURE0+V);const Re=r.get(ee);if(ee.version!==Re.__version||re===!0){t.activeTexture(i.TEXTURE0+V);const _e=$e.getPrimaries($e.workingColorSpace),be=v.colorSpace===_r?null:$e.getPrimaries(v.colorSpace),qe=v.colorSpace===_r||_e===be?i.NONE:i.BROWSER_DEFAULT_WEBGL;i.pixelStorei(i.UNPACK_FLIP_Y_WEBGL,v.flipY),i.pixelStorei(i.UNPACK_PREMULTIPLY_ALPHA_WEBGL,v.premultiplyAlpha),i.pixelStorei(i.UNPACK_ALIGNMENT,v.unpackAlignment),i.pixelStorei(i.UNPACK_COLORSPACE_CONVERSION_WEBGL,qe);let he=_(v.image,!1,n.maxTextureSize);he=le(v,he);const Se=a.convert(v.format,v.colorSpace),Ue=a.convert(v.type);let Oe=T(v.internalFormat,Se,Ue,v.colorSpace,v.isVideoTexture);De(Q,v);let Ee;const Xe=v.mipmaps,He=v.isVideoTexture!==!0,nt=Re.__version===void 0||re===!0,D=ee.dataReady,me=N(v,he);if(v.isDepthTexture)Oe=x(v.format===vi,v.type),nt&&(He?t.texStorage2D(i.TEXTURE_2D,1,Oe,he.width,he.height):t.texImage2D(i.TEXTURE_2D,0,Oe,he.width,he.height,0,Se,Ue,null));else if(v.isDataTexture)if(Xe.length>0){He&&nt&&t.texStorage2D(i.TEXTURE_2D,me,Oe,Xe[0].width,Xe[0].height);for(let Y=0,te=Xe.length;Y<te;Y++)Ee=Xe[Y],He?D&&t.texSubImage2D(i.TEXTURE_2D,Y,0,0,Ee.width,Ee.height,Se,Ue,Ee.data):t.texImage2D(i.TEXTURE_2D,Y,Oe,Ee.width,Ee.height,0,Se,Ue,Ee.data);v.generateMipmaps=!1}else He?(nt&&t.texStorage2D(i.TEXTURE_2D,me,Oe,he.width,he.height),D&&t.texSubImage2D(i.TEXTURE_2D,0,0,0,he.width,he.height,Se,Ue,he.data)):t.texImage2D(i.TEXTURE_2D,0,Oe,he.width,he.height,0,Se,Ue,he.data);else if(v.isCompressedTexture)if(v.isCompressedArrayTexture){He&&nt&&t.texStorage3D(i.TEXTURE_2D_ARRAY,me,Oe,Xe[0].width,Xe[0].height,he.depth);for(let Y=0,te=Xe.length;Y<te;Y++)if(Ee=Xe[Y],v.format!==Xt)if(Se!==null)if(He){if(D)if(v.layerUpdates.size>0){const ye=Do(Ee.width,Ee.height,v.format,v.type);for(const de of v.layerUpdates){const Ye=Ee.data.subarray(de*ye/Ee.data.BYTES_PER_ELEMENT,(de+1)*ye/Ee.data.BYTES_PER_ELEMENT);t.compressedTexSubImage3D(i.TEXTURE_2D_ARRAY,Y,0,0,de,Ee.width,Ee.height,1,Se,Ye)}v.clearLayerUpdates()}else t.compressedTexSubImage3D(i.TEXTURE_2D_ARRAY,Y,0,0,0,Ee.width,Ee.height,he.depth,Se,Ee.data)}else t.compressedTexImage3D(i.TEXTURE_2D_ARRAY,Y,Oe,Ee.width,Ee.height,he.depth,0,Ee.data,0,0);else console.warn("THREE.WebGLRenderer: Attempt to load unsupported compressed texture format in .uploadTexture()");else He?D&&t.texSubImage3D(i.TEXTURE_2D_ARRAY,Y,0,0,0,Ee.width,Ee.height,he.depth,Se,Ue,Ee.data):t.texImage3D(i.TEXTURE_2D_ARRAY,Y,Oe,Ee.width,Ee.height,he.depth,0,Se,Ue,Ee.data)}else{He&&nt&&t.texStorage2D(i.TEXTURE_2D,me,Oe,Xe[0].width,Xe[0].height);for(let Y=0,te=Xe.length;Y<te;Y++)Ee=Xe[Y],v.format!==Xt?Se!==null?He?D&&t.compressedTexSubImage2D(i.TEXTURE_2D,Y,0,0,Ee.width,Ee.height,Se,Ee.data):t.compressedTexImage2D(i.TEXTURE_2D,Y,Oe,Ee.width,Ee.height,0,Ee.data):console.warn("THREE.WebGLRenderer: Attempt to load unsupported compressed texture format in .uploadTexture()"):He?D&&t.texSubImage2D(i.TEXTURE_2D,Y,0,0,Ee.width,Ee.height,Se,Ue,Ee.data):t.texImage2D(i.TEXTURE_2D,Y,Oe,Ee.width,Ee.height,0,Se,Ue,Ee.data)}else if(v.isDataArrayTexture)if(He){if(nt&&t.texStorage3D(i.TEXTURE_2D_ARRAY,me,Oe,he.width,he.height,he.depth),D)if(v.layerUpdates.size>0){const Y=Do(he.width,he.height,v.format,v.type);for(const te of v.layerUpdates){const ye=he.data.subarray(te*Y/he.data.BYTES_PER_ELEMENT,(te+1)*Y/he.data.BYTES_PER_ELEMENT);t.texSubImage3D(i.TEXTURE_2D_ARRAY,0,0,0,te,he.width,he.height,1,Se,Ue,ye)}v.clearLayerUpdates()}else t.texSubImage3D(i.TEXTURE_2D_ARRAY,0,0,0,0,he.width,he.height,he.depth,Se,Ue,he.data)}else t.texImage3D(i.TEXTURE_2D_ARRAY,0,Oe,he.width,he.height,he.depth,0,Se,Ue,he.data);else if(v.isData3DTexture)He?(nt&&t.texStorage3D(i.TEXTURE_3D,me,Oe,he.width,he.height,he.depth),D&&t.texSubImage3D(i.TEXTURE_3D,0,0,0,0,he.width,he.height,he.depth,Se,Ue,he.data)):t.texImage3D(i.TEXTURE_3D,0,Oe,he.width,he.height,he.depth,0,Se,Ue,he.data);else if(v.isFramebufferTexture){if(nt)if(He)t.texStorage2D(i.TEXTURE_2D,me,Oe,he.width,he.height);else{let Y=he.width,te=he.height;for(let ye=0;ye<me;ye++)t.texImage2D(i.TEXTURE_2D,ye,Oe,Y,te,0,Se,Ue,null),Y>>=1,te>>=1}}else if(Xe.length>0){if(He&&nt){const Y=ge(Xe[0]);t.texStorage2D(i.TEXTURE_2D,me,Oe,Y.width,Y.height)}for(let Y=0,te=Xe.length;Y<te;Y++)Ee=Xe[Y],He?D&&t.texSubImage2D(i.TEXTURE_2D,Y,0,0,Se,Ue,Ee):t.texImage2D(i.TEXTURE_2D,Y,Oe,Se,Ue,Ee);v.generateMipmaps=!1}else if(He){if(nt){const Y=ge(he);t.texStorage2D(i.TEXTURE_2D,me,Oe,Y.width,Y.height)}D&&t.texSubImage2D(i.TEXTURE_2D,0,0,0,Se,Ue,he)}else t.texImage2D(i.TEXTURE_2D,0,Oe,Se,Ue,he);m(v)&&d(Q),Re.__version=ee.version,v.onUpdate&&v.onUpdate(v)}E.__version=v.version}function J(E,v,V){if(v.image.length!==6)return;const Q=xe(E,v),re=v.source;t.bindTexture(i.TEXTURE_CUBE_MAP,E.__webglTexture,i.TEXTURE0+V);const ee=r.get(re);if(re.version!==ee.__version||Q===!0){t.activeTexture(i.TEXTURE0+V);const Re=$e.getPrimaries($e.workingColorSpace),_e=v.colorSpace===_r?null:$e.getPrimaries(v.colorSpace),be=v.colorSpace===_r||Re===_e?i.NONE:i.BROWSER_DEFAULT_WEBGL;i.pixelStorei(i.UNPACK_FLIP_Y_WEBGL,v.flipY),i.pixelStorei(i.UNPACK_PREMULTIPLY_ALPHA_WEBGL,v.premultiplyAlpha),i.pixelStorei(i.UNPACK_ALIGNMENT,v.unpackAlignment),i.pixelStorei(i.UNPACK_COLORSPACE_CONVERSION_WEBGL,be);const qe=v.isCompressedTexture||v.image[0].isCompressedTexture,he=v.image[0]&&v.image[0].isDataTexture,Se=[];for(let te=0;te<6;te++)!qe&&!he?Se[te]=_(v.image[te],!0,n.maxCubemapSize):Se[te]=he?v.image[te].image:v.image[te],Se[te]=le(v,Se[te]);const Ue=Se[0],Oe=a.convert(v.format,v.colorSpace),Ee=a.convert(v.type),Xe=T(v.internalFormat,Oe,Ee,v.colorSpace),He=v.isVideoTexture!==!0,nt=ee.__version===void 0||Q===!0,D=re.dataReady;let me=N(v,Ue);De(i.TEXTURE_CUBE_MAP,v);let Y;if(qe){He&&nt&&t.texStorage2D(i.TEXTURE_CUBE_MAP,me,Xe,Ue.width,Ue.height);for(let te=0;te<6;te++){Y=Se[te].mipmaps;for(let ye=0;ye<Y.length;ye++){const de=Y[ye];v.format!==Xt?Oe!==null?He?D&&t.compressedTexSubImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye,0,0,de.width,de.height,Oe,de.data):t.compressedTexImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye,Xe,de.width,de.height,0,de.data):console.warn("THREE.WebGLRenderer: Attempt to load unsupported compressed texture format in .setTextureCube()"):He?D&&t.texSubImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye,0,0,de.width,de.height,Oe,Ee,de.data):t.texImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye,Xe,de.width,de.height,0,Oe,Ee,de.data)}}}else{if(Y=v.mipmaps,He&&nt){Y.length>0&&me++;const te=ge(Se[0]);t.texStorage2D(i.TEXTURE_CUBE_MAP,me,Xe,te.width,te.height)}for(let te=0;te<6;te++)if(he){He?D&&t.texSubImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,0,0,0,Se[te].width,Se[te].height,Oe,Ee,Se[te].data):t.texImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,0,Xe,Se[te].width,Se[te].height,0,Oe,Ee,Se[te].data);for(let ye=0;ye<Y.length;ye++){const de=Y[ye].image[te].image;He?D&&t.texSubImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye+1,0,0,de.width,de.height,Oe,Ee,de.data):t.texImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye+1,Xe,de.width,de.height,0,Oe,Ee,de.data)}}else{He?D&&t.texSubImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,0,0,0,Oe,Ee,Se[te]):t.texImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,0,Xe,Oe,Ee,Se[te]);for(let ye=0;ye<Y.length;ye++){const de=Y[ye];He?D&&t.texSubImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye+1,0,0,Oe,Ee,de.image[te]):t.texImage2D(i.TEXTURE_CUBE_MAP_POSITIVE_X+te,ye+1,Xe,Oe,Ee,de.image[te])}}}m(v)&&d(i.TEXTURE_CUBE_MAP),ee.__version=re.version,v.onUpdate&&v.onUpdate(v)}E.__version=v.version}function q(E,v,V,Q,re,ee){const Re=a.convert(V.format,V.colorSpace),_e=a.convert(V.type),be=T(V.internalFormat,Re,_e,V.colorSpace),qe=r.get(v),he=r.get(V);if(he.__renderTarget=v,!qe.__hasExternalTextures){const Se=Math.max(1,v.width>>ee),Ue=Math.max(1,v.height>>ee);re===i.TEXTURE_3D||re===i.TEXTURE_2D_ARRAY?t.texImage3D(re,ee,be,Se,Ue,v.depth,0,Re,_e,null):t.texImage2D(re,ee,be,Se,Ue,0,Re,_e,null)}t.bindFramebuffer(i.FRAMEBUFFER,E),pe(v)?s.framebufferTexture2DMultisampleEXT(i.FRAMEBUFFER,Q,re,he.__webglTexture,0,ue(v)):(re===i.TEXTURE_2D||re>=i.TEXTURE_CUBE_MAP_POSITIVE_X&&re<=i.TEXTURE_CUBE_MAP_NEGATIVE_Z)&&i.framebufferTexture2D(i.FRAMEBUFFER,Q,re,he.__webglTexture,ee),t.bindFramebuffer(i.FRAMEBUFFER,null)}function z(E,v,V){if(i.bindRenderbuffer(i.RENDERBUFFER,E),v.depthBuffer){const Q=v.depthTexture,re=Q&&Q.isDepthTexture?Q.type:null,ee=x(v.stencilBuffer,re),Re=v.stencilBuffer?i.DEPTH_STENCIL_ATTACHMENT:i.DEPTH_ATTACHMENT,_e=ue(v);pe(v)?s.renderbufferStorageMultisampleEXT(i.RENDERBUFFER,_e,ee,v.width,v.height):V?i.renderbufferStorageMultisample(i.RENDERBUFFER,_e,ee,v.width,v.height):i.renderbufferStorage(i.RENDERBUFFER,ee,v.width,v.height),i.framebufferRenderbuffer(i.FRAMEBUFFER,Re,i.RENDERBUFFER,E)}else{const Q=v.textures;for(let re=0;re<Q.length;re++){const ee=Q[re],Re=a.convert(ee.format,ee.colorSpace),_e=a.convert(ee.type),be=T(ee.internalFormat,Re,_e,ee.colorSpace),qe=ue(v);V&&pe(v)===!1?i.renderbufferStorageMultisample(i.RENDERBUFFER,qe,be,v.width,v.height):pe(v)?s.renderbufferStorageMultisampleEXT(i.RENDERBUFFER,qe,be,v.width,v.height):i.renderbufferStorage(i.RENDERBUFFER,be,v.width,v.height)}}i.bindRenderbuffer(i.RENDERBUFFER,null)}function ae(E,v){if(v&&v.isWebGLCubeRenderTarget)throw new Error("Depth Texture with cube render targets is not supported");if(t.bindFramebuffer(i.FRAMEBUFFER,E),!(v.depthTexture&&v.depthTexture.isDepthTexture))throw new Error("renderTarget.depthTexture must be an instance of THREE.DepthTexture");const V=r.get(v.depthTexture);V.__renderTarget=v,(!V.__webglTexture||v.depthTexture.image.width!==v.width||v.depthTexture.image.height!==v.height)&&(v.depthTexture.image.width=v.width,v.depthTexture.image.height=v.height,v.depthTexture.needsUpdate=!0),K(v.depthTexture,0);const Q=V.__webglTexture,re=ue(v);if(v.depthTexture.format===di)pe(v)?s.framebufferTexture2DMultisampleEXT(i.FRAMEBUFFER,i.DEPTH_ATTACHMENT,i.TEXTURE_2D,Q,0,re):i.framebufferTexture2D(i.FRAMEBUFFER,i.DEPTH_ATTACHMENT,i.TEXTURE_2D,Q,0);else if(v.depthTexture.format===vi)pe(v)?s.framebufferTexture2DMultisampleEXT(i.FRAMEBUFFER,i.DEPTH_STENCIL_ATTACHMENT,i.TEXTURE_2D,Q,0,re):i.framebufferTexture2D(i.FRAMEBUFFER,i.DEPTH_STENCIL_ATTACHMENT,i.TEXTURE_2D,Q,0);else throw new Error("Unknown depthTexture format")}function ne(E){const v=r.get(E),V=E.isWebGLCubeRenderTarget===!0;if(v.__boundDepthTexture!==E.depthTexture){const Q=E.depthTexture;if(v.__depthDisposeCallback&&v.__depthDisposeCallback(),Q){const re=()=>{delete v.__boundDepthTexture,delete v.__depthDisposeCallback,Q.removeEventListener("dispose",re)};Q.addEventListener("dispose",re),v.__depthDisposeCallback=re}v.__boundDepthTexture=Q}if(E.depthTexture&&!v.__autoAllocateDepthBuffer){if(V)throw new Error("target.depthTexture not supported in Cube render targets");ae(v.__webglFramebuffer,E)}else if(V){v.__webglDepthbuffer=[];for(let Q=0;Q<6;Q++)if(t.bindFramebuffer(i.FRAMEBUFFER,v.__webglFramebuffer[Q]),v.__webglDepthbuffer[Q]===void 0)v.__webglDepthbuffer[Q]=i.createRenderbuffer(),z(v.__webglDepthbuffer[Q],E,!1);else{const re=E.stencilBuffer?i.DEPTH_STENCIL_ATTACHMENT:i.DEPTH_ATTACHMENT,ee=v.__webglDepthbuffer[Q];i.bindRenderbuffer(i.RENDERBUFFER,ee),i.framebufferRenderbuffer(i.FRAMEBUFFER,re,i.RENDERBUFFER,ee)}}else if(t.bindFramebuffer(i.FRAMEBUFFER,v.__webglFramebuffer),v.__webglDepthbuffer===void 0)v.__webglDepthbuffer=i.createRenderbuffer(),z(v.__webglDepthbuffer,E,!1);else{const Q=E.stencilBuffer?i.DEPTH_STENCIL_ATTACHMENT:i.DEPTH_ATTACHMENT,re=v.__webglDepthbuffer;i.bindRenderbuffer(i.RENDERBUFFER,re),i.framebufferRenderbuffer(i.FRAMEBUFFER,Q,i.RENDERBUFFER,re)}t.bindFramebuffer(i.FRAMEBUFFER,null)}function oe(E,v,V){const Q=r.get(E);v!==void 0&&q(Q.__webglFramebuffer,E,E.texture,i.COLOR_ATTACHMENT0,i.TEXTURE_2D,0),V!==void 0&&ne(E)}function Ae(E){const v=E.texture,V=r.get(E),Q=r.get(v);E.addEventListener("dispose",A);const re=E.textures,ee=E.isWebGLCubeRenderTarget===!0,Re=re.length>1;if(Re||(Q.__webglTexture===void 0&&(Q.__webglTexture=i.createTexture()),Q.__version=v.version,o.memory.textures++),ee){V.__webglFramebuffer=[];for(let _e=0;_e<6;_e++)if(v.mipmaps&&v.mipmaps.length>0){V.__webglFramebuffer[_e]=[];for(let be=0;be<v.mipmaps.length;be++)V.__webglFramebuffer[_e][be]=i.createFramebuffer()}else V.__webglFramebuffer[_e]=i.createFramebuffer()}else{if(v.mipmaps&&v.mipmaps.length>0){V.__webglFramebuffer=[];for(let _e=0;_e<v.mipmaps.length;_e++)V.__webglFramebuffer[_e]=i.createFramebuffer()}else V.__webglFramebuffer=i.createFramebuffer();if(Re)for(let _e=0,be=re.length;_e<be;_e++){const qe=r.get(re[_e]);qe.__webglTexture===void 0&&(qe.__webglTexture=i.createTexture(),o.memory.textures++)}if(E.samples>0&&pe(E)===!1){V.__webglMultisampledFramebuffer=i.createFramebuffer(),V.__webglColorRenderbuffer=[],t.bindFramebuffer(i.FRAMEBUFFER,V.__webglMultisampledFramebuffer);for(let _e=0;_e<re.length;_e++){const be=re[_e];V.__webglColorRenderbuffer[_e]=i.createRenderbuffer(),i.bindRenderbuffer(i.RENDERBUFFER,V.__webglColorRenderbuffer[_e]);const qe=a.convert(be.format,be.colorSpace),he=a.convert(be.type),Se=T(be.internalFormat,qe,he,be.colorSpace,E.isXRRenderTarget===!0),Ue=ue(E);i.renderbufferStorageMultisample(i.RENDERBUFFER,Ue,Se,E.width,E.height),i.framebufferRenderbuffer(i.FRAMEBUFFER,i.COLOR_ATTACHMENT0+_e,i.RENDERBUFFER,V.__webglColorRenderbuffer[_e])}i.bindRenderbuffer(i.RENDERBUFFER,null),E.depthBuffer&&(V.__webglDepthRenderbuffer=i.createRenderbuffer(),z(V.__webglDepthRenderbuffer,E,!0)),t.bindFramebuffer(i.FRAMEBUFFER,null)}}if(ee){t.bindTexture(i.TEXTURE_CUBE_MAP,Q.__webglTexture),De(i.TEXTURE_CUBE_MAP,v);for(let _e=0;_e<6;_e++)if(v.mipmaps&&v.mipmaps.length>0)for(let be=0;be<v.mipmaps.length;be++)q(V.__webglFramebuffer[_e][be],E,v,i.COLOR_ATTACHMENT0,i.TEXTURE_CUBE_MAP_POSITIVE_X+_e,be);else q(V.__webglFramebuffer[_e],E,v,i.COLOR_ATTACHMENT0,i.TEXTURE_CUBE_MAP_POSITIVE_X+_e,0);m(v)&&d(i.TEXTURE_CUBE_MAP),t.unbindTexture()}else if(Re){for(let _e=0,be=re.length;_e<be;_e++){const qe=re[_e],he=r.get(qe);t.bindTexture(i.TEXTURE_2D,he.__webglTexture),De(i.TEXTURE_2D,qe),q(V.__webglFramebuffer,E,qe,i.COLOR_ATTACHMENT0+_e,i.TEXTURE_2D,0),m(qe)&&d(i.TEXTURE_2D)}t.unbindTexture()}else{let _e=i.TEXTURE_2D;if((E.isWebGL3DRenderTarget||E.isWebGLArrayRenderTarget)&&(_e=E.isWebGL3DRenderTarget?i.TEXTURE_3D:i.TEXTURE_2D_ARRAY),t.bindTexture(_e,Q.__webglTexture),De(_e,v),v.mipmaps&&v.mipmaps.length>0)for(let be=0;be<v.mipmaps.length;be++)q(V.__webglFramebuffer[be],E,v,i.COLOR_ATTACHMENT0,_e,be);else q(V.__webglFramebuffer,E,v,i.COLOR_ATTACHMENT0,_e,0);m(v)&&d(_e),t.unbindTexture()}E.depthBuffer&&ne(E)}function Te(E){const v=E.textures;for(let V=0,Q=v.length;V<Q;V++){const re=v[V];if(m(re)){const ee=S(E),Re=r.get(re).__webglTexture;t.bindTexture(ee,Re),d(ee),t.unbindTexture()}}}const Be=[],C=[];function Je(E){if(E.samples>0){if(pe(E)===!1){const v=E.textures,V=E.width,Q=E.height;let re=i.COLOR_BUFFER_BIT;const ee=E.stencilBuffer?i.DEPTH_STENCIL_ATTACHMENT:i.DEPTH_ATTACHMENT,Re=r.get(E),_e=v.length>1;if(_e)for(let be=0;be<v.length;be++)t.bindFramebuffer(i.FRAMEBUFFER,Re.__webglMultisampledFramebuffer),i.framebufferRenderbuffer(i.FRAMEBUFFER,i.COLOR_ATTACHMENT0+be,i.RENDERBUFFER,null),t.bindFramebuffer(i.FRAMEBUFFER,Re.__webglFramebuffer),i.framebufferTexture2D(i.DRAW_FRAMEBUFFER,i.COLOR_ATTACHMENT0+be,i.TEXTURE_2D,null,0);t.bindFramebuffer(i.READ_FRAMEBUFFER,Re.__webglMultisampledFramebuffer),t.bindFramebuffer(i.DRAW_FRAMEBUFFER,Re.__webglFramebuffer);for(let be=0;be<v.length;be++){if(E.resolveDepthBuffer&&(E.depthBuffer&&(re|=i.DEPTH_BUFFER_BIT),E.stencilBuffer&&E.resolveStencilBuffer&&(re|=i.STENCIL_BUFFER_BIT)),_e){i.framebufferRenderbuffer(i.READ_FRAMEBUFFER,i.COLOR_ATTACHMENT0,i.RENDERBUFFER,Re.__webglColorRenderbuffer[be]);const qe=r.get(v[be]).__webglTexture;i.framebufferTexture2D(i.DRAW_FRAMEBUFFER,i.COLOR_ATTACHMENT0,i.TEXTURE_2D,qe,0)}i.blitFramebuffer(0,0,V,Q,0,0,V,Q,re,i.NEAREST),l===!0&&(Be.length=0,C.length=0,Be.push(i.COLOR_ATTACHMENT0+be),E.depthBuffer&&E.resolveDepthBuffer===!1&&(Be.push(ee),C.push(ee),i.invalidateFramebuffer(i.DRAW_FRAMEBUFFER,C)),i.invalidateFramebuffer(i.READ_FRAMEBUFFER,Be))}if(t.bindFramebuffer(i.READ_FRAMEBUFFER,null),t.bindFramebuffer(i.DRAW_FRAMEBUFFER,null),_e)for(let be=0;be<v.length;be++){t.bindFramebuffer(i.FRAMEBUFFER,Re.__webglMultisampledFramebuffer),i.framebufferRenderbuffer(i.FRAMEBUFFER,i.COLOR_ATTACHMENT0+be,i.RENDERBUFFER,Re.__webglColorRenderbuffer[be]);const qe=r.get(v[be]).__webglTexture;t.bindFramebuffer(i.FRAMEBUFFER,Re.__webglFramebuffer),i.framebufferTexture2D(i.DRAW_FRAMEBUFFER,i.COLOR_ATTACHMENT0+be,i.TEXTURE_2D,qe,0)}t.bindFramebuffer(i.DRAW_FRAMEBUFFER,Re.__webglMultisampledFramebuffer)}else if(E.depthBuffer&&E.resolveDepthBuffer===!1&&l){const v=E.stencilBuffer?i.DEPTH_STENCIL_ATTACHMENT:i.DEPTH_ATTACHMENT;i.invalidateFramebuffer(i.DRAW_FRAMEBUFFER,[v])}}}function ue(E){return Math.min(n.maxSamples,E.samples)}function pe(E){const v=r.get(E);return E.samples>0&&e.has("WEBGL_multisampled_render_to_texture")===!0&&v.__useRenderToTexture!==!1}function ie(E){const v=o.render.frame;h.get(E)!==v&&(h.set(E,v),E.update())}function le(E,v){const V=E.colorSpace,Q=E.format,re=E.type;return E.isCompressedTexture===!0||E.isVideoTexture===!0||V!==yi&&V!==_r&&($e.getTransfer(V)===rt?(Q!==Xt||re!==cr)&&console.warn("THREE.WebGLTextures: sRGB encoded textures have to use RGBAFormat and UnsignedByteType."):console.error("THREE.WebGLTextures: Unsupported texture color space:",V)),v}function ge(E){return typeof HTMLImageElement<"u"&&E instanceof HTMLImageElement?(c.width=E.naturalWidth||E.width,c.height=E.naturalHeight||E.height):typeof VideoFrame<"u"&&E instanceof VideoFrame?(c.width=E.displayWidth,c.height=E.displayHeight):(c.width=E.width,c.height=E.height),c}this.allocateTextureUnit=U,this.resetTextureUnits=H,this.setTexture2D=K,this.setTexture2DArray=X,this.setTexture3D=$,this.setTextureCube=k,this.rebindTextures=oe,this.setupRenderTarget=Ae,this.updateRenderTargetMipmap=Te,this.updateMultisampleRenderTarget=Je,this.setupDepthRenderbuffer=ne,this.setupFrameBufferTexture=q,this.useMultisampledRTT=pe}function Cf(i,e){function t(r,n=_r){let a;const o=$e.getTransfer(n);if(r===cr)return i.UNSIGNED_BYTE;if(r===fs)return i.UNSIGNED_SHORT_4_4_4_4;if(r===ms)return i.UNSIGNED_SHORT_5_5_5_1;if(r===Jo)return i.UNSIGNED_INT_5_9_9_9_REV;if(r===$o)return i.BYTE;if(r===Zo)return i.SHORT;if(r===ki)return i.UNSIGNED_SHORT;if(r===ps)return i.INT;if(r===kr)return i.UNSIGNED_INT;if(r===Qt)return i.FLOAT;if(r===Hi)return i.HALF_FLOAT;if(r===Qo)return i.ALPHA;if(r===el)return i.RGB;if(r===Xt)return i.RGBA;if(r===tl)return i.LUMINANCE;if(r===rl)return i.LUMINANCE_ALPHA;if(r===di)return i.DEPTH_COMPONENT;if(r===vi)return i.DEPTH_STENCIL;if(r===gs)return i.RED;if(r===_s)return i.RED_INTEGER;if(r===il)return i.RG;if(r===vs)return i.RG_INTEGER;if(r===xs)return i.RGBA_INTEGER;if(r===Tn||r===wn||r===An||r===Rn)if(o===rt)if(a=e.get("WEBGL_compressed_texture_s3tc_srgb"),a!==null){if(r===Tn)return a.COMPRESSED_SRGB_S3TC_DXT1_EXT;if(r===wn)return a.COMPRESSED_SRGB_ALPHA_S3TC_DXT1_EXT;if(r===An)return a.COMPRESSED_SRGB_ALPHA_S3TC_DXT3_EXT;if(r===Rn)return a.COMPRESSED_SRGB_ALPHA_S3TC_DXT5_EXT}else return null;else if(a=e.get("WEBGL_compressed_texture_s3tc"),a!==null){if(r===Tn)return a.COMPRESSED_RGB_S3TC_DXT1_EXT;if(r===wn)return a.COMPRESSED_RGBA_S3TC_DXT1_EXT;if(r===An)return a.COMPRESSED_RGBA_S3TC_DXT3_EXT;if(r===Rn)return a.COMPRESSED_RGBA_S3TC_DXT5_EXT}else return null;if(r===Ia||r===Na||r===Ua||r===Oa)if(a=e.get("WEBGL_compressed_texture_pvrtc"),a!==null){if(r===Ia)return a.COMPRESSED_RGB_PVRTC_4BPPV1_IMG;if(r===Na)return a.COMPRESSED_RGB_PVRTC_2BPPV1_IMG;if(r===Ua)return a.COMPRESSED_RGBA_PVRTC_4BPPV1_IMG;if(r===Oa)return a.COMPRESSED_RGBA_PVRTC_2BPPV1_IMG}else return null;if(r===Fa||r===za||r===Ba)if(a=e.get("WEBGL_compressed_texture_etc"),a!==null){if(r===Fa||r===za)return o===rt?a.COMPRESSED_SRGB8_ETC2:a.COMPRESSED_RGB8_ETC2;if(r===Ba)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ETC2_EAC:a.COMPRESSED_RGBA8_ETC2_EAC}else return null;if(r===ka||r===Va||r===Ha||r===Ga||r===Wa||r===ja||r===Xa||r===Ya||r===qa||r===Ka||r===$a||r===Za||r===Ja||r===Qa)if(a=e.get("WEBGL_compressed_texture_astc"),a!==null){if(r===ka)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_4x4_KHR:a.COMPRESSED_RGBA_ASTC_4x4_KHR;if(r===Va)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_5x4_KHR:a.COMPRESSED_RGBA_ASTC_5x4_KHR;if(r===Ha)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_5x5_KHR:a.COMPRESSED_RGBA_ASTC_5x5_KHR;if(r===Ga)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_6x5_KHR:a.COMPRESSED_RGBA_ASTC_6x5_KHR;if(r===Wa)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_6x6_KHR:a.COMPRESSED_RGBA_ASTC_6x6_KHR;if(r===ja)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_8x5_KHR:a.COMPRESSED_RGBA_ASTC_8x5_KHR;if(r===Xa)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_8x6_KHR:a.COMPRESSED_RGBA_ASTC_8x6_KHR;if(r===Ya)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_8x8_KHR:a.COMPRESSED_RGBA_ASTC_8x8_KHR;if(r===qa)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_10x5_KHR:a.COMPRESSED_RGBA_ASTC_10x5_KHR;if(r===Ka)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_10x6_KHR:a.COMPRESSED_RGBA_ASTC_10x6_KHR;if(r===$a)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_10x8_KHR:a.COMPRESSED_RGBA_ASTC_10x8_KHR;if(r===Za)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_10x10_KHR:a.COMPRESSED_RGBA_ASTC_10x10_KHR;if(r===Ja)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_12x10_KHR:a.COMPRESSED_RGBA_ASTC_12x10_KHR;if(r===Qa)return o===rt?a.COMPRESSED_SRGB8_ALPHA8_ASTC_12x12_KHR:a.COMPRESSED_RGBA_ASTC_12x12_KHR}else return null;if(r===Cn||r===es||r===ts)if(a=e.get("EXT_texture_compression_bptc"),a!==null){if(r===Cn)return o===rt?a.COMPRESSED_SRGB_ALPHA_BPTC_UNORM_EXT:a.COMPRESSED_RGBA_BPTC_UNORM_EXT;if(r===es)return a.COMPRESSED_RGB_BPTC_SIGNED_FLOAT_EXT;if(r===ts)return a.COMPRESSED_RGB_BPTC_UNSIGNED_FLOAT_EXT}else return null;if(r===nl||r===rs||r===is||r===ns)if(a=e.get("EXT_texture_compression_rgtc"),a!==null){if(r===Cn)return a.COMPRESSED_RED_RGTC1_EXT;if(r===rs)return a.COMPRESSED_SIGNED_RED_RGTC1_EXT;if(r===is)return a.COMPRESSED_RED_GREEN_RGTC2_EXT;if(r===ns)return a.COMPRESSED_SIGNED_RED_GREEN_RGTC2_EXT}else return null;return r===_i?i.UNSIGNED_INT_24_8:i[r]!==void 0?i[r]:null}return{convert:t}}class Lf extends Ft{constructor(e=[]){super(),this.isArrayCamera=!0,this.cameras=e}}class gt extends Mt{constructor(){super(),this.isGroup=!0,this.type="Group"}}const Pf={type:"move"};class ma{constructor(){this._targetRay=null,this._grip=null,this._hand=null}getHandSpace(){return this._hand===null&&(this._hand=new gt,this._hand.matrixAutoUpdate=!1,this._hand.visible=!1,this._hand.joints={},this._hand.inputState={pinching:!1}),this._hand}getTargetRaySpace(){return this._targetRay===null&&(this._targetRay=new gt,this._targetRay.matrixAutoUpdate=!1,this._targetRay.visible=!1,this._targetRay.hasLinearVelocity=!1,this._targetRay.linearVelocity=new B,this._targetRay.hasAngularVelocity=!1,this._targetRay.angularVelocity=new B),this._targetRay}getGripSpace(){return this._grip===null&&(this._grip=new gt,this._grip.matrixAutoUpdate=!1,this._grip.visible=!1,this._grip.hasLinearVelocity=!1,this._grip.linearVelocity=new B,this._grip.hasAngularVelocity=!1,this._grip.angularVelocity=new B),this._grip}dispatchEvent(e){return this._targetRay!==null&&this._targetRay.dispatchEvent(e),this._grip!==null&&this._grip.dispatchEvent(e),this._hand!==null&&this._hand.dispatchEvent(e),this}connect(e){if(e&&e.hand){const t=this._hand;if(t)for(const r of e.hand.values())this._getHandJoint(t,r)}return this.dispatchEvent({type:"connected",data:e}),this}disconnect(e){return this.dispatchEvent({type:"disconnected",data:e}),this._targetRay!==null&&(this._targetRay.visible=!1),this._grip!==null&&(this._grip.visible=!1),this._hand!==null&&(this._hand.visible=!1),this}update(e,t,r){let n=null,a=null,o=null;const s=this._targetRay,l=this._grip,c=this._hand;if(e&&t.session.visibilityState!=="visible-blurred"){if(c&&e.hand){o=!0;for(const _ of e.hand.values()){const m=t.getJointPose(_,r),d=this._getHandJoint(c,_);m!==null&&(d.matrix.fromArray(m.transform.matrix),d.matrix.decompose(d.position,d.rotation,d.scale),d.matrixWorldNeedsUpdate=!0,d.jointRadius=m.radius),d.visible=m!==null}const h=c.joints["index-finger-tip"],u=c.joints["thumb-tip"],p=h.position.distanceTo(u.position),f=.02,g=.005;c.inputState.pinching&&p>f+g?(c.inputState.pinching=!1,this.dispatchEvent({type:"pinchend",handedness:e.handedness,target:this})):!c.inputState.pinching&&p<=f-g&&(c.inputState.pinching=!0,this.dispatchEvent({type:"pinchstart",handedness:e.handedness,target:this}))}else l!==null&&e.gripSpace&&(a=t.getPose(e.gripSpace,r),a!==null&&(l.matrix.fromArray(a.transform.matrix),l.matrix.decompose(l.position,l.rotation,l.scale),l.matrixWorldNeedsUpdate=!0,a.linearVelocity?(l.hasLinearVelocity=!0,l.linearVelocity.copy(a.linearVelocity)):l.hasLinearVelocity=!1,a.angularVelocity?(l.hasAngularVelocity=!0,l.angularVelocity.copy(a.angularVelocity)):l.hasAngularVelocity=!1));s!==null&&(n=t.getPose(e.targetRaySpace,r),n===null&&a!==null&&(n=a),n!==null&&(s.matrix.fromArray(n.transform.matrix),s.matrix.decompose(s.position,s.rotation,s.scale),s.matrixWorldNeedsUpdate=!0,n.linearVelocity?(s.hasLinearVelocity=!0,s.linearVelocity.copy(n.linearVelocity)):s.hasLinearVelocity=!1,n.angularVelocity?(s.hasAngularVelocity=!0,s.angularVelocity.copy(n.angularVelocity)):s.hasAngularVelocity=!1,this.dispatchEvent(Pf)))}return s!==null&&(s.visible=n!==null),l!==null&&(l.visible=a!==null),c!==null&&(c.visible=o!==null),this}_getHandJoint(e,t){if(e.joints[t.jointName]===void 0){const r=new gt;r.matrixAutoUpdate=!1,r.visible=!1,e.joints[t.jointName]=r,e.add(r)}return e.joints[t.jointName]}}const Df=`
void main() {

	gl_Position = vec4( position, 1.0 );

}`,If=`
uniform sampler2DArray depthColor;
uniform float depthWidth;
uniform float depthHeight;

void main() {

	vec2 coord = vec2( gl_FragCoord.x / depthWidth, gl_FragCoord.y / depthHeight );

	if ( coord.x >= 1.0 ) {

		gl_FragDepth = texture( depthColor, vec3( coord.x - 1.0, coord.y, 1 ) ).r;

	} else {

		gl_FragDepth = texture( depthColor, vec3( coord.x, coord.y, 0 ) ).r;

	}

}`;class Nf{constructor(){this.texture=null,this.mesh=null,this.depthNear=0,this.depthFar=0}init(e,t,r){if(this.texture===null){const n=new _t,a=e.properties.get(n);a.__webglTexture=t.texture,(t.depthNear!=r.depthNear||t.depthFar!=r.depthFar)&&(this.depthNear=t.depthNear,this.depthFar=t.depthFar),this.texture=n}}getMesh(e){if(this.texture!==null&&this.mesh===null){const t=e.cameras[0].viewport,r=new Er({vertexShader:Df,fragmentShader:If,uniforms:{depthColor:{value:this.texture},depthWidth:{value:t.z},depthHeight:{value:t.w}}});this.mesh=new Fe(new vr(20,20),r)}return this.mesh}reset(){this.texture=null,this.mesh=null}getDepthTexture(){return this.texture}}class Uf extends Mi{constructor(e,t){super();const r=this;let n=null,a=1,o=null,s="local-floor",l=1,c=null,h=null,u=null,p=null,f=null,g=null;const _=new Nf,m=t.getContextAttributes();let d=null,S=null;const T=[],x=[],N=new je;let w=null;const A=new Ft;A.viewport=new dt;const O=new Ft;O.viewport=new dt;const b=[A,O],M=new Lf;let R=null,H=null;this.cameraAutoUpdate=!0,this.enabled=!1,this.isPresenting=!1,this.getController=function(G){let J=T[G];return J===void 0&&(J=new ma,T[G]=J),J.getTargetRaySpace()},this.getControllerGrip=function(G){let J=T[G];return J===void 0&&(J=new ma,T[G]=J),J.getGripSpace()},this.getHand=function(G){let J=T[G];return J===void 0&&(J=new ma,T[G]=J),J.getHandSpace()};function U(G){const J=x.indexOf(G.inputSource);if(J===-1)return;const q=T[J];q!==void 0&&(q.update(G.inputSource,G.frame,c||o),q.dispatchEvent({type:G.type,data:G.inputSource}))}function F(){n.removeEventListener("select",U),n.removeEventListener("selectstart",U),n.removeEventListener("selectend",U),n.removeEventListener("squeeze",U),n.removeEventListener("squeezestart",U),n.removeEventListener("squeezeend",U),n.removeEventListener("end",F),n.removeEventListener("inputsourceschange",K);for(let G=0;G<T.length;G++){const J=x[G];J!==null&&(x[G]=null,T[G].disconnect(J))}R=null,H=null,_.reset(),e.setRenderTarget(d),f=null,p=null,u=null,n=null,S=null,xe.stop(),r.isPresenting=!1,e.setPixelRatio(w),e.setSize(N.width,N.height,!1),r.dispatchEvent({type:"sessionend"})}this.setFramebufferScaleFactor=function(G){a=G,r.isPresenting===!0&&console.warn("THREE.WebXRManager: Cannot change framebuffer scale while presenting.")},this.setReferenceSpaceType=function(G){s=G,r.isPresenting===!0&&console.warn("THREE.WebXRManager: Cannot change reference space type while presenting.")},this.getReferenceSpace=function(){return c||o},this.setReferenceSpace=function(G){c=G},this.getBaseLayer=function(){return p!==null?p:f},this.getBinding=function(){return u},this.getFrame=function(){return g},this.getSession=function(){return n},this.setSession=async function(G){if(n=G,n!==null){if(d=e.getRenderTarget(),n.addEventListener("select",U),n.addEventListener("selectstart",U),n.addEventListener("selectend",U),n.addEventListener("squeeze",U),n.addEventListener("squeezestart",U),n.addEventListener("squeezeend",U),n.addEventListener("end",F),n.addEventListener("inputsourceschange",K),m.xrCompatible!==!0&&await t.makeXRCompatible(),w=e.getPixelRatio(),e.getSize(N),n.renderState.layers===void 0){const J={antialias:m.antialias,alpha:!0,depth:m.depth,stencil:m.stencil,framebufferScaleFactor:a};f=new XRWebGLLayer(n,t,J),n.updateRenderState({baseLayer:f}),e.setPixelRatio(1),e.setSize(f.framebufferWidth,f.framebufferHeight,!1),S=new Vr(f.framebufferWidth,f.framebufferHeight,{format:Xt,type:cr,colorSpace:e.outputColorSpace,stencilBuffer:m.stencil})}else{let J=null,q=null,z=null;m.depth&&(z=m.stencil?t.DEPTH24_STENCIL8:t.DEPTH_COMPONENT24,J=m.stencil?vi:di,q=m.stencil?_i:kr);const ae={colorFormat:t.RGBA8,depthFormat:z,scaleFactor:a};u=new XRWebGLBinding(n,t),p=u.createProjectionLayer(ae),n.updateRenderState({layers:[p]}),e.setPixelRatio(1),e.setSize(p.textureWidth,p.textureHeight,!1),S=new Vr(p.textureWidth,p.textureHeight,{format:Xt,type:cr,depthTexture:new yl(p.textureWidth,p.textureHeight,q,void 0,void 0,void 0,void 0,void 0,void 0,J),stencilBuffer:m.stencil,colorSpace:e.outputColorSpace,samples:m.antialias?4:0,resolveDepthBuffer:p.ignoreDepthValues===!1})}S.isXRRenderTarget=!0,this.setFoveation(l),c=null,o=await n.requestReferenceSpace(s),xe.setContext(n),xe.start(),r.isPresenting=!0,r.dispatchEvent({type:"sessionstart"})}},this.getEnvironmentBlendMode=function(){if(n!==null)return n.environmentBlendMode},this.getDepthTexture=function(){return _.getDepthTexture()};function K(G){for(let J=0;J<G.removed.length;J++){const q=G.removed[J],z=x.indexOf(q);z>=0&&(x[z]=null,T[z].disconnect(q))}for(let J=0;J<G.added.length;J++){const q=G.added[J];let z=x.indexOf(q);if(z===-1){for(let ne=0;ne<T.length;ne++)if(ne>=x.length){x.push(q),z=ne;break}else if(x[ne]===null){x[ne]=q,z=ne;break}if(z===-1)break}const ae=T[z];ae&&ae.connect(q)}}const X=new B,$=new B;function k(G,J,q){X.setFromMatrixPosition(J.matrixWorld),$.setFromMatrixPosition(q.matrixWorld);const z=X.distanceTo($),ae=J.projectionMatrix.elements,ne=q.projectionMatrix.elements,oe=ae[14]/(ae[10]-1),Ae=ae[14]/(ae[10]+1),Te=(ae[9]+1)/ae[5],Be=(ae[9]-1)/ae[5],C=(ae[8]-1)/ae[0],Je=(ne[8]+1)/ne[0],ue=oe*C,pe=oe*Je,ie=z/(-C+Je),le=ie*-C;if(J.matrixWorld.decompose(G.position,G.quaternion,G.scale),G.translateX(le),G.translateZ(ie),G.matrixWorld.compose(G.position,G.quaternion,G.scale),G.matrixWorldInverse.copy(G.matrixWorld).invert(),ae[10]===-1)G.projectionMatrix.copy(J.projectionMatrix),G.projectionMatrixInverse.copy(J.projectionMatrixInverse);else{const ge=oe+ie,E=Ae+ie,v=ue-le,V=pe+(z-le),Q=Te*Ae/E*ge,re=Be*Ae/E*ge;G.projectionMatrix.makePerspective(v,V,Q,re,ge,E),G.projectionMatrixInverse.copy(G.projectionMatrix).invert()}}function Z(G,J){J===null?G.matrixWorld.copy(G.matrix):G.matrixWorld.multiplyMatrices(J.matrixWorld,G.matrix),G.matrixWorldInverse.copy(G.matrixWorld).invert()}this.updateCamera=function(G){if(n===null)return;let J=G.near,q=G.far;_.texture!==null&&(_.depthNear>0&&(J=_.depthNear),_.depthFar>0&&(q=_.depthFar)),M.near=O.near=A.near=J,M.far=O.far=A.far=q,(R!==M.near||H!==M.far)&&(n.updateRenderState({depthNear:M.near,depthFar:M.far}),R=M.near,H=M.far),A.layers.mask=G.layers.mask|2,O.layers.mask=G.layers.mask|4,M.layers.mask=A.layers.mask|O.layers.mask;const z=G.parent,ae=M.cameras;Z(M,z);for(let ne=0;ne<ae.length;ne++)Z(ae[ne],z);ae.length===2?k(M,A,O):M.projectionMatrix.copy(A.projectionMatrix),ce(G,M,z)};function ce(G,J,q){q===null?G.matrix.copy(J.matrixWorld):(G.matrix.copy(q.matrixWorld),G.matrix.invert(),G.matrix.multiply(J.matrixWorld)),G.matrix.decompose(G.position,G.quaternion,G.scale),G.updateMatrixWorld(!0),G.projectionMatrix.copy(J.projectionMatrix),G.projectionMatrixInverse.copy(J.projectionMatrixInverse),G.isPerspectiveCamera&&(G.fov=ss*2*Math.atan(1/G.projectionMatrix.elements[5]),G.zoom=1)}this.getCamera=function(){return M},this.getFoveation=function(){if(!(p===null&&f===null))return l},this.setFoveation=function(G){l=G,p!==null&&(p.fixedFoveation=G),f!==null&&f.fixedFoveation!==void 0&&(f.fixedFoveation=G)},this.hasDepthSensing=function(){return _.texture!==null},this.getDepthSensingMesh=function(){return _.getMesh(M)};let ve=null;function De(G,J){if(h=J.getViewerPose(c||o),g=J,h!==null){const q=h.views;f!==null&&(e.setRenderTargetFramebuffer(S,f.framebuffer),e.setRenderTarget(S));let z=!1;q.length!==M.cameras.length&&(M.cameras.length=0,z=!0);for(let ne=0;ne<q.length;ne++){const oe=q[ne];let Ae=null;if(f!==null)Ae=f.getViewport(oe);else{const Be=u.getViewSubImage(p,oe);Ae=Be.viewport,ne===0&&(e.setRenderTargetTextures(S,Be.colorTexture,p.ignoreDepthValues?void 0:Be.depthStencilTexture),e.setRenderTarget(S))}let Te=b[ne];Te===void 0&&(Te=new Ft,Te.layers.enable(ne),Te.viewport=new dt,b[ne]=Te),Te.matrix.fromArray(oe.transform.matrix),Te.matrix.decompose(Te.position,Te.quaternion,Te.scale),Te.projectionMatrix.fromArray(oe.projectionMatrix),Te.projectionMatrixInverse.copy(Te.projectionMatrix).invert(),Te.viewport.set(Ae.x,Ae.y,Ae.width,Ae.height),ne===0&&(M.matrix.copy(Te.matrix),M.matrix.decompose(M.position,M.quaternion,M.scale)),z===!0&&M.cameras.push(Te)}const ae=n.enabledFeatures;if(ae&&ae.includes("depth-sensing")){const ne=u.getDepthInformation(q[0]);ne&&ne.isValid&&ne.texture&&_.init(e,ne,n.renderState)}}for(let q=0;q<T.length;q++){const z=x[q],ae=T[q];z!==null&&ae!==void 0&&ae.update(z,J,c||o)}ve&&ve(G,J),J.detectedPlanes&&r.dispatchEvent({type:"planesdetected",data:J}),g=null}const xe=new vl;xe.setAnimationLoop(De),this.setAnimationLoop=function(G){ve=G},this.dispose=function(){}}}const Dr=new Yt,Of=new Qe;function Ff(i,e){function t(m,d){m.matrixAutoUpdate===!0&&m.updateMatrix(),d.value.copy(m.matrix)}function r(m,d){d.color.getRGB(m.fogColor.value,ml(i)),d.isFog?(m.fogNear.value=d.near,m.fogFar.value=d.far):d.isFogExp2&&(m.fogDensity.value=d.density)}function n(m,d,S,T,x){d.isMeshBasicMaterial||d.isMeshLambertMaterial?a(m,d):d.isMeshToonMaterial?(a(m,d),u(m,d)):d.isMeshPhongMaterial?(a(m,d),h(m,d)):d.isMeshStandardMaterial?(a(m,d),p(m,d),d.isMeshPhysicalMaterial&&f(m,d,x)):d.isMeshMatcapMaterial?(a(m,d),g(m,d)):d.isMeshDepthMaterial?a(m,d):d.isMeshDistanceMaterial?(a(m,d),_(m,d)):d.isMeshNormalMaterial?a(m,d):d.isLineBasicMaterial?(o(m,d),d.isLineDashedMaterial&&s(m,d)):d.isPointsMaterial?l(m,d,S,T):d.isSpriteMaterial?c(m,d):d.isShadowMaterial?(m.color.value.copy(d.color),m.opacity.value=d.opacity):d.isShaderMaterial&&(d.uniformsNeedUpdate=!1)}function a(m,d){m.opacity.value=d.opacity,d.color&&m.diffuse.value.copy(d.color),d.emissive&&m.emissive.value.copy(d.emissive).multiplyScalar(d.emissiveIntensity),d.map&&(m.map.value=d.map,t(d.map,m.mapTransform)),d.alphaMap&&(m.alphaMap.value=d.alphaMap,t(d.alphaMap,m.alphaMapTransform)),d.bumpMap&&(m.bumpMap.value=d.bumpMap,t(d.bumpMap,m.bumpMapTransform),m.bumpScale.value=d.bumpScale,d.side===Dt&&(m.bumpScale.value*=-1)),d.normalMap&&(m.normalMap.value=d.normalMap,t(d.normalMap,m.normalMapTransform),m.normalScale.value.copy(d.normalScale),d.side===Dt&&m.normalScale.value.negate()),d.displacementMap&&(m.displacementMap.value=d.displacementMap,t(d.displacementMap,m.displacementMapTransform),m.displacementScale.value=d.displacementScale,m.displacementBias.value=d.displacementBias),d.emissiveMap&&(m.emissiveMap.value=d.emissiveMap,t(d.emissiveMap,m.emissiveMapTransform)),d.specularMap&&(m.specularMap.value=d.specularMap,t(d.specularMap,m.specularMapTransform)),d.alphaTest>0&&(m.alphaTest.value=d.alphaTest);const S=e.get(d),T=S.envMap,x=S.envMapRotation;T&&(m.envMap.value=T,Dr.copy(x),Dr.x*=-1,Dr.y*=-1,Dr.z*=-1,T.isCubeTexture&&T.isRenderTargetTexture===!1&&(Dr.y*=-1,Dr.z*=-1),m.envMapRotation.value.setFromMatrix4(Of.makeRotationFromEuler(Dr)),m.flipEnvMap.value=T.isCubeTexture&&T.isRenderTargetTexture===!1?-1:1,m.reflectivity.value=d.reflectivity,m.ior.value=d.ior,m.refractionRatio.value=d.refractionRatio),d.lightMap&&(m.lightMap.value=d.lightMap,m.lightMapIntensity.value=d.lightMapIntensity,t(d.lightMap,m.lightMapTransform)),d.aoMap&&(m.aoMap.value=d.aoMap,m.aoMapIntensity.value=d.aoMapIntensity,t(d.aoMap,m.aoMapTransform))}function o(m,d){m.diffuse.value.copy(d.color),m.opacity.value=d.opacity,d.map&&(m.map.value=d.map,t(d.map,m.mapTransform))}function s(m,d){m.dashSize.value=d.dashSize,m.totalSize.value=d.dashSize+d.gapSize,m.scale.value=d.scale}function l(m,d,S,T){m.diffuse.value.copy(d.color),m.opacity.value=d.opacity,m.size.value=d.size*S,m.scale.value=T*.5,d.map&&(m.map.value=d.map,t(d.map,m.uvTransform)),d.alphaMap&&(m.alphaMap.value=d.alphaMap,t(d.alphaMap,m.alphaMapTransform)),d.alphaTest>0&&(m.alphaTest.value=d.alphaTest)}function c(m,d){m.diffuse.value.copy(d.color),m.opacity.value=d.opacity,m.rotation.value=d.rotation,d.map&&(m.map.value=d.map,t(d.map,m.mapTransform)),d.alphaMap&&(m.alphaMap.value=d.alphaMap,t(d.alphaMap,m.alphaMapTransform)),d.alphaTest>0&&(m.alphaTest.value=d.alphaTest)}function h(m,d){m.specular.value.copy(d.specular),m.shininess.value=Math.max(d.shininess,1e-4)}function u(m,d){d.gradientMap&&(m.gradientMap.value=d.gradientMap)}function p(m,d){m.metalness.value=d.metalness,d.metalnessMap&&(m.metalnessMap.value=d.metalnessMap,t(d.metalnessMap,m.metalnessMapTransform)),m.roughness.value=d.roughness,d.roughnessMap&&(m.roughnessMap.value=d.roughnessMap,t(d.roughnessMap,m.roughnessMapTransform)),d.envMap&&(m.envMapIntensity.value=d.envMapIntensity)}function f(m,d,S){m.ior.value=d.ior,d.sheen>0&&(m.sheenColor.value.copy(d.sheenColor).multiplyScalar(d.sheen),m.sheenRoughness.value=d.sheenRoughness,d.sheenColorMap&&(m.sheenColorMap.value=d.sheenColorMap,t(d.sheenColorMap,m.sheenColorMapTransform)),d.sheenRoughnessMap&&(m.sheenRoughnessMap.value=d.sheenRoughnessMap,t(d.sheenRoughnessMap,m.sheenRoughnessMapTransform))),d.clearcoat>0&&(m.clearcoat.value=d.clearcoat,m.clearcoatRoughness.value=d.clearcoatRoughness,d.clearcoatMap&&(m.clearcoatMap.value=d.clearcoatMap,t(d.clearcoatMap,m.clearcoatMapTransform)),d.clearcoatRoughnessMap&&(m.clearcoatRoughnessMap.value=d.clearcoatRoughnessMap,t(d.clearcoatRoughnessMap,m.clearcoatRoughnessMapTransform)),d.clearcoatNormalMap&&(m.clearcoatNormalMap.value=d.clearcoatNormalMap,t(d.clearcoatNormalMap,m.clearcoatNormalMapTransform),m.clearcoatNormalScale.value.copy(d.clearcoatNormalScale),d.side===Dt&&m.clearcoatNormalScale.value.negate())),d.dispersion>0&&(m.dispersion.value=d.dispersion),d.iridescence>0&&(m.iridescence.value=d.iridescence,m.iridescenceIOR.value=d.iridescenceIOR,m.iridescenceThicknessMinimum.value=d.iridescenceThicknessRange[0],m.iridescenceThicknessMaximum.value=d.iridescenceThicknessRange[1],d.iridescenceMap&&(m.iridescenceMap.value=d.iridescenceMap,t(d.iridescenceMap,m.iridescenceMapTransform)),d.iridescenceThicknessMap&&(m.iridescenceThicknessMap.value=d.iridescenceThicknessMap,t(d.iridescenceThicknessMap,m.iridescenceThicknessMapTransform))),d.transmission>0&&(m.transmission.value=d.transmission,m.transmissionSamplerMap.value=S.texture,m.transmissionSamplerSize.value.set(S.width,S.height),d.transmissionMap&&(m.transmissionMap.value=d.transmissionMap,t(d.transmissionMap,m.transmissionMapTransform)),m.thickness.value=d.thickness,d.thicknessMap&&(m.thicknessMap.value=d.thicknessMap,t(d.thicknessMap,m.thicknessMapTransform)),m.attenuationDistance.value=d.attenuationDistance,m.attenuationColor.value.copy(d.attenuationColor)),d.anisotropy>0&&(m.anisotropyVector.value.set(d.anisotropy*Math.cos(d.anisotropyRotation),d.anisotropy*Math.sin(d.anisotropyRotation)),d.anisotropyMap&&(m.anisotropyMap.value=d.anisotropyMap,t(d.anisotropyMap,m.anisotropyMapTransform))),m.specularIntensity.value=d.specularIntensity,m.specularColor.value.copy(d.specularColor),d.specularColorMap&&(m.specularColorMap.value=d.specularColorMap,t(d.specularColorMap,m.specularColorMapTransform)),d.specularIntensityMap&&(m.specularIntensityMap.value=d.specularIntensityMap,t(d.specularIntensityMap,m.specularIntensityMapTransform))}function g(m,d){d.matcap&&(m.matcap.value=d.matcap)}function _(m,d){const S=e.get(d).light;m.referencePosition.value.setFromMatrixPosition(S.matrixWorld),m.nearDistance.value=S.shadow.camera.near,m.farDistance.value=S.shadow.camera.far}return{refreshFogUniforms:r,refreshMaterialUniforms:n}}function zf(i,e,t,r){let n={},a={},o=[];const s=i.getParameter(i.MAX_UNIFORM_BUFFER_BINDINGS);function l(S,T){const x=T.program;r.uniformBlockBinding(S,x)}function c(S,T){let x=n[S.id];x===void 0&&(g(S),x=h(S),n[S.id]=x,S.addEventListener("dispose",m));const N=T.program;r.updateUBOMapping(S,N);const w=e.render.frame;a[S.id]!==w&&(p(S),a[S.id]=w)}function h(S){const T=u();S.__bindingPointIndex=T;const x=i.createBuffer(),N=S.__size,w=S.usage;return i.bindBuffer(i.UNIFORM_BUFFER,x),i.bufferData(i.UNIFORM_BUFFER,N,w),i.bindBuffer(i.UNIFORM_BUFFER,null),i.bindBufferBase(i.UNIFORM_BUFFER,T,x),x}function u(){for(let S=0;S<s;S++)if(o.indexOf(S)===-1)return o.push(S),S;return console.error("THREE.WebGLRenderer: Maximum number of simultaneously usable uniforms groups reached."),0}function p(S){const T=n[S.id],x=S.uniforms,N=S.__cache;i.bindBuffer(i.UNIFORM_BUFFER,T);for(let w=0,A=x.length;w<A;w++){const O=Array.isArray(x[w])?x[w]:[x[w]];for(let b=0,M=O.length;b<M;b++){const R=O[b];if(f(R,w,b,N)===!0){const H=R.__offset,U=Array.isArray(R.value)?R.value:[R.value];let F=0;for(let K=0;K<U.length;K++){const X=U[K],$=_(X);typeof X=="number"||typeof X=="boolean"?(R.__data[0]=X,i.bufferSubData(i.UNIFORM_BUFFER,H+F,R.__data)):X.isMatrix3?(R.__data[0]=X.elements[0],R.__data[1]=X.elements[1],R.__data[2]=X.elements[2],R.__data[3]=0,R.__data[4]=X.elements[3],R.__data[5]=X.elements[4],R.__data[6]=X.elements[5],R.__data[7]=0,R.__data[8]=X.elements[6],R.__data[9]=X.elements[7],R.__data[10]=X.elements[8],R.__data[11]=0):(X.toArray(R.__data,F),F+=$.storage/Float32Array.BYTES_PER_ELEMENT)}i.bufferSubData(i.UNIFORM_BUFFER,H,R.__data)}}}i.bindBuffer(i.UNIFORM_BUFFER,null)}function f(S,T,x,N){const w=S.value,A=T+"_"+x;if(N[A]===void 0)return typeof w=="number"||typeof w=="boolean"?N[A]=w:N[A]=w.clone(),!0;{const O=N[A];if(typeof w=="number"||typeof w=="boolean"){if(O!==w)return N[A]=w,!0}else if(O.equals(w)===!1)return O.copy(w),!0}return!1}function g(S){const T=S.uniforms;let x=0;const N=16;for(let A=0,O=T.length;A<O;A++){const b=Array.isArray(T[A])?T[A]:[T[A]];for(let M=0,R=b.length;M<R;M++){const H=b[M],U=Array.isArray(H.value)?H.value:[H.value];for(let F=0,K=U.length;F<K;F++){const X=U[F],$=_(X),k=x%N,Z=k%$.boundary,ce=k+Z;x+=Z,ce!==0&&N-ce<$.storage&&(x+=N-ce),H.__data=new Float32Array($.storage/Float32Array.BYTES_PER_ELEMENT),H.__offset=x,x+=$.storage}}}const w=x%N;return w>0&&(x+=N-w),S.__size=x,S.__cache={},this}function _(S){const T={boundary:0,storage:0};return typeof S=="number"||typeof S=="boolean"?(T.boundary=4,T.storage=4):S.isVector2?(T.boundary=8,T.storage=8):S.isVector3||S.isColor?(T.boundary=16,T.storage=12):S.isVector4?(T.boundary=16,T.storage=16):S.isMatrix3?(T.boundary=48,T.storage=48):S.isMatrix4?(T.boundary=64,T.storage=64):S.isTexture?console.warn("THREE.WebGLRenderer: Texture samplers can not be part of an uniforms group."):console.warn("THREE.WebGLRenderer: Unsupported uniform value type.",S),T}function m(S){const T=S.target;T.removeEventListener("dispose",m);const x=o.indexOf(T.__bindingPointIndex);o.splice(x,1),i.deleteBuffer(n[T.id]),delete n[T.id],delete a[T.id]}function d(){for(const S in n)i.deleteBuffer(n[S]);o=[],n={},a={}}return{bind:l,update:c,dispose:d}}class Bf{constructor(e={}){const{canvas:t=wc(),context:r=null,depth:n=!0,stencil:a=!1,alpha:o=!1,antialias:s=!1,premultipliedAlpha:l=!0,preserveDrawingBuffer:c=!1,powerPreference:h="default",failIfMajorPerformanceCaveat:u=!1,reverseDepthBuffer:p=!1}=e;this.isWebGLRenderer=!0;let f;if(r!==null){if(typeof WebGLRenderingContext<"u"&&r instanceof WebGLRenderingContext)throw new Error("THREE.WebGLRenderer: WebGL 1 is not supported since r163.");f=r.getContextAttributes().alpha}else f=o;const g=new Uint32Array(4),_=new Int32Array(4);let m=null,d=null;const S=[],T=[];this.domElement=t,this.debug={checkShaderErrors:!0,onShaderError:null},this.autoClear=!0,this.autoClearColor=!0,this.autoClearDepth=!0,this.autoClearStencil=!0,this.sortObjects=!0,this.clippingPlanes=[],this.localClippingEnabled=!1,this._outputColorSpace=Tt,this.toneMapping=yr,this.toneMappingExposure=1;const x=this;let N=!1,w=0,A=0,O=null,b=-1,M=null;const R=new dt,H=new dt;let U=null;const F=new We(0);let K=0,X=t.width,$=t.height,k=1,Z=null,ce=null;const ve=new dt(0,0,X,$),De=new dt(0,0,X,$);let xe=!1;const G=new ys;let J=!1,q=!1;const z=new Qe,ae=new Qe,ne=new B,oe=new dt,Ae={background:null,fog:null,environment:null,overrideMaterial:null,isScene:!0};let Te=!1;function Be(){return O===null?k:1}let C=r;function Je(y,I){return t.getContext(y,I)}try{const y={alpha:!0,depth:n,stencil:a,antialias:s,premultipliedAlpha:l,preserveDrawingBuffer:c,powerPreference:h,failIfMajorPerformanceCaveat:u};if("setAttribute"in t&&t.setAttribute("data-engine",`three.js r${us}`),t.addEventListener("webglcontextlost",te,!1),t.addEventListener("webglcontextrestored",ye,!1),t.addEventListener("webglcontextcreationerror",de,!1),C===null){const I="webgl2";if(C=Je(I,y),C===null)throw Je(I)?new Error("Error creating WebGL context with your selected attributes."):new Error("Error creating WebGL context.")}}catch(y){throw console.error("THREE.WebGLRenderer: "+y.message),y}let ue,pe,ie,le,ge,E,v,V,Q,re,ee,Re,_e,be,qe,he,Se,Ue,Oe,Ee,Xe,He,nt,D;function me(){ue=new Gd(C),ue.init(),He=new Cf(C,ue),pe=new Fd(C,ue,e,He),ie=new wf(C,ue),pe.reverseDepthBuffer&&p&&ie.buffers.depth.setReversed(!0),le=new Xd(C),ge=new uf,E=new Rf(C,ue,ie,ge,pe,He,le),v=new Bd(x),V=new Hd(x),Q=new Qc(C),nt=new Ud(C,Q),re=new Wd(C,Q,le,nt),ee=new qd(C,re,Q,le),Oe=new Yd(C,pe,E),he=new zd(ge),Re=new hf(x,v,V,ue,pe,nt,he),_e=new Ff(x,ge),be=new pf,qe=new xf(ue),Ue=new Nd(x,v,V,ie,ee,f,l),Se=new bf(x,ee,pe),D=new zf(C,le,pe,ie),Ee=new Od(C,ue,le),Xe=new jd(C,ue,le),le.programs=Re.programs,x.capabilities=pe,x.extensions=ue,x.properties=ge,x.renderLists=be,x.shadowMap=Se,x.state=ie,x.info=le}me();const Y=new Uf(x,C);this.xr=Y,this.getContext=function(){return C},this.getContextAttributes=function(){return C.getContextAttributes()},this.forceContextLoss=function(){const y=ue.get("WEBGL_lose_context");y&&y.loseContext()},this.forceContextRestore=function(){const y=ue.get("WEBGL_lose_context");y&&y.restoreContext()},this.getPixelRatio=function(){return k},this.setPixelRatio=function(y){y!==void 0&&(k=y,this.setSize(X,$,!1))},this.getSize=function(y){return y.set(X,$)},this.setSize=function(y,I,W=!0){if(Y.isPresenting){console.warn("THREE.WebGLRenderer: Can't change size while VR device is presenting.");return}X=y,$=I,t.width=Math.floor(y*k),t.height=Math.floor(I*k),W===!0&&(t.style.width=y+"px",t.style.height=I+"px"),this.setViewport(0,0,y,I)},this.getDrawingBufferSize=function(y){return y.set(X*k,$*k).floor()},this.setDrawingBufferSize=function(y,I,W){X=y,$=I,k=W,t.width=Math.floor(y*W),t.height=Math.floor(I*W),this.setViewport(0,0,y,I)},this.getCurrentViewport=function(y){return y.copy(R)},this.getViewport=function(y){return y.copy(ve)},this.setViewport=function(y,I,W,j){y.isVector4?ve.set(y.x,y.y,y.z,y.w):ve.set(y,I,W,j),ie.viewport(R.copy(ve).multiplyScalar(k).round())},this.getScissor=function(y){return y.copy(De)},this.setScissor=function(y,I,W,j){y.isVector4?De.set(y.x,y.y,y.z,y.w):De.set(y,I,W,j),ie.scissor(H.copy(De).multiplyScalar(k).round())},this.getScissorTest=function(){return xe},this.setScissorTest=function(y){ie.setScissorTest(xe=y)},this.setOpaqueSort=function(y){Z=y},this.setTransparentSort=function(y){ce=y},this.getClearColor=function(y){return y.copy(Ue.getClearColor())},this.setClearColor=function(){Ue.setClearColor.apply(Ue,arguments)},this.getClearAlpha=function(){return Ue.getClearAlpha()},this.setClearAlpha=function(){Ue.setClearAlpha.apply(Ue,arguments)},this.clear=function(y=!0,I=!0,W=!0){let j=0;if(y){let P=!1;if(O!==null){const se=O.texture.format;P=se===xs||se===vs||se===_s}if(P){const se=O.texture.type,Me=se===cr||se===kr||se===ki||se===_i||se===fs||se===ms,we=Ue.getClearColor(),Ce=Ue.getClearAlpha(),ze=we.r,ke=we.g,Ie=we.b;Me?(g[0]=ze,g[1]=ke,g[2]=Ie,g[3]=Ce,C.clearBufferuiv(C.COLOR,0,g)):(_[0]=ze,_[1]=ke,_[2]=Ie,_[3]=Ce,C.clearBufferiv(C.COLOR,0,_))}else j|=C.COLOR_BUFFER_BIT}I&&(j|=C.DEPTH_BUFFER_BIT),W&&(j|=C.STENCIL_BUFFER_BIT,this.state.buffers.stencil.setMask(4294967295)),C.clear(j)},this.clearColor=function(){this.clear(!0,!1,!1)},this.clearDepth=function(){this.clear(!1,!0,!1)},this.clearStencil=function(){this.clear(!1,!1,!0)},this.dispose=function(){t.removeEventListener("webglcontextlost",te,!1),t.removeEventListener("webglcontextrestored",ye,!1),t.removeEventListener("webglcontextcreationerror",de,!1),be.dispose(),qe.dispose(),ge.dispose(),v.dispose(),V.dispose(),ee.dispose(),nt.dispose(),D.dispose(),Re.dispose(),Y.dispose(),Y.removeEventListener("sessionstart",Es),Y.removeEventListener("sessionend",bs),br.stop()};function te(y){y.preventDefault(),console.log("THREE.WebGLRenderer: Context Lost."),N=!0}function ye(){console.log("THREE.WebGLRenderer: Context Restored."),N=!1;const y=le.autoReset,I=Se.enabled,W=Se.autoUpdate,j=Se.needsUpdate,P=Se.type;me(),le.autoReset=y,Se.enabled=I,Se.autoUpdate=W,Se.needsUpdate=j,Se.type=P}function de(y){console.error("THREE.WebGLRenderer: A WebGL context could not be created. Reason: ",y.statusMessage)}function Ye(y){const I=y.target;I.removeEventListener("dispose",Ye),ut(I)}function ut(y){St(y),ge.remove(y)}function St(y){const I=ge.get(y).programs;I!==void 0&&(I.forEach(function(W){Re.releaseProgram(W)}),y.isShaderMaterial&&Re.releaseShaderCache(y))}this.renderBufferDirect=function(y,I,W,j,P,se){I===null&&(I=Ae);const Me=P.isMesh&&P.matrixWorld.determinant()<0,we=Rl(y,I,W,j,P);ie.setMaterial(j,Me);let Ce=W.index,ze=1;if(j.wireframe===!0){if(Ce=re.getWireframeAttribute(W),Ce===void 0)return;ze=2}const ke=W.drawRange,Ie=W.attributes.position;let Ze=ke.start*ze,st=(ke.start+ke.count)*ze;se!==null&&(Ze=Math.max(Ze,se.start*ze),st=Math.min(st,(se.start+se.count)*ze)),Ce!==null?(Ze=Math.max(Ze,0),st=Math.min(st,Ce.count)):Ie!=null&&(Ze=Math.max(Ze,0),st=Math.min(st,Ie.count));const ot=st-Ze;if(ot<0||ot===1/0)return;nt.setup(P,j,we,W,Ce);let ft,lt=Ee;if(Ce!==null&&(ft=Q.get(Ce),lt=Xe,lt.setIndex(ft)),P.isMesh)j.wireframe===!0?(ie.setLineWidth(j.wireframeLinewidth*Be()),lt.setMode(C.LINES)):lt.setMode(C.TRIANGLES);else if(P.isLine){let Le=j.linewidth;Le===void 0&&(Le=1),ie.setLineWidth(Le*Be()),P.isLineSegments?lt.setMode(C.LINES):P.isLineLoop?lt.setMode(C.LINE_LOOP):lt.setMode(C.LINE_STRIP)}else P.isPoints?lt.setMode(C.POINTS):P.isSprite&&lt.setMode(C.TRIANGLES);if(P.isBatchedMesh)if(P._multiDrawInstances!==null)lt.renderMultiDrawInstances(P._multiDrawStarts,P._multiDrawCounts,P._multiDrawCount,P._multiDrawInstances);else if(ue.get("WEBGL_multi_draw"))lt.renderMultiDraw(P._multiDrawStarts,P._multiDrawCounts,P._multiDrawCount);else{const Le=P._multiDrawStarts,Tr=P._multiDrawCounts,wr=P._multiDrawCount,Vt=Ce?Q.get(Ce).bytesPerElement:1,Wr=ge.get(j).currentProgram.getUniforms();for(let It=0;It<wr;It++)Wr.setValue(C,"_gl_DrawID",It),lt.render(Le[It]/Vt,Tr[It])}else if(P.isInstancedMesh)lt.renderInstances(Ze,ot,P.count);else if(W.isInstancedBufferGeometry){const Le=W._maxInstanceCount!==void 0?W._maxInstanceCount:1/0,Tr=Math.min(W.instanceCount,Le);lt.renderInstances(Ze,ot,Tr)}else lt.render(Ze,ot)};function et(y,I,W){y.transparent===!0&&y.side===Rt&&y.forceSinglePass===!1?(y.side=Dt,y.needsUpdate=!0,ji(y,I,W),y.side=Sr,y.needsUpdate=!0,ji(y,I,W),y.side=Rt):ji(y,I,W)}this.compile=function(y,I,W=null){W===null&&(W=y),d=qe.get(W),d.init(I),T.push(d),W.traverseVisible(function(P){P.isLight&&P.layers.test(I.layers)&&(d.pushLight(P),P.castShadow&&d.pushShadow(P))}),y!==W&&y.traverseVisible(function(P){P.isLight&&P.layers.test(I.layers)&&(d.pushLight(P),P.castShadow&&d.pushShadow(P))}),d.setupLights();const j=new Set;return y.traverse(function(P){if(!(P.isMesh||P.isPoints||P.isLine||P.isSprite))return;const se=P.material;if(se)if(Array.isArray(se))for(let Me=0;Me<se.length;Me++){const we=se[Me];et(we,W,P),j.add(we)}else et(se,W,P),j.add(se)}),T.pop(),d=null,j},this.compileAsync=function(y,I,W=null){const j=this.compile(y,I,W);return new Promise(P=>{function se(){if(j.forEach(function(Me){ge.get(Me).currentProgram.isReady()&&j.delete(Me)}),j.size===0){P(y);return}setTimeout(se,10)}ue.get("KHR_parallel_shader_compile")!==null?se():setTimeout(se,10)})};let kt=null;function er(y){kt&&kt(y)}function Es(){br.stop()}function bs(){br.start()}const br=new vl;br.setAnimationLoop(er),typeof self<"u"&&br.setContext(self),this.setAnimationLoop=function(y){kt=y,Y.setAnimationLoop(y),y===null?br.stop():br.start()},Y.addEventListener("sessionstart",Es),Y.addEventListener("sessionend",bs),this.render=function(y,I){if(I!==void 0&&I.isCamera!==!0){console.error("THREE.WebGLRenderer.render: camera is not an instance of THREE.Camera.");return}if(N===!0)return;if(y.matrixWorldAutoUpdate===!0&&y.updateMatrixWorld(),I.parent===null&&I.matrixWorldAutoUpdate===!0&&I.updateMatrixWorld(),Y.enabled===!0&&Y.isPresenting===!0&&(Y.cameraAutoUpdate===!0&&Y.updateCamera(I),I=Y.getCamera()),y.isScene===!0&&y.onBeforeRender(x,y,I,O),d=qe.get(y,T.length),d.init(I),T.push(d),ae.multiplyMatrices(I.projectionMatrix,I.matrixWorldInverse),G.setFromProjectionMatrix(ae),q=this.localClippingEnabled,J=he.init(this.clippingPlanes,q),m=be.get(y,S.length),m.init(),S.push(m),Y.enabled===!0&&Y.isPresenting===!0){const se=x.xr.getDepthSensingMesh();se!==null&&zn(se,I,-1/0,x.sortObjects)}zn(y,I,0,x.sortObjects),m.finish(),x.sortObjects===!0&&m.sort(Z,ce),Te=Y.enabled===!1||Y.isPresenting===!1||Y.hasDepthSensing()===!1,Te&&Ue.addToRenderList(m,y),this.info.render.frame++,J===!0&&he.beginShadows();const W=d.state.shadowsArray;Se.render(W,y,I),J===!0&&he.endShadows(),this.info.autoReset===!0&&this.info.reset();const j=m.opaque,P=m.transmissive;if(d.setupLights(),I.isArrayCamera){const se=I.cameras;if(P.length>0)for(let Me=0,we=se.length;Me<we;Me++){const Ce=se[Me];ws(j,P,y,Ce)}Te&&Ue.render(y);for(let Me=0,we=se.length;Me<we;Me++){const Ce=se[Me];Ts(m,y,Ce,Ce.viewport)}}else P.length>0&&ws(j,P,y,I),Te&&Ue.render(y),Ts(m,y,I);O!==null&&(E.updateMultisampleRenderTarget(O),E.updateRenderTargetMipmap(O)),y.isScene===!0&&y.onAfterRender(x,y,I),nt.resetDefaultState(),b=-1,M=null,T.pop(),T.length>0?(d=T[T.length-1],J===!0&&he.setGlobalState(x.clippingPlanes,d.state.camera)):d=null,S.pop(),S.length>0?m=S[S.length-1]:m=null};function zn(y,I,W,j){if(y.visible===!1)return;if(y.layers.test(I.layers)){if(y.isGroup)W=y.renderOrder;else if(y.isLOD)y.autoUpdate===!0&&y.update(I);else if(y.isLight)d.pushLight(y),y.castShadow&&d.pushShadow(y);else if(y.isSprite){if(!y.frustumCulled||G.intersectsSprite(y)){j&&oe.setFromMatrixPosition(y.matrixWorld).applyMatrix4(ae);const se=ee.update(y),Me=y.material;Me.visible&&m.push(y,se,Me,W,oe.z,null)}}else if((y.isMesh||y.isLine||y.isPoints)&&(!y.frustumCulled||G.intersectsObject(y))){const se=ee.update(y),Me=y.material;if(j&&(y.boundingSphere!==void 0?(y.boundingSphere===null&&y.computeBoundingSphere(),oe.copy(y.boundingSphere.center)):(se.boundingSphere===null&&se.computeBoundingSphere(),oe.copy(se.boundingSphere.center)),oe.applyMatrix4(y.matrixWorld).applyMatrix4(ae)),Array.isArray(Me)){const we=se.groups;for(let Ce=0,ze=we.length;Ce<ze;Ce++){const ke=we[Ce],Ie=Me[ke.materialIndex];Ie&&Ie.visible&&m.push(y,se,Ie,W,oe.z,ke)}}else Me.visible&&m.push(y,se,Me,W,oe.z,null)}}const P=y.children;for(let se=0,Me=P.length;se<Me;se++)zn(P[se],I,W,j)}function Ts(y,I,W,j){const P=y.opaque,se=y.transmissive,Me=y.transparent;d.setupLightsView(W),J===!0&&he.setGlobalState(x.clippingPlanes,W),j&&ie.viewport(R.copy(j)),P.length>0&&Wi(P,I,W),se.length>0&&Wi(se,I,W),Me.length>0&&Wi(Me,I,W),ie.buffers.depth.setTest(!0),ie.buffers.depth.setMask(!0),ie.buffers.color.setMask(!0),ie.setPolygonOffset(!1)}function ws(y,I,W,j){if((W.isScene===!0?W.overrideMaterial:null)!==null)return;d.state.transmissionRenderTarget[j.id]===void 0&&(d.state.transmissionRenderTarget[j.id]=new Vr(1,1,{generateMipmaps:!0,type:ue.has("EXT_color_buffer_half_float")||ue.has("EXT_color_buffer_float")?Hi:cr,minFilter:Br,samples:4,stencilBuffer:a,resolveDepthBuffer:!1,resolveStencilBuffer:!1,colorSpace:$e.workingColorSpace}));const P=d.state.transmissionRenderTarget[j.id],se=j.viewport||R;P.setSize(se.z,se.w);const Me=x.getRenderTarget();x.setRenderTarget(P),x.getClearColor(F),K=x.getClearAlpha(),K<1&&x.setClearColor(16777215,.5),x.clear(),Te&&Ue.render(W);const we=x.toneMapping;x.toneMapping=yr;const Ce=j.viewport;if(j.viewport!==void 0&&(j.viewport=void 0),d.setupLightsView(j),J===!0&&he.setGlobalState(x.clippingPlanes,j),Wi(y,W,j),E.updateMultisampleRenderTarget(P),E.updateRenderTargetMipmap(P),ue.has("WEBGL_multisampled_render_to_texture")===!1){let ze=!1;for(let ke=0,Ie=I.length;ke<Ie;ke++){const Ze=I[ke],st=Ze.object,ot=Ze.geometry,ft=Ze.material,lt=Ze.group;if(ft.side===Rt&&st.layers.test(j.layers)){const Le=ft.side;ft.side=Dt,ft.needsUpdate=!0,As(st,W,j,ot,ft,lt),ft.side=Le,ft.needsUpdate=!0,ze=!0}}ze===!0&&(E.updateMultisampleRenderTarget(P),E.updateRenderTargetMipmap(P))}x.setRenderTarget(Me),x.setClearColor(F,K),Ce!==void 0&&(j.viewport=Ce),x.toneMapping=we}function Wi(y,I,W){const j=I.isScene===!0?I.overrideMaterial:null;for(let P=0,se=y.length;P<se;P++){const Me=y[P],we=Me.object,Ce=Me.geometry,ze=j===null?Me.material:j,ke=Me.group;we.layers.test(W.layers)&&As(we,I,W,Ce,ze,ke)}}function As(y,I,W,j,P,se){y.onBeforeRender(x,I,W,j,P,se),y.modelViewMatrix.multiplyMatrices(W.matrixWorldInverse,y.matrixWorld),y.normalMatrix.getNormalMatrix(y.modelViewMatrix),P.onBeforeRender(x,I,W,j,y,se),P.transparent===!0&&P.side===Rt&&P.forceSinglePass===!1?(P.side=Dt,P.needsUpdate=!0,x.renderBufferDirect(W,I,j,P,y,se),P.side=Sr,P.needsUpdate=!0,x.renderBufferDirect(W,I,j,P,y,se),P.side=Rt):x.renderBufferDirect(W,I,j,P,y,se),y.onAfterRender(x,I,W,j,P,se)}function ji(y,I,W){I.isScene!==!0&&(I=Ae);const j=ge.get(y),P=d.state.lights,se=d.state.shadowsArray,Me=P.state.version,we=Re.getParameters(y,P.state,se,I,W),Ce=Re.getProgramCacheKey(we);let ze=j.programs;j.environment=y.isMeshStandardMaterial?I.environment:null,j.fog=I.fog,j.envMap=(y.isMeshStandardMaterial?V:v).get(y.envMap||j.environment),j.envMapRotation=j.environment!==null&&y.envMap===null?I.environmentRotation:y.envMapRotation,ze===void 0&&(y.addEventListener("dispose",Ye),ze=new Map,j.programs=ze);let ke=ze.get(Ce);if(ke!==void 0){if(j.currentProgram===ke&&j.lightsStateVersion===Me)return Cs(y,we),ke}else we.uniforms=Re.getUniforms(y),y.onBeforeCompile(we,x),ke=Re.acquireProgram(we,Ce),ze.set(Ce,ke),j.uniforms=we.uniforms;const Ie=j.uniforms;return(!y.isShaderMaterial&&!y.isRawShaderMaterial||y.clipping===!0)&&(Ie.clippingPlanes=he.uniform),Cs(y,we),j.needsLights=Ll(y),j.lightsStateVersion=Me,j.needsLights&&(Ie.ambientLightColor.value=P.state.ambient,Ie.lightProbe.value=P.state.probe,Ie.directionalLights.value=P.state.directional,Ie.directionalLightShadows.value=P.state.directionalShadow,Ie.spotLights.value=P.state.spot,Ie.spotLightShadows.value=P.state.spotShadow,Ie.rectAreaLights.value=P.state.rectArea,Ie.ltc_1.value=P.state.rectAreaLTC1,Ie.ltc_2.value=P.state.rectAreaLTC2,Ie.pointLights.value=P.state.point,Ie.pointLightShadows.value=P.state.pointShadow,Ie.hemisphereLights.value=P.state.hemi,Ie.directionalShadowMap.value=P.state.directionalShadowMap,Ie.directionalShadowMatrix.value=P.state.directionalShadowMatrix,Ie.spotShadowMap.value=P.state.spotShadowMap,Ie.spotLightMatrix.value=P.state.spotLightMatrix,Ie.spotLightMap.value=P.state.spotLightMap,Ie.pointShadowMap.value=P.state.pointShadowMap,Ie.pointShadowMatrix.value=P.state.pointShadowMatrix),j.currentProgram=ke,j.uniformsList=null,ke}function Rs(y){if(y.uniformsList===null){const I=y.currentProgram.getUniforms();y.uniformsList=Ln.seqWithValue(I.seq,y.uniforms)}return y.uniformsList}function Cs(y,I){const W=ge.get(y);W.outputColorSpace=I.outputColorSpace,W.batching=I.batching,W.batchingColor=I.batchingColor,W.instancing=I.instancing,W.instancingColor=I.instancingColor,W.instancingMorph=I.instancingMorph,W.skinning=I.skinning,W.morphTargets=I.morphTargets,W.morphNormals=I.morphNormals,W.morphColors=I.morphColors,W.morphTargetsCount=I.morphTargetsCount,W.numClippingPlanes=I.numClippingPlanes,W.numIntersection=I.numClipIntersection,W.vertexAlphas=I.vertexAlphas,W.vertexTangents=I.vertexTangents,W.toneMapping=I.toneMapping}function Rl(y,I,W,j,P){I.isScene!==!0&&(I=Ae),E.resetTextureUnits();const se=I.fog,Me=j.isMeshStandardMaterial?I.environment:null,we=O===null?x.outputColorSpace:O.isXRRenderTarget===!0?O.texture.colorSpace:yi,Ce=(j.isMeshStandardMaterial?V:v).get(j.envMap||Me),ze=j.vertexColors===!0&&!!W.attributes.color&&W.attributes.color.itemSize===4,ke=!!W.attributes.tangent&&(!!j.normalMap||j.anisotropy>0),Ie=!!W.morphAttributes.position,Ze=!!W.morphAttributes.normal,st=!!W.morphAttributes.color;let ot=yr;j.toneMapped&&(O===null||O.isXRRenderTarget===!0)&&(ot=x.toneMapping);const ft=W.morphAttributes.position||W.morphAttributes.normal||W.morphAttributes.color,lt=ft!==void 0?ft.length:0,Le=ge.get(j),Tr=d.state.lights;if(J===!0&&(q===!0||y!==M)){const Lt=y===M&&j.id===b;he.setState(j,y,Lt)}let wr=!1;j.version===Le.__version?(Le.needsLights&&Le.lightsStateVersion!==Tr.state.version||Le.outputColorSpace!==we||P.isBatchedMesh&&Le.batching===!1||!P.isBatchedMesh&&Le.batching===!0||P.isBatchedMesh&&Le.batchingColor===!0&&P.colorTexture===null||P.isBatchedMesh&&Le.batchingColor===!1&&P.colorTexture!==null||P.isInstancedMesh&&Le.instancing===!1||!P.isInstancedMesh&&Le.instancing===!0||P.isSkinnedMesh&&Le.skinning===!1||!P.isSkinnedMesh&&Le.skinning===!0||P.isInstancedMesh&&Le.instancingColor===!0&&P.instanceColor===null||P.isInstancedMesh&&Le.instancingColor===!1&&P.instanceColor!==null||P.isInstancedMesh&&Le.instancingMorph===!0&&P.morphTexture===null||P.isInstancedMesh&&Le.instancingMorph===!1&&P.morphTexture!==null||Le.envMap!==Ce||j.fog===!0&&Le.fog!==se||Le.numClippingPlanes!==void 0&&(Le.numClippingPlanes!==he.numPlanes||Le.numIntersection!==he.numIntersection)||Le.vertexAlphas!==ze||Le.vertexTangents!==ke||Le.morphTargets!==Ie||Le.morphNormals!==Ze||Le.morphColors!==st||Le.toneMapping!==ot||Le.morphTargetsCount!==lt)&&(wr=!0):(wr=!0,Le.__version=j.version);let Vt=Le.currentProgram;wr===!0&&(Vt=ji(j,I,P));let Wr=!1,It=!1,bi=!1;const at=Vt.getUniforms(),qt=Le.uniforms;if(ie.useProgram(Vt.program)&&(Wr=!0,It=!0,bi=!0),j.id!==b&&(b=j.id,It=!0),Wr||M!==y){ie.buffers.depth.getReversed()?(z.copy(y.projectionMatrix),Rc(z),Cc(z),at.setValue(C,"projectionMatrix",z)):at.setValue(C,"projectionMatrix",y.projectionMatrix),at.setValue(C,"viewMatrix",y.matrixWorldInverse);const Lt=at.map.cameraPosition;Lt!==void 0&&Lt.setValue(C,ne.setFromMatrixPosition(y.matrixWorld)),pe.logarithmicDepthBuffer&&at.setValue(C,"logDepthBufFC",2/(Math.log(y.far+1)/Math.LN2)),(j.isMeshPhongMaterial||j.isMeshToonMaterial||j.isMeshLambertMaterial||j.isMeshBasicMaterial||j.isMeshStandardMaterial||j.isShaderMaterial)&&at.setValue(C,"isOrthographic",y.isOrthographicCamera===!0),M!==y&&(M=y,It=!0,bi=!0)}if(P.isSkinnedMesh){at.setOptional(C,P,"bindMatrix"),at.setOptional(C,P,"bindMatrixInverse");const Lt=P.skeleton;Lt&&(Lt.boneTexture===null&&Lt.computeBoneTexture(),at.setValue(C,"boneTexture",Lt.boneTexture,E))}P.isBatchedMesh&&(at.setOptional(C,P,"batchingTexture"),at.setValue(C,"batchingTexture",P._matricesTexture,E),at.setOptional(C,P,"batchingIdTexture"),at.setValue(C,"batchingIdTexture",P._indirectTexture,E),at.setOptional(C,P,"batchingColorTexture"),P._colorsTexture!==null&&at.setValue(C,"batchingColorTexture",P._colorsTexture,E));const Ti=W.morphAttributes;if((Ti.position!==void 0||Ti.normal!==void 0||Ti.color!==void 0)&&Oe.update(P,W,Vt),(It||Le.receiveShadow!==P.receiveShadow)&&(Le.receiveShadow=P.receiveShadow,at.setValue(C,"receiveShadow",P.receiveShadow)),j.isMeshGouraudMaterial&&j.envMap!==null&&(qt.envMap.value=Ce,qt.flipEnvMap.value=Ce.isCubeTexture&&Ce.isRenderTargetTexture===!1?-1:1),j.isMeshStandardMaterial&&j.envMap===null&&I.environment!==null&&(qt.envMapIntensity.value=I.environmentIntensity),It&&(at.setValue(C,"toneMappingExposure",x.toneMappingExposure),Le.needsLights&&Cl(qt,bi),se&&j.fog===!0&&_e.refreshFogUniforms(qt,se),_e.refreshMaterialUniforms(qt,j,k,$,d.state.transmissionRenderTarget[y.id]),Ln.upload(C,Rs(Le),qt,E)),j.isShaderMaterial&&j.uniformsNeedUpdate===!0&&(Ln.upload(C,Rs(Le),qt,E),j.uniformsNeedUpdate=!1),j.isSpriteMaterial&&at.setValue(C,"center",P.center),at.setValue(C,"modelViewMatrix",P.modelViewMatrix),at.setValue(C,"normalMatrix",P.normalMatrix),at.setValue(C,"modelMatrix",P.matrixWorld),j.isShaderMaterial||j.isRawShaderMaterial){const Lt=j.uniformsGroups;for(let wi=0,hr=Lt.length;wi<hr;wi++){const Ls=Lt[wi];D.update(Ls,Vt),D.bind(Ls,Vt)}}return Vt}function Cl(y,I){y.ambientLightColor.needsUpdate=I,y.lightProbe.needsUpdate=I,y.directionalLights.needsUpdate=I,y.directionalLightShadows.needsUpdate=I,y.pointLights.needsUpdate=I,y.pointLightShadows.needsUpdate=I,y.spotLights.needsUpdate=I,y.spotLightShadows.needsUpdate=I,y.rectAreaLights.needsUpdate=I,y.hemisphereLights.needsUpdate=I}function Ll(y){return y.isMeshLambertMaterial||y.isMeshToonMaterial||y.isMeshPhongMaterial||y.isMeshStandardMaterial||y.isShadowMaterial||y.isShaderMaterial&&y.lights===!0}this.getActiveCubeFace=function(){return w},this.getActiveMipmapLevel=function(){return A},this.getRenderTarget=function(){return O},this.setRenderTargetTextures=function(y,I,W){ge.get(y.texture).__webglTexture=I,ge.get(y.depthTexture).__webglTexture=W;const j=ge.get(y);j.__hasExternalTextures=!0,j.__autoAllocateDepthBuffer=W===void 0,j.__autoAllocateDepthBuffer||ue.has("WEBGL_multisampled_render_to_texture")===!0&&(console.warn("THREE.WebGLRenderer: Render-to-texture extension was disabled because an external texture was provided"),j.__useRenderToTexture=!1)},this.setRenderTargetFramebuffer=function(y,I){const W=ge.get(y);W.__webglFramebuffer=I,W.__useDefaultFramebuffer=I===void 0},this.setRenderTarget=function(y,I=0,W=0){O=y,w=I,A=W;let j=!0,P=null,se=!1,Me=!1;if(y){const we=ge.get(y);if(we.__useDefaultFramebuffer!==void 0)ie.bindFramebuffer(C.FRAMEBUFFER,null),j=!1;else if(we.__webglFramebuffer===void 0)E.setupRenderTarget(y);else if(we.__hasExternalTextures)E.rebindTextures(y,ge.get(y.texture).__webglTexture,ge.get(y.depthTexture).__webglTexture);else if(y.depthBuffer){const ke=y.depthTexture;if(we.__boundDepthTexture!==ke){if(ke!==null&&ge.has(ke)&&(y.width!==ke.image.width||y.height!==ke.image.height))throw new Error("WebGLRenderTarget: Attached DepthTexture is initialized to the incorrect size.");E.setupDepthRenderbuffer(y)}}const Ce=y.texture;(Ce.isData3DTexture||Ce.isDataArrayTexture||Ce.isCompressedArrayTexture)&&(Me=!0);const ze=ge.get(y).__webglFramebuffer;y.isWebGLCubeRenderTarget?(Array.isArray(ze[I])?P=ze[I][W]:P=ze[I],se=!0):y.samples>0&&E.useMultisampledRTT(y)===!1?P=ge.get(y).__webglMultisampledFramebuffer:Array.isArray(ze)?P=ze[W]:P=ze,R.copy(y.viewport),H.copy(y.scissor),U=y.scissorTest}else R.copy(ve).multiplyScalar(k).floor(),H.copy(De).multiplyScalar(k).floor(),U=xe;if(ie.bindFramebuffer(C.FRAMEBUFFER,P)&&j&&ie.drawBuffers(y,P),ie.viewport(R),ie.scissor(H),ie.setScissorTest(U),se){const we=ge.get(y.texture);C.framebufferTexture2D(C.FRAMEBUFFER,C.COLOR_ATTACHMENT0,C.TEXTURE_CUBE_MAP_POSITIVE_X+I,we.__webglTexture,W)}else if(Me){const we=ge.get(y.texture),Ce=I||0;C.framebufferTextureLayer(C.FRAMEBUFFER,C.COLOR_ATTACHMENT0,we.__webglTexture,W||0,Ce)}b=-1},this.readRenderTargetPixels=function(y,I,W,j,P,se,Me){if(!(y&&y.isWebGLRenderTarget)){console.error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not THREE.WebGLRenderTarget.");return}let we=ge.get(y).__webglFramebuffer;if(y.isWebGLCubeRenderTarget&&Me!==void 0&&(we=we[Me]),we){ie.bindFramebuffer(C.FRAMEBUFFER,we);try{const Ce=y.texture,ze=Ce.format,ke=Ce.type;if(!pe.textureFormatReadable(ze)){console.error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not in RGBA or implementation defined format.");return}if(!pe.textureTypeReadable(ke)){console.error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not in UnsignedByteType or implementation defined type.");return}I>=0&&I<=y.width-j&&W>=0&&W<=y.height-P&&C.readPixels(I,W,j,P,He.convert(ze),He.convert(ke),se)}finally{const Ce=O!==null?ge.get(O).__webglFramebuffer:null;ie.bindFramebuffer(C.FRAMEBUFFER,Ce)}}},this.readRenderTargetPixelsAsync=async function(y,I,W,j,P,se,Me){if(!(y&&y.isWebGLRenderTarget))throw new Error("THREE.WebGLRenderer.readRenderTargetPixels: renderTarget is not THREE.WebGLRenderTarget.");let we=ge.get(y).__webglFramebuffer;if(y.isWebGLCubeRenderTarget&&Me!==void 0&&(we=we[Me]),we){const Ce=y.texture,ze=Ce.format,ke=Ce.type;if(!pe.textureFormatReadable(ze))throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: renderTarget is not in RGBA or implementation defined format.");if(!pe.textureTypeReadable(ke))throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: renderTarget is not in UnsignedByteType or implementation defined type.");if(I>=0&&I<=y.width-j&&W>=0&&W<=y.height-P){ie.bindFramebuffer(C.FRAMEBUFFER,we);const Ie=C.createBuffer();C.bindBuffer(C.PIXEL_PACK_BUFFER,Ie),C.bufferData(C.PIXEL_PACK_BUFFER,se.byteLength,C.STREAM_READ),C.readPixels(I,W,j,P,He.convert(ze),He.convert(ke),0);const Ze=O!==null?ge.get(O).__webglFramebuffer:null;ie.bindFramebuffer(C.FRAMEBUFFER,Ze);const st=C.fenceSync(C.SYNC_GPU_COMMANDS_COMPLETE,0);return C.flush(),await Ac(C,st,4),C.bindBuffer(C.PIXEL_PACK_BUFFER,Ie),C.getBufferSubData(C.PIXEL_PACK_BUFFER,0,se),C.deleteBuffer(Ie),C.deleteSync(st),se}else throw new Error("THREE.WebGLRenderer.readRenderTargetPixelsAsync: requested read bounds are out of range.")}},this.copyFramebufferToTexture=function(y,I=null,W=0){y.isTexture!==!0&&(Fi("WebGLRenderer: copyFramebufferToTexture function signature has changed."),I=arguments[0]||null,y=arguments[1]);const j=Math.pow(2,-W),P=Math.floor(y.image.width*j),se=Math.floor(y.image.height*j),Me=I!==null?I.x:0,we=I!==null?I.y:0;E.setTexture2D(y,0),C.copyTexSubImage2D(C.TEXTURE_2D,W,0,0,Me,we,P,se),ie.unbindTexture()},this.copyTextureToTexture=function(y,I,W=null,j=null,P=0){y.isTexture!==!0&&(Fi("WebGLRenderer: copyTextureToTexture function signature has changed."),j=arguments[0]||null,y=arguments[1],I=arguments[2],P=arguments[3]||0,W=null);let se,Me,we,Ce,ze,ke,Ie,Ze,st;const ot=y.isCompressedTexture?y.mipmaps[P]:y.image;W!==null?(se=W.max.x-W.min.x,Me=W.max.y-W.min.y,we=W.isBox3?W.max.z-W.min.z:1,Ce=W.min.x,ze=W.min.y,ke=W.isBox3?W.min.z:0):(se=ot.width,Me=ot.height,we=ot.depth||1,Ce=0,ze=0,ke=0),j!==null?(Ie=j.x,Ze=j.y,st=j.z):(Ie=0,Ze=0,st=0);const ft=He.convert(I.format),lt=He.convert(I.type);let Le;I.isData3DTexture?(E.setTexture3D(I,0),Le=C.TEXTURE_3D):I.isDataArrayTexture||I.isCompressedArrayTexture?(E.setTexture2DArray(I,0),Le=C.TEXTURE_2D_ARRAY):(E.setTexture2D(I,0),Le=C.TEXTURE_2D),C.pixelStorei(C.UNPACK_FLIP_Y_WEBGL,I.flipY),C.pixelStorei(C.UNPACK_PREMULTIPLY_ALPHA_WEBGL,I.premultiplyAlpha),C.pixelStorei(C.UNPACK_ALIGNMENT,I.unpackAlignment);const Tr=C.getParameter(C.UNPACK_ROW_LENGTH),wr=C.getParameter(C.UNPACK_IMAGE_HEIGHT),Vt=C.getParameter(C.UNPACK_SKIP_PIXELS),Wr=C.getParameter(C.UNPACK_SKIP_ROWS),It=C.getParameter(C.UNPACK_SKIP_IMAGES);C.pixelStorei(C.UNPACK_ROW_LENGTH,ot.width),C.pixelStorei(C.UNPACK_IMAGE_HEIGHT,ot.height),C.pixelStorei(C.UNPACK_SKIP_PIXELS,Ce),C.pixelStorei(C.UNPACK_SKIP_ROWS,ze),C.pixelStorei(C.UNPACK_SKIP_IMAGES,ke);const bi=y.isDataArrayTexture||y.isData3DTexture,at=I.isDataArrayTexture||I.isData3DTexture;if(y.isRenderTargetTexture||y.isDepthTexture){const qt=ge.get(y),Ti=ge.get(I),Lt=ge.get(qt.__renderTarget),wi=ge.get(Ti.__renderTarget);ie.bindFramebuffer(C.READ_FRAMEBUFFER,Lt.__webglFramebuffer),ie.bindFramebuffer(C.DRAW_FRAMEBUFFER,wi.__webglFramebuffer);for(let hr=0;hr<we;hr++)bi&&C.framebufferTextureLayer(C.READ_FRAMEBUFFER,C.COLOR_ATTACHMENT0,ge.get(y).__webglTexture,P,ke+hr),y.isDepthTexture?(at&&C.framebufferTextureLayer(C.DRAW_FRAMEBUFFER,C.COLOR_ATTACHMENT0,ge.get(I).__webglTexture,P,st+hr),C.blitFramebuffer(Ce,ze,se,Me,Ie,Ze,se,Me,C.DEPTH_BUFFER_BIT,C.NEAREST)):at?C.copyTexSubImage3D(Le,P,Ie,Ze,st+hr,Ce,ze,se,Me):C.copyTexSubImage2D(Le,P,Ie,Ze,st+hr,Ce,ze,se,Me);ie.bindFramebuffer(C.READ_FRAMEBUFFER,null),ie.bindFramebuffer(C.DRAW_FRAMEBUFFER,null)}else at?y.isDataTexture||y.isData3DTexture?C.texSubImage3D(Le,P,Ie,Ze,st,se,Me,we,ft,lt,ot.data):I.isCompressedArrayTexture?C.compressedTexSubImage3D(Le,P,Ie,Ze,st,se,Me,we,ft,ot.data):C.texSubImage3D(Le,P,Ie,Ze,st,se,Me,we,ft,lt,ot):y.isDataTexture?C.texSubImage2D(C.TEXTURE_2D,P,Ie,Ze,se,Me,ft,lt,ot.data):y.isCompressedTexture?C.compressedTexSubImage2D(C.TEXTURE_2D,P,Ie,Ze,ot.width,ot.height,ft,ot.data):C.texSubImage2D(C.TEXTURE_2D,P,Ie,Ze,se,Me,ft,lt,ot);C.pixelStorei(C.UNPACK_ROW_LENGTH,Tr),C.pixelStorei(C.UNPACK_IMAGE_HEIGHT,wr),C.pixelStorei(C.UNPACK_SKIP_PIXELS,Vt),C.pixelStorei(C.UNPACK_SKIP_ROWS,Wr),C.pixelStorei(C.UNPACK_SKIP_IMAGES,It),P===0&&I.generateMipmaps&&C.generateMipmap(Le),ie.unbindTexture()},this.copyTextureToTexture3D=function(y,I,W=null,j=null,P=0){return y.isTexture!==!0&&(Fi("WebGLRenderer: copyTextureToTexture3D function signature has changed."),W=arguments[0]||null,j=arguments[1]||null,y=arguments[2],I=arguments[3],P=arguments[4]||0),Fi('WebGLRenderer: copyTextureToTexture3D function has been deprecated. Use "copyTextureToTexture" instead.'),this.copyTextureToTexture(y,I,W,j,P)},this.initRenderTarget=function(y){ge.get(y).__webglFramebuffer===void 0&&E.setupRenderTarget(y)},this.initTexture=function(y){y.isCubeTexture?E.setTextureCube(y,0):y.isData3DTexture?E.setTexture3D(y,0):y.isDataArrayTexture||y.isCompressedArrayTexture?E.setTexture2DArray(y,0):E.setTexture2D(y,0),ie.unbindTexture()},this.resetState=function(){w=0,A=0,O=null,ie.reset(),nt.reset()},typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}get coordinateSystem(){return or}get outputColorSpace(){return this._outputColorSpace}set outputColorSpace(e){this._outputColorSpace=e;const t=this.getContext();t.drawingBufferColorspace=$e._getDrawingBufferColorSpace(e),t.unpackColorSpace=$e._getUnpackColorSpace()}}class kf extends Mt{constructor(){super(),this.isScene=!0,this.type="Scene",this.background=null,this.environment=null,this.fog=null,this.backgroundBlurriness=0,this.backgroundIntensity=1,this.backgroundRotation=new Yt,this.environmentIntensity=1,this.environmentRotation=new Yt,this.overrideMaterial=null,typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("observe",{detail:this}))}copy(e,t){return super.copy(e,t),e.background!==null&&(this.background=e.background.clone()),e.environment!==null&&(this.environment=e.environment.clone()),e.fog!==null&&(this.fog=e.fog.clone()),this.backgroundBlurriness=e.backgroundBlurriness,this.backgroundIntensity=e.backgroundIntensity,this.backgroundRotation.copy(e.backgroundRotation),this.environmentIntensity=e.environmentIntensity,this.environmentRotation.copy(e.environmentRotation),e.overrideMaterial!==null&&(this.overrideMaterial=e.overrideMaterial.clone()),this.matrixAutoUpdate=e.matrixAutoUpdate,this}toJSON(e){const t=super.toJSON(e);return this.fog!==null&&(t.object.fog=this.fog.toJSON()),this.backgroundBlurriness>0&&(t.object.backgroundBlurriness=this.backgroundBlurriness),this.backgroundIntensity!==1&&(t.object.backgroundIntensity=this.backgroundIntensity),t.object.backgroundRotation=this.backgroundRotation.toArray(),this.environmentIntensity!==1&&(t.object.environmentIntensity=this.environmentIntensity),t.object.environmentRotation=this.environmentRotation.toArray(),t}}class Vf{constructor(e,t){this.isInterleavedBuffer=!0,this.array=e,this.stride=t,this.count=e!==void 0?e.length/t:0,this.usage=as,this.updateRanges=[],this.version=0,this.uuid=Mr()}onUploadCallback(){}set needsUpdate(e){e===!0&&this.version++}setUsage(e){return this.usage=e,this}addUpdateRange(e,t){this.updateRanges.push({start:e,count:t})}clearUpdateRanges(){this.updateRanges.length=0}copy(e){return this.array=new e.array.constructor(e.array),this.count=e.count,this.stride=e.stride,this.usage=e.usage,this}copyAt(e,t,r){e*=this.stride,r*=t.stride;for(let n=0,a=this.stride;n<a;n++)this.array[e+n]=t.array[r+n];return this}set(e,t=0){return this.array.set(e,t),this}clone(e){e.arrayBuffers===void 0&&(e.arrayBuffers={}),this.array.buffer._uuid===void 0&&(this.array.buffer._uuid=Mr()),e.arrayBuffers[this.array.buffer._uuid]===void 0&&(e.arrayBuffers[this.array.buffer._uuid]=this.array.slice(0).buffer);const t=new this.array.constructor(e.arrayBuffers[this.array.buffer._uuid]),r=new this.constructor(t,this.stride);return r.setUsage(this.usage),r}onUpload(e){return this.onUploadCallback=e,this}toJSON(e){return e.arrayBuffers===void 0&&(e.arrayBuffers={}),this.array.buffer._uuid===void 0&&(this.array.buffer._uuid=Mr()),e.arrayBuffers[this.array.buffer._uuid]===void 0&&(e.arrayBuffers[this.array.buffer._uuid]=Array.from(new Uint32Array(this.array.buffer))),{uuid:this.uuid,buffer:this.array.buffer._uuid,type:this.array.constructor.name,stride:this.stride}}}const wt=new B;class Dn{constructor(e,t,r,n=!1){this.isInterleavedBufferAttribute=!0,this.name="",this.data=e,this.itemSize=t,this.offset=r,this.normalized=n}get count(){return this.data.count}get array(){return this.data.array}set needsUpdate(e){this.data.needsUpdate=e}applyMatrix4(e){for(let t=0,r=this.data.count;t<r;t++)wt.fromBufferAttribute(this,t),wt.applyMatrix4(e),this.setXYZ(t,wt.x,wt.y,wt.z);return this}applyNormalMatrix(e){for(let t=0,r=this.count;t<r;t++)wt.fromBufferAttribute(this,t),wt.applyNormalMatrix(e),this.setXYZ(t,wt.x,wt.y,wt.z);return this}transformDirection(e){for(let t=0,r=this.count;t<r;t++)wt.fromBufferAttribute(this,t),wt.transformDirection(e),this.setXYZ(t,wt.x,wt.y,wt.z);return this}getComponent(e,t){let r=this.array[e*this.data.stride+this.offset+t];return this.normalized&&(r=Jt(r,this.array)),r}setComponent(e,t,r){return this.normalized&&(r=it(r,this.array)),this.data.array[e*this.data.stride+this.offset+t]=r,this}setX(e,t){return this.normalized&&(t=it(t,this.array)),this.data.array[e*this.data.stride+this.offset]=t,this}setY(e,t){return this.normalized&&(t=it(t,this.array)),this.data.array[e*this.data.stride+this.offset+1]=t,this}setZ(e,t){return this.normalized&&(t=it(t,this.array)),this.data.array[e*this.data.stride+this.offset+2]=t,this}setW(e,t){return this.normalized&&(t=it(t,this.array)),this.data.array[e*this.data.stride+this.offset+3]=t,this}getX(e){let t=this.data.array[e*this.data.stride+this.offset];return this.normalized&&(t=Jt(t,this.array)),t}getY(e){let t=this.data.array[e*this.data.stride+this.offset+1];return this.normalized&&(t=Jt(t,this.array)),t}getZ(e){let t=this.data.array[e*this.data.stride+this.offset+2];return this.normalized&&(t=Jt(t,this.array)),t}getW(e){let t=this.data.array[e*this.data.stride+this.offset+3];return this.normalized&&(t=Jt(t,this.array)),t}setXY(e,t,r){return e=e*this.data.stride+this.offset,this.normalized&&(t=it(t,this.array),r=it(r,this.array)),this.data.array[e+0]=t,this.data.array[e+1]=r,this}setXYZ(e,t,r,n){return e=e*this.data.stride+this.offset,this.normalized&&(t=it(t,this.array),r=it(r,this.array),n=it(n,this.array)),this.data.array[e+0]=t,this.data.array[e+1]=r,this.data.array[e+2]=n,this}setXYZW(e,t,r,n,a){return e=e*this.data.stride+this.offset,this.normalized&&(t=it(t,this.array),r=it(r,this.array),n=it(n,this.array),a=it(a,this.array)),this.data.array[e+0]=t,this.data.array[e+1]=r,this.data.array[e+2]=n,this.data.array[e+3]=a,this}clone(e){if(e===void 0){console.log("THREE.InterleavedBufferAttribute.clone(): Cloning an interleaved buffer attribute will de-interleave buffer data.");const t=[];for(let r=0;r<this.count;r++){const n=r*this.data.stride+this.offset;for(let a=0;a<this.itemSize;a++)t.push(this.data.array[n+a])}return new Bt(new this.array.constructor(t),this.itemSize,this.normalized)}else return e.interleavedBuffers===void 0&&(e.interleavedBuffers={}),e.interleavedBuffers[this.data.uuid]===void 0&&(e.interleavedBuffers[this.data.uuid]=this.data.clone(e)),new Dn(e.interleavedBuffers[this.data.uuid],this.itemSize,this.offset,this.normalized)}toJSON(e){if(e===void 0){console.log("THREE.InterleavedBufferAttribute.toJSON(): Serializing an interleaved buffer attribute will de-interleave buffer data.");const t=[];for(let r=0;r<this.count;r++){const n=r*this.data.stride+this.offset;for(let a=0;a<this.itemSize;a++)t.push(this.data.array[n+a])}return{itemSize:this.itemSize,type:this.array.constructor.name,array:t,normalized:this.normalized}}else return e.interleavedBuffers===void 0&&(e.interleavedBuffers={}),e.interleavedBuffers[this.data.uuid]===void 0&&(e.interleavedBuffers[this.data.uuid]=this.data.toJSON(e)),{isInterleavedBufferAttribute:!0,itemSize:this.itemSize,data:this.data.uuid,offset:this.offset,normalized:this.normalized}}}class ls extends Gr{static get type(){return"SpriteMaterial"}constructor(e){super(),this.isSpriteMaterial=!0,this.color=new We(16777215),this.map=null,this.alphaMap=null,this.rotation=0,this.sizeAttenuation=!0,this.transparent=!0,this.fog=!0,this.setValues(e)}copy(e){return super.copy(e),this.color.copy(e.color),this.map=e.map,this.alphaMap=e.alphaMap,this.rotation=e.rotation,this.sizeAttenuation=e.sizeAttenuation,this.fog=e.fog,this}}let ai;const Pi=new B,si=new B,oi=new B,li=new je,Di=new je,Tl=new Qe,_n=new B,Ii=new B,vn=new B,Io=new je,ga=new je,No=new je;class Uo extends Mt{constructor(e=new ls){if(super(),this.isSprite=!0,this.type="Sprite",ai===void 0){ai=new Ct;const t=new Float32Array([-.5,-.5,0,0,0,.5,-.5,0,1,0,.5,.5,0,1,1,-.5,.5,0,0,1]),r=new Vf(t,5);ai.setIndex([0,1,2,0,2,3]),ai.setAttribute("position",new Dn(r,3,0,!1)),ai.setAttribute("uv",new Dn(r,2,3,!1))}this.geometry=ai,this.material=e,this.center=new je(.5,.5)}raycast(e,t){e.camera===null&&console.error('THREE.Sprite: "Raycaster.camera" needs to be set in order to raycast against sprites.'),si.setFromMatrixScale(this.matrixWorld),Tl.copy(e.camera.matrixWorld),this.modelViewMatrix.multiplyMatrices(e.camera.matrixWorldInverse,this.matrixWorld),oi.setFromMatrixPosition(this.modelViewMatrix),e.camera.isPerspectiveCamera&&this.material.sizeAttenuation===!1&&si.multiplyScalar(-oi.z);const r=this.material.rotation;let n,a;r!==0&&(a=Math.cos(r),n=Math.sin(r));const o=this.center;xn(_n.set(-.5,-.5,0),oi,o,si,n,a),xn(Ii.set(.5,-.5,0),oi,o,si,n,a),xn(vn.set(.5,.5,0),oi,o,si,n,a),Io.set(0,0),ga.set(1,0),No.set(1,1);let s=e.ray.intersectTriangle(_n,Ii,vn,!1,Pi);if(s===null&&(xn(Ii.set(-.5,.5,0),oi,o,si,n,a),ga.set(0,1),s=e.ray.intersectTriangle(_n,vn,Ii,!1,Pi),s===null))return;const l=e.ray.origin.distanceTo(Pi);l<e.near||l>e.far||t.push({distance:l,point:Pi.clone(),uv:zt.getInterpolation(Pi,_n,Ii,vn,Io,ga,No,new je),face:null,object:this})}copy(e,t){return super.copy(e,t),e.center!==void 0&&this.center.copy(e.center),this.material=e.material,this}}function xn(i,e,t,r,n,a){li.subVectors(i,t).addScalar(.5).multiply(r),n!==void 0?(Di.x=a*li.x-n*li.y,Di.y=n*li.x+a*li.y):Di.copy(li),i.copy(e),i.x+=Di.x,i.y+=Di.y,i.applyMatrix4(Tl)}class Hf extends _t{constructor(e=null,t=1,r=1,n,a,o,s,l,c=pt,h=pt,u,p){super(null,o,s,l,c,h,n,a,u,p),this.isDataTexture=!0,this.image={data:e,width:t,height:r},this.generateMipmaps=!1,this.flipY=!1,this.unpackAlignment=1}}class Oo extends Bt{constructor(e,t,r,n=1){super(e,t,r),this.isInstancedBufferAttribute=!0,this.meshPerAttribute=n}copy(e){return super.copy(e),this.meshPerAttribute=e.meshPerAttribute,this}toJSON(){const e=super.toJSON();return e.meshPerAttribute=this.meshPerAttribute,e.isInstancedBufferAttribute=!0,e}}const ci=new Qe,Fo=new Qe,yn=[],zo=new Hr,Gf=new Qe,Ni=new Fe,Ui=new Si;class Wf extends Fe{constructor(e,t,r){super(e,t),this.isInstancedMesh=!0,this.instanceMatrix=new Oo(new Float32Array(r*16),16),this.instanceColor=null,this.morphTexture=null,this.count=r,this.boundingBox=null,this.boundingSphere=null;for(let n=0;n<r;n++)this.setMatrixAt(n,Gf)}computeBoundingBox(){const e=this.geometry,t=this.count;this.boundingBox===null&&(this.boundingBox=new Hr),e.boundingBox===null&&e.computeBoundingBox(),this.boundingBox.makeEmpty();for(let r=0;r<t;r++)this.getMatrixAt(r,ci),zo.copy(e.boundingBox).applyMatrix4(ci),this.boundingBox.union(zo)}computeBoundingSphere(){const e=this.geometry,t=this.count;this.boundingSphere===null&&(this.boundingSphere=new Si),e.boundingSphere===null&&e.computeBoundingSphere(),this.boundingSphere.makeEmpty();for(let r=0;r<t;r++)this.getMatrixAt(r,ci),Ui.copy(e.boundingSphere).applyMatrix4(ci),this.boundingSphere.union(Ui)}copy(e,t){return super.copy(e,t),this.instanceMatrix.copy(e.instanceMatrix),e.morphTexture!==null&&(this.morphTexture=e.morphTexture.clone()),e.instanceColor!==null&&(this.instanceColor=e.instanceColor.clone()),this.count=e.count,e.boundingBox!==null&&(this.boundingBox=e.boundingBox.clone()),e.boundingSphere!==null&&(this.boundingSphere=e.boundingSphere.clone()),this}getColorAt(e,t){t.fromArray(this.instanceColor.array,e*3)}getMatrixAt(e,t){t.fromArray(this.instanceMatrix.array,e*16)}getMorphAt(e,t){const r=t.morphTargetInfluences,n=this.morphTexture.source.data.data,a=r.length+1,o=e*a+1;for(let s=0;s<r.length;s++)r[s]=n[o+s]}raycast(e,t){const r=this.matrixWorld,n=this.count;if(Ni.geometry=this.geometry,Ni.material=this.material,Ni.material!==void 0&&(this.boundingSphere===null&&this.computeBoundingSphere(),Ui.copy(this.boundingSphere),Ui.applyMatrix4(r),e.ray.intersectsSphere(Ui)!==!1))for(let a=0;a<n;a++){this.getMatrixAt(a,ci),Fo.multiplyMatrices(r,ci),Ni.matrixWorld=Fo,Ni.raycast(e,yn);for(let o=0,s=yn.length;o<s;o++){const l=yn[o];l.instanceId=a,l.object=this,t.push(l)}yn.length=0}}setColorAt(e,t){this.instanceColor===null&&(this.instanceColor=new Oo(new Float32Array(this.instanceMatrix.count*3).fill(1),3)),t.toArray(this.instanceColor.array,e*3)}setMatrixAt(e,t){t.toArray(this.instanceMatrix.array,e*16)}setMorphAt(e,t){const r=t.morphTargetInfluences,n=r.length+1;this.morphTexture===null&&(this.morphTexture=new Hf(new Float32Array(n*this.count),n,this.count,gs,Qt));const a=this.morphTexture.source.data.data;let o=0;for(let c=0;c<r.length;c++)o+=r[c];const s=this.geometry.morphTargetsRelative?1:1-o,l=n*e;a[l]=s,a.set(r,l+1)}updateMorphTargets(){}dispose(){return this.dispatchEvent({type:"dispose"}),this.morphTexture!==null&&(this.morphTexture.dispose(),this.morphTexture=null),this}}class cs extends Gr{static get type(){return"PointsMaterial"}constructor(e){super(),this.isPointsMaterial=!0,this.color=new We(16777215),this.map=null,this.alphaMap=null,this.size=1,this.sizeAttenuation=!0,this.fog=!0,this.setValues(e)}copy(e){return super.copy(e),this.color.copy(e.color),this.map=e.map,this.alphaMap=e.alphaMap,this.size=e.size,this.sizeAttenuation=e.sizeAttenuation,this.fog=e.fog,this}}const Bo=new Qe,hs=new hl,Mn=new Si,Sn=new B;class ko extends Mt{constructor(e=new Ct,t=new cs){super(),this.isPoints=!0,this.type="Points",this.geometry=e,this.material=t,this.updateMorphTargets()}copy(e,t){return super.copy(e,t),this.material=Array.isArray(e.material)?e.material.slice():e.material,this.geometry=e.geometry,this}raycast(e,t){const r=this.geometry,n=this.matrixWorld,a=e.params.Points.threshold,o=r.drawRange;if(r.boundingSphere===null&&r.computeBoundingSphere(),Mn.copy(r.boundingSphere),Mn.applyMatrix4(n),Mn.radius+=a,e.ray.intersectsSphere(Mn)===!1)return;Bo.copy(n).invert(),hs.copy(e.ray).applyMatrix4(Bo);const s=a/((this.scale.x+this.scale.y+this.scale.z)/3),l=s*s,c=r.index,h=r.attributes.position;if(c!==null){const u=Math.max(0,o.start),p=Math.min(c.count,o.start+o.count);for(let f=u,g=p;f<g;f++){const _=c.getX(f);Sn.fromBufferAttribute(h,_),Vo(Sn,_,l,n,e,t,this)}}else{const u=Math.max(0,o.start),p=Math.min(h.count,o.start+o.count);for(let f=u,g=p;f<g;f++)Sn.fromBufferAttribute(h,f),Vo(Sn,f,l,n,e,t,this)}}updateMorphTargets(){const e=this.geometry.morphAttributes,t=Object.keys(e);if(t.length>0){const r=e[t[0]];if(r!==void 0){this.morphTargetInfluences=[],this.morphTargetDictionary={};for(let n=0,a=r.length;n<a;n++){const o=r[n].name||String(n);this.morphTargetInfluences.push(0),this.morphTargetDictionary[o]=n}}}}}function Vo(i,e,t,r,n,a,o){const s=hs.distanceSqToPoint(i);if(s<t){const l=new B;hs.closestPointToPoint(i,l),l.applyMatrix4(r);const c=n.ray.origin.distanceTo(l);if(c<n.near||c>n.far)return;a.push({distance:c,distanceToRay:Math.sqrt(s),point:l,index:e,face:null,faceIndex:null,barycoord:null,object:o})}}class Ho extends _t{constructor(e,t,r,n,a,o,s,l,c){super(e,t,r,n,a,o,s,l,c),this.isCanvasTexture=!0,this.needsUpdate=!0}}class In extends Ct{constructor(e=.5,t=1,r=32,n=1,a=0,o=Math.PI*2){super(),this.type="RingGeometry",this.parameters={innerRadius:e,outerRadius:t,thetaSegments:r,phiSegments:n,thetaStart:a,thetaLength:o},r=Math.max(3,r),n=Math.max(1,n);const s=[],l=[],c=[],h=[];let u=e;const p=(t-e)/n,f=new B,g=new je;for(let _=0;_<=n;_++){for(let m=0;m<=r;m++){const d=a+m/r*o;f.x=u*Math.cos(d),f.y=u*Math.sin(d),l.push(f.x,f.y,f.z),c.push(0,0,1),g.x=(f.x/t+1)/2,g.y=(f.y/t+1)/2,h.push(g.x,g.y)}u+=p}for(let _=0;_<n;_++){const m=_*(r+1);for(let d=0;d<r;d++){const S=d+m,T=S,x=S+r+1,N=S+r+2,w=S+1;s.push(T,x,w),s.push(x,N,w)}}this.setIndex(s),this.setAttribute("position",new ht(l,3)),this.setAttribute("normal",new ht(c,3)),this.setAttribute("uv",new ht(h,2))}copy(e){return super.copy(e),this.parameters=Object.assign({},e.parameters),this}static fromJSON(e){return new In(e.innerRadius,e.outerRadius,e.thetaSegments,e.phiSegments,e.thetaStart,e.thetaLength)}}class Nn extends Ct{constructor(e=1,t=32,r=16,n=0,a=Math.PI*2,o=0,s=Math.PI){super(),this.type="SphereGeometry",this.parameters={radius:e,widthSegments:t,heightSegments:r,phiStart:n,phiLength:a,thetaStart:o,thetaLength:s},t=Math.max(3,Math.floor(t)),r=Math.max(2,Math.floor(r));const l=Math.min(o+s,Math.PI);let c=0;const h=[],u=new B,p=new B,f=[],g=[],_=[],m=[];for(let d=0;d<=r;d++){const S=[],T=d/r;let x=0;d===0&&o===0?x=.5/t:d===r&&l===Math.PI&&(x=-.5/t);for(let N=0;N<=t;N++){const w=N/t;u.x=-e*Math.cos(n+w*a)*Math.sin(o+T*s),u.y=e*Math.cos(o+T*s),u.z=e*Math.sin(n+w*a)*Math.sin(o+T*s),g.push(u.x,u.y,u.z),p.copy(u).normalize(),_.push(p.x,p.y,p.z),m.push(w+x,1-T),S.push(c++)}h.push(S)}for(let d=0;d<r;d++)for(let S=0;S<t;S++){const T=h[d][S+1],x=h[d][S],N=h[d+1][S],w=h[d+1][S+1];(d!==0||o>0)&&f.push(T,x,w),(d!==r-1||l<Math.PI)&&f.push(x,N,w)}this.setIndex(f),this.setAttribute("position",new ht(g,3)),this.setAttribute("normal",new ht(_,3)),this.setAttribute("uv",new ht(m,2))}copy(e){return super.copy(e),this.parameters=Object.assign({},e.parameters),this}static fromJSON(e){return new Nn(e.radius,e.widthSegments,e.heightSegments,e.phiStart,e.phiLength,e.thetaStart,e.thetaLength)}}class ct extends Gr{static get type(){return"MeshLambertMaterial"}constructor(e){super(),this.isMeshLambertMaterial=!0,this.color=new We(16777215),this.map=null,this.lightMap=null,this.lightMapIntensity=1,this.aoMap=null,this.aoMapIntensity=1,this.emissive=new We(0),this.emissiveIntensity=1,this.emissiveMap=null,this.bumpMap=null,this.bumpScale=1,this.normalMap=null,this.normalMapType=al,this.normalScale=new je(1,1),this.displacementMap=null,this.displacementScale=1,this.displacementBias=0,this.specularMap=null,this.alphaMap=null,this.envMap=null,this.envMapRotation=new Yt,this.combine=ds,this.reflectivity=1,this.refractionRatio=.98,this.wireframe=!1,this.wireframeLinewidth=1,this.wireframeLinecap="round",this.wireframeLinejoin="round",this.flatShading=!1,this.fog=!0,this.setValues(e)}copy(e){return super.copy(e),this.color.copy(e.color),this.map=e.map,this.lightMap=e.lightMap,this.lightMapIntensity=e.lightMapIntensity,this.aoMap=e.aoMap,this.aoMapIntensity=e.aoMapIntensity,this.emissive.copy(e.emissive),this.emissiveMap=e.emissiveMap,this.emissiveIntensity=e.emissiveIntensity,this.bumpMap=e.bumpMap,this.bumpScale=e.bumpScale,this.normalMap=e.normalMap,this.normalMapType=e.normalMapType,this.normalScale.copy(e.normalScale),this.displacementMap=e.displacementMap,this.displacementScale=e.displacementScale,this.displacementBias=e.displacementBias,this.specularMap=e.specularMap,this.alphaMap=e.alphaMap,this.envMap=e.envMap,this.envMapRotation.copy(e.envMapRotation),this.combine=e.combine,this.reflectivity=e.reflectivity,this.refractionRatio=e.refractionRatio,this.wireframe=e.wireframe,this.wireframeLinewidth=e.wireframeLinewidth,this.wireframeLinecap=e.wireframeLinecap,this.wireframeLinejoin=e.wireframeLinejoin,this.flatShading=e.flatShading,this.fog=e.fog,this}}const Go={enabled:!1,files:{},add:function(i,e){this.enabled!==!1&&(this.files[i]=e)},get:function(i){if(this.enabled!==!1)return this.files[i]},remove:function(i){delete this.files[i]},clear:function(){this.files={}}};class jf{constructor(e,t,r){const n=this;let a=!1,o=0,s=0,l;const c=[];this.onStart=void 0,this.onLoad=e,this.onProgress=t,this.onError=r,this.itemStart=function(h){s++,a===!1&&n.onStart!==void 0&&n.onStart(h,o,s),a=!0},this.itemEnd=function(h){o++,n.onProgress!==void 0&&n.onProgress(h,o,s),o===s&&(a=!1,n.onLoad!==void 0&&n.onLoad())},this.itemError=function(h){n.onError!==void 0&&n.onError(h)},this.resolveURL=function(h){return l?l(h):h},this.setURLModifier=function(h){return l=h,this},this.addHandler=function(h,u){return c.push(h,u),this},this.removeHandler=function(h){const u=c.indexOf(h);return u!==-1&&c.splice(u,2),this},this.getHandler=function(h){for(let u=0,p=c.length;u<p;u+=2){const f=c[u],g=c[u+1];if(f.global&&(f.lastIndex=0),f.test(h))return g}return null}}}const Xf=new jf;class Ss{constructor(e){this.manager=e!==void 0?e:Xf,this.crossOrigin="anonymous",this.withCredentials=!1,this.path="",this.resourcePath="",this.requestHeader={}}load(){}loadAsync(e,t){const r=this;return new Promise(function(n,a){r.load(e,n,t,a)})}parse(){}setCrossOrigin(e){return this.crossOrigin=e,this}setWithCredentials(e){return this.withCredentials=e,this}setPath(e){return this.path=e,this}setResourcePath(e){return this.resourcePath=e,this}setRequestHeader(e){return this.requestHeader=e,this}}Ss.DEFAULT_MATERIAL_NAME="__DEFAULT";class Yf extends Ss{constructor(e){super(e)}load(e,t,r,n){this.path!==void 0&&(e=this.path+e),e=this.manager.resolveURL(e);const a=this,o=Go.get(e);if(o!==void 0)return a.manager.itemStart(e),setTimeout(function(){t&&t(o),a.manager.itemEnd(e)},0),o;const s=Vi("img");function l(){h(),Go.add(e,this),t&&t(this),a.manager.itemEnd(e)}function c(u){h(),n&&n(u),a.manager.itemError(e),a.manager.itemEnd(e)}function h(){s.removeEventListener("load",l,!1),s.removeEventListener("error",c,!1)}return s.addEventListener("load",l,!1),s.addEventListener("error",c,!1),e.slice(0,5)!=="data:"&&this.crossOrigin!==void 0&&(s.crossOrigin=this.crossOrigin),a.manager.itemStart(e),s.src=e,s}}class qf extends Ss{constructor(e){super(e)}load(e,t,r,n){const a=new _t,o=new Yf(this.manager);return o.setCrossOrigin(this.crossOrigin),o.setPath(this.path),o.load(e,function(s){a.image=s,a.needsUpdate=!0,t!==void 0&&t(a)},r,n),a}}class wl extends Mt{constructor(e,t=1){super(),this.isLight=!0,this.type="Light",this.color=new We(e),this.intensity=t}dispose(){}copy(e,t){return super.copy(e,t),this.color.copy(e.color),this.intensity=e.intensity,this}toJSON(e){const t=super.toJSON(e);return t.object.color=this.color.getHex(),t.object.intensity=this.intensity,this.groundColor!==void 0&&(t.object.groundColor=this.groundColor.getHex()),this.distance!==void 0&&(t.object.distance=this.distance),this.angle!==void 0&&(t.object.angle=this.angle),this.decay!==void 0&&(t.object.decay=this.decay),this.penumbra!==void 0&&(t.object.penumbra=this.penumbra),this.shadow!==void 0&&(t.object.shadow=this.shadow.toJSON()),this.target!==void 0&&(t.object.target=this.target.uuid),t}}const _a=new Qe,Wo=new B,jo=new B;class Kf{constructor(e){this.camera=e,this.intensity=1,this.bias=0,this.normalBias=0,this.radius=1,this.blurSamples=8,this.mapSize=new je(512,512),this.map=null,this.mapPass=null,this.matrix=new Qe,this.autoUpdate=!0,this.needsUpdate=!1,this._frustum=new ys,this._frameExtents=new je(1,1),this._viewportCount=1,this._viewports=[new dt(0,0,1,1)]}getViewportCount(){return this._viewportCount}getFrustum(){return this._frustum}updateMatrices(e){const t=this.camera,r=this.matrix;Wo.setFromMatrixPosition(e.matrixWorld),t.position.copy(Wo),jo.setFromMatrixPosition(e.target.matrixWorld),t.lookAt(jo),t.updateMatrixWorld(),_a.multiplyMatrices(t.projectionMatrix,t.matrixWorldInverse),this._frustum.setFromProjectionMatrix(_a),r.set(.5,0,0,.5,0,.5,0,.5,0,0,.5,.5,0,0,0,1),r.multiply(_a)}getViewport(e){return this._viewports[e]}getFrameExtents(){return this._frameExtents}dispose(){this.map&&this.map.dispose(),this.mapPass&&this.mapPass.dispose()}copy(e){return this.camera=e.camera.clone(),this.intensity=e.intensity,this.bias=e.bias,this.radius=e.radius,this.mapSize.copy(e.mapSize),this}clone(){return new this.constructor().copy(this)}toJSON(){const e={};return this.intensity!==1&&(e.intensity=this.intensity),this.bias!==0&&(e.bias=this.bias),this.normalBias!==0&&(e.normalBias=this.normalBias),this.radius!==1&&(e.radius=this.radius),(this.mapSize.x!==512||this.mapSize.y!==512)&&(e.mapSize=this.mapSize.toArray()),e.camera=this.camera.toJSON(!1).object,delete e.camera.matrix,e}}class $f extends Kf{constructor(){super(new xl(-5,5,5,-5,.5,500)),this.isDirectionalLightShadow=!0}}class Zf extends wl{constructor(e,t){super(e,t),this.isDirectionalLight=!0,this.type="DirectionalLight",this.position.copy(Mt.DEFAULT_UP),this.updateMatrix(),this.target=new Mt,this.shadow=new $f}dispose(){this.shadow.dispose()}copy(e){return super.copy(e),this.target=e.target.clone(),this.shadow=e.shadow.clone(),this}}class Jf extends wl{constructor(e,t){super(e,t),this.isAmbientLight=!0,this.type="AmbientLight"}}typeof __THREE_DEVTOOLS__<"u"&&__THREE_DEVTOOLS__.dispatchEvent(new CustomEvent("register",{detail:{revision:us}}));typeof window<"u"&&(window.__THREE__?console.warn("WARNING: Multiple instances of Three.js being imported."):window.__THREE__=us);const Qf={up:{dir:[0,1,0],vertices:[[0,1,0],[1,1,0],[1,1,1],[0,1,1]],normal:[0,1,0]},down:{dir:[0,-1,0],vertices:[[0,0,1],[1,0,1],[1,0,0],[0,0,0]],normal:[0,-1,0]},east:{dir:[1,0,0],vertices:[[1,0,1],[1,1,1],[1,1,0],[1,0,0]],normal:[1,0,0]},west:{dir:[-1,0,0],vertices:[[0,0,0],[0,1,0],[0,1,1],[0,0,1]],normal:[-1,0,0]},south:{dir:[0,0,1],vertices:[[0,0,1],[0,1,1],[1,1,1],[1,0,1]],normal:[0,0,1]},north:{dir:[0,0,-1],vertices:[[1,0,0],[1,1,0],[0,1,0],[0,0,0]],normal:[0,0,-1]}},En=[[0,1],[0,0],[1,0],[1,1]],bn=[0,2,1,0,3,2],em=[{vertices:[[.1464,0,.8536],[.1464,1,.8536],[.8536,1,.1464],[.8536,0,.1464]],normal:[-.7071,0,.7071]},{vertices:[[.8536,0,.1464],[.8536,1,.1464],[.1464,1,.8536],[.1464,0,.8536]],normal:[.7071,0,-.7071]},{vertices:[[.1464,0,.1464],[.1464,1,.1464],[.8536,1,.8536],[.8536,0,.8536]],normal:[.7071,0,.7071]},{vertices:[[.8536,0,.8536],[.8536,1,.8536],[.1464,1,.1464],[.1464,0,.1464]],normal:[-.7071,0,-.7071]}],tm=[{vertices:[[0,.01,0],[0,.01,1],[1,.01,1],[1,.01,0]],normal:[0,1,0]},{vertices:[[1,.01,0],[1,.01,1],[0,.01,1],[0,.01,0]],normal:[0,-1,0]}],rm=[{mask:1,quads:[{vertices:[[0,.99,0],[1,.99,0],[1,.99,1],[0,.99,1]],normal:[0,1,0]},{vertices:[[0,.99,1],[1,.99,1],[1,.99,0],[0,.99,0]],normal:[0,-1,0]}]},{mask:2,quads:[{vertices:[[0,.01,1],[1,.01,1],[1,.01,0],[0,.01,0]],normal:[0,-1,0]},{vertices:[[0,.01,0],[1,.01,0],[1,.01,1],[0,.01,1]],normal:[0,1,0]}]},{mask:4,quads:[{vertices:[[1,0,.01],[1,1,.01],[0,1,.01],[0,0,.01]],normal:[0,0,-1]},{vertices:[[0,0,.01],[0,1,.01],[1,1,.01],[1,0,.01]],normal:[0,0,1]}]},{mask:8,quads:[{vertices:[[0,0,.99],[0,1,.99],[1,1,.99],[1,0,.99]],normal:[0,0,1]},{vertices:[[1,0,.99],[1,1,.99],[0,1,.99],[0,0,.99]],normal:[0,0,-1]}]},{mask:16,quads:[{vertices:[[.99,0,1],[.99,1,1],[.99,1,0],[.99,0,0]],normal:[1,0,0]},{vertices:[[.99,0,0],[.99,1,0],[.99,1,1],[.99,0,1]],normal:[-1,0,0]}]},{mask:32,quads:[{vertices:[[.01,0,0],[.01,1,0],[.01,1,1],[.01,0,1]],normal:[-1,0,0]},{vertices:[[.01,0,1],[.01,1,1],[.01,1,0],[.01,0,0]],normal:[1,0,0]}]}],$t=class Ke{constructor(e){this.container=null,this.renderer=null,this.animationFrameId=null,this.initialized=!1,this.atlasTexture=null,this.atlasMapping=null,this.atlasMaterial=null,this.transparentAtlasMaterial=null,this.transparentStates=new Set,this.stateToBlockId=new Map,this.waterLevelMap=new Map,this.blockNameToState=new Map,this.itemMapping=new Map,this.hiddenStates=new Set,this.sections=new Map,this.blockMap=new Map,this.dirtySections=new Set,this.skinTextures=new Map,this.embeddedSkins=new Map,this.armorTextureCache=new Map,this.entityTextureCache=new Map,this.playerPrevPositions=new Map,this.playerBodyYaw=new Map,this._lastDebugLog=0,this.playerHeldItems=new Map,this.entityMeshes=new Map,this.entityTypeNames=new Map,this.activeExplosions=[],this.playerMeshes=new Map,this.playerArmorState=new Map,this.playerNametagHealth=new Map,this.instancedMesh=null,this.blockIndexMap=new Map,this.instanceCount=0,this.maxInstances=0,this.hiddenScale=new Qe().makeScale(0,0,0),this.mcVersion=e,this.scene=new kf,this.camera=new Ft(75,1,.1,1e4),this.camera.position.set(0,80,0)}async setTextureAtlas(e,t){this.atlasMapping=t,t.transparentStates&&(this.transparentStates=new Set(t.transparentStates)),t.stateToBlock&&(this.stateToBlockId=new Map(Object.entries(t.stateToBlock).map(([l,c])=>[Number(l),c])));const r=new qf;this.atlasTexture=await new Promise((l,c)=>{r.load(e,l,void 0,c)}),this.atlasTexture.magFilter=pt,this.atlasTexture.minFilter=pt,this.atlasTexture.colorSpace=Tt,this.atlasTexture.flipY=!1,this.atlasMaterial=new ct({map:this.atlasTexture,transparent:!1,alphaTest:.1}),this.transparentAtlasMaterial=new ct({map:this.atlasTexture,transparent:!0,depthWrite:!1,alphaTest:.1}),t.blockNameToState&&(this.blockNameToState=new Map(Object.entries(t.blockNameToState).map(([l,c])=>[l,c])));const n=["barrier","structure_void","light"];for(const l of n){const c=this.blockNameToState.get(l);if(c!==void 0){this.hiddenStates.add(c);const h=this.stateToBlockId.get(c);if(h!==void 0)for(const[u,p]of this.stateToBlockId)p===h&&this.hiddenStates.add(u)}}t.items&&(this.itemMapping=new Map(Object.entries(t.items).map(([l,c])=>[l,c]))),t.waterLevels&&(this.waterLevelMap=new Map(Object.entries(t.waterLevels).map(([l,c])=>[Number(l),c])));const a=e.substring(0,e.lastIndexOf("/")),o=[],s=l=>new Promise((c,h)=>{r.load(l,c,void 0,h)}).then(c=>(c.magFilter=pt,c.minFilter=pt,c.colorSpace=Tt,c.flipY=!1,c));if(t.armorTextures)for(const[l,c]of Object.entries(t.armorTextures))for(const[h,u]of Object.entries(c))o.push(s(`${a}/${u}`).then(p=>{this.armorTextureCache.set(`${l}_${h}`,p)}).catch(()=>{}));if(t.entityTextures)for(const[l,c]of Object.entries(t.entityTextures))o.push(s(`${a}/${c}`).then(h=>{this.entityTextureCache.set(l,h)}).catch(()=>{}));await Promise.all(o)}async init(e){this.container=e,this.renderer=new Bf({antialias:!0}),this.renderer.setPixelRatio(window.devicePixelRatio),this.renderer.setClearColor(8900331),this.updateSize(),e.appendChild(this.renderer.domElement),this.setupBasicScene(),this.initialized=!0,this.startRenderLoop()}async loadSnapshot(e){console.log("[PrismarineRenderer] loadSnapshot called:",{initialized:this.initialized,hasAtlas:!!this.atlasMapping,blockCount:e.length}),this.initialized&&(this.atlasMapping&&this.atlasMaterial&&this.transparentAtlasMaterial?await this.loadBlocksTextured(e):this.loadBlocksBasic(e))}applyBlockChanges(e){if(this.initialized)if(this.atlasMapping&&this.atlasMaterial&&this.transparentAtlasMaterial){for(const t of e)this.updateBlockTextured(t);this.rebuildDirtySections()}else for(const t of e)this.updateBlockBasic(t)}createTexturedEntityGroup(e){const t=new gt;if(e.startsWith("item:")){const s=e.substring(5);return this.createItemEntityGroup(s,e)}if(e.startsWith("falling_block:")){const s=parseInt(e.substring(14),10);if(!isNaN(s)&&this.atlasMapping){const l=this.atlasMapping.blocks[s];if(l){const c=this.createBlockMeshFromFaces(l);if(c)return c.scale.set(2.86,2.86,2.86),c.position.y=.5,t.add(c),this.addEntityNametag(t,"falling_block",1.5),t}}}const r=Ke.ENTITY_BLOCK_NAMES[e];if(r){const s=this.createHeldItemMesh(r);if(s)return s.scale.set(2.86,2.86,2.86),s.position.y=.5,t.add(s),this.addEntityNametag(t,e,1.5),t}const n=Ke.ENTITY_ITEM_TEXTURES[e];if(n){const s=this.createEntitySprite(n);if(s)return(e==="end_crystal"||e==="ender_crystal")&&(s.scale.set(3,3,3),s.position.y=1),t.add(s),t}if(e==="item"||e==="dropped_item")return this.createItemEntityGroup("",e);const a=Ke.ENTITY_STYLES[e]||Ke.DEFAULT_ENTITY_STYLE,o=this.entityTextureCache.get(e);if(o&&Ke.HUMANOID_MOBS.has(e)){const s=new ct({map:o,alphaTest:.1}),l=new gt;l.name="bodyPivot";const c=new tt(.5,.75,.25);Ke.setSkinUVs(c,{front:[20,20,8,12],right:[16,20,4,12],back:[32,20,8,12],left:[28,20,4,12],top:[20,16,8,4],bottom:[28,16,8,4]}),l.add(new Fe(c,s)),l.children[0].position.y=.375;const h=new gt;h.name="head",h.position.y=.75;const u=new tt(.5,.5,.5);Ke.setSkinUVs(u,{front:[8,8,8,8],right:[0,8,8,8],back:[24,8,8,8],left:[16,8,8,8],top:[8,0,8,8],bottom:[16,0,8,8]});const p=new Fe(u,s);p.position.y=.25,h.add(p),l.add(h);const f=new gt;f.name="rightArm",f.position.set(.375,.75,0);const g=new tt(.25,.75,.25);Ke.setSkinUVs(g,{front:[44,20,4,12],right:[40,20,4,12],back:[52,20,4,12],left:[48,20,4,12],top:[44,16,4,4],bottom:[48,16,4,4]});const _=new Fe(g,s);_.position.y=-.375,f.add(_),l.add(f);const m=new gt;m.name="leftArm",m.position.set(-.375,.75,0);const d=new tt(.25,.75,.25);Ke.setSkinUVs(d,{front:[36,52,4,12],right:[32,52,4,12],back:[44,52,4,12],left:[40,52,4,12],top:[36,48,4,4],bottom:[40,48,4,4]});const S=new Fe(d,s);S.position.y=-.375,m.add(S),l.add(m),l.position.y=.75,t.add(l);const T=new gt;T.name="rightLeg",T.position.set(.125,.75,0);const x=new tt(.25,.75,.25);Ke.setSkinUVs(x,{front:[4,20,4,12],right:[0,20,4,12],back:[12,20,4,12],left:[8,20,4,12],top:[4,16,4,4],bottom:[8,16,4,4]});const N=new Fe(x,s);N.position.y=-.375,T.add(N),t.add(T);const w=new gt;w.name="leftLeg",w.position.set(-.125,.75,0);const A=new tt(.25,.75,.25);Ke.setSkinUVs(A,{front:[20,52,4,12],right:[16,52,4,12],back:[28,52,4,12],left:[24,52,4,12],top:[20,48,4,4],bottom:[24,48,4,4]});const O=new Fe(A,s);O.position.y=-.375,w.add(O),t.add(w),this.addEntityNametag(t,e,2.3)}else if(o){const s=new ct({map:o,alphaTest:.1}),l=new tt(a.width,a.height,a.depth),c=new Fe(l,s);c.position.y=a.height/2,t.add(c),this.addEntityNametag(t,e,a.height+.5)}else if(Ke.HUMANOID_MOBS.has(e)){const s=a.height-.5,l=new tt(a.width,s,a.depth),c=new ct({color:a.color}),h=new Fe(l,c);h.position.y=s/2,t.add(h);const u=Math.max(a.width,.5),p=new tt(u,u,u),f=new We(a.color);f.offsetHSL(0,0,.15);const g=new ct({color:f}),_=new Fe(p,g);_.position.y=s+u/2,t.add(_),this.addEntityNametag(t,e,a.height+.5)}else{const s=new tt(a.width,a.height,a.depth),l=new ct({color:a.color}),c=new Fe(s,l);c.position.y=a.height/2,t.add(c),this.addEntityNametag(t,e,a.height+.5)}return t}createItemEntityGroup(e,t){const r=new gt;if(e){const s=this.createHeldItemMesh(e);if(s)return s.position.y=.25,s.scale.set(.6,.6,.6),r.add(s),r}const n=new tt(.25,.25,.25),a=new ct({color:16776960}),o=new Fe(n,a);return o.position.y=.25,r.add(o),r}createEntitySprite(e){const t=this.atlasMapping;if(!t||!this.atlasTexture)return null;const r=this.itemMapping.get(e),n=(r==null?void 0:r.texture)||`item/${e}`;let a=t.textures[n];if(!a){const h=Ke.ITEM_TEXTURE_FALLBACKS[e];h&&(a=t.textures[h])}if(!a)return null;const o=new Bi({map:this.atlasTexture,alphaTest:.1,side:Rt,transparent:!0}),s=new vr(.5,.5),l=s.getAttribute("uv");l.setXY(0,a.u,a.v),l.setXY(1,a.u+a.su,a.v),l.setXY(2,a.u,a.v+a.sv),l.setXY(3,a.u+a.su,a.v+a.sv),l.needsUpdate=!0;const c=new Fe(s,o);return c.position.y=.25,c.onBeforeRender=(h,u,p)=>{c.quaternion.copy(p.quaternion)},c}addEntityNametag(e,t,r){const n=t.replace(/_/g," "),a=document.createElement("canvas");a.width=256,a.height=48;const o=a.getContext("2d");o.fillStyle="rgba(0,0,0,0.5)",o.fillRect(0,0,256,48),o.fillStyle="#cccccc",o.font="bold 24px sans-serif",o.textAlign="center",o.fillText(n,128,32);const s=new Ho(a),l=new ls({map:s}),c=new Uo(l);c.position.y=r,c.scale.set(1.5,.35,1),e.add(c)}updateEntity(e,t,r,n,a,o,s){let l=this.entityMeshes.get(e);const c=this.entityTypeNames.get(e);if(l&&s&&s!==c&&(this.scene.remove(l),l=void 0),!l){const u=s||"";l=this.createTexturedEntityGroup(u),this.scene.add(l),this.entityMeshes.set(e,l),u&&this.entityTypeNames.set(e,u)}l.position.set(t,r,n);const h=this.entityTypeNames.get(e);h&&!h.startsWith("item")&&h!=="dropped_item"&&(l.rotation.y=Math.PI-a*(Math.PI/180))}removeEntity(e,t){const r=this.entityMeshes.get(e);if(r){const n=t||this.entityTypeNames.get(e)||"";Ke.EXPLOSIVE_TYPES.has(n)&&this.spawnExplosion(r.position.x,r.position.y+.5,r.position.z,n==="tnt"?2:1,n),this.scene.remove(r),this.entityMeshes.delete(e),this.entityTypeNames.delete(e)}}getKnownEntityIds(){return new Set(this.entityMeshes.keys())}spawnExplosion(e,t,r,n,a){if(a==="wind_charge"){this.spawnWindChargeExplosion(e,t,r);return}const o=Math.floor(40+n*20),s=new Float32Array(o*3),l=new Float32Array(o*3),c=new Float32Array(o*3),h=Math.floor(o*.5);for(let d=0;d<o;d++){const S=d*3;s[S]=e,s[S+1]=t,s[S+2]=r;const T=Math.random()*Math.PI*2,x=Math.random()*Math.PI;if(d<h){const N=(2+Math.random()*3)*n;c[S]=Math.sin(x)*Math.cos(T)*N,c[S+1]=Math.cos(x)*N*.5+Math.random()*1.5,c[S+2]=Math.sin(x)*Math.sin(T)*N;const w=Math.random();l[S]=1,l[S+1]=.3+w*.5,l[S+2]=w*.1}else{const N=(.5+Math.random()*1.5)*n;c[S]=Math.sin(x)*Math.cos(T)*N*.4,c[S+1]=1+Math.random()*2,c[S+2]=Math.sin(x)*Math.sin(T)*N*.4;const w=.15+Math.random()*.25;l[S]=w,l[S+1]=w,l[S+2]=w}}const u=new Ct;u.setAttribute("position",new ht(s,3)),u.setAttribute("color",new ht(l,3));const p=new cs({size:.3*n,vertexColors:!0,transparent:!0,opacity:1,depthWrite:!1,sizeAttenuation:!0}),f=new ko(u,p);this.scene.add(f);const g=new Nn(.3*n,12,12),_=new Bi({color:16777198,transparent:!0,opacity:.9,depthWrite:!1,side:Rt}),m=new Fe(g,_);m.position.set(e,t,r),this.scene.add(m),this.activeExplosions.push({particles:f,velocities:c,startTime:performance.now(),duration:800,extraMeshes:[m]})}spawnWindChargeExplosion(e,t,r){const n=new Float32Array(54),a=new Float32Array(18*3),o=new Float32Array(18*3);for(let u=0;u<18;u++){const p=u*3;n[p]=e,n[p+1]=t,n[p+2]=r;const f=u/18*Math.PI*2+(Math.random()-.5)*.4,g=2.5+Math.random()*1.5;o[p]=Math.cos(f)*g,o[p+1]=(Math.random()-.5)*.5,o[p+2]=Math.sin(f)*g;const _=Math.random();a[p]=.5+_*.5,a[p+1]=.7+_*.3,a[p+2]=1}const s=new Ct;s.setAttribute("position",new ht(n,3)),s.setAttribute("color",new ht(a,3));const l=new cs({size:.2,vertexColors:!0,transparent:!0,opacity:1,depthWrite:!1,sizeAttenuation:!0}),c=new ko(s,l);this.scene.add(c);const h=this.spawnWindChargeRing(e,t,r);this.activeExplosions.push({particles:c,velocities:o,startTime:performance.now(),duration:500,noGravity:!0,extraMeshes:[h]})}spawnWindChargeRing(e,t,r){const n=new In(.3,.5,32),a=new Bi({color:9031664,transparent:!0,opacity:.8,depthWrite:!1,side:Rt}),o=new Fe(n,a);return o.position.set(e,t,r),o.rotation.x=-Math.PI/2,this.scene.add(o),o}updateExplosions(){const e=performance.now();for(let t=this.activeExplosions.length-1;t>=0;t--){const r=this.activeExplosions[t],n=e-r.startTime,a=Math.min(n/r.duration,1);if(a>=1){if(this.scene.remove(r.particles),r.particles.geometry.dispose(),r.particles.material.dispose(),r.extraMeshes)for(const h of r.extraMeshes)this.scene.remove(h),h instanceof Fe&&(h.geometry.dispose(),h.material.dispose());this.activeExplosions.splice(t,1);continue}const o=r.particles.geometry.getAttribute("position"),s=.016,l=o.count;for(let h=0;h<l;h++){const u=h*3;o.array[u]+=r.velocities[u]*s,o.array[u+1]+=r.velocities[u+1]*s,o.array[u+2]+=r.velocities[u+2]*s,r.noGravity||(r.velocities[u+1]-=9.8*s)}o.needsUpdate=!0;const c=r.particles.material;if(c.opacity=1-a,c.size*=.98,r.extraMeshes)for(const h of r.extraMeshes){if(!(h instanceof Fe))continue;const u=h.material;if(h.geometry instanceof Nn){const p=Math.min(a/.2,1),f=1+p*4;h.scale.set(f,f,f),u.opacity=.9*(1-p)}else if(h.geometry instanceof In){const p=Math.min(a/.8,1),f=1+p*(2/.3-1);h.scale.set(f,f,f),u.opacity=.8*(1-p)}}}}createPlayerGroup(e){const t=new gt,r=new gt;r.name="bodyPivot";const n=new tt(.5,.75,.25),a=new ct({color:4491468}),o=new Fe(n,a);o.position.y=.375,r.add(o);const s=new gt;s.name="head",s.rotation.order="YXZ",s.position.y=.75;const l=new Fe(new tt(.5,.5,.5),new ct({color:14527095}));l.position.y=.25,s.add(l),r.add(s);const c=new gt;c.name="rightArm",c.position.set(.375,.75,0);const h=new Fe(new tt(.25,.75,.25),new ct({color:4491468}));h.position.y=-.375,c.add(h),r.add(c);const u=new gt;u.name="leftArm",u.position.set(-.375,.75,0);const p=new Fe(new tt(.25,.75,.25),new ct({color:4491468}));p.position.y=-.375,u.add(p),r.add(u),r.position.y=.75,t.add(r);const f=new gt;f.name="rightLeg",f.position.set(.125,.75,0);const g=new Fe(new tt(.25,.75,.25),new ct({color:3355494}));g.position.y=-.375,f.add(g),t.add(f);const _=new gt;_.name="leftLeg",_.position.set(-.125,.75,0);const m=new Fe(new tt(.25,.75,.25),new ct({color:3355494}));m.position.y=-.375,_.add(m),t.add(_);const d=Ke.createNametagSprite(e,20);return t.add(d),t}static createNametagSprite(e,t){const r=document.createElement("canvas");r.width=512,r.height=64;const n=r.getContext("2d");n.clearRect(0,0,512,64),n.fillStyle="#ffffff",n.font="bold 28px sans-serif",n.textAlign="center",n.textBaseline="middle",n.strokeStyle="rgba(0,0,0,0.8)",n.lineWidth=3,n.strokeText(e,256,20),n.fillText(e,256,20);const a=Math.floor(t/2),o=t%2>=1,s=10,l=12,c=256-s*(l+1)/2;for(let f=0;f<s;f++){const g=c+f*(l+1),_=42;f<a?n.fillStyle="#ff2222":f===a&&o?n.fillStyle="#ff6666":n.fillStyle="rgba(80,0,0,0.5)",n.fillRect(g,_,l,l-2),n.fillRect(g+1,_-2,l/2-1,3),n.fillRect(g+l/2+1,_-2,l/2-2,3)}const h=new Ho(r);h.magFilter=jt;const u=new ls({map:h,transparent:!0,depthTest:!1}),p=new Uo(u);return p.name="nametag",p.position.y=2.3,p.scale.set(3,.4,1),p.renderOrder=999,p}static getArmorColor(e){for(const[t,r]of Object.entries(Ke.ARMOR_COLORS))if(e.includes(t))return r;return null}static getArmorMaterial(e){const t=["diamond","iron","golden","netherite","leather","chainmail","turtle"];for(const r of t)if(e.includes(r))return r==="golden"?"gold":r;return null}static getArmorLayer(e){return e.includes("leggings")?"layer2":"layer1"}static setArmorUVs(e,t,r=!1){const n=Ke.ARMOR_UV_PARTS[t],a=e.getAttribute("uv"),o=64,s=32,l=["right","left","top","bottom","back","front"];for(let c=0;c<6;c++){let h=l[c];r&&(h==="right"?h="left":h==="left"&&(h="right"));const[u,p,f,g]=n[h],_=u/o,m=p/s,d=f/o,S=g/s,T=c*4;r&&h!=="top"&&h!=="bottom"?(a.setXY(T+0,_+d,m),a.setXY(T+1,_,m),a.setXY(T+2,_+d,m+S),a.setXY(T+3,_,m+S)):(a.setXY(T+0,_,m),a.setXY(T+1,_+d,m),a.setXY(T+2,_,m+S),a.setXY(T+3,_+d,m+S))}a.needsUpdate=!0}createArmorMesh(e,t,r,n,a,o=!1){const s=new tt(t,r,n),l=Ke.getArmorMaterial(e),c=Ke.getArmorLayer(e);if(l){const u=`${l}_${c}`,p=this.armorTextureCache.get(u);if(p)return Ke.setArmorUVs(s,a,o),new Fe(s,new ct({map:p,alphaTest:.1}))}const h=Ke.getArmorColor(e)??8947848;return new Fe(s,new ct({color:h}))}updatePlayer(e,t,r,n,a,o,s,l=!1,c=0,h=0,u,p,f=0,g,_=20){var m,d,S,T;let x=this.playerMeshes.get(e);x||(x=this.createPlayerGroup(t),this.scene.add(x),this.playerMeshes.set(e,x));const N=u||e;if(!this.skinTextures.has(N))this.skinTextures.set(N,null),this.loadSkinTexture(N).then(q=>{if(q){this.skinTextures.set(N,q);const z=this.playerMeshes.get(e);z&&this.applySkinToGroup(z,q,N)}});else{const q=this.skinTextures.get(N);q&&x.userData.skinApplied!==N&&this.applySkinToGroup(x,q,N)}x.position.set(r,n,a);const w=this.playerNametagHealth.get(e);if(w===void 0||Math.floor(w)!==Math.floor(_)){this.playerNametagHealth.set(e,_);const q=x.getObjectByName("nametag");q&&(x.remove(q),(m=q.material.map)==null||m.dispose(),q.material.dispose());const z=Ke.createNametagSprite(t,_);x.add(z)}const A=x.getObjectByName("bodyPivot");A&&(l?(A.position.y=.45,A.rotation.x=-.5):(A.position.y=.75,A.rotation.x=0));const O=x.getObjectByName("rightArm"),b=x.getObjectByName("leftArm"),M=x.getObjectByName("rightLeg"),R=x.getObjectByName("leftLeg"),H=h-c,U=c>0&&H>=0&&H<300,F=this.playerPrevPositions.get(e);let K=0,X=0;if(F){const q=r-F.x,z=a-F.z,ae=(h-F.time)/1e3,ne=ae>0?ae:.016;K=Math.sqrt(q*q+z*z)/ne,K>.5?X=-Math.atan2(-q,z)*(180/Math.PI):X=F.moveDir,ae>=.05&&this.playerPrevPositions.set(e,{x:r,z:a,time:h,moveDir:X})}else this.playerPrevPositions.set(e,{x:r,z:a,time:h,moveDir:o});const $=15;let k=this.playerBodyYaw.get(e)??o;if(K>1){let q=X-k;q=(q%360+540)%360-180,k+=q*.25}let Z=o-k;Z=(Z%360+540)%360-180,Math.abs(Z)>$&&(k=o-Math.sign(Z)*$),this.playerBodyYaw.set(e,k),x.rotation.y=Math.PI-k*(Math.PI/180);const ce=x.getObjectByName("head");if(ce){let q=o-k;q=(q%360+540)%360-180,ce.rotation.y=-q*(Math.PI/180),ce.rotation.x=-s*(Math.PI/180)}if((!this._lastDebugLog||h-this._lastDebugLog>1e3)&&(this._lastDebugLog=h,console.log(`[HeadRot] headYaw=${o.toFixed(1)} bodyYaw=${k.toFixed(1)} moveDir=${X.toFixed(1)} diff=${(o-k).toFixed(1)} speed=${K.toFixed(2)}`)),K>.5){const q=Math.sin(h*8/1e3)*(Math.PI/4)*Math.min(K/6,1);M&&(M.rotation.x=q),R&&(R.rotation.x=-q),U||(O&&(O.rotation.x=-q),b&&(b.rotation.x=q))}else M&&(M.rotation.x=0),R&&(R.rotation.x=0);if(U){const q=H/300,z=Math.sin(q*Math.PI)*(Math.PI/2);O&&(O.rotation.x=z),b&&(b.rotation.x=0)}else K<=.5&&(O&&(O.rotation.x=0),b&&(b.rotation.x=0));const ve=p??"",De=this.playerHeldItems.get(e);if(ve!==De&&(this.playerHeldItems.set(e,ve),O)){const q=O.getObjectByName("heldItemMesh");if(q&&O.remove(q),ve&&ve!=="air"){const z=this.createHeldItemMesh(ve);z&&(z.name="heldItemMesh",(d=this.itemMapping.get(ve))!=null&&d.isBlock||this.blockNameToState.has(ve)?z.position.set(0,-.75,-.2):(z.position.set(0,-.55,-.35),z.rotation.set(0,Math.PI/2,-Math.PI/4)),O.add(z))}}if(g){const q=ne=>{const oe=g.get(ne);return oe?typeof oe=="string"?oe:oe.itemName:""},z=Array.from(g.entries()).sort((ne,oe)=>ne[0]-oe[0]).map(([ne,oe])=>`${ne}:${typeof oe=="string"?oe:oe.itemName}`).join(","),ae=this.playerArmorState.get(e);if(z!==ae){this.playerArmorState.set(e,z);const ne=[];x.traverse(ue=>{ue.name.startsWith("armor_")&&ne.push(ue)});for(const ue of ne)(S=ue.parent)==null||S.remove(ue);const oe=q(2),Ae=q(3),Te=q(4),Be=q(5);if(oe&&oe!=="air"){const ue=this.createArmorMesh(oe,.56,.56,.56,"head");ue.name="armor_head",ue.position.y=.25;const pe=x.getObjectByName("head");pe&&pe.add(ue)}if(Ae&&Ae!=="air"){const ue=this.createArmorMesh(Ae,.56,.8,.3,"body");ue.name="armor_chest",ue.position.y=.375;const pe=x.getObjectByName("bodyPivot");pe&&pe.add(ue);for(const ie of["rightArm","leftArm"]){const le=x.getObjectByName(ie);if(le){const ge=this.createArmorMesh(Ae,.3,.8,.3,"arm",ie==="leftArm");ge.name=`armor_${ie}`,ge.position.y=-.375,le.add(ge)}}}if(Te&&Te!=="air")for(const ue of["rightLeg","leftLeg"]){const pe=x.getObjectByName(ue);if(pe){const ie=this.createArmorMesh(Te,.3,.8,.3,"leg",ue==="leftLeg");ie.name=`armor_${ue}`,ie.position.y=-.375,pe.add(ie)}}if(Be&&Be!=="air")for(const ue of["rightLeg","leftLeg"]){const pe=x.getObjectByName(ue);if(pe){const ie=this.createArmorMesh(Be,.3,.5,.35,"leg",ue==="leftLeg");ie.name=`armor_boot_${ue}`,ie.position.set(0,-.5,.025),pe.add(ie)}}const C=q(1),Je=x.getObjectByName("leftArm");if(Je){const ue=Je.getObjectByName("offhandItemMesh");if(ue&&Je.remove(ue),C&&C!=="air"){const pe=this.createHeldItemMesh(C);pe&&(pe.name="offhandItemMesh",(T=this.itemMapping.get(C))!=null&&T.isBlock||this.blockNameToState.has(C)?pe.position.set(0,-.75,-.2):(pe.position.set(0,-.55,-.35),pe.rotation.set(0,Math.PI/2,-Math.PI/4)),Je.add(pe))}}}}const xe=400,G=h-f,J=f>0&&G>=0&&G<xe;x.traverse(q=>{if(q.isMesh){const z=q.material;if(z.emissive)if(J){const ae=1-G/xe;z.emissive.setRGB(ae,0,0)}else z.emissive.setRGB(0,0,0)}})}removePlayer(e){const t=this.playerMeshes.get(e);t&&(this.scene.remove(t),this.playerMeshes.delete(e)),this.playerPrevPositions.delete(e),this.playerHeldItems.delete(e),this.playerBodyYaw.delete(e)}resetPlayerState(){console.log(`[ResetState] Clearing ${this.playerMeshes.size} players, ${this.entityMeshes.size} entities, armor keys: ${Array.from(this.playerArmorState.entries()).map(([e,t])=>`${e.slice(0,8)}=${t}`).join("; ")}`),this.playerMeshes.forEach(e=>this.scene.remove(e)),this.playerMeshes.clear(),this.playerPrevPositions.clear(),this.playerHeldItems.clear(),this.playerArmorState.clear(),this.playerBodyYaw.clear(),this.playerNametagHealth.clear(),this.entityMeshes.forEach(e=>this.scene.remove(e)),this.entityMeshes.clear()}createBlockMeshFromFaces(e){const t=this.atlasMapping;if(!t||!this.atlasTexture)return null;const r=.35,n=new tt(r,r,r),a=n.getAttribute("uv"),o=[e.side,e.side,e.top,e.bottom,e.side,e.side];for(let s=0;s<6;s++){const l=t.textures[o[s]];if(!l)continue;const c=s*4;a.setXY(c+0,l.u,l.v),a.setXY(c+1,l.u+l.su,l.v),a.setXY(c+2,l.u,l.v+l.sv),a.setXY(c+3,l.u+l.su,l.v+l.sv)}return a.needsUpdate=!0,new Fe(n,new ct({map:this.atlasTexture,alphaTest:.1}))}createHeldItemMesh(e){const t=this.atlasMapping,r=this.itemMapping.get(e);if(r!=null&&r.isBlock){const u=this.blockNameToState.get(e);if(u!==void 0&&t){const p=t.blocks[u];if(p&&this.atlasTexture){const f=new tt(.35,.35,.35),g=f.getAttribute("uv"),_=[p.side,p.side,p.top,p.bottom,p.side,p.side];for(let d=0;d<6;d++){const S=t.textures[_[d]];if(!S)continue;const T=d*4;g.setXY(T+0,S.u,S.v),g.setXY(T+1,S.u+S.su,S.v),g.setXY(T+2,S.u,S.v+S.sv),g.setXY(T+3,S.u+S.su,S.v+S.sv)}g.needsUpdate=!0;const m=new ct({map:this.atlasTexture,alphaTest:.1});return new Fe(f,m)}}}if(r!=null&&r.texture&&t&&this.atlasTexture){const u=t.textures[r.texture];if(u){const p=new ct({map:this.atlasTexture,alphaTest:.1,side:Rt}),f=new vr(.85,.85),g=f.getAttribute("uv");return g.setXY(0,u.u,u.v),g.setXY(1,u.u+u.su,u.v),g.setXY(2,u.u,u.v+u.sv),g.setXY(3,u.u+u.su,u.v+u.sv),g.needsUpdate=!0,new Fe(f,p)}}const n=this.blockNameToState.get(e);if(n!==void 0&&t){const u=t.blocks[n];if(u&&this.atlasTexture){const p=new tt(.35,.35,.35),f=p.getAttribute("uv"),g=[u.side,u.side,u.top,u.bottom,u.side,u.side];for(let m=0;m<6;m++){const d=t.textures[g[m]];if(!d)continue;const S=m*4;f.setXY(S+0,d.u,d.v),f.setXY(S+1,d.u+d.su,d.v),f.setXY(S+2,d.u,d.v+d.sv),f.setXY(S+3,d.u+d.su,d.v+d.sv)}f.needsUpdate=!0;const _=new ct({map:this.atlasTexture,alphaTest:.1});return new Fe(p,_)}}const a=Ke.ITEM_TEXTURE_FALLBACKS[e];if(a&&t&&this.atlasTexture){const u=t.textures[a];if(u){const p=new ct({map:this.atlasTexture,alphaTest:.1,side:Rt}),f=new vr(.85,.85),g=f.getAttribute("uv");return g.setXY(0,u.u,u.v),g.setXY(1,u.u+u.su,u.v),g.setXY(2,u.u,u.v+u.sv),g.setXY(3,u.u+u.su,u.v+u.sv),g.needsUpdate=!0,new Fe(f,p)}}const o=new vr(.85,.85);let s=0;for(let u=0;u<e.length;u++)s=(s<<5)-s+e.charCodeAt(u)|0;const l=Math.abs(s)%360/360,c=new We().setHSL(l,.7,.5),h=new ct({color:c,side:Rt});return new Fe(o,h)}setEmbeddedSkin(e,t){this.embeddedSkins.set(e,t)}async preloadEmbeddedSkins(){const e=[];for(const[t]of this.embeddedSkins)(!this.skinTextures.has(t)||this.skinTextures.get(t)===null)&&(this.skinTextures.set(t,null),e.push(this.loadSkinTexture(t).then(r=>{if(r){this.skinTextures.set(t,r);for(const[,n]of this.playerMeshes)n.userData.skinApplied!==t&&this.applySkinToGroup(n,r,t)}})));await Promise.all(e)}getCamera(){return this.camera}getScene(){return this.scene}getDomElement(){var e;return((e=this.renderer)==null?void 0:e.domElement)??null}updateSize(){if(!this.container||!this.renderer)return;const{clientWidth:e,clientHeight:t}=this.container;this.renderer.setSize(e,t),this.camera.aspect=e/t,this.camera.updateProjectionMatrix()}dispose(){var e,t,r,n,a,o;this.animationFrameId!==null&&cancelAnimationFrame(this.animationFrameId),(e=this.renderer)==null||e.dispose(),(r=(t=this.renderer)==null?void 0:t.domElement)!=null&&r.parentElement&&this.renderer.domElement.parentElement.removeChild(this.renderer.domElement),(n=this.atlasTexture)==null||n.dispose(),(a=this.atlasMaterial)==null||a.dispose(),(o=this.transparentAtlasMaterial)==null||o.dispose();for(const s of this.armorTextureCache.values())s.dispose();this.armorTextureCache.clear();for(const s of this.entityTextureCache.values())s.dispose();this.entityTextureCache.clear();for(const s of this.skinTextures.values())s==null||s.dispose();this.skinTextures.clear(),this.playerPrevPositions.clear(),this.playerHeldItems.clear(),this.playerArmorState.clear(),this.playerBodyYaw.clear(),this.playerNametagHealth.clear(),this.entityMeshes.clear(),this.entityTypeNames.clear();for(const s of this.activeExplosions)if(this.scene.remove(s.particles),s.particles.geometry.dispose(),s.particles.material.dispose(),s.extraMeshes)for(const l of s.extraMeshes)this.scene.remove(l),l instanceof Fe&&(l.geometry.dispose(),l.material.dispose());this.activeExplosions.length=0,this.scene.clear(),this.sections.clear(),this.blockMap.clear(),this.initialized=!1}setupBasicScene(){const e=new Jf(16777215,.6);this.scene.add(e);const t=new Zf(16777215,.8);t.position.set(50,100,50),this.scene.add(t)}static sectionKey(e,t,r){return`${e>>4},${t>>4},${r>>4}`}async loadBlocksTextured(e){if(e.length===0)return;for(const a of e)this.blockMap.set(`${a.x},${a.y},${a.z}`,a.stateId);const t=new Map;for(const a of e){const o=Ke.sectionKey(a.x,a.y,a.z);let s=t.get(o);s||(s=[],t.set(o,s)),s.push(a)}const r=Array.from(t.keys()),n=8;for(let a=0;a<r.length;a+=n){const o=r.slice(a,a+n);for(const s of o)this.buildSectionMesh(s);a+n<r.length&&await new Promise(s=>setTimeout(s,0))}console.log(`[PrismarineRenderer] Loaded ${e.length} blocks in ${this.sections.size} sections (textured)`)}updateBlockTextured(e){const t=`${e.x},${e.y},${e.z}`,r=Ke.sectionKey(e.x,e.y,e.z);e.stateId===0?this.blockMap.delete(t):this.blockMap.set(t,e.stateId),this.dirtySections.add(r);const n=e.x,a=e.y,o=e.z;n&15||this.dirtySections.add(Ke.sectionKey(n-1,a,o)),(n&15)===15&&this.dirtySections.add(Ke.sectionKey(n+1,a,o)),a&15||this.dirtySections.add(Ke.sectionKey(n,a-1,o)),(a&15)===15&&this.dirtySections.add(Ke.sectionKey(n,a+1,o)),o&15||this.dirtySections.add(Ke.sectionKey(n,a,o-1)),(o&15)===15&&this.dirtySections.add(Ke.sectionKey(n,a,o+1))}rebuildDirtySections(){for(const e of this.dirtySections)this.buildSectionMesh(e);this.dirtySections.clear()}buildSectionMesh(e){const t=this.sections.get(e);t&&(t.opaque&&(this.scene.remove(t.opaque),t.opaque.geometry.dispose()),t.transparent&&(this.scene.remove(t.transparent),t.transparent.geometry.dispose()),this.sections.delete(e));const[r,n,a]=e.split(",").map(Number),o=r*16,s=n*16,l=a*16,c=this.atlasMapping,h=[],u=[],p=[],f=[];let g=0;const _=[],m=[],d=[],S=[];let T=0;for(let w=0;w<16;w++)for(let A=0;A<16;A++)for(let O=0;O<16;O++){const b=o+w,M=s+A,R=l+O,H=`${b},${M},${R}`,U=this.blockMap.get(H);if(U===void 0||U===0||this.hiddenStates.has(U))continue;const F=c.blocks[U];if(!F)continue;const K=this.transparentStates.has(U),X=this.waterLevelMap.get(U);if(F.shape==="multiface"&&F.faceMask!==void 0){const xe=c.textures[F.side];if(!xe)continue;for(const G of rm)if(F.faceMask&G.mask)for(const J of G.quads){for(let q=0;q<4;q++){const z=J.vertices[q];_.push(b+z[0],M+z[1],R+z[2]),m.push(J.normal[0],J.normal[1],J.normal[2]);const ae=En[q];d.push(xe.u+ae[0]*xe.su,xe.v+ae[1]*xe.sv)}for(const q of bn)S.push(T+q);T+=4}continue}if(F.shape==="cross"){const xe=c.textures[F.side];if(!xe)continue;for(const G of em){for(let J=0;J<4;J++){const q=G.vertices[J];_.push(b+q[0],M+q[1],R+q[2]),m.push(G.normal[0],G.normal[1],G.normal[2]);const z=En[J];d.push(xe.u+z[0]*xe.su,xe.v+z[1]*xe.sv)}for(const J of bn)S.push(T+J);T+=4}continue}if(F.shape==="floor"){const xe=c.textures[F.top]||c.textures[F.side];if(!xe)continue;for(const G of tm){for(let J=0;J<4;J++){const q=G.vertices[J];_.push(b+q[0],M+q[1],R+q[2]),m.push(G.normal[0],G.normal[1],G.normal[2]);const z=En[J];d.push(xe.u+z[0]*xe.su,xe.v+z[1]*xe.sv)}for(const J of bn)S.push(T+J);T+=4}continue}const $=K?_:h,k=K?m:u,Z=K?d:p,ce=K?S:f,ve=()=>K?T:g,De=xe=>{K?T+=xe:g+=xe};for(const[xe,G]of Object.entries(Qf)){const[J,q,z]=G.dir,ae=b+J,ne=M+q,oe=R+z,Ae=`${ae},${ne},${oe}`,Te=this.blockMap.get(Ae);if(Te!==void 0&&Te!==0&&!this.hiddenStates.has(Te)){if(!this.transparentStates.has(Te))continue;if(K){const ue=this.stateToBlockId.get(U),pe=this.stateToBlockId.get(Te);if(ue!==void 0&&ue===pe){const ie=this.waterLevelMap.get(U),le=this.waterLevelMap.get(Te);if(ie===void 0||ie===le)continue}}}let Be;xe==="up"?Be=F.top:xe==="down"?Be=F.bottom:Be=F.side;const C=c.textures[Be];if(!C)continue;const Je=ve();for(let ue=0;ue<4;ue++){const pe=G.vertices[ue];let ie=M+pe[1];X!==void 0&&pe[1]===1&&(ie=M+X),$.push(b+pe[0],ie,R+pe[2]),k.push(G.normal[0],G.normal[1],G.normal[2]);const le=En[ue];Z.push(C.u+le[0]*C.su,C.v+le[1]*C.sv)}for(const ue of bn)ce.push(Je+ue);De(4)}}let x=null,N=null;if(h.length>0){const w=new Ct;w.setAttribute("position",new ht(h,3)),w.setAttribute("normal",new ht(u,3)),w.setAttribute("uv",new ht(p,2)),w.setIndex(f),x=new Fe(w,this.atlasMaterial),this.scene.add(x)}if(_.length>0){const w=new Ct;w.setAttribute("position",new ht(_,3)),w.setAttribute("normal",new ht(m,3)),w.setAttribute("uv",new ht(d,2)),w.setIndex(S),N=new Fe(w,this.transparentAtlasMaterial),this.scene.add(N)}(x||N)&&this.sections.set(e,{opaque:x,transparent:N})}async loadSkinTexture(e){const t=e.replace(/-/g,""),r=this.embeddedSkins.get(e)||this.embeddedSkins.get(t);if(r&&r.length>0)try{const n=new Blob([r],{type:"image/png"}),a=URL.createObjectURL(n),o=new Image;o.src=a,await new Promise((l,c)=>{o.onload=()=>{URL.revokeObjectURL(a),l()},o.onerror=()=>{URL.revokeObjectURL(a),c(new Error("Failed to load embedded skin"))}});const s=new _t(o);return s.magFilter=pt,s.minFilter=pt,s.colorSpace=Tt,s.needsUpdate=!0,s}catch(n){console.warn("[PrismarineRenderer] Failed to load embedded skin for",e,n)}try{const n=`https://mineskin.eu/skin/${t}`,a=new Image;a.crossOrigin="anonymous",a.src=n,await new Promise((s,l)=>{a.onload=()=>s(),a.onerror=()=>l(new Error("Failed to load skin"))});const o=new _t(a);return o.magFilter=pt,o.minFilter=pt,o.colorSpace=Tt,o.needsUpdate=!0,o}catch(n){return console.warn("[PrismarineRenderer] Failed to load skin for",e,n),null}}static setSkinUVs(e,t){const r=e.getAttribute("uv"),n=64,a=64,o=["right","left","top","bottom","back","front"];for(let s=0;s<6;s++){const[l,c,h,u]=t[o[s]],p=l/n,f=c/a,g=(l+h)/n,_=(c+u)/a,m=s*4;r.setXY(m+0,p,1-f),r.setXY(m+1,g,1-f),r.setXY(m+2,p,1-_),r.setXY(m+3,g,1-_)}r.needsUpdate=!0}applySkinToGroup(e,t,r){if(e.userData.skinApplied)return;e.userData.skinApplied=r||!0;const n=new ct({map:t,alphaTest:.1}),a={head:{front:[8,8,8,8],right:[0,8,8,8],back:[24,8,8,8],left:[16,8,8,8],top:[8,0,8,8],bottom:[16,0,8,8]},torso:{front:[20,20,8,12],right:[16,20,4,12],back:[32,20,8,12],left:[28,20,4,12],top:[20,16,8,4],bottom:[28,16,8,4]},rightArmMesh:{front:[44,20,4,12],right:[40,20,4,12],back:[52,20,4,12],left:[48,20,4,12],top:[44,16,4,4],bottom:[48,16,4,4]},leftArmMesh:{front:[36,52,4,12],right:[32,52,4,12],back:[44,52,4,12],left:[40,52,4,12],top:[36,48,4,4],bottom:[40,48,4,4]},rightLegMesh:{front:[4,20,4,12],right:[0,20,4,12],back:[12,20,4,12],left:[8,20,4,12],top:[4,16,4,4],bottom:[8,16,4,4]},leftLegMesh:{front:[20,52,4,12],right:[16,52,4,12],back:[28,52,4,12],left:[24,52,4,12],top:[20,48,4,4],bottom:[24,48,4,4]}};e.traverse(o=>{var s;if(!(o instanceof Fe)||o.name)return;let l=null;const c=((s=o.parent)==null?void 0:s.name)||"";if(c==="head"?l="head":c==="rightArm"?l="rightArmMesh":c==="leftArm"?l="leftArmMesh":c==="rightLeg"?l="rightLegMesh":c==="leftLeg"?l="leftLegMesh":c==="bodyPivot"&&!o.name&&(l="torso"),!l||!a[l])return;const h=o.geometry;h instanceof tt&&(Ke.setSkinUVs(h,a[l]),o.material=n)})}loadBlocksBasic(e){if(e.length===0)return;this.maxInstances=e.length+1024;const t=new tt(1,1,1),r=new ct({vertexColors:!1}),n=new Wf(t,r,this.maxInstances);n.count=e.length,this.instanceCount=e.length;const a=new Qe,o=new We;for(let s=0;s<e.length;s++){const l=e[s];a.identity(),a.setPosition(l.x+.5,l.y+.5,l.z+.5),n.setMatrixAt(s,a);const c=l.stateId*137%360;o.setHSL(c/360,.6,.5),n.setColorAt(s,o),this.blockIndexMap.set(`${l.x},${l.y},${l.z}`,s)}n.instanceMatrix.needsUpdate=!0,n.instanceColor&&(n.instanceColor.needsUpdate=!0),this.scene.add(n),this.instancedMesh=n,console.log(`[PrismarineRenderer] Loaded ${e.length} blocks via InstancedMesh`)}updateBlockBasic(e){if(!this.instancedMesh)return;const t=`${e.x},${e.y},${e.z}`,r=this.blockIndexMap.get(t);if(e.stateId===0)r!==void 0&&(this.instancedMesh.setMatrixAt(r,this.hiddenScale),this.instancedMesh.instanceMatrix.needsUpdate=!0,this.blockIndexMap.delete(t));else if(r!==void 0){const n=new We,a=e.stateId*137%360;n.setHSL(a/360,.6,.5),this.instancedMesh.setColorAt(r,n),this.instancedMesh.instanceColor&&(this.instancedMesh.instanceColor.needsUpdate=!0)}else if(this.instanceCount<this.maxInstances){const n=this.instanceCount++;this.instancedMesh.count=this.instanceCount;const a=new Qe;a.setPosition(e.x+.5,e.y+.5,e.z+.5),this.instancedMesh.setMatrixAt(n,a),this.instancedMesh.instanceMatrix.needsUpdate=!0;const o=new We,s=e.stateId*137%360;o.setHSL(s/360,.6,.5),this.instancedMesh.setColorAt(n,o),this.instancedMesh.instanceColor&&(this.instancedMesh.instanceColor.needsUpdate=!0),this.blockIndexMap.set(t,n)}}startRenderLoop(){const e=()=>{var t;this.animationFrameId=requestAnimationFrame(e),this.updateExplosions(),this.updateEntityAnimations(),(t=this.renderer)==null||t.render(this.scene,this.camera)};e()}updateEntityAnimations(){const e=performance.now()/1e3;this.entityMeshes.forEach((t,r)=>{const n=this.entityTypeNames.get(r);if(n&&n.startsWith("item")||n==="dropped_item"){const a=t.children[0];a&&a.type!=="Sprite"&&(a.rotation.y=e*1.5)}})}};$t.ITEM_TEXTURE_FALLBACKS={enchanted_golden_apple:"item/golden_apple",crossbow:"item/crossbow_standby",compass:"item/compass_16",recovery_compass:"item/recovery_compass_16",clock:"item/clock_00",tipped_arrow:"item/tipped_arrow_base",debug_stick:"item/stick"},$t.EXPLOSIVE_TYPES=new Set(["tnt","wind_charge","ender_crystal","end_crystal","fireball","small_fireball","wither_skull","dragon_fireball"]),$t.ENTITY_ITEM_TEXTURES={arrow:"arrow",spectral_arrow:"spectral_arrow",snowball:"snowball",ender_pearl:"ender_pearl",wind_charge:"wind_charge",fire_charge:"fire_charge",experience_bottle:"experience_bottle",egg:"egg",end_crystal:"end_crystal",ender_crystal:"end_crystal",fireball:"fire_charge",small_fireball:"fire_charge",wither_skull:"wither_skeleton_skull",dragon_fireball:"fire_charge",experience_orb:"experience_bottle",trident:"trident"},$t.ENTITY_BLOCK_NAMES={tnt:"tnt",primed_tnt:"tnt"},$t.HUMANOID_MOBS=new Set(["zombie","skeleton","creeper","enderman","blaze","witch","villager","zombie_villager","pillager","vindicator","evoker","drowned","husk","stray","wither_skeleton","piglin","piglin_brute","zombified_piglin","armor_stand","wandering_trader","vex"]),$t.ENTITY_STYLES={zombie:{width:.6,height:1.95,depth:.6,color:43520},skeleton:{width:.6,height:1.99,depth:.6,color:12632256},creeper:{width:.6,height:1.7,depth:.6,color:65280},spider:{width:1.4,height:.9,depth:1.4,color:6044190},enderman:{width:.6,height:2.9,depth:.6,color:1447446},blaze:{width:.6,height:1.8,depth:.6,color:16746496},witch:{width:.6,height:1.95,depth:.6,color:6950317},slime:{width:.5,height:.5,depth:.5,color:8978312},phantom:{width:.9,height:.5,depth:.9,color:4482730},end_crystal:{width:2,height:2,depth:2,color:16711935},ender_crystal:{width:2,height:2,depth:2,color:16711935},armor_stand:{width:.5,height:1.97,depth:.5,color:9136404},item_frame:{width:.5,height:.5,depth:.1,color:13808780},villager:{width:.6,height:1.95,depth:.6,color:11762252},iron_golem:{width:1.4,height:2.7,depth:1.4,color:13948116},wolf:{width:.6,height:.85,depth:.6,color:13684944},cat:{width:.6,height:.7,depth:.6,color:15245344},cow:{width:.9,height:1.4,depth:.9,color:4469538},pig:{width:.9,height:.9,depth:.9,color:15769760},sheep:{width:.9,height:1.3,depth:.9,color:15790320},chicken:{width:.4,height:.7,depth:.4,color:16777215},horse:{width:1.4,height:1.6,depth:1.4,color:12619872},zombie_horse:{width:1.4,height:1.6,depth:1.4,color:5929538},skeleton_horse:{width:1.4,height:1.6,depth:1.4,color:13158560},donkey:{width:1.4,height:1.5,depth:1.4,color:5918266},mule:{width:1.4,height:1.6,depth:1.4,color:5913120},llama:{width:.9,height:1.87,depth:.9,color:12626056},trader_llama:{width:.9,height:1.87,depth:.9,color:12626056},turtle:{width:1.2,height:.4,depth:1.2,color:3836462},rabbit:{width:.4,height:.5,depth:.4,color:11895642},fox:{width:.6,height:.7,depth:.6,color:13989941},bee:{width:.7,height:.6,depth:.7,color:15582019},goat:{width:.9,height:1.3,depth:.9,color:11904396},frog:{width:.5,height:.55,depth:.5,color:8299859},allay:{width:.35,height:.6,depth:.35,color:5882848},bat:{width:.5,height:.9,depth:.5,color:5127470},squid:{width:.8,height:.8,depth:.8,color:2241382},glow_squid:{width:.8,height:.8,depth:.8,color:694676},dolphin:{width:.9,height:.6,depth:.9,color:8952234},guardian:{width:.85,height:.85,depth:.85,color:6261362},elder_guardian:{width:1.99,height:1.99,depth:1.99,color:11904396},warden:{width:.9,height:2.9,depth:.9,color:739914},ravager:{width:1.95,height:2.2,depth:1.95,color:7763574},ghast:{width:4,height:4,depth:4,color:15790320},magma_cube:{width:.5,height:.5,depth:.5,color:4456448},hoglin:{width:1.4,height:1.4,depth:1.4,color:9132602},zoglin:{width:1.4,height:1.4,depth:1.4,color:8956552},wither:{width:.9,height:3.5,depth:.9,color:2236962},ender_dragon:{width:4,height:4,depth:4,color:1710638},silverfish:{width:.4,height:.3,depth:.4,color:7237230},endermite:{width:.4,height:.3,depth:.4,color:3876692},cave_spider:{width:.7,height:.5,depth:.7,color:803406},shulker:{width:1,height:1,depth:1,color:9133707},breeze:{width:.6,height:1.77,depth:.6,color:9031664},bogged:{width:.6,height:1.99,depth:.6,color:6982202},panda:{width:1.3,height:1.25,depth:1.3,color:15790320},polar_bear:{width:1.4,height:1.4,depth:1.4,color:15790320},ocelot:{width:.6,height:.7,depth:.6,color:15245344},mooshroom:{width:.9,height:1.4,depth:.9,color:11149858},snow_golem:{width:.7,height:1.9,depth:.7,color:15790320},strider:{width:.9,height:1.7,depth:.9,color:10040115},axolotl:{width:.75,height:.42,depth:.75,color:15769792},camel:{width:1.7,height:2.375,depth:1.7,color:13935962},sniffer:{width:1.9,height:1.75,depth:1.9,color:6982212},armadillo:{width:.7,height:.65,depth:.7,color:10515530},vex:{width:.4,height:.8,depth:.4,color:8952268},wandering_trader:{width:.6,height:1.95,depth:.6,color:2250154},creaking:{width:.9,height:2.7,depth:.9,color:3811866},item:{width:.25,height:.25,depth:.25,color:16776960},dropped_item:{width:.25,height:.25,depth:.25,color:16776960},experience_orb:{width:.2,height:.2,depth:.2,color:8454016},minecart:{width:.98,height:.7,depth:.98,color:6710886},chest_minecart:{width:.98,height:.7,depth:.98,color:9136404},hopper_minecart:{width:.98,height:.7,depth:.98,color:6710886},tnt_minecart:{width:.98,height:.7,depth:.98,color:13382451},furnace_minecart:{width:.98,height:.7,depth:.98,color:8947848},spawner_minecart:{width:.98,height:.7,depth:.98,color:3355494},command_block_minecart:{width:.98,height:.7,depth:.98,color:13404245},boat:{width:1.375,height:.5625,depth:1.375,color:9136404},oak_boat:{width:1.375,height:.5625,depth:1.375,color:9136404},spruce_boat:{width:1.375,height:.5625,depth:1.375,color:5978654},birch_boat:{width:1.375,height:.5625,depth:1.375,color:13154170},jungle_boat:{width:1.375,height:.5625,depth:1.375,color:10119744},acacia_boat:{width:1.375,height:.5625,depth:1.375,color:10506797},dark_oak_boat:{width:1.375,height:.5625,depth:1.375,color:3811850},mangrove_boat:{width:1.375,height:.5625,depth:1.375,color:7027258},cherry_boat:{width:1.375,height:.5625,depth:1.375,color:13928090},bamboo_raft:{width:1.375,height:.5625,depth:1.375,color:8034858},oak_chest_boat:{width:1.375,height:.5625,depth:1.375,color:9136404},chest_boat:{width:1.375,height:.5625,depth:1.375,color:9136404},fishing_bobber:{width:.25,height:.25,depth:.25,color:13382451},leash_knot:{width:.375,height:.5,depth:.375,color:9136404},painting:{width:1,height:1,depth:.0625,color:9136404}},$t.DEFAULT_ENTITY_STYLE={width:.6,height:1,depth:.6,color:11163050},$t.ARMOR_COLORS={diamond:52428,iron:12632256,gold:16766720,golden:16766720,netherite:3815994,leather:9127187,chainmail:11053224,chain:11053224,turtle:4235360},$t.ARMOR_UV_PARTS={head:{right:[0,8,8,8],left:[16,8,8,8],top:[8,0,8,8],bottom:[16,0,8,8],front:[8,8,8,8],back:[24,8,8,8]},body:{right:[16,20,4,12],left:[28,20,4,12],top:[20,16,8,4],bottom:[28,16,8,4],front:[20,20,8,12],back:[32,20,8,12]},arm:{right:[40,20,4,12],left:[48,20,4,12],top:[44,16,4,4],bottom:[48,16,4,4],front:[44,20,4,12],back:[52,20,4,12]},leg:{right:[0,20,4,12],left:[8,20,4,12],top:[4,16,4,4],bottom:[8,16,4,4],front:[4,20,4,12],back:[12,20,4,12]}};let im=$t;class nm{constructor(e,t){this.mode="free",this.moveSpeed=20,this.lookSpeed=.002,this.velocity=new B,this.euler=new Yt(0,0,0,"YXZ"),this.keys=new Set,this.isPointerLocked=!1,this.followOffset=new B(0,5,-6),this.followTarget=null,this.smoothPosition=new B,this.smoothLookAt=new B,this.camera=e,this.domElement=t,this.smoothPosition.copy(e.position),this.boundOnKeyDown=this.onKeyDown.bind(this),this.boundOnKeyUp=this.onKeyUp.bind(this),this.boundOnMouseMove=this.onMouseMove.bind(this),this.boundOnClick=this.onClick.bind(this),this.boundOnPointerLockChange=this.onPointerLockChange.bind(this),this.attach()}attach(){document.addEventListener("keydown",this.boundOnKeyDown),document.addEventListener("keyup",this.boundOnKeyUp),document.addEventListener("mousemove",this.boundOnMouseMove),this.domElement.addEventListener("click",this.boundOnClick),document.addEventListener("pointerlockchange",this.boundOnPointerLockChange)}setMode(e){this.mode=e,e==="follow"&&document.pointerLockElement&&document.exitPointerLock()}getMode(){return this.mode}setFollowTarget(e){this.followTarget=e}update(e){this.mode==="free"?this.updateFreeCamera(e):this.mode==="follow"&&this.updateFollowCamera(e)}dispose(){document.removeEventListener("keydown",this.boundOnKeyDown),document.removeEventListener("keyup",this.boundOnKeyUp),document.removeEventListener("mousemove",this.boundOnMouseMove),this.domElement.removeEventListener("click",this.boundOnClick),document.removeEventListener("pointerlockchange",this.boundOnPointerLockChange),document.pointerLockElement===this.domElement&&document.exitPointerLock()}updateFreeCamera(e){const t=new B;this.keys.has("w")&&(t.z-=1),this.keys.has("s")&&(t.z+=1),this.keys.has("a")&&(t.x-=1),this.keys.has("d")&&(t.x+=1),this.isPointerLocked&&this.keys.has("shift")&&(t.y-=1),t.normalize(),t.applyQuaternion(this.camera.quaternion),this.camera.position.addScaledVector(t,this.moveSpeed*e)}updateFollowCamera(e){if(!this.followTarget)return;const{x:t,y:r,z:n,yaw:a}=this.followTarget,o=a*Math.PI/180,s=new B(t-Math.sin(o)*this.followOffset.z,r+this.followOffset.y,n-Math.cos(o)*this.followOffset.z),l=new B(t,r+1.6,n),c=1-Math.exp(-5*e);this.smoothPosition.lerp(s,c),this.smoothLookAt.lerp(l,c),this.camera.position.copy(this.smoothPosition),this.camera.lookAt(this.smoothLookAt)}onKeyDown(e){this.keys.add(e.key.toLowerCase())}onKeyUp(e){this.keys.delete(e.key.toLowerCase())}onMouseMove(e){this.mode!=="free"||!this.isPointerLocked||(this.euler.setFromQuaternion(this.camera.quaternion),this.euler.y-=e.movementX*this.lookSpeed,this.euler.x-=e.movementY*this.lookSpeed,this.euler.x=Math.max(-Math.PI/2,Math.min(Math.PI/2,this.euler.x)),this.camera.quaternion.setFromEuler(this.euler))}onClick(){this.mode==="free"&&!this.isPointerLocked&&this.domElement.requestPointerLock()}onPointerLockChange(){this.isPointerLocked=document.pointerLockElement===this.domElement}}function Oi(i){const e=i.split(".").map(Number);return[e[0]||0,e[1]||0,e[2]||0]}function am(i,e){if(e.versions.includes(i))return i;const[t,r,n]=Oi(i),a=e.versions.filter(s=>{const[l,c]=Oi(s);return l===t&&c===r});if(a.length>0)return a.map(s=>({v:s,dist:Math.abs(Oi(s)[2]-n)})).sort((s,l)=>s.dist-l.dist)[0].v;const o=e.versions.filter(s=>{const[l]=Oi(s);return l===t});return o.length>0?o.map(s=>({v:s,dist:Math.abs(Oi(s)[1]-r)})).sort((s,l)=>s.dist-l.dist)[0].v:e.default}const sm=[.25,.5,1,2,4,8];function Xo(i){const e=Math.floor(i/1e3),t=Math.floor(e/60),r=e%60;return`${t}:${r.toString().padStart(2,"0")}`}const om=({state:i,onPlay:e,onPause:t,onSeek:r,onRateChange:n,visible:a=!0})=>{const[o,s]=Ne.useState(!1),l=Ne.useRef(null),c=i.totalDurationMs>0?i.currentTimeMs/i.totalDurationMs*100:0,h=Ne.useCallback(g=>{r(Number(g.target.value))},[r]),u=Ne.useCallback(g=>{const _=Math.max(0,Math.min(i.totalDurationMs,i.currentTimeMs+g));r(_)},[r,i.currentTimeMs,i.totalDurationMs]);Ne.useEffect(()=>{if(!o)return;const g=_=>{l.current&&!l.current.contains(_.target)&&s(!1)};return document.addEventListener("mousedown",g),()=>document.removeEventListener("mousedown",g)},[o]);const p={background:"none",border:"none",color:"white",fontSize:14,cursor:"pointer",padding:"4px 8px",lineHeight:1,borderRadius:4,display:"flex",alignItems:"center",justifyContent:"center",opacity:.85,transition:"opacity 0.15s"},f="modl-scrubber-styles";return Ne.useEffect(()=>{if(document.getElementById(f))return;const g=document.createElement("style");g.id=f,g.textContent=`
      .modl-scrubber {
        -webkit-appearance: none;
        appearance: none;
        height: 4px;
        border-radius: 2px;
        outline: none;
        cursor: pointer;
        background: transparent;
      }
      .modl-scrubber::-webkit-slider-runnable-track {
        height: 4px;
        border-radius: 2px;
        background: rgba(255,255,255,0.15);
      }
      .modl-scrubber::-webkit-slider-thumb {
        -webkit-appearance: none;
        appearance: none;
        width: 14px;
        height: 14px;
        border-radius: 50%;
        background: white;
        border: none;
        margin-top: -5px;
        box-shadow: 0 1px 4px rgba(0,0,0,0.4);
        cursor: pointer;
        transition: transform 0.1s;
      }
      .modl-scrubber::-webkit-slider-thumb:hover {
        transform: scale(1.2);
      }
      .modl-scrubber::-moz-range-track {
        height: 4px;
        border-radius: 2px;
        background: rgba(255,255,255,0.15);
        border: none;
      }
      .modl-scrubber::-moz-range-thumb {
        width: 14px;
        height: 14px;
        border-radius: 50%;
        background: white;
        border: none;
        box-shadow: 0 1px 4px rgba(0,0,0,0.4);
        cursor: pointer;
      }
      .modl-scrubber::-moz-range-progress {
        height: 4px;
        border-radius: 2px;
        background: rgba(99, 102, 241, 0.8);
      }
    `,document.head.appendChild(g)},[]),L.jsxs("div",{style:{position:"absolute",bottom:24,left:"50%",transform:a?"translateX(-50%) translateY(0)":"translateX(-50%) translateY(12px)",background:"rgba(15, 15, 30, 0.7)",backdropFilter:"blur(24px)",WebkitBackdropFilter:"blur(24px)",border:"1px solid rgba(255, 255, 255, 0.1)",borderRadius:999,padding:"8px 20px",display:"flex",alignItems:"center",gap:10,fontFamily:"system-ui, sans-serif",color:"white",fontSize:13,zIndex:5,opacity:a?1:0,transition:"opacity 0.3s ease, transform 0.3s ease",pointerEvents:a?"auto":"none",userSelect:"none"},onMouseEnter:g=>g.stopPropagation(),children:[L.jsx("button",{onClick:()=>u(-5e3),style:p,title:"Back 5s",children:L.jsxs("svg",{width:"16",height:"16",viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:"2",strokeLinecap:"round",strokeLinejoin:"round",children:[L.jsx("polygon",{points:"19,20 9,12 19,4",fill:"currentColor",stroke:"none"}),L.jsx("line",{x1:"5",y1:"4",x2:"5",y2:"20"})]})}),L.jsx("button",{onClick:i.playing?t:e,style:{...p,fontSize:18,padding:"4px 6px"},title:i.playing?"Pause":"Play",children:i.playing?L.jsxs("svg",{width:"18",height:"18",viewBox:"0 0 24 24",fill:"currentColor",children:[L.jsx("rect",{x:"6",y:"4",width:"4",height:"16",rx:"1"}),L.jsx("rect",{x:"14",y:"4",width:"4",height:"16",rx:"1"})]}):L.jsx("svg",{width:"18",height:"18",viewBox:"0 0 24 24",fill:"currentColor",children:L.jsx("polygon",{points:"6,4 20,12 6,20"})})}),L.jsx("button",{onClick:()=>u(5e3),style:p,title:"Forward 5s",children:L.jsxs("svg",{width:"16",height:"16",viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:"2",strokeLinecap:"round",strokeLinejoin:"round",children:[L.jsx("polygon",{points:"5,4 15,12 5,20",fill:"currentColor",stroke:"none"}),L.jsx("line",{x1:"19",y1:"4",x2:"19",y2:"20"})]})}),L.jsxs("div",{style:{flex:1,minWidth:120,position:"relative"},children:[L.jsx("div",{style:{position:"absolute",top:"50%",left:0,height:4,width:`${c}%`,background:"rgba(99, 102, 241, 0.8)",borderRadius:2,transform:"translateY(-50%)",pointerEvents:"none",zIndex:0}}),L.jsx("input",{type:"range",className:"modl-scrubber",min:0,max:i.totalDurationMs,value:i.currentTimeMs,onChange:h,style:{width:"100%",position:"relative",zIndex:1}})]}),L.jsxs("span",{style:{minWidth:90,textAlign:"center",fontSize:12,fontFamily:"monospace",opacity:.8,whiteSpace:"nowrap"},children:[Xo(i.currentTimeMs)," / ",Xo(i.totalDurationMs)]}),L.jsxs("div",{ref:l,style:{position:"relative"},children:[L.jsxs("button",{onClick:()=>s(!o),style:{...p,background:"rgba(255, 255, 255, 0.08)",borderRadius:6,padding:"3px 10px",fontSize:12,fontFamily:"system-ui, sans-serif"},title:"Playback speed",children:[i.rate,"x"]}),o&&L.jsx("div",{style:{position:"absolute",bottom:"100%",right:0,marginBottom:8,background:"rgba(15, 15, 30, 0.9)",backdropFilter:"blur(16px)",WebkitBackdropFilter:"blur(16px)",border:"1px solid rgba(255, 255, 255, 0.1)",borderRadius:8,padding:4,minWidth:60},children:sm.map(g=>L.jsxs("button",{onClick:()=>{n(g),s(!1)},style:{display:"block",width:"100%",padding:"5px 12px",background:g===i.rate?"rgba(99, 102, 241, 0.3)":"none",border:"none",borderRadius:4,color:"white",fontSize:12,cursor:"pointer",textAlign:"center"},children:[g,"x"]},g))})]})]})};function lm(i){return i.replace(/_/g," ").replace(/\b\w/g,e=>e.toUpperCase())}function cm(i){let e=0;for(let t=0;t<i.length;t++)e=(e<<5)-e+i.charCodeAt(t)|0;return`hsl(${Math.abs(e)%360}, 40%, 35%)`}const hm={enchanted_golden_apple:"item/golden_apple",crossbow:"item/crossbow_standby",compass:"item/compass_16",recovery_compass:"item/recovery_compass_16",clock:"item/clock_00",tipped_arrow:"item/tipped_arrow_base",debug_stick:"item/stick"};function Al(i,e){if(e.items){const a=e.items[i];if(a!=null&&a.texture){const o=e.textures[a.texture];if(o)return o}}const t=`item/${i}`;if(e.textures[t])return e.textures[t];const r=`block/${i}`;if(e.textures[r])return e.textures[r];if(e.blockNameToState&&e.blocks){const a=e.blockNameToState[i];if(a!==void 0){const o=e.blocks[a];if(o){const s=o.side||o.top;if(s&&e.textures[s])return e.textures[s]}}}const n=hm[i];return n&&e.textures[n]?e.textures[n]:null}function um(i){const e=[];for(let t=0;t<10;t++){const r=i-t*2;r>=2?e.push("full"):r>=1?e.push("half"):e.push("empty")}return e}function dm(i){const e=[];for(let t=0;t<10;t++){const r=i-t*2;r>=2?e.push("full"):r>=1?e.push("half"):e.push("empty")}return e}const pm={1:"Speed",2:"Slowness",3:"Haste",4:"Mining Fatigue",5:"Strength",6:"Instant Health",7:"Instant Damage",8:"Jump Boost",9:"Nausea",10:"Regeneration",11:"Resistance",12:"Fire Resistance",13:"Water Breathing",14:"Invisibility",15:"Blindness",16:"Night Vision",17:"Hunger",18:"Weakness",19:"Poison",20:"Wither",21:"Health Boost",22:"Absorption",23:"Saturation",24:"Glowing",25:"Levitation",26:"Luck",27:"Unluck",28:"Slow Falling",29:"Conduit Power",30:"Dolphins Grace",31:"Bad Omen",32:"Hero of the Village",33:"Darkness"},Yo=["I","II","III","IV","V","VI","VII","VIII","IX","X"];function fm(i){const e=new Set([1,3,5,6,8,10,11,12,13,16,21,22,23,26,28,29,32]),t=new Set([2,4,7,9,15,17,18,19,20,25,27,31,33]);return e.has(i)?"rgba(80, 200, 120, 0.7)":t.has(i)?"rgba(220, 80, 80, 0.7)":"rgba(100, 160, 220, 0.7)"}function mm(i){const e=Math.max(0,Math.floor(i/20)),t=Math.floor(e/60),r=e%60;return`${t}:${r.toString().padStart(2,"0")}`}const va=27,Kt=22,gm=({itemName:i,atlas:e})=>{const t=Ne.useRef(null);return Ne.useEffect(()=>{const r=t.current;if(!r)return;const n=r.getContext("2d");if(!n)return;const a=Al(i,e);if(!a){n.clearRect(0,0,Kt,Kt);return}const o=e.atlasImage;n.imageSmoothingEnabled=!1,n.clearRect(0,0,Kt,Kt),n.drawImage(o,a.u*o.width,a.v*o.height,a.su*o.width,a.sv*o.height,0,0,Kt,Kt)},[i,e]),L.jsx("canvas",{ref:t,width:Kt,height:Kt,style:{width:Kt,height:Kt,imageRendering:"pixelated"}})},Ir=({slot:i,label:e,atlas:t})=>{const r=i&&t&&Al(i.itemName,t);return L.jsxs("div",{style:{width:va,height:va,backgroundColor:i?r?"rgba(80, 80, 80, 0.5)":cm(i.itemName):"rgba(139, 139, 139, 0.2)",border:"1px solid rgba(255,255,255,0.1)",borderRadius:2,position:"relative",display:"flex",alignItems:"center",justifyContent:"center"},title:i?`${lm(i.itemName)} x${i.count}`:e||"Empty",children:[i&&L.jsxs(L.Fragment,{children:[r?L.jsx(gm,{itemName:i.itemName,atlas:t}):L.jsx("span",{style:{fontSize:6,color:"white",textAlign:"center",overflow:"hidden",maxWidth:va-4,lineHeight:"8px"},children:i.itemName.split("_").pop()}),i.count>1&&L.jsx("span",{style:{position:"absolute",bottom:0,right:1,fontSize:7,color:"white",fontWeight:"bold",textShadow:"1px 1px 1px rgba(0,0,0,0.8)"},children:i.count})]}),!i&&e&&L.jsx("span",{style:{fontSize:6,color:"rgba(255,255,255,0.3)"},children:e})]})},_m=({health:i})=>{const e=um(i),t={full:"#cc3333",half:"#cc3333",empty:"#333333"};return L.jsxs("div",{style:{display:"flex",alignItems:"center"},children:[e.map((r,n)=>L.jsx("div",{style:{display:"inline-block",width:8,height:8,margin:"0 0.5px",borderRadius:1,backgroundColor:t[r],opacity:r==="empty"?.4:1,...r==="half"?{background:`linear-gradient(to right, ${t.half} 50%, #333333 50%)`}:{}}},n)),L.jsx("span",{style:{color:"#cc3333",fontSize:10,marginLeft:4},children:i.toFixed(0)})]})},vm=({food:i})=>{const e=dm(i),t={full:"#b87333",half:"#b87333",empty:"#333333"};return L.jsxs("div",{style:{display:"flex",alignItems:"center"},children:[e.map((r,n)=>L.jsx("div",{style:{display:"inline-block",width:8,height:8,margin:"0 0.5px",borderRadius:1,backgroundColor:t[r],opacity:r==="empty"?.4:1,...r==="half"?{background:`linear-gradient(to right, ${t.half} 50%, #333333 50%)`}:{}}},n)),L.jsx("span",{style:{color:"#b87333",fontSize:10,marginLeft:4},children:i})]})},xm=({playerState:i,currentTimeMs:e,atlas:t,isFollowing:r,onFollow:n,onUnfollow:a})=>{const o=l=>i.inventory.get(l),s=i.effects.filter(l=>{const c=e-l.appliedAtMs,h=Math.floor(c/50);return l.durationTicks-h>0});return L.jsxs("div",{style:{padding:"8px 12px",borderTop:"1px solid rgba(255,255,255,0.06)"},children:[L.jsx("button",{onClick:r?a:n,style:{width:"100%",padding:"5px 10px",marginBottom:8,background:r?"rgba(99, 102, 241, 0.3)":"rgba(99, 102, 241, 0.15)",border:"1px solid rgba(99, 102, 241, 0.3)",borderRadius:6,color:"white",fontSize:11,cursor:"pointer",transition:"background 0.15s"},children:r?"Unfollow":"Follow Camera"}),L.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:3,marginBottom:8},children:[L.jsx(_m,{health:i.health}),L.jsx(vm,{food:i.food})]}),L.jsxs("div",{style:{marginBottom:8},children:[L.jsx("div",{style:{fontSize:9,color:"rgba(255,255,255,0.4)",marginBottom:4,textTransform:"uppercase",letterSpacing:"0.05em"},children:"Equipment"}),L.jsxs("div",{style:{display:"flex",gap:1},children:[L.jsx(Ir,{slot:o(39),label:"Head",atlas:t}),L.jsx(Ir,{slot:o(38),label:"Chest",atlas:t}),L.jsx(Ir,{slot:o(37),label:"Legs",atlas:t}),L.jsx(Ir,{slot:o(36),label:"Feet",atlas:t}),L.jsx("div",{style:{width:4}}),L.jsx(Ir,{slot:o(40),label:"Off",atlas:t})]})]}),s.length>0&&L.jsxs("div",{style:{marginBottom:8},children:[L.jsx("div",{style:{fontSize:9,color:"rgba(255,255,255,0.4)",marginBottom:4,textTransform:"uppercase",letterSpacing:"0.05em"},children:"Effects"}),L.jsx("div",{style:{display:"flex",flexWrap:"wrap",gap:3},children:s.map((l,c)=>{const h=pm[l.effectTypeId]||`Effect ${l.effectTypeId}`,u=l.amplifier<Yo.length?Yo[l.amplifier]:`${l.amplifier+1}`,p=e-l.appliedAtMs,f=Math.floor(p/50),g=Math.max(0,l.durationTicks-f);return L.jsxs("span",{style:{fontSize:9,padding:"1px 5px",borderRadius:3,background:"rgba(255,255,255,0.06)",color:fm(l.effectTypeId),whiteSpace:"nowrap"},children:[h," ",u," ",mm(g)]},c)})})]}),L.jsxs("div",{children:[L.jsx("div",{style:{fontSize:9,color:"rgba(255,255,255,0.4)",marginBottom:4,textTransform:"uppercase",letterSpacing:"0.05em"},children:"Inventory"}),L.jsx("div",{style:{display:"flex",flexDirection:"column",gap:1,marginBottom:4},children:[0,1,2].map(l=>L.jsx("div",{style:{display:"flex",gap:1},children:Array.from({length:9},(c,h)=>{const u=9+l*9+h;return L.jsx(Ir,{slot:o(u),atlas:t},u)})},l))}),L.jsx("div",{style:{borderTop:"1px solid rgba(255,255,255,0.06)",margin:"3px 0"}}),L.jsx("div",{style:{display:"flex",gap:1},children:Array.from({length:9},(l,c)=>L.jsx(Ir,{slot:o(c),atlas:t},c))})]})]})},ym=({players:i,followingPlayer:e,expandedPlayer:t,onFollowPlayer:r,onExpandPlayer:n,currentTimeMs:a=0,visible:o=!0,selectedPlayerState:s,atlas:l})=>i.length===0?null:L.jsxs("div",{style:{position:"absolute",top:16,right:16,bottom:80,width:280,background:"rgba(15, 15, 30, 0.5)",backdropFilter:"blur(16px)",WebkitBackdropFilter:"blur(16px)",border:"1px solid rgba(255, 255, 255, 0.08)",borderRadius:12,overflowY:"auto",fontFamily:"system-ui, sans-serif",color:"white",fontSize:13,zIndex:5,opacity:o?1:0,transform:o?"translateX(0)":"translateX(12px)",transition:"opacity 0.3s ease, transform 0.3s ease",pointerEvents:o?"auto":"none"},children:[L.jsx("div",{style:{padding:"10px 12px 6px",fontSize:10,textTransform:"uppercase",letterSpacing:"0.05em",opacity:.5},children:"Players"}),i.map(c=>{const h=e===c.uuid,u=t===c.uuid,p=(c.effects||[]).filter(f=>{const g=a-f.appliedAtMs,_=Math.floor(g/50);return f.durationTicks-_>0});return L.jsxs("div",{children:[L.jsxs("button",{onClick:()=>n(u?null:c.uuid),style:{display:"flex",alignItems:"center",justifyContent:"space-between",width:"100%",padding:"6px 12px",background:u?"rgba(99, 102, 241, 0.2)":"none",border:"none",color:"white",textAlign:"left",cursor:"pointer",fontSize:13,transition:"background 0.15s"},title:u?"Collapse":`Expand ${c.playerName}`,children:[L.jsx("span",{children:c.playerName}),h&&L.jsx("span",{style:{opacity:.5,fontSize:10},children:"following"}),!h&&p.length>0&&L.jsxs("span",{style:{opacity:.4,fontSize:10},children:[p.length," effect",p.length>1?"s":""]})]}),u&&s&&s.uuid===c.uuid&&L.jsx(xm,{playerState:s,currentTimeMs:a,atlas:l,isFollowing:h,onFollow:()=>r(c.uuid),onUnfollow:()=>r(null)})]},c.uuid)}),e&&L.jsx("button",{onClick:()=>r(null),style:{display:"block",width:"100%",padding:"6px 12px",marginTop:4,background:"none",border:"none",borderTop:"1px solid rgba(255,255,255,0.06)",color:"rgba(255,255,255,0.5)",textAlign:"center",cursor:"pointer",fontSize:12},children:"Free camera"})]}),Mm=({message:i="Loading replay..."})=>L.jsxs("div",{style:{position:"absolute",inset:0,display:"flex",flexDirection:"column",alignItems:"center",justifyContent:"center",backgroundColor:"rgba(0, 0, 0, 0.7)",color:"white",zIndex:10,fontFamily:"system-ui, sans-serif"},children:[L.jsx("div",{style:{width:40,height:40,border:"3px solid rgba(255, 255, 255, 0.3)",borderTopColor:"white",borderRadius:"50%",animation:"spin 0.8s linear infinite"}}),L.jsx("p",{style:{marginTop:16,fontSize:14},children:i}),L.jsx("style",{children:"@keyframes spin { to { transform: rotate(360deg); } }"})]}),Sm=({onClick:i,visible:e})=>L.jsx("button",{onClick:i,style:{position:"absolute",top:16,left:16,width:40,height:40,borderRadius:"50%",background:"rgba(15, 15, 30, 0.7)",backdropFilter:"blur(16px)",WebkitBackdropFilter:"blur(16px)",border:"1px solid rgba(255, 255, 255, 0.1)",color:"white",fontSize:18,fontFamily:"Georgia, serif",fontStyle:"italic",fontWeight:"bold",cursor:"pointer",display:"flex",alignItems:"center",justifyContent:"center",zIndex:6,opacity:e?1:0,transform:e?"translateY(0)":"translateY(-8px)",transition:"opacity 0.3s ease, transform 0.3s ease",pointerEvents:e?"auto":"none"},title:"Replay info",children:"i"}),Em=({playerName:i,onVerdict:e})=>L.jsxs("div",{style:{textAlign:"center"},children:[L.jsx("h3",{style:{fontSize:16,fontWeight:500,marginBottom:8},children:i}),L.jsx("p",{style:{color:"rgba(255,255,255,0.5)",fontSize:13,marginBottom:24},children:"What is your verdict for this player?"}),L.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:8},children:[L.jsx("button",{onClick:()=>e("legit"),style:{padding:"12px 16px",background:"rgba(34, 197, 94, 0.2)",border:"1px solid rgba(34, 197, 94, 0.4)",borderRadius:8,color:"#22c55e",fontSize:14,fontWeight:600,cursor:"pointer",transition:"background 0.2s",width:"100%"},children:"Legit"}),L.jsx("button",{onClick:()=>e("cheating"),style:{padding:"12px 16px",background:"rgba(239, 68, 68, 0.2)",border:"1px solid rgba(239, 68, 68, 0.4)",borderRadius:8,color:"#ef4444",fontSize:14,fontWeight:600,cursor:"pointer",transition:"background 0.2s",width:"100%"},children:"Cheating"}),L.jsx("button",{onClick:()=>e("unsure"),style:{padding:"12px 16px",background:"rgba(234, 179, 8, 0.2)",border:"1px solid rgba(234, 179, 8, 0.4)",borderRadius:8,color:"#eab308",fontSize:14,fontWeight:600,cursor:"pointer",transition:"background 0.2s",width:"100%"},children:"Unsure"})]})]}),bm=[{type:"aim",label:"Aim"},{type:"trigger",label:"Trigger"},{type:"strafe_wtap",label:"Strafe / W-Tap"},{type:"connection",label:"Connection"},{type:"scaffold",label:"Scaffold"},{type:"automation",label:"Automation"},{type:"xray_esp",label:"X-Ray / ESP"},{type:"movement",label:"Movement"},{type:"other",label:"Other"}],Tm=({selected:i,onToggle:e,onNext:t})=>L.jsxs("div",{children:[L.jsx("h3",{style:{fontSize:16,fontWeight:500,marginBottom:4},children:"Select Cheat Types"}),L.jsx("p",{style:{color:"rgba(255,255,255,0.5)",fontSize:13,marginBottom:16},children:"Select all that apply"}),L.jsx("div",{style:{display:"grid",gridTemplateColumns:"repeat(2, 1fr)",gap:8,marginBottom:20},children:bm.map(({type:r,label:n})=>{const a=i.includes(r);return L.jsx("button",{onClick:()=>e(r),style:{padding:"10px 12px",background:a?"rgba(99, 102, 241, 0.3)":"rgba(255, 255, 255, 0.05)",border:`1px solid ${a?"rgba(99, 102, 241, 0.5)":"rgba(255, 255, 255, 0.1)"}`,borderRadius:8,color:"white",fontSize:13,cursor:"pointer",transition:"all 0.15s"},children:n},r)})}),L.jsx("button",{onClick:t,disabled:i.length===0,style:{width:"100%",padding:"10px 16px",background:i.length>0?"rgba(99, 102, 241, 0.6)":"rgba(255,255,255,0.1)",border:"1px solid rgba(99, 102, 241, 0.4)",borderRadius:8,color:i.length>0?"white":"rgba(255,255,255,0.3)",fontSize:14,fontWeight:500,cursor:i.length>0?"pointer":"default"},children:"Next"})]}),wm={aim:"Aim",trigger:"Trigger",strafe_wtap:"Strafe / W-Tap",connection:"Connection",scaffold:"Scaffold",automation:"Automation",xray_esp:"X-Ray / ESP",movement:"Movement",other:"Other"};function xa(i){const e=Math.floor(i/1e3),t=Math.floor(e/60),r=e%60;return`${t}:${r.toString().padStart(2,"0")}`}const Am=({cheatType:i,ranges:e,currentTimeMs:t,totalDurationMs:r,onAddRange:n,onSetStart:a,onSetEnd:o,onRemoveRange:s,onNext:l})=>{const c=Ne.useRef(null),[h,u]=Ne.useState(null);Ne.useCallback(m=>{var d;const S=(d=c.current)==null?void 0:d.getBoundingClientRect();if(!S||S.width===0)return 0;const T=(m-S.left)/S.width;return Math.max(0,Math.min(r,T*r))},[r]);const p=Ne.useCallback((m,d,S)=>{m.preventDefault(),m.stopPropagation(),m.target.setPointerCapture(m.pointerId);const T=e[d];u({rangeIndex:d,mode:S,startX:m.clientX,origStartMs:T.startMs,origEndMs:T.endMs})},[e]),f=Ne.useCallback(m=>{var d;if(!h)return;const S=(d=c.current)==null?void 0:d.getBoundingClientRect();if(!S||S.width===0)return;const T=(m.clientX-h.startX)/S.width*r;if(h.mode==="move"){const x=h.origEndMs-h.origStartMs;let N=h.origStartMs+T,w=h.origEndMs+T;N<0&&(N=0,w=x),w>r&&(w=r,N=r-x),a(h.rangeIndex,Math.round(N)),o(h.rangeIndex,Math.round(w))}else if(h.mode==="resize-start"){const x=Math.max(0,Math.min(h.origEndMs,h.origStartMs+T));a(h.rangeIndex,Math.round(x))}else if(h.mode==="resize-end"){const x=Math.max(h.origStartMs,Math.min(r,h.origEndMs+T));o(h.rangeIndex,Math.round(x))}},[h,r,a,o]),g=Ne.useCallback(()=>{u(null)},[]),_=6;return L.jsxs("div",{children:[L.jsxs("h3",{style:{fontSize:16,fontWeight:500,marginBottom:4},children:["Time Ranges — ",wm[i]]}),L.jsx("p",{style:{color:"rgba(255,255,255,0.5)",fontSize:13,marginBottom:4},children:"Mark when this cheat is visible. Seek the replay, then click Set Start / Set End — or drag ranges on the timeline."}),L.jsxs("p",{style:{color:"rgba(99, 102, 241, 0.8)",fontSize:12,marginBottom:16},children:["Current time: ",xa(t)]}),L.jsxs("div",{ref:c,onPointerMove:f,onPointerUp:g,onPointerCancel:g,style:{height:24,background:"rgba(255,255,255,0.05)",borderRadius:4,position:"relative",marginBottom:16,touchAction:"none"},children:[e.map((m,d)=>{const S=r>0?m.startMs/r*100:0,T=r>0?(m.endMs-m.startMs)/r*100:0,x=(h==null?void 0:h.rangeIndex)===d;return L.jsxs(Pl.Fragment,{children:[L.jsx("div",{onPointerDown:N=>p(N,d,"move"),style:{position:"absolute",left:`${S}%`,width:`${Math.max(T,.5)}%`,top:0,bottom:0,background:x&&(h==null?void 0:h.mode)==="move"?"rgba(239, 68, 68, 0.6)":"rgba(239, 68, 68, 0.4)",borderRadius:2,cursor:x&&(h==null?void 0:h.mode)==="move"?"grabbing":"grab",zIndex:1}}),L.jsx("div",{onPointerDown:N=>p(N,d,"resize-start"),style:{position:"absolute",left:`${S}%`,width:_,marginLeft:-_/2,top:0,bottom:0,cursor:"ew-resize",zIndex:2,background:x&&(h==null?void 0:h.mode)==="resize-start"?"rgba(34, 197, 94, 0.8)":"rgba(34, 197, 94, 0.5)",borderRadius:2}}),L.jsx("div",{onPointerDown:N=>p(N,d,"resize-end"),style:{position:"absolute",left:`calc(${S+T}% - ${_/2}px)`,width:_,top:0,bottom:0,cursor:"ew-resize",zIndex:2,background:x&&(h==null?void 0:h.mode)==="resize-end"?"rgba(239, 68, 68, 0.8)":"rgba(239, 68, 68, 0.6)",borderRadius:2}})]},d)}),L.jsx("div",{style:{position:"absolute",left:`${r>0?t/r*100:0}%`,top:0,bottom:0,width:2,background:"white",zIndex:3,pointerEvents:"none"}})]}),L.jsx("div",{style:{display:"flex",flexDirection:"column",gap:8,marginBottom:16,maxHeight:200,overflowY:"auto"},children:e.map((m,d)=>L.jsxs("div",{style:{background:"rgba(255,255,255,0.05)",borderRadius:6,padding:"6px 10px",fontSize:13},children:[L.jsxs("div",{style:{display:"flex",alignItems:"center",justifyContent:"space-between",marginBottom:4},children:[L.jsxs("span",{style:{color:"rgba(255,255,255,0.5)"},children:["#",d+1]}),L.jsx("button",{onClick:()=>s(d),style:{background:"none",border:"none",color:"rgba(255,255,255,0.3)",cursor:"pointer",fontSize:14},children:"×"})]}),L.jsxs("div",{style:{display:"flex",alignItems:"center",gap:6},children:[L.jsx("button",{onClick:()=>a(d),style:{padding:"3px 6px",background:"rgba(34, 197, 94, 0.2)",border:"1px solid rgba(34, 197, 94, 0.3)",borderRadius:4,color:"#22c55e",fontSize:10,cursor:"pointer"},children:"Start"}),L.jsx("span",{style:{fontFamily:"monospace",fontSize:11},children:xa(m.startMs)}),L.jsx("span",{style:{color:"rgba(255,255,255,0.3)"},children:"—"}),L.jsx("button",{onClick:()=>o(d),style:{padding:"3px 6px",background:"rgba(239, 68, 68, 0.2)",border:"1px solid rgba(239, 68, 68, 0.3)",borderRadius:4,color:"#ef4444",fontSize:10,cursor:"pointer"},children:"End"}),L.jsx("span",{style:{fontFamily:"monospace",fontSize:11},children:xa(m.endMs)})]})]},d))}),L.jsxs("div",{style:{display:"flex",gap:8},children:[L.jsx("button",{onClick:n,style:{flex:1,padding:"8px 12px",background:"rgba(255,255,255,0.05)",border:"1px solid rgba(255,255,255,0.1)",borderRadius:8,color:"rgba(255,255,255,0.7)",fontSize:13,cursor:"pointer"},children:"+ Add Range"}),L.jsx("button",{onClick:l,style:{flex:1,padding:"8px 12px",background:"rgba(99, 102, 241, 0.6)",border:"1px solid rgba(99, 102, 241, 0.4)",borderRadius:8,color:"white",fontSize:13,fontWeight:500,cursor:"pointer"},children:"Next"})]})]})},Rm=({confidence:i,notes:e,onConfidenceChange:t,onNotesChange:r,onNext:n})=>L.jsxs("div",{children:[L.jsx("h3",{style:{fontSize:16,fontWeight:500,marginBottom:16},children:"Confidence & Notes"}),L.jsxs("div",{style:{marginBottom:20},children:[L.jsxs("label",{style:{display:"block",color:"rgba(255,255,255,0.5)",fontSize:12,marginBottom:8},children:["Confidence: ",i,"%"]}),L.jsx("input",{type:"range",min:0,max:100,value:i,onChange:a=>t(Number(a.target.value)),style:{width:"100%",cursor:"pointer"}}),L.jsxs("div",{style:{display:"flex",justifyContent:"space-between",fontSize:11,color:"rgba(255,255,255,0.3)",marginTop:4},children:[L.jsx("span",{children:"Unsure"}),L.jsx("span",{children:"Certain"})]})]}),L.jsxs("div",{style:{marginBottom:20},children:[L.jsx("label",{style:{display:"block",color:"rgba(255,255,255,0.5)",fontSize:12,marginBottom:8},children:"Notes (optional — e.g., hacked client name)"}),L.jsx("textarea",{value:e,onChange:a=>r(a.target.value),placeholder:"Any additional observations...",style:{width:"100%",minHeight:80,padding:"8px 12px",background:"rgba(255,255,255,0.05)",border:"1px solid rgba(255,255,255,0.1)",borderRadius:8,color:"white",fontSize:13,fontFamily:"system-ui, sans-serif",resize:"vertical"}})]}),L.jsx("button",{onClick:n,style:{width:"100%",padding:"10px 16px",background:"rgba(99, 102, 241, 0.6)",border:"1px solid rgba(99, 102, 241, 0.4)",borderRadius:8,color:"white",fontSize:14,fontWeight:500,cursor:"pointer"},children:"Next Player"})]}),Cm={aim:"Aim",trigger:"Trigger",strafe_wtap:"Strafe / W-Tap",connection:"Connection",scaffold:"Scaffold",automation:"Automation",xray_esp:"X-Ray / ESP",movement:"Movement",other:"Other"};function Lm(i){return i==="legit"?"#22c55e":i==="unsure"?"#eab308":"#ef4444"}const Pm=({labels:i,onSubmit:e,onClose:t,labelApiUrl:r,replayId:n})=>{const[a,o]=Ne.useState(!1),[s,l]=Ne.useState(!1),[c,h]=Ne.useState(null),u=!!(e||r&&n),p=async()=>{o(!0),h(null);try{if(e&&await e(i),r&&n){const f=`${r.replace(/\/+$/,"")}/v1/public/replays/${encodeURIComponent(n)}/label`,g=await fetch(f,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({players:i})});if(g.status===409)throw new Error("This replay has already been labeled");if(!g.ok)throw new Error(`Server returned ${g.status}`)}l(!0)}catch(f){h(f instanceof Error?f.message:String(f))}finally{o(!1)}};return L.jsxs("div",{children:[L.jsx("h3",{style:{fontSize:16,fontWeight:500,marginBottom:16},children:"Review Labels"}),L.jsx("div",{style:{display:"flex",flexDirection:"column",gap:8,marginBottom:20,maxHeight:260,overflowY:"auto"},children:i.map(f=>L.jsxs("div",{style:{background:"rgba(255,255,255,0.05)",borderRadius:8,padding:"10px 14px",fontSize:13},children:[L.jsxs("div",{style:{display:"flex",justifyContent:"space-between",marginBottom:4},children:[L.jsx("span",{style:{fontWeight:500},children:f.playerName}),L.jsx("span",{style:{color:Lm(f.verdict),fontWeight:600,fontSize:12,textTransform:"uppercase"},children:f.verdict})]}),f.verdict==="cheating"&&f.cheats.length>0&&L.jsx("div",{style:{display:"flex",gap:4,flexWrap:"wrap",marginBottom:4},children:f.cheats.map(g=>L.jsx("span",{style:{fontSize:10,padding:"1px 6px",background:"rgba(239, 68, 68, 0.2)",borderRadius:3,color:"#fca5a5"},children:Cm[g.type]},g.type))}),L.jsxs("div",{style:{color:"rgba(255,255,255,0.4)",fontSize:11},children:["Confidence: ",f.confidence,"%",f.notes&&` — ${f.notes}`]})]},f.uuid))}),u&&!s&&L.jsx("button",{onClick:p,disabled:a,style:{width:"100%",padding:"10px 16px",background:a?"rgba(99, 102, 241, 0.3)":"rgba(99, 102, 241, 0.6)",border:"1px solid rgba(99, 102, 241, 0.4)",borderRadius:8,color:"white",fontSize:14,fontWeight:500,cursor:a?"not-allowed":"pointer",marginBottom:8,display:"flex",alignItems:"center",justifyContent:"center",gap:8},children:a?L.jsxs(L.Fragment,{children:[L.jsx("span",{style:{display:"inline-block",width:14,height:14,border:"2px solid rgba(255,255,255,0.3)",borderTopColor:"white",borderRadius:"50%",animation:"modl-spin 0.6s linear infinite"}}),"Submitting..."]}):"Submit Labels"}),s&&L.jsx("div",{style:{padding:"10px 14px",background:"rgba(34, 197, 94, 0.15)",border:"1px solid rgba(34, 197, 94, 0.3)",borderRadius:8,color:"#22c55e",fontSize:13,marginBottom:8,textAlign:"center"},children:"Labels submitted successfully"}),c&&L.jsx("div",{style:{padding:"10px 14px",background:"rgba(239, 68, 68, 0.15)",border:"1px solid rgba(239, 68, 68, 0.3)",borderRadius:8,color:"#ef4444",fontSize:12,marginBottom:8},children:c}),L.jsx("button",{onClick:t,style:{width:"100%",padding:"10px 16px",background:"rgba(255,255,255,0.05)",border:"1px solid rgba(255,255,255,0.1)",borderRadius:8,color:"rgba(255,255,255,0.7)",fontSize:14,cursor:"pointer"},children:"Close"}),L.jsx("style",{children:"@keyframes modl-spin { to { transform: rotate(360deg); } }"})]})};function Dm(){return{step:"verdict",playerIndex:0,labels:[],currentVerdict:null,selectedCheats:[],cheatDetails:[],currentCheatIndex:0,confidence:75,notes:""}}function Im(i){return{...i,currentVerdict:null,selectedCheats:[],cheatDetails:[],currentCheatIndex:0,confidence:75,notes:""}}function Nm(i,e){switch(e.type){case"SET_VERDICT":return e.verdict==="legit"||e.verdict==="unsure"?{...i,currentVerdict:e.verdict,step:"confidence"}:{...i,currentVerdict:"cheating",step:"cheat_types"};case"TOGGLE_CHEAT":{const t=i.selectedCheats.includes(e.cheatType)?i.selectedCheats.filter(r=>r!==e.cheatType):[...i.selectedCheats,e.cheatType];return{...i,selectedCheats:t}}case"CHEATS_NEXT":{const t=i.selectedCheats.map(r=>({type:r,timeRanges:[{startMs:0,endMs:0}]}));return{...i,cheatDetails:t,currentCheatIndex:0,step:"time_ranges"}}case"ADD_RANGE":{const t=[...i.cheatDetails],r=t[i.currentCheatIndex];return t[i.currentCheatIndex]={...r,timeRanges:[...r.timeRanges,{startMs:0,endMs:0}]},{...i,cheatDetails:t}}case"SET_RANGE_START":{const t=[...i.cheatDetails],r=t[i.currentCheatIndex],n=[...r.timeRanges];return n[e.index]={...n[e.index],startMs:e.timeMs},t[i.currentCheatIndex]={...r,timeRanges:n},{...i,cheatDetails:t}}case"SET_RANGE_END":{const t=[...i.cheatDetails],r=t[i.currentCheatIndex],n=[...r.timeRanges];return n[e.index]={...n[e.index],endMs:e.timeMs},t[i.currentCheatIndex]={...r,timeRanges:n},{...i,cheatDetails:t}}case"REMOVE_RANGE":{const t=[...i.cheatDetails],r=t[i.currentCheatIndex],n=r.timeRanges.filter((a,o)=>o!==e.index);return t[i.currentCheatIndex]={...r,timeRanges:n.length>0?n:[{startMs:0,endMs:0}]},{...i,cheatDetails:t}}case"TIME_RANGES_NEXT":return i.currentCheatIndex<i.cheatDetails.length-1?{...i,currentCheatIndex:i.currentCheatIndex+1}:{...i,step:"confidence"};case"SET_CONFIDENCE":return{...i,confidence:e.value};case"SET_NOTES":return{...i,notes:e.value};case"CONFIDENCE_NEXT":{const t=e.players[i.playerIndex],r={uuid:t.uuid,playerName:t.playerName,verdict:i.currentVerdict,confidence:i.confidence,cheats:i.currentVerdict==="cheating"?i.cheatDetails:[],notes:i.notes},n=[...i.labels,r];return i.playerIndex<e.players.length-1?{...Im(i),labels:n,playerIndex:i.playerIndex+1,step:"verdict"}:{...i,labels:n,step:"review"}}default:return i}}const Um=({players:i,currentTimeMs:e,totalDurationMs:t,onClose:r,onSubmit:n,labelApiUrl:a,replayId:o})=>{const[s,l]=Ne.useReducer(Nm,void 0,Dm),c=i[s.playerIndex];return L.jsxs("div",{style:{padding:"14px 16px",color:"white",fontFamily:"system-ui, sans-serif"},children:[L.jsx("button",{onClick:r,style:{background:"none",border:"none",color:"rgba(255, 255, 255, 0.5)",fontSize:12,cursor:"pointer",padding:"0 0 8px",display:"flex",alignItems:"center",gap:4},children:"← Back"}),s.step!=="review"&&L.jsxs("div",{style:{fontSize:11,color:"rgba(255,255,255,0.4)",marginBottom:12},children:["Player ",s.playerIndex+1," of ",i.length]}),s.step==="verdict"&&c&&L.jsx(Em,{playerName:c.playerName,onVerdict:h=>l({type:"SET_VERDICT",verdict:h})}),s.step==="cheat_types"&&L.jsx(Tm,{selected:s.selectedCheats,onToggle:h=>l({type:"TOGGLE_CHEAT",cheatType:h}),onNext:()=>l({type:"CHEATS_NEXT"})}),s.step==="time_ranges"&&s.cheatDetails[s.currentCheatIndex]&&L.jsx(Am,{cheatType:s.cheatDetails[s.currentCheatIndex].type,ranges:s.cheatDetails[s.currentCheatIndex].timeRanges,currentTimeMs:e,totalDurationMs:t,onAddRange:()=>l({type:"ADD_RANGE"}),onSetStart:(h,u)=>l({type:"SET_RANGE_START",index:h,timeMs:u??e}),onSetEnd:(h,u)=>l({type:"SET_RANGE_END",index:h,timeMs:u??e}),onRemoveRange:h=>l({type:"REMOVE_RANGE",index:h}),onNext:()=>l({type:"TIME_RANGES_NEXT"})}),s.step==="confidence"&&L.jsx(Rm,{confidence:s.confidence,notes:s.notes,onConfidenceChange:h=>l({type:"SET_CONFIDENCE",value:h}),onNotesChange:h=>l({type:"SET_NOTES",value:h}),onNext:()=>l({type:"CONFIDENCE_NEXT",players:i})}),s.step==="review"&&L.jsx(Pm,{labels:s.labels,onSubmit:n,onClose:r,labelApiUrl:a,replayId:o})]})};function Om(i){const e=Math.floor(i/1e3),t=Math.floor(e/60),r=e%60;return`${t}:${r.toString().padStart(2,"0")}`}const Fm=({replayId:i,duration:e,playerCount:t,playerNames:r,onClose:n,players:a,currentTimeMs:o,totalDurationMs:s,onLabelSubmit:l,labelApiUrl:c,labeled:h})=>{const[u,p]=Ne.useState(!1);return L.jsxs("div",{style:{position:"absolute",top:16,left:16,bottom:80,width:300,background:"rgba(15, 15, 30, 0.5)",backdropFilter:"blur(16px)",WebkitBackdropFilter:"blur(16px)",border:"1px solid rgba(255, 255, 255, 0.08)",borderRadius:12,overflowY:"auto",fontFamily:"system-ui, sans-serif",color:"white",fontSize:13,zIndex:5,display:"flex",flexDirection:"column"},children:[L.jsx("button",{onClick:n,style:{position:"absolute",top:8,right:12,background:"none",border:"none",color:"rgba(255, 255, 255, 0.5)",fontSize:18,cursor:"pointer",lineHeight:1,zIndex:1},children:"×"}),u?L.jsx(Um,{players:a,currentTimeMs:o,totalDurationMs:s,onClose:()=>p(!1),onSubmit:l,labelApiUrl:c,replayId:i}):L.jsxs("div",{style:{padding:"14px 16px",display:"flex",flexDirection:"column",gap:14},children:[L.jsxs("div",{children:[L.jsx("div",{style:{fontSize:10,textTransform:"uppercase",letterSpacing:"0.05em",opacity:.5,marginBottom:8},children:"Replay Info"}),L.jsxs("div",{style:{display:"flex",flexDirection:"column",gap:10,fontSize:13},children:[L.jsxs("div",{children:[L.jsx("span",{style:{color:"rgba(255,255,255,0.5)",fontSize:10,textTransform:"uppercase",letterSpacing:"0.05em"},children:"Replay ID"}),L.jsx("div",{style:{marginTop:2,fontFamily:"monospace",fontSize:12},children:i||"Local File"})]}),L.jsxs("div",{children:[L.jsx("span",{style:{color:"rgba(255,255,255,0.5)",fontSize:10,textTransform:"uppercase",letterSpacing:"0.05em"},children:"Duration"}),L.jsx("div",{style:{marginTop:2},children:Om(e)})]}),L.jsxs("div",{children:[L.jsxs("span",{style:{color:"rgba(255,255,255,0.5)",fontSize:10,textTransform:"uppercase",letterSpacing:"0.05em"},children:["Players (",t,")"]}),L.jsx("div",{style:{marginTop:4,display:"flex",flexWrap:"wrap",gap:4},children:r.map(f=>L.jsx("span",{style:{background:"rgba(255,255,255,0.08)",borderRadius:4,padding:"2px 6px",fontSize:11},children:f},f))})]})]})]}),L.jsx("div",{style:{borderTop:"1px solid rgba(255,255,255,0.08)"}}),L.jsxs("div",{children:[L.jsx("div",{style:{fontSize:10,textTransform:"uppercase",letterSpacing:"0.05em",opacity:.5,marginBottom:8},children:"Keyboard Shortcuts"}),L.jsx("div",{style:{display:"flex",flexDirection:"column",gap:6},children:[{key:"Space",action:"Play / Pause"},{key:"←",action:"Rewind 5s"},{key:"→",action:"Forward 5s"}].map(({key:f,action:g})=>L.jsxs("div",{style:{display:"flex",alignItems:"center",gap:8},children:[L.jsx("span",{style:{background:"rgba(255,255,255,0.1)",borderRadius:4,padding:"2px 8px",fontSize:11,fontFamily:"monospace",minWidth:44,textAlign:"center"},children:f}),L.jsx("span",{style:{fontSize:12,opacity:.7},children:g})]},f))})]}),L.jsx("div",{style:{borderTop:"1px solid rgba(255,255,255,0.08)"}}),h?L.jsx("div",{style:{width:"100%",padding:"8px 14px",background:"rgba(255,255,255,0.05)",border:"1px solid rgba(255,255,255,0.1)",borderRadius:8,color:"rgba(255,255,255,0.4)",fontSize:13,fontWeight:500,textAlign:"center"},children:"Already Labeled"}):L.jsx("button",{onClick:()=>p(!0),style:{width:"100%",padding:"8px 14px",background:"rgba(99, 102, 241, 0.6)",border:"1px solid rgba(99, 102, 241, 0.4)",borderRadius:8,color:"white",fontSize:13,fontWeight:500,cursor:"pointer",transition:"background 0.2s"},children:"Label This Replay"})]})]})},zm={playing:!1,currentTimeMs:0,totalDurationMs:0,rate:1,followingPlayer:null},Bm=3e3,Vm=({replayUrl:i,mcVersion:e,atlasUrl:t,atlasMappingUrl:r,atlasBaseUrl:n="/atlas",onLoad:a,onError:o,replayId:s,labelApiUrl:l,onLabelSubmit:c,labeled:h})=>{const u=Ne.useRef(null),p=Ne.useRef(null),f=Ne.useRef(null),g=Ne.useRef(null),_=Ne.useRef(new Map),[m,d]=Ne.useState(!0),[S,T]=Ne.useState(null),[x,N]=Ne.useState(zm),[w,A]=Ne.useState([]),[O,b]=Ne.useState(null),[M,R]=Ne.useState(null),H=Ne.useRef(null),[U,F]=Ne.useState(!0),[K,X]=Ne.useState(!1),[$,k]=Ne.useState(null),Z=Ne.useRef(),ce=Ne.useRef(x);ce.current=x,Ne.useEffect(()=>{const z=u.current;if(!z)return;const ae=()=>{F(!0),Z.current&&clearTimeout(Z.current),Z.current=setTimeout(()=>{K||F(!1)},Bm)},ne=()=>ae(),oe=()=>{Z.current&&clearTimeout(Z.current),K||F(!1)};return z.addEventListener("mousemove",ne),z.addEventListener("mouseleave",oe),ae(),()=>{z.removeEventListener("mousemove",ne),z.removeEventListener("mouseleave",oe),Z.current&&clearTimeout(Z.current)}},[K]),Ne.useEffect(()=>{K&&(F(!0),Z.current&&clearTimeout(Z.current))},[K]),Ne.useEffect(()=>{let z=!1;async function ae(){try{const ne=await fetch(i);if(!ne.ok)throw new Error(`Failed to fetch replay: ${ne.status} ${ne.statusText}`);const oe=await ne.arrayBuffer();if(z)return;const Ae=await Nl(oe);if(z)return;console.log("[ReplayViewer] Parsed replay:",{header:Ae.header,snapshotBlocks:Ae.snapshot.length,events:Ae.events.length,firstBlocks:Ae.snapshot.slice(0,5)});const Te=new Fl(Ae);p.current=Te;const Be=new im(e);if(f.current=Be,u.current){let C=t,Je=r;if(!C||!Je)try{const pe=Ae.header.mcVersion||e,ie=await fetch(`${n}/versions.json`);if(ie.ok){const le=await ie.json(),ge=am(pe,le);C=C||`${n}/${ge}/atlas.png`,Je=Je||`${n}/${ge}/atlas-data.json`,console.log(`[ReplayViewer] Auto-resolved atlas version: ${ge} (replay MC: ${pe})`)}}catch(pe){console.warn("Failed to fetch atlas version manifest:",pe)}if(C&&Je)try{const pe=await fetch(Je);if(pe.ok){const ie=await pe.json();if(z)return;await Be.setTextureAtlas(C,ie);const le=new Image;le.crossOrigin="anonymous",le.src=C,le.onload=()=>{z||b({atlasImage:le,textures:ie.textures,items:ie.items,blockNameToState:ie.blockNameToState,blocks:ie.blocks})}}}catch(pe){console.warn("Failed to load texture atlas, falling back to basic rendering:",pe)}if(await Be.init(u.current),z)return;await Be.loadSnapshot(Ae.snapshot);const ue=Be.getDomElement();if(ue){const pe=new nm(Be.getCamera(),ue);g.current=pe;let ie=Ae.header.targetX,le=Ae.header.targetY,ge=Ae.header.targetZ;const E=Ae.events.find(v=>v.type===Pe.PLAYER_SPAWN);if(E)ie=E.x,le=E.y,ge=E.z;else if(ie===0&&le===0&&ge===0&&Ae.snapshot.length>0){let v=0,V=0,Q=0;for(const re of Ae.snapshot)v+=re.x,V+=re.y,Q+=re.z;ie=v/Ae.snapshot.length,le=V/Ae.snapshot.length,ge=Q/Ae.snapshot.length}Be.getCamera().position.set(ie-10,le+15,ge+10),Be.getCamera().lookAt(ie,le,ge)}}Te.subscribe((C,Je)=>{if(z)return;N(C),Je.length>0&&f.current&&f.current.applyBlockChanges(Je);const ue=Te.getPlayers(),pe=[],ie=f.current;if(ue.forEach(le=>{if(le.alive){pe.push({uuid:le.uuid,playerName:le.playerName,effects:le.effects});const ge=Te.getPlayerSkin(le.uuid);ge&&ie&&(ie.setEmbeddedSkin(le.uuid,ge),le.skinUuid&&le.skinUuid!==le.uuid&&ie.setEmbeddedSkin(le.skinUuid,ge));const E=Te.getInterpolatedPlayerPosition(le.uuid)||{x:le.x,y:le.y,z:le.z,yaw:le.yaw,pitch:le.pitch};ie==null||ie.updatePlayer(le.uuid,le.playerName,E.x,E.y,E.z,E.yaw,E.pitch,le.sneaking,le.lastSwingTime,C.currentTimeMs,le.skinUuid,le.heldItem,le.lastDamageTime,le.equipment,le.health)}else ie==null||ie.removePlayer(le.uuid)}),A(pe),ie){const le=Te.getEntities(),ge=new Set;le.forEach(E=>{ge.add(E.entityId),_.current.set(E.entityId,E.entityTypeName);const v=Te.getInterpolatedEntityPosition(E.entityId)||{x:E.x,y:E.y,z:E.z,yaw:E.yaw,pitch:E.pitch};ie.updateEntity(E.entityId,v.x,v.y,v.z,v.yaw,v.pitch,E.entityTypeName)}),ie.getKnownEntityIds().forEach(E=>{if(!ge.has(E)){const v=_.current.get(E);ie.removeEntity(E,v),_.current.delete(E)}})}if(H.current){const le=ue.get(H.current);le&&R({uuid:le.uuid,playerName:le.playerName,health:le.health,food:le.food,saturation:le.saturation,effects:le.effects,equipment:new Map(le.equipment),inventory:new Map(le.inventory)})}if(C.followingPlayer&&g.current){const le=Te.getInterpolatedPlayerPosition(C.followingPlayer);le&&(g.current.setMode("follow"),g.current.setFollowTarget(le))}}),Te.seek(0),d(!1),a==null||a(),Be&&Be.preloadEmbeddedSkins().catch(()=>{})}catch(ne){if(z)return;const oe=ne instanceof Error?ne:new Error(String(ne));T(oe.message),d(!1),o==null||o(oe)}}return ae(),()=>{var ne,oe,Ae;z=!0,(ne=p.current)==null||ne.dispose(),(oe=f.current)==null||oe.dispose(),(Ae=g.current)==null||Ae.dispose()}},[i,e,t,r,n,a,o]),Ne.useEffect(()=>{const z=()=>{var oe;return(oe=f.current)==null?void 0:oe.updateSize()};window.addEventListener("resize",z);const ae=u.current;let ne=null;return ae&&(ne=new ResizeObserver(()=>{var oe;return(oe=f.current)==null?void 0:oe.updateSize()}),ne.observe(ae)),()=>{window.removeEventListener("resize",z),ne==null||ne.disconnect()}},[]),Ne.useEffect(()=>{let z=performance.now(),ae;const ne=oe=>{var Ae;const Te=(oe-z)/1e3;z=oe,(Ae=g.current)==null||Ae.update(Te),ae=requestAnimationFrame(ne)};return ae=requestAnimationFrame(ne),()=>cancelAnimationFrame(ae)},[]),Ne.useEffect(()=>{if(w.length>0&&$===null){const z=w[0].uuid;k(z),H.current=z}},[w,$]);const ve=Ne.useCallback(()=>{var z;return(z=p.current)==null?void 0:z.play()},[]),De=Ne.useCallback(()=>{var z;return(z=p.current)==null?void 0:z.pause()},[]),xe=Ne.useCallback(z=>{var ae,ne,oe;(ae=f.current)==null||ae.resetPlayerState(),(ne=p.current)==null||ne.seek(z);const Ae=(oe=p.current)==null?void 0:oe.getPlayers();Ae&&Ae.forEach((Te,Be)=>{const C=Array.from(Te.equipment.entries()).map(([Je,ue])=>`${Je}:${typeof ue=="string"?ue:ue.itemName}`).join(",");console.log(`[PostSeek] t=${z.toFixed(0)} player=${Te.playerName} equip=[${C}]`)})},[]),G=Ne.useCallback(z=>{var ae;return(ae=p.current)==null?void 0:ae.setRate(z)},[]),J=Ne.useCallback(z=>{var ae,ne,oe;(ae=p.current)==null||ae.followPlayer(z),z?(ne=g.current)==null||ne.setMode("follow"):(oe=g.current)==null||oe.setMode("free")},[]),q=Ne.useCallback(z=>{var ae;if(k(z),H.current=z,z){const ne=(ae=p.current)==null?void 0:ae.getPlayers();if(ne){const oe=ne.get(z);oe&&R({uuid:oe.uuid,playerName:oe.playerName,health:oe.health,food:oe.food,saturation:oe.saturation,effects:oe.effects,equipment:new Map(oe.equipment),inventory:new Map(oe.inventory)})}}else R(null)},[]);return Ne.useEffect(()=>{const z=ae=>{if(ae.target instanceof HTMLInputElement||ae.target instanceof HTMLTextAreaElement)return;const ne=ce.current;switch(ae.key){case" ":ae.preventDefault(),ne.playing?De():ve();break;case"ArrowRight":ae.preventDefault(),xe(Math.min(ne.totalDurationMs,ne.currentTimeMs+5e3));break;case"ArrowLeft":ae.preventDefault(),xe(Math.max(0,ne.currentTimeMs-5e3));break}};return document.addEventListener("keydown",z),()=>document.removeEventListener("keydown",z)},[ve,De,xe]),L.jsxs("div",{ref:u,style:{position:"absolute",inset:0,overflow:"hidden",backgroundColor:"#1a1a2e",cursor:U?"default":"none"},children:[m&&L.jsx(Mm,{}),S&&L.jsxs("div",{style:{position:"absolute",inset:0,display:"flex",alignItems:"center",justifyContent:"center",backgroundColor:"rgba(0, 0, 0, 0.8)",color:"#ef4444",fontFamily:"system-ui, sans-serif",fontSize:14,zIndex:10,padding:24,textAlign:"center"},children:["Failed to load replay: ",S]}),!m&&!S&&L.jsxs(L.Fragment,{children:[L.jsx(Sm,{onClick:()=>X(z=>!z),visible:U&&!K}),K&&L.jsx(Fm,{replayId:s,duration:x.totalDurationMs,playerCount:w.length,playerNames:w.map(z=>z.playerName),onClose:()=>X(!1),players:w,currentTimeMs:x.currentTimeMs,totalDurationMs:x.totalDurationMs,onLabelSubmit:c,labelApiUrl:l,labeled:h}),L.jsx(ym,{players:w,followingPlayer:x.followingPlayer,expandedPlayer:$,onFollowPlayer:J,onExpandPlayer:q,currentTimeMs:x.currentTimeMs,visible:U,selectedPlayerState:M,atlas:O}),L.jsx(om,{state:x,onPlay:ve,onPause:De,onSeek:xe,onRateChange:G,visible:U})]})]})};export{Vm as $};
