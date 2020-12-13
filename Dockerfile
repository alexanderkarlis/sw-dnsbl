FROM golang:1.14 as server

WORKDIR /usr/src/app/server
RUN mkdir -p /usr/src/app/server
RUN mkdir -p /usr/local/bin/
COPY . /usr/src/app/server

RUN . ./config.env
RUN chmod a+x build_and_start_server.sh

ENV HOST=0.0.0.0
ENV WAIT_VERSION 2.7.2
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/$WAIT_VERSION/wait /wait
RUN chmod +x /wait

EXPOSE 8080
CMD ["./build_and_start_server.sh"]