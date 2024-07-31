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

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"openapi.peridot.resf.org/peridotopenapi"
)

var taskInfo = &cobra.Command{
	Use:  "info [name-or-buildId]",
	Args: cobra.ExactArgs(1),
	Run:  taskInfoMn,
}

var (
	showLogLink       bool
	showSubmitterInfo bool
	showDuration      bool
)

func init() {
	taskInfo.Flags().BoolVar(&succeeded, "succeeded", true, "only query successful tasks")
	taskInfo.Flags().BoolVar(&cancelled, "cancelled", false, "only query cancelled tasks")
	taskInfo.Flags().BoolVar(&failed, "failed", false, "only query failed tasks")
	taskInfo.MarkFlagsMutuallyExclusive("cancelled", "failed", "succeeded")

	taskInfo.Flags().BoolVarP(&showLogLink, "logs", "L", false, "include log link in output (table format only)")
	taskInfo.Flags().BoolVar(&showSubmitterInfo, "submitter", false, "include submitter details (table format only)")
	taskInfo.Flags().BoolVar(&showDuration, "duration", true, "include duration from start to stop (table format only)")
}

func getNextColor(colors tablewriter.Colors) tablewriter.Colors {
	bgColor := getNextBackgroundColor(colors[0])
	fgColor := colors[1]
	if bgColor == -1 {
		fgColor = getNextForegroundColor(0)
		bgColor = getNextBackgroundColor(0)
	}
	return tablewriter.Colors{bgColor, fgColor}
}

func getNextForegroundColor(color int) int {
	switch color {
	case 0:
		return tablewriter.FgGreenColor
	case tablewriter.FgCyanColor:
		return tablewriter.FgHiGreenColor
	case tablewriter.FgHiCyanColor:
		return tablewriter.FgGreenColor
	default:
		color++
		return color
	}
}

func getNextBackgroundColor(color int) int {
	switch color {
	case 0:
		return tablewriter.BgRedColor
	case tablewriter.BgCyanColor:
		return tablewriter.BgHiRedColor
	case tablewriter.BgHiCyanColor:
		return -1
	default:
		color++
		return color
	}
}

func buildHeaderAndAutoMergeCells() ([]string, []int) {
	header := []string{"ptid", "tid", "status", "type", "arch", "created", "finished"}
	mergableNames := []string{"ptid", "type", "arch"}
	var autoMergeCells []int

	// Conditional appending to header
	if showDuration {
		header = append(header, "duration")
		mergableNames = append(mergableNames, "duration")
	}
	if showSubmitterInfo {
		header = append(header, "submitter")
		mergableNames = append(mergableNames, "submitter")
	}
	if showLogLink {
		header = append(header, "logs")
	}

	// Determine dynamic indices for auto-merge cells
	for _, itemName := range mergableNames {
		index := slices.Index(header, itemName)
		if index != -1 {
			autoMergeCells = append(autoMergeCells, index)
		}
	}

	return header, autoMergeCells
}

