# => frontend build environment
FROM node:13.12.0-alpine as build

WORKDIR /frontend
ENV PATH /frontend/node_modules/.bin:$PATH
COPY frontend/package.json ./
COPY frontend/package-lock.json ./
RUN npm ci --silent
RUN npm install react-scripts@3.4.1 -g --silent
COPY frontend ./

RUN npm run build

# => backend build env
FROM golang:1.14 as builder

WORKDIR /workspace
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN make linux
RUN echo 'nobody:x:65534:65534:Nobody:/:' > passwd.minimal

# => Run container
FROM ubuntu:18.04
RUN apt-get update && apt-get install -y ffmpeg
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /workspace/dist/linux/cmd /server
COPY --from=builder /workspace/passwd.minimal /etc/passwd
USER nobody

ENTRYPOINT ["/server"]
