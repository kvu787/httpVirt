#!/usr/bin/env bash

# install docker
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates
echo "deb https://apt.dockerproject.org/repo ubuntu-trusty main" | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt-get update
sudo apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual
sudo apt-get update
sudo apt-get install -y --force-yes docker-engine

# install go
wget --quiet https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
tar -xf go1.7.3.linux-amd64.tar.gz
sudo mv go /usr/local/go

# setup go workspace environment variables
echo 'export GOPATH=~/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:/usr/local/go/bin

# build the Docker image
cd ~
mkdir -p ~/go/src/github.com/kvu787
mv ~/httpVirt ~/go/src/github.com/kvu787
echo $GOPATH/src/github.com/kvu787/httpVirt/image
cd $GOPATH/src/github.com/kvu787/httpVirt/image
sudo docker build -t httpvirt .

# install git
sudo apt-get update
sudo apt-get install -y git

# run server
cd $GOPATH/src/github.com/kvu787/httpVirt/server
go get -d
go build
sudo ./server cors >~/log.txt 2>&1 &

# wait for server to spin up
sleep 3
