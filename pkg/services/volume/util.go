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

package volume

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func which(bins ...string) ([]string, error) {
	paths := make([]string, 0, len(bins))

	out, _ := exec.Command("which", bins...).Output()
	r := bufio.NewReader(bytes.NewBuffer(out))
	for {
		p, err := r.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if p = strings.TrimSpace(p); !path.IsAbs(p) {
			continue
		}
		realPath, err := filepath.EvalSymlinks(p)
		if err != nil {
			return nil, err
		}
		paths = append(paths, realPath)
	}
	return paths, nil
}

func clone(src, dst string) error {
	// Prefer hard link, fallback to copy
	err := os.Link(src, dst)
	if err != nil {
		err = fallbackCopy(src, dst)
	}
	return err
}

func fallbackCopy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	fi, err := s.Stat()
	if err != nil {
		return err
	}

	d, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	if err := d.Chmod(fi.Mode()); err != nil {
		d.Close()
		return err
	}

	return d.Close()
}

func blacklisted(file string, obj *elf.File) (bool, error) {
	lib := regexp.MustCompile(`^.*/lib([\w-]+)\.so[\d.]*$`)
	glcore := regexp.MustCompile(`libnvidia-e?glcore\.so`)
	gldispatch := regexp.MustCompile(`libGLdispatch\.so`)

	if m := lib.FindStringSubmatch(file); m != nil {
		switch m[1] {
		// Blacklist EGL/OpenGL libraries issued by other vendors
		case "EGL":
			fallthrough
		case "GLESv1_CM":
			fallthrough
		case "GLESv2":
			fallthrough
		case "GL":
			deps, err := obj.DynString(elf.DT_NEEDED)
			if err != nil {
				return false, err
			}
			for _, d := range deps {
				if glcore.MatchString(d) || gldispatch.MatchString(d) {
					return false, nil
				}
			}
			return true, nil

		// Blacklist TLS libraries using the old ABI (!= 2.3.99)
		case "nvidia-tls":
			const abi = 0x6300000003
			s, err := obj.Section(".note.ABI-tag").Data()
			if err != nil {
				return false, err
			}
			return binary.LittleEndian.Uint64(s[24:]) != abi, nil
		}
	}
	return false, nil
}
