Import-Module PowerCLI.ViCore

function handle($context, $payload) {
    [void](Set-PowerCLIConfiguration -InvalidCertificateAction ignore -Confirm:$false)

    $username = $context.secrets.username
    $password = $context.secrets.password
    $hostname = $context.secrets.host
    $vmName = $payload.metadata.vm_name
    $guestUser = $context.secrets.guestUsername
    $guestPassword = $context.secrets.guestPassword

    # Connect to vSphere
    $server = connect-viserver -server $hostname -User $username -Password $password

    # Get Virtual Machine By name
    $vm = Get-VM -Name $vmName

    # Invoke Guest Script
    $output = Invoke-VMScript -VM $vm -GuestUser $guestUser -GuestPassword $guestPassword -ScriptType "Bash" -ScriptText "echo Script executed successfully!"

    # Refresh VM
    $vm = Get-VM -Name $vmName

    # Rename VM by Appending "-ready"
    $newName = $vm.Name + "-ready"
    Set-VM -VM $vm -Name $newName -confirm:$false

    # Retrieve IP Address
    $ipaddress = $vm.guest.IPAddress[0]

    [void](emitEvent $newName $ipaddress)

    return ($output | Select ScriptOutput)
}

function emitEvent($newName, $ipaddress) {
    $payload = @{vm_name=$newName;vm_ip=$ipaddress}
    $postParams = @{topic="vm.ready";payload=$payload}
    $headers = @{"cookie"="Cookie"}
    [void](Invoke-WebRequest -Uri "http://dispatch-event-manager.dispatch/v1/event" -Headers $headers -Method "POST" -ContentType "application/json" -Body (ConvertTo-Json $postParams))
}
