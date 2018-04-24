#!/usr/bin/env bats

set -o pipefail

load helpers
load variables

@test "Setup service catalog" {
    run helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
    echo_to_log
    assert_success

    run helm install svc-cat/catalog --name catalog --namespace ${DISPATCH_NAMESPACE}-services --wait
    echo_to_log
    assert_success

    run helm install svc-cat/ups-broker --name ups-broker --namespace ${DISPATCH_NAMESPACE}-services --wait
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
}

@test "List service classes" {
    # Give the service catalog a chance to sync with the broker
    run_with_retry "dispatch get serviceclasses user-provided-service-with-schemas --json | jq -r .status" "READY" 6 10
}

@test "Create service instance" {
    run dispatch create serviceinstance ups-with-schema user-provided-service-with-schemas default --params '{"param-1": "foo", "param-2": "bar"}'
    echo_to_log
    assert_success

    run_with_retry "dispatch get serviceinstances ups-with-schema --json | jq -r .status" "READY" 6 10
    run_with_retry "dispatch get serviceinstances ups-with-schema --json | jq -r .binding.status" "READY" 6 10
}

@test "Batch load images" {
    batch_create_images images.yaml
}

@test "Create a function which echos the service context" {
    run dispatch create function nodejs6 node-echo ${DISPATCH_ROOT}/examples/nodejs6/echo.js --service ups-with-schema
    echo_to_log
    assert_success

    run_with_retry "dispatch get function node-echo --json | jq -r .status" "READY" 8 5
}

@test "Execute a function which echos the service context" {
    run_with_retry "dispatch exec node-echo --wait --input '{\"in-1\": \"baz\"}' | jq -r '.output.context.serviceBindings.\"ups-with-schema\".\"special-key-1\"'" "special-value-1" 2 10
    echo_to_log
    assert_success
}

@test "Delete service instance" {
    run dispatch delete serviceinstance ups-with-schema
    echo_to_log
    assert_success
}

@test "Tear down catalog" {
    run helm delete --purge ups-broker
    echo_to_log
    assert_success

    run helm delete --purge catalog
    echo_to_log
    assert_success

    run kubectl delete namespace ${DISPATCH_NAMESPACE}-services
    echo_to_log
    assert_success
}

@test "Cleanup" {
    cleanup
}

