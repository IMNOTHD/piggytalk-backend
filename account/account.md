# 帐号规则


## 密码传输
前端 ---> 后端 

md5(password+```salt```"piggytalk") 

后端 ---> 数据库

bcrypt(password+```salt``` ```uuid```)