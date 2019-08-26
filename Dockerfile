FROM golang:1.12-alpine 

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN go get -u github.com/gin-gonic/gin
RUN go get -u github.com/jinzhu/gorm
RUN go get -u github.com/jinzhu/gorm/dialects/postgres  
RUN go get -u github.com/swaggo/gin-swagger
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN go get -u github.com/alecthomas/template
RUN go get -u github.com/swaggo/files
RUN go get -u github.com/swaggo/http-swagger
RUN go get -u github.com/gin-contrib/cors
RUN go get -u github.com/rs/cors/wrapper/gin
RUN go get -u github.com/dgrijalva/jwt-go

# Add Maintainer Info
LABEL maintainer="Fernando Feliu"

# Set the Current Working Directory inside the container
WORKDIR /go/src/projetoapi

# Copy everything from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN swag init

RUN go build -o main .

# Expose port 8081 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./main"]


