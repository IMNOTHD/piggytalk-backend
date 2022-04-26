# RabbitMQ数据格式

## event
| 类型          | 数据                 |
|-------------|--------------------|
| ContentType | "text/plain"       |
| MessageId   | eventId            |
| Body        | event message json |
| type        | event type         |

## message
| 类型            | 数据                               |
|---------------|----------------------------------|
| CorrelationId | json(sender, talk, ReceiverUuid) |
| MessageId     | messageId                        |
| Body          | protobuf marshal bytes           |
| ContentType   | "binary"                         |
| type          | single message                   |