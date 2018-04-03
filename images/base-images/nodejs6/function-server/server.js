/*
 * Copyright (c) 2018 VMware, Inc. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const {FUNCTION_MODULE, PORT} = process.env;

const util = require('util');

function printTo(logs) {
    return (str, ...args) => {
        logs.push(util.format(str, ...args));
    }
}

function patchLog(logs) {
    console.info = printTo(logs);
    console.warn = printTo(logs);
    console.error = printTo(logs);
    console.log = printTo(logs);
}

function wrap(f) {
    return async ({context, payload}) => {
        let [logs, r, err] = [[], null, null];
        try {
            patchLog(logs);
            r = await f(context, payload);
        } catch (e) {
            printTo(e.stack);
            err = e;
        }
        return {context: {logs: logs, error: err}, payload: r}
    }
}

const fun = wrap(require(FUNCTION_MODULE));
const createApp = require('./http-api');

const app = createApp(fun);
app.listen(PORT);

console.log("Function Runtime API started");

process.on('SIGTERM', process.exit);
process.on('SIGINT', process.exit);
