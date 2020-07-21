[![Go Report](https://goreportcard.com/badge/github.com/zerodayz/gowiki)](https://goreportcard.com/badge/github.com/zerodayz/gowiki)

# GoWiki

Wiki written in Go

This is free wiki written in Go, for personal development purposes. 

Initial Wiki started off the Go Learning project (Web Go Application)[https://golang.org/doc/articles/wiki/]
Insipiration of some features from (jmoiron's gowiki)[https://github.com/jmoiron/gowiki] project last updated 4 years ago and from (ieyasu's go-bwiki)[https://github.com/ieyasu/go-bwiki] which had updates couple of months back.

The idea is to keep the Wiki as close to the standard libraries as possible, avoiding using any unnecessary third party libraries.

![Wiki Homepage](https://github.com/zerodayz/gowiki/blob/master/screenshots/WikiHome.png?raw=true)

See (screenshots)[https://github.com/zerodayz/gowiki/tree/master/screenshots] folder for more.

# How to use

Currently you will need mysql, that is dependency and used for Pages, History and User management.

```
docker run --name gowiki-mysql -v $PWD/data/mysql:/var/lib/mysql:Z -p 3306:3306/tcp -e MYSQL_ROOT_PASSWORD=roottoor -e MYSQL_DATABASE=gowiki -e MYSQL_USER=gowiki -e MYSQL_PASSWORD=gowiki55 -d mariadb:latest
```

The default port the wiki is listening on is `8080`.

## Pre-requirements
Some basics are already provided in the `install_gowiki.sh` script.
~~~
Go
~~~

## Installation
~~~
go get https://github.com/zerodayz/gowiki/
~~~

Navigate to `http://localhost:8080` and enjoy.
