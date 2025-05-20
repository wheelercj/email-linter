// Copyright 2025 Chris Wheeler

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove your API token from your system's keyring",
	Args:  logoutArgsFunc,
	RunE:  logoutRunFunc,
}

// logoutArgsFunc ensures no unexpected arguments were provided to the logout command.
func logoutArgsFunc(cmd *cobra.Command, args []string) error {
	return cobra.NoArgs(cmd, args)
}

// logoutRunFunc removes the user's API token from the system's keyring.
func logoutRunFunc(cmd *cobra.Command, args []string) error {
	return keyring.Delete(serviceName, userName)
}
