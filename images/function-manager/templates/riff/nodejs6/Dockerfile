FROM projectriff/node-function-invoker:0.0.3 AS riff
FROM {{ .DockerURL }}

WORKDIR /usr/app
COPY --from=riff /usr/app/ .

# Function
RUN mkdir function
COPY index.js .
COPY {{ .FunctionFile }} function/func.js
ENV FUNCTION_URI ./index.js

EXPOSE 8080
CMD ["node", "server.js"]
