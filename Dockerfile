# golang:1.19.3-alpine3.16
FROM golang@sha256:d171aa333fb386089206252503bc6ab545072670e0286e3d1bbc644362825c6e as builder

RUN ln -s /usr/local/go/bin/go /usr/local/bin/go

WORKDIR /src/github.com/RichardoC/kube-rest-audit

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

# TODO add tests and uncomment
# RUN CGO_ENABLED=0 go test -timeout 30s .

RUN go build .

# RUN CGO_ENABLED=0 GOOS=linux go build  -ldflags "-s" -a  -installsuffix 'static' .
RUN CGO_ENABLED=0 GOOS=linux go build .


# TODO offer scratch as an option

# alpine:3.16
FROM alpine@sha256:b95359c2505145f16c6aa384f9cc74eeff78eb36d308ca4fd902eeeb0a0b161b

EXPOSE 9090

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /src/github.com/RichardoC/kube-rest-audit/kube-rest-audit /src/github.com/RichardoC/kube-rest-audit/kube-rest-audit

USER 255999

ENTRYPOINT ["/bin/sh", "-c"]

CMD ["/src/github.com/RichardoC/kube-rest-audit/kube-rest-audit"]