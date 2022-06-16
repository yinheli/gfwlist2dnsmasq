FROM golang:1.18 AS builder
LABEL maintainer="mingcheng<mingcheng@outlook.com>"

ENV PACKAGE github.com/yinheli/gfwlist2dnsmasq
ENV BUILD_DIR ${GOPATH}/src/${PACKAGE}
ENV GOPROXY https://goproxy.cn,direct

# Build
COPY . ${BUILD_DIR}
WORKDIR ${BUILD_DIR}
RUN go build ./cmd/gfwlist2dnsmasq \
    && cp ./gfwlist2dnsmasq /bin/gfwlist2dnsmasq

# Stage2
FROM debian:bullseye

ENV TZ "Asia/Shanghai"

RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
	&& sed -i 's/security.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list \
	&& echo "Asia/Shanghai" > /etc/timezone \
	&& apt -y update \
	&& apt -y upgrade \
	&& apt -y install ca-certificates openssl tzdata curl netcat dumb-init \
	&& apt -y autoremove

COPY --from=builder /bin/gfwlist2dnsmasq /bin/gfwlist2dnsmasq

RUN mkdir -p /etc/dnsmasq.d
VOLUME /etc/dnsmasq.d/

ENTRYPOINT ["/bin/gfwlist2dnsmasq"]
