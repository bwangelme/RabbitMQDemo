# 轮询分发

默认情况下, RabbitMQ 将会顺序地给每个消费者发送消息。平均下来，每个消费者获得的消息数量是相同的。
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

__[代码见 Github@]()__

可以看到2号窗口和3号窗口中的 Worker 都分配到了耗时较长的任务。
