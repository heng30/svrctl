package svrctl

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/Heng30/logger"
)

var (
	sigChan = make(chan os.Signal, 1)
	appName = path.Base(os.Args[0])
	pidDir  = string("/var/run/")
	pidPath = pidDir + appName + ".pid"
)

func waitStopSignal() {
	signal.Notify(sigChan, os.Interrupt)
	s := <-sigChan
	logger.Warnf("got signal %v", s)
	os.Exit(1)
}

func startService() {
	var err error

	if err = os.MkdirAll(pidDir, 0755); err != nil {
		logger.Warnf("make %s failed, error: %v", pidDir, err)
	}

	file, err := os.Create(pidPath)
	if err != nil {
		logger.Warnf("open %s failed, error: %v", pidPath, err)
        return
	}
	defer file.Close()

	pid := os.Getpid()
	_, err = file.WriteString(fmt.Sprintf("%d\n", pid))
	if err != nil {
		logger.Warnf("save pid in %s failed, error: %v", pidPath, err)
        return 
	}

	go waitStopSignal()
}

func runAsDaemon(chpidDir, closefd bool) bool {
	if ret, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0); err != 0 {
		logger.Warnln("syscall fork error")
		return false
	} else {
		switch ret {
		case 0:
			break
		default:
			os.Exit(0)
		}

	}

	if _, err := syscall.Setsid(); err != nil {
		logger.Warnf("syscall setsid failed: %v", err)
		return false
	}

	if chpidDir {
		os.Chdir("/")
		return false
	}

	if closefd {
		if f, err := os.Open("/dev/null"); err == nil {
			fd := int(f.Fd())
			syscall.Dup2(fd, int(os.Stdin.Fd()))
			syscall.Dup2(fd, int(os.Stdout.Fd()))
			syscall.Dup2(fd, int(os.Stderr.Fd()))
		} else {
			logger.Warnf("open /dev/null failed: %v", err)
            return false
		}
	}
	return true
}

func daemonRun() {
	runAsDaemon(true, true)
}

func stopService() {
	file, err := os.Open(pidPath)
	if err != nil {
		logger.Warnf("open %s failed: %v", pidPath, err)
		return
	}
	defer file.Close()

	var pid int = 0
	if n, err := fmt.Fscanf(file, "%d", &pid); n != 1 && err != nil {
		logger.Warnf("Get pid from %s failed: %v", pidPath, err)
		return
	}

	if err := syscall.Kill(pid, syscall.SIGINT); err != nil {
		logger.Warnf("send SIGINT to pid: %d failed: %v", pid, err)
	}
}

func getCtl() string {
	start := flag.Bool("start", false, "start the service")
	restart := flag.Bool("restart", false, "restart the service")
	daemon := flag.Bool("daemon", false, "service run as daemon")
	stop := flag.Bool("stop", false, "stop the service")

	flag.Parse()
	if *start {
		return "start"
	} else if *restart {
		return "restart"
	} else if *daemon {
		return "daemon"
	} else if *stop {
		return "stop"
	} else {
		return ""
	}
}

func Run() {
	if len(os.Args) < 1 {
		return
	}

	verb := getCtl()
	if len(verb) <= 0 {
		return
	}

	switch verb {
	case "start":
		logger.Traceln("Start...")
		startService()
		return

	case "daemon":
		logger.Traceln("Daemon...")
		daemonRun()
		startService()
		return

	case "stop":
		logger.Traceln("Stop...")
		stopService()
        os.Exit(0);
		return

	case "restart":
		logger.Traceln("Restart...")
		stopService()

		daemonRun()
		startService()
		return
	}
}
