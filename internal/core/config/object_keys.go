package config

import "fmt"

const (
	UploadKeyFmt      = "uploads/%s/%s/audio.wav"        // groupID, taskID
	DenoisedKeyFmt    = "results/%s/%s/denoised.wav"     // groupID, taskID
	TranscriptKeyFmt  = "results/%s/%s/transcript.srt"   // groupID, taskID
	DiarizationKeyFmt = "results/%s/%s/diarization.json" // groupID, taskID
	SummaryKeyFmt     = "results/%s/%s/summary.md"       // groupID, taskID
)

func UploadKey(groupID, taskID string) string {
	return fmt.Sprintf(UploadKeyFmt, groupID, taskID)
}

func DenoisedKey(groupID, taskID string) string {
	return fmt.Sprintf(DenoisedKeyFmt, groupID, taskID)
}

func TranscriptKey(groupID, taskID string) string {
	return fmt.Sprintf(DenoisedKeyFmt, groupID, taskID)
}
func DiarizationKey(groupID, taskID string) string {
	return fmt.Sprintf(DenoisedKeyFmt, groupID, taskID)
}
func SummaryKey(groupID, taskID string) string {
	return fmt.Sprintf(DenoisedKeyFmt, groupID, taskID)
}

