local job = redis.call("RPOP",KEY[1])
if not job then
    return nil 
end 

local decoded = cjson.decode(job)
local job_id = decoded.job_id

local running_key = KEYS[2] .. job_id

redis.call("HSET", running_key,
  "job_id", job_id,
  "app_id", decoded.app_id,
  "job_type", decoded.type,
  "worker_id", ARGV[1],
  "retry", decoded.retry or 0,
  "max_retry", decoded.max_retry or 0,
  "started_at", ARGV[3],
  "last_beat", ARGV[3]
)

redis.call("EXPIRE",running_key,ARGV[2])

return job 