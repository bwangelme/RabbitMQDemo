# 轮询分发

默认情况下, RabbitMQ 将会顺序地给绑定同一个队列的每个消费者发送消息。平均下来，每个消费者获得的消息数量是相同的。
这种方式就叫做轮询(round-robin)，可以参考下面这个例子。

![round-robin Example](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-12-014029.png)

完整代码见 [Github@41c82064](https://github.com/bwangelme/RabbitMQDemo/tree/41c82064edbde8c3bf74c1474420da821f7ac6dc) 。

# 消息确认

当消费者收到消息后，可能会遇到某种异常崩溃了，此时这条消息就会丢失了。

为了避免这种情况，我们可以使用 RabbitMQ 提供的消息确认机制。

消费者在消费完消息后，再向 RabbitMQ 发送 ack。收到 ack 之后，RabbitMQ 才会把这条消息标记为可删除的，并择机删除。
如果 RabbitMQ 没有收到 ack，消费者就死掉了(channel 关闭，连接关闭，或者 TCP 连接关闭)。
那么 RabbitMQ 就会认为这条消息没有被消费完，它就会重新入队，然后被快速发送给其他消费者。

使用了消息确认机制后，我们就可以确保及时存在消费者偶尔崩溃的情况，我们的消息也不会丢失。

在消息确认机制中，没有任何的超时限制，所以即使客户端花费很长的时间去处理消息，也不用担心消息会被误重发。

## Example

我们修改 worker 的代码，将`auto_ack`关掉，在处理完消息后，我们手动发送 Ack。

```go
	msgs, err := ch.Consume(
		q.Name, // name
		"",     // consumer
		false,   // 关闭掉 autoack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dot_count := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dot_count)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false) // 手动发送 ack
		}
	}()
```

示例如下，第一个消费者在 `09:59:25` 收到消息后，就被我们 kill 掉了，在`09:59:38`另一个消费者就收到了这个消息。

![](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-12-015958.png)

完整代码见[Github@9cd87dd](https://github.com/bwangelme/RabbitMQDemo/tree/9cd87dd)

## 注意事项

如果我们忘记了返回 ack，那么后果是很严重的，在消费者退出的时候，RabbitMQ 将会重发消息。在消费者未退出之前， RabbitMQ 将会不断占用内存存储未被确认的消息，直到内存被吃光，然后 RabbitMQ 将不会再发送任何消息了。

为了监控 RabbitMQ 中消息的情况，可以使用如下的命令:

```sh
sudo rabbitmqctl list_queues name messages_ready messages_unacknowledged
```

# 持久化

通过消息确认，我们可以在消费者崩溃的情况下，让我们的消息不丢失。但如果 RabbitMQ 崩溃，或者 RabbitMQ 所在的节点宕机的话，消息仍然可能会丢失。

在这种情况下，我们可以使用消息和队列持久化，这样消息即使 RabbitMQ 退出，消息和队列也回被持久化到磁盘中，不会丢失。

## 声明队列持久化

声明队列持久化的代码如下:

```go
	q, err := ch.QueueDeclare(
		"task_queue", // Queue name
		true,   // durable  持久性
		false,   // delete when unused
		false,   // exclusive 独占
		false,   // no-wait
		nil,     // arguments
	)
```

__注意:__ 在声明队列时，如果我们声明一个已经存在的队列，但是初始化参数不同的时候，`QueueDeclare`会失败并返回一个 err，因此这里我们可以给queue起另外一个名字`task_queue`

__注意:__ 生产者和消费者在声明队列时，传递的参数应该是相同的。


## 声明消息持久化

在发送消息的时候，我们可以设置一个 `amqp.Persistent` 选项，来表明这个消息应该被持久化。

```go
	err = ch.Publish(
		"",     //exchange
		q.Name, // routing key
		false,  // mandatory 强制的
		false,  // immediate 即时的
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // 声明消息持久化
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
```

## 关于消息持久化注意事项

上述声明消息持久化的代码，并不能完全保证消息不会丢失，有以下两方面的原因:

1. 在 RabbitMQ 收到消息和 RabbitMQ 将消息写入磁盘这两个事件中仍然有一个短暂的时间窗口。
2. RabbitMQ 并不会每次收到消息后，都调用 `fsync(2)`，消息可能被存储在缓存中，过一段时间后才被写入到磁盘中。

因此，这个持久化并不是强健的，但是对于我们这个 demo 已经够用了，如果你想使用更强健的持久化策略，可以考虑使用 [Publisher Confirms](https://www.rabbitmq.com/confirms.html)。

# 公平分发

如果按照轮询分发的策略，那么可能会出现一个 Worker 特别忙，但是另外一个 Worker 很闲的情况。

例如有两个 Worker，我们分发的消息中，奇数的都是轻松的，偶数的都是困难的，那样就造成第一个 Worker 很轻松，但是第二个 Worker 特别繁忙。

![](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-13-141545.png)

__[上图代码见 Github@14a0414 ](https://github.com/bwangelme/RabbitMQDemo/tree/14a0414)__

在上面的例子中我们可以看到，2号窗口中的 Worker 接收到的都是繁忙任务，3号窗口中接收到的都是轻松任务。

为了避免这种极端情况的发生，我们可以设置预取值(`prefetch count`)为1。这样的话，相当于告诉 RabbitMQ，在 Worker 消费完一个消息之前，不要再给他分发新的消息了，这样的话随后的消息就会被分发给其他的空闲的 Worker 了。

在具体代码如下:

```go
// 在 Worker 的代码中设置
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
```

运行效果如下:

![](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-13-142622.png)

__[代码见 Github@ad5507e](https://github.com/bwangelme/RabbitMQDemo/tree/ad5507e)__

可以看到2号窗口和3号窗口中的 Worker 都分配到了耗时较长的任务。

__注意事项:__ 这样操作容易让 RabbitMQ 消息被塞满，需要有合适的监控机制来监控消息的数量。

# Exchange

在 RabbitMQ 中，生产者不会将消息直接发到队列中，而是发送到 exchange 上。

Exchange 一端接收消息，另外一端发送消息到队列中，Exchange 处理消息的方式是由它的类型决定的，Exchange 共有这么几种类型:
`direct`, `topic`, `headers`, `fanout`。

## 查看 Exchange

通过命令 `rabbitmqctl list_exchanges` 就可以列出所有的 Exchange:

```sh
>>> rabbitmqctl list_exchanges
Listing exchanges for vhost / ...
name    type
amq.fanout      fanout
amq.headers     headers
amq.match       headers
amq.direct      direct
amq.topic       topic
        direct
amq.rabbitmq.log        topic
amq.rabbitmq.trace      topic
```

我们可以看到 RabbitMQ 中已经预置了 `amq.*` Exchange 和一个没有名字的 Exchange。无名 Exchange 是默认的，所有未指定 Exchange 的消息都会发到这里来。

## Direct Exchange

当我们像下面这样，发送消息时没有指定 Exchange 时，消息会发到默认的`""` Exchange 上。`""` Exchange 是`direct`类型，它将会根据 `routing key` 参数将消息转发到对应的同名队列上。

+ 生产者发送消息到 Direct Exchange

```go
	err = ch.Publish(
		"",     //exchange
		q.Name, // routing key，这里指定的是队列的名称。
		false,  // mandatory 强制的
		false,  // immediate 即时的
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		},
	)
```

同样，消费者在声明队列的时候，如果手动给队列绑定 Exchange，那么队列会默认绑定到 `""` Exchange 上。这个队列对应的 `Bind key` 就是队列名称。

+ 消费者将 queue 绑定到 direct exchange，并消费消息的代码

```go
err = ch.QueueBind(
    q.Name,        // queue name
    "some routing",         // routing key
    "", // exchange name
    false,         // no-wait
    nil,           // args
)

msgs, err := ch.Consume(
    q.Name, // name
    "",     // consumer
    true,   // auto ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
)
```

我们将`Publish`时指定的`routing key`称为`Route Key`, 将消费者 Bind `direct` Exchange 时指定的`routing key`成为 `Bind key`。

Direct Exchange 的转发逻辑很简单，就是寻找 `Bind key` == `Route key` 队列，然后将消息转发过去。它支持一个队列使用多个 `Bind Key`，也支持多个队列使用同一个 `Bind key`。

### Example

以下示例的代码可以在 [Github@3593d57](https://github.com/bwangelme/RabbitMQDemo/tree/3593d57) 上找到。

在下面这个例子中，我们使用日志等级(warning|info|error等)作为`routing key`，对应等级的日志，就只会发到对应`Bind Key`的队列上。

![](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-18-134234.png)

## Fanout Exchange

Fanout 类型的 Exchange 策略非常简单，它就是将收到的消息发送给所有绑定的队列，这点从它的名字就可以看出。

### Example

示例如下:

![](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-14-051451.png)

__代码见 [Github@cf8f902](https://github.com/bwangelme/RabbitMQDemo/tree/cf8f902)__

我们可以看到二号窗口和三号窗口中的消费者都收到了生成者发送的消息。

同时我们也可以通过命令查看我们创建的队列，Exchange 和 Binding。

## Topic Exchange

下面是一个 Topic Exchange 的例子，代码见 [Github@b56c3b1](https://github.com/bwangelme/RabbitMQDemo/tree/b56c3b1)

![](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-18-142810.png)

注意红框处的消息，2号窗口中的消费者订阅了 `*.info` 和 `net.*` 两个模式，当1号窗口的生产者发送了 `net.info` 模式的消息时，2号窗口的消费者只收到了一条。


```sh
>>> rabbitmqctl list_bindings
Listing bindings for vhost /...
source_name     source_kind     destination_name        destination_kind        routing_key     arguments
logs    exchange        amq.gen-Q0YVQhBf8hBKPnU5uH4mOA  queue           []
logs    exchange        amq.gen-jKZ7HxokfjOGh-NHVeB8Ew  queue           []
...
>>> rabbitmqctl list_exchanges
Listing exchanges for vhost / ...
name    type
logs    fanout
...
>>> rabbitmqctl list_queues
Timeout: 60.0 seconds ...
Listing queues for vhost / ...
amq.gen-jKZ7HxokfjOGh-NHVeB8Ew  0
amq.gen-Q0YVQhBf8hBKPnU5uH4mOA  0
...
```

# 队列

## 临时队列

当我们需要创建一个只用一次的队列时，可以通过指定 `exclusive` 参数来实现:

```go
q, err := ch.QueueDeclare(
  "",    // name
  false, // durable
  false, // delete when unused
  true,  // exclusive
  false, // no-wait
  nil,   // arguments
)
```

上面的代码创建了一个临时队列，在 channel 关闭之后，该队列就会被自动删除，注意我们设置了三个参数，`name`, `durable`, `exclusive`。

上述代码返回的队列中，队列的名字类似于这样: `amq.gen-JzTY20BRgKO-HjmUJj0wLg`。想要了解 Exclusive Queue 的更多信息，请参考文档 [Guide on Queues](https://www.rabbitmq.com/queues.html)

# Binding

当队列创建好了以后，我们需要通过 binding 将队列和 Exchange 连接起来，具体使用的代码如下:

```go
err = ch.QueueBind(
  q.Name, // queue name
  "",     // routing key
  "logs", // exchange
  false,
  nil,
)
```

通过 `rabbitmqctl list_bindgs` 命令我们可以查看 RabbitMQ 中所有的 binding。
