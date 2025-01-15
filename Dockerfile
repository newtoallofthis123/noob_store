FROM golang:1.23

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux cd cmd && go build -o ../noob_store

EXPOSE 8000

CMD ["./noob_store"]
