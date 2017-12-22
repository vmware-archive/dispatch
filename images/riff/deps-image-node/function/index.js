"use strict";

const util = require('util');

let logs = [];

function print(str, ...args) {
    logs.push(util.format(str, ...args));
}

const console_info = console.info;
const console_warn = console.warn;
const console_error = console.error;
const console_log = console.log;

function patchLog() {
    console.info = print;
    console.warn = print;
    console.error = print;
    console.log = print;
}

function unpatchLog() {
    console.info = console_info;
    console.warn = console_warn;
    console.error = console_error;
    console.log = console_log;
}

let func = require('./func');

module.exports = ({context, payload}) => {
    logs = [];

    let r = null;
    try {
        patchLog();
        r = func(context, payload);
    } catch (e) {
        print(e.stack);
    } finally {
        unpatchLog();
    }

    return {context: {logs: logs}, payload: r};
};
