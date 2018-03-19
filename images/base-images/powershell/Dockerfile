## builder image
FROM microsoft/powershell:ubuntu16.04 as builder

RUN mkdir /powershell

RUN pwsh -command 'Save-Module -Name PowerShellGet -RequiredVersion "1.6.0" -Path /powershell' > /dev/null

# PSDepend dependency manager
RUN pwsh -command 'Save-Module -Name PSDepend -RequiredVersion "0.1.64" -Path /powershell' > /dev/null

# Fix for https://github.com/RamblingCookieMonster/PSDepend/issues/74
RUN mv /powershell/PSDepend/0.1.64/PSDependScripts/Noop.ps1 /powershell/PSDepend/0.1.64/PSDependScripts/noop.ps1


## base image
FROM vmware/photon2:20180302

WORKDIR /root/

RUN tdnf install -y powershell-6.0.1-1.ph2
COPY --from=builder /powershell/ /root/.local/share/powershell/Modules/

ARG IMAGE_TEMPLATE=/image-template
ARG FUNCTION_TEMPLATE=/function-template

LABEL io.dispatchframework.imageTemplate="${IMAGE_TEMPLATE}" \
      io.dispatchframework.functionTemplate="${FUNCTION_TEMPLATE}"

COPY image-template ${IMAGE_TEMPLATE}/
COPY function-template ${FUNCTION_TEMPLATE}/

ENV FUNCTION_MODULE=/root/function/handler.ps1 PORT=8080
EXPOSE ${PORT}

COPY ./function/handler.ps1 ${FUNCTION_MODULE}
COPY ./index.ps1 .

CMD ["pwsh", "-NoLogo", "-File", "index.ps1"]
