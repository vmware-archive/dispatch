#!/usr/bin/env bats

set -o pipefail

load ${E2E_TEST_ROOT}/helpers.bash
load ${E2E_TEST_ROOT}/variables.bash

@test "Setup service catalog" {

    run helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
    echo_to_log
    assert_success

    run helm upgrade -i catalog svc-cat/catalog --namespace ${DISPATCH_NAMESPACE}-services --wait
    echo_to_log
    assert_success

    run helm upgrade -i ups-broker svc-cat/ups-broker --namespace ${DISPATCH_NAMESPACE}-services --wait
    echo_to_log
    assert_success
}


@test "Register ups broker" {


cat << EOF > ups-broker.yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
    name: ups-broker
spec:
    url: http://ups-broker-ups-broker.${DISPATCH_NAMESPACE}-services.svc.cluster.local
EOF

    # Seems the service isn't quite ready when it says it is
    retry_simple "kubectl create -f ups-broker.yaml" 30 6
    echo_to_log
    assert_success

    run_with_retry "kubectl get clusterserviceclasses -o json | jq '.items | length'" 3 12 10
}

@test "List service classes" {

    # Give the service catalog a chance to sync with the broker
    run_with_retry "dispatch get serviceclasses user-provided-service-with-schemas --json | jq -r .status" "READY" 6 10
}

@test "Create service instance" {

    run dispatch create serviceinstance ups-with-schema user-provided-service-with-schemas default --params '{"param-1": "foo", "param-2": "bar"}'
    echo_to_log
    assert_success

    run_with_retry "dispatch get serviceinstances ups-with-schema --json | jq -r .status" "READY" 12 10
    run_with_retry "dispatch get serviceinstances ups-with-schema --json | jq -r .binding.status" "READY" 12 10
    run dispatch get serviceinstances ups-with-schema --json
    echo_to_log
    assert_success

    run dispatch get secrets --json
    echo_to_log
    assert_success
}

@test "Create a function which echos the service context" {

    run dispatch create function --image=nodejs node-echo-service ${DISPATCH_ROOT}/examples/nodejs --handler=./echo.js --service ups-with-schema
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-echo-service --json | jq -r .status" "READY" 8 5
}