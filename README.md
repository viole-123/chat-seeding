# Uniscore Seeding Bot (scaffold)

This repository is a scaffold for a seeding bot project. Fill in implementations.

---

## 🚀 Quick Test Guide

### Start Bot & Dashboard
```bash
# Terminal 1: Bot
go build -o seeding-bot.exe ./cmd/seeding-bot
.\seeding-bot.exe

# Terminal 2: Web Dashboard  
cd web
python -m http.server 3000
# Open http://localhost:3000/index.html

# Terminal 3: Send test events
cd tests
go run producer.go
```

### Sample JSON (1 dòng để test)

**PREMATCH event:**
```
{"eventID":"evt001","matchID":"match001","leagueID":"league001","roomID":"room-match001","phase":"PREMATCH","offsetMinutes":-30,"eventType":"PREMATCH","info":"Match sắp bắt đầu","timestamp":1709712000}
```

**GOAL event:**
```
{"match_id":"match001","league_id":"league001","type":"GOAL","minute":23,"player_name":"Bruno Fernandes","position":10,"home_score":1,"away_score":0,"home_team":"Man United","away_team":"Arsenal","timestamp":1709714000,"room_id":"room-match001","event_type":"GOAL"}
```

**RED_CARD event:**
```
{"match_id":"match001","league_id":"league001","type":"RED_CARD","minute":58,"player_name":"Gabriel","position":6,"home_score":1,"away_score":0,"home_team":"Man United","away_team":"Arsenal","timestamp":1709716000,"room_id":"room-match001","event_type":"RED_CARD"}
```
