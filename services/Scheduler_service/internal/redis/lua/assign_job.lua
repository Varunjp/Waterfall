local running = tonumber(redis.call("GET", KEYS[1]) or "0")
local limit = tonumber(ARGV[1])

if running >= limit then
  return {err="CONCURRENCY_LIMIT"}
end

redis.call("INCR", KEYS[1])

redis.call("XADD", KEYS[2], "*",
  "job_id", ARGV[2],
  "payload", ARGV[3],
  "attempt", ARGV[4]
)

return "OK"