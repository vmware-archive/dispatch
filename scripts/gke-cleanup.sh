#!/usr/bin/env bash
# Script to clean up orphaned resources from GKE cluster. It deletes orphaned volumes and Load Balancers.
# You must have your project configured in gcloud CLI to use this script.
# Usage:
# ./gke-script.sh - by default it will ask before deleting
# Flags:
# -f,--force - delete without confirmation
# -v,--volume-name-part - string to use when filtering volume names (default: dispatch-ci)
# -r,--region - GCE region to use (default: us-west1)
#
# Notes:
# * Orphaned volumes are volumes not attached to any virtual machine. to avoid false positives (volumes purposely
# detached), additional filtering is applied. By default, volume name must also include "dispatch-ci" string.
# * Orphaned Load Balancers are detected using their health checks. If a target pool has 0 healthy nodes, is considered
# orphaned and subject to deletion. Related objects: firewall rules, forwarding rules and http health checks are also
# detected and deleted.

set -e -o pipefail

## Default values for flag variables
: ${FORCE:=NO}
: ${VOLUME_NAME_PART:="dispatch-ci"}
: ${REGION:="us-west1"}
## End Default values

## HELPER FUNCTIONS
# ask asks user for confirmation, returns success if accepted, fails otherwise
ask() {
    local prompt default reply
    while true; do

        if [ "${2:-}" = "Y" ]; then
            prompt="Y/n"
            default=Y
        elif [ "${2:-}" = "N" ]; then
            prompt="y/N"
            default=N
        else
            prompt="y/n"
            default=
        fi

        # Ask the question (not using "read -p" as it uses stderr not stdout)
        echo -n "$1 [${prompt}] "

        # Read the answer (use /dev/tty in case stdin is redirected from somewhere else)
        read reply </dev/tty

        # Default?
        if [ -z "$reply" ]; then
            reply=${default}
        fi

        # Check if the reply is valid
        case "${reply}" in
            Y*|y*) return 0 ;;
            N*|n*) return 1 ;;
        esac
    done
}

# join_by uses character specified as first argument to join the rest of arguments, e.g.
# join_by "," "1" "2" "3" produces "1,2,3"
join_by() {
    local IFS="$1";
    shift;
    echo "$*";
}

# force returns success if FORCE flag was set.
force() {
    if [[ ${FORCE} == "YES" ]]; then
        return 0;
    else
        return 1;
    fi
}
## END HELPER FUNCTIONS

## FLAG PARSING
POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case ${key} in
    -f|--force)
    FORCE=YES
    shift # past argument
    ;;
    -v|--volume-name-part)
    VOLUME_NAME_PART="$2"
    shift # past argument
    shift # past value
    ;;
    -r|--region)
    REGION="$2"
    shift # past argument
    shift # past value
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters
## END FLAG PARSING ##

## VOLUMES ##
# list names of all volumes with no attachments and name matching VOLUME_NAME_PART string
VOLUMES_TO_DELETE=$(gcloud compute disks list --filter "name~${VOLUME_NAME_PART} AND users=null" --format "value(name)")

if [[ -n ${VOLUMES_TO_DELETE} ]]; then
    echo "Volumes to delete: "
    echo ${VOLUMES_TO_DELETE} | tr ' ' '\n'

    if force || ask "Are you sure you want to delete the volumes" "N"; then
        echo "Deleting volumes..."
        gcloud compute disks delete -q $(echo ${VOLUMES_TO_DELETE} | tr '\n' ' ')
    else
        echo "Aborting..."
       # exit 1
    fi
else
    echo "No volumes to delete, skipping..."
fi
## END VOLUMES ##

## LOAD BALANCERS ##
# Load balancers actually consist of 4 different objects:
# * firewall rules
# * forwarding rules (external IP)
# * target pools
# * http health checks
# They must be deleted in order specified above to avoid dependency conflicts.

# We collect resources into the arrays and delete them at the end all together to reduce the number to DELETE API calls.
HEALTH_CHECKS_TO_DELETE=()
FIREWALL_RULES_TO_DELETE=()
FORWARDING_RULES_TO_DELETE=()
TARGET_POOLS_TO_DELETE=()

