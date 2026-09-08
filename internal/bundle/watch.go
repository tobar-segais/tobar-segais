package bundle

import (
	"context"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watch keeps the library in step with its directory until ctx is done.
//
// fsnotify is used where the platform supports it, and a poll loop where it
// does not -- which in practice means network filesystems and containers with
// certain volume drivers, exactly the places documentation archives tend to
// live. Either way the caller gets the same behaviour: archives added,
// replaced or removed are picked up without a restart.
func Watch(ctx context.Context, l *Library, poll time.Duration, log *slog.Logger) {
	if err := l.Reload(); err != nil {
		log.Error("initial scan failed", "dir", l.Dir(), "err", err)
	}

	w, err := fsnotify.NewWatcher()
	if err == nil {
		if err = w.Add(l.Dir()); err != nil {
			w.Close()
		}
	}
	if err != nil {
		log.Warn("filesystem notifications unavailable, polling instead",
			"dir", l.Dir(), "interval", poll, "err", err)
		pollLoop(ctx, l, poll, log)
		return
	}
	defer w.Close()
	log.Info("watching content directory", "dir", l.Dir())

	// Writes arrive in bursts -- a large archive copied in generates many
	// events, and a rename generates several. Settle before reloading so a
	// file is not read while it is still being written.
	const settle = 300 * time.Millisecond
	var timer *time.Timer
	var fire <-chan time.Time

	// Poll slowly even when watching: it catches anything the platform
	// silently drops, and costs one readdir.
	backstop := time.NewTicker(poll)
	defer backstop.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			// A metadata file is content too: writing one beside an
			// archive changes what that archive is.
			if !isArchive(ev.Name) && !isSidecar(ev.Name) {
				continue
			}
			if timer == nil {
				timer = time.NewTimer(settle)
				fire = timer.C
			} else {
				timer.Reset(settle)
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Warn("watch error", "err", err)
		case <-fire:
			timer, fire = nil, nil
			if err := l.Reload(); err != nil {
				log.Error("reload failed", "err", err)
			}
		case <-backstop.C:
			if err := l.Reload(); err != nil {
				log.Error("reload failed", "err", err)
			}
		}
	}
}

func pollLoop(ctx context.Context, l *Library, every time.Duration, log *slog.Logger) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := l.Reload(); err != nil {
				log.Error("reload failed", "err", err)
			}
		}
	}
}
