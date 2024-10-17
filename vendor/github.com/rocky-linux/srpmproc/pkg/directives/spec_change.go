// Copyright (c) 2021 The Srpmproc Authors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package directives

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/rocky-linux/srpmproc/pkg/data"
)

const (
	sectionChangelog = "%changelog"
)

var sections = []string{"%description", "%prep", "%build", "%install", "%files", "%changelog"}

type sourcePatchOperationInLoopRequest struct {
	cfg           *srpmprocpb.Cfg
	field         string
	value         *string
	longestField  int
	lastNum       *int
	in            *string
	expectedField string
	operation     srpmprocpb.SpecChange_FileOperation_Type
}

type sourcePatchOperationAfterLoopRequest struct {
	cfg           *srpmprocpb.Cfg
	inLoopNum     int
	lastNum       *int
	longestField  int
	newLines      *[]string
	in            *string
	expectedField string
	operation     srpmprocpb.SpecChange_FileOperation_Type
}

func sourcePatchOperationInLoop(req *sourcePatchOperationInLoopRequest) error {
	if strings.HasPrefix(req.field, req.expectedField) {
		for _, file := range req.cfg.SpecChange.File {
			if file.Type != req.operation {
				continue
			}

			switch file.Mode.(type) {
			case *srpmprocpb.SpecChange_FileOperation_Delete:
				if file.Name == *req.value {
					*req.value = ""
				}
				break
			}
		}
		if req.field != req.expectedField {
			sourceNum, err := strconv.Atoi(strings.Split(req.field, req.expectedField)[1])
			if err != nil {
				return errors.New(fmt.Sprintf("INVALID_%s_NUM:%s", strings.ToUpper(req.expectedField), req.field))
			}
			*req.lastNum = sourceNum
		}
	}

	return nil
}

func sourcePatchOperationAfterLoop(req *sourcePatchOperationAfterLoopRequest) (bool, error) {
	if req.inLoopNum == *req.lastNum && *req.in == req.expectedField {
		for _, file := range req.cfg.SpecChange.File {
			if file.Type != req.operation {
				continue
			}

			switch file.Mode.(type) {
			case *srpmprocpb.SpecChange_FileOperation_Add:
				fieldNum := *req.lastNum + 1
				field := fmt.Sprintf("%s%d", req.expectedField, fieldNum)
				spaces := calculateSpaces(req.longestField, len(field), req.cfg.SpecChange.DisableAutoAlign)
				*req.newLines = append(*req.newLines, fmt.Sprintf("%s:%s%s", field, spaces, file.Name))

				if req.expectedField == "Patch" && file.AddToPrep {
					val := fmt.Sprintf("%%patch -P%d", fieldNum)
					if file.NPath > 0 {
						val = fmt.Sprintf("%s -p%d", val, file.NPath)
					}

					req.cfg.SpecChange.Append = append(req.cfg.SpecChange.Append, &srpmprocpb.SpecChange_AppendOperation{
						Field: "%prep",
						Value: val,
					})
				}

				*req.lastNum++
				break
			}
		}
		*req.in = ""

		return true, nil
	}

	return false, nil
}

func calculateSpaces(longestField int, fieldLength int, disableAutoAlign bool) string {
	if disableAutoAlign {
		return " "
	}
	return strings.Repeat(" ", longestField+8-fieldLength)
}

func searchAndReplaceLine(line string, sar []*srpmprocpb.SpecChange_SearchAndReplaceOperation) string {
	for _, searchAndReplace := range sar {
		switch searchAndReplace.Identifier.(type) {
		case *srpmprocpb.SpecChange_SearchAndReplaceOperation_Any:
			line = strings.Replace(line, searchAndReplace.Find, searchAndReplace.Replace, int(searchAndReplace.N))
			break
		case *srpmprocpb.SpecChange_SearchAndReplaceOperation_StartsWith:
			if strings.HasPrefix(strings.TrimSpace(line), searchAndReplace.Find) {
				line = strings.Replace(line, searchAndReplace.Find, searchAndReplace.Replace, int(searchAndReplace.N))
			}
			break
		case *srpmprocpb.SpecChange_SearchAndReplaceOperation_EndsWith:
			if strings.HasSuffix(strings.TrimSpace(line), searchAndReplace.Find) {
				line = strings.Replace(line, searchAndReplace.Find, searchAndReplace.Replace, int(searchAndReplace.N))
			}
			break
		}
	}

	return line
}

func isNextLineSection(lineNum int, lines []string) bool {
	if len(lines)-1 > lineNum {
		if strings.HasPrefix(strings.TrimSpace(lines[lineNum+1]), "%") {
			return true
		}

		return false
	}

	return true
}

func setFASlice(futureAdditions map[int][]string, key int, addition string) {
	if futureAdditions[key] == nil {
		futureAdditions[key] = []string{}
	}
	futureAdditions[key] = append(futureAdditions[key], addition)
}

func strSliceContains(slice []string, str string) bool {
	for _, x := range slice {
		if str == x {
			return true
		}
	}

	return false
}

