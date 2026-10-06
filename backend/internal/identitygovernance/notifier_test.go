// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package identitygovernance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	notifcommon "github.com/thunder-id/thunderid/internal/notification/common"
	"github.com/thunder-id/thunderid/internal/notificationtemplate"
	syscontext "github.com/thunder-id/thunderid/internal/system/context"
	"github.com/thunder-id/thunderid/internal/system/deployment"
	"github.com/thunder-id/thunderid/internal/system/log"
	"github.com/thunder-id/thunderid/internal/system/log/rollingfile"
	tidcommon "github.com/thunder-id/thunderid/pkg/thunderidengine/common"
	"github.com/thunder-id/thunderid/pkg/thunderidengine/providers"
	"github.com/thunder-id/thunderid/tests/mocks/notification/notificationmock"
)

type NotifierTestSuite struct {
	suite.Suite

	profiles  *entityStateProviderMock
	templates *notificationTemplateRendererMock
	senders   *notificationmock.NotificationSenderServiceInterfaceMock
	sender    *defaultEmailSenderSourceMock
	sent      chan notifcommon.EmailData
}

// testSenderID is the default email sender the suite configures.
const testSenderID = "sender-1"

func TestNotifierTestSuite(t *testing.T) {
	suite.Run(t, new(NotifierTestSuite))
}

func (s *NotifierTestSuite) SetupTest() {
	s.profiles = newEntityStateProviderMock(s.T())
	s.templates = newNotificationTemplateRendererMock(s.T())
	s.senders = notificationmock.NewNotificationSenderServiceInterfaceMock(s.T())
	s.sender = newDefaultEmailSenderSourceMock(s.T())
	s.sender.EXPECT().DefaultEmailSenderID(mock.Anything).Return(testSenderID).Maybe()
	s.sent = make(chan notifcommon.EmailData, 4)
}

func testUser() GovernedEntity {
	return GovernedEntity{ID: "user-1", Category: providers.EntityCategoryUser, Type: "member"}
}

func testLockEpisode() LockEpisode {
	return LockEpisode{Duration: 5 * time.Minute, UnlockAt: "2026-10-08T10:05:00Z", LockCount: 1}
}

func lockEmailConfig(email accountAccessLockEmail) accountAccessValue {
	return accountAccessValue{User: &accountAccessCategory{
		Notifications: &accountAccessNotifications{
			OnLock: &accountAccessLockNotification{Email: &email},
		},
	}}
}

// newNotifier builds a notifier over the suite's mocks, reading the given settings through the source.
func (s *NotifierTestSuite) newNotifier(settings accountAccessLockEmail) *lockEmailNotifier {
	return newLockEmailNotifier(newAccountAccessSource(lockEmailConfig(settings)),
		s.profiles, s.templates, s.senders, s.sender)
}

// notifierFor builds a notifier with newNotifier and closes it when the test ends.
func (s *NotifierTestSuite) notifierFor(settings accountAccessLockEmail) *lockEmailNotifier {
	notifier := s.newNotifier(settings)
	s.T().Cleanup(notifier.Close)
	return notifier
}

// expectProfile answers the profile read with the given string attributes.
func (s *NotifierTestSuite) expectProfile(entity GovernedEntity, attributes map[string]string) {
	raw, err := json.Marshal(attributes)
	s.Require().NoError(err)
	s.expectProfileJSON(entity, string(raw))
}

// expectProfileJSON answers the profile read with the given attributes document.
func (s *NotifierTestSuite) expectProfileJSON(entity GovernedEntity, attributes string) {
	s.profiles.EXPECT().GetEntityProfile(mock.Anything, entity.ID).Return(
		&providers.Entity{ID: entity.ID, Category: entity.Category, Attributes: json.RawMessage(attributes)}, nil)
}

// expectRender returns a rendered mail and captures the handle and template data.
func (s *NotifierTestSuite) expectRender(handle *string, data *map[string]string) {
	s.templates.EXPECT().Resolve(mock.Anything, notificationtemplate.ChannelTypeEmail, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ notificationtemplate.ChannelType, h string,
			in notificationtemplate.RenderInput) (*notificationtemplate.ResolvedContent, *tidcommon.ServiceError) {
			*handle, *data = h, in.Data
			return &notificationtemplate.ResolvedContent{Subject: "Sign-in was locked", Body: "body"}, nil
		})
}

