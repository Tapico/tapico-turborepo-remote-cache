FROM golang as builder

RUN mkdir /build

WORKDIR /build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o tapico-turborepo-remote-cache .


FROM scratch

COPY --from=builder /build/tapico-turborepo-remote-cache /app/

WORKDIR /app

ENV PORT=8080
ENV LISTEN_ADDRESS=0.0.0.0:${PORT}

EXPOSE $PORT

ENV PATH=$PATH:/app

ENTRYPOINT ["./tapico-turborepo-remote-cache"]
