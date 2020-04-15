#!/bin/bash
set -x

cd $GOPATH/src/gowiki-upstream
git pull
rsync -r --exclude 'data' $GOPATH/src/gowiki-upstream/ $GOPATH/src/gowiki/
cd $GOPATH/src/gowiki/
/root/.local/go/bin/go build wiki.go
pkill wiki
nohup ./wiki &