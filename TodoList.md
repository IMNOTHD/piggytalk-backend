# TODO清单

[x] 登录  
[x] 心跳  
[x] 图片上传  
[x] rabbitmq添加  
[x] 添加/删除好友
[&ensp;] 群聊邀请  
[x] 私聊  
[&ensp;] 群聊  
[&ensp;] 微服务监测维护模块  
[x] 死信处理  
[x] 消息存储  
[x] 头像设置  
[&ensp;] 修改密码  
[&ensp;] 用户设置  
[&ensp;] 修改messageId和eventId的生成流程
```
具体流程如下:
1. client发出message后, 对该uuid陷入等待, 1s检查一次是否ack, 3s后提示是否重发
2. 到达message再生成id, 并把id发回, 并用uuid做幂等
```
[&ensp;] 修改gorm的log插件