#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

gcloud compute --project="${project}" instances delete ${user}-wks-{1,2,3,4,5,6} \
  --zone="${zone}" \
  --quiet
