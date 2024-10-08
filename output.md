# ApacheBench Performance Comparison

## Test 1: Without Singleflight

```
b -n 1000 -c 100 -vvv -A user1:password1 http://localhost:8080/login
```

## Test 2: With Singleflight

```
ab -n 1000 -c 100 -vvv -A user1:password1 http://localhost:8080/login-singleflight
```

| Metric | Without Singleflight | With Singleflight |
|--------|----------------------|-------------------|
| Requests per second (mean) | 116.34 | 517.15 |
| Time per request (mean) | 859.552 ms | 193.368 ms|
| Time per request (mean, across all concurrent requests) | 8.596 ms | 1.934 ms|
| Requests per second | 116.34 [#/sec] (mean) | 517.15 [#/sec] (mean) |
| 50% of requests served within | 781 ms | 175 ms |
| 75% of requests served within | 950 ms | 176 ms |
| 99% of requests served within | 1664 ms | 178 ms |
| 100% of requests served within | 1759 ms | 188 ms |

 

