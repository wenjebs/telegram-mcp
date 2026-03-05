inspect:
    -lsof -ti:6277,6274 | xargs kill -9
    npx @modelcontextprotocol/inspector ./bin/telegram-mcp
