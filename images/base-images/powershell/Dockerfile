### Downloader image: Download & install Power{CLI,NSX,vRA}
FROM vmware/photon:2.0 as downloader

RUN tdnf install -y unzip

WORKDIR /powershell
ARG POWERCLI_PACKAGE=PowerCLI.ViCore.zip
ARG POWERCLI_VDS_PACKAGE=PowerCLI.Vds.zip
ARG POWERCLI_CIS_PACKAGE=PowerCLI.Cis.zip

# Add PowerCLI
ADD https://download3.vmware.com/software/vmw-tools/powerclicore/PowerCLI_Core.zip /powershell
RUN unzip /powershell/PowerCLI_Core.zip -d /powershell && \
    mkdir -p /root/.local/share/powershell/Modules && \
    unzip /powershell/$POWERCLI_PACKAGE -d /root/.local/share/powershell/Modules && \
    unzip /powershell/$POWERCLI_VDS_PACKAGE -d /root/.local/share/powershell/Modules && \
    unzip /powershell/$POWERCLI_CIS_PACKAGE -d /root/.local/share/powershell/Modules

# Add PowerNSX
ADD https://raw.githubusercontent.com/vmware/powernsx/master/module/platform/core/PowerNSX/PowerNSX.psd1 /root/.local/share/powershell/Modules/PowerNSX/
ADD https://raw.githubusercontent.com/vmware/powernsx/master/module/platform/core/PowerNSX/PowerNSX.psm1 /root/.local/share/powershell/Modules/PowerNSX/

# Add PowervRA
ADD https://github.com/jakkulabs/PowervRA/archive/v2.2.0.zip /powershell
RUN unzip /powershell/v2.2.0.zip -d /powershell
RUN mv /powershell/PowervRA-2.2.0/PowervRA ~/.local/share/powershell/Modules/
RUN rm -rf /powershell/PowervRA-2.2.0 v2.2.0.zip

### Actual image
FROM vmware/photon:2.0

# Add PowerShell repository location to Photon OS
RUN echo $'[powershell]\n\
name=VMware PowerShell Repo\n\
baseurl=https://vmware.bintray.com/powershell\n\
gpgcheck=0\n\
enabled=1\n\
skip_if_unavailable=True\n '\
>> /etc/yum.repos.d/powershell.repo && \
 # Install PowerShell on Photon
 tdnf install -y util-linux-libs powershell curl openssl

COPY --from=downloader /root/.local/share/powershell/Modules /root/.local/share/powershell/Modules

WORKDIR /root/