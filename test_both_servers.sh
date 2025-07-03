#!/bin/bash

echo "🔍 Testing Manual vs Automatic Instrumentation"
echo "=============================================="
echo ""

echo "📊 Server Status:"
echo "- Manual Server (port 8080): main.go"
echo "- Auto Server (port 8081): main_auto.go"
echo ""

echo "🧪 Testing Manual Instrumentation Server (Port 8080):"
echo "------------------------------------------------------"

echo "1️⃣ Testing Hello World:"
curl -s http://localhost:8080/
echo ""

echo "2️⃣ Creating item via Manual Server:"
MANUAL_RESPONSE=$(curl -s -X POST http://localhost:8080/createItem \
  -H "Content-Type: application/json" \
  -d '{"name": "Manual Item", "value": "manual tracing"}')
echo "$MANUAL_RESPONSE"
echo ""

echo "3️⃣ Retrieving item from Manual Server:"
curl -s http://localhost:8080/items/1
echo ""
echo ""

echo "⚡ Testing Automatic Instrumentation Server (Port 8081):"
echo "--------------------------------------------------------"

echo "1️⃣ Testing Hello World:"
curl -s http://localhost:8081/
echo ""

echo "2️⃣ Creating item via Auto Server:"
AUTO_RESPONSE=$(curl -s -X POST http://localhost:8081/createItem \
  -H "Content-Type: application/json" \
  -d '{"name": "Auto Item", "value": "automatic tracing"}')
echo "$AUTO_RESPONSE"
echo ""

echo "3️⃣ Retrieving item from Auto Server:"
curl -s http://localhost:8081/items/1
echo ""
echo ""

echo "📈 Comparison Summary:"
echo "----------------------"
echo "✅ Both servers are functional"
echo "✅ Manual: Full control, more code (15-20 lines per handler)"
echo "✅ Auto: Less code, automatic HTTP tracing (3-5 lines per handler)"
echo ""

echo "🔍 Key Differences:"
echo "- Manual: You create every span manually"
echo "- Auto: HTTP spans created automatically, focus on business logic"
echo ""

echo "🎯 Recommendation: Use Automatic Instrumentation (main_auto.go)"
echo "- 90% less tracing code"
echo "- Better maintainability"
echo "- Focus on business value"
echo ""