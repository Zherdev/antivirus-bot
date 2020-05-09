FROM golang

RUN apt-get update
RUN apt-get install sudo
RUN sudo apt-get update -y
RUN sudo apt-get install -y clamdscan
RUN sudo touch /etc/clamav/clamd.conf
RUN sudo echo "TCPSocket 3310" > /etc/clamav/clamd.conf
RUN sudo echo "TCPAddr clamav" >> /etc/clamav/clamd.conf

ADD . /go/src/antivirus-bot/

RUN mkdir /build
WORKDIR /build/

ADD config.json /build/

RUN go get /go/src/antivirus-bot/...
RUN go build -o /build/antivirus-bot /go/src/antivirus-bot/cmd/app/main.go
>>>>>>> dev

ENTRYPOINT ./antivirus-bot
