package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	ort "github.com/yalue/onnxruntime_go"
	"gocv.io/x/gocv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CameraURL            string  `yaml:"camera_url"`
	YOLOModel            string  `yaml:"yolo_model"`
	YOLOConf             float32 `yaml:"yolo_conf"`
	YOLOImgsz            int     `yaml:"yolo_imgsz"`
	MotionThreshold      float64 `yaml:"motion_threshold"`
	FrameInterval        float64 `yaml:"frame_interval"`
	CooldownSec          int     `yaml:"cooldown_sec"`
	VideoDuration        int     `yaml:"video_duration"`
	VideoMaxDuration     int     `yaml:"video_max_duration"`
	GeminiAPIKey         string  `yaml:"gemini_api_key"`
	GeminiNightMinMotion float64 `yaml:"gemini_night_min_motion"`
	GLMAPIKey            string  `yaml:"glm_api_key"`
	TelegramBotToken     string  `yaml:"telegram_bot_token"`
	TelegramChatIDs      []int64 `yaml:"telegram_chat_ids"`
	MediaRelayHost       string  `yaml:"media_relay_host"`
	SnapshotDir          string  `yaml:"snapshot_dir"`
	DatasetV3Dir         string  `yaml:"dataset_v3_dir"`
	DatasetV3Max         int     `yaml:"dataset_v3_max"`
	DatasetV3Interval    int     `yaml:"dataset_v3_interval"`
	DatasetV3DaytimeOnly bool    `yaml:"dataset_v3_daytime_only"`
	PresenceInterval     int     `yaml:"presence_interval"`    // seconds between presence checks; 0=off
	PresenceDiffThresh   float64 `yaml:"presence_diff_thresh"` // fraction of pixels changed vs empty ref
	PresenceAdaptRate    float64 `yaml:"presence_adapt_rate"`  // slow background adaptation when no dog
	ONNXRuntimeLibPath   string  `yaml:"onnxruntime_lib_path"`
	GrassMinGreenFrac    float64 `yaml:"grass_min_green_frac"`
	NotifyOnlyGrass      bool    `yaml:"notify_only_grass"`
	GeminiOnNegative     bool    `yaml:"gemini_on_negative"`
	GeminiMinMotion      float64 `yaml:"gemini_min_motion"`
	VLMMode              string  `yaml:"vlm_mode"`              // on | shadow | off — VLM bootstrap lifecycle
	VisitVideo           bool    `yaml:"visit_video"`           // record visit (arrival→departure) from daemon frames
	VisitPrerollSec      int     `yaml:"visit_preroll_sec"`     // pre-arrival ring buffer, seconds
	VisitMaxSec          int     `yaml:"visit_max_sec"`         // per-video cap; longer visits split into parts
	VisitEndGraceSec     int     `yaml:"visit_end_grace_sec"`   // no dog evidence for this long -> visit over
	VisitArmTimeoutSec   int     `yaml:"visit_arm_timeout_sec"` // armed (motion-started, unconfirmed) visit drop window
	SegmentSource        bool    `yaml:"segment_source"`        // v2: pull SD segments via SSH instead of RTSP
	SegmentHost          string  `yaml:"segment_host"`          // camera ssh host
	SegmentRemoteDir     string  `yaml:"segment_remote_dir"`    // recordings root on camera
	SegmentLocalDir      string  `yaml:"segment_local_dir"`     // local mirror of segments
	SegmentPollSec       int     `yaml:"segment_poll_sec"`      // poll interval
	SegmentKeepDays      int     `yaml:"segment_keep_days"`     // local retention
	SnapshotURL          string  `yaml:"snapshot_url"`          // live JPEG endpoint (arrival photos)
	// blob-concentration trigger: a still dog lying down produces little global
	// motion but a HIGHLY concentrated blob of change. 2026-08-19 case: motion
	// 0.5% (below 1.2% gate) yet 9-20x local/global concentration at ground
	// level — missed. Calibrated on real frames: dog blob sits in grid rows>=2
	// (ground), while IR sensor noise concentrates in rows 0-1 (sky/wall).
	BlobTriggerEnabled *bool   `yaml:"blob_trigger_enabled"` // nil (absent) = true
	BlobTriggerRatio   float64 `yaml:"blob_trigger_ratio"`   // min maxCell/global ratio (6x6 grid @320x240)
	BlobTriggerMinPx   float64 `yaml:"blob_trigger_min_px"`  // min changed px at 320x240
	BlobTriggerMinRow  int     `yaml:"blob_trigger_min_row"` // min grid row (0-5) for the max cell — ground zone
}

type Stats struct {
	frames         int64
	analyzed       int64
	dogs           int64
	yoloDogs       int64
	gemDogs        int64
	grassDogs      int64
	presenceChecks int64
	presenceDogs   int64
	vlmAuditRuns   int64 // shadow-mode: VLM audits run
	vlmAuditMisses int64 // shadow-mode: dogs the VLM saw that YOLO missed
	blobTriggers   int64 // motion gate bypassed by concentrated blob
}

type DetectorResult struct {
	Detected  bool
	Source    string
	Err       error
	Latency   time.Duration
	OnGrass   bool
	GrassFrac float64
}

type YoloDetector struct {
	session     *ort.DynamicAdvancedSession
	modelInput  int
	confidence  float32
	inputNames  []string
	outputNames []string
}

type Daemon struct {
	cfg          Config
	loc          *time.Location
	httpClient   *http.Client
	yolo         *YoloDetector
	cooldownTill time.Time
	stats        Stats
	lastDataset  time.Time
	// presence: background-subtraction vs empty-yard references
	bgRefDay          gocv.Mat // empty-yard reference, daytime colors
	bgRefNight        gocv.Mat // empty-yard reference, IR mode
	bgRefInitDay      bool
	bgRefInitNight    bool
	lastPresenceCheck time.Time
	lastPresenceDog   bool
	// visit recording: buffers JPEGs from the main frame loop (no 2nd RTSP)
	visitActive    bool
	visitAnnounced bool
	lastArmDrop    time.Time // suppress immediate re-arm after an armed drop
	visitThumb     []byte    // last confirmed-dog JPEG — becomes the visit video thumbnail
	// segment source v2: mirror SD segments and analyze on Pi
	prevAnalyze   gocv.Mat // previous small frame for motion diff (shared)
	segmentSeen   map[string]bool // remote segments already mirrored (existence-based — immune to camera clock jumps)
	segmentDate   string   // current date dir being watched
	segFrame      gocv.Mat // reusable decode mat
	visitStart    time.Time
	visitLastSeen time.Time
	visitPart     int
	visitFrames   [][]byte
	visitPreroll  [][]byte
}

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	testFrame := flag.String("test-frame", "", "Run YOLO+grass analysis on a single JPEG and exit")
	testBlob := flag.String("test-blob", "", "Run motion+blob-concentration analysis on FRAMEA,FRAMEB and exit")
	flag.Parse()

	cfg, err := readConfig(*configPath)
	if err != nil {
		log.Fatalf("read config: %v", err)
	}
	applyDefaults(&cfg)

	loc := mustBrazilLocation()
	if err := os.MkdirAll(cfg.SnapshotDir, 0o755); err != nil {
		log.Fatalf("mkdir snapshot dir: %v", err)
	}
	if cfg.DatasetV3Dir != "" {
		if err := os.MkdirAll(cfg.DatasetV3Dir, 0o755); err != nil {
			log.Fatalf("mkdir dataset dir: %v", err)
		}
	}

	ort.SetSharedLibraryPath(cfg.ONNXRuntimeLibPath)
	if err := ort.InitializeEnvironment(); err != nil {
		log.Fatalf("onnx init: %v", err)
	}
	defer func() {
		if err := ort.DestroyEnvironment(); err != nil {
			log.Printf("onnx destroy: %v", err)
		}
	}()

	if *testBlob != "" {
		parts := strings.SplitN(*testBlob, ",", 2)
		if len(parts) != 2 {
			log.Fatalf("--test-blob expects FRAMEA,FRAMEB (comma-separated)")
		}
		runBlobTest(cfg, parts[0], parts[1])
		return
	}

	yolo, err := NewYoloDetector(cfg)
	if err != nil {
		log.Fatalf("yolo init: %v", err)
	}
	if *testFrame != "" {
		runFrameTest(cfg, yolo, *testFrame)
		return
	}
	defer func() {
		if err := yolo.Close(); err != nil {
			log.Printf("yolo close: %v", err)
		}
	}()

	d := &Daemon{
		cfg: cfg,
		loc: loc,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		yolo: yolo,
	}

	log.Printf("daemon started: interval=%.1fs motion=%.4f cooldown=%ds", cfg.FrameInterval, cfg.MotionThreshold, cfg.CooldownSec)
	if err := d.Run(context.Background()); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}

func readConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("open %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.YOLOImgsz <= 0 {
		cfg.YOLOImgsz = 640
	}
	if cfg.FrameInterval <= 0 {
		cfg.FrameInterval = 2.0
	}
	if cfg.CooldownSec < 0 {
		cfg.CooldownSec = 0
	}
	if cfg.VideoDuration <= 0 {
		cfg.VideoDuration = 15
	}
	if cfg.VideoMaxDuration <= 0 {
		cfg.VideoMaxDuration = 90
	}
	if cfg.VideoDuration > cfg.VideoMaxDuration {
		cfg.VideoDuration = cfg.VideoMaxDuration
	}
	if cfg.DatasetV3Interval <= 0 {
		cfg.DatasetV3Interval = 30
	}
	// blob-concentration trigger defaults: enabled unless config says otherwise.
	// Calibrated 2026-08-19 on the real 03:08 arrival (fires on all 4 dog pairs,
	// silent on all noise pairs; noise lives in grid rows 0-1).
	if cfg.BlobTriggerRatio == 0 {
		cfg.BlobTriggerRatio = 8.0
	}
	if cfg.BlobTriggerMinPx == 0 {
		cfg.BlobTriggerMinPx = 130.0
	}
	if cfg.BlobTriggerMinRow == 0 {
		cfg.BlobTriggerMinRow = 2
	}
	if cfg.BlobTriggerEnabled == nil {
		t := true
		cfg.BlobTriggerEnabled = &t
	}
	if cfg.DatasetV3Max <= 0 {
		cfg.DatasetV3Max = 200
	}
	if cfg.PresenceInterval <= 0 {
		cfg.PresenceInterval = 0 // off by default; set in config.yaml
	}
	if cfg.VLMMode == "" {
		cfg.VLMMode = "on"
	}
	cfg.VLMMode = strings.ToLower(strings.TrimSpace(cfg.VLMMode))
	if cfg.VLMMode != "on" && cfg.VLMMode != "shadow" && cfg.VLMMode != "off" {
		log.Printf("vlm_mode %q invalid; defaulting to on", cfg.VLMMode)
		cfg.VLMMode = "on"
	}
	log.Printf("vlm_mode=%s (on=VLM rescues YOLO misses | shadow=YOLO decides, VLM audits silently | off=no VLM API calls)", cfg.VLMMode)
	if cfg.VisitPrerollSec <= 0 {
		cfg.VisitPrerollSec = 60
	}
	if cfg.VisitMaxSec <= 0 {
		cfg.VisitMaxSec = 600
	}
	if cfg.VisitEndGraceSec <= 0 {
		cfg.VisitEndGraceSec = 360
	}
	if cfg.VisitArmTimeoutSec <= 0 {
		cfg.VisitArmTimeoutSec = 180
	}
	if cfg.SegmentHost == "" {
		cfg.SegmentHost = "192.168.100.64"
	}
	if cfg.SegmentRemoteDir == "" {
		cfg.SegmentRemoteDir = "/mnt/mmcblk0p1/recordings"
	}
	if cfg.SegmentLocalDir == "" {
		cfg.SegmentLocalDir = "/root/clawd/segments"
	}
	if cfg.SegmentPollSec <= 0 {
		cfg.SegmentPollSec = 8
	}
	if cfg.SegmentKeepDays <= 0 {
		cfg.SegmentKeepDays = 2
	}
	if cfg.PresenceDiffThresh <= 0 {
		cfg.PresenceDiffThresh = 0.02 // 2% of pixels differing = presence candidate
	}
	if cfg.PresenceAdaptRate <= 0 {
		cfg.PresenceAdaptRate = 0.05 // slow: ~20 checks to absorb an object
	}
	if cfg.YOLOConf <= 0 {
		cfg.YOLOConf = 0.45
	}
	if cfg.GrassMinGreenFrac <= 0 {
		cfg.GrassMinGreenFrac = 0.22
	}
	if cfg.GeminiMinMotion <= 0 {
		cfg.GeminiMinMotion = 0.015
	}
}

