# HealthConnect — Technical Architecture Document
**Go Backend · React Vite PWA**
Version 1.0 · Hackathon Build

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Repository Structure](#2-repository-structure)
3. [Backend — Go](#3-backend--go)
4. [Frontend — React Vite PWA](#4-frontend--react-vite-pwa)
5. [AI & Multimodal Integration](#5-ai--multimodal-integration)
6. [Voice Pipeline](#6-voice-pipeline)
7. [Animated Doctor Avatar](#7-animated-doctor-avatar)
8. [Camera Scan Feature](#8-camera-scan-feature)
9. [Pharmacy Map & Geolocation](#9-pharmacy-map--geolocation)
10. [Email Transcript](#10-email-transcript)
11. [Mocked Drug Reservation](#11-mocked-drug-reservation)
12. [Safety — Disclaimers & Emergency Escalation](#12-safety--disclaimers--emergency-escalation)
13. [Data — Seeded Pharmacies](#13-data--seeded-pharmacies)
14. [PWA Configuration](#14-pwa-configuration)
15. [Environment Variables](#15-environment-variables)
16. [API Reference](#16-api-reference)
17. [Local Development Setup](#17-local-development-setup)
18. [Deployment Notes](#18-deployment-notes)
19. [Demo Runbook](#19-demo-runbook)

---

## 1. System Overview

```
Browser (React Vite PWA)
        │
        │  HTTPS / REST + SSE
        ▼
Go HTTP Server  (Chi router)
        │
        ├──► Anthropic Claude API  (chat + vision)
        ├──► Resend API            (email transcript)
        └──► Seeded JSON           (pharmacy data, no live DB)
```

**Key design principles for the hackathon:**

- All pharmacy data is bundled as static JSON — zero live DB dependency during the demo.
- AI calls are the only external HTTP dependency; a canned "fallback response" mode can be toggled for offline demos.
- The Go backend is a single binary (`healthconnect-server`). No Docker required to run locally.
- HTTPS is served via a self-signed cert locally and a real cert in staging (required for browser mic/camera).

---

## 2. Repository Structure

```
healthconnect/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go            # Entry point
│   ├── internal/
│   │   ├── ai/
│   │   │   ├── client.go          # Anthropic SDK wrapper
│   │   │   ├── prompts.go         # System prompts, red-flag list
│   │   │   └── vision.go          # Image analysis handler
│   │   ├── email/
│   │   │   └── resend.go          # Resend API integration
│   │   ├── pharmacy/
│   │   │   ├── store.go           # In-memory pharmacy store
│   │   │   └── seed.go            # Loads data/pharmacies.json
│   │   └── server/
│   │       ├── router.go          # Chi routes
│   │       ├── middleware.go      # CORS, logger, recovery
│   │       └── handlers.go        # HTTP handlers
│   ├── data/
│   │   └── pharmacies.json        # Seeded Lagos pharmacy data
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── public/
│   │   ├── manifest.webmanifest
│   │   └── icons/                 # PWA icons
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── assets/
│   │   │   └── avatar/            # Lottie JSON files
│   │   ├── components/
│   │   │   ├── Landing.tsx        # Voice / Text choice screen
│   │   │   ├── VoiceSession.tsx   # Voice mode UI
│   │   │   ├── TextChat.tsx       # Text mode UI
│   │   │   ├── DoctorAvatar.tsx   # Animated avatar (Lottie)
│   │   │   ├── CameraScan.tsx     # Camera capture + scan
│   │   │   ├── PharmacyMap.tsx    # Leaflet map + list
│   │   │   ├── Transcript.tsx     # Running transcript panel
│   │   │   ├── EmailPanel.tsx     # Email transcript form
│   │   │   ├── Disclaimer.tsx     # One-time modal + inline badge
│   │   │   ├── EmergencyAlert.tsx # Red-flag notice
│   │   │   └── DrugReservation.tsx# Mocked reservation flow
│   │   ├── hooks/
│   │   │   ├── useSpeechRecognition.ts
│   │   │   ├── useSpeechSynthesis.ts
│   │   │   └── useGeolocation.ts
│   │   ├── lib/
│   │   │   ├── api.ts             # Typed fetch wrappers
│   │   │   ├── redFlags.ts        # Red-flag keyword list (client-side pre-check)
│   │   │   └── constants.ts
│   │   ├── store/
│   │   │   └── session.ts         # Zustand session state
│   │   └── types/
│   │       └── index.ts
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
│
└── README.md
```

---

## 3. Backend — Go

### 3.1 Stack

| Concern | Choice | Reason |
|---|---|---|
| HTTP router | `go-chi/chi v5` | Lightweight, middleware-friendly |
| AI client | `anthropics-sdk-go` (official) | Streaming + vision support |
| Email | `resend-go` | Simple API, generous free tier |
| Config | `godotenv` + `os.Getenv` | No extra framework needed |
| Logging | `log/slog` (stdlib) | Zero dependency |

### 3.2 `main.go` Skeleton

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    "healthconnect/internal/server"
)

func main() {
    _ = godotenv.Load() // no-op if .env absent (production)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    r := server.NewRouter()
    slog.Info("HealthConnect server starting", "port", port)
    if err := http.ListenAndServeTLS(":"+port, "certs/cert.pem", "certs/key.pem", r); err != nil {
        slog.Error("server error", "err", err)
        os.Exit(1)
    }
}
```

> **Local HTTP fallback:** swap `ListenAndServeTLS` for `ListenAndServe` when running behind a TLS-terminating reverse proxy or for non-permission local dev (browser permissions won't work without HTTPS though — use `mkcert` to generate trusted local certs instead).

### 3.3 Router

```go
// internal/server/router.go
func NewRouter() http.Handler {
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(corsMiddleware())

    r.Post("/api/chat",       handlers.Chat)
    r.Post("/api/scan",       handlers.Scan)
    r.Post("/api/email",      handlers.SendEmail)
    r.Get ("/api/pharmacies", handlers.ListPharmacies)

    // Serve compiled Vite build
    r.Handle("/*", http.FileServer(http.Dir("./frontend/dist")))

    return r
}
```

### 3.4 CORS Middleware

```go
func corsMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", os.Getenv("ALLOWED_ORIGIN"))
            w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

## 4. Frontend — React Vite PWA

### 4.1 Stack

| Concern | Choice |
|---|---|
| Build tool | Vite 5 |
| Framework | React 18 + TypeScript |
| PWA plugin | `vite-plugin-pwa` |
| State | Zustand |
| Styling | Tailwind CSS v3 |
| Map | Leaflet + `react-leaflet` |
| Avatar animation | `lottie-react` |
| Icons | `lucide-react` |

### 4.2 `vite.config.ts`

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      manifest: {
        name: 'HealthConnect',
        short_name: 'HealthConnect',
        description: 'AI Doctor Health Assistant',
        theme_color: '#0ea5e9',
        background_color: '#ffffff',
        display: 'standalone',
        start_url: '/',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png' },
        ],
      },
      workbox: {
        // Cache the Vite build assets; skip caching API calls
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        runtimeCaching: [],
      },
    }),
  ],
  server: {
    proxy: {
      '/api': 'https://localhost:8080',
    },
    https: true, // required for mic/camera permissions in dev
  },
})
```

### 4.3 Session State (Zustand)

```ts
// src/store/session.ts
import { create } from 'zustand'

interface Message {
  role: 'user' | 'assistant'
  content: string
  timestamp: number
}

interface SessionState {
  mode: 'landing' | 'voice' | 'text'
  messages: Message[]
  isListening: boolean
  isSpeaking: boolean
  disclaimerAcknowledged: boolean
  emergencyTriggered: boolean
  setMode: (mode: SessionState['mode']) => void
  addMessage: (msg: Message) => void
  setListening: (v: boolean) => void
  setSpeaking: (v: boolean) => void
  acknowledgeDisclaimer: () => void
  triggerEmergency: () => void
}

export const useSession = create<SessionState>((set) => ({
  mode: 'landing',
  messages: [],
  isListening: false,
  isSpeaking: false,
  disclaimerAcknowledged: false,
  emergencyTriggered: false,
  setMode: (mode) => set({ mode }),
  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),
  setListening: (v) => set({ isListening: v }),
  setSpeaking: (v) => set({ isSpeaking: v }),
  acknowledgeDisclaimer: () => set({ disclaimerAcknowledged: true }),
  triggerEmergency: () => set({ emergencyTriggered: true }),
}))
```

---

## 5. AI & Multimodal Integration

### 5.1 System Prompt

```go
// internal/ai/prompts.go
const SystemPrompt = `You are HealthConnect, a friendly AI health assistant
serving users in Nigeria. Your role is to provide general health information
and guidance — never a medical diagnosis or prescription.

Rules:
1. Always be warm, clear, and plain-spoken.
2. Every response MUST end with:
   "⚠️ This is general information, not a medical diagnosis.
    Please consult a qualified healthcare professional."
3. If the user describes any of the following, respond ONLY with the
   EMERGENCY_FLAG token and nothing else:
   chest pain, difficulty breathing, severe bleeding, stroke symptoms
   (face drooping, arm weakness, speech difficulty), loss of consciousness,
   suicidal thoughts, or any life-threatening emergency.
4. After giving medication-related advice, include the token PHARMACY_PROMPT
   so the UI can offer to show nearby pharmacies.
5. Keep responses concise — the user may be on a mobile device.`

const EmergencyFlag = "EMERGENCY_FLAG"
const PharmacyPrompt = "PHARMACY_PROMPT"
```

### 5.2 Chat Handler

```go
// internal/server/handlers.go
type ChatRequest struct {
    Messages []ai.Message `json:"messages"`
}

type ChatResponse struct {
    Reply           string `json:"reply"`
    IsEmergency     bool   `json:"isEmergency"`
    ShowPharmacyBtn bool   `json:"showPharmacyBtn"`
}

func Chat(w http.ResponseWriter, r *http.Request) {
    var req ChatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    // Server-side red-flag pre-check (fast, before hitting Claude)
    lastUserMsg := req.Messages[len(req.Messages)-1].Content
    if ai.ContainsRedFlag(lastUserMsg) {
        writeJSON(w, ChatResponse{IsEmergency: true})
        return
    }

    reply, err := ai.Chat(r.Context(), req.Messages)
    if err != nil {
        http.Error(w, "ai error", http.StatusInternalServerError)
        return
    }

    resp := ChatResponse{
        Reply:           reply,
        IsEmergency:     strings.Contains(reply, ai.EmergencyFlag),
        ShowPharmacyBtn: strings.Contains(reply, ai.PharmacyPrompt),
    }
    // Strip control tokens from the reply shown to users
    resp.Reply = strings.ReplaceAll(resp.Reply, ai.EmergencyFlag, "")
    resp.Reply = strings.ReplaceAll(resp.Reply, ai.PharmacyPrompt, "")

    writeJSON(w, resp)
}
```

### 5.3 Anthropic Client

```go
// internal/ai/client.go
package ai

import (
    "context"
    anthropic "github.com/anthropics/anthropic-sdk-go"
)

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func Chat(ctx context.Context, history []Message) (string, error) {
    client := anthropic.NewClient() // reads ANTHROPIC_API_KEY from env

    msgs := make([]anthropic.MessageParam, len(history))
    for i, m := range history {
        if m.Role == "user" {
            msgs[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content))
        } else {
            msgs[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content))
        }
    }

    resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.ModelClaude3_5HaikuLatest, // fast + cheap for demo
        MaxTokens: 512,
        System:    []anthropic.TextBlockParam{{Text: SystemPrompt}},
        Messages:  msgs,
    })
    if err != nil {
        return "", err
    }
    return resp.Content[0].Text, nil
}
```

> Use `claude-haiku` for fast chat latency. Switch to `claude-sonnet` for the vision scan where reasoning quality matters more.

---

## 6. Voice Pipeline

### 6.1 Architecture

```
User speaks
    │
    ▼
Web Speech API (SpeechRecognition)   ← browser-native, free
    │  transcript string
    ▼
POST /api/chat                       ← Go backend → Claude
    │  reply text
    ▼
Web Speech API (SpeechSynthesis)     ← browser-native TTS
    │
    ▼
Avatar mouth animation triggered
```

### 6.2 `useSpeechRecognition` Hook

```ts
// src/hooks/useSpeechRecognition.ts
export function useSpeechRecognition(onResult: (text: string) => void) {
  const recognitionRef = useRef<SpeechRecognition | null>(null)
  const [isListening, setIsListening] = useState(false)

  const start = useCallback(() => {
    const SR = window.SpeechRecognition ?? window.webkitSpeechRecognition
    if (!SR) { alert('Speech recognition not supported in this browser.'); return }

    const rec = new SR()
    rec.lang = 'en-NG'
    rec.interimResults = false
    rec.maxAlternatives = 1

    rec.onresult = (e) => onResult(e.results[0][0].transcript)
    rec.onend = () => setIsListening(false)
    rec.onerror = () => setIsListening(false)

    rec.start()
    recognitionRef.current = rec
    setIsListening(true)
  }, [onResult])

  const stop = useCallback(() => {
    recognitionRef.current?.stop()
    setIsListening(false)
  }, [])

  return { isListening, start, stop }
}
```

### 6.3 `useSpeechSynthesis` Hook

```ts
// src/hooks/useSpeechSynthesis.ts
export function useSpeechSynthesis(onStart: () => void, onEnd: () => void) {
  const speak = useCallback((text: string) => {
    if (!window.speechSynthesis) return
    window.speechSynthesis.cancel()

    const utt = new SpeechSynthesisUtterance(text)
    utt.lang = 'en-NG'
    utt.rate = 0.95
    utt.pitch = 1.05

    // Prefer a female voice for the doctor avatar if available
    const voices = window.speechSynthesis.getVoices()
    const preferred = voices.find(v => v.lang.startsWith('en') && v.name.toLowerCase().includes('female'))
    if (preferred) utt.voice = preferred

    utt.onstart = onStart
    utt.onend = onEnd
    window.speechSynthesis.speak(utt)
  }, [onStart, onEnd])

  return { speak }
}
```

---

## 7. Animated Doctor Avatar

### 7.1 Approach: Lottie + State Machine

Three Lottie animation files drive the avatar states:

| File | State | Trigger |
|---|---|---|
| `avatar-idle.json` | Subtle breathing / blinking | Default |
| `avatar-listening.json` | Ear/attention animation | `isListening === true` |
| `avatar-talking.json` | Mouth movement loop | `isSpeaking === true` |

### 7.2 `DoctorAvatar` Component

```tsx
// src/components/DoctorAvatar.tsx
import Lottie from 'lottie-react'
import idleAnim from '../assets/avatar/avatar-idle.json'
import listeningAnim from '../assets/avatar/avatar-listening.json'
import talkingAnim from '../assets/avatar/avatar-talking.json'

interface Props { isListening: boolean; isSpeaking: boolean }

export function DoctorAvatar({ isListening, isSpeaking }: Props) {
  const anim = isSpeaking ? talkingAnim : isListening ? listeningAnim : idleAnim

  return (
    <div className="relative flex flex-col items-center">
      <Lottie
        animationData={anim}
        loop
        className="w-48 h-48 md:w-64 md:h-64"
      />
      <span className="mt-2 text-sm font-medium text-sky-600">
        {isSpeaking ? 'Speaking…' : isListening ? 'Listening…' : 'Dr. Ife'}
      </span>
      {isListening && (
        <span className="absolute bottom-8 flex h-3 w-3">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-sky-400 opacity-75" />
          <span className="relative inline-flex rounded-full h-3 w-3 bg-sky-500" />
        </span>
      )}
    </div>
  )
}
```

> **Avatar asset source:** Use a free medical/doctor Lottie from [LottieFiles.com](https://lottiefiles.com) with a CC0 or free licence. Search "doctor" or "medical". Download and save three variants (idle, listening, talking) or use one file with segment control via `lottie-react`'s `goToAndPlay`.

---

## 8. Camera Scan Feature

### 8.1 Frontend — `CameraScan.tsx`

```tsx
export function CameraScan({ onResult }: { onResult: (text: string) => void }) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [status, setStatus] = useState<'idle'|'streaming'|'captured'|'loading'|'done'>('idle')

  async function startCamera() {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
      videoRef.current!.srcObject = stream
      setStatus('streaming')
    } catch {
      alert('Camera permission denied. Please allow camera access and try again.')
    }
  }

  function capture() {
    const canvas = canvasRef.current!
    const video = videoRef.current!
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    canvas.getContext('2d')!.drawImage(video, 0, 0)
    ;(video.srcObject as MediaStream).getTracks().forEach(t => t.stop())
    setStatus('captured')
  }

  async function analyze() {
    setStatus('loading')
    const base64 = canvasRef.current!.toDataURL('image/jpeg', 0.8).split(',')[1]
    const res = await fetch('/api/scan', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ image: base64 }),
    })
    const data = await res.json()
    onResult(data.reply)
    setStatus('done')
  }

  return (
    <div className="flex flex-col items-center gap-4">
      <video ref={videoRef} autoPlay playsInline className={status === 'streaming' ? 'block' : 'hidden'} />
      <canvas ref={canvasRef} className={status === 'captured' ? 'block max-w-xs rounded-xl' : 'hidden'} />
      {status === 'idle' && <button onClick={startCamera}>📷 Scan face / injury</button>}
      {status === 'streaming' && <button onClick={capture}>📸 Capture</button>}
      {status === 'captured' && <button onClick={analyze}>🔍 Analyze</button>}
      {status === 'loading' && <p>Analyzing image…</p>}
    </div>
  )
}
```

### 8.2 Backend — Scan Handler

```go
// POST /api/scan
type ScanRequest struct {
    Image string `json:"image"` // base64 JPEG
}

func Scan(w http.ResponseWriter, r *http.Request) {
    var req ScanRequest
    json.NewDecoder(r.Body).Decode(&req)

    reply, err := ai.AnalyzeImage(r.Context(), req.Image)
    if err != nil {
        http.Error(w, "vision error", http.StatusInternalServerError)
        return
    }
    writeJSON(w, map[string]string{"reply": reply})
}
```

```go
// internal/ai/vision.go
func AnalyzeImage(ctx context.Context, base64JPEG string) (string, error) {
    client := anthropic.NewClient()

    resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.ModelClaude3_5SonnetLatest,
        MaxTokens: 512,
        System: []anthropic.TextBlockParam{{
            Text: SystemPrompt + `\n\nFor image analysis: describe visible symptoms in plain language,
            suggest possible general causes, and always recommend professional evaluation.
            Do NOT diagnose. If the image shows severe injury or emergency signs, output EMERGENCY_FLAG.`,
        }},
        Messages: []anthropic.MessageParam{
            anthropic.NewUserMessage(
                anthropic.NewImageBlockBase64("image/jpeg", base64JPEG),
                anthropic.NewTextBlock("Please analyze this image and provide general health observations."),
            ),
        },
    })
    if err != nil {
        return "", err
    }
    return resp.Content[0].Text, nil
}
```

---

## 9. Pharmacy Map & Geolocation

### 9.1 Data Model

```go
// internal/pharmacy/store.go
type Drug struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

