#!/bin/bash
set -x

# COPY DL URL FROM https://golang.org/dl/
curl -o $HOME/go1.14.1.linux-amd64.tar.gz https://dl.google.com/go/go1.14.1.linux-amd64.tar.gz
tar -C $HOME/.local -xzf  $HOME/go1.14.1.linux-amd64.tar.gz
