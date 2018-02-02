"use strict";

let print = console.log;

console.info = console.warn;
console.log = console.warn;

let getStdin = require('get-stdin');
let func = require('./function/func');

getStdin().then(input => {
    let ctxAndIn = JSON.parse(input);
    Promise.resolve(func(ctxAndIn.context, ctxAndIn.input)).then(obj => {
        print(JSON.stringify(obj))
    }).catch(e => {
        console.error(e)
    })
}).catch(e => {
    console.error(e.stack);
});
