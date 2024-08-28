FROM golang:1.23 AS builder
LABEL intermediateStageToBeDeleted=true

WORKDIR /build
COPY . .

# CGO_ENABLED must be disabled to run go binary in Alpine
RUN CGO_ENABLED=0 go build -o bin/client ./client


FROM busybox:latest
COPY --from=builder /build/bin/client /client
ENTRYPOINT ["/bin/sh"]
