#!/usr/bin/env bash

. .env
aws s3 sync --delete plot/ s3://$MEGGIEL_BUCKET_NAME/maryland-farms/
