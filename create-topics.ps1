Write-Host "Waiting for Redpanda..."

while ($true) {
    docker exec redpanda rpk cluster info *> $null
    if ($LASTEXITCODE -eq 0) { break }
    Start-Sleep -Seconds 2
}

Write-Host "Creating topics..."

docker exec redpanda rpk topic create job_requests --partitions 3 --replicas 1 2>$null
docker exec redpanda rpk topic create scheduler_queue --partitions 3 --replicas 1 2>$null
docker exec redpanda rpk topic create retry_queue --partitions 3 --replicas 1 2>$null
docker exec redpanda rpk topic create worker_jobs --partitions 1 --replicas 1 2>$null
docker exec redpanda rpk topic create job_status --partitions 1 --replicas 1 2>$null

Write-Host "Done"
