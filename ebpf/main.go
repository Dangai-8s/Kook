package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"go.uber.org/zap"

	connections "eTrace/connections"
	"eTrace/settings"
)

var objs = bpfObjects{}

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS -no-global-types -target $TARGET bpf keploy_ebpf.c -- -I../headers -I../headers/$TARGET

func getlogger() *zap.Logger {
	// logger init
	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := logCfg.Build()
	if err != nil {
		log.Panic("failed to start the logger for the CLI")
		return nil
	}
	return logger
}

func main() {

	// start a profiler
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	println("Ebpf Loader PID:", os.Getpid())

	if err := settings.InitRealTimeOffset(); err != nil {
		log.Printf("Failed fixing BPF clock, timings will be offseted: %v", err)
	}

	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	// objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading objects: %+v", err)
		var ve *ebpf.VerifierError
		if errors.As(err, &ve) {
			log.Fatal("verifier log: %s", strings.Join(ve.Log, "\n"))
		}
	}

	defer objs.Close()

	logger := getlogger()

	connectionFactory := connections.NewFactory(time.Minute, logger)
	go func() {
		for {
			connectionFactory.HandleReadyConnections()
			time.Sleep(1 * time.Second)
		}
	}()

	// ------------ For Ingress using Kprobes --------------

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program.
	ac, err := link.Kprobe("sys_accept", objs.SyscallProbeEntryAccept, nil)
	if err != nil {
		log.Fatalf("opening accept kprobe: %s", err)
	}
	defer ac.Close()

	// Open a Kprobe at the exit point of the kernel function and attach the
	// pre-compiled program.
	ac_, err := link.Kretprobe("sys_accept", objs.SyscallProbeRetAccept, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening accept kretprobe: %s", err)
	}
	defer ac_.Close()

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program.
	ac4, err := link.Kprobe("sys_accept4", objs.SyscallProbeEntryAccept4, nil)
	if err != nil {
		log.Fatalf("opening accept4 kprobe: %s", err)
	}
	defer ac4.Close()

	// Open a Kprobe at the exit point of the kernel function and attach the
	// pre-compiled program.
	ac4_, err := link.Kretprobe("sys_accept4", objs.SyscallProbeRetAccept4, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening accept4 kretprobe: %s", err)
	}
	defer ac4_.Close()

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program.
	rd, err := link.Kprobe("sys_read", objs.SyscallProbeEntryRead, nil)
	if err != nil {
		log.Fatalf("opening read kprobe: %s", err)
	}
	defer rd.Close()

	// Open a Kprobe at the exit point of the kernel function and attach the
	// pre-compiled program.
	rd_, err := link.Kretprobe("sys_read", objs.SyscallProbeRetRead, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening read kretprobe: %s", err)
	}
	defer rd_.Close()

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program.
	wt, err := link.Kprobe("sys_write", objs.SyscallProbeEntryWrite, nil)
	if err != nil {
		log.Fatalf("opening write kprobe: %s", err)
	}
	defer wt.Close()

	// Open a Kprobe at the exit point of the kernel function and attach the
	// pre-compiled program.
	wt_, err := link.Kretprobe("sys_write", objs.SyscallProbeRetWrite, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening write kretprobe: %s", err)
	}
	defer wt_.Close()

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program for writev. (javascript/typescript specific)
	wtv, err := link.Kprobe("sys_writev", objs.SyscallProbeEntryWritev, nil)
	if err != nil {
		log.Fatalf("opening writev kprobe: %s", err)
	}
	defer wtv.Close()

	// Open a Kprobe at the exit point of the kernel function and attach the
	// pre-compiled program for writev.
	wtv_, err := link.Kretprobe("sys_writev", objs.SyscallProbeRetWritev, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening writev kretprobe: %s", err)
	}
	defer wtv_.Close()

	//python specific sys calls (sys_sendto, sys_recvfrom)

	//Open a kprobe at the entry of sendto syscall
	snd, err := link.Kprobe("sys_sendto", objs.SyscallProbeEntrySendto, nil)
	if err != nil {
		log.Fatalf("opening sendto kprobe: %s", err)
	}
	defer snd.Close()

	//Opening a kretprobe at the exit of sendto syscall
	sndr, err := link.Kretprobe("sys_sendto", objs.SyscallProbeRetSendto, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening sendto kretprobe: %s", err)
	}
	defer sndr.Close()

	//Attaching a kprobe at the entry of recvfrom syscall
	rcv, err := link.Kprobe("sys_recvfrom", objs.SyscallProbeEntryRecvfrom, nil)
	if err != nil {
		log.Fatalf("opening recvfrom kprobe: %s", err)
	}
	defer rcv.Close()

	//Attaching a kretprobe at the exit of recvfrom syscall
	rcvr, err := link.Kretprobe("sys_recvfrom", objs.SyscallProbeRetRecvfrom, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening recvfrom kretprobe: %s", err)
	}
	defer rcvr.Close()

	// Open a Kprobe at the entry point of the kernel function and attach the
	// pre-compiled program.
	cl, err := link.Kprobe("sys_close", objs.SyscallProbeEntryClose, nil)
	if err != nil {
		log.Fatalf("opening write kprobe: %s", err)
	}
	defer cl.Close()

	// Open a Kprobe at the exit point of the kernel function and attach the
	// pre-compiled program.
	cl_, err := link.Kretprobe("sys_close", objs.SyscallProbeRetClose, &link.KprobeOptions{RetprobeMaxActive: 2048})
	if err != nil {
		log.Fatalf("opening write kretprobe: %s", err)
	}
	defer cl_.Close()

	var appPidFlag int
	flag.IntVar(&appPidFlag, "pid", 0, "Application PID")
	flag.Parse()
	appPid := uint32(appPidFlag)

	key := 0
	//send application pid to kernel to filter.
	log.Printf("Application pid sending to kernel:%v", appPid)
	err = objs.AppPidMap.Update(uint32(key), &appPid, ebpf.UpdateAny)
	if err != nil {
		log.Fatalf("failed to send application pid to kernel %v", err)
	}

	LaunchPerfBufferConsumers(objs, connectionFactory, stopper, logger)

	log.Printf("Probes added to the kernel.\n")

	<-stopper
	log.Println("Received signal, exiting program..")

	// closing all readers.
	for _, reader := range PerfEventReaders {
		if err := reader.Close(); err != nil {
			log.Fatalf("closing perf reader: %s", err)
		}
	}
	for _, reader := range RingEventReaders {
		if err := reader.Close(); err != nil {
			log.Fatalf("closing ringbuf reader: %s", err)
		}
	}

}
