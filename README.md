# Golang + gRPC + Protocol Buffers API

### HTTP2 和 HTTP1.1 比较
    https://imagekit.io/demo/http2-vs-http1

### 安装
    go get -u google.golang.org/grpc
    go get -u github.com/golang/protobuf/protoc-gen-go

### Four kinds of service method
![](http://qiniu.rocbj.com/Jietu20190923-155613.png)

##### Unary
Unary RPC calls are the basic Request/Response
The client will send one message to the server and will receive one  response from the server.

##### Server Streaming
The client will send one message to the server and will receive many  response from the server.

server streaming are well suited for
* when the server needs to send a lot of data (big data) 
* when the server needs to "PUSH" data to the client without having the client request for more (think live feed, chat etc)

##### Client Streaming
The client will send many message to the server and will receive one response from the server.
client streaming are well suited for
* when the client needs to send a lot of data (big data)
* when the server processing is expensive and should happen as the client sends data
* when the client needs to "PUSH" data to the server without really expecting a response

##### Bi Directional Streaming
The client will send many message to the server and will receive many response from the server.

the number of requests and responses does not have to match.

Bi Directional Streaming are well suited for
* when the client and the server needs to send a lot of data asynchronously
* "Chat" protocol
* Long running connections

##### Test CRUD
Evans CLI

##### 参考连接
* https://grpc.io/
* http://avi.im/grpc-errors/
* https://github.com/ktr0731/evans
