# Experiment 
Goal of this experiment is to compare the performance of a singleflight and without singleflight in cache hydration counter.

The TTL of the cache is 1 millisecond to simulate a cache that is always expired and needs to be hydrated
from the main datastore (dynamodb in this case).

The experiment demonstrates that singleflight can significantly reduce the number of cache hydration requests and improve the performance of the system
compared to not using singleflight where the number of cache hydration requests is very high.

The experiment was done with 100 concurrent connections and 100000 requests.

# Command: Benchmarking without singleflight

```
ab -n 100000 -c 100 http://127.0.0.1:8080/customer/868f4667-0707-4391-9fe3-f479f9be1952
```

# Command: Benchmarking with singleflight

```
ab -n 100000 -c 100 http://127.0.0.1:8080/customer/868f4667-0707-4391-9fe3-f479f9be1952/singleflight
```

# Results
| Metric | Without Singleflight | With Singleflight |
|--------|----------------------|-------------------|
| cache_hydration_counter | 14358 | 903 |
| Requests per second (mean) | 14578.44 | 30070.35 |
| Time per request (mean) | 6.859 ms | 3.326 ms |
| go_memstats_alloc_bytes (Number of bytes allocated in heap) | 7.736496e+06 | 3.248128e+06 |
| p50 | 2ms | 3ms |
| p66 | 3ms | 3ms |
| p75 | 4ms | 4ms |
| p80 | 5ms | 5ms |
| p90 | 27ms | 4ms |
| p95 | 38ms | 5ms |
| p98 | 48ms | 6ms |
| p99 | 57ms | 7ms |
| 100% | 154ms | 87ms |
