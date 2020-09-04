# NOTE: Must be run from root for file paths to be valid
FROM golang:1.14 as builder

WORKDIR /workspace
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN make linux
RUN echo 'nobody:x:65534:65534:Nobody:/:' > passwd.minimal

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y ffmpeg
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/dist/linux/cmd /server
COPY --from=builder /workspace/passwd.minimal /etc/passwd
USER nobody

ENTRYPOINT ["/server"]