func runFrameTest(cfg Config, yolo *YoloDetector, path string) {
	img := gocv.IMRead(path, gocv.IMReadColor)
	if img.Empty() {
		log.Fatalf("test-frame: cannot read %s", path)
	}
	defer img.Close()
	d := &Daemon{cfg: cfg, loc: mustBrazilLocation(), httpClient: &http.Client{Timeout: 30 * time.Second}, yolo: yolo}
	res := d.runYolo(img)
	log.Printf("test-frame %s: detected=%v onGrass=%v grassFrac=%.3f source=%s err=%v latency=%s",
		path, res.Detected, res.OnGrass, res.GrassFrac, res.Source, res.Err, res.Latency)
	if !res.Detected && cfg.GeminiOnNegative && cfg.VLMMode != "off" {
		jpegBytes, err := encodeJPEG(img)
		if err == nil {
			gem := d.runVisionLLM([][]byte{jpegBytes})
			log.Printf("test-frame visionllm: detected=%v onGrass=%v err=%v latency=%s", gem.Detected, gem.OnGrass, gem.Err, gem.Latency)
		}
	}
}

func mustBrazilLocation() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Printf("time zone fallback: %v", err)
		return time.FixedZone("Brasilia", -3*3600)
	}
	return loc
}

func NewYoloDetector(cfg Config) (*YoloDetector, error) {
	inputNames := []string{"images"}
	outputNames := []string{"output0"}
	session, err := ort.NewDynamicAdvancedSession(cfg.YOLOModel, inputNames, outputNames, nil)
	if err != nil {
		return nil, err
	}
	return &YoloDetector{
		session:     session,
		modelInput:  cfg.YOLOImgsz,
		confidence:  cfg.YOLOConf,
		inputNames:  inputNames,
		outputNames: outputNames,
	}, nil
}

func (y *YoloDetector) Close() error {
	if y == nil || y.session == nil {
		return nil
	}
	return y.session.Destroy()
}

