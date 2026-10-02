package locales

import (
	"crypto/rand"
	"strings"
)

// ThrowawayAppName is a suggested Twitch application name: "throwaway-" plus
// 16 lowercase letters or digits.
func ThrowawayAppName() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "throwaway-app"
	}
	for i, c := range raw {
		raw[i] = alphabet[int(c)%len(alphabet)]
	}
	return "throwaway-" + string(raw)
}

// GameSearchInstructions fills {{NAME}} in the Game Search setup guide.
// An empty appName gets a new ThrowawayAppName.
func GameSearchInstructions(instructions, appName string) string {
	if appName == "" {
		appName = ThrowawayAppName()
	}
	return strings.ReplaceAll(instructions, "{{NAME}}", appName)
}
