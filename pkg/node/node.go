package node

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type NodePkg struct {
	NodeVer    func(node *NodePkg) error
	CheckNode  func(str string) error
	CmdBuilder func(name string, arg ...string) *exec.Cmd
}

func NewNode() *NodePkg {
	return &NodePkg{
		NodeVer:    nodeVersion,
		CheckNode:  checkNode,
		CmdBuilder: exec.Command,
	}
}

func nodeVersion(node *NodePkg) error {
	cmd := node.CmdBuilder("node", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return errors.New(NODE_NOT_INSTALLED)
	}
	return node.CheckNode(out.String())
}

func checkNode(str string) error {
	versionOutput := strings.TrimSpace(str)
	if len(versionOutput) > 0 && versionOutput[0] == 'v' {
		versionOutput = versionOutput[1:]
	}

	versionParts := strings.Split(versionOutput, ".")
	if len(versionParts) > 0 {
		majorVersion := versionParts[0]

		var major int
		fmt.Sscanf(majorVersion, "%d", &major)

		if major < 18 {
			return errors.New(NODE_OLDER_VERSION)
		}
	}
	return nil
}
