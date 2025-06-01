#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
CONCURRENT_USERS=20
REQUESTS_PER_USER=10

echo "🚀 Starting Load Test: $CONCURRENT_USERS concurrent users, $REQUESTS_PER_USER requests each"
echo "Total requests: $((CONCURRENT_USERS * REQUESTS_PER_USER))"
echo "=============================================="

# Function to run load test for one user
run_user_load() {
    user_id=$1
    start_time=$(date +%s)
    success_count=0
    error_count=0
    
    for i in $(seq 1 $REQUESTS_PER_USER); do
        response=$(curl -s -w "%{http_code}:%{time_total}" -X POST $BASE_URL/auth/register \
            -H "Content-Type: application/json" \
            -d "{
                \"email\": \"loadtest${user_id}_${i}@example.com\",
                \"password\": \"password123\",
                \"display_name\": \"Load Test User ${user_id}_${i}\"
            }" 2>/dev/null)
        
        http_code=$(echo $response | cut -d: -f1)
        response_time=$(echo $response | cut -d: -f2)
        
        if [[ $http_code =~ ^2[0-9][0-9]$ ]]; then
            success_count=$((success_count + 1))
        else
            error_count=$((error_count + 1))
        fi
        
        echo "User $user_id Request $i: HTTP $http_code (${response_time}s)"
    done
    
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    echo "User $user_id completed: $success_count success, $error_count errors in ${duration}s"
}

# Start concurrent users
for user in $(seq 1 $CONCURRENT_USERS); do
    run_user_load $user &
done

# Wait for all background jobs to complete
wait

echo "✅ Load test completed!"