# iterate over target pools. Only using target pool as starting point we can find other dependent resources.
for target_pool in $(gcloud compute target-pools list --filter "region:(${REGION})" --format "value(name)" | tr '\n' ' ');
do
    # Get target pool health. Returns "[]" for unhealthy cluster. We must add || true as the command returns 1
    # when cluster is unhealthy
    health=$(gcloud compute target-pools get-health --region ${REGION} ${target_pool} --format json 2>/dev/null || true);
    if [[ ${health} != '[]' ]];
    then
        echo "${target_pool} is a healthy target-pool, skipping"
        continue
    fi
    echo "${target_pool} is a candidate for deletion, retrieving dependent resources..."

    # Get related firewall rule(s)
    firewall_rules=$(gcloud compute firewall-rules list --filter "name~${target_pool}" --format "value(name)")
    if [[ -n ${firewall_rules} ]];
    then
        echo "Found related firewall rule(s) ${firewall_rules}, adding for later deletion"
        FIREWALL_RULES_TO_DELETE+=(${firewall_rules})
    else
        echo "No related firewall rule(s) found"
    fi

    # Get related forwarding rule(s)
    forwarding_rules=$(gcloud compute forwarding-rules list --filter "target~${target_pool} AND region:${REGION}" --format "value(name)")
    if [[ -n ${forwarding_rules} ]];
    then
        echo "Found related forwarding rule(s) ${forwarding_rules}, adding for later deletion"
        FORWARDING_RULES_TO_DELETE+=(${forwarding_rules})
    else
        echo "No related firewall rule(s) found"
    fi

    TARGET_POOLS_TO_DELETE+=(${target_pool})

    # Get related health checks
    health_check=$(gcloud compute target-pools describe --region ${REGION} ${target_pool} --format "value(healthChecks,0)")

    if [[ -n ${health_check} ]];
    then
        # Strip the health check URL, leave only the name, e.g.
        # "https://www.googleapis.com/compute/v1/projects/project-id/global/httpHealthChecks/k8s-b075c3c1b7e5a05f-node"
        # becomes "k8s-b075c3c1b7e5a05f-node"
        # also remove the whitespaces.
        health_check=$(echo ${health_check##*/} | sed 's/ *$//g')
        echo "Found a related health check ${health_check}, adding for later deletion"
        HEALTH_CHECKS_TO_DELETE+=(${health_check})
    else
        echo "No related health checks found"
    fi
done

# We've collected all dependent resources, now we need to delete them in proper order.

# Delete the firewall rules, if any
if [[ ${#FIREWALL_RULES_TO_DELETE[@]} > 0 ]];
then
    echo "Following firewall rules will be deleted:"
    echo ${FIREWALL_RULES_TO_DELETE[@]}
    if force || ask "Are you sure you want to delete these firewall rules:" "N";
    then
        echo "Deleting firewall rules..."
        gcloud compute firewall-rules delete -q $(join_by ' ' "${FIREWALL_RULES_TO_DELETE[@]}")
    else
        echo "Aborting..."
        exit 1
    fi
else
    echo "No firewall rules to delete"
fi

# Delete the forwarding rules, if any
if [[ ${#FORWARDING_RULES_TO_DELETE[@]} > 0 ]];
then
    echo "Following forwarding rules will be deleted:"
    echo ${FORWARDING_RULES_TO_DELETE[@]}
    if force || ask "Are you sure you want to delete these forwarding rules:" "N";
    then
        echo "Deleting forwarding rules..."
        gcloud compute forwarding-rules delete --region ${REGION} -q $(join_by ' ' "${FORWARDING_RULES_TO_DELETE[@]}")
    else
        echo "Aborting..."
        exit 1
    fi
else
    echo "No forwarding rules to delete"
fi

# Delete the target pools, if any
if [[ ${#TARGET_POOLS_TO_DELETE[@]} > 0 ]];
then
    echo "Following target pools are going to be deleted:"
    echo ${TARGET_POOLS_TO_DELETE[@]}
    if force || ask "Are you sure you want to delete these target pools:" "N";
    then
       echo "Deleting target pools..."
       gcloud compute target-pools delete --region ${REGION}  -q $(join_by ' ' "${TARGET_POOLS_TO_DELETE[@]}")
    else
        echo "Aborting..."
        exit 1
    fi
else
    echo "No target pools to delete"
fi

# Delete the health check(s), if any
if [[ ${#HEALTH_CHECKS_TO_DELETE[@]} > 0 ]];
then
    echo "Following health checks are going to be deleted:"
    echo ${HEALTH_CHECKS_TO_DELETE[@]}
    if force || ask "Are you sure you want to delete these health checks:" "N";
    then
       echo "Deleting health checks..."
       gcloud compute http-health-checks delete -q $(join_by ' ' "${HEALTH_CHECKS_TO_DELETE[@]}")
    else
        echo "Aborting..."
        exit 1
    fi
else
    echo "No health checks to delete"
fi

## END LOAD BALANCERS ##




