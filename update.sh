#!/bin/bash
set -e
set -u

TIME=`date +"%Y-%m-%d"`

git pull
git add .
git commit -m "$TIME"
git push

