## adding a multistage build to get better performance out of our container
## the first stage will be the build stage and the second stage will be the runtime stage

# Stage 1: Build Environment
FROM golang:1.23-alpine AS build-stage
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main .

# Stage 2: Runtime environment
FROM alpine:latest AS final-stage
WORKDIR /app
COPY --from=build-stage /app/main .
EXPOSE 8080
CMD ["./main"]
