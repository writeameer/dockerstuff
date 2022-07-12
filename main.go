package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dapr/cli/pkg/print"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {

	print.PendingStatusEvent(os.Stdout, "Making the jump to hyperspace...")
	print.PendingStatusEvent(os.Stdout, "Starting up systems...")

	docker := NewDockerUtil()

	imageName := "postgres:14.2-alpine"
	running, _, _ := docker.IsRunning(imageName)

	if !running {
		docker.pull(imageName)
		docker.run(imageName, []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
		})
		print.InfoStatusEvent(os.Stdout, "All systems are online.")
	} else {
		//fmt.Printf("Container: %s, is already running, skipping!\n", imageName)
		print.InfoStatusEvent(os.Stdout, "All systems are online.")
	}

	print.SuccessStatusEvent(os.Stdout, "Ark is up and running !")

}

type DockerUtils struct {
	cli *client.Client
	ctx context.Context
}

func NewDockerUtil() *DockerUtils {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &DockerUtils{
		cli: cli,
		ctx: ctx,
	}
}

func (d *DockerUtils) run(imageName string, envvars []string) (err error) {
	resp, err := d.cli.ContainerCreate(d.ctx, &container.Config{
		Image: imageName,
		Env:   envvars,
	}, nil, nil, nil, "")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	if err := d.cli.ContainerStart(d.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	var state string
	var running bool
	for i := 1; i <= 10; i++ {
		running, state, _ = d.IsRunning(imageName)
		if !running {
			fmt.Printf("Container state is :%v\n", state)
			time.Sleep(5 * time.Second)
		} else {
			// fmt.Printf("Container started successfuly:%s\n", resp.ID)
			break
		}
	}

	return err
}

func (d *DockerUtils) pull(imageName string) (err error) {
	_, err = d.cli.ImagePull(d.ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	} else {
		//fmt.Printf("Pulled Image: \t%s\n", imageName)
	}
	return err
}

func (d *DockerUtils) IsRunning(imageName string) (running bool, state string, id string) {
	running = false
	id = ""
	state = "none"

	containers, err := d.cli.ContainerList(d.ctx, types.ContainerListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	for _, container := range containers {
		if container.Image == imageName && container.State == "running" {
			id = container.ID[:12]
			running = true
			state = container.State
		} else if container.Image == imageName {
			state = container.State
		}
	}

	return running, state, id
}
