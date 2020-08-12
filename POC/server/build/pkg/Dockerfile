FROM golang:1.13 as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make linux
RUN echo 'nobody:x:65534:65534:Nobody:/:' > passwd.minimal

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/dist/linux/cmd /server
COPY --from=builder /workspace/passwd.minimal /etc/passwd
USER nobody

CMD ["/server"]