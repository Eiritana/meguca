"use strict";System.register([],function(_export,_context){return {setters:[],execute:function(){const message={disconnect:0};_export("message",message);function fetchJSON(url){return fetch("api/"+url).then(res => res.json());}_export("fetchJSON",fetchJSON);function randomID(len){let id='';for(let i=0;i<len;i++){let char=(Math.random()*36).toString(36)[0];if(Math.random()<0.5){char=char.toUpperCase();}id+=char;}return id;}_export("randomID",randomID);class WeakSetMap{constructor(){this.map=new Map();}add(key,item){if(!this.map.has(key)){this.map.set(key,new WeakSet());}this.map.get(key).add(item);}remove(key,item){const set=this.map.get(key);if(!set){return;}set.delete(item);if(set.size===0){this.map.delete(key);}}forEach(key,fn){const set=this.map.get(key);if(!set){return;}set.forEach(fn);}}_export("WeakSetMap",WeakSetMap);}};});
//# sourceMappingURL=maps/common.js.map