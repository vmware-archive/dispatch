///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////


module.exports = function (context, params) {
    let name = params.name;
    if (context.secrets["password"] === undefined) {
        return {message: "I know nothing"}
    }
    return {message: "The password is " + context.secrets["password"]}
};
