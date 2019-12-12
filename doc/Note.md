# 轮询分发

默认情况下, RabbitMQ 将会顺序地给每个消费者发送消息。平均下来，每个消费者获得的消息数量是相同的。
这种方式就叫做轮询(round-robin)，可以参考下面这个例子。

![round-robin Example](https://passage-1253400711.cos-website.ap-beijing.myqcloud.com/2019-12-12-014029.png)

示例代码在 [Github]() 上。
