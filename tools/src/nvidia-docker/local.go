// Copyright (c) 2015-2016, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"

	"github.com/NVIDIA/nvidia-docker/tools/src/docker"
	"github.com/NVIDIA/nvidia-docker/tools/src/nvidia"
)

func GenerateLocalArgs(image string, vols []string) ([]string, error) {
	cv, err := nvidia.GetCUDAVersion()
	if err != nil {
		return nil, err
	}
	if err := cudaSupported(image, cv); err != nil {
		return nil, err
	}

	d, err := devicesArgs()
	if err != nil {
		return nil, err
	}
	v, err := volumesArgs(vols)
	if err != nil {
		return nil, err
	}
	return append(d, v...), nil
}

func devicesArgs() ([]string, error) {
	var args []string

	args = append(args, fmt.Sprintf("--device=%s", nvidia.DeviceCtl))
	args = append(args, fmt.Sprintf("--device=%s", nvidia.DeviceUVM))

	devs, err := nvidia.LookupDevices(nvidia.LookupMinimal)
	if err != nil {
		return nil, err
	}

	if len(GPU) == 0 {
		for i := range devs {
			args = append(args, fmt.Sprintf("--device=%s", devs[i].Path))
		}
	} else {
		devs, err := nvidia.FilterDevices(devs, GPU)
		if err != nil {
			return nil, err
		}
		for i := range devs {
			args = append(args, fmt.Sprintf("--device=%s", devs[i].Path))
		}
	}
	return args, nil
}

func volumesArgs(vols []string) ([]string, error) {
	args := make([]string, 0, len(vols))

	drv, err := nvidia.GetDriverVersion()
	if err != nil {
		return nil, err
	}
	for _, vol := range nvidia.Volumes {
		for _, v := range vols {
			if v == vol.Name {
				// Check if the volume exists locally otherwise fallback to using the plugin
				n := fmt.Sprintf("%s_%s", vol.Name, drv)
				if _, err := docker.VolumeInspect(n); err == nil {
					args = append(args, fmt.Sprintf("--volume=%s:%s:ro", n, vol.Mountpoint))
				} else {
					args = append(args, fmt.Sprintf("--volume-driver=%s", nvidia.DockerPlugin))
					args = append(args, fmt.Sprintf("--volume=%s:%s:ro", n, vol.Mountpoint))
				}
				break
			}
		}
	}
	return args, nil
}
