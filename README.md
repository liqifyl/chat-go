<h1>用go开发的一个简单聊天服务端，包括用户、好友、朋友圈、聊天功能，对外输出API都是http协议(暂不支持https,rpc暂也不支持后续再考虑)，日志收集系统支持graylog, graylog客户端使用udp gelf

# 准备工作

## mysql准备工作
```sql
CREATE DATABASE xx;
```
## 创建表
```sql
//用户表
create table user(
	id BIGINT AUTO_INCREMENT,
	nick varchar(200),
	password varchar(30),
	age tinyint,
	birthday datetime,
	sign varchar(100),
	country varchar(20),
	sex tinyint,
	rtime datetime,
	pnumber char(11),
	primary key(id)
);

//通讯录表
create table friend (
	id BIGINT AUTO_INCREMENT primary key,
	uid BIGINT NOT NULL,
	fid BIGINT NOT NULL,
	etime datetime NOT NULL,
	fnick varchar(200) NOT NULL,
	foreign key(uid) REFERENCES user(id),
	foreign key(fid) REFERENCES user(id)
);

//朋友圈表
create table friend_circle (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	uid BIGINT NOT NULL,
	ptime datetime,
	title varchar(200),
	url varchar(200),
	foreign key(uid) REFERENCES user(id)
);

```

## graylog搭建
请参考[graylog单机搭建](https://cloud.tencent.com/developer/article/1628850)进行部署
### 在部署中注意事项
1.由于elasticsearch无法以root模式运行，所以需要创建elasticesearch组和用户运行elasticsearch;第二种方式就是以root方式运行并且将elasticsearch运行时需要用到的文件夹全部chown -R到
elasticsearch:elasticsearch
2.如果需要将graylog web需要以公网方式输出，请修改/etc/graylog/server/server.conf中的http_publish_uri参数，将其配置外网web地址
3.如何开启gelf udp监听端口，通过graylog web页面system->inputs->gelf udp启动

# 编译输出bin

