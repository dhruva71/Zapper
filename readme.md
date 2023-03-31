# Zapper

Zapper is a tiny microservice built on top of [zap](https://github.com/uber-go/zap) to function as a logging service.

## Usage

### HTTP
Submit a POST request to the `/log` endpoint with a JSON body that looks like:
```json
{
    "application": "my-app", 
    "message": "hello, world", 
    "level": "info"
}
```
### AMQP
Send a JSON message to the log_queue queue with the following format:
```json
{
  "application": "my_app",
  "level": "info",
  "message": "This is a test log message"
}
```

## Testing
You can test the microservice using tools like curl for HTTP or create a custom AMQP producer to send messages to the log_queue queue.

### HTTP
```bash
curl -X POST -H "Content-Type: application/json" -d '{"application":"my_app","level":"info","message":"This is a test log message"}' http://localhost:8080/log
```
### AMQP
Refer to the `log_amqp_producer.go` file for an example of how to send messages to the log_queue queue.