func (d *Daemon) Run(ctx context.Context) error {
	statsTicker := time.NewTicker(1 * time.Minute)
	defer statsTicker.Stop()

	// presence references must be allocated gocv Mats before use
	d.bgRefDay = gocv.NewMat()
	defer d.bgRefDay.Close()
	d.bgRefNight = gocv.NewMat()
	defer d.bgRefNight.Close()

	if d.cfg.SegmentSource {
		if err := d.runSegmentLoop(ctx, statsTicker); err != nil && err != context.Canceled {
			log.Printf("segment loop ended: %v — falling back to RTSP", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := d.runCameraLoop(ctx, statsTicker); err != nil {
			log.Printf("camera loop ended: %v", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

// ---- Segment source v2 ----
// The camera records 24/7 to its SD card at full quality (2K@~15fps). The
// daemon no longer holds an RTSP session open; instead it polls the camera
// over SSH (key auth), mirrors finished segments locally, and analyzes
// decoded frames from them. Delay: segment length + poll (~15s typical).

func segRemoteList(host, dir string) ([]string, error) {
	// camera-local time first line, then segment list — one round trip.
	// Segments still being written lack the mp4 moov atom; only list
	// files whose start-time is >= 25s in the past (10s record + margin).
	// CHEAP-LISTING RULE (2026-08-18): never `ls -t` the whole recordings
	// tree — with ~4.3k segments across 3 day-dirs, mtime sort stats every
	// file and pinned the T31 at 91% sys time / load 4.2 (single core).
	// Instead: list ONLY the last two day-dirs (today + yesterday, midnight
	// rollover), plain `ls -1` (no stat), sort by filename in Go —
	// vid_HHMMSS zero-padded is lexicographic == chronological within a day.
	// HOUR-PREFIX GLOB (v6, 2026-08-18): fixes the tail -60 regression — in
	// the mixed-era day-dir (UTC-era vid_22xx before the camera TZ fix +
	// BRT-era vid_19xx after), the lexicographic tail selected the OLD UTC
	// names and the mirror froze for the rest of the day. Instead: ask the
	// CAMERA its own clock first, then glob only the current + previous
	// hour prefixes of today's dir (era-proof: date and filenames share
	// whatever clock the camera runs). ≤124 entries, no stat, no big output.
	// Hour 00 also globs yesterday's vid_23* (midnight rollover); the age
	// check below filters same-hour-yesterday collisions (age ≈ 0 → skip).
	// Outages >2h leave older segments on the camera SD (manual recovery).
	nowOut, err := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=8",
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=/tmp/kikaseg-%r@%h:%p",
		"-o", "ControlPersist=600",
		"root@"+host,
		"date +%H%M%S").Output()
	if err != nil || len(nowOut) == 0 {
		return nil, fmt.Errorf("camera clock probe: %v", err)
	}
	var nowH, nowM, nowS int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(nowOut)), "%2d%2d%2d", &nowH, &nowM, &nowS); err != nil {
		return nil, fmt.Errorf("camera clock: %w", err)
	}
	nowSec := nowH*3600 + nowM*60 + nowS
	h0, h1 := nowH, nowH-1
	ydir := ""
	if h1 < 0 {
		h1 += 24
		ydir = time.Now().In(mustBrazilLocation()).AddDate(0, 0, -1).Format("2006-01-02")
	}
	todayDir := time.Now().In(mustBrazilLocation()).Format("2006-01-02")
	var sb strings.Builder
	first := true
	addGlob := func(dayDir string, hh int) {
		if !first {
			sb.WriteString("; ")
		}
		first = false
		sb.WriteString(fmt.Sprintf("ls -1 %s/%s/vid_%02d*.mp4 2>/dev/null", dir, dayDir, hh))
	}
	addGlob(todayDir, h1) // previous hour first (chronological)
	addGlob(todayDir, h0)
	if ydir != "" {
		addGlob(ydir, h1) // 23:xx of yesterday when now is 00:xx
	}
	out, err := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=8",
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=/tmp/kikaseg-%r@%h:%p",
		"-o", "ControlPersist=600",
		"root@"+host,
		sb.String()).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var segs []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if !strings.HasSuffix(l, ".mp4") {
			continue
		}
		base := filepath.Base(l)
		var hh, mm, ss int
		if _, err := fmt.Sscanf(base, "vid_%2d%2d%2d.mp4", &hh, &mm, &ss); err != nil {
			continue
		}
		age := nowSec - (hh*3600 + mm*60 + ss)
		if age < 0 {
			age += 86400 // midnight rollover
		}
		if age < 25 {
			continue // still recording (or just closed)
		}
		segs = append(segs, l)
	}
	// candidates arrive grouped by day-dir (yesterday's batch first, each in
	// filename order). Sort by full path: day-dir prefix + vid_HHMMSS is
	// chronologically sound. Replaces the old mtime-based reverse.
	sort.Strings(segs)
	return segs, nil
}

func segFetch(host, remote, local string) error {
	// thingino dropbear has no sftp-server: stream via ssh cat (works always)
	f, err := os.Create(local + ".part")
	if err != nil {
		return err
	}
	cmd := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=8",
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=/tmp/kikaseg-%r@%h:%p",
		"-o", "ControlPersist=600",
		"root@"+host,
		"cat "+remote)
	cmd.Stdout = f
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	err = cmd.Run()
	f.Close()
	if err != nil {
		os.Remove(local + ".part")
		return fmt.Errorf("ssh cat: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return os.Rename(local+".part", local)
}

// decodeSegmentFrames: extract up to n evenly-spaced frames from a segment.
func decodeSegmentFrames(path string, want int) ([]gocv.Mat, error) {
	cap, err := gocv.VideoCaptureFile(path)
	if err != nil || !cap.IsOpened() {
		return nil, fmt.Errorf("open segment %s: %v", path, err)
	}
	defer cap.Close()
	var frames []gocv.Mat
	total := int(cap.Get(gocv.VideoCaptureFrameCount))
	if total <= 0 {
		// stream-less fallback: read sequentially
		m := gocv.NewMat()
		for cap.Read(&m) && len(frames) < want {
			if !m.Empty() {
				frames = append(frames, m.Clone())
			}
		}
		m.Close()
		return frames, nil
	}
	m := gocv.NewMat()
	defer m.Close()
	for i := 0; i < want; i++ {
		pos := total * (i + 1) / (want + 1) // skip first/last (mid-GOP noise)
		cap.Set(gocv.VideoCapturePosFrames, float64(pos))
		if !cap.Read(&m) || m.Empty() {
			continue
		}
		frames = append(frames, m.Clone())
	}
	return frames, nil
}

func (d *Daemon) runSegmentLoop(ctx context.Context, statsTicker *time.Ticker) error {
	log.Printf("segment loop: host=%s remote=%s local=%s poll=%ds", d.cfg.SegmentHost, d.cfg.SegmentRemoteDir, d.cfg.SegmentLocalDir, d.cfg.SegmentPollSec)
	os.MkdirAll(d.cfg.SegmentLocalDir, 0o755)
	d.segFrame = gocv.NewMat()
	defer d.segFrame.Close()

	ticker := time.NewTicker(time.Duration(d.cfg.SegmentPollSec) * time.Second)
	defer ticker.Stop()

	// priming poll: mark everything already on the camera as seen — mirror
	// only genuinely new segments. Existence-based, so a camera clock jump
	// (timezone change, rescue flash) never freezes or replays the mirror.
	d.segmentSeen = map[string]bool{}
	if segs, err := segRemoteList(d.cfg.SegmentHost, d.cfg.SegmentRemoteDir); err == nil && len(segs) > 0 {
		for _, s := range segs {
			d.segmentSeen[s] = true
		}
		log.Printf("segment loop: primed at %s (%d historical marked seen)", filepath.Base(segs[len(segs)-1]), len(segs))
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-statsTicker.C:
			d.logStats()
		case <-ticker.C:
		}

		segs, err := segRemoteList(d.cfg.SegmentHost, d.cfg.SegmentRemoteDir)
		if err != nil {
			log.Printf("segment poll: %v", err)
			continue
		}
		if len(segs) == 0 {
			continue
		}
		if d.segmentSeen == nil {
			d.segmentSeen = map[string]bool{}
		}
		// bound the seen-map: keep only names still present on the camera
		if len(d.segmentSeen) > 5000 {
			keep := map[string]bool{}
			for _, s := range segs {
				keep[s] = true
			}
			d.segmentSeen = keep
		}
		fresh := []string(nil)
		for _, s := range segs {
			if !d.segmentSeen[s] {
				fresh = append(fresh, s)
			}
		}
		if len(fresh) == 0 {
			continue
		}
		// fetch all new segments (bounded by tail -40)
		type item struct {
			remote, local string
		}
		var items []item
		for _, r := range fresh {
			rel := strings.TrimPrefix(r, d.cfg.SegmentRemoteDir+"/")
			local := filepath.Join(d.cfg.SegmentLocalDir, rel)
			os.MkdirAll(filepath.Dir(local), 0o755)
			if err := segFetch(d.cfg.SegmentHost, r, local); err != nil {
				log.Printf("segment fetch %s: %v", filepath.Base(r), err)
				continue
			}
			items = append(items, item{r, local})
			d.segmentSeen[r] = true
		}
		if len(items) == 0 {
			continue
		}
		// analyze: one representative frame per segment
		for _, it := range items {
			mats, err := decodeSegmentFrames(it.local, 1)
			if err != nil || len(mats) == 0 {
				log.Printf("segment decode %s: %v", filepath.Base(it.local), err)
				continue
			}
			frame := mats[0]
			d.stats.frames++
			if d.cfg.VisitVideo {
				if vj, vjerr := encodeJPEG(frame); vjerr == nil {
					d.pushVisitFrame(vj)
					d.checkVisitTimeout()
				}
			}
			d.processFrame(frame)
			frame.Close()
			for _, m := range mats[1:] {
				m.Close()
			}
		}
		d.pruneSegments()
	}
}

func stringsHas(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// pruneSegments: drop local mirror files older than SegmentKeepDays.
func (d *Daemon) pruneSegments() {
	cutoff := time.Now().AddDate(0, 0, -d.cfg.SegmentKeepDays)
	entries, err := os.ReadDir(d.cfg.SegmentLocalDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dayDir := filepath.Join(d.cfg.SegmentLocalDir, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.RemoveAll(dayDir)
			log.Printf("segments: pruned %s", e.Name())
			continue
		}
		files, _ := os.ReadDir(dayDir)
		if len(files) > 900 { // safety: cap per-day files
			for _, f := range files[:len(files)-600] {
				os.Remove(filepath.Join(dayDir, f.Name()))
			}
		}
	}
}

// processFrame: shared analysis path for a decoded frame (RTSP or segment).
// motion/presence -> YOLO -> VLM -> visit hooks -> notify. Consumes nothing;
// caller owns the frame Mat.
func (d *Daemon) processFrame(frame gocv.Mat) {
	currSmall := gocv.NewMat()
	defer currSmall.Close()
	prevSmall := gocv.NewMat()
	defer prevSmall.Close()

	if d.prevAnalyze.Ptr() == nil {
		d.prevAnalyze = gocv.NewMat() // allocate C-side first: CopyTo needs valid dst
	}
	gocv.Resize(frame, &currSmall, image.Pt(320, 240), 0, 0, gocv.InterpolationLinear)
	if d.prevAnalyze.Empty() {
		currSmall.CopyTo(&d.prevAnalyze)
		d.tryDatasetCapture(frame)
		return
	}

	diff := gocv.NewMat()
	defer diff.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	thresh := gocv.NewMat()
	defer thresh.Close()

	motion := motionPercent(d.prevAnalyze, currSmall, &diff, &gray, &thresh)
	// Concentration must be read from the RAW threshold mask (as calibrated),
	// before presenceBlobBBox applies its morphological cleanup to thresh.
	_, _, blobRatio, blobPx, blobR, blobC := blobConcentration(thresh, 6)
	// Blob bbox must be computed BEFORE prevAnalyze is overwritten below —
	// the crop for the VLM needs the OLD→NEW change region.
	bx, by, bw, bh, bok := presenceBlobBBox(d.prevAnalyze, currSmall, &diff, &gray, &thresh)
	currSmall.CopyTo(&d.prevAnalyze)
	d.armVisit(motion)
	d.tryDatasetCapture(frame)
	d.maybePresenceCheck(frame, currSmall)

	// Blob-concentration gate: motion below threshold can still be a still dog
	// if the change is small but highly concentrated (lying down, barely moving)
	// and sits in the ground zone (grid rows >= BlobTriggerMinRow).
	blobFired := false
	if motion < d.cfg.MotionThreshold && *d.cfg.BlobTriggerEnabled &&
		blobRatio >= d.cfg.BlobTriggerRatio && blobPx >= d.cfg.BlobTriggerMinPx && blobR >= d.cfg.BlobTriggerMinRow {
		blobFired = true
		d.stats.blobTriggers++
		log.Printf("blob trigger: motion=%.2f%% below thresh but concentrated (ratio=%.0fx, px=%d, cell=(%d,%d))",
			motion*100, blobRatio, int(blobPx), blobR, blobC)
	}
	if motion < d.cfg.MotionThreshold && !blobFired {
		return
	}
	if time.Now().Before(d.cooldownTill) {
		return
	}

	d.stats.analyzed++
	jpegBytes, err := encodeJPEG(frame)
	if err != nil {
		log.Printf("encode jpeg: %v", err)
		return
	}

	yoloRes := d.runYolo(frame)
	d.logMotion(motion, yoloRes)

	result := yoloRes
	geminiTrigger := (result.Err != nil && motion >= d.cfg.GeminiNightMinMotion) ||
		(d.cfg.GeminiOnNegative && !result.Detected && result.Err == nil && motion >= d.cfg.GeminiMinMotion)
	// Still/lying dog cases arrive with LOW motion: extend the VLM trigger to
	// blob-triggered frames (concentrated change below the motion gate).
	if blobFired {
		geminiTrigger = d.vlmEnabled()
	}
	vlmImages := d.motionVLMBundle(frame, jpegBytes, bok, bx, by, bw, bh)
	switch {
	case d.vlmShadow() && result.Err == nil && !result.Detected && motion >= d.cfg.GeminiMinMotion:
		gemRes := d.runVisionLLM(vlmImages)
		d.stats.vlmAuditRuns++
		if gemRes.Detected {
			d.stats.vlmAuditMisses++
			log.Printf("shadow: VLM saw a dog YOLO missed (motion=%.1f%%); sample harvested, not notifying", motion*100)
			d.saveShadowSample(frame, motion)
		}
	case d.vlmShadow() && result.Err != nil && motion >= d.cfg.GeminiNightMinMotion:
		gemRes := d.runVisionLLM(vlmImages)
		if gemRes.Detected {
			result = gemRes
			d.stats.gemDogs++
		}
	case geminiTrigger && d.vlmEnabled():
		gemRes := d.runVisionLLM(vlmImages)
		if gemRes.Detected {
			result = gemRes
			d.stats.gemDogs++
		}
	}

	if !result.Detected {
		return
	}
	d.visitThumb = jpegBytes
	if d.cfg.NotifyOnlyGrass && !result.OnGrass {
		log.Printf("kika detected off-grass (grass=%.1f%%); notify_only_grass=true, skipping notify", result.GrassFrac*100)
		return
	}

	d.stats.dogs++
	if result.Source == "yolo" {
		d.stats.yoloDogs++
	}
	if result.OnGrass {
		d.stats.grassDogs++
	}

	if d.visitActive && d.visitAnnounced {
		d.keepAliveVisit()
		d.cooldownTill = time.Now().Add(time.Duration(d.cfg.CooldownSec) * time.Second)
		return
	}
	if d.cfg.VisitVideo {
		d.beginVisit()
		d.visitAnnounced = true
	}

	ts := time.Now().In(d.loc)
	pct := motion * 100
	photoCaption := fmt.Sprintf("@klka *Kika no quintal!* @%s (movimento %.1f%%)", ts.Format("15:04:05"), pct)
	if d.cfg.VisitVideo {
		photoCaption += "\n🎬 Gravando a visita — vídeo completo quando ela sair"
	}
	if result.Source != "yolo" {
		photoCaption += " · VLM"
	}
	photoJPEG := d.arrivalPhotoJPEG(jpegBytes)
	snapshotPath, err := d.saveSnapshot(photoJPEG, ts, motion)
	if err == nil {
		if err := d.sendPhoto(snapshotPath, photoCaption); err != nil {
			log.Printf("send photo: %v", err)
		}
	}
	d.cooldownTill = time.Now().Add(time.Duration(d.cfg.CooldownSec) * time.Second)
}

func (d *Daemon) runCameraLoop(ctx context.Context, statsTicker *time.Ticker) error {
	if err := os.Setenv("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp"); err != nil {
		log.Printf("set ffmpeg capture opts: %v", err)
	}

	cap, err := gocv.VideoCaptureFile(d.cfg.CameraURL)
	if err != nil {
		return fmt.Errorf("open camera: %w", err)
	}
	defer cap.Close()
	if !cap.IsOpened() {
		return errors.New("camera not opened")
	}

	frame := gocv.NewMat()
	defer frame.Close()
	currSmall := gocv.NewMat()
	defer currSmall.Close()
	prevSmall := gocv.NewMat()
	defer prevSmall.Close()
	diff := gocv.NewMat()
	defer diff.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	thresh := gocv.NewMat()
	defer thresh.Close()

	interval := time.Duration(d.cfg.FrameInterval * float64(time.Second))
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-statsTicker.C:
			d.logStats()
		case <-ticker.C:
		}

		if ok := cap.Read(&frame); !ok {
			return errors.New("read failed")
		}
		if frame.Empty() {
			log.Printf("empty frame")
			continue
		}

		d.stats.frames++
		if d.cfg.VisitVideo {
			if vj, vjerr := encodeJPEG(frame); vjerr == nil {
				d.pushVisitFrame(vj)
				d.checkVisitTimeout()
			}
		}
		gocv.Resize(frame, &currSmall, image.Pt(320, 240), 0, 0, gocv.InterpolationLinear)
		if prevSmall.Empty() {
			currSmall.CopyTo(&prevSmall)
			d.tryDatasetCapture(frame)
			continue
		}

		motion := motionPercent(prevSmall, currSmall, &diff, &gray, &thresh)
		currSmall.CopyTo(&prevSmall)
		d.armVisit(motion)
		d.tryDatasetCapture(frame)
		d.maybePresenceCheck(frame, currSmall)

		if motion < d.cfg.MotionThreshold {
			continue
		}
		if time.Now().Before(d.cooldownTill) {
			continue
		}

		d.stats.analyzed++
		jpegBytes, err := encodeJPEG(frame)
		if err != nil {
			log.Printf("encode jpeg: %v", err)
			continue
		}

		yoloRes := d.runYolo(frame)
		d.logMotion(motion, yoloRes)

		result := yoloRes
		geminiTrigger := (result.Err != nil && motion >= d.cfg.GeminiNightMinMotion) ||
			(d.cfg.GeminiOnNegative && !result.Detected && result.Err == nil && motion >= d.cfg.GeminiMinMotion)
		switch {
		case d.vlmShadow() && result.Err == nil && !result.Detected && motion >= d.cfg.GeminiMinMotion:
			// SHADOW audit: YOLO said no — VLM checks silently. A miss = free training sample.
			gemRes := d.runVisionLLM([][]byte{jpegBytes})
			d.stats.vlmAuditRuns++
			if gemRes.Detected {
				d.stats.vlmAuditMisses++
				log.Printf("shadow: VLM saw a dog YOLO missed (motion=%.1f%%); sample harvested, not notifying", motion*100)
				d.saveShadowSample(frame, motion)
			}
		case d.vlmShadow() && result.Err != nil && motion >= d.cfg.GeminiNightMinMotion:
			// SHADOW + YOLO error: no local decision exists — VLM rescues this one frame.
			gemRes := d.runVisionLLM([][]byte{jpegBytes})
			if gemRes.Detected {
				result = gemRes
				d.stats.gemDogs++
			}
		case geminiTrigger && d.vlmEnabled():
			gemRes := d.runVisionLLM([][]byte{jpegBytes})
			if gemRes.Detected {
				result = gemRes
				d.stats.gemDogs++
			}
		}

		if !result.Detected {
			continue
		}
		d.visitThumb = jpegBytes
		if d.cfg.NotifyOnlyGrass && !result.OnGrass {
			log.Printf("kika detected off-grass (grass=%.1f%%); notify_only_grass=true, skipping notify", result.GrassFrac*100)
			continue
		}

		d.stats.dogs++
		if result.Source == "yolo" {
			d.stats.yoloDogs++
		}
		if result.OnGrass {
			d.stats.grassDogs++
		}

		if d.visitActive && d.visitAnnounced {
			// same visit already announced — keep the recording alive only
			d.keepAliveVisit()
			d.cooldownTill = time.Now().Add(time.Duration(d.cfg.CooldownSec) * time.Second)
			continue
		}
		if d.cfg.VisitVideo {
			d.beginVisit()
			d.visitAnnounced = true
		}
		d.keepAliveVisit()

		now := time.Now().In(d.loc)
		snapshotPath, err := d.saveSnapshot(jpegBytes, now, motion)
		if err != nil {
			log.Printf("save snapshot: %v", err)
			continue
		}

		detectorLabel := "YOLO"
		if result.Source != "yolo" {
			detectorLabel = "VLM"
		}
		place := "na área interna 🏠"
		if result.OnGrass {
			place = "NA GRAMA 🌱"
		}
		photoCaption := fmt.Sprintf("🐕 *Kika %s*\nHorário: *%s (Brasília)*\nMovimento: *%.1f%%*\nGrama: *%.0f%%*\nDetector: %s",
			place, now.Format("15:04:05"), motion*100, result.GrassFrac*100, detectorLabel)
		if d.cfg.VisitVideo {
			photoCaption += "\n🎬 Gravando a visita — vídeo completo quando ela sair"
		}

		if err := d.sendPhoto(snapshotPath, photoCaption); err != nil {
			log.Printf("notify error: %v", err)
		}

		d.cooldownTill = time.Now().Add(time.Duration(d.cfg.CooldownSec) * time.Second)
	}
}

