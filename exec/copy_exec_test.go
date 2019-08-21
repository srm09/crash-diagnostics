package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gitlab.eng.vmware.com/vivienv/flare/script"
)

func TestExecCOPY(t *testing.T) {
	tests := []execTest{
		{
			name: "COPY single files",
			source: func() string {
				return "COPY /tmp/foo0.txt"
			},
			exec: func(s *script.Script) error {
				machine := s.Preambles[script.CmdFrom][0].(*script.FromCommand).Machines()[0].Address
				workdir := s.Preambles[script.CmdWorkDir][0].(*script.WorkdirCommand)

				cpCmd := s.Actions[0].(*script.CopyCommand)
				srcFile := cpCmd.Args()[0]
				if err := makeTestFakeFile(t, srcFile, "HelloFoo"); err != nil {
					return err
				}
				defer os.Remove(srcFile)

				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}

				fileName := filepath.Join(workdir.Dir(), machine, filepath.Base(srcFile))
				if _, err := os.Stat(fileName); err != nil {
					return err
				}

				return nil
			},
		},

		{
			name: "COPY multiple files",
			source: func() string {
				return "COPY /tmp/foo0.txt\nCOPY /tmp/foo1.txt /tmp/foo2.txt"
			},
			exec: func(s *script.Script) error {
				machine := s.Preambles[script.CmdFrom][0].(*script.FromCommand).Machines()[0].Address
				workdir := s.Preambles[script.CmdWorkDir][0].(*script.WorkdirCommand)

				var srcFiles []string
				cpCmd0 := s.Actions[0].(*script.CopyCommand)
				srcFiles = append(srcFiles, cpCmd0.Args()[0])
				cpCmd1 := s.Actions[1].(*script.CopyCommand)
				srcFiles = append(srcFiles, cpCmd1.Args()[0])
				srcFiles = append(srcFiles, cpCmd1.Args()[1])

				for i, srcFile := range srcFiles {
					if err := makeTestFakeFile(t, srcFile, fmt.Sprintf("HelloFoo-%d", i)); err != nil {
						return err
					}
					defer os.Remove(srcFile)

				}

				e := New(s)
				if err := e.Execute(); err != nil {
					return err
				}

				for _, srcFile := range srcFiles {
					fileName := filepath.Join(workdir.Dir(), machine, filepath.Base(srcFile))
					if _, err := os.Stat(fileName); err != nil {
						return err
					}
				}

				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runExecutorTest(t, test)
		})
	}
}
