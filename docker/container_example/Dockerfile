FROM flexconstructor/docker-centos-golang
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN echo  "[program:system_monitor]" >> /etc/supervisord.conf
RUN echo  "command = go run /go/src/container_monitor_setup.go" >> /etc/supervisord.conf 
CMD ["/usr/bin/supervisord","-n", "-c", "/etc/supervisord.conf"]