// VLM bootstrap lifecycle (vlm_mode). The VLM is scaffolding: it labels and
// verifies while the local YOLO learns, then is retired.
//
//	on     — VLM rescues YOLO misses (training phase, current default)
//	shadow — local YOLO is the decisor; VLM runs silently to audit and harvest
//	         miss samples as training data. No VLM-driven notifications.
//	off    — zero VLM API calls; fully local detector (post-graduation).
func (d *Daemon) vlmEnabled() bool { return d.cfg.VLMMode != "off" }
func (d *Daemon) vlmShadow() bool  { return d.cfg.VLMMode == "shadow" }

// saveShadowSample: motion-path frame the VLM confirmed as a dog but YOLO
// missed (no bbox on this path) — saved unlabeled for a later labeling pass.
func (d *Daemon) saveShadowSample(frame gocv.Mat, motion float64) {
	dir := d.cfg.DatasetV3Dir
	if dir == "" {
		return
	}
	shDir := filepath.Join(dir, "shadow_pos")
	if err := os.MkdirAll(shDir, 0o755); err != nil {
		return
	}
	entries, err := os.ReadDir(shDir)
	if err == nil && len(entries) >= d.cfg.DatasetV3Max {
		return
	}
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, frame)
	if err != nil {
		return
	}
	defer buf.Close()
	ts := time.Now().In(d.loc)
	stem := fmt.Sprintf("shadow_%s_%d", ts.Format("20060102_150405"), ts.UnixNano())
	if err := os.WriteFile(filepath.Join(shDir, stem+".jpg"), append([]byte(nil), buf.GetBytes()...), 0o644); err != nil {
		return
	}
	meta := fmt.Sprintf("{\"motion\": %.4f, \"vlm_confirmed\": true, \"ts\": \"%s\"}\n", motion, ts.Format(time.RFC3339))
	_ = os.WriteFile(filepath.Join(shDir, stem+".json"), []byte(meta), 0o644)
	log.Printf("shadow: training sample saved %s.jpg (motion=%.1f%%)", stem, motion*100)
}

// ---- Visit recording (from the daemon's own frame loop) ----
// The previous flow opened a SECOND RTSP connection via ffmpeg right after a
// detection; the camera routinely refused it (332-byte empty "clips"). The
// visit recorder instead buffers JPEG frames straight from the main loop:
// pre-roll before she arrives, the whole visit while she stays, then muxes
// them locally with ffmpeg image2 (files only — no camera connection) and
// sends ONE video when she leaves. One visit = one arrival photo + one
// departure video; no duplicate motion/presence photos.

func (d *Daemon) visitFrameBudget() int {
	perFrame := time.Duration(d.cfg.FrameInterval * float64(time.Second))
	if perFrame <= 0 {
		perFrame = 2 * time.Second
	}
	n := int(time.Duration(d.cfg.VisitMaxSec)*time.Second/perFrame) + 1
	if n < 2 {
		n = 2
	}
	return n
}

func (d *Daemon) pushVisitFrame(jpeg []byte) {
	if !d.cfg.VisitVideo {
		return
	}
	if d.visitActive {
		if len(d.visitFrames) >= d.visitFrameBudget() {
			d.finalizeVisitPart() // over-long visit: send a part, keep recording
		}
		d.visitFrames = append(d.visitFrames, jpeg)
		return
	}
	perFrame := time.Duration(d.cfg.FrameInterval * float64(time.Second))
	if perFrame <= 0 {
		perFrame = 2 * time.Second
	}
	preN := int(time.Duration(d.cfg.VisitPrerollSec)*time.Second/perFrame) + 1
	d.visitPreroll = append(d.visitPreroll, jpeg)
	if len(d.visitPreroll) > preN {
		d.visitPreroll = d.visitPreroll[len(d.visitPreroll)-preN:]
	}
}

// beginVisit: first dog confirmation of a session (motion or presence path).
func (d *Daemon) beginVisit() {
	if !d.cfg.VisitVideo || d.visitActive {
		return
	}
	d.visitActive = true
	d.visitStart = time.Now()
	d.visitLastSeen = d.visitStart
	d.visitPart = 0
	d.visitThumb = nil
	d.visitFrames = append([][]byte(nil), d.visitPreroll...)
	d.visitPreroll = nil
	log.Printf("visit: recording started (pre-roll=%d frames)", len(d.visitFrames))
}

// armVisit: start recording on the first suspicious motion, BEFORE any model
// confirms a dog. On 2026-08-18 confirmation lagged arrival by 1-2 min, so the
// visit video started mid-visit and the arrival was lost. An armed visit stays
// silent; with no dog confirmation within VisitArmTimeoutSec it is dropped
// without notifying (ghost motions must never produce videos).
func (d *Daemon) armVisit(motion float64) {
	if !d.cfg.VisitVideo || d.visitActive {
		return
	}
	if motion < d.cfg.MotionThreshold {
		return
	}
	if time.Since(d.lastArmDrop) < 60*time.Second {
		return
	}
	d.beginVisit()
	log.Printf("visit: armed on motion %.1f%% — gravando silenciosamente, aguardando confirmação", motion*100)
}

func (d *Daemon) keepAliveVisit() {
	if d.visitActive {
		d.visitLastSeen = time.Now()
	}
}

func (d *Daemon) checkVisitTimeout() {
	if !d.visitActive {
		return
	}
	grace := d.cfg.VisitEndGraceSec
	if !d.visitAnnounced { // armed visit: motion started, model never confirmed
		grace = d.cfg.VisitArmTimeoutSec
	}
	if time.Since(d.visitLastSeen) > time.Duration(grace)*time.Second {
		reason := fmt.Sprintf("sem cachorro por %ds", grace)
		if !d.visitAnnounced {
			reason = fmt.Sprintf("sem confirmação por %ds", grace)
		}
		d.endVisit(reason)
	}
}

// finalizeVisitPart: visit exceeded VisitMaxSec — send what we have as a
// part and keep recording (visit stays active).
func (d *Daemon) finalizeVisitPart() {
	frames := d.visitFrames
	d.visitFrames = nil
	if len(frames) < 2 || !d.visitAnnounced {
		return // armed (unconfirmed) visits are never sent as parts
	}
	thumbFrame := d.visitThumb
	d.visitPart++
	part := d.visitPart
	cfg := d.cfg
	elapsed := time.Since(d.visitStart).Round(time.Second)
	go func() {
		path, err := muxVisitVideo(frames, cfg)
		if err != nil {
			log.Printf("visit: mux part %d: %v", part, err)
			return
		}
		caption := fmt.Sprintf("📹 *Kika continua no quintal* — parte %d (visita já durando %s)", part, elapsed)
		thumb := makeVideoThumb(thumbFrame, path)
		if err := d.sendVideo(path, caption, thumb); err != nil {
			log.Printf("visit: send part %d: %v", part, err)
		}
		os.Remove(path)
		if thumb != "" {
			os.Remove(thumb)
		}
	}()
}

// endVisit: she left (presence back to baseline, or grace expired). Mux the
// buffered frames and send the visit video.
func (d *Daemon) endVisit(reason string) {
	if !d.visitActive {
		return
	}
	frames := d.visitFrames
	d.visitFrames = nil
	d.visitActive = false
	announced := d.visitAnnounced // capture BEFORE clearing — endVisit must
	// not mistake a real, announced visit for an unconfirmed armed one
	d.visitAnnounced = false
	dur := time.Since(d.visitStart).Round(time.Second)
	log.Printf("visit: ended (%s), lasted %s, %d frames", reason, dur, len(frames))
	if !announced {
		// motion-armed recording that never got a dog confirmation: discard
		// silently — ghosts must not produce videos
		d.lastArmDrop = time.Now()
		d.visitThumb = nil
		return
	}
	if len(frames) < 2 {
		return
	}
	thumbFrame := d.visitThumb
	d.visitThumb = nil
	d.visitPart++
	cfg := d.cfg
	arrival := d.visitStart.In(d.loc)
	visitStart := d.visitStart
	segmentMode := cfg.SegmentSource && cfg.SegmentLocalDir != ""
	go func() {
		var path string
		if segmentMode {
			// wait for trailing segments to be mirrored
			time.Sleep(30 * time.Second)
			out := filepath.Join(os.TempDir(), fmt.Sprintf("visit_%d.mp4", time.Now().Unix()))
			if mErr := muxVisitSegments(cfg.SegmentLocalDir, visitStart, time.Now().In(d.loc), out); mErr == nil {
				path = out
			} else {
				log.Printf("visit: segment mux failed (%v) — fallback to jpeg mux", mErr)
			}
		}
		if path == "" {
			var err error
			path, err = muxVisitVideo(frames, cfg)
			if err != nil {
				log.Printf("visit: mux: %v", err)
				return
			}
		}
		caption := fmt.Sprintf("📹 *Visita da Kika encerrada* 🐕\nChegada: *%s* · Ficou: *%s*\nVídeo: da chegada à saída (%s)",
			arrival.Format("15:04:05"), dur, reason)
		thumb := makeVideoThumb(thumbFrame, path)
		if err := d.sendVideo(path, caption, thumb); err != nil {
			log.Printf("visit: send video: %v", err)
		}
		os.Remove(path)
		if thumb != "" {
			os.Remove(thumb)
		}
	}()
}

