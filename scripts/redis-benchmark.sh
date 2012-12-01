#!/usr/bin/env bash
echo "Direct redis"
redis-benchmark -q -p 6379 -c 10 -n 1000 | tee benchmark.redis.txt
echo "Proxied redis"
redis-benchmark -q -p 9999 -c 10 -n 100 | tee benchmark.redproxy.txt
