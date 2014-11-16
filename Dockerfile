FROM google/golang:1.3

MAINTAINER Máximo Cuadros <mcuadros@gmail.com>

RUN mkdir -p $GOPATH/src/github.com/mcuadros/dockership
WORKDIR $GOPATH/src/github.com/mcuadros/dockership/

RUN git clone https://github.com/mcuadros/dockership.git .
RUN make
RUN make install

VOLUME /etc/dockership

EXPOSE 80
CMD ["dockershipd"]