// makeVideoThumb: Telegram-safe thumbnail (JPEG, 320px wide, <200kB).
// Prefers the last confirmed-dog frame (d.visitThumb); falls back to a frame
// 2s into the muxed video. Returns "" on any failure — the send must never
// fail because of the thumbnail.
func makeVideoThumb(dogFrame []byte, videoPath string) string {
	out := filepath.Join(os.TempDir(), fmt.Sprintf("kikathumb_%d.jpg", time.Now().UnixNano()))
	src := videoPath
	seek := []string{"-ss", "2"}
	if len(dogFrame) > 0 {
		in := filepath.Join(os.TempDir(), fmt.Sprintf("kikathumb_in_%d.jpg", time.Now().UnixNano()))
		if err := os.WriteFile(in, dogFrame, 0o644); err == nil {
			src = in
			seek = nil
			defer os.Remove(in)
		}
	}
	args := []string{"-y", "-loglevel", "error"}
	args = append(args, seek...)
	args = append(args, "-i", src, "-frames:v", "1", "-vf", "scale=320:-2", "-q:v", "4", out)
	if o, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		log.Printf("visit thumb: %v (%s)", err, strings.TrimSpace(string(o)))
		os.Remove(out)
		return ""
	}
	if fi, err := os.Stat(out); err != nil || fi.Size() == 0 || fi.Size() > 195*1024 {
		if fi != nil && fi.Size() > 195*1024 {
			// re-encode harder — Telegram caps thumbnails at 200kB
			if o, err := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-i", out, "-vf", "scale=320:-2", "-q:v", "9", out+".small.jpg").CombinedOutput(); err == nil {
				os.Remove(out)
				return out + ".small.jpg"
			} else {
				log.Printf("visit thumb re-encode: %v (%s)", err, strings.TrimSpace(string(o)))
			}
		}
		os.Remove(out)
		return ""
	}
	return out
}

// muxVisitSegments: concat mirrored SD segments into the visit video
// (stream copy — no re-encode, full 2K quality, real 15fps motion).
// selectSegmentsInZone: pick mirrored segments whose vid_HHMMSS names fall
// inside [start, end) when interpreted in the given wall-clock zone.
func selectSegmentsInZone(localDir string, base []os.DirEntry, start, end time.Time, zone *time.Location) []string {
	var files []string
	for _, day := range base {
		if !day.IsDir() {
			continue
		}
		fs, _ := os.ReadDir(filepath.Join(localDir, day.Name()))
		for _, f := range fs {
			n := f.Name() // vid_HHMMSS.mp4
			if !strings.HasSuffix(n, ".mp4") {
				continue
			}
			var hh, mm, ss int
			if _, err := fmt.Sscanf(n, "vid_%2d%2d%2d.mp4", &hh, &mm, &ss); err != nil {
				continue
			}
			dayTime, err := time.ParseInLocation("2006-01-02 15:04:05", day.Name()+" "+fmt.Sprintf("%02d:%02d:%02d", hh, mm, ss), zone)
			if err != nil {
				continue
			}
			if dayTime.After(start) && dayTime.Before(end) {
				files = append(files, filepath.Join(localDir, day.Name(), n))
			}
		}
	}
	return files
}

func muxVisitSegments(localDir string, startTs, endTs time.Time, out string) error {
	// Camera clock runs ~20s behind the Pi and mirrored segments land with
	// poll+fetch latency; widen both edges so the visit window is fully covered.
	start := startTs.Add(-45 * time.Second) // pre-roll margin + clock skew
	end := endTs.Add(40 * time.Second)
	base, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}
	// Segment filenames carry the CAMERA's wall clock. The camera ran UTC
	// until 2026-08-18 and was then set to Brasília time (-03), so parse each
	// name in both zones and keep whichever zone yields segments — a camera
	// that reverts to GMT0 (rescue flash, factory reset) still muxes right.
	var files []string
	zoneUsed := "none"
	for _, zone := range []struct {
		name string
		z    *time.Location
	}{{"brt", mustBrazilLocation()}, {"utc", time.Local}} {
		fs := selectSegmentsInZone(localDir, base, start, end, zone.z)
		if len(fs) > 0 {
			files, zoneUsed = fs, zone.name
			break
		}
	}
	if len(files) < 1 {
		return fmt.Errorf("no segments in visit window")
	}
	log.Printf("segment mux: %d segments (parsed as %s)", len(files), zoneUsed)
	sort.Strings(files)
	list := out + ".txt"
	var sb strings.Builder
	for _, f := range files {
		fmt.Fprintf(&sb, "file '%s'\n", f)
	}
	if err := os.WriteFile(list, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	defer os.Remove(list)
	cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0",
		"-i", list,
		"-c", "copy",
		"-movflags", "+faststart",
		out)
	if o, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg concat: %w (%s)", err, strings.TrimSpace(string(o)))
	}
	return nil
}

// muxVisitVideo: JPEG frames -> MP4 via ffmpeg image2 (local files only).
// Output framerate matches the capture rate (1/FrameInterval) so the video
// plays at real time.
func muxVisitVideo(frames [][]byte, cfg Config) (string, error) {
	if len(frames) < 2 {
		return "", errors.New("visit: not enough frames")
	}
	tmp, err := os.MkdirTemp(cfg.SnapshotDir, "visit_*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	for i, f := range frames {
		p := filepath.Join(tmp, fmt.Sprintf("f%05d.jpg", i))
		if err := os.WriteFile(p, f, 0o644); err != nil {
			return "", err
		}
	}
	fps := 0.5
	if cfg.FrameInterval > 0 {
		fps = 1.0 / cfg.FrameInterval
	}
	if fps <= 0 || fps > 30 {
		fps = 0.5
	}
	out := filepath.Join(cfg.SnapshotDir, fmt.Sprintf("visit_%d.mp4", time.Now().UnixNano()))
	cmd := exec.Command(
		"ffmpeg",
		"-framerate", strconv.FormatFloat(fps, 'f', -1, 64),
		"-i", filepath.Join(tmp, "f%05d.jpg"),
		"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-y",
		out,
	)
	if o, err := cmd.CombinedOutput(); err != nil {
		os.Remove(out)
		return "", fmt.Errorf("%w (%s)", err, strings.TrimSpace(string(o)))
	}
	return out, nil
}

func (d *Daemon) logStats() {
	log.Printf("stats: vlm_mode=%s frames=%d analyzed=%d dogs=%d (yolo=%d gemini=%d) grass=%d presence[checks=%d dogs=%d cur=%v] shadow[audits=%d misses=%d] blob[triggers=%d]",
		d.cfg.VLMMode, d.stats.frames, d.stats.analyzed, d.stats.dogs, d.stats.yoloDogs, d.stats.gemDogs, d.stats.grassDogs,
		d.stats.presenceChecks, d.stats.presenceDogs, d.lastPresenceDog, d.stats.vlmAuditRuns, d.stats.vlmAuditMisses, d.stats.blobTriggers)
}

func motionPercent(prev, curr gocv.Mat, diff, gray, thresh *gocv.Mat) float64 {
	gocv.AbsDiff(prev, curr, diff)
	gocv.CvtColor(*diff, gray, gocv.ColorBGRToGray)
	gocv.Threshold(*gray, thresh, 25, 255, gocv.ThresholdBinary)

	nonZero := gocv.CountNonZero(*thresh)
	total := thresh.Rows() * thresh.Cols()
	if total <= 0 {
		return 0
	}
	return float64(nonZero) / float64(total)
}

// blobConcentration: 6x6-grid analysis of a binary threshold mask. Returns
// (maxCellPct, globalPct, ratio, changedPx, maxRow, maxCol). A still dog
// lying down = small changed area but maxCell/global ratio far above 1
// (18-22x in the 2026-08-19 03:08 case at 320x240).
func blobConcentration(thresh gocv.Mat, grid int) (maxCellPct, globalPct, ratio, changedPx float64, maxR, maxC int) {
	total := thresh.Rows() * thresh.Cols()
	changed := gocv.CountNonZero(thresh)
	if total <= 0 || changed <= 0 {
		return 0, 0, 0, 0, 0, 0
	}
	globalPct = float64(changed) / float64(total)
	ch := thresh.Rows() / grid
	cw := thresh.Cols() / grid
	maxCellPct = 0.0
	for r := 0; r < grid; r++ {
		for c := 0; c < grid; c++ {
			cell := thresh.Region(image.Rect(c*cw, r*ch, (c+1)*cw, (r+1)*ch))
			n := gocv.CountNonZero(cell)
			pct := float64(n) / float64(cw*ch)
			if pct > maxCellPct {
				maxCellPct = pct
				maxR, maxC = r, c
			}
		}
	}
	ratio = maxCellPct / math.Max(globalPct, 1e-6)
	return maxCellPct, globalPct, ratio, float64(changed), maxR, maxC
}

// motionVLMBundle: full frame + zoomed crop of the motion blob, mirroring the
// presence path. In IR darkness a lying dog is a faint smudge in the full
// frame — the crop is what lets the VLM actually see it (2026-08-19 22:57 case).
func (d *Daemon) motionVLMBundle(frame gocv.Mat, jpegBytes []byte, bok bool, bx, by, bw, bh float64) [][]byte {
	images := [][]byte{jpegBytes}
	if !bok {
		return images
	}
	if crop, err := cropPresenceRegion(frame, bx, by, bw, bh, 0.15); err == nil {
		cb, eerr := encodeJPEG(crop)
		crop.Close()
		if eerr == nil && len(cb) > 0 {
			images = append(images, cb)
		}
	}
	return images
}

// runBlobTest: debug mode --test-blob FRAMEA,FRAMEB — prints motion, changed
// px, 6x6 grid, concentration ratio and whether the blob trigger would fire.
func runBlobTest(cfg Config, pathA, pathB string) {
	imgA := gocv.IMRead(pathA, gocv.IMReadColor)
	if imgA.Empty() {
		log.Fatalf("read %s failed", pathA)
	}
	defer imgA.Close()
	imgB := gocv.IMRead(pathB, gocv.IMReadColor)
	if imgB.Empty() {
		log.Fatalf("read %s failed", pathB)
	}
	defer imgB.Close()

	smallA := gocv.NewMat()
	defer smallA.Close()
	smallB := gocv.NewMat()
	defer smallB.Close()
	gocv.Resize(imgA, &smallA, image.Pt(320, 240), 0, 0, gocv.InterpolationLinear)
	gocv.Resize(imgB, &smallB, image.Pt(320, 240), 0, 0, gocv.InterpolationLinear)

	diff := gocv.NewMat()
	defer diff.Close()
	gray := gocv.NewMat()
	defer gray.Close()
	thresh := gocv.NewMat()
	defer thresh.Close()

	motion := motionPercent(smallA, smallB, &diff, &gray, &thresh)
	maxCell, glob, ratio, px, mr, mc := blobConcentration(thresh, 6)
	fired := *cfg.BlobTriggerEnabled && ratio >= cfg.BlobTriggerRatio && px >= cfg.BlobTriggerMinPx && mr >= cfg.BlobTriggerMinRow

	ch := thresh.Rows() / 6
	cw := thresh.Cols() / 6
	log.Printf("frame A=%s B=%s", filepath.Base(pathA), filepath.Base(pathB))
	log.Printf("motion=%.3f%% (threshold %.3f%%)  changed_px=%d  max_cell=(%d,%d)=%.1f%%  global=%.2f%%  ratio=%.1fx",
		motion*100, cfg.MotionThreshold*100, int(px), mr, mc, maxCell*100, glob*100, ratio)
	var gridStr strings.Builder
	for r := 0; r < 6; r++ {
		gridStr.WriteString("  [")
		for c := 0; c < 6; c++ {
			cell := thresh.Region(image.Rect(c*cw, r*ch, (c+1)*cw, (r+1)*ch))
			fmt.Fprintf(&gridStr, "%5.1f", float64(gocv.CountNonZero(cell))/float64(cw*ch)*100)
		}
		gridStr.WriteString("]\n")
	}
	log.Printf("6x6 grid (%% changed px per cell):\n%s", gridStr.String())
	log.Printf("blob trigger: ratio %.1fx >= %.1fx && px %d >= %d && row %d >= %d → %v",
		ratio, cfg.BlobTriggerRatio, int(px), int(cfg.BlobTriggerMinPx), mr, cfg.BlobTriggerMinRow, fired)
}

