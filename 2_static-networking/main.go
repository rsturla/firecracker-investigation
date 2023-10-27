package main

import (
	"context"
	sdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"net"
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
				StaticConfiguration: &sdk.StaticNetworkConfiguration{
					HostDevName: "ftap0",
					IPConfiguration: &sdk.IPConfiguration{
						IPAddr: net.IPNet{
							IP: net.ParseIP("172.16.0.2"),
							// 255.255.255.0
							Mask: net.IPv4Mask(255, 255, 255, 0),
						},
						IfName:  "eth0",
						Gateway: net.ParseIP("172.16.0.1"),
					},
				},
			},
		},
		KernelImagePath: kernelImagePath,
		LogLevel:        "Debug",
	}

	return &cfg
}
