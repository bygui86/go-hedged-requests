
# go-tail-latency
Tail latency sample project in Golang

7 techniques to deal with tail latency:

- Hedged requests
- Tied requests
- Micro partitions
- Selective replication
- Latency-induced probation
- Good enough responses
- Canary requests

## run

### hedged requests

```shell script
# in the first shell
cd server
go build
./server -port=8090
./server -port=8091
./server -port=8092

# in the second shell
cd client
go build
./client

# in the third shell
curl localhost:8080/simple
curl localhost:8080/fanout
curl localhost:8080/hedged
```

## load test with [vegeta](https://github.com/tsenart/vegeta)

### server
```shell script
# in the first shell
cd server
go build
./server

# in the second shell
echo "GET http://localhost:8090/ishealthy" | \
    vegeta attack -duration 5s -rate 100/1s -workers 10 -connections 10 --insecure | \
    tee results.bin | \
    vegeta report

# look at the line "Latencies"
# in my case:   Latencies     [min, mean, 50, 90, 95, 99, max]  15.257ms, 19.454ms, 16.572ms, 17.196ms, 17.373ms, 116.514ms, 117.112ms
```

### client with fanout

```shell script
# in the first shell
cd server
go build
./server -port=8090
./server -port=8091
./server -port=8092

# in the second shell
cd client
go build
./client > client-fanout.log

# in the third shell
echo "GET http://localhost:8080/fanout" | \
    vegeta attack -duration 5s -rate 100/1s -workers 10 -connections 10 --insecure | \
    tee results.bin | \
    vegeta report

# look at the line "Latencies"
# in my case:   Latencies     [min, mean, 50, 90, 95, 99, max]  15.545ms, 17.056ms, 16.879ms, 17.73ms, 17.892ms, 18.113ms, 117.263ms

# count total requests
wc -l client-fanout.log
# in my case:   2001
```

### client with hedged

```shell script
# in the first shell
cd server
go build
./server -port=8090
./server -port=8091
./server -port=8092

# in the second shell
cd client
go build
./client > client-hedged.log

# in the third shell
echo "GET http://localhost:8080/hedged" | \
    vegeta attack -duration 5s -rate 100/1s -workers 10 -connections 10 --insecure | \
    tee results.bin | \
    vegeta report

# look at the line "Latencies"
# in my case:   Latencies     [min, mean, 50, 90, 95, 99, max]  15.588ms, 17.238ms, 16.934ms, 17.679ms, 17.883ms, 33.792ms, 35.123ms

# count total requests
wc -l client-hedged.log
# in my case:   1351
```

## links

- https://medium.com/swlh/hedged-requests-tackling-tail-latency-9cea0a05f577
- https://blog.golang.org/concurrency-timeouts
