#!/bin/bash

cd $GOPATH
go get -u github.com/go-sql-driver/mysql
go get -u golang.org/x/crypto/bcrypt
cd $GOPATH/src/gowiki/
./mysql_run.sh
