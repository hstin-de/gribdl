# Stage 1: Generate weights
FROM ghcr.io/hstin-de/cdo AS weights

RUN apt update && apt install -y wget bzip2

WORKDIR /app

COPY ./weights ./weights
COPY ./generateWeights.sh .

RUN chmod +x generateWeights.sh
RUN ./generateWeights.sh



# Stage 2: Build the binary
FROM golang:1.21.5-alpine AS build

WORKDIR /app
COPY ./src .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o gribdl .


# Stage 3: Final stage
FROM ghcr.io/hstin-de/cdo

WORKDIR /app

COPY --from=weights /app/weights /app/weights
COPY ./weights/*.txt /var/tmp/gribdl/dwd/weights
COPY --from=build /app/gribdl .

RUN chmod +x gribdl

ENV PATH="/app:${PATH}"
ENTRYPOINT ["/app/gribdl"]