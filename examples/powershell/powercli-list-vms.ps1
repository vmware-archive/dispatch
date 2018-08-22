#######################################################################
## Copyright (c) 2017 VMware, Inc. All Rights Reserved.
## SPDX-License-Identifier: Apache-2.0
#######################################################################

# Import-Module PowerCLI.ViCore

function handle($context, $payload) {
    [void](Set-PowerCLIConfiguration -InvalidCertificateAction ignore -Confirm:$false)

    $username = $context.secrets.username
    $password = $context.secrets.password
    $hostname = $context.secrets.host

    $server = connect-viserver -server $hostname -User $username -Password $password

    return Get-VM -Server $server | Select Name, MemoryGB, NumCpu, @{Name="HostName"; Expression={$_.VMHost.Name}}
}


