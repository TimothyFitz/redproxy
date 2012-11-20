import redis.client

def unproxied_redis():
    return redis.client.StrictRedis(port=6379)

def proxied_redis():
    return redis.client.StrictRedis(port=9999)
