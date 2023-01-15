# Ygaros Discovery Server

Basic discovery server for microservice architecture with `rand.Intn` "load balancer :D".

[**CLIENT**](https://github.com/ygaros/discovery-client)

*The idea of this project is to try mimic the `Eureka-Server` from `Spring-Cloud` in **Go**.*

### To get started

```
go get github.com/ygaros/discovery-server
```

*Fully functional **gRPC** discovery-server with in memory storing and horizonal scaling enabled.*

This means that on default multi-value-map is used for storing registered service instances and on `rand.Intn` the particular instance is given on request. 


```
func main(){
    server.NewServer()
}
```

<sup>*There is slice implementation ready to disable registering multiple service instances*</sup>

### Default timers

*Every 90 seconds after registration service instance is considered unhealthy and its deleted.*

