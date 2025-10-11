package log

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugw(str string, kv ...any)
	Infow(str string, kv ...any)
	Warnw(str string, kv ...any)
	Errorw(str string, kv ...any)

	Debugf(str string, args ...any)
	Infof(str string, args ...any)
	Warnf(str string, args ...any)
	Errorf(str string, args ...any)

	With(kv ...any) Logger
}

// Returns cleanup, error
// If err != nil, cleanup will be nil
func Init(discordPrefix, discordWebhookURL string) (func() error, error) {
	if discordWebhookURL != "" && !strings.HasPrefix(discordWebhookURL, "https://discord.com/api/webhooks/") {
		return nil, fmt.Errorf("bad Discord URL")
	}

	cfg := zap.NewProductionConfig()
	if os.Getenv("ENV") == "DEV" {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	cfg.OutputPaths = []string{"stdout"}
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true

	opts := []zap.Option{}
	if discordWebhookURL != "" {
		opts = append(opts, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return &DiscordCore{
				hookURL: discordWebhookURL,
				pfx:     discordPrefix,
				l:       &sync.RWMutex{},
				Core:    c,
			}
		}))
	}

	z, err := cfg.Build(opts...)
	if err != nil {
		return nil, err
	}

	l := &log{
		base: z.Sugar(),
	}

	GlobalLogger = l

	return z.Sync, nil
}

type log struct {
	base *zap.SugaredLogger
}

func (l *log) Debugw(str string, kv ...any)   { l.base.Debugw(str, kv...) }
func (l *log) Infow(str string, kv ...any)    { l.base.Infow(str, kv...) }
func (l *log) Warnw(str string, kv ...any)    { l.base.Warnw(str, kv...) }
func (l *log) Errorw(str string, kv ...any)   { l.base.Errorw(str, kv...) }
func (l *log) Debugf(str string, args ...any) { l.base.Debugf(str, args...) }
func (l *log) Infof(str string, args ...any)  { l.base.Infof(str, args...) }
func (l *log) Warnf(str string, args ...any)  { l.base.Warnf(str, args...) }
func (l *log) Errorf(str string, args ...any) { l.base.Errorf(str, args...) }

func (l *log) With(kv ...any) Logger {
	return &log{
		base: l.base.With(kv...),
	}
}

var GlobalLogger *log
