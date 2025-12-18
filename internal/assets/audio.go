package assets

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
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

// AudioManager handles loading and playing sounds and music.
type AudioManager struct {
	context *audio.Context
	sounds  map[string]*audio.Player
	music   map[string]*audio.Player
	fs      fs.FS
}

// NewAudioManager creates an audio manager.
func NewAudioManager(filesystem fs.FS) *AudioManager {
	return &AudioManager{
		context: audio.NewContext(DefaultSampleRate),
		sounds:  make(map[string]*audio.Player),
		music:   make(map[string]*audio.Player),
		fs:      filesystem,
	}
}

// LoadSound loads a sound effect (short audio clip).
func (m *AudioManager) LoadSound(name, path string) error {
	data, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read audio %s: %w", path, err)
	}

	stream, err := m.decodeAudio(path, data)
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

// LoadMusic loads a music track (long audio, often streamed).
func (m *AudioManager) LoadMusic(name, path string) error {
	data, err := fs.ReadFile(m.fs, path)
	if err != nil {
		return fmt.Errorf("failed to read music %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))

	var player *audio.Player

	switch ext {
	case ".wav":
		stream, err := wav.DecodeWithSampleRate(DefaultSampleRate, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decode WAV: %w", err)
		}
		loop := audio.NewInfiniteLoop(stream, stream.Length())
		player, err = m.context.NewPlayer(loop)
		if err != nil {
			return fmt.Errorf("failed to create music player: %w", err)
		}

	case ".mp3":
		stream, err := mp3.DecodeWithSampleRate(DefaultSampleRate, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decode MP3: %w", err)
		}
		loop := audio.NewInfiniteLoop(stream, stream.Length())
		player, err = m.context.NewPlayer(loop)
		if err != nil {
			return fmt.Errorf("failed to create music player: %w", err)
		}

	case ".ogg":
		stream, err := vorbis.DecodeWithSampleRate(DefaultSampleRate, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decode OGG: %w", err)
		}
		loop := audio.NewInfiniteLoop(stream, stream.Length())
		player, err = m.context.NewPlayer(loop)
		if err != nil {
			return fmt.Errorf("failed to create music player: %w", err)
		}

	default:
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	m.music[name] = player
	return nil
}

// decodeAudio decodes audio based on file extension.
func (m *AudioManager) decodeAudio(path string, data []byte) (io.Reader, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".wav":
		stream, err := wav.DecodeWithSampleRate(DefaultSampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode WAV: %w", err)
		}
		return stream, nil

	case ".mp3":
		stream, err := mp3.DecodeWithSampleRate(DefaultSampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode MP3: %w", err)
		}
		return stream, nil

	case ".ogg":
		stream, err := vorbis.DecodeWithSampleRate(DefaultSampleRate, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode OGG: %w", err)
		}
		return stream, nil

	default:
		return nil, fmt.Errorf("unsupported audio format: %s", ext)
	}
}

// PlaySound plays a sound effect once.
func (m *AudioManager) PlaySound(name string) {
	if player, ok := m.sounds[name]; ok {
		player.Rewind()
		player.Play()
	}
}

// PlaySoundWithVolume plays a sound with a specific volume (0.0 to 1.0).
func (m *AudioManager) PlaySoundWithVolume(name string, volume float64) {
	if player, ok := m.sounds[name]; ok {
		player.Rewind()
		player.SetVolume(volume)
		player.Play()
	}
}

// PlayMusic starts playing a music track.
func (m *AudioManager) PlayMusic(name string) {
	if player, ok := m.music[name]; ok {
		player.Rewind()
		player.Play()
	}
}

// StopMusic stops the current music.
func (m *AudioManager) StopMusic(name string) {
	if player, ok := m.music[name]; ok {
		player.Pause()
	}
}

// SetMusicVolume sets the volume for a music track.
func (m *AudioManager) SetMusicVolume(name string, volume float64) {
	if player, ok := m.music[name]; ok {
		player.SetVolume(volume)
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
		volumeStep := (targetVolume - startVolume) / float64(steps)

		for i := 0; i < steps; i++ {
			time.Sleep(stepDuration)
			player.SetVolume(startVolume + volumeStep*float64(i+1))
		}
		player.SetVolume(targetVolume)
	}()
}

// Context returns the audio context.
func (m *AudioManager) Context() *audio.Context {
	return m.context
}
