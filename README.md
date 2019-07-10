# Description
This project contains a simple service for hashing passwords using SHA512.

The project was created as a coding exercise for [JumpCloud](http://www.jumpcloud.com).

# Endpoints

## POST hash/

Create a new password hash.

### Request Parameters

| Name     | Description |
|----------|-------------|
| password | string      |
|          | The password string to SHA512 encode. |
|          | *required* |



### Example
```
> curl -X POST -F "password=angryMonkey" http://localhost:8080/hash
> 1
```

## GET hash/{id}

### Example
```
> curl http://localhost:8080/hash/1
> ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==
```

## GET stats

### Example
```
> curl http://localhost:8080/stats
> {"total": 4, "average": 1195}
```

## GET shutdown

### Example
```
> curl http://localhost:8080/shutdown
```
