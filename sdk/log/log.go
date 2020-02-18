package log

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	loghook "github.com/ovh/cds/sdk/log/hook"
	"github.com/prometheus/common/log"

	"github.com/sirupsen/logrus"
)

// Conf contains log configuration
type Conf struct {
	Level                      string
	GraylogHost                string
	GraylogPort                string
	GraylogProtocol            string
	GraylogExtraKey            string
	GraylogExtraValue          string
	GraylogFieldCDSServiceType string
	GraylogFieldCDSServiceName string
	GraylogFieldCDSVersion     string
	GraylogFieldCDSOS          string
	GraylogFieldCDSArch        string
	Ctx                        context.Context
}

var (
	customLogger Logger
	hook         *loghook.Hook
)

// Logger defines the logs levels used
type Logger interface {
	Logf(fmt string, values ...interface{})
	Errorf(fmt string, values ...interface{})
	Fatalf(fmt string, values ...interface{})
}

// SetLogger replace logrus logger with custom one.
func SetLogger(l Logger) {
	customLogger = l
}

// Initialize init log level
func Initialize(conf *Conf) {
	switch conf.Level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.SetFormatter(&CDSFormatter{})

	if conf.GraylogHost != "" && conf.GraylogPort != "" {
		graylogcfg := &loghook.Config{
			Addr:      fmt.Sprintf("%s:%s", conf.GraylogHost, conf.GraylogPort),
			Protocol:  conf.GraylogProtocol,
			TLSConfig: &tls.Config{ServerName: conf.GraylogHost},
		}

		extra := map[string]interface{}{}
		if conf.GraylogExtraKey != "" && conf.GraylogExtraValue != "" {
			keys := strings.Split(conf.GraylogExtraKey, ",")
			values := strings.Split(conf.GraylogExtraValue, ",")
			if len(keys) != len(values) {
				log.Errorf("Error while initialize log: extraKey (len:%d) does not have same corresponding number of values on extraValue (len:%d)", len(keys), len(values))
			} else {
				for i := range keys {
					extra[keys[i]] = values[i]
				}
			}
		}

		if conf.GraylogFieldCDSServiceName != "" {
			extra["CDSName"] = conf.GraylogFieldCDSServiceName
		}
		if conf.GraylogFieldCDSServiceName != "" {
			extra["CDSService"] = conf.GraylogFieldCDSServiceType
		}
		if conf.GraylogFieldCDSVersion != "" {
			extra["CDSVersion"] = conf.GraylogFieldCDSVersion
		}
		if conf.GraylogFieldCDSOS != "" {
			extra["CDSOS"] = conf.GraylogFieldCDSOS
		}
		if conf.GraylogFieldCDSArch != "" {
			extra["CDSArch"] = conf.GraylogFieldCDSArch
		}

		// no need to check error here
		hostname, _ := os.Hostname()
		extra["CDSHostname"] = hostname

		var errhook error
		hook, errhook = loghook.NewHook(graylogcfg, extra)

		if errhook != nil {
			logrus.Errorf("Error while initialize graylog hook: %v", errhook)
		} else {
			logrus.AddHook(hook)
			logrus.SetOutput(ioutil.Discard)
		}
	}

	if conf.Ctx == nil {
		conf.Ctx = context.Background()
	}

	go func() {
		<-conf.Ctx.Done()
		logrus.Info(conf.Ctx, "Draining logs")
		if hook != nil {
			hook.Flush()
		}
	}()
}

// Debug prints debug log
func Debug(format string, values ...interface{}) {
	if logger != nil {
    logger.Logf("[DEBUG] "+format, values...)
    return
  }
			logrus.Debugf(format, values...)
}

// Info prints information log
func Info(ctx context.Context, format string, values ...interface{}) {
	if logger != nil {
		logger.Logf("[INFO] "+format, values...)
		return
  }
  logrus.WithFields(fields Fields).
	logrus.Infof(format, values...)
}

// InfoWithoutCtx prints information log
func InfoWithoutCtx(format string, values ...interface{}) {
	if logger != nil {
		logger.Logf("[INFO] "+format, values...)
		return
	}
	logrus.Infof(format, values...)
}

// Warning prints warnings for user
func Warning(ctx context.Context, format string, values ...interface{}) {
	if logger != nil {
    logger.Logf("[WARN] "+format, values...)
    return
	}
			logrus.Warnf(format, values...)
}

// Error prints error informations
func Error(ctx context.Context, format string, values ...interface{}) {
	if logger != nil {
		logger.Logf("[ERROR] "+format, values...)
return
  }
			logrus.Errorf(format, values...)
}

// Fatalf prints fatal informations, then os.Exit(1)
func Fatalf(format string, values ...interface{}) {
	if logger != nil {
		logger.Logf("[FATAL] "+format, values...)
		return
	}
	logrus.Fatalf(format, values...)
}
