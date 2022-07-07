// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"strings"
	"unicode/utf8"
)

type Progress struct {
	Instance   *mpb.Progress
	Bar        *mpb.Bar
	StatusBars []*mpb.Bar
	KeepLogs   bool

	MaxLinesNum            int
	MaxLinesNumAfterFinish int

	statusHeader string
	progressText string
	successText  string
	failedText   string
	statusLines  []string
	isFinished   bool
	didFail      bool
}

func NewProgress(instance *mpb.Progress, progressText string, successText string, failedText string) *Progress {
	if instance == nil {
		return &Progress{}
	}

	p := &Progress{
		Instance:               instance,
		MaxLinesNum:            15,
		MaxLinesNumAfterFinish: 10,
		progressText:           progressText,
		successText:            successText,
		failedText:             failedText,
	}

	p.setStatusHeader(p.progressText)
	p.Bar = p.Instance.Add(1,
		mpb.NewBarFiller(mpb.BarStyle().Tip(`-`, `\`, `|`, `/`).Filler("").Padding(" ").Lbound("").Rbound("")),
		mpb.PrependDecorators(decor.Any(func(_ decor.Statistics) string {
			if p.isFinished {
				if p.didFail {
					return color.RedString(p.statusHeader)
				}
				return color.GreenString(p.statusHeader)
			} else {
				return color.CyanString(p.statusHeader)
			}
		})),
	)

	for i := 0; i < p.MaxLinesNum; i++ {
		p.statusLines = append(p.statusLines, "...")
	}

	for i := 0; i < p.MaxLinesNum; i++ {
		currentNum := i
		statusBar := p.Instance.Add(1, mpb.NewBarFiller(mpb.BarStyle().Lbound("").Rbound("").Padding(" ").Filler(" ")), mpb.PrependDecorators(
			decor.Any(func(s decor.Statistics) string {
				return fmt.Sprintf("    => %s", p.statusLines[currentNum])
			}),
		))
		p.StatusBars = append(p.StatusBars, statusBar)
	}

	return p
}

func (p *Progress) AppendStatus(format string, args ...interface{}) {
	if p.Instance == nil {
		return
	}

	statusLine := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(fmt.Sprintf(format, args...)), "\n", ""), "\r\n", ""), "\r", ""), "\x00", "")
	if !utf8.ValidString(statusLine) {
		v := make([]rune, 0, len(statusLine))
		for i, r := range statusLine {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(statusLine[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		statusLine = string(v)
	}

	p.statusLines = append(p.statusLines[1:], statusLine)
}

func (p *Progress) Finish(didFail bool) {
	if p.Instance == nil {
		return
	}

	p.didFail = didFail
	if didFail {
		p.setStatusHeader(p.failedText)
	} else {
		p.setStatusHeader(p.successText)
	}

	for i, bar := range p.StatusBars {
		if len(p.StatusBars) > p.MaxLinesNumAfterFinish && i < p.MaxLinesNum-p.MaxLinesNumAfterFinish-1 {
			bar.Abort(true)
		} else {
			bar.Abort(!p.KeepLogs)
		}
	}

	p.Bar.Increment()

	p.isFinished = true
}

func (p *Progress) setStatusHeader(header string) {
	p.statusHeader = fmt.Sprintf("  ==> %s", header)
}
