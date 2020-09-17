# => frontend build environment
FROM node:13.12.0-alpine as frontend-builder

WORKDIR /frontend
ENV PATH /frontend/node_modules/.bin:$PATH
COPY frontend/package.json ./
COPY frontend/yarn.lock ./
RUN yarn install --pure-lockfile

COPY frontend ./
RUN yarn build

# => backend build env
FROM golang:1.14 as backend-builder

WORKDIR /workspace
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN make linux
RUN echo 'nobody:x:65534:65534:Nobody:/:' > passwd.minimal

# => Run container
FROM ubuntu:18.04
RUN apt-get update && apt-get install -y ffmpeg
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /workspace/dist/linux/cmd /server
COPY --from=backend-builder /workspace/passwd.minimal /etc/passwd
# TODO: remove this line below
COPY --from=backend-builder /workspace/internal/app/http/templates /internal/app/http/templates

COPY --from=frontend-builder /frontend/build /frontend/build
USER nobody

ENTRYPOINT ["/server"]
