#!/usr/bin/env bash

source ~/.bashisms/s3_upload.bash

set -e

echo "Compiling for linux..."
GOOS=linux GOARCH=amd64 ginkgo build
mv fezzik.test fezzik
echo "Uploading..."
upload_to_s3 fezzik
rm fezzik