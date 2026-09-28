#!/bin/sh
echo "🛑 Stopping TailRouter..."
pkill -f "tailrouter" 2>/dev/null || killall tailrouter 2>/dev/null || true
pkill -f "server.py" 2>/dev/null || true
sleep 1
echo "✅ TailRouter stopped."