// expectSend records what reached the default email sender.
func (s *NotifierTestSuite) expectSend(svcErr *tidcommon.ServiceError) {
	s.senders.EXPECT().SendEmail(mock.Anything, testSenderID, mock.Anything).
		RunAndReturn(func(_ context.Context, _ string, data notifcommon.EmailData) *tidcommon.ServiceError {
			s.sent <- data
			return svcErr
		})
}

// assertNothingSent checks the sender was never called.
func (s *NotifierTestSuite) assertNothingSent() {
	s.senders.AssertNotCalled(s.T(), "SendEmail", mock.Anything, mock.Anything, mock.Anything)
}

// awaitSend waits for the worker goroutine to send.
func (s *NotifierTestSuite) awaitSend() notifcommon.EmailData {
	select {
	case data := <-s.sent:
		return data
	case <-time.After(2 * time.Second):
		s.FailNow("timed out waiting for the lock notification to be sent")
		return notifcommon.EmailData{}
	}
}

// captureLogs runs fn with the logger writing to a file at the given level and returns what it wrote.
func (s *NotifierTestSuite) captureLogs(level string, fn func()) string {
	path := filepath.Join(s.T().TempDir(), "notifier.log")
	logger := log.GetLogger()
	s.Require().NoError(logger.Configure(log.OutputOptions{FileEnabled: true, File: rollingfile.Config{Path: path}}))
	s.Require().NoError(logger.SetLevel(level))
	defer func() {
		s.Require().NoError(logger.SetLevel("info"))
		s.Require().NoError(logger.Configure(log.OutputOptions{ConsoleEnabled: true}))
	}()
	fn()
	// #nosec G304 -- path is in the test's temp directory
	out, err := os.ReadFile(path)
	s.Require().NoError(err)
	return string(out)
}

// With no attribute configured, the source's default attribute is used.
func (s *NotifierTestSuite) TestSendsToTheDefaultAttribute() {
	entity := testUser()
	s.expectProfile(entity, map[string]string{"email": "member@example.com"})
	var handle string
	var data map[string]string
	s.expectRender(&handle, &data)
	s.expectSend(nil)

	notifier := s.notifierFor(accountAccessLockEmail{Enabled: new(true)})
	notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential, testLockEpisode())

	sent := s.awaitSend()
	s.Equal([]string{"member@example.com"}, sent.To)
	s.True(sent.IsHTML)
	s.Equal(lockTemplates[providers.AuthenticationMethodCredential].expiring, handle)
}

// deliveryContext is what one delivery step saw of its context, read while the step ran.
type deliveryContext struct {
	traceID, deploymentID string
	callerValue           any
	err                   error
	bounded               bool
}

// callerKey is a context key only the failed sign-in's context carries.
type callerKey struct{}

func observedContext(ctx context.Context) deliveryContext {
	_, bounded := ctx.Deadline()
	return deliveryContext{traceID: syscontext.GetTraceID(ctx), deploymentID: deployment.Resolve(ctx),
		callerValue: ctx.Value(callerKey{}), err: ctx.Err(), bounded: bounded}
}

