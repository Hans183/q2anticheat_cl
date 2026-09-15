package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/user/anticheat_cl/database"
)

// PushManager handles VAPID key generation/storage and Web Push message delivery
type PushManager struct {
	db         *database.DB
	publicKey  string
	privateKey string
}

// NewPushManager initializes PushManager, generating or loading VAPID keys
func NewPushManager(db *database.DB) (*PushManager, error) {
	pm := &PushManager{db: db}
	if err := pm.initKeys(); err != nil {
		return nil, fmt.Errorf("init VAPID keys: %w", err)
	}
	return pm, nil
}

func (pm *PushManager) initKeys() error {
	// 1. Check environment variables first (allows manual override)
	envPub := strings.TrimSpace(os.Getenv("VAPID_PUBLIC_KEY"))
	envPriv := strings.TrimSpace(os.Getenv("VAPID_PRIVATE_KEY"))
	if envPub != "" && envPriv != "" {
		pm.publicKey = envPub
		pm.privateKey = envPriv
		log.Printf("[PUSH] VAPID keys loaded from environment variables")
		return nil
	}

	// 2. Check database for previously persisted keys
	pub, errPub := pm.db.GetConfigValue("vapid_public_key")
	priv, errPriv := pm.db.GetConfigValue("vapid_private_key")

	if errPub == nil && errPriv == nil && pub != "" && priv != "" {
		pm.publicKey = pub
		pm.privateKey = priv
		log.Printf("[PUSH] VAPID keys loaded from database")
		return nil
	}

	// 3. Generate new VAPID key pair
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return fmt.Errorf("generate VAPID keys: %w", err)
	}

	if err := pm.db.SetConfigValue("vapid_public_key", publicKey); err != nil {
		return fmt.Errorf("save public VAPID key: %w", err)
	}
	if err := pm.db.SetConfigValue("vapid_private_key", privateKey); err != nil {
		return fmt.Errorf("save private VAPID key: %w", err)
	}

	pm.publicKey = publicKey
	pm.privateKey = privateKey
	log.Printf("[PUSH] Generated and saved new VAPID key pair")
	return nil
}

// PublicKey returns the VAPID public key
func (pm *PushManager) PublicKey() string {
	return pm.publicKey
}

// PushNotificationPayload represents the JSON body sent to the Service Worker
type PushNotificationPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
	Badge string `json:"badge,omitempty"`
	URL   string `json:"url,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

// PushSendResult contains detailed diagnostic metrics for push dispatches
type PushSendResult struct {
	TotalSubscriptions int      `json:"total_subscriptions"`
	SuccessCount       int      `json:"success_count"`
	FailedCount        int      `json:"failed_count"`
	Errors             []string `json:"errors,omitempty"`
}

func (pm *PushManager) sendToSingle(data []byte, subscriberEmail string, subRecord database.PushSubscriptionRecord) error {
	sub := &webpush.Subscription{
		Endpoint: subRecord.Endpoint,
		Keys: webpush.Keys{
			P256dh: subRecord.P256dh,
			Auth:   subRecord.Auth,
		},
	}

	resp, err := webpush.SendNotification(data, sub, &webpush.Options{
		Subscriber:      subscriberEmail,
		VAPIDPublicKey:  pm.publicKey,
		VAPIDPrivateKey: pm.privateKey,
		TTL:             86400,
	})

	if resp != nil {
		defer resp.Body.Close()
		io.Copy(io.Discard, resp.Body)
	}

	if resp != nil {
		statusCode := resp.StatusCode
		if statusCode == http.StatusGone || statusCode == http.StatusNotFound || statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden || statusCode == http.StatusBadRequest {
			log.Printf("[PUSH] Subscription invalid/expired (status %d), removing endpoint: %s", statusCode, subRecord.Endpoint)
			_ = pm.db.DeletePushSubscription(subRecord.Endpoint)
			return fmt.Errorf("subscription expired or invalid (status %d)", statusCode)
		}
		if statusCode != http.StatusCreated && statusCode != http.StatusOK {
			log.Printf("[PUSH] Unexpected status %d for subscription ID %d (%s)", statusCode, subRecord.ID, subRecord.Endpoint)
			return fmt.Errorf("push service returned status %d", statusCode)
		}
		log.Printf("[PUSH] Notification sent OK (status %d) to subscription ID %d", statusCode, subRecord.ID)
		return nil
	}

	if err != nil {
		log.Printf("[PUSH] Error sending notification to subscription ID %d: %v", subRecord.ID, err)
		return err
	}

	return nil
}

// SendToAllSync sends Web Push notifications synchronously and returns detailed diagnostic results
func (pm *PushManager) SendToAllSync(payload PushNotificationPayload) PushSendResult {
	res := PushSendResult{}
	subs, err := pm.db.GetPushSubscriptions()
	if err != nil {
		log.Printf("[PUSH] Error retrieving push subscriptions: %v", err)
		res.Errors = append(res.Errors, fmt.Sprintf("db error: %v", err))
		return res
	}

	res.TotalSubscriptions = len(subs)
	if len(subs) == 0 {
		log.Printf("[PUSH] SendToAllSync called but 0 active subscriptions found in DB")
		return res
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[PUSH] Error marshaling push payload: %v", err)
		res.Errors = append(res.Errors, fmt.Sprintf("marshal error: %v", err))
		return res
	}

	subscriberEmail := os.Getenv("VAPID_SUBSCRIBER")
	if subscriberEmail == "" {
		subscriberEmail = "mailto:admin@q2anticheat.com"
	}

	for _, s := range subs {
		if err := pm.sendToSingle(data, subscriberEmail, s); err != nil {
			res.FailedCount++
			res.Errors = append(res.Errors, fmt.Sprintf("sub %d: %v", s.ID, err))
		} else {
			res.SuccessCount++
		}
	}

	return res
}

// SendToAll sends a Web Push notification asynchronously to all registered subscriptions
func (pm *PushManager) SendToAll(payload PushNotificationPayload) {
	go pm.SendToAllSync(payload)
}
