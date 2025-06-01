#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
DURATION=60  # Test duration in seconds
CONCURRENT_USERS=10

echo "🔥 Mixed Workload Test - ${DURATION}s duration, ${CONCURRENT_USERS} users"

# Pre-create some users for login testing
echo "Creating test users..."
for i in $(seq 1 5); do
    curl -s -X POST $BASE_URL/auth/register \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"testuser${i}@example.com\",
            \"password\": \"password123\",
            \"display_name\": \"Test User ${i}\"
        }" > /dev/null
done

# Mixed workload function
mixed_workload() {
    user_id=$1
    end_time=$(($(date +%s) + DURATION))
    request_count=0
    
    while [ $(date +%s) -lt $end_time ]; do
        request_count=$((request_count + 1))
        
        # 40% registration, 40% login, 20% health checks
        rand=$((RANDOM % 10))
        
        if [ $rand -lt 4 ]; then
            # Registration
            curl -s -w "User$user_id-R$request_count: %{http_code} (%{time_total}s)\n" \
                -X POST $BASE_URL/auth/register \
                -H "Content-Type: application/json" \
                -d "{
                    \"email\": \"mixed${user_id}_${request_count}@example.com\",
                    \"password\": \"password123\",
                    \"display_name\": \"Mixed User ${user_id}_${request_count}\"
                }" -o /dev/null
        elif [ $rand -lt 8 ]; then
            # Login
            curl -s -w "User$user_id-L$request_count: %{http_code} (%{time_total}s)\n" \
                -X POST $BASE_URL/auth/login \
                -H "Content-Type: application/json" \
                -d "{
                    \"email\": \"testuser$((user_id % 5 + 1))@example.com\",
                    \"password\": \"password123\",
                    \"device_id\": \"load-test-${user_id}\"
                }" -o /dev/null
        else
            # Health check
            curl -s -w "User$user_id-H$request_count: %{http_code} (%{time_total}s)\n" \
                -X GET http://localhost:8080/health -o /dev/null
        fi
        
        # Small delay to simulate real usage
        sleep 0.1
    done
    
    echo "User $user_id completed $request_count requests"
}

# Start concurrent users
for user in $(seq 1 $CONCURRENT_USERS); do
    mixed_workload $user &
done

wait
echo "✅ Mixed workload test completed!"