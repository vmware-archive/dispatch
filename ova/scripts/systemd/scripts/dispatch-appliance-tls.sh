#!/usr/bin/env bash
# Copyright 2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -euf -o pipefail

umask 077

keytool="/usr/bin/keytool"

cert_dir="/storage/data/certs"

ca_cert="${cert_dir}/ca.crt"
ca_key="${cert_dir}/ca.key"

csr="${cert_dir}/server.csr"
ext="${cert_dir}/extfile.cnf"
cert="${cert_dir}/server.crt"
key="${cert_dir}/server.key"

flag="${cert_dir}/cert_gen_type"
jks="${cert_dir}/trustedcertificates.jks"

# From dispatch-appliance-environment
tls_cert="${APPLIANCE_TLS_CERT}"
tls_private_key="${APPLIANCE_TLS_PRIVATE_KEY}"
tls_ca_cert="${APPLIANCE_TLS_CA_CERT}"

# Format cert file
function formatCert {
  content=$1
  file=$2
  echo "$content" | sed -r 's/(-{5}BEGIN [A-Z ]+-{5})/&\n/g; s/(-{5}END [A-Z ]+-{5})/\n&\n/g' | sed -r 's/.{64}/&\n/g; /^\s*$/d' | sed -r '/^$/d' > "$file"
}

# Check if private key is valid
function checkKey {
  file=$1
  openssl rsa -in "$file" -check
}

# Generate self signed cert
function genCert {
  echo "Generating self signed certificate"
  if [ ! -e $ca_cert ] || [ ! -e $ca_key ]
  then
    openssl req -newkey rsa:4096 -nodes -sha256 -keyout $ca_key \
      -x509 -days 1095 -out $ca_cert -subj \
      "/C=US/ST=California/L=Palo Alto/O=VMware, Inc./OU=Dispatch/CN=Self-signed by VMware, Inc."
  fi
  openssl req -newkey rsa:4096 -nodes -sha256 -keyout $key \
    -out $csr -subj \
    "/C=US/ST=California/L=Palo Alto/O=VMware/OU=Dispatch/CN=${HOSTNAME}"

  if [ -n "${HOSTNAME}" ] && [ "${HOSTNAME}" != "${IP_ADDRESS}" ]; then
    san="subjectAltName = DNS:${HOSTNAME},IP:${IP_ADDRESS}"
  else
    san="subjectAltName = IP:${IP_ADDRESS}"
  fi
  echo "Add subjectAltName $san to certificate"
  echo "$san" > $ext

  openssl x509 -req -days 1095 -in $csr -CA $ca_cert -CAkey $ca_key -CAcreateserial -extfile $ext -out $cert

  echo "Creating certificate chain for $cert"
  cat $ca_cert >> $cert

  echo "self-signed" > $flag
}

function secure {
  if [ -n "$tls_private_key" ] && [ "$tls_private_key" == "*ENCRYPTED*" ]; then
    echo "Private key is encrypted - will generate a self-signed certificate"
    genCert
    return
  fi

  if [ -n "$tls_cert" ] && [ -n "$tls_private_key" ] && [ -n "$tls_ca_cert" ]; then
    echo "TLS certificate, private key, and CA certificate are set, using customized certificate"
    formatCert "$tls_cert" $cert
    formatCert "$tls_private_key" $key
    formatCert "$tls_ca_cert" $ca_cert

    ensurePKCS8Format

    echo "customized" > $flag

    return
  fi

  if [ ! -e $cert ] || [ ! -e $key ] || [ ! -e $ca_cert ]; then
    echo "TLS certificate, private key, or CA certificate file does not exist - will generate a self-signed certificate"
    genCert
    return
  fi

  if [ ! -e $flag ]; then
    echo "The file which records the certificate source (user provided or self generated) does not exist - will generate a new self-signed certificate"
    genCert
    return
  fi

  if [ ! "$(cat $flag)" = "self-signed" ]; then
    echo "The certificate source (user provided or self generated) changed - will generate a new self-signed certificate"
    genCert
    return
  fi

  cn=$(openssl x509 -noout -subject -in $cert | sed -n '/^subject/s/^.*CN=//p') || true
  if [ "${HOSTNAME}" !=  "$cn" ]; then
    echo "Common name changed: $cn -> ${HOSTNAME} - will generate a new self-signed certificate"
    genCert
    return
  fi

  ip_in_cert=$(openssl x509 -noout -text -in $cert | sed -n '/IP Address:/s/.*IP Address://p') || true
  if [ "${IP_ADDRESS}" !=  "$ip_in_cert" ]; then
    echo "IP changed: $ip_in_cert -> ${IP_ADDRESS} - will generate a new self-signed certificate"
    genCert
    return
  fi

  echo "Using the existing CA, certificate and key file"
}

# Check if private key is in PKCS8 format, convert if possible
function ensurePKCS8Format() {
  local header=""
  header=$(head -n 1 $key)
  if [ "$header" == "-----BEGIN PRIVATE KEY-----" ]; then
    echo "Private key $key is in PKCS8 format"
    return
  fi
  echo "Private key $key is not in PKCS8 format"

  cp $key $key.tmp
  rm $key

  openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in $key.tmp -out $key
  if [ $? -ne 0 ]; then
    echo "ERROR Failed to convert $key to PKCS8"
    mv $key.tmp $key
    return
  fi
  # Override value from environment file
  tls_private_key="$(cat $key)"
  echo "Converted $key to PKCS8"

  # Cleanup
  rm -f $key.tmp
}

# Warn if expiration in less than 60 days
function checkCertExpiration() {
  local certDate=""
  certDate=$(openssl x509 -noout -enddate -in $cert | cut -d = -f2 | xargs -i date -d {} +%s)
  warnDate=$(date -d "+60 day" +%s)
  if [ "$certDate" -lt "$warnDate" ]; then
    echo "WARNING certificate expires is less than 60 days"
  fi
  echo "Certificate expiration date OK"
}

# File permissions
mkdir -p ${cert_dir}
chown -R "${APPLIANCE_SERVICE_UID}":"${APPLIANCE_SERVICE_UID}" ${cert_dir}

# Init certs
secure
checkCertExpiration

# File permissions - components can use shared TLS cert
chown "${APPLIANCE_SERVICE_UID}":"${APPLIANCE_SERVICE_UID}" ${ca_cert}
chown "${APPLIANCE_SERVICE_UID}":"${APPLIANCE_SERVICE_UID}" ${key}
chown "${APPLIANCE_SERVICE_UID}":"${APPLIANCE_SERVICE_UID}" ${cert}

# Log
if [ -f ${cert} ]; then
  cert_fp="$(openssl x509 -in ${cert} -noout -sha256 -fingerprint)"
  echo "${cert} fingerprint: ${cert_fp}"
  cert_fp="$(openssl x509 -in ${cert} -noout -sha1 -fingerprint)"
  echo "${cert} fingerprint: ${cert_fp}"
fi
if [ -f ${ca_cert} ]; then
  ca_cert_fp="$(openssl x509 -in ${ca_cert} -noout -sha256 -fingerprint)"
  echo "${ca_cert} fingerprint: ${ca_cert_fp}"
  ca_cert_fp="$(openssl x509 -in ${ca_cert} -noout -sha1 -fingerprint)"
  echo "${ca_cert} fingerprint: ${ca_cert_fp}"
fi

echo "Finished dispatch-appliance-tls config"
