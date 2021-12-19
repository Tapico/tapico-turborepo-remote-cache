FROM golang

RUN useradd -m tapico

RUN mkdir /build

RUN chown -Rvf tapico: /build

USER tapico


WORKDIR /build

COPY . .

RUN go build

RUN go install

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["tapico-turborepo-remote-cache"]
