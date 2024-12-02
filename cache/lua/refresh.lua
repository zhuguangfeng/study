if redis.call('get',KYES[1])==ARGV[1] then
    return redis.call('expire',KEYS[1],ARGV[2])
else
    return 0
end