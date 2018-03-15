FROM microsoft/powershell:ubuntu16.04

LABEL authors="renoufa@vmware.com,jaker@vmware.com"

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y unzip && apt-get clean

# Set working directory so stuff doesn't end up in /
WORKDIR /root

# Install VMware modules from PSGallery
SHELL [ "pwsh", "-command" ]
RUN Set-PSRepository -Name PSGallery -InstallationPolicy Trusted
RUN Install-Module VMware.PowerCLI,PowerNSX,PowervRA

# Add the PowerCLI Example Scripts and Modules
# using ZIP instead of a git pull to save at least 100MB
SHELL [ "bash", "-c"]
RUN curl -o ./PowerCLI-Example-Scripts.zip -J -L https://github.com/vmware/PowerCLI-Example-Scripts/archive/master.zip && \
    unzip PowerCLI-Example-Scripts.zip && \
    rm -f PowerCLI-Example-Scripts.zip && \
    mv ./PowerCLI-Example-Scripts-master ./PowerCLI-Example-Scripts && \
    mv ./PowerCLI-Example-Scripts/Modules/* /usr/local/share/powershell/Modules/

CMD ["/usr/bin/pwsh"]