ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG SYSTEM_PACKAGES_FILE=system-packages.txt
COPY ${SYSTEM_PACKAGES_FILE} .

RUN packages=$(cat ${SYSTEM_PACKAGES_FILE} | sed 's/ /-/' | uniq | paste -d' ' -); \
    [ -n "${packages}" ] && tdnf install -y ${packages} || :

ARG PACKAGES_FILE=packages.txt
COPY ${PACKAGES_FILE} /root/requirements.psd1

RUN pj=$(cat requirements.psd1); [ -n "${pj}" ] && \
    pwsh -command '$ErrorActionPreference="Stop"; Invoke-PSDepend -Path "/root/requirements.psd1" -Confirm:$false' || :