// presenceDiffPercent: fraction of pixels differing from the empty-yard
// reference. A dog lying still shows up as a persistent blob of difference.
func presenceDiffPercent(ref, curr gocv.Mat, diff, gray, thresh *gocv.Mat) float64 {
	gocv.AbsDiff(ref, curr, diff)
	gocv.CvtColor(*diff, gray, gocv.ColorBGRToGray)
	gocv.Threshold(*gray, thresh, 30, 255, gocv.ThresholdBinary)
	gocv.MedianBlur(*thresh, thresh, 5)
	nonZero := gocv.CountNonZero(*thresh)
	total := thresh.Rows() * thresh.Cols()
	if total <= 0 {
		return 0
	}
	return float64(nonZero) / float64(total)
}

// adaptBackground: slowly blend current frame into the reference when the
// yard is empty, so gradual lighting/scene changes don't trigger presence.
func adaptBackground(ref *gocv.Mat, curr gocv.Mat, rate float64) {
	if ref.Empty() || rate <= 0 {
		return
	}
	alpha := 1.0 - rate
	gocv.AddWeighted(curr, rate, *ref, alpha, 0, ref)
}

// maybePresenceCheck: periodic check for a STILL dog (motionless presence).
// Uses background subtraction against the empty-yard reference; on trigger,
// confirms with YOLO then the vision LLM (GLM first, Gemini fallback).
// presenceBlobBBox: largest contour of the presence diff (>= 60 px² at
// 320x240) → normalized [0,1] bbox in the ORIGINAL frame coordinates.
func presenceBlobBBox(ref, curr gocv.Mat, diff, gray, thresh *gocv.Mat) (x, y, w, h float64, ok bool) {
	gocv.AbsDiff(ref, curr, diff)
	gocv.CvtColor(*diff, gray, gocv.ColorBGRToGray)
	gocv.Threshold(*gray, thresh, 25, 255, gocv.ThresholdBinary)
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	defer kernel.Close()
	gocv.Erode(*thresh, thresh, kernel)
	gocv.Dilate(*thresh, thresh, kernel)
	gocv.Dilate(*thresh, thresh, kernel)
	gocv.Erode(*thresh, thresh, kernel)
	contours := gocv.FindContours(*thresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()
	var best image.Rectangle
	bestArea := 0.0
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := gocv.ContourArea(c)
		if area > bestArea && area >= 60 {
			bestArea = area
			best = gocv.BoundingRect(c)
		}
	}
	if bestArea <= 0 {
		return 0, 0, 0, 0, false
	}
	sw := float64(curr.Cols())
	sh := float64(curr.Rows())
	return float64(best.Min.X) / sw, float64(best.Min.Y) / sh,
		float64(best.Dx()) / sw, float64(best.Dy()) / sh, true
}

// cropPresenceRegion: cut a zoomed crop from the original frame around the
// normalized presence bbox, expanded by margin (fraction of frame size),
// clamped to frame bounds. Caller must Close() the returned Mat.
func cropPresenceRegion(frame gocv.Mat, bx, by, bw, bh, margin float64) (gocv.Mat, error) {
	fw, fh := float64(frame.Cols()), float64(frame.Rows())
	mx := margin * fw
	my := margin * fh
	x1 := int(math.Max(0, bx*fw-mx))
	y1 := int(math.Max(0, by*fh-my))
	x2 := int(math.Min(fw, (bx+bw)*fw+mx))
	y2 := int(math.Min(fh, (by+bh)*fh+my))
	if x2-x1 < 8 || y2-y1 < 8 {
		return gocv.Mat{}, errors.New("presence crop too small")
	}
	return frame.Region(image.Rect(x1, y1, x2, y2)), nil
}

func (d *Daemon) maybePresenceCheck(frame gocv.Mat, currSmall gocv.Mat) {
	if d.cfg.PresenceInterval <= 0 {
		return
	}
	now := time.Now()
	if now.Sub(d.lastPresenceCheck) < time.Duration(d.cfg.PresenceInterval)*time.Second {
		return
	}
	d.lastPresenceCheck = now

	day := isDaytime(now.In(d.loc))
	ref := &d.bgRefDay
	if !day {
		ref = &d.bgRefNight
	}
	if ref.Empty() {
		currSmall.CopyTo(ref)
		if day {
			d.bgRefInitDay = true
		} else {
			d.bgRefInitNight = true
		}
		log.Printf("presence: %s reference captured", map[bool]string{true: "day", false: "night"}[day])
		return
	}

	diff := gocv.NewMat()
	gray := gocv.NewMat()
	thresh := gocv.NewMat()
	pdiff, pgray, pthresh := gocv.NewMat(), gocv.NewMat(), gocv.NewMat()
	defer func() {
		diff.Close()
		gray.Close()
		thresh.Close()
		pdiff.Close()
		pgray.Close()
		pthresh.Close()
	}()
	pct := presenceDiffPercent(*ref, currSmall, &diff, &gray, &thresh)
	d.stats.presenceChecks++

	// Same dog still present (confirmed recently) -> adapt reference to her
	// position so lighting drift doesn't accumulate, but don't spam-notify.
	if pct < d.cfg.PresenceDiffThresh {
		adaptBackground(ref, currSmall, d.cfg.PresenceAdaptRate)
		if d.lastPresenceDog {
			log.Printf("presence: dog left (diff=%.2f%%)", pct*100)
			d.lastPresenceDog = false
			d.endVisit("presença desapareceu da cena")
		}
		return
	}

	// Something persistently different from the empty yard.
	jpegBytes, err := encodeJPEG(frame)
	if err != nil {
		log.Printf("presence: encode jpeg: %v", err)
		return
	}

	// Blob of persistent difference (largest contour vs empty reference).
	// Computed BEFORE the VLM call so the same bbox feeds the zoomed crop
	// (second VLM image) and the training sample below.
	bx, by, bw, bh, bok := presenceBlobBBox(*ref, currSmall, &pdiff, &pgray, &pthresh)

	res := d.runYolo(frame)
	source := "yolo"
	if !res.Detected && d.vlmEnabled() {
		images := [][]byte{jpegBytes}
		if bok {
			if crop, cerr := cropPresenceRegion(frame, bx, by, bw, bh, 0.15); cerr == nil {
				cb, eerr := encodeJPEG(crop)
				crop.Close()
				if eerr == nil && len(cb) > 0 {
					images = append(images, cb)
				}
			}
		}
		vres := d.runVisionLLM(images)
		if vres.Err != nil {
			log.Printf("presence: vision err: %v", vres.Err)
		}
		if d.vlmShadow() {
			// SHADOW: YOLO is the decisor. Harvest the miss, do not adopt VLM verdict.
			d.stats.vlmAuditRuns++
			if vres.Err == nil && vres.Detected {
				d.stats.vlmAuditMisses++
				log.Printf("shadow: presence VLM saw a dog YOLO missed; sample harvested, not notifying")
				d.savePresenceSample(frame, bx, by, bw, bh, "shadow_vlm")
			}
		} else if vres.Detected {
			res = vres
			source = "visionllm"
		}
	}

	if !res.Detected {
		// Unknown persistent object (chair moved, etc.) — adopt it into the
		// reference slowly so it stops triggering.
		log.Printf("presence: %.2f%% diff, no dog -> adapting (source=%s)", pct*100, source)
		adaptBackground(ref, currSmall, d.cfg.PresenceAdaptRate*4)
		return
	}

	d.stats.presenceDogs++
	wasDog := d.lastPresenceDog
	d.lastPresenceDog = true
	d.keepAliveVisit()
	d.visitThumb = jpegBytes
	log.Printf("presence: DOG still present (diff=%.2f%%, source=%s, onGrass=%v)", pct*100, source, res.OnGrass)

	// auto-capture training sample: full-res frame + YOLO label from diff blob
	if bok {
		d.savePresenceSample(frame, bx, by, bw, bh, source)
	}

	nowBr := now.In(d.loc)
	shouldPhoto := !wasDog && !(d.visitActive && d.visitAnnounced)
	if d.cfg.VisitVideo && !d.visitActive {
		// The visit was grace-ended while she stayed put: quiet dogs stop
		// triggering motion and YOLO can't see her lying down. Presence says
		// she is still here — resume recording silently, without a second
		// arrival photo for the same presence streak.
		d.beginVisit()
		d.visitAnnounced = true
		log.Printf("visit: resumed by presence confirm (quiet stay survived grace)")
	}
	if shouldPhoto {
		// first confirmation of this visit — skipped when the motion path
		// already announced it (one arrival photo per visit)
		if d.cfg.VisitVideo {
			if !d.visitActive {
				d.beginVisit()
			}
			d.visitAnnounced = true
		}
		photoJPEG := d.arrivalPhotoJPEG(jpegBytes)
		snapshotPath, err := d.saveSnapshot(photoJPEG, nowBr, pct)
		if err != nil {
			log.Printf("presence: save snapshot: %v", err)
			return
		}
		place := "na área interna 🏠"
		if res.OnGrass {
			place = "NA GRAMA 🌱"
		}
		caption := fmt.Sprintf("🐕 *Kika parada no quintal* (presença)\nHorário: *%s (Brasília)*\nLocal: *%s*\nDiff: *%.1f%%*\nGrama: *%.0f%%*\nDetector: %s",
			nowBr.Format("15:04:05"), place, pct*100, res.GrassFrac*100, source)
		if d.cfg.VisitVideo {
			caption += "\n🎬 Gravando — vídeo completo quando ela sair"
		}
		if err := d.sendPhoto(snapshotPath, caption); err != nil {
			log.Printf("presence: notify error: %v", err)
		}
	}
	// While she stays: keep reference as-is (don't adapt her blob away) but
	// still blend lightly to survive slow light drift around her.
	adaptBackground(ref, currSmall, d.cfg.PresenceAdaptRate*0.25)
}

func encodeJPEG(frame gocv.Mat) ([]byte, error) {
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, frame)
	if err != nil {
		return nil, err
	}
	defer buf.Close()
	out := make([]byte, buf.Len())
	copy(out, buf.GetBytes())
	return out, nil
}

