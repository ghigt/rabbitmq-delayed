This project demonstrates how to use the [`rabbitmq_delayed_message_exchange`](https://www.rabbitmq.com/blog/2015/04/16/scheduling-messages-with-rabbitmq/)'s plugin.

1. An exchange is created with the `x-delayed-message` type.
2. Then we declare a queue `user-published-queue` that we bind to the routing key `user.event.publish`.
3. Then we send an event with into `user.event.publish` with a 'delay' of 10 sec.
4. Finally, we consume the queue `user-published-queue` and we can see that the event sent is only received after the delay of 10 sec.

We also handle a disconnection and reconnection of rabbitmq trying every 5 sec if a new connection is available.

Uses `dep` tool for dependency management:
```sh
$ go get -u github.com/golang/dep/cmd/dep
$ dep ensure
```
