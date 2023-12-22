FROM python:3

RUN apt update
RUN pip install ja_timex

ENV ARCH amd64
ENV GOVERSION 1.20.5

RUN set -x \
    && cd /tmp \
    && wget https://dl.google.com/go/go$GOVERSION.linux-$ARCH.tar.gz \
    && tar -C /usr/local -xzf go$GOVERSION.linux-$ARCH.tar.gz \
    && rm /tmp/go$GOVERSION.linux-$ARCH.tar.gz

ENV PATH $PATH:/usr/local/go/bin

WORKDIR /work

COPY . .

RUN go build -o app

ENTRYPOINT ./app
