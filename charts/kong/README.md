Usage
==========

Firstly install helm: [docs here](https://helm.sh/)

1. Clone the repository
2. cd ./charts/kong/
3. edit the "values.yaml" doc to fit your environment
3. helm package .
4. helm install kong-0.1.1.tgz --name yourreleasename

You will then see notes on how to connect and use your Kong deployment. 

Todo:
--------

- Fix cassandra deployment issues 