func convertSubTaskSliceToCSV(task peridotopenapi.V1AsyncTask) {
	subtasks, ok := task.GetSubtasksOk()
	if !ok {
		errFatal(fmt.Errorf("error getting subtasks: %v", ok))
	}

	var parentTask = (*subtasks)[0]

	var table = tablewriter.NewWriter(os.Stdout)
	// var data [][]string

	var header, autoMergeCells = buildHeaderAndAutoMergeCells()

	var lastColor = tablewriter.Colors{0, tablewriter.FgWhiteColor}
	var seenTasksColors = make(map[string]tablewriter.Colors)

	var parentTaskIds []string // cache parentTaskIds for colorizing

	// precache all the subtask's parent tasks so we know if we should color them
	for _, subtask := range *subtasks {
		parentTaskIds = append(parentTaskIds, subtask.GetParentTaskId())
	}

	for _, subtask := range *subtasks {
		json, err := subtask.MarshalJSON()
		if err != nil {
			errFatal(err)
		}

		if debug() {
			err = PrettyPrintJSON(json)
			if err != nil {
				errFatal(err)
			}
			// taskResponse, _ := subtask.GetResponse().MarshalJSON()
			// taskMetadata, _ := subtask.GetMetadata().MarshalJSON()
		}

		subtaskId := subtask.GetId()
		subtaskParentTaskId := subtask.GetParentTaskId()
		createdAt := subtask.GetCreatedAt()
		finishedAt := subtask.GetFinishedAt()

		row := []string{
			subtaskParentTaskId,
			subtaskId,
			string(subtask.GetStatus()),
			string(subtask.GetType()),
			subtask.GetArch(),
			formatTime(createdAt),
			formatTime(finishedAt),
		}

		if showDuration {
			row = append(row, formatDuration(createdAt, finishedAt))
		}

		if showSubmitterInfo {
			effectiveSubmitter := fmt.Sprintf("%s <%s>", parentTask.GetSubmitterId(), parentTask.GetSubmitterEmail())
			row = append(row, effectiveSubmitter)
		}

		if showLogLink {
			row = append(row, getLogLink(subtaskId))
		}

		if !color() {
			table.Append(row)
			continue
		}

		nextColor := tablewriter.Colors{tablewriter.BgBlackColor, tablewriter.FgWhiteColor}
		needsColor := hasAny(parentTaskIds, subtaskId)
		if _, seen := seenTasksColors[subtaskId]; !seen && needsColor {
			debugP("before: lastcolor: %v next: %v", lastColor, nextColor)
			nextColor = getNextColor(lastColor)
			debugP("after: lastcolor: %v next: %v", lastColor, nextColor)
			lastColor = nextColor
			seenTasksColors[subtaskId] = nextColor
		}

		tidColors := nextColor

		ptidColors := tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.BgBlackColor}
		if seenColor, seen := seenTasksColors[subtaskParentTaskId]; seen {
			ptidColors = seenColor
		}

		var colors = make([]tablewriter.Colors, len(row))

		debugP("tidcolor: %v ptidcolor %d", tidColors, ptidColors)
		for i, v := range header {
			switch v {
			case "ptid":
				colors[i] = ptidColors
			case "tid":
				colors[i] = tidColors
			default:
				colors[i] = tablewriter.Colors{}
			}
		}

		table.Rich(row, colors)
	}

	table.SetHeader(header)
	table.SetAutoMergeCellsByColumnIndex(autoMergeCells)
	table.SetRowLine(true)
	table.Render()

}

func debugP(s string, args ...any) {
	if debug() {
		log.Printf(s, args...)
	}
}

func hasAny(slice []string, target string) bool {
	if idx := slices.Index(slice, target); idx >= 0 {
		return true
	}
	return false
}

func taskInfoMn(_ *cobra.Command, args []string) {
	// Ensure project id exists
	projectId := mustGetProjectID()

	taskId := args[0]

	err := uuid.Validate(taskId)
	if err != nil {
		errFatal(errors.New("invalid task id"))
	}

	taskCl := getClient(serviceTask).(peridotopenapi.TaskServiceApi)
	log.Printf("Searching for task %s in project %s\n", taskId, projectId)

	var waiting = false
	for {
		res, _, err := taskCl.GetTask(getContext(), projectId, taskId).Execute()
		if err != nil {
			errFatal(fmt.Errorf("error getting task: %s", err.Error()))
		}

		task := res.GetTask()

		switch output() {
		case "table":
			if !waiting || task.GetDone() {
				convertSubTaskSliceToCSV(task)
			}
			if wait() && !task.GetDone() {
				waiting = true
				log.Printf("Waiting for task %s to complete", task.GetTaskId())
				time.Sleep(5 * time.Second)
				continue
			}

		case "json":
			taskJSON, err := res.MarshalJSON()
			if err != nil {
				errFatal(err)
			}

			err = PrettyPrintJSON(taskJSON)
			if err != nil {
				errFatal(err)
			}
		}
		break
	}
}
