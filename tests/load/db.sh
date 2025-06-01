#!/bin/bash

echo "💾 DATABASE PERFORMANCE ANALYSIS"
echo "================================"

# Check database performance
echo "📊 Database Connection Performance:"
for i in {1..10}; do
    start_time=$(date +%s%3N)
    docker exec chat-platform-postgres-1 psql -U chatuser -d chatdb -c "SELECT 1;" > /dev/null 2>&1
    end_time=$(date +%s%3N)
    duration=$((end_time - start_time))
    echo "   Connection $i: ${duration}ms"
done

echo ""
echo "📈 Database Statistics:"
docker exec chat-platform-postgres-1 psql -U chatuser -d chatdb -c "
SELECT 
    'Total Users' as metric,
    COUNT(*) as value
FROM users
UNION ALL
SELECT 
    'Active Sessions' as metric,
    COUNT(*) as value
FROM user_sessions
WHERE expires_at > NOW()
UNION ALL
SELECT 
    'Users Created Today' as metric,
    COUNT(*) as value
FROM users 
WHERE created_at >= CURRENT_DATE;
"

echo ""
echo "⚡ Query Performance Test:"
start_time=$(date +%s%3N)
docker exec chat-platform-postgres-1 psql -U chatuser -d chatdb -c "
SELECT u.email, u.display_name, us.device_id 
FROM users u 
LEFT JOIN user_sessions us ON u.id = us.user_id 
LIMIT 100;" > /dev/null
end_time=$(date +%s%3N)
duration=$((end_time - start_time))
echo "   Complex JOIN query (100 records): ${duration}ms"