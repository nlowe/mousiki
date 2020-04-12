package pandora

type AudioFormat string

const (
	AudioFormatAACPlus    AudioFormat = "aacplus"
	AudioFormatMP3        AudioFormat = "mp3"
	AudioFormatPremiumMP3 AudioFormat = "mp3-hifi"
)

func IsValidAudioFormat(f string) bool {
	switch AudioFormat(f) {
	case AudioFormatMP3:
		fallthrough
	case AudioFormatPremiumMP3:
		fallthrough
	case AudioFormatAACPlus:
		return true
	default:
		return false
	}
}
