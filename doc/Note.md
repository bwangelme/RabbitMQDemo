# 轮询分发

默认情况下, RabbitMQ 将会顺序地给每个消费者发送消息。平均下来，每个消费者获得的消息数量是相同的。
这种方式就叫做轮询(round-robin)，可以参考下面这个例子。

![round-robin Example](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-12-014029.png)

示例代码在 [Github](https://github.com/bwangelme/RabbitMQDemo/tree/41c82064edbde8c3bf74c1474420da821f7ac6dc) 上。

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

完整代码见[Github]()
