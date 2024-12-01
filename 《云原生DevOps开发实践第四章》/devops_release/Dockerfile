FROM harbor.avlyun.org/inf/golang:1.17 as builder
ENV GOPROXY=https://registry-mirrors.exayun.org/repository/goproxy/,https://goproxy.io,direct
ENV GONOPROXY=*.exayun.org
ENV GONOSUMDB=*.exayun.org
ENV GOPRIVATE=*.exayun.org

ARG GOMODCACHEDIR
WORKDIR /build
COPY . .
RUN go build -o nighting-release main.go

FROM harbor.exayun.org/inf/debian:bullseye

WORKDIR /opt/nighting-release
COPY --from=builder /build/devops-release ./devops-release
ENTRYPOINT ["/opt/devops-release/devops-release"]
CMD ["server", "-c=config/settings.yaml"]
EXPOSE 5000