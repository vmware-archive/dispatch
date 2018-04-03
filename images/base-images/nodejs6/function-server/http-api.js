/*
 * Copyright (c) 2018 VMware, Inc. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const express = require('express');
const bodyParser = require('body-parser');

module.exports = (fun) => {
    const app = express();

    app.get('/healthz', (req, res) => {
        res.json({});
    });

    app.use(/.*/, bodyParser.json({strict: false}));

    app.post(/.*/, async (req, res) => {
        res.json(await fun(req.body));
    });

    return app;
};
