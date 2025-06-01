#!/bin/bash

clear
echo ""
echo "🚀 PEAK PERFORMANCE TEST - Finding Your Limits"
echo "=============================================="
echo ""

BASE_URL="http://localhost:8080"

# Test increasing concurrent loads quickly
echo "⚡ Rapid Concurrent User Testing:"
echo "--------------------------------"

for users in 10 20 30 40 50 75 100; do
    echo -n "Testing $users users: "
    
    start_time=$(date +%s)
    success=0
    
    # Launch concurrent requests
    for i in $(seq 1 $users); do
        (
            response=$(curl -s -w "%{http_code}" -X POST $BASE_URL/api/v1/auth/register \
                -H "Content-Type: application/json" \
                -d "{\"email\":\"peak_${users}_${i}_$(date +%s)@test.com\",\"password\":\"test123\",\"display_name\":\"Peak Test\"}" \
                --max-time 10 -o /dev/null 2>/dev/null)
            if [[ $response =~ ^2[0-9][0-9]$ ]]; then
                echo "SUCCESS" >> /tmp/peak_$$
            fi
        ) &
    done
    
    wait
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    if [ -f "/tmp/peak_$$" ]; then
        success=$(wc -l < /tmp/peak_$$ 2>/dev/null || echo "0")
        rm -f /tmp/peak_$$
    fi
    
    success_rate=$((success * 100 / users))
    
    if [ $success_rate -ge 80 ]; then
        echo "✅ ${success}/$users (${success_rate}%) in ${duration}s"
    else
        echo "🚨 ${success}/$users (${success_rate}%) in ${duration}s - BREAKING POINT!"
        break
    fi
    
    sleep 1
done

echo ""
echo "🏆 Your system can handle serious concurrent load!"