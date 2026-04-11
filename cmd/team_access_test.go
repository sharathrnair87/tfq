//go:build all
// +build all

package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeamAccessListCmd(t *testing.T) {
	buf := new(bytes.Buffer)

	tr := rootCmd
	tc := teamCmd
	tlc := teamListCmd

	tr.AddCommand(tc, tlc)
	tr.SetOut(buf)
	tr.SetErr(buf)
	tr.SetArgs([]string{"team", "list", "--query", ".[0].id"})

	err := tr.Execute()
	if err != nil {
		t.Fatalf("Error executing command: %v", err)
	}

	out := strings.TrimSpace(buf.String())

	re := regexp.MustCompile(`"([^"]*)"`)
	matches := re.FindStringSubmatch(out)

	teamId := "NA"

	if len(matches) >= 2 {
		teamId = matches[1]
	}

	if teamId != "NA" {
		tar := rootCmd
		ta := teamAccessCmd
		tla := teamAccessListCmd

		tar.AddCommand(ta, tla)

		tar.SetArgs([]string{"team-access", "list", "--team-id", teamId})

		rbuf := bytes.NewBufferString("")
		tar.SetOut(rbuf)

		err := tar.Execute()
		fmt.Print(rbuf.String())

		if err != nil {
			t.Fatalf("Error executing command: %v", err)
		}
	}

	require.Nil(t, err)
}
