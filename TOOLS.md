# TOOLS.md - Local Notes

Skills define _how_ tools work. This file is for _your_ specifics — the stuff that's unique to your setup.

### Telegram

- João Vitor Guidi → [Chat ID não salvo - precisa adicionar]

### Camera Snapshot (cam-snap)

- **Serviço:** `cam-snap.service` (systemd, ativo) — Go, porta 9378, só localhost
- **Código:** `/root/clawd/scripts/cam-snap/` (main.go + binário)
- **Uso quando o João pedir foto da câmera:**
  - Foto nova: `curl -s -o /root/clawd/cam_fresh.jpg 127.0.0.1:9378/fresh` (~400ms)
  - Aceita cache: `curl -s -o f.jpg "127.0.0.1:9378/snap?max_age=10s"` (~10ms)
  - Depois anexar com `MEDIA:/root/clawd/cam_fresh.jpg`
- **Captura:** HTTP `http://thingino:thingino@192.168.100.64/image.jpg` (fallback RTSP+ffmpeg)
- **Cache:** aquecido a cada 30s; `/last` = frame em cache instantâneo; `/healthz` = status

## What Goes Here

Things like:

- Camera names and locations
- SSH hosts and aliases
- Preferred voices for TTS
- Speaker/room names
- Device nicknames
- Anything environment-specific

## Examples

```markdown
### Home Assistant

- **URL:** http://127.0.0.1:8123 (host network, Docker)
- **Token:** /root/clawd/.ha_token (long-lived access token)
- **API:** curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8123/api/...

### Lâmpadas (Tuya Blaupunkt Smart E27 MC)

- 4 lâmpadas: `light.blaupunkt_smart_e27_mc` [+ _2, _3, _4]
- Plataforma: Tuya (via Home Assistant)
- Color temp: 2000K–6500K | Modos: color_temp, hs
- API: POST /api/services/light/turn_on {"entity_id": "...", "color_temp_kelvin": 2000}

### Cameras

- living-room → Main area, 180° wide angle
- front-door → Entrance, motion-triggered

### SSH

- home-server → 192.168.1.100, user: admin

### TTS

- Preferred voice: "Nova" (warm, slightly British)
- Default speaker: Kitchen HomePod
```

## Why Separate?

Skills are shared. Your setup is yours. Keeping them apart means you can update skills without losing your notes, and share skills without leaking your infrastructure.

---

Add whatever helps you do your job. This is your cheat sheet.
