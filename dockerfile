# syntax=docker/dockerfile:1

FROM golang:1.24.4 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-fileEditor ./cmd

FROM build-stage AS run-test-stage

RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

# Binary
COPY --from=build-stage /docker-fileEditor /docker-fileEditor

# Swagger
COPY --from=build-stage /app/api/swagger.yaml /api/swagger.yaml
COPY --from=build-stage /app/api/swagger-ui /api/swagger-ui

# Email templates
COPY --from=build-stage /app/templates/forgotPassword.html /templates/forgotPassword.html
COPY --from=build-stage /app/templates/welcome.html /templates/invite.html
COPY --from=build-stage /app/templates/welcome.html /templates/accountCompletion.html
COPY --from=build-stage /app/templates/welcome.html /templates/welcome.html

# Font assets
COPY --from=build-stage /app/assets/fonts/Roboto-Bold.ttf /assets/fonts/Roboto-Bold.ttf
COPY --from=build-stage /app/assets/fonts/Roboto-Regular.ttf /assets/fonts/Roboto-Regular.ttf

# Other
COPY --from=build-stage /app/ca.pem /ca.pem

USER nonroot:nonroot

EXPOSE 9091

ENTRYPOINT ["/docker-fileEditor"]
