# WebRTCGame
A template for making real time online games with a server-client structure, web browsers as clients and a pion backend as a server.

The biggest challenge with making real time online browser games (AKA .io games) is that there is no way to send an arbitray UDP packet to a browser client.
WebRTC solves this problem by allowing you to make a secure connection between browser clients and a [pion](https://pion.ly/) client and then send UDP packets between them.


### If you are working on a unix OS based computer

## To preview application i.e run executable binary like in production

    use command `make all`

## To build application executable only

    use command `make build`

## To run application in debug mode

    use command `make run`

### else you are working on a windows os based computer

# cd into directory `cmd` and use `go run main.go`



### Then visit `http://localhost:80` in your web browser to interact with go server using javascript client