В проекте используется [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool), это защищенный от параллелизма пул соединений для pgx, библиотеки для работы с PostgreSQL.
По умолчанию максимальное количество соединений равно 10.

1. Запустил тест на 1 одновременное подключение в течение 5 минут:

```
  scenarios: (100.00%) 1 scenario, 1 max VUs, 5m30s max duration (incl. graceful stop):
           * default: 1 looping VUs for 5m0s (gracefulStop: 30s)

     ✓ status was 200

     █ setup

     checks.........................: 100.00% ✓ 527     ✗ 0  
     data_received..................: 81 kB   268 B/s
     data_sent......................: 90 kB   300 B/s
     http_req_blocked...............: avg=5.93µs   min=3.11µs   med=4.85µs   max=328.82µs p(90)=6.06µs   p(95)=7.23µs  
     http_req_connecting............: avg=334ns    min=0s       med=0s       max=176.33µs p(90)=0s       p(95)=0s      
     http_req_duration..............: avg=569.78ms min=440.43ms med=554.85ms max=961.41ms p(90)=653.91ms p(95)=689.02ms
       { expected_response:true }...: avg=569.78ms min=440.43ms med=554.85ms max=961.41ms p(90)=653.91ms p(95)=689.02ms
     http_req_failed................: 0.00%   ✓ 0       ✗ 527
     http_req_receiving.............: avg=88.81µs  min=58.7µs   med=82.66µs  max=306.29µs p(90)=112.91µs p(95)=122.54µs
     http_req_sending...............: avg=23.23µs  min=13.77µs  med=21.67µs  max=161.29µs p(90)=29.16µs  p(95)=31.6µs  
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s       max=0s       p(90)=0s       p(95)=0s      
     http_req_waiting...............: avg=569.67ms min=440.34ms med=554.74ms max=961.31ms p(90)=653.81ms p(95)=688.92ms
     http_reqs......................: 527     1.75441/s
     iteration_duration.............: avg=568.89ms min=22.96µs  med=555ms    max=961.6ms  p(90)=654.04ms p(95)=689.07ms
     iterations.....................: 527     1.75441/s
     vus............................: 1       min=1     max=1
     vus_max........................: 1       min=1     max=1

started: Sun Apr 23 2023 22:22:11 GMT+0300 (MSK)
ended: Sun Apr 23 2023 22:27:11 GMT+0300 (MSK)

running (5m00.4s), 0/1 VUs, 527 complete and 0 interrupted iterations
default ✓ [======================================] 1 VUs  5m0s
```

