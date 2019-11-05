/*
 * Tencent is pleased to support the open source community by making TKEStack available.
 *
 * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

package checkpoint

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// Name prefix for the temporary files.
	tmpPrefix = "."
)

var (
	// ErrKeyNotFound is the error returned if key is not found in Store.
	ErrKeyNotFound = fmt.Errorf("key is not found")
)

// CheckpointManager stores checkpoint in file.
type Manager struct {
	// Absolute path to the base directory for storing checkpoint files.
	directoryPath string
	// File name of the storing checkpoint file.
	file string
}

// NewManager returns an instance of CheckpointManager.
func NewManager(path string, file string) (*Manager, error) {
	if err := ensureDirectory(path); err != nil {
		return nil, err
	}

	return &Manager{directoryPath: path, file: file}, nil
}

// Write writes the given checkpoint to file.
func (f *Manager) Write(data []byte) error {
	if err := ensureDirectory(f.directoryPath); err != nil {
		return err
	}

	return writeFile(f.getPathOfFile(), data)
}

// Read reads the checkpoint from the file.
func (f *Manager) Read() ([]byte, error) {
	bytes, err := ioutil.ReadFile(f.getPathOfFile())
	if os.IsNotExist(err) {
		return bytes, ErrKeyNotFound
	}
	return bytes, err
}

// Delete deletes the file.
func (f *Manager) Delete() error {
	return removePath(f.getPathOfFile())
}

// getPathOfFile returns the full path of the file.
func (f *Manager) getPathOfFile() string {
	return filepath.Join(f.directoryPath, f.file)
}

// ensureDirectory creates the directory if it does not exist.
func ensureDirectory(path string) error {
	if _, err := os.Stat(path); err != nil {
		// MkdirAll returns nil if directory already exists.
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// writeFile writes checkpoint to path in a single transaction.
func writeFile(path string, data []byte) (retErr error) {
	// Create a temporary file in the base directory of `path` with a prefix.
	tmpFile, err := ioutil.TempFile(filepath.Dir(path), tmpPrefix)
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	shouldClose := true

	defer func() {
		// Close the file.
		if shouldClose {
			if err := tmpFile.Close(); err != nil {
				if retErr == nil {
					retErr = fmt.Errorf("close error: %v", err)
				} else {
					retErr = fmt.Errorf("failed to close temp file after error %v; close error: %v", retErr, err)
				}
			}
		}

		// Clean up the temp file on error.
		if retErr != nil && tmpPath != "" {
			if err := removePath(tmpPath); err != nil {
				retErr = fmt.Errorf("failed to remove the temporary file (%q) after error %v; remove error: %v", tmpPath, retErr, err)
			}
		}
	}()

	// Write checkpoint.
	if _, err := tmpFile.Write(data); err != nil {
		return err
	}

	// Sync file.
	if err := tmpFile.Sync(); err != nil {
		return err
	}

	// Closing the file before renaming.
	err = tmpFile.Close()
	shouldClose = false
	if err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func removePath(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
