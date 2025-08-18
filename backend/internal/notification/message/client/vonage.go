/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asgardeo/thunder/internal/notification/message/constants"
	"github.com/asgardeo/thunder/internal/notification/message/model"
	httpservice "github.com/asgardeo/thunder/internal/system/http"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	vonageURL                 = "https://api.nexmo.com/v1/messages"
	vonageLoggerComponentName = "VonageClient"
)

// VonageClient implements the MessageClientInterface for sending messages via Vonage API.
type VonageClient struct {
	name      string
	url       string
	apiKey    string
	apiSecret string
	senderID  string
}

// NewVonageClient creates a new instance of VonageClient.
func NewVonageClient(sender model.MessageNotificationSender) (MessageClientInterface, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, vonageLoggerComponentName))

	client := &VonageClient{}
	client.name = sender.Name
	client.url = vonageURL

	for _, prop := range sender.Properties {
		switch prop.Name {
		case constants.VonagePropKeyAPIKey:
			client.apiKey = prop.Value
		case constants.VonagePropKeyAPISecret:
			client.apiSecret = prop.Value
		case constants.VonagePropKeySenderID:
			client.senderID = prop.Value
		default:
			logger.Warn("Unknown property for Vonage client", log.String("property", prop.Name))
		}
	}

	return client, nil
}

// GetName returns the name of the Vonage client.
func (v *VonageClient) GetName() string {
	return v.name
}

// SendSMS sends an SMS using the Vonage API.
func (v *VonageClient) SendSMS(sms model.SMSData) error {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, vonageLoggerComponentName))
	logger.Debug("Sending SMS via Vonage", log.String("to", log.MaskString(sms.To)))

	// Format the phone number according to Vonage requirements
	formattedPhoneNumber := v.formatPhoneNumber(sms.To)

	payload := map[string]interface{}{
		"message_type": "text",
		"channel":      "sms",
		"from":         v.senderID,
		"to":           formattedPhoneNumber,
		"text":         sms.Body,
		"sms": map[string]interface{}{
			"encoding_type": "text",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, v.url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(v.apiKey, v.apiSecret)

	// Send the HTTP request
	client := httpservice.NewHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Failed to close response body", log.Error(closeErr))
		}
	}()

	logger.Debug("Received response from Vonage", log.Int("statusCode", resp.StatusCode))

	// Check the response status
	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error("Failed to send SMS", log.Int("statusCode", resp.StatusCode),
			log.String("response", string(bodyBytes)))
		return fmt.Errorf("failed to send SMS, status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// formatPhoneNumber formats a phone number to comply with Vonage E.164 requirements
// by removing any leading '+' or '00' from the number.
func (v *VonageClient) formatPhoneNumber(phoneNumber string) string {
	if len(phoneNumber) > 0 && phoneNumber[0] == '+' {
		return phoneNumber[1:]
	}
	if len(phoneNumber) > 1 && phoneNumber[0:2] == "00" {
		return phoneNumber[2:]
	}
	return phoneNumber
}
