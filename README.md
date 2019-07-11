# Description
This project contains a simple service for hashing passwords.  Given a password, the service will hash the password using [SHA512](https://en.wikipedia.org/wiki/SHA-2) and convert it to [base64](https://en.wikipedia.org/wiki/Base64) encoding.

The project was created as a coding exercise for [JumpCloud](http://www.jumpcloud.com).

# Assumptions

* The hashed passwords are persisted in-memory: there is no long-term persistence. When the service is stopped, all hashed passwords are gone.

* The object identifiers are not random.  They increment monotonically by 1 with every POST request. When the service is restarted, the identifiers start at 1 again.  This is not particularly useful for a real production service, but it's a good exercise in thread locking.

* Since the passwords are persisted in-memory, the service is not designed to scale behind a load balancer.  You can only run one instance of the service.

* The service is not secure.  The current implementation doesn't use TLS: passwords should always be transported using TLS.  Also, the endpoints don't include any token validation for authentication/authorization.

* The service doesn't place any limits on the request body size.  In general, there is an assumption that the service will not be accessible by evil clients.  Adding request limits is not too difficult, if needed.

* The service status data is very limited.  The status should include: P95 and P99 times, memory usage/stress, cpu usage/stress.

* Life is to be enjoyed.

# Endpoints

## POST hash/

Create a new password hash.

The service will return immediately, but the hash will not be available for 5 secs (yes, this is unusual, but it was a requirement for the coding exercise). The service returns the identifier of the hash that can be fetched using the GET endpoint below.

### Example
```
> curl -X POST -F "password=angryMonkey" http://localhost:8080/hash
> 1
```

### Request Parameters

| Name     | Description |
| :---     | :---        |
| password | The password string to SHA512 encode. type: string. *required*. |

### Response

| Name     | Description |
| :---     | :---        |
| id       | The identifer of the hash created.  The id auto-increments from 1. type: int64. |

### Error Codes

| Error     | Description |
| :---     | :---        |
| 200       | OK          |
| 400       | Bad Request. No password was found int the request. |
| 405       | Method not allowed. Only POST methods are allowed. |
| 503       | Service unavailabl. The service is in the process of shutting down and new hash requests are not allowed. |

## GET hash/{id}]

Fetch the password hash for the {id} record.

### Path Parameters

| Name     | Description |
| :---     | :---        |
| id       | The identifier returned by a previous call to POST. type: int64. *required*. |

### Response

| Name     | Description |
| :---     | :---        |
| hash     | The SHA512 hash of the password base64 encoded. type: string. |

### Error Codes

| Error     | Description |
| :---      | :---        |
| 200       | OK          |
| 400       | Bad Request. The path didn't contain an id. |
| 404       | Not found. The id was not found.  Remember, it takes 5 secs for the hash to be available. |
| 405       | Method not allowed. Only GET methods are allowed. |
| 503       | Service unavailable. The service is in the process of shutting down and new hash requests are not allowed. |

### Example
```
> curl http://localhost:8080/hash/1
> ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==
```

## GET stats

Fetch the stats for the service.

### Example
```
> curl http://localhost:8080/stats
> {"total": 4, "average": 1195}
```

### Response

| Name     | Description |
| :---     | :---        |
| total    | The total number of times POST hash has been called since the service started. type: int64. |
| average  | The average time to execute all the POST requests in *microseconds*. type: int64. |

### Error Codes

| Error     | Description |
| :---      | :---        |
| 200       | OK          |
| 405       | Method not allowed. Only GET methods are allowed. |
| 503       | Service unavailable. The service is in the process of shutting down and new hash requests are not allowed. |

## GET shutdown

Shutdown the service.  Wait for all pending hashes to complete.

### Example
```
> curl http://localhost:8080/shutdown
```
### Error Codes

| Error     | Description |
| :---      | :---        |
| 200       | OK        |
| 405       | Method not allowed. Only GET methods are allowed. |
| 503       | Service unavailable. The service is in the process of shutting down and new hash requests are not allowed. |
