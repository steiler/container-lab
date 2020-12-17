package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/srl-wim/container-lab/clab"
)

var cleanup bool
var graceful bool

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Short:   "destroy a lab",
	Aliases: []string{"des"},
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := []clab.ClabOption{
			clab.WithDebug(debug),
			clab.WithTimeout(timeout),
			clab.WithTopoFile(topo),
			clab.WithEnvDockerClient(),
			clab.WithGracefulShutdown(graceful),
		}
		c := clab.NewContainerLab(opts...)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var err error
		// Parse topology information
		if err = c.ParseTopology(); err != nil {
			return err
		}

		if cleanup {
			err = os.RemoveAll(c.Dir.Lab)
			if err != nil {
				log.Errorf("error deleting lab directory: %v", err)
			}
		}
		containers, err := c.ListContainers(ctx, []string{fmt.Sprintf("containerlab=lab-%s", c.Config.Name)})
		if err != nil {
			return fmt.Errorf("could not list containers: %v", err)
		}

		log.Infof("Destroying container lab: %s", topo)
		wg := new(sync.WaitGroup)
		wg.Add(len(containers))
		for _, cont := range containers {
			go func(cont types.Container) {
				defer wg.Done()
				name := cont.ID
				if len(cont.Names) > 0 {
					name = strings.TrimLeft(cont.Names[0], "/")
				}
				log.Infof("Stopping container: %s", name)
				err := c.DeleteContainer(ctx, name)
				if err != nil {
					log.Errorf("could not remove container '%s': %v", name, err)
				}
			}(cont)
		}
		wg.Wait()
		log.Info("Removing container entries from /etc/hosts file")
		err = deleteEntriesFromHostsFile(containers, c.Config.Mgmt.Network)
		if err != nil {
			return err
		}

		// delete container management bridge
		log.Info("Deleting docker bridge ...")
		if err = c.DeleteBridge(ctx); err != nil {
			log.Error(err)
		}
		// delete container network namespaces symlinks
		c.DeleteNetnsSymlinks()
		c.InitVirtualWiring()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	destroyCmd.Flags().BoolVarP(&cleanup, "cleanup", "", false, "delete lab directory")
	destroyCmd.Flags().BoolVarP(&graceful, "graceful", "", false, "attempt to stop containers before removing")
}

func deleteEntriesFromHostsFile(containers []types.Container, bridgeName string) error {
	if bridgeName == "" {
		return fmt.Errorf("missing bridge name")
	}
	f, err := os.OpenFile("/etc/hosts", os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	data := hostsEntries(containers, bridgeName)
	remainingLines := make([][]byte, 0)
	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		found := false
		sLine := strings.Join(strings.Fields(string(line)), " ")
		for _, dl := range strings.Split(string(data), "\n") {
			sdl := strings.Join(strings.Fields(string(dl)), " ")
			if strings.Compare(sLine, sdl) == 0 {
				found = true
				break
			}
		}
		if !found {
			remainingLines = append(remainingLines, line)
		}
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	for _, l := range remainingLines {
		f.Write(l)
		f.Write([]byte("\n"))
	}
	return nil
}