func (d *Daemon) runYolo(frame gocv.Mat) DetectorResult {
	start := time.Now()
	tensorData, err := preprocessYOLO(frame, d.cfg.YOLOImgsz)
	if err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}

	inputTensor, err := ort.NewTensor(ort.NewShape(1, 3, int64(d.cfg.YOLOImgsz), int64(d.cfg.YOLOImgsz)), tensorData)
	if err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}
	defer inputTensor.Destroy()

	inputs := []ort.Value{inputTensor}
	outputs := []ort.Value{nil}
	if err := d.yolo.session.Run(inputs, outputs); err != nil {
		return DetectorResult{Err: fmt.Errorf("onnx run: %w", err), Latency: time.Since(start)}
	}
	if len(outputs) == 0 || outputs[0] == nil {
		return DetectorResult{Err: errors.New("empty yolo output"), Latency: time.Since(start)}
	}
	defer outputs[0].Destroy()

	t, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return DetectorResult{Err: fmt.Errorf("unexpected output type %T", outputs[0]), Latency: time.Since(start)}
	}

	dets, err := parseYoloDetections(t.GetData(), t.GetShape(), d.cfg.YOLOConf)
	if err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}
	if len(dets) == 0 {
		return DetectorResult{Detected: false, Source: "yolo", Latency: time.Since(start)}
	}

	// Map best detection from letterboxed input space to frame space.
	size := d.cfg.YOLOImgsz
	fw, fh := frame.Cols(), frame.Rows()
	r := math.Min(float64(size)/float64(fw), float64(size)/float64(fh))
	padX := (float64(size) - math.Round(float64(fw)*r)) / 2
	padY := (float64(size) - math.Round(float64(fh)*r)) / 2

	best := dets[0]
	for _, dt := range dets[1:] {
		if dt.Conf > best.Conf {
			best = dt
		}
	}
	toFrame := func(v, pad float64, max int) int {
		out := (v - pad) / r
		if out < 0 {
			out = 0
		}
		if out > float64(max-1) {
			out = float64(max - 1)
		}
		return int(out)
	}
	x1 := toFrame(float64(best.X1), padX, fw)
	x2 := toFrame(float64(best.X2), padX, fw)
	y1 := toFrame(float64(best.Y1), padY, fh)
	y2 := toFrame(float64(best.Y2), padY, fh)
	if x2 <= x1 {
		x2 = x1 + 1
	}
	if y2 <= y1 {
		y2 = y1 + 1
	}

	// Grass check: region around the dog's lower body (where she stands/lies).
	gy1 := y1 + (y2-y1)/2
	gy2 := y2 + (y2-y1)/4
	if gy2 > fh {
		gy2 = fh
	}
	frac := grassGreenFraction(frame, x1, gy1, x2, gy2)

	return DetectorResult{
		Detected:  true,
		OnGrass:   frac >= d.cfg.GrassMinGreenFrac,
		GrassFrac: frac,
		Source:    "yolo",
		Latency:   time.Since(start),
	}
}

type yoloDetection struct {
	Conf           float32
	X1, Y1, X2, Y2 float32
}

func parseYoloDetections(data []float32, shape ort.Shape, conf float32) ([]yoloDetection, error) {
	if len(shape) < 3 {
		return nil, fmt.Errorf("unexpected yolo shape: %v", shape)
	}

	// YOLOv8 output shape: [1, 4+C, N] where C=num_classes, N=num_detections
	// Data layout: data[channel * N + detection_idx]
	numClasses := shape[1] - 4 // e.g. 1 for fine-tuned kika, 80 for COCO
	numDets := shape[2]
	if numClasses < 1 || numDets < 1 {
		return nil, fmt.Errorf("unexpected yolo shape: %v", shape)
	}
	totalElements := int64(len(data))
	var out []yoloDetection
	for j := int64(0); j < numDets; j++ {
		var bestConf float32
		bestClass := int64(-1)
		for c := int64(0); c < numClasses; c++ {
			idx := (4+c)*numDets + j
			if idx >= 0 && idx < totalElements && data[idx] >= conf && data[idx] > bestConf {
				bestConf = data[idx]
				bestClass = c
			}
		}
		if bestClass < 0 {
			continue
		}
		cx := data[0*numDets+j]
		cy := data[1*numDets+j]
		w := data[2*numDets+j]
		h := data[3*numDets+j]
		out = append(out, yoloDetection{
			Conf: bestConf,
			X1:   cx - w/2,
			Y1:   cy - h/2,
			X2:   cx + w/2,
			Y2:   cy + h/2,
		})
	}
	return out, nil
}

// grassGreenFraction returns the fraction of green "grass-like" pixels (HSV)
// inside the given pixel region of a BGR frame.
func grassGreenFraction(frame gocv.Mat, x1, y1, x2, y2 int) float64 {
	fw, fh := frame.Cols(), frame.Rows()
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > fw {
		x2 = fw
	}
	if y2 > fh {
		y2 = fh
	}
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	roi := frame.Region(image.Rect(x1, y1, x2, y2))
	defer roi.Close()

	hsv := gocv.NewMat()
	defer hsv.Close()
	gocv.CvtColor(roi, &hsv, gocv.ColorBGRToHSV)

	mask := gocv.NewMat()
	defer mask.Close()
	gocv.InRangeWithScalar(hsv, gocv.NewScalar(35, 40, 30, 0), gocv.NewScalar(85, 255, 255, 0), &mask)

	total := (x2 - x1) * (y2 - y1)
	if total <= 0 {
		return 0
	}
	return float64(gocv.CountNonZero(mask)) / float64(total)
}

func preprocessYOLO(frame gocv.Mat, size int) ([]float32, error) {
	if frame.Empty() {
		return nil, errors.New("empty frame")
	}
	w := frame.Cols()
	h := frame.Rows()
	if w <= 0 || h <= 0 {
		return nil, errors.New("invalid frame size")
	}

	r := math.Min(float64(size)/float64(w), float64(size)/float64(h))
	newW := int(math.Round(float64(w) * r))
	newH := int(math.Round(float64(h) * r))
	if newW <= 0 {
		newW = 1
	}
	if newH <= 0 {
		newH = 1
	}

	resized := gocv.NewMat()
	defer resized.Close()
	gocv.Resize(frame, &resized, image.Pt(newW, newH), 0, 0, gocv.InterpolationLinear)

	canvas := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(128, 128, 128, 0), size, size, gocv.MatTypeCV8UC3)
	defer canvas.Close()

	x := (size - newW) / 2
	y := (size - newH) / 2
	roi := canvas.Region(image.Rect(x, y, x+newW, y+newH))
	resized.CopyTo(&roi)
	roi.Close()

	pix, err := canvas.DataPtrUint8()
	if err != nil {
		return nil, err
	}

	plane := size * size
	out := make([]float32, 3*plane)
	for i := 0; i < plane; i++ {
		b := pix[i*3+0]
		g := pix[i*3+1]
		rv := pix[i*3+2]
		out[0*plane+i] = float32(rv) / 255.0
		out[1*plane+i] = float32(g) / 255.0
		out[2*plane+i] = float32(b) / 255.0
	}
	return out, nil
}

func (d *Daemon) runGemini(images [][]byte) DetectorResult {
	start := time.Now()
	if strings.TrimSpace(d.cfg.GeminiAPIKey) == "" {
		return DetectorResult{Err: errors.New("empty gemini api key"), Latency: time.Since(start)}
	}

	parts := make([]any, 0, len(images)+1)
	parts = append(parts, map[string]any{"text": "You are detecting a small dog in a surveillance frame. Reply with strict JSON: {\"detected\":bool,\"grass\":bool} where detected=true only if a DOG is clearly visible (canine anatomy: head with snout and ears, four legs, tail), and grass=true means the dog is on or over living grass (lawn). Known false positives — detected=false for: dark blurry mass in bottom-left corner (fixed object near lens), dark shaggy texture on the white-bordered rug (rug/pet-bed, not a dog), shadows, bags, clothes, furniture. Not a cat. When in doubt, detected=false."})
	for _, img := range images {
		parts = append(parts, map[string]any{"inline_data": map[string]any{"mime_type": "image/jpeg", "data": base64.StdEncoding.EncodeToString(img)}})
	}
	payload := map[string]any{
		"contents": []any{
			map[string]any{
				"parts": parts,
			},
		},
	}
	body, _ := json.Marshal(payload)
	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-latest:generateContent?key=" + url.QueryEscape(d.cfg.GeminiAPIKey)
	req, _ := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return DetectorResult{Err: fmt.Errorf("gemini status: %s", resp.Status), Latency: time.Since(start)}
	}

	txt := extractGeminiText(respBody)
	if strings.TrimSpace(txt) == "" {
		return DetectorResult{Err: errors.New("empty gemini response"), Latency: time.Since(start)}
	}
	jsonText := extractJSON(txt)
	var parsed struct {
		Detected bool `json:"detected"`
		Grass    bool `json:"grass"`
	}
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}
	grassFrac := 0.0
	if parsed.Grass {
		grassFrac = 1.0
	}
	return DetectorResult{Detected: parsed.Detected, OnGrass: parsed.Grass, GrassFrac: grassFrac, Source: "gemini", Latency: time.Since(start)}
}

// runGLM queries Z.AI glm-4.5v (coding plan) — no 15 RPM cap like Gemini.
func (d *Daemon) runGLM(images [][]byte) DetectorResult {
	start := time.Now()
	if strings.TrimSpace(d.cfg.GLMAPIKey) == "" {
		return DetectorResult{Err: errors.New("empty glm api key"), Latency: time.Since(start)}
	}
	content := make([]any, 0, len(images)+1)
	for _, img := range images {
		content = append(content, map[string]any{"type": "image_url", "image_url": map[string]any{
			"url": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(img)}})
	}
	if len(images) > 1 {
		content = append(content, map[string]any{"type": "text", "text": `You are detecting a specific small dog named Kika in surveillance images. Kika is a small black-and-white dog (mostly black body, white chest/paws, reddish collar), much smaller than a person. If more than one image is given, later images are zoomed crops of the region that changed. Reply with strict JSON only: {"detected":bool,"grass":bool} where detected=true only if a DOG is clearly visible in any image (visible canine anatomy: head with snout and ears, four legs, tail, fur), and grass=true means the dog is on or over living grass (lawn). IMPORTANT known false positives in this scene — reply detected=false for: the large dark blurry out-of-focus mass in the bottom-left corner (a fixed object near the lens, NOT a dog); the dark shaggy area on the white-bordered rug (a coarse rug/pet-bed texture, NOT a dog); shadows, bags, clothes, shoes, furniture. Not a cat. When in doubt, reply detected=false.`})
	} else {
		content = append(content, map[string]any{"type": "text", "text": `You are detecting a specific small dog named Kika in a surveillance frame. Kika is a small black-and-white dog (mostly black body, white chest/paws, reddish collar), much smaller than a person. Reply with strict JSON only: {"detected":bool,"grass":bool} where detected=true only if a DOG is clearly visible (visible canine anatomy: head with snout and ears, four legs, tail, fur), and grass=true means the dog is on or over living grass (lawn). IMPORTANT known false positives in this scene — reply detected=false for: the large dark blurry out-of-focus mass in the bottom-left corner (a fixed object near the lens, NOT a dog); the dark shaggy area on the white-bordered rug (a coarse rug/pet-bed texture, NOT a dog); shadows, bags, clothes, shoes, furniture. Not a cat. When in doubt, reply detected=false.`})
	}
	payload := map[string]any{
		"model":      "glm-4.5v",
		"messages":   []any{map[string]any{"role": "user", "content": content}},
		"max_tokens": 300,
		"thinking":   map[string]any{"type": "disabled"},
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "https://api.z.ai/api/coding/paas/v4/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.cfg.GLMAPIKey)
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return DetectorResult{Err: fmt.Errorf("glm status: %s body: %.200s", resp.Status, string(respBody)), Latency: time.Since(start)}
	}
	var cr struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &cr); err != nil || len(cr.Choices) == 0 {
		return DetectorResult{Err: fmt.Errorf("glm parse: %v", err), Latency: time.Since(start)}
	}
	jsonText := extractJSON(cr.Choices[0].Message.Content)
	var parsed struct {
		Detected bool `json:"detected"`
		Grass    bool `json:"grass"`
	}
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return DetectorResult{Err: err, Latency: time.Since(start)}
	}
	grassFrac := 0.0
	if parsed.Grass {
		grassFrac = 1.0
	}
	return DetectorResult{Detected: parsed.Detected, OnGrass: parsed.Grass, GrassFrac: grassFrac, Source: "glm", Latency: time.Since(start)}
}

