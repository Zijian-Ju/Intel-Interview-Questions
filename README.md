# 英特尔面试题

## 基本信息
- 姓名： 鞠子健
- 意向职位：GoLang后端开发
- 毕业时间：2021年12月
- 毕业院校：墨尔本大学（软件工程 硕士）
- 本科院校：哈尔滨工业大学 材料科学与工程 本科）

## Solutionns
### Question 1
基于errgroup实现一个http server的启动和关闭，以及linux signal 信号的注册和处理，要保证能够一个退出，全部注销退出。

**解答：** 
根据题目要求，在初始化server后，我注册了两个handler function，分别是：
- **WelcomeHandler** 用于模拟正常的http请求，路径：localhost: ${port}/
- **TerminationHandler** 服务器接到请求后会关闭，用于模拟服务器内部发生错误，意外退出，路径：localhost：${port}/terminate

同时会在errgroup里注册三个goroutine，每个goroutine都有对应的defer函数，会打印出goroutine的退出信息，用于证明goroutine的返回
- Goroutine 1：用于跑http server
- Goroutine 2：用于在server接收到terminate请求或接收到ctx.Done()信号时将server关闭，并返回
- Goroutine 3：用于在接收到系统的关闭信号或接收到ctx.Done()信号时返回

在server接收到terminate请求时，会向serverSigChan channel里传递信号，goroutine 2接收到信号后，会关闭server，并返回，这里操作会触发ctx.Done(), 使得goroutine 3也返回。
程序收到linux信号时同理，会打印退出信息后返回并触发goroutine 2的ctx.Done, 同样使得goroutine 2关闭server后返回。

**开始**

编译：
```
make build
```
运行(并指定端口号，不指定时默认在8081运行)
```
./question1 8092
```

服务端输出
```
2021/08/04 15:52:02 Goroutine 1 is starting server listening on 8092
```
**正常请求**

请求
```
curl 'localhost:8092'
```
服务端输出
```
Hi, I am Zijain. Thanks for assessing my interview questions!
```
**接收到系统中止信号时的输出（ctrl + c）**
```
2021/08/04 15:52:53 Goroutine 1 is starting server listening on 8092
2021/08/04 15:53:03 Goroutine 3 has returned due to: Received os signal(interrupt)
2021/08/04 15:53:03 Goroutine 1 (HTTP server) has been closed
2021/08/04 15:53:04 Goroutine 2 has been closed due to: Errgroup is called to be terminated
2021/08/04 15:53:04 Main process is existing due to:  Received os signal(interrupt)
```

**接收到terminate请求**

请求 
```
curl 'localhost:8092/terminate'
```

客户端输出
```
I am going to stop the server and all goroutines
```
服务器端输出
```
2021/08/04 15:54:39 Goroutine 1 (HTTP server) has been closed
2021/08/04 15:54:39 Goroutine 3 has returned due to: Something wrong happened in server
2021/08/04 15:54:40 Goroutine 2 has been closed due to: Something wrong happening in server
2021/08/04 15:54:40 Main process is existing due to:  http: Server closed
```

## Question 2
我们在数据库操作的时候，比如Data Access Object(DAO)层中当遇到一个sql.ErrNoRows的时候，是否应该wrap这个error,抛给上层。为什么？应该怎么做，请写出代码。

**解答**

我们应该返回给上层，但是返回应该区别于一般的err。这是因为对于业务层而言，根据业务场景的不同，对于查询的记录不存在的处理逻辑也可能不同，有时可能属于正常现象，有的则可能需要特殊处理。但是无论如何我们不应将sql语句返回的err直接抛给上层，由上层识别，这不利于业务层和持久层的解耦， 可以通过在dao层定义error的枚举类型的方式来处理，当查询不到记录时返回这个定义在dao层的error，并在上层对这个error进行识别。
同时还可以按照查询传参的类型来处理，例如现在广泛使用的orm包gorm的处理方式是赋值的类型是slice时不报这个错误，单结构体时则会抛出这个异常。

代码示例
```go
var RecordNotFoundErr Error = errors.New("Record not found")

func QuerySomethingById(id string, obj *AnyStruct) error {
	row, err := database.Query("Select * from ....")
	row.Scan(obj)
	if err != nil {
	    if err == sql.ErrNoRows {
	        err = RecordNotFoundErr 	
        }
        return err
    }
    return  nil
}
```



