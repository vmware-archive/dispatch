---
layout: default
title:  Harden VMs in vSphere
---

# Introduction

This tutorial will walk you through executing functions in response to vSphere events.
_Note:_ Ensure you have followed the [Quickstart]({{ '/documentation/front/quickstart' | relative_url }}) guide and have a working Dispatch Virtual Appliance.

Our goal is to execute a PowerShell function called hardenVM, which will apply additional hardening configuration on every virtual machine created in vSphere.

# Create Seed Images

Dispatch is bundled with a set of seed images for multiple languages to get you started easily. If you followed the guide, you should
have a set of images created in your Dispatch VM. Execute the following command to check:

```bash
$ dispatch get images
     NAME    |   URL                |    BASEIMAGE    | STATUS | CREATED DATE
-----------------------------------------------------------------------------
  java       | dispatch/f23f029e... | java-base       | READY  | ...
  nodejs     | dispatch/6f04f67d... | nodejs-base     | READY  | ...
  powershell | dispatch/edcbdda8... | powershell-base | READY  | ...
  python3    | dispatch/1937b329... | python3-base    | READY  | ...
```

If the list is empty, you can create the seed images using the following command:

```bash
$ dispatch create seed-images
Created BaseImage: nodejs-base
Created BaseImage: python3-base
Created BaseImage: powershell-base
Created BaseImage: java-base
Created Image: nodejs
Created Image: python3
Created Image: powershell
Created Image: java
```

**Note**: You need to wait a short while for the images to be in the `READY` state. You can always check the status using the `dispatch get images` command.

# Create a PowerShell image with PowerCLI dependency

Our function needs a library that knows how to talk to vSphere. We will use `PowerCLI` SDK to do that. To bring the dependency for our function, we will prepare a dedicated image.

1. Create a file named `deps.ps1` with PowerCLI dependency:

```bash
cat << EOF > deps.psd1
@{
  'VMware.PowerCLI' = 'latest'
}
EOF
```

2. create a new image using PowerShell base image:

```
$ dispatch create image powershell-powercli powershell-base --runtime-deps deps.psd1
Created image: powershell-powercli
```

# Create a secret with vSphere credentials

Our function also needs to know how to connect to vSphere, and what credentials to use. In order to do that, we will create a Dispatch secret.

1. Create a JSON file with secret contents:

```bash
cat << EOF > vsphere.json
{
  "host": "<vCenter host URL, e.g. https://vcenter.example.com>",
  "username": "<username>",
  "password": "<password>"
  "vcenterurl": "<full vCenter URL, e.g. username:password@vcenter.example.com:443>"
}
EOF
```

2. Create a dispatch secret called `vsphere`:
```
$ dispatch create secret vsphere vsphere.json
Created secret: vsphere
```

# Create a function

We will use [`ApplyHardening` function](https://github.com/vmware/PowerCLI-Example-Scripts/blob/cb9e57e185a268559747fdb679299499dea816b5/Modules/apply-hardening/apply-hardening.psm1)
from PowerCLI-Example-Scripts repo. We slightly modify it to include the entrypoint handler (comments  removed for brevity.

1. Create function file with following contents:
```powershell
Import-Module PowerCLI.ViCore

function Apply-Hardening {    
    [CmdletBinding()]
    param( 
        [Parameter(Mandatory=$true,
        ValueFromPipeline=$True,
                    Position=0)]
        [VMware.VimAutomation.ViCore.Impl.V1.Inventory.InventoryItemImpl[]]
        $VMs
    ) 
    Process { 
        $ExtraOptions = @{
            "isolation.tools.diskShrink.disable"="true";
            "isolation.tools.diskWiper.disable"="true";
            "isolation.tools.copy.disable"="true";
            "isolation.tools.paste.disable"="true";
            "isolation.tools.dnd.disable"="true";
            "isolation.tools.setGUIOptions.enable"="false"; 
            "log.keepOld"="10";
            "log.rotateSize"="100000"
            "RemoteDisplay.maxConnections"="2";
            "RemoteDisplay.vnc.enabled"="false";  
        
        }
        if ($DebugPreference -eq "Inquire") {
            Write-Output "VM Hardening Options:"
            $ExtraOptions | Format-Table -AutoSize
        }
        
        $VMConfigSpec = New-Object VMware.Vim.VirtualMachineConfigSpec
        
        Foreach ($Option in $ExtraOptions.GetEnumerator()) {
            $OptionValue = New-Object VMware.Vim.optionvalue
            $OptionValue.Key = $Option.Key
            $OptionValue.Value = $Option.Value
            $VMConfigSpec.extraconfig += $OptionValue
        }
        ForEach ($VM in $VMs){
                $VMv = Get-VM $VM | Get-View
            $state = $VMv.Summary.Runtime.PowerState
            Write-Output "...Starting Reconfiguring VM: $VM "
            $TaskConf = ($VMv).ReconfigVM_Task($VMConfigSpec)
                if ($state -eq "poweredOn") {
                    Write-Output "...Migrating VM: $VM "
                    $TaskMig = $VMv.MigrateVM_Task($null, $_.Runtime.Host, 'highPriority', $null)
                    }
            }
        }
}
    
function handle($context, $payload) {
    [void](Set-PowerCLIConfiguration -InvalidCertificateAction ignore -Confirm:$false)

    $username = $context.secrets.username
    $password = $context.secrets.password
    $hostname = $context.secrets.host
    $vmName = $payload.metadata.vm_name

    # Connect to vSphere
    Write-Host "Checking VC Connection is active"
    if (-not $global:defaultviservers) {
        Write-Host "Connecting to $hostname"
        $server = connect-viserver -server $hostname -User $username -Password $password
    } else {
        Write-Host "Already connected to $hostname"
    }

    Write-Host "Get Virtual Machine By name"
    $vm = Get-VM -Name $vmName

    Write-Host "Security Harden our VM"
    $vm | Apply-Hardening 
    
    return "success"
}
```

2. Create Dispatch function using previously created image and secret (assuming the function file is called hardenvm.ps1):
```
$ dispatch create function harden-vm hardenvm.ps1 --image=powershell-powercli --secret vsphere
Created function: harden-vm
```

If you have a VM in your vSphere already, you can test the function before wiring it to vSphere events by running it manually:

```
$ dispatch exec harden-vm --wait --input '{"metadata": {"vm_name": "myvm"}}' | jq -r '.output'
[
  "...Starting Reconfiguring VM: myvm ",
  "success"
]
``` 

Replace `myvm` with the name of your vm. 

# Create an event driver
In order to receive vSphere events we need to create an event driver.

1. Create a vcenter event driver type which registers the docker image implementing the driver:
```
$ dispatch create event-driver-type vcenter dispatchframework/dispatch-events-vcenter:solo-auth
Created event driver type: vcenter
```

2. Create the driver using previously created secret:
```
$ dispatch create event-driver vcenter --secret vsphere --name vcenter
Created event driver: vcenter
```

# Wiring function and event - subscription

The final step is to wire the function and event together. to do that, we create a subscription.  We will use vsphere event
`VMDeployedEvent`, which is emitted when VM finishes deploying:

```
$ dispatch create subscription --event-type vm.deployed harden-vm --name harden_deployed
created subscription: harden_deployed
```

Now, when you create a VM in vSphere, after a while you should see that the `harden-vm` function is executed, and in vSphere UI
yo can see that VM has now extra configured applied to it!

![Hardening options applied to the vm.]({{ '/assets/images/harden-options.png' | relative_url }}){:width="600px"}



 
