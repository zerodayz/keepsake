[![Go Report Card](https://goreportcard.com/badge/github.com/zerodayz/keepsake)](https://goreportcard.com/report/github.com/zerodayz/keepsake)

# Keepsake
This is Open Source Wiki written in Go, aims for simplicity over complexity. It is actively maintained and serves as Go development project.

Initial Wiki started off the Go Learning project (Web Go Application)[https://golang.org/doc/articles/wiki/]
Insipiration of some features from (jmoiron's gowiki)[https://github.com/jmoiron/gowiki] project last updated 4 years ago and from (ieyasu's go-bwiki)[https://github.com/ieyasu/go-bwiki] which had updates couple of months back.

The idea is to keep the Wiki as close to the standard libraries as possible, avoiding using any unnecessary third party libraries.

# How to use
## Run the DB container
```
docker run --name gowiki-mysql -v $PWD/data/mysql:/var/lib/mysql:Z -p 3306:3306/tcp -e MYSQL_ROOT_PASSWORD=roottoor -e MYSQL_DATABASE=gowiki -e MYSQL_USER=gowiki -e MYSQL_PASSWORD=gowiki55 -d mariadb:latest
```

## Clone the keepsake into your own Filesystem
~~~
git clone git@github.com:zerodayz/keepsake.git
~~~
OR use HTTPS
~~~
https://github.com/zerodayz/keepsake.git
~~~

### Change directory to keepsake
~~~
cd keepsake
~~~

## Install TLS/SSL Cert
~~~
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
~~~
This will generate server.key and server.crt for your Keepsake server.

## Build Keepsake server
~~~
go build wiki.go
~~~

## Run keepsake
~~~
./wiki
~~~

The default port the wiki is listening on is `443`.
Navigate to `https://localhost` and enjoy.

## Dashboard
![Dashboard](screenshots/Dashboard.png)

## Create new user
![Create_User](screenshots/Create_User.png)

![Create_User_2](screenshots/Create_User_2.png)

## Login to the Wiki
![Login](screenshots/Login.png)

## Create new category
![Create_Category](screenshots/Create_Category.png)

## Create new page
![Create_Page](screenshots/Create_Page.png)

