FROM golang:1.27.2 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOBIN=/out go install -tags='no_clickhouse no_mssql no_mysql no_sqlite3 no_vertica no_ydb no_libsql' github.com/pressly/goose/v3/cmd/goose@v3.28.0

FROM scratch AS tools
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/goose /usr/local/bin/goose
COPY migrations /app/migrations
USER 65532:65532
WORKDIR /app

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/api /usr/local/bin/api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["api"]
