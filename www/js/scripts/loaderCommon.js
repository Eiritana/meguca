!function(){function check(func){try{return eval("(function(){"+func+"})();")}catch(e){return!1}}for(var tests=["return (()=>5)()===5;",'"use strict";  const bar = 123; {const bar = 456;} return bar===123;','"use strict"; let bar = 123;{ let bar = 456; }return bar === 123;',"var x='y';return ({ [x]: 1 }).y === 1;","var a=7,b=8,c={a,b};return c.a===7 && c.b===8;",'var a = "ba"; return `foo bar${a + "z"}` === "foo barbaz";',"var arr = [5]; for (var item of arr) return item === 5;","return Math.max(...[1, 2, 3]) === 3",'"use strict"; class C {}; return typeof C === "function"','"use strict"; var passed = false;class B {constructor(a) {  passed = (a === "barbaz")}};class C extends B {constructor(a) {super("bar" + a)}};new C("baz"); return passed;'],legacy,i=0;i<tests.length;i++)if(!check(tests[i])){legacy=!0;break}this.initModuleLoader=function(r){return new Promise(function(r,e){var t={};t["es5/*"]=t["es6/*"]={format:"register"};var a="/lang/en_GB";a=legacy?"es5"+a:"es6"+a,System.config({baseURL:"/ass/js",defaultJSExtensions:!0,map:{lang:a,underscore:"vendor/underscore","js-cookie":"vendor/js-cookie"},meta:t}),legacy?System["import"]("vendor/corejs").then(function(){r(legacy)},function(r){e(r)}):r(legacy)})}}();
//# sourceMappingURL=maps/loaderCommon.js.map