ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG SYSTEM_PACKAGES_FILE=system-packages.txt
COPY ${SYSTEM_PACKAGES_FILE} .

RUN packages=$(cat ${SYSTEM_PACKAGES_FILE} | sed 's/ /-/' | uniq | paste -d' ' -); \
    [ -n "${packages}" ] && tdnf install -y ${packages} || :

ARG PACKAGES_FILE=packages.txt
COPY ${PACKAGES_FILE} package.json

RUN pj=$(cat package.json); [ -n "${pj}" ] && npm install || :
