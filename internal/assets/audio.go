package assets

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// DefaultSampleRate is the standard sample rate for audio.
	DefaultSampleRate = 44100
)

// ReadSeekerLength combines ReadSeeker with Length method.
type ReadSeekerLength interface {
	io.ReadSeeker
	Length() int64
}

// AudioManager handles loading and playing sounds and music.
type AudioManager struct {
	context *audio.Context
	sounds  map[string]*audio.Player
	music   map[string]*audio.Player
	pools   map[string]*SoundPool
	fs      fs.FS

	// Volume controls (0.0 to 1.0)
	masterVolume float64
	sfxVolume    float64
	musicVolume  float64
}

// SoundPool manages multiple players for concurrent playback of the same sound.
type SoundPool struct {
	players []*audio.Player
	current int
}

// NewAudioManager creates an audio manager.
// filesystem can be nil if you only plan to load from bytes.
func NewAudioManager(filesystem fs.FS) *AudioManager {
	return &AudioManager{
		context:      audio.NewContext(DefaultSampleRate),
		sounds:       make(map[string]*audio.Player),
		music:        make(map[string]*audio.Player),
		pools:        make(map[string]*SoundPool),
		fs:           filesystem,
		masterVolume: 1.0,
		sfxVolume:    1.0,
		musicVolume:  1.0,
	}
}

// SetMasterVolume sets the global master volume.
func (m *AudioManager) SetMasterVolume(volume float64) {
	m.masterVolume = m.clampVolume(volume)
	m.updateMusicVolume()
}

// SetSFXVolume sets the global sound effects volume.
func (m *AudioManager) SetSFXVolume(volume float64) {
	m.sfxVolume = m.clampVolume(volume)
}

// SetMusicVolume sets the global music volume.
func (m *AudioManager) SetMusicVolume(volume float64) {
	m.musicVolume = m.clampVolume(volume)
	m.updateMusicVolume()
}

// GetMasterVolume returns the global master volume.
func (m *AudioManager) GetMasterVolume() float64 {
	return m.masterVolume
}

// GetSFXVolume returns the global sound effects volume.
func (m *AudioManager) GetSFXVolume() float64 {
	return m.sfxVolume
}

// GetMusicVolume returns the global music volume.
func (m *AudioManager) GetMusicVolume() float64 {
	return m.musicVolume
}

// SetMusicTrackVolume sets the volume for a specific music track.
func (m *AudioManager) SetMusicTrackVolume(name string, volume float64) {
	if player, ok := m.music[name]; ok {
		// Note: This overrides global music volume scaling for this track if used directly?
		// Or should it be relative?
		// For simplicity/compat, let's set it directly but considering global?
		// The previous implementation was direct.
		// If we want it to work with global volume, we should store a "base volume" for the track.
		// But we don't store per-track base volume.
		// So let's just set it relative to master * global music?
		// If the user calls this, they likely want a specific level.
		val := m.clampVolume(volume)
		player.SetVolume(val * m.masterVolume * m.musicVolume)
	}
}

func (m *AudioManager) clampVolume(v float64) float64 {
	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return v
}

func (m *AudioManager) updateMusicVolume() {
	vol := m.masterVolume * m.musicVolume
	for _, p := range m.music {
		if p.IsPlaying() {
			p.SetVolume(vol)
		}
	}
}

// LoadSound loads a sound effect from file.
func (m *AudioManager) LoadSound(name, path string) error {
	if m.fs == nil {
		return errors.New("filesystem is nil")
	}

	data, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read audio %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	return m.LoadSoundFromBytes(name, data, ext)
}

// LoadSoundFromBytes loads a sound effect from raw data.
// format should be ".wav", ".mp3", or ".ogg".
func (m *AudioManager) LoadSoundFromBytes(name string, data []byte, format string) error {
	stream, err := m.decodeStream(format, data)
	if err != nil {
		return err
	}

	player, err := m.context.NewPlayer(stream)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %w", err)
	}

	m.sounds[name] = player

	return nil
}

// LoadMusic loads a music track from file.
func (m *AudioManager) LoadMusic(name, path string) error {
	if m.fs == nil {
		return errors.New("filesystem is nil")
	}

	data, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read music %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	return m.LoadMusicFromBytes(name, data, ext)
}

// LoadMusicFromBytes loads a music track from raw data.
func (m *AudioManager) LoadMusicFromBytes(name string, data []byte, format string) error {
	stream, err := m.decodeStream(format, data)
	if err != nil {
		return err
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())

	player, err := m.context.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("failed to create music player: %w", err)
	}

	m.music[name] = player

	return nil
}

// CreatePool creates a sound pool from file.
func (m *AudioManager) CreatePool(name, path string, size int) error {
	if m.fs == nil {
		return errors.New("filesystem is nil")
	}

	data, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read audio %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	return m.CreatePoolFromBytes(name, data, size, ext)
}

// CreatePoolFromBytes creates a sound pool from raw data.
func (m *AudioManager) CreatePoolFromBytes(name string, data []byte, size int, format string) error {
	pool := &SoundPool{
		players: make([]*audio.Player, size),
		current: 0,
	}

	for i := range size {
		stream, err := m.decodeStream(format, data)
		if err != nil {
			return err
		}

		player, err := m.context.NewPlayer(stream)
		if err != nil {
			return fmt.Errorf("failed to create pool player: %w", err)
		}

		pool.players[i] = player
	}

	m.pools[name] = pool

	return nil
}

