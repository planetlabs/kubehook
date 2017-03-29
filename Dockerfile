FROM alpine:3.5
MAINTAINER Nic Cope <n+docker@rk0n.org>

ENV APP /kubehook

RUN mkdir -p "${APP}"
COPY "dist/kubehook" "${APP}"