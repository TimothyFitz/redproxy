#!/usr/bin/env bash
redis-benchmark -p 6379 -c 10 -n 1000 | tee benchmark.redis.txt
redis-benchmark -p 9999 -c 10 -n 100 | tee benchmark.redproxy.txt

