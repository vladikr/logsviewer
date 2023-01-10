# syntax=docker/dockerfile:1
FROM golang:1.17 AS builder

### Install NodeJS and yarn
ENV NODE_VERSION="v14.16.0"
ENV YARN_VERSION="v1.22.10"

# yarn needs a home writable by any user running the container
ENV HOME /opt/home
RUN mkdir -p ${HOME}
RUN mkdir -p /frontend/build
RUN chmod 777 -R ${HOME}

RUN cd /tmp && \
    wget --quiet -O /tmp/node.tar.gz http://nodejs.org/dist/${NODE_VERSION}/node-${NODE_VERSION}-linux-x64.tar.gz && \
    tar xf node.tar.gz && \
    rm -f /tmp/node.tar.gz && \
    cd node-* && \
    cp -r lib/node_modules /usr/local/lib/node_modules && \
    cp bin/node /usr/local/bin && \
    ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm
# so any container user can install global node modules if needed
RUN chmod 777 /usr/local/lib/node_modules
# cleanup
RUN rm -rf /tmp/node-v*

RUN cd /tmp && \
    wget --quiet -O /tmp/yarn.tar.gz https://github.com/yarnpkg/yarn/releases/download/${YARN_VERSION}/yarn-${YARN_VERSION}.tar.gz && \
    tar xf yarn.tar.gz && \
    rm -f /tmp/yarn.tar.gz && \
    mv /tmp/yarn-${YARN_VERSION} /usr/local/yarn && \
    ln -s /usr/local/yarn/bin/yarn /usr/local/bin/yarn

WORKDIR /app

COPY build-frontend.sh ./
COPY go.mod ./
COPY go.sum ./
USER 0
RUN go mod download

COPY pkg/ pkg/
COPY frontend/ frontend/
COPY cmd/ cmd/
RUN CGO_ENABLED=0 go build -o backend cmd/backend/backend.go
RUN ./build-frontend.sh

FROM alpine:3.15
WORKDIR /
RUN ls -lsaR
COPY --from=builder app/backend /
COPY --from=builder app/frontend/dist /frontend/build
