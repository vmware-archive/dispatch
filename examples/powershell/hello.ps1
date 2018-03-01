#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################


function handle($context, $payload) {

    $name = $payload.name
    if (!$name) {
        $name = "Noone"
    }
    $place = $payload.place
    if (!$place) {
        $place = "Nowhere"
    }

    return @{myField="Hello, $name from $place"}
}


