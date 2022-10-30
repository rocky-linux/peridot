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
import { RouteComponentProps } from 'react-router';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import { ProjectContext } from 'peridot/ui/src/context/ProjectContext';
import { fetchRemoteResource, suspenseRemoteResource } from 'common/ui/remote';
import { packageApi, taskApi } from 'peridot/ui/src/api';
import {
  V1GetTaskResponse,
  V1TaskType,
  V1TaskStatus,
} from 'bazel-bin/peridot/proto/v1/client_typescript';
import { ToolbarHeader } from 'common/mui/ToolbarHeader';
import Divider from '@mui/material/Divider';
import { PeridotLink } from 'common/ui/PeridotLink';
import {
  transformTaskStatusToIcon,
  transformTaskType,
  translateTaskTypeToText,
} from 'peridot/ui/src/components/ProjectTasks';
import { generateSuspensedComponent } from 'common/ui/SuspensedComponent';
import TabContext from '@mui/lab/TabContext';
import TabList from '@mui/lab/TabList';
import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Button from '@mui/material/Button';
import TabPanel from '@mui/lab/TabPanel';
import ProjectTaskFullLog from 'peridot/ui/src/components/ProjectTaskFullLog';
import { RemoteErrors } from 'common/ui/types';
import Tabs from '@mui/material/Tabs';
import Toolbar from '@mui/material/Toolbar';
import { ProjectTasksSubtasks } from 'peridot/ui/src/components/ProjectTasksSubtasks';
import { reqap } from 'common/ui/reqap';

export interface ProjectPackageDetailsRouteProps {
  id: string;
}

export default function (
  props: RouteComponentProps<ProjectPackageDetailsRouteProps>
) {
  const project = React.useContext(ProjectContext);
  if (!project) {
    return null;
  }

  const [taskRes, setTaskRes] = React.useState<
    V1GetTaskResponse | undefined | null | RemoteErrors
  >(undefined);
  const [tabValue, setTabValue] = React.useState('1');
  const [submitting, setSubmitting] = React.useState(false);

  const handleTabValueChange = (event, newValue) => {
    setTabValue(newValue);
  };

  fetchRemoteResource(
    () =>
      taskApi.getTask({
        id: props.match.params.id,
        projectId: project.id || '',
      }),
    setTaskRes
  );

  const cancelTask = async () => {
    if (!confirm('Are you sure you want to cancel this task?')) {
      return;
    }
    setSubmitting(true);
    const [err, res] = await reqap(() =>
      taskApi.cancelTask({
        id: props.match.params.id,
        projectId: project.id || '',
      })
    );
    if (err) {
      alert(err);
      setSubmitting(false);
      return;
    }

    window.location.reload();
  };

  return (
    <>
      {suspenseRemoteResource(taskRes, (res) => {
        const subtasks = res.task?.subtasks;
        if (!subtasks) {
          return null;
        }

        const parentTask = subtasks[0];
        const parentTaskMetadata = parentTask?.metadata as any;

        return (
          res.task && (
            <>
              <ToolbarHeader>
                <div className="flex space-x-4 items-center">
                  <div className="flex flex-col">
                    <h5 className="font-bold text-xs">Project</h5>
                    <h5 className="font-light text-xs">{project.name}</h5>
                  </div>
                  <ChevronRightIcon />
                  <div className="flex flex-col">
                    <h5 className="font-bold text-xs">Task ID</h5>
                    <h5 className="font-light text-xs">{res.task.taskId}</h5>
                  </div>
                </div>
              </ToolbarHeader>
              <Divider />
              <div className="w-full flex items-center justify-between bg-white">
                <div className="w-full bg-white flex divide-x divide-gray-100">
                  <div className="p-5 w-128 space-y-1">
                    <h2 className="font-bold text-lg">Type</h2>
                    <div className="flex space-x-2 h-6">
                      {translateTaskTypeToText(
                        parentTask.type || V1TaskType.Unspecified
                      )}
                    </div>
                  </div>
                  <div className="p-5 w-128 space-y-1">
                    <h2 className="font-bold text-lg">Status</h2>
                    <div className="flex space-x-2 h-6">
                      {transformTaskStatusToIcon(parentTask.status, true)}
                    </div>
                  </div>
                  {[V1TaskType.Import, V1TaskType.Build].includes(
                    parentTask.type || V1TaskType.Unspecified
                  ) && (
                    <div className="p-5 w-128 space-y-1">
                      <h2 className="font-bold text-lg">Package</h2>
                      <div className="flex space-x-2 h-6">
                        {parentTaskMetadata?.packageName}
                      </div>
                    </div>
                  )}
                </div>
                {!parentTask.parentTaskId &&
                  parentTask.status == V1TaskStatus.Running && (
                    <div className="p-5">
                      <Button
                        variant="contained"
                        disabled={submitting}
                        color="error"
                        onClick={cancelTask}
                      >
                        Cancel
                      </Button>
                    </div>
                  )}
              </div>
              <Divider />
              <Box className="bg-white">
                <Tabs value={tabValue} onChange={handleTabValueChange}>
                  <Tab label="Subtasks" value="1" />
                  <Tab label="Aggregated logs" value="2" />
                </Tabs>
              </Box>
              {tabValue === '1' && (
                <ProjectTasksSubtasks subtasks={res.task?.subtasks || []} />
              )}
              {tabValue === '2' && (
                <ProjectTaskFullLog taskId={props.match.params.id} parent />
              )}
            </>
          )
        );
      })}
    </>
  );
}