// Every delivery step runs under the failed sign-in's trace ID and deployment, bounded by the delivery
// timeout, and under none of the caller's other values or its cancellation.
func (s *NotifierTestSuite) TestDeliveryKeepsOnlyTheTraceAndDeploymentOfTheRequest() {
	entity := testUser()
	seen := make(chan deliveryContext, 3)
	s.profiles.EXPECT().GetEntityProfile(mock.Anything, entity.ID).RunAndReturn(
		func(ctx context.Context, _ string) (*providers.Entity, error) {
			seen <- observedContext(ctx)
			return &providers.Entity{ID: entity.ID, Attributes: json.RawMessage(`{"email":"member@example.com"}`)}, nil
		})
	s.templates.EXPECT().
		Resolve(mock.Anything, notificationtemplate.ChannelTypeEmail,
			lockTemplates[providers.AuthenticationMethodCredential].expiring, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ notificationtemplate.ChannelType, _ string,
			_ notificationtemplate.RenderInput) (*notificationtemplate.ResolvedContent, *tidcommon.ServiceError) {
			seen <- observedContext(ctx)
			return &notificationtemplate.ResolvedContent{Subject: "Sign-in was locked", Body: "body"}, nil
		})
	delivered := make(chan struct{})
	s.senders.EXPECT().SendEmail(mock.Anything, testSenderID, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ notifcommon.EmailData) *tidcommon.ServiceError {
			seen <- observedContext(ctx)
			close(delivered)
			return nil
		})
	notifier := s.notifierFor(accountAccessLockEmail{Enabled: new(true)})
	reqCtx, cancel := context.WithCancel(context.WithValue(
		deployment.WithID(syscontext.WithTraceID(context.Background(), "trace-1"), "deployment-1"),
		callerKey{}, "caller only"))
	cancel()

	notifier.NotifyLockFormed(reqCtx, entity, providers.AuthenticationMethodCredential, testLockEpisode())

	select {
	case <-delivered:
	case <-time.After(2 * time.Second):
		s.FailNow("timed out waiting for the lock notification to be sent")
	}
	for _, step := range []string{"profile read", "render", "send"} {
		s.Equal(deliveryContext{traceID: "trace-1", deploymentID: "deployment-1", bounded: true}, <-seen, step)
	}
}

// The named attribute is used even when an "email" attribute also exists.
func (s *NotifierTestSuite) TestSendsToTheNamedAttribute() {
	entity := testUser()
	s.expectProfile(entity, map[string]string{
		"email":     "personal@example.com",
		"workEmail": "work@example.com",
	})
	var handle string
	var data map[string]string
	s.expectRender(&handle, &data)
	s.expectSend(nil)

	notifier := s.notifierFor(accountAccessLockEmail{
		Enabled: new(true), RecipientAttribute: new("workEmail"),
	})
	notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential, testLockEpisode())

	s.Equal([]string{"work@example.com"}, s.awaitSend().To)
}

// A dotted attribute name is a path into the profile's nested attributes.
func (s *NotifierTestSuite) TestADottedAttributeIsAPathIntoTheProfile() {
	entity := testUser()
	s.expectProfileJSON(entity, `{"email":"personal@example.com","contact":{"email":"member@example.com"}}`)
	var handle string
	var data map[string]string
	s.expectRender(&handle, &data)
	s.expectSend(nil)

	notifier := s.notifierFor(accountAccessLockEmail{
		Enabled: new(true), RecipientAttribute: new("contact.email"),
	})
	notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential, testLockEpisode())

	s.Equal([]string{"member@example.com"}, s.awaitSend().To)
}

// A recipient value that is not a usable address is skipped and recorded, never sent and never logged.
// A value that is absent or not text is a missing contact; text or a number that is not an address is
// an invalid one. A missing attribute never falls back to another.
func (s *NotifierTestSuite) TestAnUnusableRecipientIsNotSent() {
	for name, tc := range map[string]struct {
		attribute string
		profile   string
		outcome   noticeOutcome
	}{
		"text that is not an address": {"email", `{"email":"not an address"}`, outcomeInvalidContact},
		"a number":                    {"email", `{"email":5550142}`, outcomeInvalidContact},
		"an object":                   {"email", `{"email":{"value":"a@example.com"}}`, outcomeMissingContact},
		"a boolean":                   {"email", `{"email":true}`, outcomeMissingContact},
		"null":                        {"email", `{"email":null}`, outcomeMissingContact},
		"absent, with another address": {"workEmail", `{"email":"personal@example.com"}`,
			outcomeMissingContact},
		"a path through text": {"contact.email", `{"contact":"a@example.com"}`, outcomeMissingContact},
	} {
		s.Run(name, func() {
			s.SetupTest()
			entity := testUser()
			s.expectProfileJSON(entity, tc.profile)
			notifier := s.newNotifier(accountAccessLockEmail{Enabled: new(true), RecipientAttribute: &tc.attribute})

			logs := s.captureLogs("debug", func() {
				notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential,
					testLockEpisode())
				notifier.Close()
			})

			s.Contains(logs, "outcome="+string(tc.outcome))
			s.NotContains(logs, "not an address")
			s.NotContains(logs, "5550142")
			s.assertNothingSent()
		})
	}
}

