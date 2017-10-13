"use strict";

let getStdin = require('get-stdin');

let func = require('./function/func');

let handler = (input, callback) => {
    callback(undefined, func(JSON.parse(input)));
};

getStdin().then(val => {
    handler(val, (err, res) => {
        if (err) {
            return console.error(err);
        }
        if(isArray(res) || isObject(res)) {
            console.log(JSON.stringify(res));
        } else {
            process.stdout.write(res);
        }
    });
}).catch(e => {
    console.error(e.stack);
});

let isArray = (a) => {
    return (!!a) && (a.constructor === Array);
};

let isObject = (a) => {
    return (!!a) && (a.constructor === Object);
};