// runVisionLLM: GLM first (higher quota), Gemini fallback.
// Accepts one or more images (full frame + zoomed presence crop).
func (d *Daemon) runVisionLLM(images [][]byte) DetectorResult {
	if len(images) == 0 {
		return DetectorResult{Err: errors.New("no images"), Latency: 0}
	}
	if strings.TrimSpace(d.cfg.GLMAPIKey) != "" {
		var lastErr error
		for attempt := 1; attempt <= 2; attempt++ {
			res := d.runGLM(images)
			if res.Err == nil {
				return res
			}
			lastErr = res.Err
			if attempt == 1 {
				log.Printf("vision glm attempt 1 failed (%v), retrying once", res.Err)
				time.Sleep(2 * time.Second)
			}
		}
		log.Printf("vision glm failed after retry (%v), falling back to gemini", lastErr)
	}
	return d.runGemini(images)
}

func extractGeminiText(raw []byte) string {
	var r map[string]any
	if err := json.Unmarshal(raw, &r); err != nil {
		return ""
	}
	cands, ok := r["candidates"].([]any)
	if !ok || len(cands) == 0 {
		return ""
	}
	cand, ok := cands[0].(map[string]any)
	if !ok {
		return ""
	}
	content, ok := cand["content"].(map[string]any)
	if !ok {
		return ""
	}
	parts, ok := content["parts"].([]any)
	if !ok {
		return ""
	}
	var out strings.Builder
	for _, p := range parts {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		t, _ := m["text"].(string)
		out.WriteString(t)
	}
	return out.String()
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

func (d *Daemon) logMotion(motion float64, r DetectorResult) {
	ms := r.Latency.Seconds() * 1000
	if r.Err != nil {
		if motion >= d.cfg.GeminiNightMinMotion {
			log.Printf("motion %.1f%% without dog (YOLO(night,fallback))", motion*100)
		} else {
			log.Printf("motion %.1f%% without dog (YOLO(error))", motion*100)
		}
		return
	}
	if r.Detected {
		log.Printf("motion %.1f%% with dog (%s pos,%.0fms, grass=%.0f%%, onGrass=%v)", motion*100, r.Source, ms, r.GrassFrac*100, r.OnGrass)
		return
	}
	log.Printf("motion %.1f%% without dog (YOLO(neg,%.0fms))", motion*100, ms)
}

func (d *Daemon) saveSnapshot(jpeg []byte, now time.Time, motion float64) (string, error) {
	name := fmt.Sprintf("%s_%d_%.0fpct.jpg", now.Format("20060102_150405"), time.Now().UnixNano(), motion*100)
	path := filepath.Join(d.cfg.SnapshotDir, name)
	if err := os.WriteFile(path, jpeg, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// grabLiveJPEG: fetch a fresh frame from the camera's HTTP snapshot endpoint.
// Used for arrival photos so the photo reflects the moment of detection even
// when the analysis frame is a mirrored SD segment (~25-95s behind live).
// Falls back to (nil, err); caller keeps the analysis frame on failure.
func (d *Daemon) grabLiveJPEG() ([]byte, error) {
	if d.cfg.SnapshotURL == "" {
		return nil, errors.New("no snapshot_url configured")
	}
	client := &http.Client{Timeout: 6 * time.Second}
	req, err := http.NewRequest(http.MethodGet, d.cfg.SnapshotURL, nil)
	if err != nil {
		return nil, err
	}
	if u, p, ok := basicAuthFromURL(d.cfg.SnapshotURL); ok {
		req.SetBasicAuth(u, p)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("snapshot http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if len(b) < 1024 {
		return nil, fmt.Errorf("snapshot too small (%d bytes)", len(b))
	}
	return b, nil
}

// arrivalPhotoJPEG: live frame if available, else the analyzed frame.
func (d *Daemon) arrivalPhotoJPEG(analyzed []byte) []byte {
	if live, err := d.grabLiveJPEG(); err == nil {
		return live
	} else {
		log.Printf("arrival photo: live grab failed (%v) — using analyzed frame", err)
	}
	if len(analyzed) > 0 {
		return analyzed
	}
	return nil
}

func basicAuthFromURL(raw string) (user, pass string, ok bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", false
	}
	if u.User == nil {
		return "", "", false
	}
	p, _ := u.User.Password()
	return u.User.Username(), p, true
}

func (d *Daemon) sendPhoto(path, caption string) error {
	for _, chatID := range d.cfg.TelegramChatIDs {
		if d.cfg.MediaRelayHost != "" {
			if err := d.sendViaRelay("sendPhoto", "photo", path, chatID, caption, ""); err == nil {
				continue
			} else {
				log.Printf("relay sendPhoto failed, falling back to direct: %v", err)
			}
		}
		if err := d.sendTelegramFile("sendPhoto", "photo", path, chatID, caption, ""); err != nil {
			return err
		}
	}
	return nil
}

func (d *Daemon) sendVideo(path, caption, thumbPath string) error {
	for _, chatID := range d.cfg.TelegramChatIDs {
		if d.cfg.MediaRelayHost != "" {
			if err := d.sendViaRelay("sendVideo", "video", path, chatID, caption, thumbPath); err == nil {
				continue
			} else {
				log.Printf("relay sendVideo failed, falling back to direct: %v", err)
			}
		}
		if err := d.sendTelegramFile("sendVideo", "video", path, chatID, caption, thumbPath); err != nil {
			return err
		}
	}
	return nil
}

func (d *Daemon) sendViaRelay(method, fieldName, path string, chatID int64, caption, thumbPath string) error {
	remote := d.cfg.MediaRelayHost + ":/tmp/kika_relay_" + filepath.Base(path)
	scpCmd := exec.Command("scp", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=accept-new", path, remote)
	if out, err := scpCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("relay scp: %v: %s", err, strings.TrimSpace(string(out)))
	}
	base := filepath.Base(path)
	thumbPart := ""
	thumbRemote := ""
	if thumbPath != "" {
		thumbRemote = "/tmp/kika_relay_thumb_" + filepath.Base(thumbPath)
		if out, err := exec.Command("scp", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=accept-new", thumbPath, d.cfg.MediaRelayHost+":"+thumbRemote).CombinedOutput(); err != nil {
			log.Printf("relay thumb scp: %v (%s) — enviando sem thumbnail", err, strings.TrimSpace(string(out)))
		} else {
			thumbPart = fmt.Sprintf(" -F thumbnail=@%s", thumbRemote)
		}
	}
	curlScript := fmt.Sprintf(
		"curl -s -m 120 -X POST https://api.telegram.org/bot%s/%s -F chat_id=%d -F caption=%q -F parse_mode=Markdown -F %s=@/tmp/kika_relay_%s%s",
		d.cfg.TelegramBotToken, method, chatID, caption, fieldName, base, thumbPart)
	sshCmd := exec.Command("ssh", "-o", "ConnectTimeout=10", d.cfg.MediaRelayHost, curlScript)
	if out, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("relay ssh-curl: %v: %s", err, strings.TrimSpace(string(out)))
	}
	cleanupCmd := exec.Command("ssh", "-o", "ConnectTimeout=10", d.cfg.MediaRelayHost, "rm -f /tmp/kika_relay_"+base+" "+thumbRemote)
	_ = cleanupCmd.Run()
	return nil
}

func (d *Daemon) sendTelegramFile(method, fieldName, path string, chatID int64, caption, thumbPath string) error {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	_ = mw.WriteField("caption", caption)
	_ = mw.WriteField("parse_mode", "Markdown")
	fw, err := mw.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	if _, err := fw.Write(fileData); err != nil {
		return err
	}
	if thumbPath != "" {
		// Telegram video thumbnail: JPEG, <=320px wide, <200kB
		if tdata, terr := os.ReadFile(thumbPath); terr == nil {
			if tw, twerr := mw.CreateFormFile("thumbnail", filepath.Base(thumbPath)); twerr == nil {
				_, _ = tw.Write(tdata)
			}
		}
	}
	if err := mw.Close(); err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/%s", d.cfg.TelegramBotToken, method)
	req, _ := http.NewRequest(http.MethodPost, endpoint, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram %s: %s %s", method, resp.Status, strings.TrimSpace(string(b)))
	}
	return nil
}

func (d *Daemon) tryDatasetCapture(frame gocv.Mat) {
	if d.cfg.DatasetV3Dir == "" || d.cfg.DatasetV3Max <= 0 {
		return
	}
	now := time.Now().In(d.loc)
	if d.cfg.DatasetV3DaytimeOnly && !isDaytime(now) {
		return
	}
	if !d.lastDataset.IsZero() && now.Sub(d.lastDataset) < time.Duration(d.cfg.DatasetV3Interval)*time.Second {
		return
	}

	files, err := os.ReadDir(d.cfg.DatasetV3Dir)
	if err == nil && len(files) >= d.cfg.DatasetV3Max {
		return
	}
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, frame)
	if err != nil {
		return
	}
	defer buf.Close()

	name := fmt.Sprintf("motion_%d.jpg", now.Unix())
	path := filepath.Join(d.cfg.DatasetV3Dir, name)
	if err := os.WriteFile(path, append([]byte(nil), buf.GetBytes()...), 0o644); err != nil {
		return
	}
	d.lastDataset = now
}

// savePresenceSample: saves full-res frame + normalized YOLO label derived
// from the presence diff blob — labeled training data for free, any position.
func (d *Daemon) savePresenceSample(frame gocv.Mat, bx, by, bw, bh float64, source string) {
	dir := d.cfg.DatasetV3Dir
	if dir == "" {
		return
	}
	posDir := filepath.Join(dir, "presence_pos")
	if err := os.MkdirAll(posDir, 0o755); err != nil {
		return
	}
	entries, err := os.ReadDir(posDir)
	if err == nil && len(entries) >= d.cfg.DatasetV3Max {
		return // cap reached
	}
	buf, err := gocv.IMEncode(gocv.JPEGFileExt, frame)
	if err != nil {
		return
	}
	defer buf.Close()
	ts := time.Now().In(d.loc)
	stem := fmt.Sprintf("pres_%s_%d", ts.Format("20060102_150405"), ts.UnixNano())
	imgPath := filepath.Join(posDir, stem+".jpg")
	if err := os.WriteFile(imgPath, append([]byte(nil), buf.GetBytes()...), 0o644); err != nil {
		return
	}
	// YOLO label: class 0 (kika), center + size normalized
	cx := bx + bw/2
	cy := by + bh/2
	label := fmt.Sprintf("0 %.4f %.4f %.4f %.4f\n", cx, cy, bw, bh)
	lblPath := filepath.Join(posDir, stem+".txt")
	if err := os.WriteFile(lblPath, []byte(label), 0o644); err != nil {
		return
	}
	log.Printf("presence: training sample saved %s (%s, bbox=%.2f,%.2f %.2fx%.2f)", stem+".jpg", source, bx, by, bw, bh)
}

func isDaytime(t time.Time) bool {
	h := t.Hour()
	return h >= 6 && h < 18
}
