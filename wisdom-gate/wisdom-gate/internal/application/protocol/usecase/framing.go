package usecase

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Message struct {
	Command string
	Body    string
}

func (m *Message) String() string {
	if m.Body == "" {
		return m.Command
	}
	return fmt.Sprintf("%s %d |%s", m.Command, len(m.Body), m.Body)
}

func WriteMessage(w io.Writer, msg *Message) error {
	_, err := fmt.Fprintf(w, "%s\n", msg.String())
	return err
}

func ReadMessage(r *bufio.Reader) (*Message, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)

	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 1 {
		// Command only
		return &Message{Command: parts[0]}, nil
	}

	bodyPart := parts[1]
	bodyParts := strings.SplitN(bodyPart, " |", 2)
	if len(bodyParts) != 2 {
		return nil, fmt.Errorf("invalid message format")
	}

	length, err := strconv.Atoi(bodyParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid body length: %w", err)
	}

	body := bodyParts[1]
	if len(body) != length {
		return nil, fmt.Errorf("body length mismatch: expected %d, got %d", length, len(body))
	}

	return &Message{
		Command: parts[0],
		Body:    body,
	}, nil
}
