FROM golang:1.26.2-trixie@sha256:c0074c718b473f3827043f86532c4c0ff537e3fe7a81b8219b0d1ccfcc2c9a09

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends\
        gettext-base=0.21-4 \
        iptables=1.8.7-1 \
        libgl1-mesa-dev=20.3.5-1 \
        xorg-dev=1:7.7+22 \
        libayatana-appindicator3-dev=0.5.5-2+deb11u2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && go install -v golang.org/x/tools/gopls@latest


WORKDIR /app