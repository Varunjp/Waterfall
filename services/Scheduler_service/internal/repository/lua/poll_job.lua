local job = redis.call("RPOP", KEYS[1])
if not job then
    return nil 
end 

local decoded = cjson.decode(job)
local job_id = decoded.JobID

if not job_id then
  return redis.error_reply("JobID missing in job payload")
end

local running_key = KEYS[2] .. job_id

redis.call("HSET", running_key,
  "job_id", job_id,
  "app_id", decoded.AppID,
  "job_type", decoded.Type,
  "worker_id", ARGV[1],
  "retry", decoded.Retry or 0,
  "max_retry", decoded.MaxRetries or 0,
  "started_at", ARGV[3],
  "last_beat", ARGV[3]
)

redis.call("EXPIRE",running_key,ARGV[2])

return job 