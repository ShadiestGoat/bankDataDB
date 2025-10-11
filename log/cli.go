package log

import "fmt"

type CLILogger struct {kvs []any}

func (t *CLILogger) Debugw(str string, kv ...any)   {} // noop
func (t *CLILogger) Infow(str string, kv ...any)    {fmt.Println(str, append(t.kvs, kv...))}
func (t *CLILogger) Warnw(str string, kv ...any)    {fmt.Println(str, append(t.kvs, kv...))}
func (t *CLILogger) Errorw(str string, kv ...any)   {fmt.Println(str, append(t.kvs, kv...))}

func (t *CLILogger) Debugf(str string, args ...any) {} // noop
func (t *CLILogger) Infof(str string, args ...any)  {fmt.Printf(str, args...)}
func (t *CLILogger) Warnf(str string, args ...any)  {fmt.Printf(str, args...)}
func (t *CLILogger) Errorf(str string, args ...any) {fmt.Printf(str, args...)}

func (t *CLILogger) With(kv ...any) Logger {
	cur := make([]any, len(t.kvs) + len(kv))
	copy(t.kvs, cur)

	return &CLILogger{kvs: append(cur, kv...)}
}

func NewCLICtxLogger() CtxLogger {
	return NewCtxLoggerComplex(&CLILogger{})
}
