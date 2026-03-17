package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tenlisboa/tempo/internal/config"
	"github.com/tenlisboa/tempo/internal/ipc"
	"github.com/tenlisboa/tempo/internal/scheduler"
	"github.com/tenlisboa/tempo/internal/store"
)

func Run(cfg *config.Config) error {
	st, err := store.NewSQLite(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer st.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := scheduler.NewRunner(st, cfg.ClaudeBin)
	sched, err := scheduler.New(ctx, runner)
	if err != nil {
		return fmt.Errorf("create scheduler: %w", err)
	}

	tasks, err := st.ListTasks()
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}
	sched.LoadAll(tasks)
	sched.Start()
	sched.StartWatchdog(st)
	defer sched.Stop()

	srv := ipc.NewServer(cfg.SocketPath)
	h := &handlers{st: st, sched: sched, startAt: time.Now()}
	h.register(srv)

	if err := srv.Listen(); err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer srv.Close()

	fmt.Fprintf(os.Stderr, "tempod %s listening on %s\n", config.Version, cfg.SocketPath)

	go srv.Serve()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Fprintln(os.Stderr, "shutting down")
	return nil
}
