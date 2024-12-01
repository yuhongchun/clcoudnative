FROM harbor.avlyun.org/inf/golang:1.17 as builder
ENV GOPROXY=https://registry-mirrors.exayun.org/repository/goproxy/,https://goproxy.io,direct
ENV GONOPROXY=*.exayun.org
ENV GONOSUMDB=*.exayun.org
ENV GOPRIVATE=*.exayun.org

WORKDIR /build
COPY . .
RUN go build -o nighting-build main.go

FROM harbor.avlyun.org/inf/debian:bullseye

WORKDIR /opt/devops-build
COPY --from=builder /build/devops-build ./devops-build
ENTRYPOINT ["/opt/devops-build/devops-build"]
CMD ["server", "-c=config/settings.yaml"]
EXPOSE 5000