func (m *AudioManager) decodeStream(format string, data []byte) (ReadSeekerLength, error) {
	format = strings.ToLower(format)
	if !strings.HasPrefix(format, ".") {
		format = "." + format
	}

	reader := bytes.NewReader(data)

	switch format {
	case ".wav":
		return wav.DecodeWithSampleRate(DefaultSampleRate, reader)
	case ".mp3":
		return mp3.DecodeWithSampleRate(DefaultSampleRate, reader)
	case ".ogg":
		return vorbis.DecodeWithSampleRate(DefaultSampleRate, reader)
	default:
		return nil, fmt.Errorf("unsupported audio format: %s", format)
	}
}

// PlaySound plays a sound effect once.
func (m *AudioManager) PlaySound(name string) {
	if player, ok := m.sounds[name]; ok {
		if !player.IsPlaying() {
			player.Rewind()
			player.SetVolume(m.masterVolume * m.sfxVolume)
			player.Play()
		} else {
			player.Rewind()
			player.SetVolume(m.masterVolume * m.sfxVolume)
			player.Play()
		}
	}
}

// PlaySoundWithVolume plays a sound with a specific volume (0.0 to 1.0).
func (m *AudioManager) PlaySoundWithVolume(name string, volume float64) {
	if player, ok := m.sounds[name]; ok {
		player.Rewind()
		player.SetVolume(volume * m.masterVolume * m.sfxVolume)
		player.Play()
	}
}

// PlayPooled plays a sound from a pool.
func (m *AudioManager) PlayPooled(name string) {
	pool, ok := m.pools[name]
	if !ok {
		return
	}

	player := pool.players[pool.current]
	pool.current = (pool.current + 1) % len(pool.players)

	player.Rewind()
	player.SetVolume(m.masterVolume * m.sfxVolume)
	player.Play()
}

// PlayMusic starts playing a music track.
func (m *AudioManager) PlayMusic(name string) {
	if player, ok := m.music[name]; ok {
		player.SetVolume(m.masterVolume * m.musicVolume)

		if !player.IsPlaying() {
			player.Rewind()
			player.Play()
		}
	}
}

// StopMusic stops the current music.
func (m *AudioManager) StopMusic(name string) {
	if player, ok := m.music[name]; ok {
		player.Pause()
	}
}

// IsMusicPlaying returns true if the named music is playing.
func (m *AudioManager) IsMusicPlaying(name string) bool {
	if player, ok := m.music[name]; ok {
		return player.IsPlaying()
	}

	return false
}

// FadeMusic fades music volume over duration.
func (m *AudioManager) FadeMusic(name string, targetVolume float64, duration time.Duration) {
	player, ok := m.music[name]
	if !ok {
		return
	}

	go func() {
		startVolume := player.Volume()
		steps := 20
		stepDuration := duration / time.Duration(steps)
		// Global volume scaling
		finalTarget := targetVolume * m.masterVolume * m.musicVolume

		volumeStep := (finalTarget - startVolume) / float64(steps)

		for i := range steps {
			time.Sleep(stepDuration)
			player.SetVolume(startVolume + volumeStep*float64(i+1))
		}

		player.SetVolume(finalTarget)
	}()
}

// CrossfadeMusic fades out current music and fades in new music.
func (m *AudioManager) CrossfadeMusic(fromName, toName string, duration time.Duration) {
	// Fade out current music
	if fromName != "" {
		m.FadeMusic(fromName, 0, duration)
	}

	// Start and fade in new music
	if player, ok := m.music[toName]; ok {
		player.SetVolume(0)
		player.Rewind()
		player.Play()
		// Fade to full (which is 1.0 * global)
		// We pass 1.0 to FadeMusic, it calculates the global scaling
		m.FadeMusic(toName, 1.0, duration)
	}
}

// Context returns the audio context.
func (m *AudioManager) Context() *audio.Context {
	return m.context
}

// PlaySoundAt plays a sound with 2D spatial panning.
func (m *AudioManager) PlaySoundAt(name string, x, y, listenerX, listenerY, maxDistance float64) {
	pool, ok := m.pools[name]
	if !ok {
		// Try single sound if pool not found
		if player, ok := m.sounds[name]; ok {
			dx := x - listenerX
			dy := y - listenerY
			dist := math.Sqrt(dx*dx + dy*dy)

			volume := 1.0 - (dist / maxDistance)
			if volume < 0 {
				volume = 0
			}

			player.Rewind()
			player.SetVolume(volume * m.masterVolume * m.sfxVolume)
			player.Play()
		}

		return
	}

	// Pooled playback
	player := pool.players[pool.current]
	pool.current = (pool.current + 1) % len(pool.players)

	dx := x - listenerX
	dy := y - listenerY
	dist := math.Sqrt(dx*dx + dy*dy)

	volume := 1.0 - (dist / maxDistance)
	if volume < 0 {
		volume = 0
	}

	player.Rewind()
	player.SetVolume(volume * m.masterVolume * m.sfxVolume)
	player.Play()
}
