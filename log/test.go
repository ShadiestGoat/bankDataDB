package log

import "testing"

type testingLogger struct {
	t *testing.T
	kvs []any
}

func (t *testingLogger) Debugw(str string, kv ...any)   {t.t.Log(str, t.kvs, kv)}
func (t *testingLogger) Infow(str string, kv ...any)    {t.t.Log(str, t.kvs, kv)}
func (t *testingLogger) Warnw(str string, kv ...any)    {t.t.Log(str, t.kvs, kv)}
func (t *testingLogger) Errorw(str string, kv ...any)   {t.t.Log(str, t.kvs, kv)}

func (t *testingLogger) Debugf(str string, args ...any) {t.t.Logf(str, args...)}
func (t *testingLogger) Infof(str string, args ...any)  {t.t.Logf(str, args...)}
func (t *testingLogger) Warnf(str string, args ...any)  {t.t.Logf(str, args...)}
func (t *testingLogger) Errorf(str string, args ...any) {t.t.Logf(str, args...)}

func (t *testingLogger) With(kv ...any) Logger {
	cur := make([]any, len(t.kvs))
	copy(t.kvs, cur)

	return &testingLogger{
		t:   &testing.T{},
		kvs: append(cur, kv...),
	}
}

func NewTestLogger(t *testing.T) Logger {
	return &testingLogger{t, []any{"test_name", t.Name()}}
}

func NewTestCtxLogger(t *testing.T) CtxLogger {
	return NewCtxLoggerComplex(NewTestLogger(t))
}
