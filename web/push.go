package web

import (
	"encoding/json"
	"fmt"
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
	URL   string `json:"url,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

// SendToAll sends a Web Push notification asynchronously to all registered subscriptions
func (pm *PushManager) SendToAll(payload PushNotificationPayload) {
	subs, err := pm.db.GetPushSubscriptions()
	if err != nil {
		log.Printf("[PUSH] Error retrieving push subscriptions: %v", err)
		return
	}
	if len(subs) == 0 {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[PUSH] Error marshaling push payload: %v", err)
		return
	}

	subscriberEmail := os.Getenv("VAPID_SUBSCRIBER")
	if subscriberEmail == "" {
		subscriberEmail = "mailto:admin@q2anticheat.local"
	}

	for _, s := range subs {
		sub := &webpush.Subscription{
			Endpoint: s.Endpoint,
			Keys: webpush.Keys{
				P256dh: s.P256dh,
				Auth:   s.Auth,
			},
		}

		go func(subRecord database.PushSubscriptionRecord, s *webpush.Subscription) {
			resp, err := webpush.SendNotification(data, s, &webpush.Options{
				Subscriber:      subscriberEmail,
				VAPIDPublicKey:  pm.publicKey,
				VAPIDPrivateKey: pm.privateKey,
				TTL:             86400,
			})
			if err != nil {
				log.Printf("[PUSH] Error sending notification to subscription ID %d: %v", subRecord.ID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
				log.Printf("[PUSH] Subscription expired or revoked (%d), deleting: %s", resp.StatusCode, subRecord.Endpoint)
				_ = pm.db.DeletePushSubscription(subRecord.Endpoint)
			}
		}(s, sub)
	}
}
