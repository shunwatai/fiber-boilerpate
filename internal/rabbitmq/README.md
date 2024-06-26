# Rabbitmq

I try it to handle the request's log in the `internal/middleware/logging/main.go`.
Can check the QueueLog() for how it publish the logs into the queue.

## Publisher
`rabbitmq.go` contains the helper functions for its operations.

## Worker(consumer)
`worker.go` is for consuming the message from the queues.

Run the worker by following:
```
go run main.go run-rbmq-worker
```
