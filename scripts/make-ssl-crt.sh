openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=serverless.vmware.com/O=serverless.vmware.com"

kubectl create secret tls serverless-vmware-tls --key tls.key --cert tls.crt
