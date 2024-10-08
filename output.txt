// ==============================================================================
// ab -n 1000 -c 100 -vvv -A user1:password1 http://localhost:8080/login
// ==============================================================================

This is ApacheBench, Version 2.3 <$Revision: 1913912 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Completed 600 requests
Completed 700 requests
Completed 800 requests
Completed 900 requests
Completed 1000 requests
Finished 1000 requests


Server Software:        
Server Hostname:        localhost
Server Port:            8080

Document Path:          /login
Document Length:        16 bytes

Concurrency Level:      100
Time taken for tests:   8.596 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      133000 bytes
HTML transferred:       16000 bytes
Requests per second:    116.34 [#/sec] (mean)
Time per request:       859.552 [ms] (mean)
Time per request:       8.596 [ms] (mean, across all concurrent requests)
Transfer rate:          15.11 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    1   1.5      0      10
Processing:   183  809 274.0    780    1759
Waiting:      181  806 273.9    778    1758
Total:        183  810 273.8    781    1759

Percentage of the requests served within a certain time (ms)
  50%    781
  66%    884
  75%    950
  80%   1010
  90%   1155
  95%   1326
  98%   1577
  99%   1664
 100%   1759 (longest request)

// ==============================================================================
// ab -n 1000 -c 100 -vvv -A user1:password1 http://localhost:8080/login-singleflight
// ==============================================================================

This is ApacheBench, Version 2.3 <$Revision: 1913912 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking localhost (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Completed 600 requests
Completed 700 requests
Completed 800 requests
Completed 900 requests
Completed 1000 requests
Finished 1000 requests


Server Software:        
Server Hostname:        localhost
Server Port:            8080

Document Path:          /login-singleflight
Document Length:        16 bytes

Concurrency Level:      100
Time taken for tests:   1.934 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      133000 bytes
HTML transferred:       16000 bytes
Requests per second:    517.15 [#/sec] (mean)
Time per request:       193.368 [ms] (mean)
Time per request:       1.934 [ms] (mean, across all concurrent requests)
Transfer rate:          67.17 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    7   1.3      7       8
Processing:   160  168   2.7    168     188
Waiting:      160  168   2.7    168     186
Total:        166  174   2.8    175     188

Percentage of the requests served within a certain time (ms)
  50%    175
  66%    176
  75%    176
  80%    177
  90%    177
  95%    177
  98%    178
  99%    178
 100%    188 (longest request)
 