type Pharmacy struct {
    ID      string  `json:"id"`
    Name    string  `json:"name"`
    Address string  `json:"address"`
    Area    string  `json:"area"`       // LGA / district
    City    string  `json:"city"`
    Lat     float64 `json:"lat"`
    Lng     float64 `json:"lng"`
    Phone   string  `json:"phone"`
    Drugs   []Drug  `json:"drugs"`
}
```

### 9.2 Pharmacy API

```
GET /api/pharmacies
    ?lat=6.5244&lng=3.3792    (geolocation — return 8 nearest)
    ?area=Ikeja               (manual search — return matches)
    (no params)               (return all — for map initialisation)
```

```go
func ListPharmacies(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    latStr, lngStr := q.Get("lat"), q.Get("lng")
    area := q.Get("area")

    all := pharmacy.All()

    switch {
    case latStr != "" && lngStr != "":
        lat, _ := strconv.ParseFloat(latStr, 64)
        lng, _ := strconv.ParseFloat(lngStr, 64)
        writeJSON(w, pharmacy.Nearest(all, lat, lng, 8))
    case area != "":
        writeJSON(w, pharmacy.ByArea(all, area))
    default:
        writeJSON(w, all)
    }
}
```

### 9.3 Frontend — `PharmacyMap.tsx`

```tsx
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'
import 'leaflet/dist/leaflet.css'