// Settings are re-read at delivery, so a switch turned off after queuing is honored.
func (s *NotifierTestSuite) TestSwitchedOffBeforeDeliveryIsNotSent() {
	source := newAccountAccessSourceInterfaceMock(s.T())
	source.EXPECT().Current(mock.Anything).
		Return(lockEmailConfig(accountAccessLockEmail{Enabled: new(true)})).Once()
	source.EXPECT().Current(mock.Anything).Return(lockEmailConfig(accountAccessLockEmail{})).Once()
	notifier := &lockEmailNotifier{
		source:    source,
		profiles:  s.profiles,
		templates: s.templates,
		senders:   s.senders,
		sender:    s.sender,
		queue:     make(chan lockNotice, 1),
		timeout:   noticeTimeout,
		ctx:       context.Background(),
		logger:    log.GetLogger(),
	}

	notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
		testLockEpisode())
	notifier.deliver(<-notifier.queue)

	s.profiles.AssertNotCalled(s.T(), "GetEntityProfile", mock.Anything, mock.Anything)
	s.assertNothingSent()
}

// A notice after Close is dropped instead of sending on the closed queue.
func (s *NotifierTestSuite) TestANoticeAfterCloseIsDropped() {
	notifier := s.notifierFor(accountAccessLockEmail{Enabled: new(true)})
	notifier.Close()

	s.NotPanics(func() {
		notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
			testLockEpisode())
	})
	s.profiles.AssertNotCalled(s.T(), "GetEntityProfile", mock.Anything, mock.Anything)
	s.assertNothingSent()
}

// Close waits only its budget for sends that do not return. It then cancels them, records the notice
// still queued as dropped instead of delivering it, and drops a notice that comes after.
func (s *NotifierTestSuite) TestCloseIsBoundedWhenSendsBlock() {
	s.profiles.EXPECT().GetEntityProfile(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, id string) (*providers.Entity, error) {
			return &providers.Entity{ID: id, Attributes: json.RawMessage(`{"email":"member@example.com"}`)}, nil
		})
	s.templates.EXPECT().Resolve(mock.Anything, notificationtemplate.ChannelTypeEmail, mock.Anything, mock.Anything).
		Return(&notificationtemplate.ResolvedContent{Subject: "Sign-in was locked", Body: "body"}, nil)
	sending := make(chan struct{}, noticeWorkers+1)
	s.senders.EXPECT().SendEmail(mock.Anything, testSenderID, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ notifcommon.EmailData) *tidcommon.ServiceError {
			sending <- struct{}{}
			<-ctx.Done()
			return &tidcommon.InternalServerError
		})
	// Not closed on cleanup: if Close hangs, a second Close would wait on the first.
	notifier := s.newNotifier(accountAccessLockEmail{Enabled: new(true)})
	notifier.shutdownBudget = 100 * time.Millisecond

	// Every worker blocks in a send, and one more notice waits in the queue behind them.
	for i := range noticeWorkers + 1 {
		user := testUser()
		user.ID = fmt.Sprintf("user-%d", i)
		notifier.NotifyLockFormed(context.Background(), user, providers.AuthenticationMethodCredential,
			testLockEpisode())
	}
	for range noticeWorkers {
		select {
		case <-sending:
		case <-time.After(2 * time.Second):
			s.FailNow("timed out waiting for the workers to start sending")
		}
	}

	logs := s.captureLogs("debug", func() {
		closed := make(chan struct{})
		go func() {
			notifier.Close()
			close(closed)
		}()
		select {
		case <-closed:
		case <-time.After(5 * time.Second):
			s.FailNow("Close did not return after its shutdown budget")
		}
		notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
			testLockEpisode())
	})

	s.profiles.AssertNumberOfCalls(s.T(), "GetEntityProfile", noticeWorkers)
	s.senders.AssertNumberOfCalls(s.T(), "SendEmail", noticeWorkers)
	s.Equal(noticeWorkers, strings.Count(logs, "outcome="+string(outcomeSendFailed)), "the cancelled sends")
	s.Equal(1, strings.Count(logs, "outcome="+string(outcomeShutdown)), "the queued notice")
	s.Equal(1, strings.Count(logs, "outcome="+string(outcomeClosed)), "the notice after Close")
}

