package logstore

import (
	"context"
	"fmt"
	"forester/internal/config"
	"github.com/djherbis/times"
	"golang.org/x/exp/slog"
	"gopkg.in/mcuadros/go-syslog.v2"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"time"
)

type Directory struct {
	srv               *syslog.Server
	handlerCancelFunc context.CancelFunc
}

func Start(ctx context.Context) (*Directory, error) {
	dir := Directory{
		srv: syslog.NewServer(),
	}

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	dir.srv.SetFormat(syslog.Automatic)
	dir.srv.SetHandler(handler)
	err := dir.srv.ListenUDP(fmt.Sprintf("0.0.0.0:%d", config.Application.SyslogPort))
	if err != nil {
		return nil, fmt.Errorf("cannot listen on UDP port %d: %w", config.Application.SyslogPort, err)
	}
	err = dir.srv.ListenTCP(fmt.Sprintf("0.0.0.0:%d", config.Application.SyslogPort))
	if err != nil {
		return nil, fmt.Errorf("cannot listen on TCP port %d: %w", config.Application.SyslogPort, err)
	}
	err = dir.srv.Boot()
	if err != nil {
		return nil, fmt.Errorf("cannot start syslog server: %w", err)
	}

	err = os.MkdirAll(config.Logging.SyslogDir, 0770)
	if err != nil {
		return nil, fmt.Errorf("cannot create directory %s: %w", config.Logging.SyslogDir, err)
	}

	var handlerCtx context.Context
	handlerCtx, dir.handlerCancelFunc = context.WithCancel(ctx)
	if config.Logging.Syslog {
		go dir.fileHandler(handlerCtx, channel)
	} else {
		go dir.noopHandler(handlerCtx, channel)
	}

	return &dir, nil
}

func (d *Directory) Stop() {
	slog.Debug("stopping syslog listeners")
	err := d.srv.Kill()
	if err != nil {
		slog.Warn("cannot stop syslog server", "err", err)
	}
	d.srv.Wait()
	d.handlerCancelFunc()
}

type LogEntry struct {
	Path       string
	Size       int64
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func LogsForSystem(ctx context.Context, systemID int64) ([]LogEntry, error) {
	result := make([]LogEntry, 0)
	files, err := filepath.Glob(path.Join(config.Logging.SyslogDir, fmt.Sprintf("f-%d-*", systemID)))
	if err != nil {
		return nil, fmt.Errorf("error while listing log entries: %w", err)
	}

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("error while reading stat log entry %s: %w", file, err)
		}
		t, err := times.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("error while reading atime log entry %s: %w", file, err)
		}
		e := LogEntry{
			Path:       stat.Name(),
			Size:       stat.Size(),
			CreatedAt:  t.BirthTime(),
			ModifiedAt: stat.ModTime(),
		}
		result = append(result, e)
	}

	return result, nil
}

func closeFiles(files map[string]SyslogWriter) {
	for k, f := range files {
		if f.File == nil {
			delete(files, k)
			continue
		}

		if f.LastWrite.Before(time.Now().Add(time.Duration(-5) * time.Second)) {
			slog.Debug("closing syslog file after timeout", "file", f.File.Name())
			err := f.File.Close()
			if err != nil {
				slog.Error("cannot close", "file", f.File.Name(), "err", err.Error())
			}
			delete(files, k)
		}
	}
}

type SyslogWriter struct {
	File      *os.File
	LastWrite time.Time
}

func (d *Directory) recoverAndLog(ctx context.Context) {
	if rec := recover(); rec != nil {
		slog.WarnContext(ctx, "fatal error in syslog subsystem (panic)", "payload", rec, "stack", debug.Stack())
		d.srv.Kill()
	}
}

func (d *Directory) fileHandler(ctx context.Context, channel syslog.LogPartsChannel) {
	defer d.recoverAndLog(ctx)

	files := make(map[string]SyslogWriter)
	defer closeFiles(files)
	closeTicker := time.Tick(time.Second * 5)

	for {
		select {
		case logParts := <-channel:
			hpart, ok := logParts["hostname"]
			if !ok {
				slog.DebugContext(ctx, "log entry does not contain valid hostname, skipping")
				continue
			}
			hostname := hpart.(string)
			sw, ok := files[hostname]
			if !ok {
				var err error
				name := fmt.Sprintf("%s.log", hostname)
				slog.DebugContext(ctx, "opening syslog file", "file", name)
				// join and clean the path to prevent path traversal attacks
				fp := path.Join(config.Logging.SyslogDir, name)
				sw.File, err = os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
				if err != nil {
					slog.ErrorContext(ctx, "cannot open file for appending", "file", fp, "err", err.Error())
				}
				files[hostname] = sw
				slog.DebugContext(ctx, "file map", "size", len(files))
			}

			sw.LastWrite = time.Now()
			tag, ok := logParts["tag"]
			if !ok {
				tag = "-"
			}
			timestamp := "00:00:00"
			ts, ok := logParts["timestamp"]
			if ok {
				t, _ := time.Parse(time.RFC3339, ts.(string))
				timestamp = fmt.Sprintf("%02d:%02d:%02d", t.Hour(), t.Minute(), t.Second())
			}
			str := fmt.Sprintf("%s %s t:%s\n", timestamp, logParts["content"], tag)
			if _, err := sw.File.WriteString(str); err != nil {
				slog.ErrorContext(ctx, "cannot append to file", "file", sw.File.Name(), "err", err.Error())
			}

			if config.Logging.Syslog {
				var attrs []slog.Attr
				for k, v := range logParts {
					if k != "content" {
						attrs = append(attrs, slog.Any(k, v))
					}
				}
				slog.DebugContext(ctx, fmt.Sprintf("%s", logParts["content"]), "syslog", attrs)
			}
		case <-closeTicker:
			closeFiles(files)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Directory) noopHandler(ctx context.Context, channel syslog.LogPartsChannel) {
	for {
		select {
		case _ = <-channel:
			// do nothing
		case <-ctx.Done():
			return
		}
	}
}
