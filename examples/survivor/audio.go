package main

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"

	"td/internal/assets"
)

const (
	sampleRate = 44100
)

type AudioPlayer struct {
	manager *assets.AudioManager
}

func NewAudioPlayer() *AudioPlayer {
	// No filesystem needed for procedural audio
	return &AudioPlayer{
		manager: assets.NewAudioManager(nil),
	}
}

func (ap *AudioPlayer) PlaySound(name string) {
	if ap == nil || ap.manager == nil {
		return
	}
	// Use PlayPooled for everything by default if it's a pool,
	// but the manager distinguishes pools vs singles.
	// In GenerateSounds I will register them as pools.
	ap.manager.PlayPooled(name)
}

func (ap *AudioPlayer) PlayBGM() {
	if ap.manager != nil {
		ap.manager.PlayMusic("bgm")
	}
}

func (ap *AudioPlayer) SetSFXVolume(vol float64) {
	if ap.manager != nil {
		ap.manager.SetSFXVolume(vol)
	}
}

func (ap *AudioPlayer) SetMusicVolume(vol float64) {
	if ap.manager != nil {
		ap.manager.SetMusicVolume(vol)
	}
}

func (ap *AudioPlayer) SFXVolume() float64 {
	if ap.manager != nil {
		return ap.manager.GetSFXVolume()
	}

	return 0
}

func (ap *AudioPlayer) MusicVolume() float64 {
	if ap.manager != nil {
		return ap.manager.GetMusicVolume()
	}

	return 0
}

// Generators

func (ap *AudioPlayer) GenerateSounds() {
	// Register sounds as Pools (8 concurrent)
	// We need to pass valid WAV data.
	// Since gen... returns WAV bytes, we pass "wav" as format.
	ap.manager.CreatePoolFromBytes("shoot", genShootSound(), 8, "wav")
	ap.manager.CreatePoolFromBytes("hit", genHitSound(), 8, "wav")
	ap.manager.CreatePoolFromBytes("levelup", genLevelUpSound(), 4, "wav")
	ap.manager.CreatePoolFromBytes("select", genSelectSound(), 4, "wav")

	// BGM
	// Load as Music (streaming/looping)
	ap.manager.LoadMusicFromBytes("bgm", genBGM(), "wav")
}

// Waveform Generators

func genShootSound() []byte {
	// Short noise burst + square wave decay
	seconds := 0.15

	return genWavHeaderAndData(seconds, func(t float64) float64 {
		// Frequency sweep down
		freq := 800.0 - t*4000.0

		val := math.Sin(2 * math.Pi * freq * t)
		if val > 0 {
			val = 1
		} else {
			val = -1
		} // Square

		// Noise
		noise := rand.Float64()*2 - 1

		// Mix
		mix := val*0.5 + noise*0.5

		// Envelope
		env := 1.0 - t/seconds

		return mix * env * 0.3
	})
}

func genHitSound() []byte {
	seconds := 0.1

	return genWavHeaderAndData(seconds, func(t float64) float64 {
		noise := rand.Float64()*2 - 1
		env := 1.0 - t/seconds

		return noise * env * 0.3
	})
}

func genSelectSound() []byte {
	seconds := 0.1

	return genWavHeaderAndData(seconds, func(t float64) float64 {
		freq := 440.0 + t*1000.0 // Chirp up
		val := math.Sin(2 * math.Pi * freq * t)

		return val * 0.3
	})
}

func genLevelUpSound() []byte {
	seconds := 1.0
	// Major arpeggio: C4, E4, G4, C5
	freqs := []float64{261.63, 329.63, 392.00, 523.25}

	return genWavHeaderAndData(seconds, func(t float64) float64 {
		idx := int(t * 8) // change note every 1/8th sec roughly
		if idx >= len(freqs) {
			return 0
		}

		freq := freqs[idx]

		val := math.Sin(2 * math.Pi * freq * t)
		// Add some harmonics (square-ish)
		val2 := math.Sin(2*math.Pi*freq*2*t) * 0.5

		return (val + val2) * 0.2
	})
}

func genBGM() []byte {
	// Simple 4-bar loop, ~120 BPM -> 2s per bar -> 8s total
	seconds := 6.4 // 3.2s loop
	bpm := 150.0
	beatDur := 60.0 / bpm

	return genWavHeaderAndData(seconds, func(t float64) float64 {
		beat := int(t / beatDur)
		localT := t - float64(beat)*beatDur

		// Bass: C - G - Am - F (Standard progression)
		// 0-3 beats each
		progression := []float64{65.41, 98.00, 110.00, 87.31} // C2, G2, A2, F2
		noteIdx := (beat / 4) % 4
		baseFreq := progression[noteIdx]

		// Bass wave (Sawtooth)
		bass := (localT * baseFreq) - math.Floor(localT*baseFreq)
		bass = bass*2 - 1

		// Simple melody (Arpeggio)
		arpFreq := baseFreq * 4 // 2 octaves up
		if beat%2 == 0 {
			arpFreq *= 1.5 // Fifth
		}

		melody := math.Sin(2 * math.Pi * arpFreq * t)

		// Drums (Noise burst on beat)
		drum := 0.0
		if localT < 0.1 {
			drum = (rand.Float64()*2 - 1) * (1.0 - localT/0.1)
		}

		return (bass*0.3 + melody*0.1 + drum*0.3) * 0.2
	})
}

func genWavHeaderAndData(seconds float64, generator func(float64) float64) []byte {
	numSamples := int(seconds * sampleRate)
	dataSize := numSamples * 4 // 16-bit * 2 channels
	fileSize := 36 + dataSize

	buf := new(bytes.Buffer)

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, int32(fileSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, int32(16)) // Chunk size
	binary.Write(buf, binary.LittleEndian, int16(1))  // PCM
	binary.Write(buf, binary.LittleEndian, int16(2))  // Channels (Stereo)
	binary.Write(buf, binary.LittleEndian, int32(sampleRate))
	binary.Write(buf, binary.LittleEndian, int32(sampleRate*4)) // Byte rate
	binary.Write(buf, binary.LittleEndian, int16(4))            // Block align
	binary.Write(buf, binary.LittleEndian, int16(16))           // Bits per sample

	// data chunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, int32(dataSize))

	// Data
	for i := range numSamples {
		t := float64(i) / float64(sampleRate)
		val := generator(t)
		// Clamp
		if val > 1 {
			val = 1
		}

		if val < -1 {
			val = -1
		}

		intVal := int16(val * 32767)
		// Write twice for Stereo (L + R)
		binary.Write(buf, binary.LittleEndian, intVal)
		binary.Write(buf, binary.LittleEndian, intVal)
	}

	return buf.Bytes()
}
