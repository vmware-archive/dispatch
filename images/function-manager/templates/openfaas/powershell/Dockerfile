FROM vmware/dispatch-openfaas-watchdog:revbf667b8 AS watchdog
FROM {{ .DockerURL }}
COPY --from=watchdog /go/src/github.com/openfaas/faas/watchdog/watchdog /usr/bin/fwatchdog

WORKDIR /root/

COPY index.ps1 .

COPY {{ .FunctionFile }} function/handler.ps1

ENV fprocess="powershell -NoLogo -File index.ps1"

HEALTHCHECK --interval=1s CMD [ -e /tmp/.lock ] || exit 1

CMD ["fwatchdog"]