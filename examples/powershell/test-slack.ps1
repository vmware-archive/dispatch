#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################

# Executes Test-SlackApi cmdlet and returns result
function handle($context, $payload) {
    $result=Test-SlackApi
    return @{result=$result}
}


