package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap/zapcore"
)

const MAX_DISC_MSG_SIZE = 1900

// TODO: Kms

func encodeDiscMessage(prefix string, e zapcore.Entry, t []zapcore.Field) string {
	msg := &discEnc{
		Builder: strings.Builder{},
		strRepl:  strings.NewReplacer(
			`\`, `\\`,
			`"`, `\"`,
			"```", "\\`\\`\\`",
		),
	}

	msg.WriteString(prefix + strings.ReplaceAll(e.Message, "```", "\\`\\`\\`") + "\n```\n")
	msg.AddTime("Source Time", e.Time)

	for _, v := range t {
		v.AddTo(msg)
		if msg.Builder.Len() >= MAX_DISC_MSG_SIZE {
			break
		}
	}

	resMsg := msg.Builder.String()
	if len(resMsg) > MAX_DISC_MSG_SIZE {
		resMsg = strings.TrimSpace(resMsg[:MAX_DISC_MSG_SIZE]) + "\n... clipped\n"
	}

	return resMsg + "```"
}

type discordWebhook struct {
	Content string `json:"content"`
}

func prepareDiscMsg(prefix, webhookURL string, e zapcore.Entry, f []zapcore.Field) (*http.Request, error) {
	msg := encodeDiscMessage(prefix, e, f)

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(&discordWebhook{msg})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(`POST`, webhookURL, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func sendDiscMsg(req *http.Request) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Discord says 204 requests that don't wait for data.
	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("bad Discord HTTP Status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// fuck you, its the easiest way to do logging
type DiscordCore struct {
	hookURL string
	pfx string

	l *sync.RWMutex

	zapcore.Core
}

func (d *DiscordCore) sendDiscord(e zapcore.Entry, f []zapcore.Field) error {
	d.l.RLock()

	req, err := prepareDiscMsg(d.pfx, d.hookURL, e, f)
	if err != nil {
		d.l.RUnlock()
		return err
	}

	go func ()  {
		// TODO: Figure out how to handle errors here
		err := sendDiscMsg(req)
		if err != nil {
			fmt.Println("Discord ERR", err)
		}

		d.l.RUnlock()
	}()

	return nil
}

func (d *DiscordCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	ce = d.Core.Check(ent, ce)
	if ent.Level >= zapcore.ErrorLevel {
		ce = ce.AddCore(ent, d)
	}

	return ce
}

func (d *DiscordCore) Write(e zapcore.Entry, f []zapcore.Field) error {
	return d.sendDiscord(e, f)
}

func (d *DiscordCore) Sync() error {
	d.l.Lock()
	err := d.Core.Sync()
	d.l.Unlock()

	return err
}