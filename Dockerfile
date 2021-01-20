FROM golang:1.15 AS build

ADD ./ /opt/build/golang
WORKDIR /opt/build/golang
RUN go install ./main/server.go

FROM ubuntu:20.04 AS release

RUN apt-get -y update && apt-get install -y tzdata

ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt -y update && apt install -y postgresql-12

USER postgres
COPY ./init ./init
RUN /etc/init.d/postgresql start &&\
    psql -f ./init/db_init.sql &&\
    /etc/init.d/postgresql stop
EXPOSE 5432

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root
EXPOSE 5000

COPY --from=build go/bin/server /usr/bin/
CMD service postgresql start && server