export function PharmacyMap() {
  const [pharmacies, setPharmacies] = useState<Pharmacy[]>([])
  const [search, setSearch] = useState('')
  const { coords, error } = useGeolocation()

  useEffect(() => {
    const params = coords
      ? `?lat=${coords.latitude}&lng=${coords.longitude}`
      : ''
    fetch(`/api/pharmacies${params}`)
      .then(r => r.json())
      .then(setPharmacies)
  }, [coords])

  const center: [number, number] = coords
    ? [coords.latitude, coords.longitude]
    : [6.5244, 3.3792] // Lagos default

  return (
    <div className="flex flex-col gap-4">
      <input
        placeholder="Search by area (e.g. Ikeja, Lekki)…"
        value={search}
        onChange={e => setSearch(e.target.value)}
        onKeyDown={e => e.key === 'Enter' && fetchByArea(search)}
        className="input"
      />
      <MapContainer center={center} zoom={13} className="h-72 rounded-2xl">
        <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
        {pharmacies.map(p => (
          <Marker key={p.id} position={[p.lat, p.lng]}>
            <Popup>
              <strong>{p.name}</strong><br />
              {p.address}<br />
              <button onClick={() => openReservation(p)}>Reserve a drug →</button>
            </Popup>
          </Marker>
        ))}
      </MapContainer>
      {/* Distance-sorted list below the map */}
      {pharmacies.map(p => <PharmacyCard key={p.id} pharmacy={p} />)}
    </div>
  )
}
```

---

## 10. Email Transcript

### 10.1 Backend Handler

```go
// POST /api/email
type EmailRequest struct {
    To         string `json:"to"`
    Transcript string `json:"transcript"`
    Summary    string `json:"summary"`
}

