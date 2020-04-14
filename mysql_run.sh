#!/bin/bash

mkdir $PWD/data/mysql
podman run --name gowiki-mysql \
-v $PWD/data/mysql:/var/lib/mysql:Z \
-p 3306:3306/tcp \
-e MYSQL_ROOT_PASSWORD=roottoor \
-e MYSQL_DATABASE=gowiki \
-e MYSQL_USER=gowiki \
-e MYSQL_PASSWORD=gowiki55 \
-d \
mariadb:latest
