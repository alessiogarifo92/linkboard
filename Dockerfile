# Un solo Dockerfile per api e worker: scegli il binario con --build-arg SERVICE=api|worker.
# Stage 1: compila un binario statico.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE=api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

# Stage 2: immagine minima (~15 MB), senza shell, utente non root.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
