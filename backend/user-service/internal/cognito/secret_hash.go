package cognito

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// SecretHash is required when the Cognito App Client has a client secret.
// AWS expects: Base64(HMAC_SHA256(clientSecret, username+clientId)).
func SecretHash(username, clientID, clientSecret string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(username + clientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
