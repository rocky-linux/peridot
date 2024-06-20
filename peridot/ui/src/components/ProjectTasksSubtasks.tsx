/*
 * Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
 * Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
 * Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 * this list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 * this list of conditions and the following disclaimer in the documentation
 * and/or other materials provided with the distribution.
 *
 * 3. Neither the name of the copyright holder nor the names of its contributors
 * may be used to endorse or promote products derived from this software without
 * specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

import React from 'react';
import classnames from 'classnames';
import Toolbar from '@mui/material/Toolbar';
import Divider from '@mui/material/Divider';
import {
  V1Subtask,
  V1TaskType,
  V1TaskStatus,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import ProjectTaskFullLog from 'peridot/ui/src/components/ProjectTaskFullLog';
import {
  transformTaskStatusToIcon,
  translateTaskTypeToText,
} from './ProjectTasks';


function formatDuration(ms) {
  const seconds = Math.floor((ms / 1000) % 60);
  const minutes = Math.floor((ms / (1000 * 60)) % 60);
  const hours = Math.floor((ms / (1000 * 60 * 60)) % 24);

  return [hours, minutes, seconds]
    .map(val => (val < 10 ? `0${val}` : val)) // Adding leading zeros if necessary
    .join(':');
}

export interface ProjectTasksSubtasksProps {
  subtasks: V1Subtask[];
}

export const ProjectTasksSubtasks = (props: ProjectTasksSubtasksProps) => {
  const [selectedTask, setSelectedTask] = React.useState<V1Subtask | undefined>(
    undefined
  );
  const [tabValue, setTabValue] = React.useState('1');

  const handleTabValueChange = (event, newValue) => {
    setTabValue(newValue);
  };

  return (
    <>
      <Divider />
      <div className="flex items-start h-full">
        <div className="w-2/6 h-full overflow-y-scroll">
          <Toolbar className="bg-gray-100 flex justify-between text-sm border-b">
            <span>Subtask</span>
            <span>Duration</span>
          </Toolbar>
          <div className="h-full bg-white text-sm divide-y">
            {props.subtasks.map((subtask) => {
              let subtaskDuration = <></>;
              if (subtask.finishedAt && subtask.createdAt) {
                const difference =
                  (new Date(subtask.finishedAt) as any) -
                  (new Date(subtask.createdAt) as any);
                subtaskDuration = (
                  <>{formatDuration(difference)}</>
                );
              }

              return (
                <div
                  className={classnames(
                    'px-6 py-3 hover:bg-gray-100 focus:bg-gray-100 cursor-pointer flex justify-between',
                    selectedTask &&
                      selectedTask.id == subtask.id &&
                      'bg-gray-100'
                  )}
                  onClick={() => setSelectedTask(subtask)}
                >
                  <div>
                    {translateTaskTypeToText(
                      subtask.type || V1TaskType.Unspecified
                    )}
                    <div className="text-xs text-gray-600">{subtask.arch}</div>
                  </div>
                  <div className="text-xs flex items-center space-x-2">
                    <div>{subtaskDuration}</div>
                    <div>{transformTaskStatusToIcon(subtask.status, false)}</div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
        {selectedTask && (
          <div className="w-4/6 h-full">
            <Box className="bg-white">
              <Tabs value={tabValue} onChange={handleTabValueChange}>
                <Tab label="Logs" value="1" />
                <Tab label="Details" value="2" />
              </Tabs>
            </Box>
            {tabValue === '1' && (
              <ProjectTaskFullLog taskId={selectedTask.id ?? ''} />
            )}
          </div>
        )}
      </div>
    </>
  );
};