func SendEmail(w http.ResponseWriter, r *http.Request) {
    var req EmailRequest
    json.NewDecoder(r.Body).Decode(&req)

    if err := email.Send(req.To, req.Transcript, req.Summary); err != nil {
        http.Error(w, "email error", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

```go
// internal/email/resend.go
func Send(to, transcript, summary string) error {
    client := resend.NewClient(os.Getenv("RESEND_API_KEY"))

    body := fmt.Sprintf(`
<h2>Your HealthConnect Session Summary</h2>
<p>%s</p>
<hr>
<h3>Full Transcript</h3>
<pre>%s</pre>
<p><small>⚠️ This is general health information, not a medical diagnosis.
Consult a qualified healthcare professional.</small></p>
`, summary, transcript)

    _, err := client.Emails.Send(&resend.SendEmailRequest{
        From:    "HealthConnect <noreply@yourdomain.com>",
        To:      []string{to},
        Subject: "Your HealthConnect Health Session Transcript",
        Html:    body,
    })
    return err
}
```

> **Demo fallback:** If `RESEND_API_KEY` is empty, the handler returns HTTP 204 (success) and logs the email body to stdout. The frontend always sees "Email sent!" — the feature never breaks the demo.

---

## 11. Mocked Drug Reservation

No backend call is needed — this is entirely frontend-only.

```tsx
// src/components/DrugReservation.tsx
function generateRef(): string {
  return 'HC-' + Math.random().toString(36).slice(2, 8).toUpperCase()
}

export function DrugReservation({ drug, pharmacy }: Props) {
  const [confirmed, setConfirmed] = useState(false)
  const [ref] = useState(generateRef)

  if (confirmed) {
    return (
      <div className="rounded-2xl bg-green-50 p-6 text-center">
        <p className="text-2xl">✅</p>
        <p className="font-semibold text-green-700">Reservation confirmed!</p>
        <p className="text-sm text-gray-500 mt-1">Reference: <strong>{ref}</strong></p>
        <p className="text-xs text-gray-400 mt-2">
          Show this reference when you visit {pharmacy.name}.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-2xl border p-4">
      <p>Reserve <strong>{drug.name}</strong> at {pharmacy.name}?</p>
      <p className="text-sm text-gray-500">₦{drug.price.toLocaleString()}</p>
      <button onClick={() => setConfirmed(true)} className="btn-primary mt-3">
        Confirm reservation (demo)
      </button>
    </div>
  )
}
```

---

## 12. Safety — Disclaimers & Emergency Escalation

### 12.1 Red-Flag List

```ts
// src/lib/redFlags.ts (client-side pre-check — also mirrored in backend)
export const RED_FLAG_PATTERNS = [
  /chest\s*pain/i,
  /can'?t?\s*breathe/i,
  /difficulty\s*breath/i,
  /severe\s*bleed/i,
  /stroke/i,
  /face\s*drop/i,
  /arm\s*weak/i,
  /loss\s+of\s+consciousness/i,
  /fainted/i,
  /unconscious/i,
  /want\s+to\s+(kill|hurt)\s+(my)?self/i,
  /suicid/i,
]

export function containsRedFlag(text: string): boolean {
  return RED_FLAG_PATTERNS.some(p => p.test(text))
}
```

### 12.2 `EmergencyAlert` Component

```tsx
export function EmergencyAlert() {
  return (
    <div className="rounded-2xl bg-red-600 text-white p-6 text-center shadow-xl">
      <p className="text-3xl mb-2">🚨</p>
      <p className="text-xl font-bold">Seek Emergency Care Immediately</p>
      <p className="mt-2 text-sm">
        Based on your symptoms, you need in-person emergency care right away.
        Do not wait.
      </p>
      <a href="tel:112" className="mt-4 inline-block text-2xl font-bold underline">
        📞 Call 112 (Nigeria Emergency)
      </a>
      <p className="mt-2 text-xs opacity-75">
        HealthConnect cannot assist with emergencies.
      </p>
    </div>
  )
}
```

### 12.3 Disclaimer Component

```tsx
// One-time modal on first session start
export function DisclaimerModal({ onAck }: { onAck: () => void }) {
  return (
    <dialog open className="modal">
      <div className="modal-box">
        <h3 className="font-bold text-lg">Before we begin</h3>
        <p className="py-4 text-sm text-gray-600">
          HealthConnect provides <strong>general health information only</strong>,
          not a medical diagnosis, prescription, or treatment plan. Always consult
          a qualified healthcare professional for medical advice.
        </p>
        <button onClick={onAck} className="btn btn-primary w-full">
          I understand — let's continue
        </button>
      </div>
    </dialog>
  )
}

// Inline badge shown on every AI reply
export const DISCLAIMER_BADGE = (
  <p className="text-xs text-amber-600 mt-2 border-l-2 border-amber-400 pl-2">
    ⚠️ General information only — not a medical diagnosis.
    Consult a healthcare professional.
  </p>
)
```

---

## 13. Data — Seeded Pharmacies

`backend/data/pharmacies.json` — 10 realistic Lagos pharmacies:

```json
[
  {
    "id": "ph-001",
    "name": "HealthPlus Pharmacy — Ikeja",
    "address": "2 Allen Avenue, Ikeja, Lagos",
    "area": "Ikeja",
    "city": "Lagos",
    "lat": 6.6018,
    "lng": 3.3515,
    "phone": "+234-800-000-0001",
    "drugs": [
      { "name": "Paracetamol 500mg (24 tabs)", "price": 350 },
      { "name": "Amoxicillin 250mg (21 caps)", "price": 1200 },
      { "name": "ORS Sachet (10 pack)",         "price": 500 }
    ]
  },
  {
    "id": "ph-002",
    "name": "Medplus Pharmacy — Victoria Island",
    "address": "23 Adeola Odeku St, Victoria Island, Lagos",
    "area": "Victoria Island",
    "city": "Lagos",
    "lat": 6.4281,
    "lng": 3.4219,
    "phone": "+234-800-000-0002",
    "drugs": [
      { "name": "Ibuprofen 400mg (24 tabs)", "price": 400 },
      { "name": "Ciprofloxacin 500mg (10 tabs)", "price": 1800 }
    ]
  },
  {
    "id": "ph-003",
    "name": "Alpha Pharmacy — Lekki Phase 1",
    "address": "14 Admiralty Way, Lekki Phase 1, Lagos",
    "area": "Lekki",
    "city": "Lagos",
    "lat": 6.4495,
    "lng": 3.4700,
    "phone": "+234-800-000-0003",
    "drugs": [
      { "name": "Artemether-Lumefantrine (Coartem)", "price": 1500 },
      { "name": "Vitamin C 1000mg (30 tabs)", "price": 700 }
    ]
  },
  {
    "id": "ph-004",
    "name": "Sani Pharmacy — Surulere",
    "address": "45 Bode Thomas St, Surulere, Lagos",
    "area": "Surulere",
    "city": "Lagos",
    "lat": 6.4969,
    "lng": 3.3552,
    "phone": "+234-800-000-0004",
    "drugs": [
      { "name": "Metronidazole 400mg (21 tabs)", "price": 600 },
      { "name": "Zinc Sulphate 20mg (10 tabs)", "price": 300 }
    ]
  },
  {
    "id": "ph-005",
    "name": "Green Cross Pharmacy — Yaba",
    "address": "12 Commercial Ave, Yaba, Lagos",
    "area": "Yaba",
    "city": "Lagos",
    "lat": 6.5050,
    "lng": 3.3710,
    "phone": "+234-800-000-0005",
    "drugs": [
      { "name": "Omeprazole 20mg (28 caps)", "price": 900 },
      { "name": "Loratadine 10mg (10 tabs)", "price": 450 }
    ]
  }
]
```

> Add 5–7 more entries covering Ajah, Ikorodu, Mainland, Apapa, and Festac to reach 10–12 total. Use Google Maps to get accurate coordinates.

---

## 14. PWA Configuration

The app ships as a PWA so it can be added to the demo phone's home screen for a native feel:

- **`manifest.webmanifest`** — registered via `vite-plugin-pwa` (see §4.2).
- **Service worker** — caches all static assets; API calls are NOT cached (always live).
- **HTTPS** — required for both PWA install prompts and browser mic/camera permissions. Use `mkcert` locally.

```bash
# Generate trusted local certs
mkcert -install
mkcert localhost 127.0.0.1
# Move generated files to backend/certs/cert.pem and backend/certs/key.pem
```

---

## 15. Environment Variables

### Backend (`.env`)

```
ANTHROPIC_API_KEY=sk-ant-...
RESEND_API_KEY=re_...
ALLOWED_ORIGIN=https://localhost:5173
PORT=8080
# Optional: enable canned fallback replies (no Claude call) for offline demos
OFFLINE_MODE=false
```

### Frontend (`.env.local`)

```
VITE_API_BASE=https://localhost:8080
```

> The Vite dev server proxies `/api/*` to the Go backend, so `VITE_API_BASE` is only needed for production builds served from a different origin.

---

## 16. API Reference

| Method | Path | Request Body | Response |
|---|---|---|---|
| `POST` | `/api/chat` | `{ messages: [{role, content}] }` | `{ reply, isEmergency, showPharmacyBtn }` |
| `POST` | `/api/scan` | `{ image: string }` (base64 JPEG) | `{ reply: string }` |
| `POST` | `/api/email` | `{ to, transcript, summary }` | `204 No Content` |
| `GET` | `/api/pharmacies` | — | `Pharmacy[]` |
| `GET` | `/api/pharmacies?lat=&lng=` | — | `Pharmacy[]` (nearest 8) |
| `GET` | `/api/pharmacies?area=Ikeja` | — | `Pharmacy[]` (filtered) |

All responses are `application/json`. All error responses are plain text with an appropriate HTTP status code.

---

## 17. Local Development Setup

### Prerequisites

- Go 1.22+
- Node 20+
- `mkcert` (for local HTTPS)
- Anthropic API key

### Steps

```bash
# 1. Clone and install
git clone https://github.com/your-org/healthconnect.git
cd healthconnect

# 2. Generate local HTTPS certs
mkcert -install
mkcert localhost 127.0.0.1
mkdir -p backend/certs
mv localhost+1.pem backend/certs/cert.pem
mv localhost+1-key.pem backend/certs/key.pem

# 3. Configure environment
cp backend/.env.example backend/.env
# → Fill in ANTHROPIC_API_KEY (required), RESEND_API_KEY (optional)

# 4. Start the Go backend
cd backend
go mod download
go run ./cmd/server

# 5. Start the Vite frontend (new terminal)
cd frontend
npm install
npm run dev
# → Opens https://localhost:5173

# 6. Accept the self-signed cert warning once in Chrome
# → Navigate to https://localhost:5173
```

### Building for Production

```bash
# Build frontend
cd frontend && npm run build
# Output goes to frontend/dist/

# Build backend binary (serves dist/ as static files)
cd backend && go build -o healthconnect-server ./cmd/server

# Run
./healthconnect-server
```

---

## 18. Deployment Notes

For a public staging URL (useful to share with judges):

- **Option A — Railway.app:** Connect the repo; set env vars in the Railway dashboard. Add a `Procfile`: `web: ./healthconnect-server`. Railway provides HTTPS automatically.
- **Option B — Render.com:** Free tier Go web service. Build command: `go build -o healthconnect-server ./backend/cmd/server`. Start command: `./healthconnect-server`.
- **Option C — Fly.io:** `fly launch` from the repo root. Set secrets with `fly secrets set ANTHROPIC_API_KEY=...`.

In all cases, set `ALLOWED_ORIGIN` to the deployed frontend URL.

---

## 19. Demo Runbook

A step-by-step checklist for a live demo:

```
PRE-DEMO (30 min before)
□ Confirm HTTPS is working and mic/camera permissions are pre-granted in Chrome
□ Test voice round-trip (speak → avatar talks back)
□ Test camera scan on a visible skin area
□ Confirm pharmacy map loads with Lagos markers
□ Pre-warm the app (one message to avoid cold-start latency)
□ Have a backup screen recording ready

DEMO FLOW (~5 minutes)
1. Open https://healthconnect.yourdomain.com on the demo laptop/phone
2. Acknowledge disclaimer modal
3. Tap "Talk" (Voice mode) → speak a symptom ("I have a headache and fever")
   → Show avatar animating while replying
4. Tap "Scan" → point camera at arm → show AI analysis response
5. Tap "Find nearby pharmacy" → show map with Lagos markers
6. Tap a pharmacy → tap "Reserve" on a drug → show confirmation screen
7. Enter email address → tap "Email me this transcript" → show success toast
8. Switch to Text mode (secondary demo if time allows)

FALLBACKS
• If mic fails: switch to Text mode
• If camera fails: skip scan, show a pre-captured screenshot
• If Claude API is slow: toggle OFFLINE_MODE=true (canned responses)
• If map fails to load: show the pharmacy list view instead
```

---

*Document maintained by the HealthConnect hackathon team. Last updated: June 2026.*
