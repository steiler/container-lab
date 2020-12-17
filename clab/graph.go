package clab

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/awalterschulze/gographviz"
	log "github.com/sirupsen/logrus"
)

var g *gographviz.Graph

// GenerateGraph generates a graph of the lab topology
func (c *cLab) GenerateGraph(topo string) error {
	log.Info("Generating lab graph...")
	g = gographviz.NewGraph()
	if err := g.SetName(c.TopoFile.name); err != nil {
		return err
	}
	if err := g.SetDir(false); err != nil {
		return err
	}

	var attr map[string]string

	// Process the Nodes
	for nodeName, node := range c.Nodes {
		attr = make(map[string]string)
		attr["color"] = "red"
		attr["style"] = "filled"
		attr["fillcolor"] = "red"

		attr["label"] = nodeName
		attr["xlabel"] = node.Kind
		if len(strings.TrimSpace(node.Group)) != 0 {
			attr["group"] = node.Group
			if strings.Contains(node.Group, "bb") {
				attr["fillcolor"] = "blue"
				attr["color"] = "blue"
				attr["fontcolor"] = "white"
			} else if strings.Contains(node.Kind, "srl") {
				attr["fillcolor"] = "green"
				attr["color"] = "green"
				attr["fontcolor"] = "black"
			}
		}
		if err := g.AddNode(c.TopoFile.name, node.ShortName, attr); err != nil {
			return err
		}

	}

	// Process the links inbetween Nodes
	for _, link := range c.Links {
		attr = make(map[string]string)
		attr["color"] = "black"

		if (strings.Contains(link.A.Node.ShortName, "client")) || (strings.Contains(link.B.Node.ShortName, "client")) {
			attr["color"] = "blue"
		}
		if err := g.AddEdge(link.A.Node.ShortName, link.B.Node.ShortName, false, attr); err != nil {
			return err
		}
		//log.Info(link.A.Node.ShortName, " <-> ", link.B.Node.ShortName)
	}

	// create graph directory
	CreateDirectory(c.Dir.Lab, 0755)
	CreateDirectory(c.Dir.LabGraph, 0755)

	// create graph filename
	dotfile := c.Dir.LabGraph + "/" + c.TopoFile.name + ".dot"
	createFile(dotfile, g.String())
	log.Infof("Created %s", dotfile)

	pngfile := c.Dir.LabGraph + "/" + c.TopoFile.name + ".png"

	// Only try to create png
	if commandExists("dot") {
		generatePngFromDot(dotfile, pngfile)
		log.Info("Created ", pngfile)
	}
	return nil
}

// generatePngFromDot generated PNG from the provided dot file
func generatePngFromDot(dotfile string, outfile string) (err error) {
	_, err = exec.Command("dot", "-o", outfile, "-Tpng", dotfile).CombinedOutput()
	if err != nil {
		log.Errorf("failed to generate png (%v) from dot file (%v), with error (%v)", outfile, dotfile, err)
		return fmt.Errorf("failed to generate png (%v) from dot file (%v), with error (%v)", outfile, dotfile, err)
	}
	return nil
}

// commandExists checks for the existence of the given command on the system
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	if err == nil {
		log.Debugf("executable %s exists!", cmd)
	} else {
		log.Debugf("executable %s doesn't exist!", cmd)
	}
	return err == nil
}
