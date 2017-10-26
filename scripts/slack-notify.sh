#! /bin/bash

STATUS=${1}
COLOR=${2} # valid values are good, warning, danger (or a hex value)
PIPELINES_BASE_URL="https://gitlab.eng.vmware.com/serverless/serverless/pipelines"

PAYLOAD="{\
    \"channel\": \"#serverless\",\
    \"username\": \"gitlab-ci\",\
    \"attachments\": [\
        {\
            \"title\": \"[$STATUS] Pipeline $CI_PIPELINE_ID\",\
            \"text\": \"<!here> <$PIPELINES_BASE_URL/$CI_PIPELINE_ID|Pipeline $CI_PIPELINE_ID>\",\
            \"color\": \"$COLOR\"\
        }\
    ]\
}"

curl -X POST --data-urlencode "payload=$PAYLOAD" $SLACK_NOTIFY_WEBHOOK