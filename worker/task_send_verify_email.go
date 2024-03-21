package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	
	db "feyin/digital-fintech-api/db/sqlc"
	"feyin/digital-fintech-api/util"
	
)

// TaskSendVerifyEmail is a constant representing the task type for sending email verification
const TaskSendVerifyEmail = "task:send_verify_email"

// PayloadSendVerifyEmail represents the payload for the TaskSendVerifyEmail task
type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

// DistributeTaskSendVerifyEmail enqueues the TaskSendVerifyEmail task
func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	// Marshal the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// Create a new task and enqueue it
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Log information about the enqueued task
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}

// ProcessTaskSendVerifyEmail processes the TaskSendVerifyEmail task
func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	// Unmarshal the payload from JSON
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	// Retrieve user information from the database
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Create a new email verification record in the database
	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	// Compose email content
	subject := "Welcome to Simple Bank"
	// TODO: replace this URL with an environment variable that points to a front-end page
	verifyUrl := fmt.Sprintf("http://localhost:8080/verify_email?email_id=%d&secret_code=%s",
		verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.Username, verifyUrl)
	to := []string{user.Email}

	// Send the email
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	// Log information about the processed task
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}









/*


ask Processor (RedisTaskProcessor):

It implements the TaskProcessor interface to process tasks using Asynq.
Initializes Asynq server with configured queues, error handler, and logger.
Start method starts the server with a handler for the TaskSendVerifyEmail task.
Task Distributor (RedisTaskDistributor):

Implements the TaskDistributor interface to distribute tasks using Asynq.
DistributeTaskSendVerifyEmail method enqueues the TaskSendVerifyEmail task with the specified payload.
Payload and Constants:

Defines the PayloadSendVerifyEmail struct for the payload of the TaskSendVerifyEmail task.
Defines constants like TaskSendVerifyEmail and queue names.
Task Processing (ProcessTaskSendVerifyEmail):

Processes the TaskSendVerifyEmail task by retrieving user information, creating a verification email, and sending it.
Implements error handling and logging for each step.


*/