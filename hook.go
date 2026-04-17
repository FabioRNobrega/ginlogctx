package ginlogctx

import "github.com/sirupsen/logrus"

type hook struct{}

// NewHook creates a Logrus hook that injects the currently bound
// request-scoped fields into log entries emitted on the active request
// goroutine.
func NewHook(_ Config) logrus.Hook {
	return hook{}
}

// Install adds the ginlogctx hook to the provided logger.
//
// If logger is nil, the standard Logrus logger is used.
func Install(logger *logrus.Logger, cfg Config) {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	logger.AddHook(NewHook(cfg))
}

func (hook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook) Fire(entry *logrus.Entry) error {
	fields := currentFields()
	if len(fields) == 0 {
		return nil
	}

	for key, value := range fields {
		if _, exists := entry.Data[key]; exists {
			continue
		}
		entry.Data[key] = value
	}

	return nil
}
