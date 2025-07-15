package handlers

import (
	"bytes"
	"crypto/tls"
	"realTimeEditor/pkg/constants"

	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"path/filepath"
)

// func LoadEnvironment() (*MailerVars, error) {
// 	env, err := constants.LoadEnv()
// 	if err != nil {
// 		return nil, fmt.Errorf("error loading environment variables: %s", err)
// 	}
// 	return &MailerVars{
// 		SMTP_HOST: env.SMTP_HOST,
// 		SMTP_PORT: env.SMTP_PORT,
// 		SMTP_USER: env.SMTP_USER,
// 		SMTP_PASS: env.SMTP_PASS,
// 	}, nil
// }

func ParseTemplate[T any](fileName string, data T) (string, error) {
	// Automatically look in templates directory
	templatePath := filepath.Join("templates", fileName+".html")

	tpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("Failed to parse template %s: %v", fileName, err)
		return "", fmt.Errorf("template parsing error: %w", err)
	}

	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		log.Printf("Failed to execute template %s: %v", fileName, err)
		return "", fmt.Errorf("template execution error: %w", err)
	}
	return body.String(), nil
}

func SendMail(to, templatePath, subject string, data any) error {
	// Load config
	config, err := constants.LoadEnv()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("config load failed: %w", err)
	}

	// Validate config
	if config.SMTP_HOST == "" || config.SMTP_PORT == "" || config.SMTP_USER == "" || config.SMTP_PASS == "" {
		return fmt.Errorf("incomplete SMTP configuration")
	}

	body, err := ParseTemplate(templatePath, data)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("template processing failed: %w", err)
	}

	// Message construction
	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject:"+subject+"\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"%s",
		config.SMTP_USER,
		to,
		body,
	)

	// TLS config
	tlsConfig := &tls.Config{
		ServerName: config.SMTP_HOST,
	}

	// Connect to SMTP server
	conn, err := tls.Dial("tcp", config.SMTP_HOST+":"+config.SMTP_PORT, tlsConfig)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, config.SMTP_HOST)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()
	auth := smtp.PlainAuth("", config.SMTP_USER, config.SMTP_PASS, config.SMTP_HOST)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	if err := client.Mail(config.SMTP_USER); err != nil {
		log.Println(err)
		return fmt.Errorf("sender set failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		log.Println(err)
		return fmt.Errorf("recipient set failed: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("data command failed: %w", err)
	}
	defer w.Close()

	if _, err := w.Write([]byte(msg)); err != nil {
		log.Println(err)
		return fmt.Errorf("message write failed: %w", err)
	}

	log.Printf("Email successfully sent to %s", to)
	return nil
}
