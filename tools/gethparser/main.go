package main

import (
    "bufio"
    "encoding/json"
    "os"
    "bytes"

    "github.com/spf13/cobra"
    "github.com/whiteblock/genesis-cli/pkg/util"
)

func main() {
    if err := root.Execute(); err != nil {
        util.ErrorFatal(err)
    }
}

var root = &cobra.Command{
    Use:     "parser <source-file> <destination-directory>",
    Version: "1.0.0",
    Run: func(cmd *cobra.Command, args []string) {
        util.CheckArguments(cmd, args, 2, 2)
        source, err := os.Open(args[0])
        if err != nil {
            util.ErrorFatal(err)
        }
        defer source.Close()

        out := &Outputter{
            Destination: args[1],
            Split:      true,
        }
        err = out.Setup()
        if err != nil {
            util.ErrorFatal(err)
        }

        rdr := bufio.NewReader(source)

        for {
            data, err := rdr.ReadBytes(byte('\n'))
            if err != nil {
                break   // assumes new line at EOF
            }
            data = data[bytes.IndexByte(data, '{'):] // trim timestamp
            var item SyslogngOutput
            err = json.Unmarshal(data, &item)
            if err != nil {
                util.ErrorFatal(err)
            }
            err = out.handleInput(item, util.GetStringFlagValue(cmd, "test-id"))
            if err != nil {
                util.ErrorFatal(err)
            }
        }
    },
}

func init() {
    root.Flags().StringP("test-id", "t", "", "Genesis test ID to parse (required)")
    root.MarkFlagRequired("test-id")
}