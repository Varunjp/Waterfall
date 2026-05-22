local running = tonumber(redis.call("GET", KEYS[1]) or "0")

redis.call("INCR", KEYS[1])

redis.call("XADD", KEYS[2], "*",
  "job_id", ARGV[1],
  "payload", ARGV[2],
  "attempt", ARGV[3]
)

return "OK"
