#!/bin/bash

. .env

#yarn build
AWS_PROFILE=unimeg_meggie aws s3 sync build/ s3://$MEGGIEL_BUCKET_NAME/maryland-farms/
