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

RUN echo "host all  all 0.0.0.0/0  md5" >> /etc/postgresql/12/main/pg_hba.conf
RUN echo "listen_addresses='*'" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "fsync = off" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "synchronous_commit = off" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "work_mem = 8MB" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "shared_buffers = 512MB" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "random_page_cost = 1.0" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "effective_cache_size = 1024MB" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "maintenance_work_mem = 128MB" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "wal_level = minimal" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "wal_buffers = 1MB" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "max_wal_senders = 0" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "log_statement = none" >> /etc/postgresql/12/main/postgresql.conf
RUN echo "log_duration = off " >> /etc/postgresql/12/main/postgresql.conf
RUN echo "log_lock_waits = on" >> /etc/postgresql/12/main/postgresql.conf

EXPOSE 5432

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root
EXPOSE 5000

COPY --from=build go/bin/server /usr/bin/
CMD service postgresql start && server
