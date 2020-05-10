FROM golang

RUN apt-get update -y
RUN apt-get install -y clamdscan
RUN touch /etc/clamav/clamd.conf
RUN echo "TCPSocket 3310" >> /etc/clamav/clamd.conf
RUN echo "TCPAddr clamav" >> /etc/clamav/clamd.conf

ADD . /go/src/antivirus-bot/

RUN mkdir /build
WORKDIR /build/

ADD config.json /build/

RUN go get /go/src/antivirus-bot/...
RUN go build -o /build/antivirus-bot /go/src/antivirus-bot/cmd/app/main.go

ENTRYPOINT ./antivirus-bot
