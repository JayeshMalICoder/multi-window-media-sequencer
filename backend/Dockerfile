FROM golang:1.23-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server .
COPY --from=build /app/data ./data
ENV PORT=8080
EXPOSE 8080
CMD ["./server"]