// When disabled, nothing is queued and no profile is read.
func (s *NotifierTestSuite) TestDisabledIsNotEvenQueued() {
	notifier := s.notifierFor(accountAccessLockEmail{})
	notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
		testLockEpisode())

	notifier.Close()
	s.profiles.AssertNotCalled(s.T(), "GetEntityProfile", mock.Anything, mock.Anything)
}

// A failed profile read skips the notice without retry, and the error is logged with the outcome.
func (s *NotifierTestSuite) TestProfileReadFailureSkipsDelivery() {
	s.profiles.EXPECT().GetEntityProfile(mock.Anything, "user-1").Return(nil, errors.New("store down")).Once()
	notifier := s.newNotifier(accountAccessLockEmail{Enabled: new(true)})

	logs := s.captureLogs("info", func() {
		notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
			testLockEpisode())
		notifier.Close()
	})

	s.Contains(logs, "outcome="+string(outcomeReadFailed))
	s.Contains(logs, "store down")
	s.assertNothingSent()
}

// A missing template sends nothing.
func (s *NotifierTestSuite) TestTemplateFailureSendsNothing() {
	entity := testUser()
	s.expectProfile(entity, map[string]string{"email": "member@example.com"})
	s.templates.EXPECT().Resolve(mock.Anything, notificationtemplate.ChannelTypeEmail,
		lockTemplates[providers.AuthenticationMethodCredential].expiring,
		mock.Anything).Return(nil, &tidcommon.ServiceError{Code: "NTPL-404"})

	notifier := s.notifierFor(accountAccessLockEmail{Enabled: new(true)})
	notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential, testLockEpisode())

	notifier.Close()
	s.assertNothingSent()
}

// A refused send is recorded, not retried.
func (s *NotifierTestSuite) TestSendFailureIsNotRetried() {
	entity := testUser()
	s.expectProfile(entity, map[string]string{"email": "member@example.com"})
	var handle string
	var data map[string]string
	s.expectRender(&handle, &data)
	s.expectSend(&tidcommon.InternalServerError)

	notifier := s.notifierFor(accountAccessLockEmail{Enabled: new(true)})
	notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential, testLockEpisode())

	s.awaitSend()
	notifier.Close()
	s.senders.AssertNumberOfCalls(s.T(), "SendEmail", 1)
}

// Without a default email sender nothing is rendered or sent.
func (s *NotifierTestSuite) TestNoDefaultSenderSendsNothing() {
	entity := testUser()
	s.expectProfile(entity, map[string]string{"email": "member@example.com"})
	sender := newDefaultEmailSenderSourceMock(s.T())
	sender.EXPECT().DefaultEmailSenderID(mock.Anything).Return("")

	notifier := newLockEmailNotifier(newAccountAccessSource(lockEmailConfig(
		accountAccessLockEmail{Enabled: new(true)})),
		s.profiles, s.templates, s.senders, sender)
	notifier.NotifyLockFormed(context.Background(), entity, providers.AuthenticationMethodCredential, testLockEpisode())

	notifier.Close()
	s.templates.AssertNotCalled(s.T(), "Resolve", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	s.assertNothingSent()
}

// A full queue drops the notice rather than blocking the caller.
func (s *NotifierTestSuite) TestFullQueueDropsTheNotice() {
	notifier := &lockEmailNotifier{
		source: newAccountAccessSource(lockEmailConfig(
			accountAccessLockEmail{Enabled: new(true)})),
		profiles:  s.profiles,
		templates: s.templates,
		senders:   s.senders,
		sender:    s.sender,
		// No workers and a queue of one, so the second notice is dropped.
		queue:   make(chan lockNotice, 1),
		timeout: noticeTimeout,
		logger:  log.GetLogger(),
	}

	notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
		testLockEpisode())
	notifier.NotifyLockFormed(context.Background(), testUser(), providers.AuthenticationMethodCredential,
		testLockEpisode())

	s.Len(notifier.queue, 1)
}
