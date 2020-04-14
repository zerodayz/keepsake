#!/bin/bash
set -x

# COPY DL URL FROM https://golang.org/dl/
curl -o $HOME/go1.14.1.linux-amd64.tar.gz https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz
tar -C $HOME/.local -xzf  $HOME/go1.14.1.linux-amd64.tar.gz
export PATH=$PATH:$HOME/.local/go/bin

mkdir -p $HOME/go
export GOPATH=$HOME/go
cd $GOPATH
go get -u github.com/go-sql-driver/mysql
go get -u golang.org/x/crypto/bcrypt
git clone https://github.com/zerodayz/gowiki.git $GOPATH/src/gowiki/
cd $GOPATH/src/gowiki/
/root/.local/go/bin/go build wiki.go

mkdir $PWD/data/mysql

docker run --name gowiki-mysql \
-v $PWD/data/mysql:/var/lib/mysql:Z \
-p 3306:3306/tcp \
-e MYSQL_ROOT_PASSWORD=roottoor \
-e MYSQL_DATABASE=gowiki \
-e MYSQL_USER=gowiki \
-e MYSQL_PASSWORD=gowiki55 \
-d \
mariadb:latest

cd $GOPATH/src/gowiki/
nohup ./wiki &