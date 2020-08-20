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
    Use:     "parser <source-file> <destination-file>",
    Version: "1.0.0",
    Run: func(cmd *cobra.Command, args []string) {
        util.CheckArguments(cmd, args, 2, 2)
        source, err := os.Open(args[0])
        if err != nil {
            util.ErrorFatal(err)
        }
        defer source.Close()

        out := &Outputter{
            Destination: util.GetStringFlagValue(cmd, "test-id"),
            Split:       util.GetBoolFlagValue(cmd, "split"),
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
    // TODO: Clean up this workaround. "split" option no longer needed. Should
    // be a default behavior.
    root.Flags().BoolP("split", "s", true, "treat destination file as a directory and output to a different file for each container")
    root.Flags().StringP("test-id", "t", "", "filter logs from specific test ID")
}