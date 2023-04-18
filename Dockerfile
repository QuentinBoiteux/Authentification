FROM golang:1.18.4

LABEL version="1.0"
LABEL authors="Quentin Boiteux/Daryl Parisi/ Evan Duvivier"
LABEL name="Campus SAINT MARC"
RUN apt update
RUN apt install sqlite3

WORKDIR /forum
COPY . .

RUN go mod tidy
RUN go build .

EXPOSE 80
EXPOSE 443

CMD ["./forum"]