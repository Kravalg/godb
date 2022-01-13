# GoDB
Key value storage for data presented as strings

## Features
 - Add data
 - Edit data
 - Get data
 - Delete data

## How to run the GoDB

Install the latest [Golang](https://go.dev/dl/) version 

Clone repo and build database from source code
```
go build src/main.go
```

Launch database
```
./main
```

## How to use the GoDB

GoDB uses TCP to communicate with the clients and available at host `localhost` and port `6666`

TCP client available to send the following commands to the GoDB server

### Add or update record
```SET <key> <value>;```

### Read record
```GET <key>;```

### Remove record
```DELETE <key>;```

### Also, client has the opportunity to send multiple commands in one batch request
```SET <key> <value>;GET <key>;DELETE <key>;```

## Examples of using

```SET name John;```

```GET name;```

```DELETE name;```