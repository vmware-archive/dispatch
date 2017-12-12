FROM kong:0.11.1

# please create api-tls.crt and make sure they are at project root
ADD api-tls.crt /usr/local/kong/tls/tls.crt
ADD api-tls.key /usr/local/kong/tls/tls.key
ADD charts/kong/kong.conf /etc/kong/kong.conf
ADD charts/kong/nginx.conf /nginx.conf
ADD charts/kong/serverless-transformer /usr/local/kong/custom/kong/plugins/serverless-transformer
