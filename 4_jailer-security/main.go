package main

import (
	"context"
	sdk "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"io"
	"os"
)

var (
	// Path to the kernel image
	kernelImagePath = "/home/admin/Firecracker/vmlinux"
	// Path to the root file system image
	rootDrivePath = "/home/admin/Firecracker/rootfs.img"
	// Path to the sandbox directory
	sandboxPath = "/home/admin/Firecracker/sandbox"
	// The UID and GID of the sandbox user
	sandboxUID = 1000
	sandboxGID = 1000
)

func main() {
	ctx := context.Background()

	sandbox, err := NewSandbox(sandboxPath)
	if err != nil {
		panic(err)
	}
	defer sandbox.Cleanup()

	cfg := sandbox.NewConfig()

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

func (s *Sandbox) NewConfig() *sdk.Config {
	cfg := sdk.Config{
		SocketPath: "firecracker.sock",
		KernelArgs: "console=ttyS0 reboot=k panic=1 pci=off init=/init",
		Drives: []models.Drive{
			{
				DriveID:      sdk.String("rootfs"),
				PathOnHost:   sdk.String(s.RootDrivePath),
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
				},
			},
		},
		//NetworkInterfaces: sdk.NetworkInterfaces{
		//	{
		//		StaticConfiguration: &sdk.StaticNetworkConfiguration{
		//			HostDevName: "ftap0",
		//			IPConfiguration: &sdk.IPConfiguration{
		//				IPAddr: net.IPNet{
		//					IP: net.ParseIP("172.16.0.2"),
		//					// 255.255.255.0
		//					Mask: net.IPv4Mask(255, 255, 255, 0),
		//				},
		//				IfName:  "eth0",
		//				Gateway: net.ParseIP("172.16.0.1"),
		//			},
		//		},
		//	},
		//},
		KernelImagePath: s.KernelImagePath,
		LogLevel:        "Debug",
		JailerCfg: &sdk.JailerConfig{
			ID:             "fc-sandbox",
			UID:            sdk.Int(sandboxUID),
			GID:            sdk.Int(sandboxGID),
			NumaNode:       sdk.Int(0),
			CgroupVersion:  "2",
			ChrootBaseDir:  sandboxPath,
			ChrootStrategy: sdk.NewNaiveChrootStrategy(s.KernelImagePath),
			ExecFile:       "/usr/local/bin/firecracker",
		},
	}

	return &cfg
}

type Sandbox struct {
	KernelImagePath string
	RootDrivePath   string
}

func NewSandbox(path string) (*Sandbox, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	if err := copyFile(kernelImagePath, path+"/vmlinux", sandboxUID, sandboxGID); err != nil {
		return nil, err
	}

	if err := copyFile(rootDrivePath, path+"/rootfs.img", sandboxUID, sandboxGID); err != nil {
		return nil, err
	}

	return &Sandbox{
		KernelImagePath: path + "/vmlinux",
		RootDrivePath:   path + "/rootfs.img",
	}, nil
}

func (s *Sandbox) Cleanup() error {
	return os.RemoveAll(sandboxPath)
}

func copyFile(src, dst string, uid, gid int) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFd.Close()

	dstFd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFd.Close()

	if _, err = io.Copy(dstFd, srcFd); err != nil {
		return err
	}

	if err := os.Chown(dst, uid, gid); err != nil {
		return err
	}

	return nil
}
