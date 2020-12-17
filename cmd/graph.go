package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/srl-wim/container-lab/clab"
)

const (
	defaultTemplatePath = "/etc/containerlab/templates/graph/index.html"
)

var srv string
var tmpl string

type graphTopo struct {
	Nodes []containerDetails `json:"nodes,omitempty"`
	Links []link             `json:"links,omitempty"`
}
type link struct {
	Source         string `json:"source,omitempty"`
	SourceEndpoint string `json:"source_endpoint,omitempty"`
	Target         string `json:"target,omitempty"`
	TargetEndpoint string `json:"target_endpoint,omitempty"`
}

type topoData struct {
	Name string
	Data template.JS
}

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "generate a topology graph",

	RunE: func(cmd *cobra.Command, args []string) error {
		opts := []clab.ClabOption{
			clab.WithDebug(debug),
			clab.WithTimeout(timeout),
			clab.WithTopoFile(topo),
			clab.WithEnvDockerClient(),
		}
		c := clab.NewContainerLab(opts...)

		// Parse topology information
		if err := c.ParseTopology(); err != nil {
			return err
		}

		if srv == "" {
			if err := c.GenerateGraph(topo); err != nil {
				return err
			}
			return nil
		}
		gtopo := graphTopo{
			Nodes: make([]containerDetails, 0, len(c.Nodes)),
			Links: make([]link, 0, len(c.Links)),
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		containers, err := c.ListContainers(ctx, []string{fmt.Sprintf("containerlab=lab-%s", c.Config.Name)})
		if err != nil {
			log.Errorf("could not list containers: %v", err)
		}
		log.Debugf("found %d containers", len(containers))
		for _, cont := range containers {
			var name string
			if len(cont.Names) > 0 {
				name = strings.TrimPrefix(cont.Names[0], fmt.Sprintf("/clab-%s-", c.Config.Name))
			}
			log.Debugf("looking for node name %s", name)
			if node, ok := c.Nodes[name]; ok {
				gtopo.Nodes = append(gtopo.Nodes, containerDetails{
					Name:        name,
					Kind:        node.Kind,
					Image:       cont.Image,
					Group:       node.Group,
					State:       fmt.Sprintf("%s/%s", cont.State, cont.Status),
					IPv4Address: getContainerIPv4(cont, c.Config.Mgmt.Network),
					IPv6Address: getContainerIPv6(cont, c.Config.Mgmt.Network),
				})
			}
		}
		sort.Slice(gtopo.Nodes, func(i, j int) bool {
			return gtopo.Nodes[i].Name < gtopo.Nodes[j].Name
		})
		for _, l := range c.Links {
			gtopo.Links = append(gtopo.Links, link{
				Source:         l.A.Node.ShortName,
				SourceEndpoint: l.A.EndpointName,
				Target:         l.B.Node.ShortName,
				TargetEndpoint: l.B.EndpointName,
			})
		}
		b, err := json.Marshal(gtopo)
		if err != nil {
			return err
		}
		log.Debugf("generating graph using data: %s", string(b))
		topoD := topoData{
			Name: c.Config.Name,
			Data: template.JS(string(b)),
		}
		tmpl := template.Must(template.ParseFiles(tmpl))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			tmpl.Execute(w, topoD)
		})

		log.Infof("Listening on %s...", srv)
		err = http.ListenAndServe(srv, nil)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().StringVarP(&srv, "srv", "s", "", "HTTP server address to view, customize and export your topology")
	graphCmd.Flags().StringVarP(&tmpl, "template", "", defaultTemplatePath, "Go html template used to generate the graph")
}
