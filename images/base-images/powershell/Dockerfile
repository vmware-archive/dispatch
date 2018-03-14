FROM microsoft/powershell:ubuntu16.04 as builder

RUN mkdir /powershell

RUN pwsh -command 'Save-Module -Name PowerShellGet -RequiredVersion "1.6.0" -Path /powershell' > /dev/null

# PSDepend dependency manager
RUN pwsh -command 'Save-Module -Name PSDepend -RequiredVersion "0.1.64" -Path /powershell' > /dev/null

# Fix for https://github.com/RamblingCookieMonster/PSDepend/issues/74
RUN mv /powershell/PSDepend/0.1.64/PSDependScripts/Noop.ps1 /powershell/PSDepend/0.1.64/PSDependScripts/noop.ps1

# Using photon image as it has smaller footprint than default microsoft/powershell image
FROM vmware/photon:2.0

RUN tdnf makecache && tdnf install -y powershell && tdnf clean all
COPY --from=builder /powershell/ /root/.local/share/powershell/Modules/

CMD ["pwsh"]