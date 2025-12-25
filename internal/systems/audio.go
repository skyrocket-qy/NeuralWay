package systems

import (
	"td/internal/assets"
)

// AudioSystem wraps the AudioManager for ECS-style usage.
type AudioSystem struct {
	manager *assets.AudioManager
}

// NewAudioSystem creates an audio system.
func NewAudioSystem(manager *assets.AudioManager) *AudioSystem {
	return &AudioSystem{
		manager: manager,
	}
}

// Manager returns the underlying AudioManager.
func (s *AudioSystem) Manager() *assets.AudioManager {
	return s.manager
}

// PlaySound plays a sound effect.
func (s *AudioSystem) PlaySound(name string) {
	s.manager.PlaySound(name)
}

// PlaySoundWithVolume plays a sound with specified volume.
func (s *AudioSystem) PlaySoundWithVolume(name string, volume float64) {
	s.manager.PlaySoundWithVolume(name, volume)
}

// PlayMusic starts playing music.
func (s *AudioSystem) PlayMusic(name string) {
	s.manager.PlayMusic(name)
}

// StopMusic stops the music.
func (s *AudioSystem) StopMusic(name string) {
	s.manager.StopMusic(name)
}

// SetMusicVolume sets music volume.
func (s *AudioSystem) SetMusicVolume(name string, volume float64) {
	s.manager.SetMusicTrackVolume(name, volume)
}

// Update is a no-op but satisfies the System interface.
func (s *AudioSystem) Update(world any) {
	// Audio playback is handled by Ebitengine internally
}
