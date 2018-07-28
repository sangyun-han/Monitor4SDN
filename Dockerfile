FROM sangyunhan/ubuntu-for-network-test
MAINTAINER Sangyun Han <sangyun628@gmail.com>

# setup golang
RUN add-apt-repository ppa:gophers/archive
RUN apt-get update
RUN apt-get install golang-1.10-go -y

# setup controller