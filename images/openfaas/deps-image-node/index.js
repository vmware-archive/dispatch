"use strict";

let print = console.log;

console.info = console.warn;
console.log = console.warn;

let getStdin = require('get-stdin');
let func = require('./function/func');

getStdin().then(input => {
    let ctxAndIn = JSON.parse(input);
    print(JSON.stringify(func(ctxAndIn.context, ctxAndIn.input)));
}).catch(e => {
    console.error(e.stack);
});
