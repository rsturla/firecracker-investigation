package main

import (
	"context"
	sdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

var (
	// Path to the kernel image
	kernelImagePath = "/home/admin/Firecracker/vmlinux"
	// Path to the root file system image
	rootDrivePath = "/home/admin/Firecracker/rootfs.img"
)

func main() {
	ctx := context.Background()

	cfg := NewConfig()

	machine, err := sdk.NewMachine(ctx, *cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := machine.StopVMM(); err != nil {
			panic(err)
		}
	}()

	if err := machine.Start(ctx); err != nil {
		panic(err)
	}

	if err := machine.Wait(ctx); err != nil {
		panic(err)
	}
}

func NewConfig() *sdk.Config {
	cfg := sdk.Config{
		SocketPath: "/tmp/firecracker.sock",
		KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off init=/init",
		Drives: []models.Drive{
			{
				DriveID:      sdk.String("rootfs"),
				PathOnHost:   sdk.String(rootDrivePath),
				IsRootDevice: sdk.Bool(true),
				IsReadOnly:   sdk.Bool(false),
			},
		},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  sdk.Int64(1),
			MemSizeMib: sdk.Int64(128),
			Smt:        sdk.Bool(false),
		},
		NetworkInterfaces: sdk.NetworkInterfaces{
			{
				CNIConfiguration: &sdk.CNIConfiguration{
					NetworkName: "fcnet",
					IfName:      "eth0",
					ConfDir:     ".",
				},
			},
		},
		KernelImagePath: kernelImagePath,
		LogLevel:        "Debug",
	}

	return &cfg
}
