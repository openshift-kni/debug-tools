/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2022 Red Hat, Inc.
 */

package ethtool

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-kni/debug-tools/internal/cli/knit"
	goethtool "github.com/safchain/ethtool"
)

type ethtoolOptions struct {
	showFeatures bool
	showChannels bool
}

func NewEthtoolCommand(knitOpts *knit.KnitOptions) *cobra.Command {
	opts := &ethtoolOptions{}
	eInfo := &cobra.Command{
		Use:   "ethtool",
		Short: "subset of ethtool query capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showEthtool(cmd, knitOpts, opts, args)
		},
		Args: cobra.MaximumNArgs(1),
	}
	eInfo.Flags().BoolVarP(&opts.showFeatures, "show-features", "k", false, "show the features of the selected devices")
	eInfo.Flags().BoolVarP(&opts.showChannels, "show-channels", "l", false, "show the channels of the selected devices")
	return eInfo
}

func showEthtool(cmd *cobra.Command, knitOpts *knit.KnitOptions, opts *ethtoolOptions, args []string) error {
	var err error
	ifaces := args
	if len(ifaces) == 0 {
		ifaces, err = getAllInterfaceNames()
	}
	if err != nil {
		return err
	}

	needSeparator := false
	if len(ifaces) > 1 {
		needSeparator = true
	}

	ethHandle, err := goethtool.NewEthtool()
	if err != nil {
		return err
	}
	defer ethHandle.Close()

	for _, iface := range ifaces {
		if opts.showFeatures {
			if err := showEthtoolFeatures(ethHandle, iface, knitOpts.JsonOutput); err != nil {
				return err
			}
		}
		if opts.showChannels {
			if err := showEthtoolChannels(ethHandle, iface, knitOpts.JsonOutput); err != nil {
				return err
			}
		}
		if needSeparator {
			fmt.Printf("\n")
		}
	}
	return nil
}

func showEthtoolFeatures(et *goethtool.Ethtool, iface string, jsonMode bool) error {
	feats, err := et.Features(iface)
	if err != nil {
		return err
	}
	if jsonMode {
		json.NewEncoder(os.Stdout).Encode(feats)
		return nil
	}
	fmt.Printf("Features for %s:\n", iface)
	for key, val := range feats {
		fmt.Printf("%s: %s\n", key, toggle(val))
	}
	return nil
}

func showEthtoolChannels(et *goethtool.Ethtool, iface string, jsonMode bool) error {
	chans, err := et.GetChannels(iface)
	if err != nil {
		return err
	}
	if jsonMode {
		json.NewEncoder(os.Stdout).Encode(chans)
		return nil
	}
	fmt.Printf("Channel parameters for %s:\n", iface)
	fmt.Printf("Pre-set maximums:\n")
	fmt.Printf("RX:		%d\n", chans.MaxRx)
	fmt.Printf("TX:		%d\n", chans.MaxTx)
	fmt.Printf("Other:		%d\n", chans.MaxOther)
	fmt.Printf("Combined:	%d\n", chans.MaxCombined)
	fmt.Printf("Current hardware settings:\n")
	fmt.Printf("RX:		%d\n", chans.RxCount)
	fmt.Printf("TX:		%d\n", chans.TxCount)
	fmt.Printf("Other:		%d\n", chans.OtherCount)
	fmt.Printf("Combined:	%d\n", chans.CombinedCount)
	return nil
}

func toggle(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func getAllInterfaceNames() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var names []string
	for _, iface := range ifaces {
		if (iface.Flags & net.FlagUp) != net.FlagUp {
			continue
		}
		if (iface.Flags & net.FlagLoopback) == net.FlagLoopback {
			continue
		}
		names = append(names, iface.Name)
	}
	return names, nil
}
