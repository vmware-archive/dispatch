"use strict";

let getStdin = require('get-stdin');

let func = require('./function/func');

let handler = (input, callback) => {
    let ctxAndIn = JSON.parse(input);
    callback(undefined, func(ctxAndIn.context, ctxAndIn.input));
};

getStdin().then(input => {
    handler(input, (err, output) => {
        if (err) {
            return console.error(err);
        }
        console.log(JSON.stringify(output));
    });
}).catch(e => {
    console.error(e.stack);
});