func specChange(cfg *srpmprocpb.Cfg, pd *data.ProcessData, md *data.ModeData, _ *git.Worktree, pushTree *git.Worktree) error {
	// no spec change operations present
	// skip parsing spec
	if cfg.SpecChange == nil {
		return nil
	}

	specFiles, err := pushTree.Filesystem.ReadDir("SPECS")
	if err != nil {
		return errors.New("COULD_NOT_READ_SPECS_DIR")
	}

	if len(specFiles) != 1 {
		return errors.New("ONLY_ONE_SPEC_FILE_IS_SUPPORTED")
	}

	filePath := filepath.Join("SPECS", specFiles[0].Name())
	stat, err := pushTree.Filesystem.Stat(filePath)
	if err != nil {
		return errors.New("COULD_NOT_STAT_SPEC_FILE")
	}

	specFile, err := pushTree.Filesystem.OpenFile(filePath, os.O_RDONLY, 0o644)
	if err != nil {
		return errors.New("COULD_NOT_READ_SPEC_FILE")
	}

	specBts, err := io.ReadAll(specFile)
	if err != nil {
		return errors.New("COULD_NOT_READ_ALL_BYTES")
	}

	specStr := string(specBts)
	lines := strings.Split(specStr, "\n")

	var newLines []string
	futureAdditions := map[int][]string{}
	newFieldMemory := map[string]map[string]int{}
	lastSourceNum := 0
	lastPatchNum := 0
	inSection := ""
	inField := ""
	lastSource := ""
	lastPatch := ""
	hasPatch := false

	version := ""
	importName := strings.Replace(pd.Importer.ImportName(pd, md), md.Name, "1", 1)
	importNameSplit := strings.SplitN(importName, "-", 2)
	if len(importNameSplit) == 2 {
		versionSplit := strings.SplitN(importNameSplit[1], ".el", 2)
		if len(versionSplit) == 2 {
			version = versionSplit[0]
		} else {
			versionSplit := strings.SplitN(importNameSplit[1], ".module", 2)
			if len(versionSplit) == 2 {
				version = versionSplit[0]
			}
		}
	}

	fieldValueRegex := regexp.MustCompile("^[a-zA-Z0-9]+:")

	longestField := 0
	for lineNum, line := range lines {
		if fieldValueRegex.MatchString(line) {
			fieldValue := strings.SplitN(line, ":", 2)
			field := strings.TrimSpace(fieldValue[0])
			longestField = int(math.Max(float64(len(field)), float64(longestField)))

			if strings.HasPrefix(field, "Source") {
				lastSource = field
			} else if strings.HasPrefix(field, "Patch") {
				lastPatch = field
				hasPatch = true
			} else {
				for _, nf := range cfg.SpecChange.NewField {
					if field == nf.Key {
						if newFieldMemory[field] == nil {
							newFieldMemory[field] = map[string]int{}
						}
						newFieldMemory[field][nf.Value] = lineNum
					}
				}
			}
		}
	}
	for _, nf := range cfg.SpecChange.NewField {
		if newFieldMemory[nf.Key] == nil {
			newFieldMemory[nf.Key] = map[string]int{}
			newFieldMemory[nf.Key][nf.Value] = 0
		}
	}

	for field, nfm := range newFieldMemory {
		for value, lineNum := range nfm {
			if lineNum != 0 {
				newLine := fmt.Sprintf("%s:%s%s", field, calculateSpaces(longestField, len(field), cfg.SpecChange.DisableAutoAlign), value)
				setFASlice(futureAdditions, lineNum+1, newLine)
			}
		}
	}

	for lineNum, line := range lines {
		inLoopSourceNum := lastSourceNum
		inLoopPatchNum := lastPatchNum
		prefixLine := strings.TrimSpace(line)

		for i, additions := range futureAdditions {
			if lineNum == i {
				for _, addition := range additions {
					newLines = append(newLines, addition)
				}
			}
		}

		if fieldValueRegex.MatchString(line) {
			line = searchAndReplaceLine(line, cfg.SpecChange.SearchAndReplace)
			fieldValue := strings.SplitN(line, ":", 2)
			field := strings.TrimSpace(fieldValue[0])
			value := strings.TrimSpace(fieldValue[1])

			if field == lastSource {
				inField = "Source"
			} else if field == lastPatch {
				inField = "Patch"
			}

			if field == "Version" && version == "" {
				version = value
			}

			for _, searchAndReplace := range cfg.SpecChange.SearchAndReplace {
				switch identifier := searchAndReplace.Identifier.(type) {
				case *srpmprocpb.SpecChange_SearchAndReplaceOperation_Field:
					if field == identifier.Field {
						value = strings.Replace(value, searchAndReplace.Find, searchAndReplace.Replace, int(searchAndReplace.N))
					}
					break
				}
			}

			for _, appendOp := range cfg.SpecChange.Append {
				if field == appendOp.Field {
					value = value + appendOp.Value

					if field == "Release" {
						version = version + appendOp.Value
					}
				}
			}

			spaces := calculateSpaces(longestField, len(field), cfg.SpecChange.DisableAutoAlign)

			err := sourcePatchOperationInLoop(&sourcePatchOperationInLoopRequest{
				cfg:           cfg,
				field:         field,
				value:         &value,
				lastNum:       &lastSourceNum,
				longestField:  longestField,
				in:            &inField,
				expectedField: "Source",
				operation:     srpmprocpb.SpecChange_FileOperation_Source,
			})
			if err != nil {
				return err
			}

			err = sourcePatchOperationInLoop(&sourcePatchOperationInLoopRequest{
				cfg:           cfg,
				field:         field,
				value:         &value,
				longestField:  longestField,
				lastNum:       &lastPatchNum,
				in:            &inField,
				expectedField: "Patch",
				operation:     srpmprocpb.SpecChange_FileOperation_Patch,
			})
			if err != nil {
				return err
			}

			if value != "" {
				newLines = append(newLines, fmt.Sprintf("%s:%s%s", field, spaces, value))
			}
		} else {
			executed, err := sourcePatchOperationAfterLoop(&sourcePatchOperationAfterLoopRequest{
				cfg:           cfg,
				inLoopNum:     inLoopSourceNum,
				lastNum:       &lastSourceNum,
				longestField:  longestField,
				newLines:      &newLines,
				expectedField: "Source",
				in:            &inField,
				operation:     srpmprocpb.SpecChange_FileOperation_Source,
			})
			if err != nil {
				return err
			}

			if executed && !hasPatch {
				newLines = append(newLines, "")
				inField = "Patch"
			}

			executed, err = sourcePatchOperationAfterLoop(&sourcePatchOperationAfterLoopRequest{
				cfg:           cfg,
				inLoopNum:     inLoopPatchNum,
				lastNum:       &lastPatchNum,
				longestField:  longestField,
				newLines:      &newLines,
				expectedField: "Patch",
				in:            &inField,
				operation:     srpmprocpb.SpecChange_FileOperation_Patch,
			})
			if err != nil {
				return err
			}

			if executed {
				var innerNewLines []string
				for field, nfm := range newFieldMemory {
					for value, ln := range nfm {
						newLine := fmt.Sprintf("%s:%s%s", field, calculateSpaces(longestField, len(field), cfg.SpecChange.DisableAutoAlign), value)
						if ln == 0 {
							if isNextLineSection(lineNum, lines) {
								innerNewLines = append(innerNewLines, newLine)
							}
						}
					}
				}
				if len(innerNewLines) > 0 {
					newLines = append(newLines, "")
					for _, il := range innerNewLines {
						newLines = append(newLines, il)
					}
				}
			}

			if executed && !strings.Contains(specStr, "%changelog") {
				newLines = append(newLines, "")
				newLines = append(newLines, "%changelog")
				inSection = sectionChangelog
			}

			if inSection == sectionChangelog {
				now := time.Now().Format("Mon Jan 02 2006")
				for _, changelog := range cfg.SpecChange.Changelog {
					newLines = append(newLines, fmt.Sprintf("* %s %s <%s> - %s", now, changelog.AuthorName, changelog.AuthorEmail, version))
					for _, msg := range changelog.Message {
						newLines = append(newLines, fmt.Sprintf("- %s", msg))
					}
					newLines = append(newLines, "")
				}
				inSection = ""
			} else {
				line = searchAndReplaceLine(line, cfg.SpecChange.SearchAndReplace)
			}

			if strings.HasPrefix(prefixLine, "%") {
				inSection = prefixLine

				for _, appendOp := range cfg.SpecChange.Append {
					if inSection == appendOp.Field {
						insertedLine := 0
						for i, x := range lines[lineNum+1:] {
							if strSliceContains(sections, strings.TrimSpace(x)) {
								insertedLine = lineNum + i
								setFASlice(futureAdditions, insertedLine, appendOp.Value)
								break
							}
						}
						if insertedLine == 0 {
							for i, x := range lines[lineNum+1:] {
								if strings.TrimSpace(x) == "" {
									insertedLine = lineNum + i + 2
									setFASlice(futureAdditions, insertedLine, appendOp.Value)
									break
								}
							}
						}
					}
				}
			}

			newLines = append(newLines, line)
		}
	}

	err = pushTree.Filesystem.Remove(filePath)
	if err != nil {
		return errors.New(fmt.Sprintf("COULD_NOT_REMOVE_OLD_SPEC_FILE:%s", filePath))
	}

	f, err := pushTree.Filesystem.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return errors.New(fmt.Sprintf("COULD_NOT_OPEN_REPLACEMENT_SPEC_FILE:%s", filePath))
	}

	_, err = f.Write([]byte(strings.Join(newLines, "\n")))
	if err != nil {
		return errors.New("COULD_NOT_WRITE_NEW_SPEC_FILE")
	}

	return nil
}
