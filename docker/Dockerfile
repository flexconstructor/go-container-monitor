FROM golang:1.7
MAINTAINER FlexConstructor <flexconstructor@gmail.com>
COPY container_monitor_setup.go /go/src
RUN go-wrapper download github.com/shirou/gopsutil                       \
                        github.com/sevlyar/go-daemon                     \
                        gopkg.in/redis.v4                                \
                        github.com/flexconstructor/go-container-monitor  \

  && go-wrapper install  github.com/shirou/gopsutil                      \
                        github.com/sevlyar/go-daemon                     \
                        gopkg.in/redis.v4                                \
                        github.com/flexconstructor/go-container-monitor  \
  && mkdir /go/logs                                                      \
  && chmod 775 /go/logs
VOLUME ["/go","/usr/local/go"]